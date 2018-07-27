package controllers

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"sync"
	"time"

	"emotibot.com/emotigo/module/admin-api/ELKStats/data"
	"emotibot.com/emotigo/module/admin-api/ELKStats/services"
	"emotibot.com/emotigo/module/admin-api/util"
)

var	visitStatsQueryHandlers = map[string]data.VisitStatsQueryHandler{
	data.VisitStatsMetricConversations:   services.ConversationCounts,
	data.VisitStatsMetricUniqueUsers:     services.UniqueUserCounts,
	data.VisitStatsMetricActiveUsers:     services.ActiveUserCounts,
	data.VisitStatsMetricNewUsers:        services.NewUserCounts,
	data.VisitStatsMetricTotalAsks:       services.TotalAskCounts,
	data.VisitStatsMetricNormalResponses: services.NormalResponseCounts,
	data.VisitStatsMetricChats:           services.ChatCounts,
	data.VisitStatsMetricOthers:          services.OtherCounts,
	data.VisitStatsMetricUnknownQnA:      services.UnknownQnACounts,
	// Note: 'Success rate' and 'Conversations per Session' are called separately
}

func VisitStatsGetHandler(w http.ResponseWriter, r *http.Request) {
	enterpriseID := util.GetEnterpriseID(r)
	appID := util.GetAppID(r)
	statsType := r.URL.Query().Get("type")
	statsFilter := r.URL.Query().Get("filter")
	t1 := r.URL.Query().Get("t1")
	t2 := r.URL.Query().Get("t2")

	if enterpriseID == "" && appID == "" {
		errResp := data.ErrorResponse{
			Message: fmt.Sprintf("Both headers %s and %s are not specified",
				data.EnterpriseIDHeaderKey, data.AppIDHeaderKey),
		}
		w.WriteHeader(http.StatusBadRequest)
		writeResponseJSON(w, errResp)
		return
	} else if statsType == "" {
		errResponse := data.NewBadRequestResponse(data.ErrCodeInvalidParameterType, "type")
		returnBadRequest(w, errResponse)
		return
	} else if t1 == "" {
		errResponse := data.NewBadRequestResponse(data.ErrCodeInvalidParameterT1, "t1")
		returnBadRequest(w, errResponse)
		return
	} else if t2 == "" {
		errResponse := data.NewBadRequestResponse(data.ErrCodeInvalidParameterT2, "t2")
		returnBadRequest(w, errResponse)
		return
	}

	startTime, endTime, err := util.CreateTimeRangeFromString(t1, t2, "20060102")
	if err != nil {
		errResponse := data.NewBadRequestResponse(data.ErrCodeInvalidParameterT2, "t1/t2")
		returnBadRequest(w, errResponse)
		return
	}

	var aggInterval string

	if t1 == t2 {
		aggInterval = data.IntervalHour
	} else {
		aggInterval = data.IntervalDay
	}

	var query data.VisitStatsQuery

	switch statsType {
	case data.VisitStatsTypeTime:
		query = data.VisitStatsQuery{
			CommonQuery: data.CommonQuery{
				EnterpriseID: enterpriseID,
				AppID:        appID,
				StartTime:    startTime,
				EndTime:      endTime,
			},
			AggBy:       data.AggByTime,
			AggInterval: aggInterval,
		}
	case data.VisitStatsTypeBarchart:
		if statsFilter == "" {
			errResponse := data.NewBadRequestResponse(data.ErrCodeInvalidParameterFilter, "filter")
			returnBadRequest(w, errResponse)
			return
		}

		switch statsFilter {
		case data.VisitStatsFilterCategory:
			statsCategory := r.URL.Query().Get("category")
			if statsCategory == "" {
				errResponse := data.NewBadRequestResponse(data.ErrCodeInvalidParameterCategory, "category")
				returnBadRequest(w, errResponse)
				return
			}

			query = data.VisitStatsQuery{
				CommonQuery: data.CommonQuery{
					EnterpriseID: enterpriseID,
					AppID:        appID,
					StartTime:    startTime,
					EndTime:      endTime,
				},
				AggBy:       data.AggByTag,
				AggInterval: aggInterval,
				AggTagType:  statsCategory,
			}
		case data.VisitStatsFilterQType:
			query = data.VisitStatsQuery{
				CommonQuery: data.CommonQuery{
					EnterpriseID: enterpriseID,
					AppID:        appID,
					StartTime:    startTime,
					EndTime:      endTime,
				},
			}
		default:
			errResponse := data.NewBadRequestResponse(data.ErrCodeInvalidParameterFilter, "filter")
			returnBadRequest(w, errResponse)
			return
		}
	default:
		errResponse := data.NewBadRequestResponse(data.ErrCodeInvalidParameterType, "type")
		returnBadRequest(w, errResponse)
		return
	}

	esCtx, esClient := util.GetElasticsearch()

	if statsType == data.VisitStatsTypeTime ||
		(statsType == data.VisitStatsTypeBarchart && statsFilter == data.VisitStatsFilterCategory) {
		statsCounts, err := fetchVisitStats(query)
		if err != nil {
			errResponse := data.NewErrorResponse(err.Error())
			returnInternalServerError(w, errResponse)
			return
		}

		switch statsType {
		case data.VisitStatsTypeTime:
			response, err := createVisitStatsResponse(query, statsCounts)
			if err != nil {
				errResponse := data.NewErrorResponse(err.Error())
				returnInternalServerError(w, errResponse)
				return
			}
			returnOK(w, response)
		case data.VisitStatsTypeBarchart:
			response, err := createVisitStatsTagResponse(query, statsCounts)
			if err != nil {
				errResponse := data.NewErrorResponse(err.Error())
				returnInternalServerError(w, errResponse)
				return
			}
			returnOK(w, response)
		default:
			errResponse := data.NewBadRequestResponse(data.ErrCodeInvalidParameterType, "type")
			returnBadRequest(w, errResponse)
			return
		}
	} else if statsType == data.VisitStatsTypeBarchart && statsFilter == data.VisitStatsFilterQType {
		// Return answer category counts
		statCounts, err := services.AnswerCategoryCounts(esCtx, esClient, query)
		if err != nil {
			errResponse := data.NewErrorResponse(err.Error())
			returnInternalServerError(w, errResponse)
			return
		}

		response, err := createAnswerCategoryStatsResponse(statCounts)
		if err != nil {
			errResponse := data.NewErrorResponse(err.Error())
			returnInternalServerError(w, errResponse)
			return
		}
		returnOK(w, response)
	} else {
		errResponse := data.NewBadRequestResponse(data.ErrCodeInvalidParameterType, "type")
		returnBadRequest(w, errResponse)
	}
}

