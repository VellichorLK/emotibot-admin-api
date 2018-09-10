package controllers

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"emotibot.com/emotigo/pkg/api/dal/v1"
	"github.com/gorilla/mux"

	"emotibot.com/emotigo/module/admin-api/ELKStats/data"
	"emotibot.com/emotigo/module/admin-api/ELKStats/services"
	"emotibot.com/emotigo/module/admin-api/util/requestheader"
	"emotibot.com/emotigo/pkg/logger"
)

//VisitRecordsGetHandler handle advanced query for records.
//Limit & Page should be given by r's query string.
func VisitRecordsGetHandler(w http.ResponseWriter, r *http.Request) {
	// DISCARD THIS SINCE appID and enterpriseID should be handled by middleware
	// if enterpriseID == "" && appID == "" {
	// 	errResp := data.ErrorResponse{
	// 		Message: fmt.Sprintf("Both headers %s and %s are not specified",
	// 			data.EnterpriseIDHeaderKey, data.AppIDHeaderKey),
	// 	}
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	writeResponseJSON(w, errResp)
	// 	return
	// }
	defer r.Body.Close()
	query, err := newRecordQuery(r)
	if err != nil {
		errResponse := data.NewErrorResponseWithCode(data.ErrCodeInvalidRequestBody, err.Error())
		returnBadRequest(w, errResponse)
		return
	}
	limit, found := r.URL.Query()["limit"]
	if !found {
		errResponse := data.NewErrorResponseWithCode(data.ErrCodeInvalidRequestBody, "limit should be greater than zero")
		returnBadRequest(w, errResponse)
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

	// TODO: merge esclient error into current design
	// esCtx, esClient := elasticsearch.GetClient()
	// if esClient == nil {
	// 	returnInternalServerError(w, data.NewErrorResponse(data.ErrNotInit.Error()))
	// 	return
	// }

	result, err := services.VisitRecordsQuery(*query, services.AggregateFilterIgnoredRecord, services.AggregateFilterMarkedRecord)
	if err != nil {
		errResponse := data.NewErrorResponse(err.Error())
		returnInternalServerError(w, errResponse)
		return
	}

	response := data.VisitRecordsResponse{
		Data:        result.Hits,
		Limit:       query.Limit,
		Page:        int(query.From)/query.Limit + 1,
		TableHeader: data.VisitRecordsTableHeader,
		TotalSize:   result.Aggs["total_size"].(int64),
		MarkedSize:  result.Aggs["isMarked"].(int64),
		IgnoredSize: result.Aggs["isIgnored"].(int64),
	}
	w.Header().Set("content-type", "application/json")
	returnOK(w, response)
}

//RecordsDownloadHandler handle record downloading request.
//It ignore limit and page, and write csv format records to the response body
func RecordsDownloadHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	query, err := newRecordQuery(r)
	if err != nil {
		errResponse := data.NewErrorResponseWithCode(data.ErrCodeInvalidRequestBody,
			data.ErrInvalidRequestBody.Error())
		returnBadRequest(w, errResponse)
		return
	}

	query.Limit = 10000
	query.From = 0
	//FIXME: use scroll API to rewrite
	result, err := services.VisitRecordsQuery(*query)
	if err != nil {
		errResponse := data.NewErrorResponse(err.Error())
		returnInternalServerError(w, errResponse)
		return
	}

	totalRecords := result.Hits
	//This total size may over the es limit, 10000
	logger.Trace.Printf("download total size: %d", result.Aggs["total_size"].(int64))

	// Everything goes right, start writting CSV response
	// FIXME, ADD windows byte mark for utf-8 support on old EXCEL
	out := csv.NewWriter(w)
	out.Write(data.VisitRecordsExportHeader)

	for _, record := range totalRecords {
		csvData := []string{
			record.UserID,
			record.UserQ,
			strconv.FormatFloat(record.Score, 'f', -1, 64),
			record.StdQ,
			record.Answer,
			record.LogTime,
			record.Emotion,
			record.QType,
		}

		err = out.Write(csvData)
		if err != nil {
			errResponse := data.NewErrorResponse(err.Error())
			returnInternalServerError(w, errResponse)
			return
		}
	}

	out.Flush()
	err = out.Error()
	if err != nil {
		errResponse := data.NewErrorResponse(err.Error())
		returnInternalServerError(w, errResponse)
		return
	}
	w.Header().Set("content-type", "text/csv;charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

type markResponse struct {
	Records []string `json:"records"`
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
		returnBadRequest(w, data.NewErrorResponseWithCode(http.StatusBadRequest, err.Error()))
		return
	}
	if request.Ignore == nil {
		returnBadRequest(w, data.NewErrorResponseWithCode(http.StatusBadRequest, "ignore is required"))
		return
	}
	if request.Records == nil {
		returnBadRequest(w, data.NewErrorResponseWithCode(http.StatusBadRequest, "records is required"))
		return
	}
	q := data.RecordQuery{
		AppID:   requestheader.GetAppID(r),
		Records: make([]interface{}, len(request.Records)),
	}
	for i, r := range request.Records {
		q.Records[i] = r
	}
	err = services.UpdateRecords(q, services.UpdateRecordIgnore(*request.Ignore))
	if err != nil {
		returnBadRequest(w, data.NewErrorResponseWithCode(http.StatusInternalServerError, "update records failed, "+err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
}

//NewRecordsMarkUpdateHandler create a handler to handle records mark & unmark request.
//The handler will update record store & ssm store.
//Because we separate the Init() function and Controller, the only way to pass dal.Client is from parameters.
func NewRecordsMarkUpdateHandler(client *dal.Client) func(w http.ResponseWriter, r *http.Request) {
	type internalError struct {
		IsRollbacked bool `json:"rollbacked"`
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
			returnBadRequest(w, data.NewErrorResponseWithCode(http.StatusBadRequest, err.Error()))
			return
		}
		if request.Mark == nil {
			returnBadRequest(w, data.NewErrorResponseWithCode(http.StatusBadRequest, "mark is required."))
			return
		}
		if request.Records == nil {
			returnBadRequest(w, data.NewErrorResponseWithCode(http.StatusBadRequest, "records is required"))
			return
		}
		appID := requestheader.GetAppID(r)
		q := data.RecordQuery{
			AppID:   appID,
			Records: make([]interface{}, len(request.Records)),
			Limit:   len(request.Records),
		}
		for i, r := range request.Records {
			q.Records[i] = r
		}
		//Need retrive record's userq
		result, err := services.VisitRecordsQuery(q)
		if err != nil {
			logger.Error.Printf("check record content failed, %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(internalError{
				IsRollbacked: true,
			})
			return
		}
		records_content := make([]string, 0)
		for _, record := range result.Hits {
			var isSQ bool
			isSQ, err = client.IsStandardQuestion(appID, record.UserQ)
			if err != nil {
				logger.Error.Printf("Get isStd failed, %v\n", err)
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(internalError{
					IsRollbacked: false,
				})
				return
			}

			if isSQ {
				w.WriteHeader(http.StatusConflict)
				json.NewEncoder(w).Encode(map[string]string{
					"conflicted_content": record.UserQ,
				})
				return
			}
			records_content = append(records_content, record.UserQ)
		}
		if *request.Mark {
			for _, content := range records_content {
				var isLQ bool
				isLQ, err = client.IsSimilarQuestion(appID, content)
				if err != nil {
					logger.Error.Printf("Get isSimQ failed, %v\n", err)
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(internalError{
						IsRollbacked: true,
					})
					return
				}
				if isLQ {
					err = client.DeleteSimilarQuestions(appID, content)
					if err != nil {
						//TODO: FIX AS ROLLBACKABLE
						logger.Error.Printf("delete sim q %v\n", err)
						w.WriteHeader(http.StatusInternalServerError)
						json.NewEncoder(w).Encode(internalError{
							IsRollbacked: false,
						})
						return
					}
				}
			}
			err = client.SetSimilarQuestion(requestheader.GetAppID(r), request.Content, records_content...)
			if err != nil {
				logger.Error.Printf("set simQ failed, %v\n", err)
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(internalError{
					IsRollbacked: false,
				})
				return
			}
		} else { //unmarking should remove the records from ssm store

			err = client.DeleteSimilarQuestions(appID, records_content...)
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

		err = services.UpdateRecords(q, services.UpdateRecordMark(*request.Mark))
		if err != nil {
			logger.Error.Printf("service update record failed, %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(internalError{
				IsRollbacked: true,
			})
			return
		}
		w.WriteHeader(http.StatusOK)
	}

}

// NewRecordSSMHandler create a handler for retriving ssm info of certain record.
func NewRecordSSMHandler(client *dal.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := mux.Vars(r)["id"]
		if !ok {
			returnBadRequest(w, data.NewErrorResponseWithCode(http.StatusBadRequest, "no path variable id"))
			return
		}
		appID := requestheader.GetAppID(r)
		query := data.RecordQuery{
			AppID:   appID,
			Records: []interface{}{id},
			Limit:   1,
		}
		result, err := services.VisitRecordsQuery(query)
		if err != nil {
			logger.Error.Printf("fetch es records failed, %v\n", err)
			returnInternalServerError(w, data.NewErrorResponse("internal server error"))
			return
		}
		if size := len(result.Hits); size == 0 || size > 1 {
			returnBadRequest(w, data.NewErrorResponseWithCode(http.StatusBadRequest, "record id "+id+" is ambiguous(results: "+strconv.Itoa(size)+")"))
			return
		}
		if !result.Hits[0].IsMarked {
			returnBadRequest(w, data.NewErrorResponseWithCode(http.StatusBadRequest, "record is not marked"))
		}
		lq := result.Hits[0].UserQ
		var response = struct {
			Content string `json:"marked_content"`
		}{}
		isLQ, err := client.IsSimilarQuestion(appID, lq)
		if err != nil {
			logger.Error.Printf("get isSimilarQ failed, %v", err)
			returnInternalServerError(w, data.NewErrorResponse("internal server error"))
			return
		}
		if !isLQ {
			response.Content = ""
			returnOK(w, response)
			return
		}

		response.Content, err = client.Question(appID, lq)
		if err != nil {
			logger.Error.Printf("get question for lq %s failed, %v", lq, err)
			returnInternalServerError(w, data.NewErrorResponse("internal server error"))
			return
		}
		returnOK(w, response)
	}
}

//newRecordQuery create a new *data.RecordQuery with given r.
//It should handle all the logic to retrive data,
//the only error should be returned is if request itself is invalided.
//It handled the limit & page logic for now, should move up to controller level.
func newRecordQuery(r *http.Request) (*data.RecordQuery, error) {
	// enterpriseID := requestheader.GetEnterpriseID(r)
	appID := requestheader.GetAppID(r)
	var query data.RecordQuery
	err := json.NewDecoder(r.Body).Decode(&query)
	if err != nil {
		return nil, fmt.Errorf("record query: request handled failed, %v", err)
	}
	query.AppID = appID

	return &query, nil
}
