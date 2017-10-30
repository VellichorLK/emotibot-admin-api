package Robot

import (
	"encoding/json"
	"fmt"

	"github.com/kataras/iris"

	"emotibot.com/emotigo/module/vipshop-admin/ApiError"
	"emotibot.com/emotigo/module/vipshop-admin/util"
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

	ret, errCode, err := GetFunctions(appid)
	errMsg := ApiError.GetErrorMsg(errCode)
	if errCode != ApiError.SUCCESS {
		ctx.JSON(util.GenRetObj(errCode, err))
		addAudit(ctx, util.AuditModuleFunctionSwitch, util.AuditOperationEdit,
			fmt.Sprintf("Get orig setting of %s error", function), result)
	}

	origInfo := ret[function]
	newInfo := loadFunctionFromContext(ctx)
	if newInfo == nil {
		addAudit(ctx, util.AuditModuleFunctionSwitch, util.AuditOperationEdit, "Bad request", result)
		ctx.StatusCode(iris.StatusBadRequest)
		return
	}

	errCode, err = UpdateFunction(appid, function, newInfo)
	errMsg = ApiError.GetErrorMsg(errCode)

	origStr, _ := json.Marshal(origInfo)
	newStr, _ := json.Marshal(newInfo)
	auditLog := ""

	if errCode != ApiError.SUCCESS {
		ctx.JSON(util.GenRetObj(errCode, err))
		auditLog = fmt.Sprintf("Error[%s] %s => %s", errMsg, string(origStr), string(newStr))
	} else {
		// http request to multicustomer
		// NOTE: no matter multicustomer return, return success
		// Terriable flow in old houta
		ctx.JSON(util.GenSimpleRetObj(errCode))

		auditLog = fmt.Sprintf("%s: [%s] => [%s]", function, string(origStr), string(newStr))
		result = 1
		util.McUpdateFunction(appid)
	}
	addAudit(ctx, util.AuditModuleFunctionSwitch, util.AuditOperationEdit, auditLog, result)
}

func handleUpdateAllFunction(ctx context.Context) {
	appid := util.GetAppID(ctx)
	result := 0

	origInfos, errCode, err := GetFunctions(appid)
	errMsg := ApiError.GetErrorMsg(errCode)
	if errCode != ApiError.SUCCESS {
		ctx.JSON(util.GenRetObj(errCode, err))
		addAudit(ctx, util.AuditModuleFunctionSwitch, util.AuditOperationEdit, "Get orig setting error", result)
		return
	}

	newInfos := loadFunctionsFromContext(ctx)
	if newInfos == nil {
		addAudit(ctx, util.AuditModuleFunctionSwitch, util.AuditOperationEdit, "Bad request", result)
		ctx.StatusCode(iris.StatusBadRequest)
		return
	}

	errCode, err = UpdateFunctions(appid, newInfos)
	errMsg = ApiError.GetErrorMsg(errCode)

	origStr, _ := json.Marshal(origInfos)
	newStr, _ := json.Marshal(newInfos)
	auditLog := ""

	if errCode != ApiError.SUCCESS {
		ctx.JSON(util.GenRetObj(errCode, err))
		auditLog = fmt.Sprintf("Error[%s] %s => %s", errMsg, string(origStr), string(newStr))
	} else {
		// http request to multicustomer
		// NOTE: no matter multicustomer return, return success
		// Terriable flow in old houta
		ctx.JSON(util.GenSimpleRetObj(errCode))
		auditLog = fmt.Sprintf("%s => %s", string(origStr), string(newStr))
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
