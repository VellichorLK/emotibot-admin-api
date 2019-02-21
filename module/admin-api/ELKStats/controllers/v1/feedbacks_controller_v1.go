package v1

import (
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"emotibot.com/emotigo/module/admin-api/ELKStats/controllers"
	"emotibot.com/emotigo/module/admin-api/ELKStats/data"
	dataCommon "emotibot.com/emotigo/module/admin-api/ELKStats/data/common"
	dataV1 "emotibot.com/emotigo/module/admin-api/ELKStats/data/v1"
	services "emotibot.com/emotigo/module/admin-api/ELKStats/services"
	servicesV1 "emotibot.com/emotigo/module/admin-api/ELKStats/services/v1"
	"emotibot.com/emotigo/module/admin-api/util/elasticsearch"
	"emotibot.com/emotigo/module/admin-api/util/requestheader"
)

func FeedbacksGetHandler(w http.ResponseWriter, r *http.Request) {
	appID := requestheader.GetAppID(r)
	startTimeString := r.URL.Query().Get("startTime")
	endTimeString := r.URL.Query().Get("endTime")
	statsType := r.URL.Query().Get("type")
	platform := r.URL.Query().Get("platform")
	gender := r.URL.Query().Get("gender")

	if startTimeString == "" {
		errResponse := data.NewBadRequestResponse(data.ErrCodeInvalidParameterStartTime, "startTime")
		controllers.ReturnBadRequest(w, errResponse)
		return
	} else if endTimeString == "" {
		errResponse := data.NewBadRequestResponse(data.ErrCodeInvalidParameterEndTime, "endTime")
		controllers.ReturnBadRequest(w, errResponse)
		return
	}

	startTimeUnix, err := strconv.Atoi(startTimeString)
	if err != nil {
		errResponse := data.NewBadRequestResponse(data.ErrCodeInvalidParameterStartTime, "startTime")
		controllers.ReturnBadRequest(w, errResponse)
		return
	}

	endTimeUnix, err := strconv.Atoi(endTimeString)
	if err != nil {
		errResponse := data.NewBadRequestResponse(data.ErrCodeInvalidParameterEndTime, "endTime")
		controllers.ReturnBadRequest(w, errResponse)
		return
	}

	startTime := time.Unix(int64(startTimeUnix), 0).UTC()
	endTime := time.Unix(int64(endTimeUnix), 0).UTC()

	query := dataV1.FeedbacksQuery{
		CommonQuery: data.CommonQuery{
			AppID:     appID,
			StartTime: startTime,
			EndTime:   endTime,
		},
	}

	if statsType != "" {
		if statsType != dataCommon.FeedbacksStatsTypeSessions &&
			statsType != dataCommon.FeedbacksStatsTypeRecords &&
			statsType != dataCommon.FeedbacksStatsTypeTERecords {
			errResponse := data.NewBadRequestResponse(data.ErrCodeInvalidParameterType, "type")
			controllers.ReturnBadRequest(w, errResponse)
			return
		}

		query.Type = statsType
	} else {
		// Default sessions feedbacks
		query.Type = dataCommon.FeedbacksStatsTypeSessions
	}

	if platform != "" {
		platformName, found := services.GetTagNameByID(query.AppID, "platform", platform)
		if !found {
			errResponse := data.NewBadRequestResponse(data.ErrCodeInvalidParameterPlatform, "platform")
			controllers.ReturnBadRequest(w, errResponse)
			return
		}

		query.Platform = platformName
	}

	if gender != "" {
		genderName, found := services.GetTagNameByID(query.AppID, "sex", gender)
		if !found {
			errResponse := data.NewBadRequestResponse(data.ErrCodeInvalidParameterGender, "gender")
			controllers.ReturnBadRequest(w, errResponse)
			return
		}

		query.Gender = genderName
	}

	statsCounts, err := fetchFeedbackStats(query)
	if err != nil {
		var errResponse data.ErrorResponse
		if rootCauseErrors, ok := elasticsearch.ExtractElasticsearchRootCauseErrors(err); ok {
			errResponse = data.NewErrorResponse(strings.Join(rootCauseErrors, "; "))
		} else {
			errResponse = data.NewErrorResponse(err.Error())
		}
		controllers.ReturnInternalServerError(w, errResponse)
		return
	}

	response, err := createFeedbacksResponse(query, statsCounts)
	if err != nil {
		var errResponse data.ErrorResponse
		if rootCauseErrors, ok := elasticsearch.ExtractElasticsearchRootCauseErrors(err); ok {
			errResponse = data.NewErrorResponse(strings.Join(rootCauseErrors, "; "))
		} else {
			errResponse = data.NewErrorResponse(err.Error())
		}
		controllers.ReturnInternalServerError(w, errResponse)
		return
	}

	controllers.ReturnOK(w, response)
}

