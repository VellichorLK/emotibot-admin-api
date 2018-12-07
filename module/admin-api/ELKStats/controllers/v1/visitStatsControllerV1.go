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

var visitStatsQueryHandlers = map[string]dataV1.VisitStatsQueryHandler{
	dataCommon.VisitStatsMetricConversations:   servicesV1.ConversationCounts,
	dataCommon.VisitStatsMetricUniqueUsers:     servicesV1.UniqueUserCounts,
	dataCommon.VisitStatsMetricNewUsers:        servicesV1.NewUserCounts,
	dataCommon.VisitStatsMetricTotalAsks:       servicesV1.TotalAskCounts,
	dataCommon.VisitStatsMetricNormalResponses: servicesV1.NormalResponseCounts,
	dataCommon.VisitStatsMetricChats:           servicesV1.ChatCounts,
	dataCommon.VisitStatsMetricOthers:          servicesV1.OtherCounts,
	dataCommon.VisitStatsMetricUnknownQnA:      servicesV1.UnknownQnACounts,
	// Note: 'Success rate' and 'Conversations per Session' are called separately
}

func VisitStatsGetHandler(w http.ResponseWriter, r *http.Request) {
	appID := requestheader.GetAppID(r)
	statsType := r.URL.Query().Get("type")
	statsFilter := r.URL.Query().Get("filter")
	t1 := r.URL.Query().Get("t1")
	t2 := r.URL.Query().Get("t2")

	if statsType == "" {
		errResponse := data.NewBadRequestResponse(data.ErrCodeInvalidParameterType, "type")
		controllers.ReturnBadRequest(w, errResponse)
		return
	} else if t1 == "" {
		errResponse := data.NewBadRequestResponse(data.ErrCodeInvalidParameterStartTime, "t1")
		controllers.ReturnBadRequest(w, errResponse)
		return
	} else if t2 == "" {
		errResponse := data.NewBadRequestResponse(data.ErrCodeInvalidParameterEndTime, "t2")
		controllers.ReturnBadRequest(w, errResponse)
		return
	}

	startTime, endTime, err := elasticsearch.CreateTimeRangeFromString(t1, t2, "20060102")
	if err != nil {
		errResponse := data.NewBadRequestResponse(data.ErrCodeInvalidParameterEndTime, "t1/t2")
		controllers.ReturnBadRequest(w, errResponse)
		return
	}

	var aggInterval string

	if t1 == t2 {
		aggInterval = data.IntervalHour
	} else {
		aggInterval = data.IntervalDay
	}

	var query dataV1.VisitStatsQuery

	switch statsType {
	case dataCommon.VisitStatsTypeTime:
		query = dataV1.VisitStatsQuery{
			StatsQuery: data.StatsQuery{
				CommonQuery: data.CommonQuery{
					AppID:     appID,
					StartTime: startTime,
					EndTime:   endTime,
				},
				AggBy:       data.AggByTime,
				AggInterval: aggInterval,
			},
		}
	case dataCommon.VisitStatsTypeBarchart:
		if statsFilter == "" {
			errResponse := data.NewBadRequestResponse(data.ErrCodeInvalidParameterFilter, "filter")
			controllers.ReturnBadRequest(w, errResponse)
			return
		}

		switch statsFilter {
		case dataCommon.VisitStatsFilterCategory:
			statsCategory := r.URL.Query().Get("category")
			if statsCategory == "" {
				errResponse := data.NewBadRequestResponse(data.ErrCodeInvalidParameterCategory, "category")
				controllers.ReturnBadRequest(w, errResponse)
				return
			}

			query = dataV1.VisitStatsQuery{
				StatsQuery: data.StatsQuery{
					CommonQuery: data.CommonQuery{
						AppID:     appID,
						StartTime: startTime,
						EndTime:   endTime,
					},
					AggBy:       data.AggByTag,
					AggInterval: aggInterval,
					AggTagType:  statsCategory,
				},
			}
		case dataCommon.VisitStatsFilterQType:
			query = dataV1.VisitStatsQuery{
				StatsQuery: data.StatsQuery{
					CommonQuery: data.CommonQuery{
						AppID:     appID,
						StartTime: startTime,
						EndTime:   endTime,
					},
				},
			}
		default:
			errResponse := data.NewBadRequestResponse(data.ErrCodeInvalidParameterFilter, "filter")
			controllers.ReturnBadRequest(w, errResponse)
			return
		}
	default:
		errResponse := data.NewBadRequestResponse(data.ErrCodeInvalidParameterType, "type")
		controllers.ReturnBadRequest(w, errResponse)
		return
	}

	if statsType == dataCommon.VisitStatsTypeTime ||
		(statsType == dataCommon.VisitStatsTypeBarchart && statsFilter == dataCommon.VisitStatsFilterCategory) {
		statsCounts, err := fetchVisitStats(query)
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
		case dataCommon.VisitStatsTypeTime:
			response, err := createVisitStatsResponse(query, statsCounts)
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
		case dataCommon.VisitStatsTypeBarchart:
			response, err := createVisitStatsTagResponse(query, statsCounts)
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
	} else if statsType == dataCommon.VisitStatsTypeBarchart && statsFilter == dataCommon.VisitStatsFilterQType {
		// Return answer category counts
		statCounts, err := servicesV1.AnswerCategoryCounts(query)
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

		response, err := createAnswerCategoryStatsResponse(statCounts)
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
	} else {
		errResponse := data.NewBadRequestResponse(data.ErrCodeInvalidParameterType, "type")
		controllers.ReturnBadRequest(w, errResponse)
	}
}

