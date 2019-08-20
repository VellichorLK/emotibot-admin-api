package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"emotibot.com/emotigo/module/token-auth/internal/audit"
	"emotibot.com/emotigo/module/token-auth/internal/data"
	"emotibot.com/emotigo/module/token-auth/internal/util"
	"emotibot.com/emotigo/module/token-auth/service"
	"github.com/gorilla/mux"
)

func AppAddHandlerV5(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var app *data.AppDetailV5
	var err error

	defer func() {
		// Add audit log
		var auditMessage string
		if app != nil {
			auditMessage = fmt.Sprintf("[%s]: %s", data.AuditContentAppAdd, app.Name)
		} else {
			auditMessage = data.AuditContentAppAdd
		}

		addAuditLog(r, audit.AuditModuleManageRobot, audit.AuditOperationAdd,
			auditMessage, err)
	}()

	enterpriseID := vars["enterpriseID"]
	if !util.IsValidUUID(enterpriseID) {
		err = util.ErrInvalidParameter
		returnBadRequest(w, "enterprise-id")
		return
	}

	app, err = parseAppFromRequestV5(r)
	if err != nil {
		returnBadRequest(w, err.Error())
		return
	}

	if app.Name == "" {
		err = util.ErrInvalidParameter
		returnBadRequest(w, "name")
		return
	}

	id, err := service.AddAppV5(enterpriseID, app)
	if err != nil {
		switch err {
		case util.ErrAppInfoExists:
			returnBadRequest(w, "name")
		case util.ErrOperationForbidden:
			returnForbiddenWithMsg(w, err.Error())
		default:
			returnInternalError(w, err.Error())
		}
		return
	} else if id == "" {
		returnBadRequest(w, "enterprise-id")
		return
	}

	newApp, err := service.GetAppV5(enterpriseID, id)
	if err != nil {
		returnInternalError(w, err.Error())
		return
	} else if newApp == nil {
		err = util.ErrInteralServer
		returnInternalError(w, err.Error())
		return
	}
	returnSuccess(w, newApp)
}

func AppsGetHandlerV5(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	enterpriseID := vars["enterpriseID"]
	if !util.IsValidUUID(enterpriseID) {
		returnBadRequest(w, "enterprise-id")
		return
	}

	retData, err := service.GetAppsV5(enterpriseID)
	if err != nil {
		returnInternalError(w, err.Error())
		return
	} else if retData == nil {
		returnNotFound(w)
	}

	returnSuccess(w, retData)
}

func AppGetHandlerV5(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	enterpriseID := vars["enterpriseID"]
	if !util.IsValidUUID(enterpriseID) {
		returnBadRequest(w, "enterprise-id")
		return
	}

	appID := vars["appID"]
	if !util.IsValidUUID(appID) {
		returnBadRequest(w, "app-id")
		return
	}

	retData, err := service.GetAppV5(enterpriseID, appID)
	if err != nil {
		returnInternalError(w, err.Error())
		return
	} else if retData == nil {
		returnNotFound(w)
		return
	}

	returnSuccess(w, retData)
}

func parseAppFromRequestV5(r *http.Request) (*data.AppDetailV5, error) {
	name := strings.TrimSpace(r.FormValue("name"))
	description := r.FormValue("description")
	appType, _ := strconv.Atoi(r.FormValue("app_type"))

	var props []*data.AppPropV5
	err := json.Unmarshal([]byte(r.FormValue("props")), &props)
	if err != nil {
		return nil, err
	}

	ret := data.AppDetailV5{
		AppV5: data.AppV5{
			Name:    name,
			AppType: appType,
			Props:   props,
		},
		Description: description,
	}

	return &ret, nil
}

func AppPropsGetHandlerV5(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var pKey string
	if _, ok := vars["key"]; ok {
		pKey = vars["key"]
	}

	retData, err := service.AppPropsGetV5(pKey)

	if err != nil {
		returnInternalError(w, err.Error())
		return
	} else if retData == nil {
		returnNotFound(w)
		return
	}

	returnSuccess(w, retData)
}
