package common

import (
	"net/http"

	"emotibot.com/emotigo/module/admin-api/ELKStats/controllers"
	"emotibot.com/emotigo/module/admin-api/ELKStats/data"
	dataCommon "emotibot.com/emotigo/module/admin-api/ELKStats/data/common"
	servicesCommon "emotibot.com/emotigo/module/admin-api/ELKStats/services/common"
	"emotibot.com/emotigo/module/admin-api/util"
)

func VisitRecordsExportDownloadHandler(w http.ResponseWriter, r *http.Request) {
	exportID := util.GetMuxVar(r, "export_id")
	if exportID == "" {
		errResponse := data.NewBadRequestResponse(data.ErrCodeInvalidParameterExportID, "export_id")
		controllers.ReturnBadRequest(w, errResponse)
		return
	}

	err := servicesCommon.VisitRecordsExportDownload(w, exportID)
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

func VisitRecordsExportDeleteHandler(w http.ResponseWriter, r *http.Request) {
	exportID := util.GetMuxVar(r, "export_id")
	if exportID == "" {
		errResponse := data.NewBadRequestResponse(data.ErrCodeInvalidParameterExportID, "export_id")
		controllers.ReturnBadRequest(w, errResponse)
		return
	}

	err := servicesCommon.VisitRecordsExportDelete(exportID)
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

func VisitRecordsExportStatusHandler(w http.ResponseWriter, r *http.Request) {
	exportID := util.GetMuxVar(r, "export_id")
	if exportID == "" {
		errResponse := data.NewBadRequestResponse(data.ErrCodeInvalidParameterExportID, "export_id")
		controllers.ReturnBadRequest(w, errResponse)
		return
	}

	status, err := servicesCommon.VisitRecordsExportStatus(exportID)
	if err != nil {
		if err == data.ErrExportTaskNotFound {
			controllers.ReturnNotFoundRequest(w, data.NewErrorResponse(err.Error()))
		} else {
			controllers.ReturnInternalServerError(w, data.NewErrorResponse(data.ErrNotInit.Error()))
		}
		return
	}

	response := dataCommon.VisitRecordsExportStatusResponse{
		Status: status,
	}

	controllers.ReturnOK(w, response)
}