func QuestionStatsGetHandler(w http.ResponseWriter, r *http.Request) {
	enterpriseID := util.GetEnterpriseID(r)
	appID := util.GetAppID(r)
	questionsType := r.URL.Query().Get("type")
	t1 := r.URL.Query().Get("t1")
	t2 := r.URL.Query().Get("t2")

	if enterpriseID == "" && appID == "" {
		errResp := data.ErrorResponse{
			Message: fmt.Sprintf("Both headers %s and %s are not specified",
				data.EnterpriseIDHeaderKey, data.AppIDHeaderKey),
		}
		w.WriteHeader(http.StatusBadRequest)
		writeResponseJSON(w, errResp)
		return
	} else if questionsType == "" {
		errResponse := data.NewBadRequestResponse(data.ErrCodeInvalidParameterType, "type")
		returnBadRequest(w, errResponse)
		return
	} else if t1 == "" {
		errResponse := data.NewBadRequestResponse(data.ErrCodeInvalidParameterT1, "t1")
		returnBadRequest(w, errResponse)
		return
	} else if t2 == "" {
		errResponse := data.NewBadRequestResponse(data.ErrCodeInvalidParameterT2, "t2")
		returnBadRequest(w, errResponse)
		return
	}

	startTime, endTime, err := util.CreateTimeRangeFromString(t1, t2, "20060102")
	if err != nil {
		errResponse := data.NewBadRequestResponse(data.ErrCodeInvalidParameterT2, "t1/t2")
		returnBadRequest(w, errResponse)
		return
	}

	query := data.VisitStatsQuery{
		CommonQuery: data.CommonQuery{
			EnterpriseID: enterpriseID,
			AppID:        appID,
			StartTime:    startTime,
			EndTime:      endTime,
		},
	}

	esCtx, esClient := util.GetElasticsearch()

	switch questionsType {
	case data.VisitQuestionsTypeTop:
		questions, err := services.TopQuestions(esCtx, esClient, query, 20)
		if err != nil {
			errResponse := data.NewErrorResponse(err.Error())
			returnInternalServerError(w, errResponse)
			return
		}

		response := createTopQuestionsResponse(questions)
		returnOK(w, response)
	case data.VisitQuestionsTypeUnused:
		var aggInterval string

		if t1 == t2 {
			aggInterval = data.IntervalHour
		} else {
			aggInterval = data.IntervalDay
		}

		query.AggInterval = aggInterval

		questions, err := services.TopUnmatchQuestions(esCtx, esClient, query, 20)
		if err != nil {
			errResponse := data.NewErrorResponse(err.Error())
			returnInternalServerError(w, errResponse)
			return
		}

		response, err := createTopUnmatchedQuestionsResponse(query, questions)
		if err != nil {
			errResponse := data.NewErrorResponse(err.Error())
			returnInternalServerError(w, errResponse)
			return
		}

		returnOK(w, response)
	default:
		errResponse := data.NewBadRequestResponse(data.ErrCodeInvalidParameterType, "type")
		returnBadRequest(w, errResponse)
	}
}