func QuestionStatsGetHandler(w http.ResponseWriter, r *http.Request) {
	appID := requestheader.GetAppID(r)
	questionsType := r.URL.Query().Get("type")
	t1 := r.URL.Query().Get("t1")
	t2 := r.URL.Query().Get("t2")

	if questionsType == "" {
		errResponse := data.NewBadRequestResponse(data.ErrCodeInvalidParameterType, "type")
		controllers.ReturnBadRequest(w, errResponse)
		return
	} else if t1 == "" {
		errResponse := data.NewBadRequestResponse(data.ErrCodeInvalidParameterStartTime, "t1")
		controllers.ReturnBadRequest(w, errResponse)
		return
	} else if t2 == "" {
		errResponse := data.NewBadRequestResponse(data.ErrCodeInvalidParameterEndTime, "t2")
		controllers.ReturnBadRequest(w, errResponse)
		return
	}

	startTime, endTime, err := elasticsearch.CreateTimeRangeFromString(t1, t2, "20060102")
	if err != nil {
		errResponse := data.NewBadRequestResponse(data.ErrCodeInvalidParameterEndTime, "t1/t2")
		controllers.ReturnBadRequest(w, errResponse)
		return
	}

	query := dataV1.VisitStatsQuery{
		StatsQuery: data.StatsQuery{
			CommonQuery: data.CommonQuery{
				AppID:     appID,
				StartTime: startTime,
				EndTime:   endTime,
			},
		},
	}

	switch questionsType {
	case dataCommon.VisitQuestionsTypeTop:
		questions, err := servicesV1.TopQuestions(query, 20)
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

		response := createTopQuestionsResponse(questions)
		controllers.ReturnOK(w, response)
	case dataCommon.VisitQuestionsTypeUnused:
		var aggInterval string

		if t1 == t2 {
			aggInterval = data.IntervalHour
		} else {
			aggInterval = data.IntervalDay
		}

		query.AggInterval = aggInterval

		questions, err := servicesV1.TopUnmatchQuestions(query, 20)
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

		response, err := createTopUnmatchedQuestionsResponse(query, questions)
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
	}
}

