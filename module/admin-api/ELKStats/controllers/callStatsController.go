package controllers

import (
	"net/http"
	"sort"
	"strconv"
	"sync"
	"time"

	"emotibot.com/emotigo/module/admin-api/ELKStats/data"
	"emotibot.com/emotigo/module/admin-api/ELKStats/services"
	"emotibot.com/emotigo/module/admin-api/util/elasticsearch"
	"emotibot.com/emotigo/module/admin-api/util/requestheader"
)

var callStatsQueryHandlers = map[string]data.CallStatsQueryHandler{
	data.CallsMetricTotals:    services.TotalCallCounts,
	data.CallsMetricCompletes: services.CompleteCallCounts,
	data.CallsMetricToHumans:  services.ToHumanCallCounts,
	data.CallsMetricTimeouts:  services.TimeoutCallCounts,
	data.CallsMetricCancels:   services.CancelCallCounts,
	data.CallsMetricUnknowns:  services.Unknowns,
	// Note: 'Completes rate', 'To Humans rate', 'Timeouts rate' and 'Cancels rate' are called separately
}

func CallStatsGetHandler(w http.ResponseWriter, r *http.Request) {
	appID := requestheader.GetAppID(r)
	statsType := r.URL.Query().Get("type")
	t1 := r.URL.Query().Get("t1")
	t2 := r.URL.Query().Get("t2")

	if statsType == "" {
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

	startTime, endTime, err := elasticsearch.CreateTimeRangeFromString(t1, t2, "20060102")
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

	query := data.CallStatsQuery{
		CommonQuery: data.CommonQuery{
			AppID:        appID,
			StartTime:    startTime,
			EndTime:      endTime,
		},
		AggInterval: aggInterval,
	}

	esCtx, esClient := elasticsearch.GetClient()
	if esClient == nil {
		returnInternalServerError(w, data.NewErrorResponse(data.ErrNotInit.Error()))
		return
	}

	switch statsType {
	case data.CallStatsTypeTime:
		callsCounts, err := fetchCallStats(query)
		if err != nil {
			errResponse := data.NewErrorResponse(err.Error())
			returnInternalServerError(w, errResponse)
			return
		}

		response, err := createCallStatsResponse(query, callsCounts)
		if err != nil {
			errResponse := data.NewErrorResponse(err.Error())
			returnInternalServerError(w, errResponse)
			return
		}

		returnOK(w, response)
	case data.CallStatsTypeAnswers:
		answers, err := services.TopToHumanAnswers(esCtx, esClient, query, 20)
		if err != nil {
			errResponse := data.NewErrorResponse(err.Error())
			returnInternalServerError(w, errResponse)
			return
		}

		response := createTopToHumanAnswersResponse(answers)
		returnOK(w, response)
	default:
		errResponse := data.NewBadRequestResponse(data.ErrCodeInvalidParameterType, "type")
		returnBadRequest(w, errResponse)
	}
}

func fetchCallStats(query data.CallStatsQuery) (map[string]map[string]interface{}, error) {
	esCtx, esClient := elasticsearch.GetClient()
	if esClient == nil {
		return nil, data.ErrNotInit
	}

	var callStatsCountsSync sync.Map // Use sync.Map to avoid concurrent map writes
	callStatsCounts := make(map[string]map[string]interface{})
	done := make(chan error, len(callStatsQueryHandlers))
	var queryError error

	// Fetch statistics concurrently
	for queryKey, queryHandler := range callStatsQueryHandlers {
		go func(key string, handler data.CallStatsQueryHandler) {
			counts, err := handler(esCtx, esClient, query)
			if err != nil {
				done <- err
				return
			}

			callStatsCountsSync.Store(key, counts)
			done <- nil
		}(queryKey, queryHandler)
	}

	// Wait for all queries complete
	for range callStatsQueryHandlers {
		err := <-done
		if err != nil {
			queryError = err
		}
	}

	if queryError != nil {
		return nil, queryError
	}

	// Copy the values from sync.Map to normal map
	callStatsCountsSync.Range(func(key, value interface{}) bool {
		callStatsCounts[key.(string)] = value.(map[string]interface{})
		return true
	})

	completesCallRates := services.CompleteCallRates(callStatsCounts[data.CallsMetricCompletes],
		callStatsCounts[data.CallsMetricTotals])
	trasferToHumanCallRates := services.ToHumanCallRates(callStatsCounts[data.CallsMetricToHumans],
		callStatsCounts[data.CallsMetricTotals])
	timeoutCallRates := services.TimeoutCallRates(callStatsCounts[data.CallsMetricTimeouts],
		callStatsCounts[data.CallsMetricTotals])
	cancelCallRates := services.CancelCallRates(callStatsCounts[data.CallsMetricCancels],
		callStatsCounts[data.CallsMetricTotals])

	callStatsCounts[data.CallsMetricCompletesRate] = completesCallRates
	callStatsCounts[data.CallsMetricToHumansRate] = trasferToHumanCallRates
	callStatsCounts[data.CallsMetricTimeoutsRate] = timeoutCallRates
	callStatsCounts[data.CallsMetricCancelsRate] = cancelCallRates

	return callStatsCounts, nil
}

func createCallStatsResponse(query data.CallStatsQuery,
	callsCounts map[string]map[string]interface{}) (*data.CallStatsResponse, error) {
	callStatsQ, totalCallStatsQ, err := createCallStatsQ(callsCounts)
	if err != nil {
		return nil, err
	}

	callStatsQuantities := make(data.CallStatsQuantities, 0)

	for datetime, q := range callStatsQ {
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

		callStatsQuantity := data.CallStatsQuantity{
			CallStatsQ: *q,
			TimeText:   timeText,
			Time:       strconv.FormatInt(timeUnixSec, 10),
		}

		callStatsQuantities = append(callStatsQuantities, callStatsQuantity)
	}

	// Sort callStatsQuantities
	sort.Sort(callStatsQuantities)

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

	response := data.CallStatsResponse{
		TableHeader: data.CallStatsTableHeader,
		Data: data.CallStatData{
			Quantities: callStatsQuantities,
			Type:       query.AggInterval,
			Name:       "场景数",
		},
		Total: data.CallStatsTotal{
			CallStatsQ: *totalCallStatsQ,
			TimeText:   "合计",
			Time:       totalTime,
		},
	}

	return &response, nil
}

func createTopToHumanAnswersResponse(answers data.ToHumanAnswers) *data.TopToHumanAnswersResponse {
	answersData := make([]data.TopToHumanAnswer, 0)
	rank := 1

	for _, answer := range answers {
		d := data.TopToHumanAnswer{
			ToHumanAnswer: data.ToHumanAnswer{
				Answer: answer.Answer,
				Count:  answer.Count,
			},
			Rank: rank,
		}

		answersData = append(answersData, d)
		rank++
	}

	response := data.TopToHumanAnswersResponse{
		Answers: answersData,
	}

	return &response
}

// createCallStatsQ converts callsCounts = {
// 	   totals: {
// 	       2018-01-01 00:00:00: 2,
// 	       2018-01-01 01:00:00: 4,
// 	       2018-01-01 02:00:00: 9,
// 	       ...
// 	   },
// 	   completes: {
// 		   2018-01-01 00:00:00: 1,
// 		   2018-01-01 01:00:00: 13,
// 		   2018-01-01 02:00:00: 5,
// 		   ...
// 	   },
// 	   ...
// }
//
// to the form: {
// 	   2018-01-01 00:00:00: {
// 		   totals: 2,
// 		   completes: 1,
// 		   ...
// 	   },
// 	   2018-01-01 01:00:00: {
// 		   totals: 4,
// 		   completes: 13,
// 		   ...
// 	   },
// 	   2018-01-01 02:00:00: {
// 		   totals: 9,
// 		   completes: 5,
// 		   ...
// 	   },
// 	   ...
// }
func createCallStatsQ(callsCounts map[string]map[string]interface{}) (callStatsQ map[string]*data.CallStatsQ,
	totalCallStatsQ *data.CallStatsQ, err error) {
	callStatsQ = make(map[string]*data.CallStatsQ)
	totalCallStatsQ = &data.CallStatsQ{}

	for callMetric, counts := range callsCounts {
		for key, count := range counts {
			if _, ok := callStatsQ[key]; !ok {
				callStatsQ[key] = &data.CallStatsQ{}
			}

			switch callMetric {
			case data.CallsMetricTotals:
				callStatsQ[key].Totals = count.(int64)
				totalCallStatsQ.Totals += count.(int64)
			case data.CallsMetricCompletes:
				callStatsQ[key].Completes = count.(int64)
				totalCallStatsQ.Completes += count.(int64)
			case data.CallsMetricCompletesRate:
				callStatsQ[key].CompletesRate = count.(string)
			case data.CallsMetricToHumans:
				callStatsQ[key].ToHumans = count.(int64)
				totalCallStatsQ.ToHumans += count.(int64)
			case data.CallsMetricToHumansRate:
				callStatsQ[key].ToHumansRate = count.(string)
			case data.CallsMetricTimeouts:
				callStatsQ[key].Timeouts = count.(int64)
				totalCallStatsQ.Timeouts += count.(int64)
			case data.CallsMetricTimeoutsRate:
				callStatsQ[key].TimeoutsRate = count.(string)
			case data.CallsMetricCancels:
				callStatsQ[key].Cancels = count.(int64)
				totalCallStatsQ.Cancels += count.(int64)
			case data.CallsMetricCancelsRate:
				callStatsQ[key].CancelsRate = count.(string)
			case data.CallsMetricUnknowns:
				callStatsQ[key].Unknowns = count.(int64)
				totalCallStatsQ.Unknowns += count.(int64)
			}
		}
	}

	if totalCallStatsQ.Totals != 0 {
		totalCallStatsQ.CompletesRate = strconv.
			FormatFloat(float64(totalCallStatsQ.Completes)/float64(totalCallStatsQ.Totals), 'f', 2, 64)

		totalCallStatsQ.ToHumansRate = strconv.
			FormatFloat(float64(totalCallStatsQ.ToHumans)/float64(totalCallStatsQ.Totals), 'f', 2, 64)

		totalCallStatsQ.TimeoutsRate = strconv.
			FormatFloat(float64(totalCallStatsQ.Timeouts)/float64(totalCallStatsQ.Totals), 'f', 2, 64)

		totalCallStatsQ.CancelsRate = strconv.
			FormatFloat(float64(totalCallStatsQ.Cancels)/float64(totalCallStatsQ.Totals), 'f', 2, 64)
	} else {
		totalCallStatsQ.CompletesRate = "N/A"
		totalCallStatsQ.ToHumansRate = "N/A"
		totalCallStatsQ.TimeoutsRate = "N/A"
		totalCallStatsQ.CancelsRate = "N/A"
	}

	return callStatsQ, totalCallStatsQ, nil
}
