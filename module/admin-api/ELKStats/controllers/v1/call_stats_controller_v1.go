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
	servicesV1 "emotibot.com/emotigo/module/admin-api/ELKStats/services/v1"
	"emotibot.com/emotigo/module/admin-api/util/elasticsearch"
	"emotibot.com/emotigo/module/admin-api/util/requestheader"
)

var callStatsQueryHandlers = map[string]dataV1.CallStatsQueryHandler{
	dataCommon.CallsMetricTotals:    servicesV1.TotalCallCounts,
	dataCommon.CallsMetricCompletes: servicesV1.CompleteCallCounts,
	dataCommon.CallsMetricToHumans:  servicesV1.ToHumanCallCounts,
	dataCommon.CallsMetricTimeouts:  servicesV1.TimeoutCallCounts,
	dataCommon.CallsMetricCancels:   servicesV1.CancelCallCounts,
	dataCommon.CallsMetricUnknowns:  servicesV1.Unknowns,
	// Note: 'Completes rate', 'To Humans rate', 'Timeouts rate' and 'Cancels rate' are called separately
}

func CallStatsGetHandler(w http.ResponseWriter, r *http.Request) {
	appID := requestheader.GetAppID(r)
	statsType := r.URL.Query().Get("type")
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

	query := dataV1.CallStatsQuery{
		CommonQuery: data.CommonQuery{
			AppID:     appID,
			StartTime: startTime,
			EndTime:   endTime,
		},
		AggInterval: aggInterval,
	}

	switch statsType {
	case dataCommon.CallStatsTypeTime:
		callsCounts, err := fetchCallStats(query)
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

		response, err := createCallStatsResponse(query, callsCounts)
		if err != nil {
			errResponse := data.NewErrorResponse(err.Error())
			controllers.ReturnInternalServerError(w, errResponse)
			return
		}

		controllers.ReturnOK(w, response)
	case dataCommon.CallStatsTypeAnswers:
		answers, err := servicesV1.TopToHumanAnswers(query, 20)
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

		response := createTopToHumanAnswersResponse(answers)
		controllers.ReturnOK(w, response)
	default:
		errResponse := data.NewBadRequestResponse(data.ErrCodeInvalidParameterType, "type")
		controllers.ReturnBadRequest(w, errResponse)
	}
}

func fetchCallStats(query dataV1.CallStatsQuery) (map[string]map[string]interface{}, error) {
	var callStatsCountsSync sync.Map // Use sync.Map to avoid concurrent map writes
	callStatsCounts := make(map[string]map[string]interface{})
	done := make(chan error, len(callStatsQueryHandlers))
	var queryError error

	// Fetch statistics concurrently
	for queryKey, queryHandler := range callStatsQueryHandlers {
		go func(key string, handler dataV1.CallStatsQueryHandler) {
			counts, err := handler(query)
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

	completesCallRates := servicesV1.CompleteCallRates(callStatsCounts[dataCommon.CallsMetricCompletes],
		callStatsCounts[dataCommon.CallsMetricTotals])
	trasferToHumanCallRates := servicesV1.ToHumanCallRates(callStatsCounts[dataCommon.CallsMetricToHumans],
		callStatsCounts[dataCommon.CallsMetricTotals])
	timeoutCallRates := servicesV1.TimeoutCallRates(callStatsCounts[dataCommon.CallsMetricTimeouts],
		callStatsCounts[dataCommon.CallsMetricTotals])
	cancelCallRates := servicesV1.CancelCallRates(callStatsCounts[dataCommon.CallsMetricCancels],
		callStatsCounts[dataCommon.CallsMetricTotals])

	callStatsCounts[dataCommon.CallsMetricCompletesRate] = completesCallRates
	callStatsCounts[dataCommon.CallsMetricToHumansRate] = trasferToHumanCallRates
	callStatsCounts[dataCommon.CallsMetricTimeoutsRate] = timeoutCallRates
	callStatsCounts[dataCommon.CallsMetricCancelsRate] = cancelCallRates

	return callStatsCounts, nil
}

func createCallStatsResponse(query dataV1.CallStatsQuery,
	callsCounts map[string]map[string]interface{}) (*dataV1.CallStatsResponse, error) {
	callStatsQ, totalCallStatsQ, err := createCallStatsQ(callsCounts)
	if err != nil {
		return nil, err
	}

	callStatsQuantities := make(dataV1.CallStatsQuantities, 0)

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

		callStatsQuantity := dataV1.CallStatsQuantity{
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

	response := dataV1.CallStatsResponse{
		TableHeader: dataV1.CallStatsTableHeader,
		Data: dataV1.CallStatData{
			Quantities: callStatsQuantities,
			Type:       query.AggInterval,
			Name:       "场景数",
		},
		Total: dataV1.CallStatsTotal{
			CallStatsQ: *totalCallStatsQ,
			TimeText:   "合计",
			Time:       totalTime,
		},
	}

	return &response, nil
}

func createTopToHumanAnswersResponse(
	answers dataV1.ToHumanAnswers) *dataV1.TopToHumanAnswersResponse {
	answersData := make([]dataV1.TopToHumanAnswer, 0)
	rank := 1

	for _, answer := range answers {
		d := dataV1.TopToHumanAnswer{
			ToHumanAnswer: dataV1.ToHumanAnswer{
				Answer: answer.Answer,
				Count:  answer.Count,
			},
			Rank: rank,
		}

		answersData = append(answersData, d)
		rank++
	}

	response := dataV1.TopToHumanAnswersResponse{
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
func createCallStatsQ(
	callsCounts map[string]map[string]interface{}) (callStatsQ map[string]*dataV1.CallStatsQ,
	totalCallStatsQ *dataV1.CallStatsQ, err error) {
	callStatsQ = make(map[string]*dataV1.CallStatsQ)
	totalCallStatsQ = &dataV1.CallStatsQ{}

	for callMetric, counts := range callsCounts {
		for key, count := range counts {
			if _, ok := callStatsQ[key]; !ok {
				callStatsQ[key] = &dataV1.CallStatsQ{}
			}

			switch callMetric {
			case dataCommon.CallsMetricTotals:
				callStatsQ[key].Totals = count.(int64)
				totalCallStatsQ.Totals += count.(int64)
			case dataCommon.CallsMetricCompletes:
				callStatsQ[key].Completes = count.(int64)
				totalCallStatsQ.Completes += count.(int64)
			case dataCommon.CallsMetricCompletesRate:
				callStatsQ[key].CompletesRate = count.(string)
			case dataCommon.CallsMetricToHumans:
				callStatsQ[key].ToHumans = count.(int64)
				totalCallStatsQ.ToHumans += count.(int64)
			case dataCommon.CallsMetricToHumansRate:
				callStatsQ[key].ToHumansRate = count.(string)
			case dataCommon.CallsMetricTimeouts:
				callStatsQ[key].Timeouts = count.(int64)
				totalCallStatsQ.Timeouts += count.(int64)
			case dataCommon.CallsMetricTimeoutsRate:
				callStatsQ[key].TimeoutsRate = count.(string)
			case dataCommon.CallsMetricCancels:
				callStatsQ[key].Cancels = count.(int64)
				totalCallStatsQ.Cancels += count.(int64)
			case dataCommon.CallsMetricCancelsRate:
				callStatsQ[key].CancelsRate = count.(string)
			case dataCommon.CallsMetricUnknowns:
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