func fetchVisitStats(query dataV1.VisitStatsQuery) (map[string]map[string]interface{}, error) {
	var visitStatsCountsSync sync.Map // Use sync.Map to avoid concurrent map writes
	visitStatsCounts := make(map[string]map[string]interface{})
	done := make(chan error, len(visitStatsQueryHandlers))
	var queryError error

	// Fetch statistics concurrently
	for queryKey, queryHandler := range visitStatsQueryHandlers {
		go func(key string, handler dataV1.VisitStatsQueryHandler) {
			counts, err := handler(query)
			if err != nil {
				done <- err
				return
			}

			visitStatsCountsSync.Store(key, counts)
			done <- nil
		}(queryKey, queryHandler)
	}

	// Wait for all queries complete
	for range visitStatsQueryHandlers {
		err := <-done
		if err != nil {
			queryError = err
		}
	}

	if queryError != nil {
		return nil, queryError
	}

	// Copy the values from sync.Map to normal map
	visitStatsCountsSync.Range(func(key, value interface{}) bool {
		visitStatsCounts[key.(string)] = value.(map[string]interface{})
		return true
	})

	successRates := servicesV1.SuccessRates(visitStatsCounts[dataCommon.VisitStatsMetricUnknownQnA],
		visitStatsCounts[dataCommon.VisitStatsMetricTotalAsks])
	conversationsPerSessions := servicesV1.
		CoversationsPerSessionCounts(visitStatsCounts[dataCommon.VisitStatsMetricConversations],
			visitStatsCounts[dataCommon.VisitStatsMetricTotalAsks])

	visitStatsCounts[dataCommon.VisitStatsMetricSuccessRate] = successRates
	visitStatsCounts[dataCommon.VisitStatsMetricConversationsPerSession] = conversationsPerSessions

	return visitStatsCounts, nil
}

func createVisitStatsResponse(query dataV1.VisitStatsQuery,
	statsCounts map[string]map[string]interface{}) (*dataV1.VisitStatsResponse, error) {
	visitStatsQ, totalVisitStatsQ, err := createVisitStatsQ(statsCounts)
	if err != nil {
		return nil, err
	}

	visitStatsQuantities := make(dataV1.VisitStatsQuantities, 0)

	for datetime, q := range visitStatsQ {
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

		visitStatQuantity := dataV1.VisitStatsQuantity{
			VisitStatsQ: *q,
			TimeText:    timeText,
			Time:        strconv.FormatInt(timeUnixSec, 10),
		}

		visitStatsQuantities = append(visitStatsQuantities, visitStatQuantity)
	}

	// Sort visitStatsQuantities
	sort.Sort(visitStatsQuantities)

	var totalTime string
	switch query.AggInterval {
	case data.IntervalYear:
		fallthrough
	case data.IntervalMonth:
		fallthrough
	case data.IntervalDay:
		totalTime = ""
	case data.IntervalHour:
		fallthrough
	case data.IntervalMinute:
		fallthrough
	case data.IntervalSecond:
		totalTime = strconv.FormatInt(query.StartTime.Unix(), 10)
	default:
		return nil, data.ErrInvalidAggTimeInterval
	}

	response := dataV1.VisitStatsResponse{
		TableHeader: dataV1.VisitStatsTableHeader,
		Data: dataV1.VisitStatsData{
			VisitStatsQuantities: visitStatsQuantities,
			Type:                 query.AggInterval,
			Name:                 "提问数",
		},
		Total: dataV1.VisitStatsTotal{
			VisitStatsQ: *totalVisitStatsQ,
			TimeText:    "合计",
			Time:        totalTime,
		},
	}

	return &response, nil
}

