package Robot

import (
	"bytes"
	"fmt"

	"github.com/kataras/iris"

	"emotibot.com/emotigo/module/admin-api/ApiError"
	"emotibot.com/emotigo/module/admin-api/util"
	"github.com/kataras/iris/context"
)

// handleList will show robot function list and it's status
func handleFunctionList(ctx context.Context) {
	appid := util.GetAppID(ctx)

	ret, errCode, err := GetFunctions(appid)
	if errCode != ApiError.SUCCESS {
		ctx.JSON(util.GenRetObj(errCode, err))
	} else {
		ctx.JSON(util.GenRetObj(errCode, ret))
	}
}

func handleUpdateFunction(ctx context.Context) {
	appid := util.GetAppID(ctx)
	function := ctx.Params().GetEscape("name")
	result := 0

	funcName, ok := util.Msg[function]
	if !ok {
		funcName = function
	}

	ret, errCode, err := GetFunctions(appid)
	if errCode != ApiError.SUCCESS {
		ctx.JSON(util.GenRetObj(errCode, err))
		errMsg := fmt.Sprintf("%s [%s] %s", util.Msg["Read"], funcName, util.Msg["Error"])
		addAudit(ctx, util.AuditModuleFunctionSwitch, util.AuditOperationEdit,
			errMsg, result)
	}

	origInfo := ret[function]
	newInfo := loadFunctionFromContext(ctx)
	if newInfo == nil {
		errMsg := fmt.Sprintf("%s%s", util.Msg["Request"], util.Msg["Error"])
		addAudit(ctx, util.AuditModuleFunctionSwitch, util.AuditOperationEdit, errMsg, result)
		ctx.StatusCode(iris.StatusBadRequest)
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
		ctx.JSON(util.GenRetObj(errCode, err))
		errMsg := ApiError.GetErrorMsg(errCode)
		auditLog = fmt.Sprintf("%s%s, %s: [%s]->[%s] (%s)",
			util.Msg["Modify"], util.Msg["Error"], funcName, origStatus, newStatus, errMsg)
	} else {
		// http request to multicustomer
		// NOTE: no matter multicustomer return, return success
		// Terriable flow in old houta
		ctx.JSON(util.GenSimpleRetObj(errCode))
		auditLog = fmt.Sprintf("%s%s, %s: [%s]->[%s]",
			util.Msg["Modify"], util.Msg["Success"], funcName, origStatus, newStatus)
		result = 1
		util.McUpdateFunction(appid)
	}
	addAudit(ctx, util.AuditModuleFunctionSwitch, util.AuditOperationEdit, auditLog, result)
}

func handleUpdateAllFunction(ctx context.Context) {
	appid := util.GetAppID(ctx)
	result := 0

	origInfos, errCode, err := GetFunctions(appid)
	if errCode != ApiError.SUCCESS {
		errMsg := fmt.Sprintf("Get orig setting error: %s", ApiError.GetErrorMsg(errCode))
		ctx.JSON(util.GenRetObj(errCode, err))
		addAudit(ctx, util.AuditModuleFunctionSwitch, util.AuditOperationEdit, errMsg, result)
		return
	}

	newInfos := loadFunctionsFromContext(ctx)
	if newInfos == nil {
		addAudit(ctx, util.AuditModuleFunctionSwitch, util.AuditOperationEdit, "Bad request", result)
		ctx.StatusCode(iris.StatusBadRequest)
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
		ctx.JSON(util.GenRetObj(errCode, err))
		errMsg := ApiError.GetErrorMsg(errCode)
		auditLog = fmt.Sprintf("%s%s: (%s)\n%s",
			util.Msg["Modify"], util.Msg["Error"], errMsg, buffer.String())
	} else {
		// http request to multicustomer
		// NOTE: no matter multicustomer return, return success
		// Terriable flow in old houta
		ctx.JSON(util.GenSimpleRetObj(errCode))
		auditLog = fmt.Sprintf("%s%s:\n%s",
			util.Msg["Modify"], util.Msg["Success"], buffer.String())
		result = 1
		util.McUpdateFunction(appid)
	}
	addAudit(ctx, util.AuditModuleFunctionSwitch, util.AuditOperationEdit, auditLog, result)
}

func loadFunctionFromContext(ctx context.Context) *FunctionInfo {
	input := &FunctionInfo{}
	err := ctx.ReadJSON(input)
	if err != nil {
		util.LogInfo.Printf("Bad request when loading from input: %s", err.Error())
		return nil
	}

	return input
}

func loadFunctionsFromContext(ctx context.Context) map[string]*FunctionInfo {
	input := make(map[string]*FunctionInfo)
	err := ctx.ReadJSON(&input)
	if err != nil {
		util.LogInfo.Printf("Bad request when loading from input: %s", err.Error())
		return nil
	}

	return input
}
