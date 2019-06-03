package v1

import (
	"emotibot.com/emotigo/module/admin-api/ELKStats/controllers"
	"emotibot.com/emotigo/module/admin-api/ELKStats/data"
	dataV1 "emotibot.com/emotigo/module/admin-api/ELKStats/data/v1"
	servicesV1 "emotibot.com/emotigo/module/admin-api/ELKStats/services/v1"
	"emotibot.com/emotigo/module/admin-api/util/elasticsearch"
	"emotibot.com/emotigo/module/admin-api/util/requestheader"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

func newRecordCcsQuery(r *http.Request) (*dataV1.CcsRecordQuery, error) {
	enterpriseID := requestheader.GetEnterpriseID(r)
	appID := requestheader.GetAppID(r)
	var query dataV1.CcsRecordQuery
	err := json.NewDecoder(r.Body).Decode(&query)
	if err != nil {
		return nil, fmt.Errorf("record query: request handled failed, %v", err)
	}

	query.EnterpriseID = enterpriseID
	query.AppID = appID

	return &query, nil
}


func VisitCcsRecordsGetHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	query, err := newRecordCcsQuery(r)
	if err != nil {
		errResponse := data.NewErrorResponseWithCode(data.ErrCodeInvalidRequestBody, err.Error())
		controllers.ReturnBadRequest(w, errResponse)
		return
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

	taskId, found := r.URL.Query()["task_id"]
	if found {
		query.TaskID = taskId[0]
	}

	if query.From+int64(query.Limit) > data.ESMaxResultWindow {
		errResponse := data.NewBadRequestResponse(data.ErrCodeInvalidParameterPage, "page")
		controllers.ReturnBadRequest(w, errResponse)
		return
	}

	result, err := servicesV1.VisitCcsRecordsQuery(*query)
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

	response := dataV1.VisitCcsRecordsResponse{
		Data:        result.Hits,
		Limit:       query.Limit,
		Page:        int(query.From)/query.Limit + 1,
		TotalSize:   result.Aggs["total_size"].(int64),
	}
	controllers.ReturnOK(w, response)
}