func createVisitStatsTagResponse(query dataV1.VisitStatsQuery,
	statsCounts map[string]map[string]interface{}) (*dataV1.VisitStatsTagResponse, error) {
	visitStatsQ, totalVisitStatsQ, err := createVisitStatsQ(statsCounts)
	if err != nil {
		return nil, err
	}

	tagData := make([]dataV1.VisitStatsTagData, 0)

	for tagName, q := range visitStatsQ {
		tagID, found := services.GetTagIDByName(query.AppID, query.AggTagType, tagName)
		if !found {
			continue
		}

		t := dataV1.VisitStatsTagData{
			Q:    *q,
			ID:   tagID,
			Name: tagName,
		}

		tagData = append(tagData, t)
	}

	response := dataV1.VisitStatsTagResponse{
		TableHeader: dataV1.VisitStatsTagTableHeader,
		Data:        tagData,
		Total:       *totalVisitStatsQ,
	}

	return &response, nil
}

func createAnswerCategoryStatsResponse(
	statCounts map[string]interface{}) (*dataV1.AnswerCategoryStatsResponse, error) {
	businessStatData := dataV1.NewAnswerCategoryStatData(dataCommon.CategoryBusiness, "业务类")
	businessStatData.Q.TotalAsks = statCounts[dataCommon.CategoryBusiness].(int64)

	chatStatData := dataV1.NewAnswerCategoryStatData(dataCommon.CategoryChat, "聊天类")
	chatStatData.Q.TotalAsks = statCounts[dataCommon.CategoryChat].(int64)

	otherStatData := dataV1.NewAnswerCategoryStatData(dataCommon.CategoryOther, "其他")
	otherStatData.Q.TotalAsks = statCounts[dataCommon.CategoryOther].(int64)

	answerCategoryStatsData := []dataV1.AnswerCategoryStatData{
		*businessStatData,
		*chatStatData,
		*otherStatData,
	}

	response := dataV1.AnswerCategoryStatsResponse{
		TableHeader: dataV1.AnswerCategoryTableHeader,
		Data:        answerCategoryStatsData,
		Total:       *dataV1.NewVisitStatsQ(),
	}

	return &response, nil
}

func createTopQuestionsResponse(questions dataV1.Questions) *dataV1.TopQuestionsResponse {
	// Sort top questions
	sort.Sort(sort.Reverse(questions))

	questionsData := make([]dataV1.TopQuestionData, 0)
	rank := 1

	for _, question := range questions {
		d := dataV1.TopQuestionData{
			Q:        question.Count,
			Path:     "N/A",
			Question: question.Question,
			Rank:     rank,
		}

		questionsData = append(questionsData, d)
		rank++
	}

	return &dataV1.TopQuestionsResponse{
		Data: questionsData,
	}
}

func createTopUnmatchedQuestionsResponse(query dataV1.VisitStatsQuery,
	questions []*dataV1.UnmatchQuestion) (*dataV1.TopUnmatchedQuestionsResponse, error) {
	questionData := make([]dataV1.TopUnmatchedQuestionData, 0)
	rank := 1

	for _, question := range questions {
		firstTime, err := time.Parse(data.ESTimeFormat, question.MinLogTime)
		if err != nil {
			return nil, err
		}

		lastTime, err := time.Parse(data.ESTimeFormat, question.MaxLogTime)
		if err != nil {
			return nil, err
		}

		d := dataV1.TopUnmatchedQuestionData{
			Question:      question.Question,
			Rank:          rank,
			Q:             question.Count,
			FirstTime:     strconv.FormatInt(firstTime.Unix(), 10),
			FirstTimeText: question.MinLogTime,
			LastTime:      strconv.FormatInt(lastTime.Unix(), 10),
			LastTimeText:  question.MaxLogTime,
		}

		questionData = append(questionData, d)
		rank++
	}

	response := dataV1.TopUnmatchedQuestionsResponse{
		Data: questionData,
	}

	return &response, nil
}

