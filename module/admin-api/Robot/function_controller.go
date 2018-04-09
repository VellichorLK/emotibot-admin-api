package Robot

import (
	"bytes"
	"fmt"
	"net/http"

	"emotibot.com/emotigo/module/admin-api/ApiError"
	"emotibot.com/emotigo/module/admin-api/util"
)

// handleList will show robot function list and it's status
func handleFunctionList(w http.ResponseWriter, r *http.Request) {
	appid := util.GetAppID(r)

	ret, errCode, err := GetFunctions(appid)
	if errCode != ApiError.SUCCESS {
		util.WriteJSON(w, util.GenRetObj(errCode, err))
	} else {
		util.WriteJSON(w, util.GenRetObj(errCode, ret))
	}
}

func handleUpdateFunction(w http.ResponseWriter, r *http.Request) {
	appid := util.GetAppID(r)
	function := util.GetMuxVar(r, "name")
	result := 0

	funcName, ok := util.Msg[function]
	if !ok {
		funcName = function
	}

	ret, errCode, err := GetFunctions(appid)
	if errCode != ApiError.SUCCESS {
		util.WriteJSON(w, util.GenRetObj(errCode, err))
		errMsg := fmt.Sprintf("%s [%s] %s", util.Msg["Read"], funcName, util.Msg["Error"])
		addAudit(r, util.AuditModuleFunctionSwitch, util.AuditOperationEdit,
			errMsg, result)
	}

	origInfo := ret[function]
	newInfo := loadFunctionFromContext(r)
	if newInfo == nil {
		errMsg := fmt.Sprintf("%s%s", util.Msg["Request"], util.Msg["Error"])
		addAudit(r, util.AuditModuleFunctionSwitch, util.AuditOperationEdit, errMsg, result)
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	errCode, err = UpdateFunction(appid, function, newInfo)
	origStatus := util.Msg["Close"]
	if origInfo.Status {
		origStatus = util.Msg["Open"]
	}
	newStatus := util.Msg["Close"]
	if newInfo.Status {
		newStatus = util.Msg["Open"]
	}
	auditLog := ""

	if errCode != ApiError.SUCCESS {
		util.WriteJSON(w, util.GenRetObj(errCode, err))
		errMsg := ApiError.GetErrorMsg(errCode)
		auditLog = fmt.Sprintf("%s%s, %s: [%s]->[%s] (%s)",
			util.Msg["Modify"], util.Msg["Error"], funcName, origStatus, newStatus, errMsg)
	} else {
		// http request to multicustomer
		// NOTE: no matter multicustomer return, return success
		// Terriable flow in old houta
		util.WriteJSON(w, util.GenSimpleRetObj(errCode))
		auditLog = fmt.Sprintf("%s%s, %s: [%s]->[%s]",
			util.Msg["Modify"], util.Msg["Success"], funcName, origStatus, newStatus)
		result = 1
		util.McUpdateFunction(appid)
	}
	addAudit(r, util.AuditModuleFunctionSwitch, util.AuditOperationEdit, auditLog, result)
	consulRet, err := util.ConsulUpdateFunctionStatus(appid)
	if err != nil {
		util.LogInfo.Printf("Update consul result: %d, %s", consulRet, err.Error())
	}
}

func handleUpdateAllFunction(w http.ResponseWriter, r *http.Request) {
	appid := util.GetAppID(r)
	result := 0

	origInfos, errCode, err := GetFunctions(appid)
	if errCode != ApiError.SUCCESS {
		errMsg := fmt.Sprintf("Get orig setting error: %s", ApiError.GetErrorMsg(errCode))
		util.WriteJSON(w, util.GenRetObj(errCode, err))
		addAudit(r, util.AuditModuleFunctionSwitch, util.AuditOperationEdit, errMsg, result)
		return
	}

	newInfos := loadFunctionsFromContext(r)
	if newInfos == nil {
		addAudit(r, util.AuditModuleFunctionSwitch, util.AuditOperationEdit, "Bad request", result)
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	var buffer bytes.Buffer
	for name, info := range newInfos {
		funcName, ok := util.Msg[name]
		if !ok {
			funcName = name
		}

		origStatus := util.Msg["Close"]
		origInfo, ok := origInfos[name]
		if ok && origInfo.Status {
			origStatus = util.Msg["Open"]
		}
		newStatus := util.Msg["Close"]
		if info.Status {
			newStatus = util.Msg["Open"]
		}
		buffer.WriteString(fmt.Sprintf("\n%s: [%s]->[%s]", funcName, origStatus, newStatus))
	}

	errCode, err = UpdateFunctions(appid, newInfos)
	auditLog := ""
	if errCode != ApiError.SUCCESS {
		util.WriteJSON(w, util.GenRetObj(errCode, err))
		errMsg := ApiError.GetErrorMsg(errCode)
		auditLog = fmt.Sprintf("%s%s: (%s)\n%s",
			util.Msg["Modify"], util.Msg["Error"], errMsg, buffer.String())
	} else {
		// http request to multicustomer
		// NOTE: no matter what multicustomer return, always return success
		// Terriable flow in old houta
		util.WriteJSON(w, util.GenSimpleRetObj(errCode))
		auditLog = fmt.Sprintf("%s%s:\n%s",
			util.Msg["Modify"], util.Msg["Success"], buffer.String())
		result = 1
		util.McUpdateFunction(appid)
	}
	addAudit(r, util.AuditModuleFunctionSwitch, util.AuditOperationEdit, auditLog, result)
	ret, err := util.ConsulUpdateRobotChat(appid)
	if err != nil {
		util.LogInfo.Printf("Update consul result: %d, %s", ret, err.Error())
	}
}

func loadFunctionFromContext(r *http.Request) *FunctionInfo {
	input := &FunctionInfo{}
	err := util.ReadJSON(r, input)
	if err != nil {
		util.LogInfo.Printf("Bad request when loading from input: %s", err.Error())
		return nil
	}

	return input
}

func loadFunctionsFromContext(r *http.Request) map[string]*FunctionInfo {
	input := make(map[string]*FunctionInfo)
	err := util.ReadJSON(r, &input)
	if err != nil {
		util.LogInfo.Printf("Bad request when loading from input: %s", err.Error())
		return nil
	}

	return input
}
