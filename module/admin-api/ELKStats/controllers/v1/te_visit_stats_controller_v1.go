package v1

import (
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"emotibot.com/emotigo/module/admin-api/ELKStats/controllers"
	"emotibot.com/emotigo/module/admin-api/ELKStats/data"
	dataCommon "emotibot.com/emotigo/module/admin-api/ELKStats/data/common"
	dataV1 "emotibot.com/emotigo/module/admin-api/ELKStats/data/v1"
	"emotibot.com/emotigo/module/admin-api/ELKStats/services"
	servicesV1 "emotibot.com/emotigo/module/admin-api/ELKStats/services/v1"
	"emotibot.com/emotigo/module/admin-api/util/elasticsearch"
	"emotibot.com/emotigo/module/admin-api/util/requestheader"
)

var teVisitStatsQueryHandlers = map[string]dataV1.TEVisitStatsQueryHandler{
	dataCommon.TEVisitStatsMetricTriggers:   servicesV1.TriggerCounts,
	dataCommon.TEVisitStatsMetricUnfinished: servicesV1.UnfinishedCounts,
}

func TEVisitStatsGetHandler(w http.ResponseWriter, r *http.Request) {
	appID := requestheader.GetAppID(r)
	startTimeString := r.URL.Query().Get("startTime")
	endTimeString := r.URL.Query().Get("endTime")
	scenarioID := r.URL.Query().Get("scenarioID")
	scenarioName := r.URL.Query().Get("scenarioName")
	statsType := r.URL.Query().Get("type")

	if startTimeString == "" {
		errResponse := data.NewBadRequestResponse(data.ErrCodeInvalidParameterStartTime, "startTime")
		controllers.ReturnBadRequest(w, errResponse)
		return
	} else if endTimeString == "" {
		errResponse := data.NewBadRequestResponse(data.ErrCodeInvalidParameterEndTime, "endTime")
		controllers.ReturnBadRequest(w, errResponse)
		return
	} else if statsType == "" {
		errResponse := data.NewBadRequestResponse(data.ErrCodeInvalidParameterType, "type")
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

	var aggBy string

	switch statsType {
	case dataCommon.TEVisitStatsTypeTime:
		aggBy = data.AggByTime
	case dataCommon.TEVisitStatsTypeDimension:
		aggBy = data.AggByTag
	default:
		errResponse := data.NewBadRequestResponse(data.ErrCodeInvalidParameterType, "type")
		controllers.ReturnBadRequest(w, errResponse)
		return
	}

	var dimension string

	if statsType == dataCommon.TEVisitStatsTypeDimension {
		dimension = r.URL.Query().Get("dimension")

		if dimension == "" {
			errResponse := data.NewBadRequestResponse(data.ErrCodeInvalidParameterDimension, "dimension")
			controllers.ReturnBadRequest(w, errResponse)
			return
		}
	}

	startTime := time.Unix(int64(startTimeUnix), 0).UTC()
	endTime := time.Unix(int64(endTimeUnix), 0).UTC()

	if startTime.After(endTime) {
		errResponse := data.NewErrorResponseWithCode(data.ErrCodeInvalidParameterStartTime,
			"startTime cannot greater than endTime")
		controllers.ReturnBadRequest(w, errResponse)
		return
	}

	var aggInterval string

	if endTime.Sub(startTime) < data.DurationDay {
		aggInterval = data.IntervalHour
	} else {
		aggInterval = data.IntervalDay
	}

	query := dataV1.TEVisitStatsQuery{
		StatsQuery: data.StatsQuery{
			CommonQuery: data.CommonQuery{
				AppID:     appID,
				StartTime: startTime,
				EndTime:   endTime,
			},
			AggBy:       aggBy,
			AggInterval: aggInterval,
			AggTagType:  dimension,
		},
		ScenarioID:   scenarioID,
		ScenarioName: scenarioName,
	}

	teStatsCounts, err := fetchTEVisitStats(query)
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

	switch statsType {
	case dataCommon.TEVisitStatsTypeTime:
		response, err := createTEVisitStatsResponse(query, teStatsCounts)
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
	case dataCommon.TEVisitStatsTypeDimension:
		response, err := createTEVisitStatsTagResponse(query, teStatsCounts)
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
	default:
		errResponse := data.NewBadRequestResponse(data.ErrCodeInvalidParameterType, "type")
		controllers.ReturnBadRequest(w, errResponse)
		return
	}
}

func fetchTEVisitStats(query dataV1.TEVisitStatsQuery) (map[string]map[string]interface{}, error) {
	var teVisitStatsCountsSync sync.Map // Use sync.Map to avoid concurrent map writes
	teVisitStatsCounts := make(map[string]map[string]interface{})
	done := make(chan error, len(teVisitStatsQueryHandlers))
	var queryError error

	// Fetch statistics concurrently
	for queryKey, queryHandler := range teVisitStatsQueryHandlers {
		go func(key string, handler dataV1.TEVisitStatsQueryHandler) {
			counts, err := handler(query)
			if err != nil {
				done <- err
				return
			}

			teVisitStatsCountsSync.Store(key, counts)
			done <- nil
		}(queryKey, queryHandler)
	}

	// Wait for all queries complete
	for range teVisitStatsQueryHandlers {
		err := <-done
		if err != nil {
			queryError = err
		}
	}

	if queryError != nil {
		return nil, queryError
	}

	// Copy the values from sync.Map to normal map
	teVisitStatsCountsSync.Range(func(key, value interface{}) bool {
		teVisitStatsCounts[key.(string)] = value.(map[string]interface{})
		return true
	})

	return teVisitStatsCounts, nil
}

func createTEVisitStatsResponse(query dataV1.TEVisitStatsQuery,
	teStatsCounts map[string]map[string]interface{}) (*dataV1.TEVisitStatsResponse, error) {
	teVisitStatsQ, totalTEVisitStatsQ, err := createTEVisitStatsQ(teStatsCounts)
	if err != nil {
		return nil, err
	}

	teVisitStatsQuantities := make(dataV1.TEVisitStatsQuantities, 0)

	for datetime, q := range teVisitStatsQ {
		var timeText string

		_time, err := time.Parse(data.ESTimeFormat, datetime)
		if err != nil {
			return nil, err
		}

		timeUnixSec := _time.Unix()

		switch query.AggInterval {
		case data.IntervalDay:
			timeText = _time.Format("2006-01-02")
		case data.IntervalHour:
			timeText = _time.Format("15:04")
		default:
			return nil, data.ErrInvalidAggTimeInterval
		}

		teVisitStatQuantity := dataV1.TEVisitStatsQuantity{
			TEVisitStatsQ: *q,
			TimeText:      timeText,
			Time:          strconv.FormatInt(timeUnixSec, 10),
		}

		teVisitStatsQuantities = append(teVisitStatsQuantities, teVisitStatQuantity)
	}

	// Sort teVisitStatsQuantities
	sort.Sort(teVisitStatsQuantities)

	response := dataV1.TEVisitStatsResponse{
		TableHeader: dataV1.TEVisitStatsTableHeader,
		Data: dataV1.TEVisitStatsData{
			TEVisitStatsQuantities: teVisitStatsQuantities,
			Type:                   query.AggInterval,
		},
		Total: dataV1.TEVisitStatsTotal{
			TEVisitStatsQ: *totalTEVisitStatsQ,
		},
	}

	return &response, nil
}

func createTEVisitStatsTagResponse(query dataV1.TEVisitStatsQuery,
	teStatsCounts map[string]map[string]interface{}) (*dataV1.TEVisitStatsTagResponse, error) {
	teVisitStatsQ, totalTEVisitStatsQ, err := createTEVisitStatsQ(teStatsCounts)
	if err != nil {
		return nil, err
	}

	tagData := make([]dataV1.TEVisitStatsTagData, 0)

	for tagName, q := range teVisitStatsQ {
		tagID, found := services.GetTagIDByName(query.AppID, query.AggTagType, tagName)
		if !found {
			continue
		}

		t := dataV1.TEVisitStatsTagData{
			Q:    *q,
			ID:   tagID,
			Name: tagName,
		}

		tagData = append(tagData, t)
	}

	response := dataV1.TEVisitStatsTagResponse{
		TableHeader: dataV1.TEVisitStatsTagTableHeader,
		Data:        tagData,
		Total:       *totalTEVisitStatsQ,
	}

	return &response, nil
}

// createTEVisitStatsQ converts statsCounts = {
// 	   triggers: {
// 	       2018-01-01 00:00:00 / android: 2,
// 	       2018-01-01 01:00:00 / ios: 4,
// 	       2018-01-01 02:00:00 / 微信: 9,
// 	       ...
// 	   },
// 	   unfinished: {
// 		   2018-01-01 00:00:00 / android: 1,
// 		   2018-01-01 01:00:00 / io): 13,
// 		   2018-01-01 02:00:00 / 微信: 5,
// 		   ...
// 	   },
// 	   ...
// }
//
// to the form: {
// 	   2018-01-01 00:00:00 / android: {
// 		   triggers: 2,
// 		   unfinished: 1,
// 		   ...
// 	   },
// 	   2018-01-01 01:00:00 / ios: {
// 		   triggers: 4,
// 		   unfinished: 13,
// 		   ...
// 	   },
// 	   2018-01-01 02:00:00 / 微信: {
// 		   triggers: 9,
// 		   unfinished: 5,
// 		   ...
// 	   },
// 	   ...
// }
func createTEVisitStatsQ(
	teStatsCounts map[string]map[string]interface{}) (teVisitStatsQ map[string]*dataV1.TEVisitStatsQ,
	totalTEVisitStatsQ *dataV1.TEVisitStatsQ, err error) {
	teVisitStatsQ = make(map[string]*dataV1.TEVisitStatsQ)
	totalTEVisitStatsQ = &dataV1.TEVisitStatsQ{}

	for teStatMetric, counts := range teStatsCounts {
		for key, count := range counts {
			if _, ok := teVisitStatsQ[key]; !ok {
				teVisitStatsQ[key] = &dataV1.TEVisitStatsQ{}
			}

			switch teStatMetric {
			case dataCommon.TEVisitStatsMetricTriggers:
				teVisitStatsQ[key].Triggers = count.(int64)
				totalTEVisitStatsQ.Triggers += count.(int64)
			case dataCommon.TEVisitStatsMetricUnfinished:
				teVisitStatsQ[key].Unfinished = count.(int64)
				totalTEVisitStatsQ.Unfinished += count.(int64)
			}
		}
	}

	return teVisitStatsQ, totalTEVisitStatsQ, nil
}