func fetchVisitStats(query data.VisitStatsQuery) (map[string]map[string]interface{}, error) {
	esCtx, esClient := util.GetElasticsearch()

	var visitStatsCountsSync sync.Map // Use sync.Map to avoid concurrent map writes
	visitStatsCounts := make(map[string]map[string]interface{})
	done := make(chan error, len(visitStatsQueryHandlers))
	var queryError error

	// Fetch statistics concurrently
	for queryKey, queryHandler := range visitStatsQueryHandlers {
		go func(key string, handler data.VisitStatsQueryHandler) {
			counts, err := handler(esCtx, esClient, query)
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

	successRates := services.SuccessRates(visitStatsCounts[data.VisitStatsMetricUnknownQnA],
		visitStatsCounts[data.VisitStatsMetricTotalAsks])
	conversationsPerSessions := services.CoversationsPerSessionCounts(visitStatsCounts[data.VisitStatsMetricConversations],
		visitStatsCounts[data.VisitStatsMetricTotalAsks])

	visitStatsCounts[data.VisitStatsMetricSuccessRate] = successRates
	visitStatsCounts[data.VisitStatsMetricConversationsPerSession] = conversationsPerSessions

	return visitStatsCounts, nil
}

func createVisitStatsResponse(query data.VisitStatsQuery,
	statsCounts map[string]map[string]interface{}) (*data.VisitStatsResponse, error) {
	visitStatsQ, totalVisitStatsQ, err := createVisitStatsQ(statsCounts)
	if err != nil {
		return nil, err
	}

	visitStatsQuantities := make(data.VisitStatsQuantities, 0)

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

		visitStatQuantity := data.VisitStatsQuantity{
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

	response := data.VisitStatsResponse{
		TableHeader: data.VisitStatsTableHeader,
		Data: data.VisitStatsData{
			VisitStatsQuantities: visitStatsQuantities,
			Type:                 query.AggInterval,
			Name:                 "提问数",
		},
		Total: data.VisitStatsTotal{
			VisitStatsQ: *totalVisitStatsQ,
			TimeText:    "合计",
			Time:        totalTime,
		},
	}

	return &response, nil
}

func createVisitStatsTagResponse(query data.VisitStatsQuery,
	statsCounts map[string]map[string]interface{}) (*data.VisitStatsTagResponse, error) {
	visitStatsQ, totalVisitStatsQ, err := createVisitStatsQ(statsCounts)
	if err != nil {
		return nil, err
	}

	tagData := make([]data.VisitStatsTagData, 0)

	for tagName, q := range visitStatsQ {
		tagID, found := services.GetTagIDByName(query.AggTagType, tagName)
		if !found {
			continue
		}

		t := data.VisitStatsTagData{
			Q:    *q,
			ID:   tagID,
			Name: tagName,
		}

		tagData = append(tagData, t)
	}

	response := data.VisitStatsTagResponse{
		TableHeader: data.VisitStatsTableHeader,
		Data:        tagData,
		Total:       *totalVisitStatsQ,
	}

	return &response, nil
}

func createAnswerCategoryStatsResponse(statCounts map[string]interface{}) (*data.AnswerCategoryStatsResponse, error) {
	businessStatData := data.NewAnswerCategoryStatData(data.CategoryBusiness, "业务类")
	businessStatData.Q.TotalAsks = statCounts[data.CategoryBusiness].(int64)

	chatStatData := data.NewAnswerCategoryStatData(data.CategoryChat, "聊天类")
	chatStatData.Q.TotalAsks = statCounts[data.CategoryChat].(int64)

	otherStatData := data.NewAnswerCategoryStatData(data.CategoryOther, "其他")
	otherStatData.Q.TotalAsks = statCounts[data.CategoryOther].(int64)

	answerCategoryStatsData := []data.AnswerCategoryStatData{
		*businessStatData,
		*chatStatData,
		*otherStatData,
	}

	response := data.AnswerCategoryStatsResponse{
		TableHeader: data.AnswerCategoryTableHeader,
		Data:        answerCategoryStatsData,
		Total:       *data.NewVisitStatsQ(),
	}

	return &response, nil
}

func createTopQuestionsResponse(questions data.Questions) *data.TopQuestionsResponse {
	questionsData := make([]data.TopQuestionData, 0)
	rank := 1

	for _, question := range questions {
		d := data.TopQuestionData{
			Q:        question.Count,
			Path:     "N/A",
			Question: question.Question,
			Rank:     rank,
		}

		questionsData = append(questionsData, d)
		rank++
	}

	return &data.TopQuestionsResponse{
		Data: questionsData,
	}
}

func createTopUnmatchedQuestionsResponse(query data.VisitStatsQuery,
	questions []*data.UnmatchQuestion) (*data.TopUnmatchedQuestionsResponse, error) {
	questionData := make([]data.TopUnmatchedQuestionData, 0)
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

		d := data.TopUnmatchedQuestionData{
			Question:      question.Question,
			Rank:          rank,
			Q:             question.Count,
			FirstTime:     strconv.FormatInt(firstTime.Unix(), 10),
			FirstTimeText: firstTime.Format(data.ESTimeFormat),
			LastTime:      strconv.FormatInt(lastTime.Unix(), 10),
			LastTimeText:  lastTime.Format(data.ESTimeFormat),
		}

		questionData = append(questionData, d)
		rank++
	}

	response := data.TopUnmatchedQuestionsResponse{
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
func createVisitStatsQ(statsCounts map[string]map[string]interface{}) (visitStatsQ map[string]*data.VisitStatsQ,
	totalVisitStatsQ *data.VisitStatsQ, err error) {
	visitStatsQ = make(map[string]*data.VisitStatsQ)
	totalVisitStatsQ = &data.VisitStatsQ{
		Unsolved:   0,
		SolvedRate: "N/A",
	}

	for statMetric, counts := range statsCounts {
		for key, count := range counts {
			if _, ok := visitStatsQ[key]; !ok {
				visitStatsQ[key] = &data.VisitStatsQ{
					Unsolved:   0,
					SolvedRate: "N/A",
				}
			}

			switch statMetric {
			case data.VisitStatsMetricConversations:
				visitStatsQ[key].Conversations = count.(int64)
				totalVisitStatsQ.Conversations += count.(int64)
			case data.VisitStatsMetricUniqueUsers:
				visitStatsQ[key].UniqueUsers = count.(int64)
				totalVisitStatsQ.UniqueUsers += count.(int64)
			case data.VisitStatsMetricActiveUsers:
				visitStatsQ[key].ActiveUsers = count.(int64)
				totalVisitStatsQ.ActiveUsers += count.(int64)
			case data.VisitStatsMetricNewUsers:
				visitStatsQ[key].NewUsers = count.(int64)
				totalVisitStatsQ.NewUsers += count.(int64)
			case data.VisitStatsMetricTotalAsks:
				visitStatsQ[key].TotalAsks = count.(int64)
				totalVisitStatsQ.TotalAsks += count.(int64)
			case data.VisitStatsMetricNormalResponses:
				visitStatsQ[key].NormalResponses = count.(int64)
				totalVisitStatsQ.NormalResponses += count.(int64)
			case data.VisitStatsMetricChats:
				visitStatsQ[key].Chats = count.(int64)
				totalVisitStatsQ.Chats += count.(int64)
			case data.VisitStatsMetricOthers:
				visitStatsQ[key].Others = count.(int64)
				totalVisitStatsQ.Others += count.(int64)
			case data.VisitStatsMetricUnknownQnA:
				visitStatsQ[key].UnknownQnA = count.(int64)
				totalVisitStatsQ.UnknownQnA += count.(int64)
			case data.VisitStatsMetricSuccessRate:
				visitStatsQ[key].SuccessRate = count.(string)
			case data.VisitStatsMetricConversationsPerSession:
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
