package v1

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"emotibot.com/emotigo/pkg/api/dal/v1"
	"github.com/gorilla/mux"

	"emotibot.com/emotigo/module/admin-api/ELKStats/controllers"
	"emotibot.com/emotigo/module/admin-api/ELKStats/controllers/common"
	"emotibot.com/emotigo/module/admin-api/ELKStats/data"
	dataCommon "emotibot.com/emotigo/module/admin-api/ELKStats/data/common"
	dataV1 "emotibot.com/emotigo/module/admin-api/ELKStats/data/v1"
	servicesCommon "emotibot.com/emotigo/module/admin-api/ELKStats/services/common"
	servicesV1 "emotibot.com/emotigo/module/admin-api/ELKStats/services/v1"
	"emotibot.com/emotigo/module/admin-api/util/elasticsearch"
	"emotibot.com/emotigo/module/admin-api/util/requestheader"
	"emotibot.com/emotigo/pkg/logger"
	dac "emotibot.com/emotigo/pkg/api/dac/v1"
)

//VisitRecordsGetHandler handle advanced query for records.
//Limit & Page should be given by r's query string.
func VisitRecordsGetHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	query, err := newRecordQuery(r)
	if err != nil {
		errResponse := data.NewErrorResponseWithCode(data.ErrCodeInvalidRequestBody, err.Error())
		controllers.ReturnBadRequest(w, errResponse)
		return
	}

	if query.StartTime != nil && query.EndTime != nil {
		if *query.StartTime > *query.EndTime {
			errResponse := data.NewErrorResponseWithCode(data.ErrCodeInvalidParameterStartTime,
				"start_time cannot greater than end_time")
			controllers.ReturnBadRequest(w, errResponse)
			return
		}
	}

	limit, found := r.URL.Query()["limit"]
	if !found {
		errResponse := data.NewErrorResponseWithCode(data.ErrCodeInvalidRequestBody, "limit should be greater than zero")
		controllers.ReturnBadRequest(w, errResponse)
		return
	} else {
		query.Limit, _ = strconv.Atoi(limit[0])
	}

	page := r.URL.Query().Get("page")
	p, _ := strconv.ParseInt(page, 10, 64)
	if p <= 0 {
		p = 1
	}

	query.From = (p - 1) * int64(query.Limit)

	if query.From+int64(query.Limit) > data.ESMaxResultWindow {
		errResponse := data.NewBadRequestResponse(data.ErrCodeInvalidParameterPage, "page")
		controllers.ReturnBadRequest(w, errResponse)
		return
	}

	result, err := servicesV1.VisitRecordsQuery(*query,
		servicesCommon.AggregateFilterIgnoredRecord, servicesCommon.AggregateFilterMarkedRecord)
	if err != nil {
		var errResponse data.ErrorResponse
		if rootCauseErrors, ok := elasticsearch.ExtractElasticsearchRootCauseErrors(err); ok {
			errResponse = data.NewErrorResponse(strings.Join(rootCauseErrors, ", "))
		} else {
			errResponse = data.NewErrorResponse(err.Error())
		}
		controllers.ReturnInternalServerError(w, errResponse)
		return
	}

	response := dataV1.VisitRecordsResponse{
		Data:        result.Hits,
		Limit:       query.Limit,
		Page:        int(query.From)/query.Limit + 1,
		TableHeader: dataV1.VisitRecordsTableHeader,
		TotalSize:   result.Aggs["total_size"].(int64),
		MarkedSize:  result.Aggs["isMarked"].(int64),
		IgnoredSize: result.Aggs["isIgnored"].(int64),
	}

	controllers.ReturnOK(w, response)
}

func VisitRecordsExportHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	query, err := newRecordQuery(r)
	if err != nil {
		errResponse := data.NewErrorResponseWithCode(data.ErrCodeInvalidRequestBody,
			data.ErrInvalidRequestBody.Error())
		controllers.ReturnBadRequest(w, errResponse)
		return
	}

	if query.StartTime != nil && query.EndTime != nil {
		if *query.StartTime > *query.EndTime {
			errResponse := data.NewErrorResponseWithCode(data.ErrCodeInvalidParameterStartTime,
				"start_time cannot greater than end_time")
			controllers.ReturnBadRequest(w, errResponse)
			return
		}
	}

	locale := requestheader.GetLocale(r)
	exportTaskID, err := servicesV1.VisitRecordsExport(query, locale)
	if err != nil {
		var errResponse data.ErrorResponse
		if rootCauseErrors, ok := elasticsearch.ExtractElasticsearchRootCauseErrors(err); ok {
			errResponse = data.NewErrorResponse(strings.Join(rootCauseErrors, "; "))
		} else {
			errResponse = data.NewErrorResponse(err.Error())

			switch err {
			case data.ErrExportTaskInProcess:
				controllers.ReturnForbiddenRequest(w, errResponse)
			default:
				controllers.ReturnInternalServerError(w, errResponse)
			}
		}
		return
	}

	response := dataCommon.VisitRecordsExportResponse{
		ExportID: exportTaskID,
	}

	controllers.ReturnOK(w, response)
}

