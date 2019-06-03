package v2

import (
	"emotibot.com/emotigo/module/admin-api/ELKStats/controllers"
	"emotibot.com/emotigo/module/admin-api/ELKStats/data"
	dataCommon "emotibot.com/emotigo/module/admin-api/ELKStats/data/common"
	dataV2 "emotibot.com/emotigo/module/admin-api/ELKStats/data/v2"
	servicesV2 "emotibot.com/emotigo/module/admin-api/ELKStats/services/v2"
	"emotibot.com/emotigo/module/admin-api/util/elasticsearch"
	"emotibot.com/emotigo/module/admin-api/util/requestheader"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

func VisitCcsRecordsGetHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	query, err, errCode := NewCcsRecordQuery(r)
	if err != nil {
		errResponse := data.NewErrorResponseWithCode(errCode, err.Error())
		controllers.ReturnBadRequest(w, errResponse)
		return
	}

	result, err := servicesV2.VisitCcsRecordsQuery(query)
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

	response := dataV2.VisitCcsRecordsResponse{
		TableHeader: dataV2.VisitCcsRecordsTableHeader,
		Data:        result.Data,
		TotalSize:   result.TotalSize,
		Page:        query.From/query.Limit + 1,
		Limit:       query.Limit,
	}

	controllers.ReturnOK(w, response)
}

func NewCcsRecordQuery(r *http.Request) (query *dataV2.VisitCcsRecordsQuery, err error, errCode int) {
	request := dataV2.VisitCcsRecordsRequest{}
	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		errCode = data.ErrCodeInvalidRequestBody
		return
	}

	enterpriseID := requestheader.GetEnterpriseID(r)
	appID := requestheader.GetAppID(r)

	startTime := time.Unix(request.StartTime, 0).UTC()
	endTime := time.Unix(request.EndTime, 0).UTC()

	var page, limit int64

	if request.Page == nil {
		page = dataCommon.RecordsDefaultPage
	} else {
		page = *request.Page
	}

	if request.Limit == nil {
		limit = dataCommon.RecordsDefaultPageLimit
	} else {
		limit = *request.Limit
	}

	query = &dataV2.VisitCcsRecordsQuery{
		CommonQuery: data.CommonQuery{
			EnterpriseID: enterpriseID,
			AppID:        appID,
			StartTime:    startTime,
			EndTime:      endTime,
		},
		SessionID:     request.SessionID,
		From:          (page - 1) * int64(limit),
		Limit:         limit,
	}

	if query.From+query.Limit > data.ESMaxResultWindow {
		err = data.ErrInvalidRequestBody
		errCode = data.ErrCodeInvalidParameterPage
		return
	}

	return
}