// createVisitStatsQ converts statsCounts = {
// 	   conversations: {
// 	       2018-01-01 00:00:00 / android: 2,
// 	       2018-01-01 01:00:00 / ios: 4,
// 	       2018-01-01 02:00:00 / 微信: 9,
// 	       ...
// 	   },
// 	   unique_users: {
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
// 		   conversations: 2,
// 		   unique_users: 1,
// 		   ...
// 	   },
// 	   2018-01-01 01:00:00 / ios: {
// 		   conversations: 4,
// 		   unique_users: 13,
// 		   ...
// 	   },
// 	   2018-01-01 02:00:00 / 微信: {
// 		   conversations: 9,
// 		   unique_users: 5,
// 		   ...
// 	   },
// 	   ...
// }
func createVisitStatsQ(
	statsCounts map[string]map[string]interface{}) (visitStatsQ map[string]*dataV1.VisitStatsQ,
	totalVisitStatsQ *dataV1.VisitStatsQ, err error) {
	visitStatsQ = make(map[string]*dataV1.VisitStatsQ)
	totalVisitStatsQ = &dataV1.VisitStatsQ{
		Unsolved:   0,
		SolvedRate: "N/A",
	}

	for statMetric, counts := range statsCounts {
		for key, count := range counts {
			if _, ok := visitStatsQ[key]; !ok {
				visitStatsQ[key] = &dataV1.VisitStatsQ{
					Unsolved:   0,
					SolvedRate: "N/A",
				}
			}

			switch statMetric {
			case dataCommon.VisitStatsMetricConversations:
				visitStatsQ[key].Conversations = count.(int64)
				totalVisitStatsQ.Conversations += count.(int64)
			case dataCommon.VisitStatsMetricUniqueUsers:
				visitStatsQ[key].UniqueUsers = count.(int64)
				totalVisitStatsQ.UniqueUsers += count.(int64)
			case dataCommon.VisitStatsMetricNewUsers:
				visitStatsQ[key].NewUsers = count.(int64)
				totalVisitStatsQ.NewUsers += count.(int64)
			case dataCommon.VisitStatsMetricTotalAsks:
				visitStatsQ[key].TotalAsks = count.(int64)
				totalVisitStatsQ.TotalAsks += count.(int64)
			case dataCommon.VisitStatsMetricNormalResponses:
				visitStatsQ[key].NormalResponses = count.(int64)
				totalVisitStatsQ.NormalResponses += count.(int64)
			case dataCommon.VisitStatsMetricChats:
				visitStatsQ[key].Chats = count.(int64)
				totalVisitStatsQ.Chats += count.(int64)
			case dataCommon.VisitStatsMetricOthers:
				visitStatsQ[key].Others = count.(int64)
				totalVisitStatsQ.Others += count.(int64)
			case dataCommon.VisitStatsMetricUnknownQnA:
				visitStatsQ[key].UnknownQnA = count.(int64)
				totalVisitStatsQ.UnknownQnA += count.(int64)
			case dataCommon.VisitStatsMetricSuccessRate:
				visitStatsQ[key].SuccessRate = count.(string)
			case dataCommon.VisitStatsMetricConversationsPerSession:
				visitStatsQ[key].ConversationPerSession = count.(string)
			}
		}
	}

	if totalVisitStatsQ.TotalAsks != 0 {
		totalVisitStatsQ.SuccessRate = strconv.
			FormatFloat((float64(totalVisitStatsQ.TotalAsks-totalVisitStatsQ.UnknownQnA) /
				float64(totalVisitStatsQ.TotalAsks)), 'f', 2, 64)
	} else {
		totalVisitStatsQ.SuccessRate = "N/A"
	}

	if totalVisitStatsQ.Conversations != 0 {
		totalVisitStatsQ.ConversationPerSession = strconv.
			FormatFloat(float64(totalVisitStatsQ.TotalAsks)/float64(totalVisitStatsQ.Conversations), 'f', 2, 64)
	} else {
		totalVisitStatsQ.ConversationPerSession = "N/A"
	}

	return visitStatsQ, totalVisitStatsQ, nil
}