func VisitRecordsExportDownloadHandler(w http.ResponseWriter, r *http.Request) {
	common.VisitRecordsExportDownloadHandler(w, r)
}

func VisitRecordsExportDeleteHandler(w http.ResponseWriter, r *http.Request) {
	common.VisitRecordsExportDeleteHandler(w, r)
}

func VisitRecordsExportStatusHandler(w http.ResponseWriter, r *http.Request) {
	common.VisitRecordsExportStatusHandler(w, r)
}

// RecordsIgnoredUpdateHandler handle the request for updating record's ignore column
func RecordsIgnoredUpdateHandler(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Ignore  *bool
		Records []string `json:"records"`
	}
	defer r.Body.Close()

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		controllers.ReturnBadRequest(w, data.NewErrorResponseWithCode(http.StatusBadRequest, err.Error()))
		return
	}

	if request.Ignore == nil {
		controllers.ReturnBadRequest(w, data.NewErrorResponseWithCode(http.StatusBadRequest, "ignore is required"))
		return
	}

	if request.Records == nil {
		controllers.ReturnBadRequest(w, data.NewErrorResponseWithCode(http.StatusBadRequest, "records is required"))
		return
	}

	q := dataV1.RecordQuery{
		AppID:   requestheader.GetAppID(r),
		Records: make([]interface{}, len(request.Records)),
	}
	for i, r := range request.Records {
		q.Records[i] = r
	}
	err = servicesV1.UpdateRecords(q, servicesV1.UpdateRecordIgnore(*request.Ignore))
	if err != nil {
		if rootCauseErrors, ok := elasticsearch.ExtractElasticsearchRootCauseErrors(err); ok {
			errResponse := data.NewErrorResponse(strings.Join(rootCauseErrors, "; "))
			controllers.ReturnInternalServerError(w, errResponse)
		} else {
			controllers.ReturnBadRequest(w, data.NewErrorResponseWithCode(http.StatusInternalServerError,
				"update records failed, "+err.Error()))
		}

		return
	} else {
		controllers.ReturnOK(w, data.NewSuccessStatusResponse(http.StatusOK, "update success"))
		return
	}

}

