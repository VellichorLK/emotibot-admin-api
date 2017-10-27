package Robot

import (
	"encoding/json"
	"fmt"

	"github.com/kataras/iris"

	"emotibot.com/emotigo/module/vipshop-admin/ApiError"
	"emotibot.com/emotigo/module/vipshop-admin/util"
	"github.com/kataras/iris/context"
)

var (
	// ModuleInfo is needed for module define
	ModuleInfo util.ModuleInfo
)

func init() {
	ModuleInfo = util.ModuleInfo{
		ModuleName: "robot",
		EntryPoints: []util.EntryPoint{
			util.NewEntryPoint("GET", "functions", []string{"view"}, handleList),
			util.NewEntryPoint("POST", "functions", []string{"edit"}, handleUpdateAllFunction),
			util.NewEntryPoint("POST", "function/{name:string}", []string{"edit"}, handleUpdateFunction),
		},
	}
}

func getEnvironments() map[string]string {
	return util.GetEnvOf(ModuleInfo.ModuleName)
}

func getEnvironment(key string) string {
	envs := util.GetEnvOf(ModuleInfo.ModuleName)
	if envs != nil {
		if val, ok := envs[key]; ok {
			return val
		}
	}
	return ""
}

func getGlobalEnv(key string) string {
	envs := util.GetEnvOf("server")
	if envs != nil {
		if val, ok := envs[key]; ok {
			return val
		}
	}
	return ""
}

// handleList will show robot function list and it's status
func handleList(ctx context.Context) {
	appid := util.GetAppID(ctx)

	ret, errCode, err := GetFunctions(appid)
	errMsg := ApiError.GetErrorMsg(errCode)
	if errCode != ApiError.SUCCESS {
		ctx.JSON(util.GenRetObj(errCode, errMsg, err))
	} else {
		ctx.JSON(util.GenRetObj(errCode, errMsg, ret))
	}
}

func handleUpdateFunction(ctx context.Context) {
	appid := util.GetAppID(ctx)
	function := ctx.Params().GetEscape("name")
	result := 0

	ret, errCode, err := GetFunctions(appid)
	errMsg := ApiError.GetErrorMsg(errCode)
	if errCode != ApiError.SUCCESS {
		ctx.JSON(util.GenRetObj(errCode, errMsg, err))
		addAudit(ctx, util.AuditOperationEdit,
			fmt.Sprintf("Get orig setting of %s error", function), result)
	}

	origInfo := ret[function]
	newInfo := loadFunctionFromContext(ctx)
	if newInfo == nil {
		addAudit(ctx, util.AuditOperationEdit, "Bad request", result)
		ctx.StatusCode(iris.StatusBadRequest)
		return
	}

	errCode, err = UpdateFunction(appid, function, newInfo)
	errMsg = ApiError.GetErrorMsg(errCode)

	origStr, _ := json.Marshal(origInfo)
	newStr, _ := json.Marshal(newInfo)
	auditLog := ""

	if errCode != ApiError.SUCCESS {
		ctx.JSON(util.GenRetObj(errCode, errMsg, err))
		auditLog = fmt.Sprintf("Error[%s] %s => %s", errMsg, string(origStr), string(newStr))
	} else {
		// http request to multicustomer
		// NOTE: no matter multicustomer return, return success
		// Terriable flow in old houta
		errMsg := ApiError.GetErrorMsg(errCode)
		ctx.JSON(util.GenSimpleRetObj(errCode, errMsg))

		auditLog = fmt.Sprintf("%s: [%s] => [%s]", function, string(origStr), string(newStr))
		result = 1
		util.McUpdateFunction(appid)
	}
	addAudit(ctx, util.AuditOperationEdit, auditLog, result)
}

func handleUpdateAllFunction(ctx context.Context) {
	appid := util.GetAppID(ctx)
	result := 0

	origInfos, errCode, err := GetFunctions(appid)
	errMsg := ApiError.GetErrorMsg(errCode)
	if errCode != ApiError.SUCCESS {
		ctx.JSON(util.GenRetObj(errCode, errMsg, err))
		addAudit(ctx, util.AuditOperationEdit, "Get orig setting error", result)
		return
	}

	newInfos := loadFunctionsFromContext(ctx)
	if newInfos == nil {
		addAudit(ctx, util.AuditOperationEdit, "Bad request", result)
		ctx.StatusCode(iris.StatusBadRequest)
		return
	}

	errCode, err = UpdateFunctions(appid, newInfos)
	errMsg = ApiError.GetErrorMsg(errCode)

	origStr, _ := json.Marshal(origInfos)
	newStr, _ := json.Marshal(newInfos)
	auditLog := ""

	if errCode != ApiError.SUCCESS {
		ctx.JSON(util.GenRetObj(errCode, errMsg, err))
		auditLog = fmt.Sprintf("Error[%s] %s => %s", errMsg, string(origStr), string(newStr))
	} else {
		// http request to multicustomer
		// NOTE: no matter multicustomer return, return success
		// Terriable flow in old houta
		errMsg := ApiError.GetErrorMsg(errCode)
		ctx.JSON(util.GenSimpleRetObj(errCode, errMsg))
		auditLog = fmt.Sprintf("%s => %s", string(origStr), string(newStr))
		result = 1
		util.McUpdateFunction(appid)
	}
	addAudit(ctx, util.AuditOperationEdit, auditLog, result)
}

func addAudit(ctx context.Context, operation string, msg string, result int) {
	userID := util.GetUserID(ctx)
	userIP := util.GetUserIP(ctx)

	util.AddAuditLog(userID, userIP, util.AuditModuleSwitchList, operation, msg, result)
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
