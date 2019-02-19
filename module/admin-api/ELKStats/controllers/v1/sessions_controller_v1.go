package v1

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"emotibot.com/emotigo/module/admin-api/ELKStats/controllers"
	"emotibot.com/emotigo/module/admin-api/ELKStats/data"
	dataCommon "emotibot.com/emotigo/module/admin-api/ELKStats/data/common"
	dataV1 "emotibot.com/emotigo/module/admin-api/ELKStats/data/v1"
	services "emotibot.com/emotigo/module/admin-api/ELKStats/services"
	servicesV1 "emotibot.com/emotigo/module/admin-api/ELKStats/services/v1"
	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/admin-api/util/elasticsearch"
	"emotibot.com/emotigo/module/admin-api/util/requestheader"
)

func SessionsGetHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	query, err, errCode := newSessionQuery(r)
	if err != nil {
		errResponse := data.NewErrorResponseWithCode(errCode, err.Error())
		controllers.ReturnBadRequest(w, errResponse)
		return
	}

	result, totalSize, err := servicesV1.SessionsQuery(query)
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

	locale := requestheader.GetLocale(r)
	response := dataV1.SessionsResponse{
		TableHeader: dataV1.SessionsTableHeader[locale],
		Data:        result,
		Limit:       query.Limit,
		Page:        query.From/query.Limit + 1,
		TotalSize:   totalSize,
	}

	controllers.ReturnOK(w, response)
}

func SessionsExportHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	query, err, errCode := newSessionQuery(r)
	if err != nil {
		errResponse := data.NewErrorResponseWithCode(errCode, err.Error())
		controllers.ReturnBadRequest(w, errResponse)
		return
	}

	exportTaskID, err := servicesV1.SessionsExport(query)
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

	response := dataV1.SessionsExportResponse{
		ExportID: exportTaskID,
	}

	controllers.ReturnOK(w, response)
}

func SessionsExportDownloadHandler(w http.ResponseWriter, r *http.Request) {
	exportID := util.GetMuxVar(r, "export_id")
	if exportID == "" {
		errResponse := data.NewBadRequestResponse(data.ErrCodeInvalidParameterExportID, "export_id")
		controllers.ReturnBadRequest(w, errResponse)
		return
	}

	err := servicesV1.SessionsExportDownload(w, exportID)
	if err != nil {
		switch err {
		case data.ErrExportTaskNotFound:
			controllers.ReturnNotFoundRequest(w, data.NewErrorResponse(err.Error()))
		case data.ErrExportTaskInProcess:
			controllers.ReturnForbiddenRequest(w, data.NewErrorResponse(err.Error()))
		case data.ErrExportTaskEmpty:
			w.WriteHeader(http.StatusNoContent)
		default:
			controllers.ReturnInternalServerError(w, data.NewErrorResponse(err.Error()))
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

func SessionsExportDeleteHandler(w http.ResponseWriter, r *http.Request) {
	exportID := util.GetMuxVar(r, "export_id")
	if exportID == "" {
		errResponse := data.NewBadRequestResponse(data.ErrCodeInvalidParameterExportID, "export_id")
		controllers.ReturnBadRequest(w, errResponse)
		return
	}

	err := servicesV1.SessionsExportDelete(exportID)
	if err != nil {
		if err == data.ErrExportTaskNotFound {
			controllers.ReturnNotFoundRequest(w, data.NewErrorResponse(err.Error()))
		} else {
			controllers.ReturnInternalServerError(w, data.NewErrorResponse(data.ErrNotInit.Error()))
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

func SessionsExportStatusHandler(w http.ResponseWriter, r *http.Request) {
	exportID := util.GetMuxVar(r, "export_id")
	if exportID == "" {
		errResponse := data.NewBadRequestResponse(data.ErrCodeInvalidParameterExportID, "export_id")
		controllers.ReturnBadRequest(w, errResponse)
		return
	}

	status, err := servicesV1.SessionsExportStatus(exportID)
	if err != nil {
		if err == data.ErrExportTaskNotFound {
			controllers.ReturnNotFoundRequest(w, data.NewErrorResponse(err.Error()))
		} else {
			controllers.ReturnInternalServerError(w, data.NewErrorResponse(data.ErrNotInit.Error()))
		}
		return
	}

	response := dataV1.SessionsExportStatusResponse{
		Status: status,
	}

	controllers.ReturnOK(w, response)
}

func newSessionQuery(r *http.Request) (query *dataV1.SessionsQuery, err error, errCode int) {
	request := dataV1.SessionsRequest{}
	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		errCode = data.ErrCodeInvalidRequestBody
		return
	}

	enterpriseID := requestheader.GetEnterpriseID(r)
	appID := requestheader.GetAppID(r)

	startTime := time.Unix(request.StartTime, 0).UTC()
	endTime := time.Unix(request.EndTime, 0).UTC()

	var platforms []string
	var sex []string

	if request.Platforms != nil {
		for _, platform := range request.Platforms {
			p, found := services.GetTagNameByID(appID, "platform", platform)
			if !found {
				err = data.ErrInvalidRequestBody
				errCode = data.ErrCodeInvalidParameterPlatform
				return
			}

			platforms = append(platforms, p)
		}
	}

	if request.Sex != nil {
		for _, _sex := range request.Sex {
			s, found := services.GetTagNameByID(appID, "sex", _sex)
			if !found {
				err = data.ErrInvalidRequestBody
				errCode = data.ErrCodeInvalidParameterGender
				return
			}

			sex = append(sex, s)
		}
	}

	var page, limit int64

	if request.Page == nil {
		page = dataCommon.SessionsDefaultPage
	} else {
		page = *request.Page
	}
	if page <= 0 {
		page = dataCommon.SessionsDefaultPage
	}

	if request.Limit == nil {
		limit = dataCommon.SessionsDefaultPageLimit
	} else {
		limit = *request.Limit
	}

	query = &dataV1.SessionsQuery{
		CommonQuery: data.CommonQuery{
			EnterpriseID: enterpriseID,
			AppID:        appID,
			StartTime:    startTime,
			EndTime:      endTime,
		},
		Platforms: platforms,
		Sex:       sex,
		UserID:    request.UserID,
		RatingMax: request.RatingMax,
		RatingMin: request.RatingMin,
		Feedback:  request.Feedback,
		From:      (page - 1) * int64(limit),
		Limit:     limit,
		SessionID: request.SessionID,
	}

	if request.FeedbackStartTime != nil {
		feedbackStartTime := time.Unix(*request.FeedbackStartTime, 0).UTC()
		query.FeedbackStartTime = &feedbackStartTime
	}

	if request.FeedbackEndTime != nil {
		feedbackEndTime := time.Unix(*request.FeedbackEndTime, 0).UTC()
		query.FeedbackEndTime = &feedbackEndTime
	}

	if query.From+query.Limit > data.ESMaxResultWindow {
		err = data.ErrInvalidRequestBody
		errCode = data.ErrCodeInvalidParameterPage
		return
	}

	return
}