// NewRecordsMarkUpdateHandler create a handler to handle records mark & unmark request.
// The handler will update record store & ssm store.
// Because we separate the Init() function and Controller, the only way to pass dal.Client is from parameters.
func NewRecordsMarkUpdateHandler(client *dal.Client) func(w http.ResponseWriter, r *http.Request) {
	if client == nil {
		return func(w http.ResponseWriter, r *http.Request) {
			controllers.ReturnInternalServerError(w, data.NewErrorResponse("no valid dal client"))
		}
	}

	type internalError struct {
		IsRollbacked bool `json:"rollbacked"`
	}
	type response struct {
		Done []interface{} `json:"done"`
		Skip []string      `json:"skip"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Content string   `json:"content"`
			Mark    *bool    `json:"mark"`
			Records []string `json:"records"`
		}
		defer r.Body.Close()
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			controllers.ReturnBadRequest(w, data.NewErrorResponseWithCode(http.StatusBadRequest, err.Error()))
			return
		}
		if request.Mark == nil {
			controllers.ReturnBadRequest(w, data.NewErrorResponseWithCode(http.StatusBadRequest, "mark is required."))
			return
		}
		if request.Records == nil {
			controllers.ReturnBadRequest(w, data.NewErrorResponseWithCode(http.StatusBadRequest, "records is required"))
			return
		}
		appID := requestheader.GetAppID(r)
		q := dataV1.RecordQuery{
			AppID:   appID,
			Records: make([]interface{}, len(request.Records)),
			Limit:   len(request.Records),
		}
		for i, r := range request.Records {
			q.Records[i] = r
		}
		// Need retrive record's userq
		result, err := servicesV1.VisitRecordsQuery(q)
		if err != nil {
			logger.Error.Printf("check record content failed, %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(internalError{
				IsRollbacked: true,
			})
			return
		}
		todoRecords := make([]*dataV1.VisitRecordsData, 0)
		skipRecordsID := make([]string, 0)
		// Filter skip records
		for _, record := range result.Hits {
			var isSQ bool
			isSQ, err = client.IsStandardQuestion(appID, record.UserQ)
			if err != nil {
				logger.Error.Printf("Get isStd failed, %v\n", err)
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(internalError{
					IsRollbacked: true,
				})
				return
			}

			if isSQ && record.UserQ != request.Content {
				skipRecordsID = append(skipRecordsID, record.UniqueID)
				continue
			}
			todoRecords = append(todoRecords, record)
		}

		logger.Trace.Println("Skip records:", skipRecordsID)
		logger.Trace.Println("To-do records:", todoRecords)

		// Deduplicate input userQ, because dalClient can not handle duplicate request.
		questions := map[string]struct{}{}
		for _, r := range todoRecords {
			_, found := questions[r.UserQ]
			if found {
				continue
			}
			questions[r.UserQ] = struct{}{}
		}
		uniqueUserQ := []string{}
		for r := range questions {
			uniqueUserQ = append(uniqueUserQ, r)
		}

		if len(uniqueUserQ) == 0 {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response{
				Done: []interface{}{},
				Skip: skipRecordsID,
			})
			return
		}

		if *request.Mark {
			// If delete Simq have problem, we will know at next step setSimQ, so dont bother to check delete simQ operation at all.
			client.DeleteSimilarQuestions(appID, uniqueUserQ...)
			err = client.SetSimilarQuestion(appID, request.Content, uniqueUserQ...)
			if err != nil {
				logger.Error.Printf("set simQ failed, %v\n", err)
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(internalError{
					IsRollbacked: false,
				})
				return
			}
		} else {
			// Unmark should remove the records from ssm store
			err = client.DeleteSimilarQuestions(appID, uniqueUserQ...)
			//Note: because ssm wont sync with record, so it can be someone delete the ssm
			//We only need to return error if the problem is not "not exist"
			if dErr, ok := err.(*dal.DetailError); ok {
				for _, op := range dErr.Results {
					if op != "NOT_EXIST" {
						logger.Error.Printf("delete sim q failed, %v", err)
						w.WriteHeader(http.StatusInternalServerError)
						json.NewEncoder(w).Encode(internalError{
							IsRollbacked: false,
						})
						return
					}
				}
			} else if err != nil {
				logger.Error.Printf("delete sim q failed, %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(internalError{
					IsRollbacked: false,
				})
				return
			}
		}

		newQuery := dataV1.RecordQuery{
			AppID:   appID,
			Records: make([]interface{}, len(todoRecords)),
		}
		for i, r := range todoRecords {
			newQuery.Records[i] = r.UniqueID
		}

		err = servicesV1.UpdateRecords(newQuery, servicesV1.UpdateRecordMark(*request.Mark))
		if err != nil {
			logger.Error.Printf("service update record failed, %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(internalError{
				IsRollbacked: true,
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response{
			Done: newQuery.Records,
			Skip: skipRecordsID,
		})
	}

}

// NewRecordsMarkUpdateHandler create a handler to handle records mark & unmark request.
// The handler will update record store & ssm store.
// Because we separate the Init() function and Controller, the only way to pass dac.Client is from parameters.
func NewRecordsMarkUpdateHandlerV2(client *dac.Client) func(w http.ResponseWriter, r *http.Request) {
	if client == nil {
		return func(w http.ResponseWriter, r *http.Request) {
			controllers.ReturnInternalServerError(w, data.NewErrorResponse("no valid dac client"))
		}
	}

	type internalError struct {
		IsRollbacked bool `json:"rollbacked"`
	}
	type response struct {
		Done []interface{} `json:"done"`
		Skip []string      `json:"skip"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Content string   `json:"content"`
			Mark    *bool    `json:"mark"`
			Records []string `json:"records"`
		}
		defer r.Body.Close()
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			controllers.ReturnBadRequest(w, data.NewErrorResponseWithCode(http.StatusBadRequest, err.Error()))
			return
		}
		if request.Mark == nil {
			controllers.ReturnBadRequest(w, data.NewErrorResponseWithCode(http.StatusBadRequest, "mark is required."))
			return
		}
		if request.Records == nil {
			controllers.ReturnBadRequest(w, data.NewErrorResponseWithCode(http.StatusBadRequest, "records is required"))
			return
		}
		appID := requestheader.GetAppID(r)
		userId := requestheader.GetUserID(r)
		q := dataV1.RecordQuery{
			AppID:   appID,
			Records: make([]interface{}, len(request.Records)),
			Limit:   len(request.Records),
		}
		for i, r := range request.Records {
			q.Records[i] = r
		}
		// Need retrive record's userq
		result, err := servicesV1.VisitRecordsQuery(q)
		if err != nil {
			logger.Error.Printf("check record content failed, %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(internalError{
				IsRollbacked: true,
			})
			return
		}
		todoRecords := make([]*dataV1.VisitRecordsData, 0)
		skipRecordsID := make([]string, 0)
		// Filter skip records
		for _, record := range result.Hits {
			var isSQ bool
			isSQ, err = client.IsStandardQuestion(appID, record.UserQ)
			if err != nil {
				logger.Error.Printf("Get isStd failed, %v\n", err)
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(internalError{
					IsRollbacked: true,
				})
				return
			}

			if isSQ && record.UserQ != request.Content {
				skipRecordsID = append(skipRecordsID, record.UniqueID)
				continue
			}
			todoRecords = append(todoRecords, record)
		}

		logger.Trace.Println("Skip records:", skipRecordsID)
		logger.Trace.Println("To-do records:", todoRecords)

		// Deduplicate input userQ, because dacClient can not handle duplicate request.
		questions := map[string]struct{}{}
		for _, r := range todoRecords {
			_, found := questions[r.UserQ]
			if found {
				continue
			}
			questions[r.UserQ] = struct{}{}
		}
		uniqueUserQ := []string{}
		for r := range questions {
			uniqueUserQ = append(uniqueUserQ, r)
		}

		if len(uniqueUserQ) == 0 {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response{
				Done: []interface{}{},
				Skip: skipRecordsID,
			})
			return
		}

		if *request.Mark {
			// If delete Simq have problem, we will know at next step setSimQ, so dont bother to check delete simQ operation at all.
			client.DeleteSimilarQuestionsWithUser(appID, userId, uniqueUserQ...)
			err = client.SetSimilarQuestionWithUser(appID, userId, request.Content, uniqueUserQ...)
			if err != nil {
				logger.Error.Printf("set simQ failed, %v\n", err)
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(internalError{
					IsRollbacked: false,
				})
				return
			}
		} else {
			// Unmark should remove the records from ssm store
			err = client.DeleteSimilarQuestionsWithUser(appID, userId, uniqueUserQ...)
			//Note: because ssm wont sync with record, so it can be someone delete the ssm
			//We only need to return error if the problem is not "not exist"
			if dErr, ok := err.(*dac.DetailError); ok {
				for _, op := range dErr.Results {
					if op != "NOT_EXIST" {
						logger.Error.Printf("delete sim q failed, %v", err)
						w.WriteHeader(http.StatusInternalServerError)
						json.NewEncoder(w).Encode(internalError{
							IsRollbacked: false,
						})
						return
					}
				}
			} else if err != nil {
				logger.Error.Printf("delete sim q failed, %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(internalError{
					IsRollbacked: false,
				})
				return
			}
		}

		newQuery := dataV1.RecordQuery{
			AppID:   appID,
			Records: make([]interface{}, len(todoRecords)),
		}
		for i, r := range todoRecords {
			newQuery.Records[i] = r.UniqueID
		}

		err = servicesV1.UpdateRecords(newQuery, servicesV1.UpdateRecordMark(*request.Mark))
		if err != nil {
			logger.Error.Printf("service update record failed, %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(internalError{
				IsRollbacked: true,
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response{
			Done: newQuery.Records,
			Skip: skipRecordsID,
		})
	}

}