func fetchFeedbackStats(query dataV1.FeedbacksQuery) (map[string]interface{}, error) {
	var feedbackStatsCountsSync sync.Map // Use sync.Map to avoid concurrent map writes
	feedbackStatsCounts := make(map[string]interface{})
	handlers := make([]func(), 0)
	var done chan error
	var queryError error

	// Average rating
	handlers = append(handlers, func() {
		avgRating, err := servicesV1.AvgRating(query)
		if err != nil {
			done <- err
			return
		}

		feedbackStatsCountsSync.Store(dataCommon.FeedbacksMetricAvgRating, avgRating)
		done <- nil
	})

	// Top 20 ratings
	handlers = append(handlers, func() {
		ratings, err := servicesV1.Ratings(query, 20)
		if err != nil {
			done <- err
			return
		}

		feedbackStatsCountsSync.Store(dataCommon.FeedbacksMetricRatings, ratings)
		done <- err
	})

	// Top 20 feedbacks
	handlers = append(handlers, func() {
		feedbacks, err := servicesV1.Feedbacks(query, 20)
		if err != nil {
			done <- err
			return
		}

		feedbackStatsCountsSync.Store(dataCommon.FeedbacksMetricFeedbacks, feedbacks)
		done <- err
	})

	done = make(chan error, len(handlers))

	// Fetch statistics concurrently
	for _, handler := range handlers {
		go handler()
	}

	// Wait for all queries complete
	for range handlers {
		err := <-done
		if err != nil {
			queryError = err
		}
	}

	if queryError != nil {
		return nil, queryError
	}

	// Copy the values from sync.Map to normal map
	feedbackStatsCountsSync.Range(func(key, value interface{}) bool {
		feedbackStatsCounts[key.(string)] = value
		return true
	})

	return feedbackStatsCounts, nil
}

func createFeedbacksResponse(query dataV1.FeedbacksQuery,
	feedbacks map[string]interface{}) (*dataV1.FeedbacksResponse, error) {

	response := dataV1.FeedbacksResponse{
		TableHeader: dataV1.FeedbacksTableHeader,
	}

	for key, value := range feedbacks {
		switch key {
		case dataCommon.FeedbacksMetricAvgRating:
			response.Data.AvgRating = value.(float64)
		case dataCommon.FeedbacksMetricRatings:
			ratingsMap := make(map[string]int64)

			ratings := value.(dataV1.FeedbackRatings)
			for _, rating := range ratings {
				ratingsMap[rating.Rating] = rating.Count
			}

			response.Data.Ratings = ratingsMap
		case dataCommon.FeedbacksMetricFeedbacks:
			feedbacksmap := make(map[string]int64)

			feedbacks := value.(dataV1.FeedbackCounts)
			for _, feedback := range feedbacks {
				feedbacksmap[feedback.Feedback] = feedback.Count
			}

			response.Data.Feedbacks = feedbacksmap
		}
	}

	return &response, nil
}

func FeedbackRatingAvgGetHandler(w http.ResponseWriter, r *http.Request) {
	appID := requestheader.GetAppID(r)
	startTimeString := r.URL.Query().Get("startTime")
	endTimeString := r.URL.Query().Get("endTime")
	platform := r.URL.Query().Get("platform")
	gender := r.URL.Query().Get("gender")

	if startTimeString == "" {
		errResponse := data.NewBadRequestResponse(data.ErrCodeInvalidParameterStartTime, "startTime")
		controllers.ReturnBadRequest(w, errResponse)
		return
	} else if endTimeString == "" {
		errResponse := data.NewBadRequestResponse(data.ErrCodeInvalidParameterEndTime, "endTime")
		controllers.ReturnBadRequest(w, errResponse)
		return
	}

	startTimeUnix, err := strconv.Atoi(startTimeString)
	if err != nil {
		errResponse := data.NewBadRequestResponse(data.ErrCodeInvalidParameterStartTime, "startTime")
		controllers.ReturnBadRequest(w, errResponse)
		return
	}

	endTimeUnix, err := strconv.Atoi(endTimeString)
	if err != nil {
		errResponse := data.NewBadRequestResponse(data.ErrCodeInvalidParameterEndTime, "endTime")
		controllers.ReturnBadRequest(w, errResponse)
		return
	}

	startTime := time.Unix(int64(startTimeUnix), 0).UTC()
	endTime := time.Unix(int64(endTimeUnix), 0).UTC()

	query := dataV1.FeedbacksQuery{
		CommonQuery: data.CommonQuery{
			AppID:     appID,
			StartTime: startTime,
			EndTime:   endTime,
		},
		Type: dataCommon.FeedbacksStatsTypeSessions,
	}

	if platform != "" {
		platformName, found := services.GetTagNameByID(query.AppID, "platform", platform)
		if !found {
			errResponse := data.NewBadRequestResponse(data.ErrCodeInvalidParameterPlatform, "platform")
			controllers.ReturnBadRequest(w, errResponse)
			return
		}

		query.Platform = platformName
	}

	if gender != "" {
		genderName, found := services.GetTagNameByID(query.AppID, "sex", gender)
		if !found {
			errResponse := data.NewBadRequestResponse(data.ErrCodeInvalidParameterGender, "gender")
			controllers.ReturnBadRequest(w, errResponse)
			return
		}

		query.Gender = genderName
	}

	avgRatingInfo, err := servicesV1.DailyAvgRating(query)
	if err != nil {
		var errResponse data.ErrorResponse
		if rootCauseErrors, ok := elasticsearch.ExtractElasticsearchRootCauseErrors(err); ok {
			errResponse = data.NewErrorResponse(strings.Join(rootCauseErrors, "; "))
		} else {
			errResponse = data.NewErrorResponse(err.Error())
		}
		controllers.ReturnInternalServerError(w, errResponse)
		return
	}

	controllers.ReturnOK(w, avgRatingInfo)
}
