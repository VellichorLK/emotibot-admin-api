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

func TEVisitRecordsGetHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	query, err, errCode := newTERecordQuery(r)
	if err != nil {
		errResponse := data.NewErrorResponseWithCode(errCode, err.Error())
		controllers.ReturnBadRequest(w, errResponse)
		return
	}

	if query.StartTime.After(query.EndTime) {
		errResponse := data.NewErrorResponseWithCode(data.ErrCodeInvalidParameterStartTime,
			"start_time cannot greater than end_time")
		controllers.ReturnBadRequest(w, errResponse)
		return
	}

	result, totalSize, err := servicesV1.TEVisitRecordsQuery(query)
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
	response := dataV1.TEVisitRecordsResponse{
		TableHeader: dataV1.TEVisitRecordsTableHeader[locale],
		Data:        result,
		Limit:       query.Limit,
		Page:        query.From/query.Limit + 1,
		TotalSize:   totalSize,
	}

	controllers.ReturnOK(w, response)
}

func TEVisitRecordsExportHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	query, err, errCode := newTERecordQuery(r)
	if err != nil {
		errResponse := data.NewErrorResponseWithCode(errCode, err.Error())
		controllers.ReturnBadRequest(w, errResponse)
		return
	}

	if query.StartTime.After(query.EndTime) {
		errResponse := data.NewErrorResponseWithCode(data.ErrCodeInvalidParameterStartTime,
			"start_time cannot greater than end_time")
		controllers.ReturnBadRequest(w, errResponse)
		return
	}

	exportTaskID, err := servicesV1.TEVisitRecordsExport(query)
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

	response := dataV1.TEVisitRecordsExportResponse{
		ExportID: exportTaskID,
	}

	controllers.ReturnOK(w, response)
}

func TEVisitRecordsExportDownloadHandler(w http.ResponseWriter, r *http.Request) {
	exportID := util.GetMuxVar(r, "export_id")
	if exportID == "" {
		errResponse := data.NewBadRequestResponse(data.ErrCodeInvalidParameterExportID, "export_id")
		controllers.ReturnBadRequest(w, errResponse)
		return
	}

	err := servicesV1.TEVisitRecordsExportDownload(w, exportID)
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

func TEVisitRecordsExportDeleteHandler(w http.ResponseWriter, r *http.Request) {
	exportID := util.GetMuxVar(r, "export_id")
	if exportID == "" {
		errResponse := data.NewBadRequestResponse(data.ErrCodeInvalidParameterExportID, "export_id")
		controllers.ReturnBadRequest(w, errResponse)
		return
	}

	err := servicesV1.TEVisitRecordsExportDelete(exportID)
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

func TEVisitRecordsExportStatusHandler(w http.ResponseWriter, r *http.Request) {
	exportID := util.GetMuxVar(r, "export_id")
	if exportID == "" {
		errResponse := data.NewBadRequestResponse(data.ErrCodeInvalidParameterExportID, "export_id")
		controllers.ReturnBadRequest(w, errResponse)
		return
	}

	status, err := servicesV1.TEVisitRecordsExportStatus(exportID)
	if err != nil {
		if err == data.ErrExportTaskNotFound {
			controllers.ReturnNotFoundRequest(w, data.NewErrorResponse(err.Error()))
		} else {
			controllers.ReturnInternalServerError(w, data.NewErrorResponse(data.ErrNotInit.Error()))
		}
		return
	}

	response := dataV1.TEVisitRecordsExportStatusResponse{
		Status: status,
	}

	controllers.ReturnOK(w, response)
}

func newTERecordQuery(r *http.Request) (query *dataV1.TEVisitRecordsQuery, err error, errCode int) {
	request := dataV1.TEVisitRecordsRequest{}
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
	var genders []string

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

	if request.Genders != nil {
		for _, gender := range request.Genders {
			g, found := services.GetTagNameByID(appID, "sex", gender)
			if !found {
				err = data.ErrInvalidRequestBody
				errCode = data.ErrCodeInvalidParameterGender
				return
			}

			genders = append(genders, g)
		}
	}

	var page, limit int64

	if request.Page == nil {
		page = dataCommon.TEVisitRecordsDefaultPage
	} else {
		page = *request.Page
	}

	if request.Limit == nil {
		limit = dataCommon.TEVisitRecordsDefaultPageLimit
	} else {
		limit = *request.Limit
	}

	query = &dataV1.TEVisitRecordsQuery{
		CommonQuery: data.CommonQuery{
			EnterpriseID: enterpriseID,
			AppID:        appID,
			StartTime:    startTime,
			EndTime:      endTime,
		},
		ScenarioName: request.ScenarioName,
		Platforms:    platforms,
		Genders:      genders,
		UserID:       request.UserID,
		Feedback:     request.Feedback,
		From:         (page - 1) * int64(limit),
		Limit:        limit,
	}

	if query.From+query.Limit > data.ESMaxResultWindow {
		err = data.ErrInvalidRequestBody
		errCode = data.ErrCodeInvalidParameterPage
		return
	}

	return
}