// NewRecordSSMHandler create a handler for retriving ssm info of certain record.
func NewRecordSSMHandler(client *dal.Client) func(http.ResponseWriter, *http.Request) {
	if client == nil {
		return func(w http.ResponseWriter, r *http.Request) {
			controllers.ReturnInternalServerError(w, data.NewErrorResponse("no valid dal client"))
		}
	}
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := mux.Vars(r)["id"]
		if !ok {
			controllers.ReturnBadRequest(w, data.NewErrorResponseWithCode(http.StatusBadRequest, "no path variable id"))
			return
		}
		appID := requestheader.GetAppID(r)
		query := dataV1.RecordQuery{
			AppID:   appID,
			Records: []interface{}{id},
			Limit:   1,
		}
		result, err := servicesV1.VisitRecordsQuery(query)
		if err != nil {
			logger.Error.Printf("fetch es records failed, %v\n", err)
			controllers.ReturnInternalServerError(w, data.NewErrorResponse("internal server error"))
			return
		}
		if size := len(result.Hits); size == 0 || size > 1 {
			controllers.ReturnBadRequest(w, data.NewErrorResponseWithCode(http.StatusBadRequest, "record id "+id+" is ambiguous(results: "+strconv.Itoa(size)+")"))
			return
		}
		logger.Trace.Printf("Get record: %+v\n", *result.Hits[0])
		if !result.Hits[0].IsMarked {
			controllers.ReturnBadRequest(w, data.NewErrorResponseWithCode(http.StatusBadRequest, "record is not marked"))
			return
		}
		lq := result.Hits[0].UserQ
		var response = struct {
			Content string `json:"marked_content"`
		}{}
		isLQ, err := client.IsSimilarQuestion(appID, lq)
		if err != nil {
			logger.Error.Printf("get isSimilarQ failed, %v", err)
			controllers.ReturnInternalServerError(w, data.NewErrorResponse("internal server error"))
			return
		}
		if !isLQ {
			response.Content = ""
			controllers.ReturnOK(w, response)
			return
		}

		response.Content, err = client.Question(appID, lq)
		if err != nil {
			logger.Error.Printf("get question for lq %s failed, %v", lq, err)
			controllers.ReturnInternalServerError(w, data.NewErrorResponse("internal server error"))
			return
		}
		controllers.ReturnOK(w, response)
	}
}


