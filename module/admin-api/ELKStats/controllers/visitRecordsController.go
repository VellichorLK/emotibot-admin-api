package controllers

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

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

	result, err := services.VisitRecordsQuery(*query, services.ElasticFilterIgnoredRecord, services.ElasticFilterMarkedRecord)
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

//RecordsMarkUpdateHandler handle records mark & unmark request.
func RecordsMarkUpdateHandler(w http.ResponseWriter, r *http.Request) {
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
	q := data.RecordQuery{
		AppID:   requestheader.GetAppID(r),
		Records: make([]interface{}, len(request.Records)),
	}
	for i, r := range request.Records {
		q.Records[i] = r
	}

	if *request.Mark {
		//FIXME: Implement DAL accessing
	}
	err = services.UpdateRecords(q, services.UpdateRecordMark(*request.Mark))
	if err != nil {
		returnBadRequest(w, data.NewErrorResponseWithCode(http.StatusInternalServerError, "update records failed, "+err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
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