// NewRecordSSMHandler create a handler for retriving ssm info of certain record.
func NewRecordSSMHandlerV2(client *dac.Client) func(http.ResponseWriter, *http.Request) {
	if client == nil {
		return func(w http.ResponseWriter, r *http.Request) {
			controllers.ReturnInternalServerError(w, data.NewErrorResponse("no valid dac client"))
		}
	}
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := mux.Vars(r)["id"]
		if !ok {
			controllers.ReturnBadRequest(w, data.NewErrorResponseWithCode(http.StatusBadRequest, "no path variable id"))
			return
		}
		appID := requestheader.GetAppID(r)
		query := dataV1.RecordQuery{
			AppID:   appID,
			Records: []interface{}{id},
			Limit:   1,
		}
		result, err := servicesV1.VisitRecordsQuery(query)
		if err != nil {
			logger.Error.Printf("fetch es records failed, %v\n", err)
			controllers.ReturnInternalServerError(w, data.NewErrorResponse("internal server error"))
			return
		}
		if size := len(result.Hits); size == 0 || size > 1 {
			controllers.ReturnBadRequest(w, data.NewErrorResponseWithCode(http.StatusBadRequest, "record id "+id+" is ambiguous(results: "+strconv.Itoa(size)+")"))
			return
		}
		logger.Trace.Printf("Get record: %+v\n", *result.Hits[0])
		if !result.Hits[0].IsMarked {
			controllers.ReturnBadRequest(w, data.NewErrorResponseWithCode(http.StatusBadRequest, "record is not marked"))
			return
		}
		lq := result.Hits[0].UserQ
		var response = struct {
			Content string `json:"marked_content"`
		}{}
		isLQ, err := client.IsSimilarQuestion(appID, lq)
		if err != nil {
			logger.Error.Printf("get isSimilarQ failed, %v", err)
			controllers.ReturnInternalServerError(w, data.NewErrorResponse("internal server error"))
			return
		}
		if !isLQ {
			response.Content = ""
			controllers.ReturnOK(w, response)
			return
		}

		response.Content, err = client.Question(appID, lq)
		if err != nil {
			logger.Error.Printf("get question for lq %s failed, %v", lq, err)
			controllers.ReturnInternalServerError(w, data.NewErrorResponse("internal server error"))
			return
		}
		controllers.ReturnOK(w, response)
	}
}



//newRecordQuery create a new *dataCommon.RecordQuery with given r.
//It should handle all the logic to retrive data,
//the only error should be returned is if request itself is invalided.
//It handled the limit & page logic for now, should move up to controller level.
func newRecordQuery(r *http.Request) (*dataV1.RecordQuery, error) {
	enterpriseID := requestheader.GetEnterpriseID(r)
	appID := requestheader.GetAppID(r)
	var query dataV1.RecordQuery
	err := json.NewDecoder(r.Body).Decode(&query)
	if err != nil {
		return nil, fmt.Errorf("record query: request handled failed, %v", err)
	}

	query.EnterpriseID = enterpriseID
	query.AppID = appID

	return &query, nil
}
