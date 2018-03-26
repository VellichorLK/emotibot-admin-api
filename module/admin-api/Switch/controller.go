package Switch

import (
	"bytes"
	"fmt"
	"reflect"
	"time"

	"emotibot.com/emotigo/module/admin-api/ApiError"
	"emotibot.com/emotigo/module/admin-api/util"
	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
)

var (
	// ModuleInfo is needed for module define
	ModuleInfo util.ModuleInfo
)

func init() {
	ModuleInfo = util.ModuleInfo{
		ModuleName: "switch-manage",
		EntryPoints: []util.EntryPoint{
			util.NewEntryPoint("GET", "switches", []string{"view"}, handleList),
			util.NewEntryPoint("GET", "switch/{id:int}", []string{"view"}, handleSwitch),
			util.NewEntryPoint("POST", "switch/{id:int}", []string{"edit"}, handleUpdateSwitch),
			util.NewEntryPoint("DELETE", "switch/{id:int}", []string{"delete"}, handleDeleteSwitch),
			util.NewEntryPoint("PUT", "switch/create", []string{"create"}, handleNewSwitch),
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

// handleList will show onoff list in database
func handleList(ctx context.Context) {
	appid := util.GetAppID(ctx)

	list, errCode, err := GetSwitches(appid)
	if errCode != ApiError.SUCCESS {
		ctx.JSON(util.GenRetObj(errCode, err))
	} else {
		ctx.JSON(util.GenRetObj(errCode, list))
	}
}

func handleSwitch(ctx context.Context) {
	id, _ := ctx.Params().GetInt("id")
	appid := util.GetAppID(ctx)

	ret, errCode, err := GetSwitch(appid, id)
	if errCode != ApiError.SUCCESS {
		ctx.JSON(util.GenRetObj(errCode, err))
	} else {
		ctx.JSON(util.GenRetObj(errCode, ret))
	}
}

func handleNewSwitch(ctx context.Context) {
	appid := util.GetAppID(ctx)

	input := loadSwitchFromContext(ctx)
	if input == nil {
		ctx.StatusCode(iris.StatusBadRequest)
		return
	}

	errCode, err := InsertSwitch(appid, input)
	errMsg := ApiError.GetErrorMsg(errCode)
	if errCode != ApiError.SUCCESS {
		ctx.JSON(util.GenRetObj(errCode, err))
		addAudit(ctx, util.AuditOperationAdd, fmt.Sprintf("Add fail: %s (%s)", errMsg, err.Error()), 0)
	} else {
		ctx.JSON(util.GenRetObj(errCode, input))
		addAudit(ctx, util.AuditOperationAdd, fmt.Sprintf("Add success %#v", input), 1)
	}
}

func diffSwitchInfo(switchA *SwitchInfo, switchB *SwitchInfo) string {
	var buf bytes.Buffer
	fields := []string{"Name", "Status", "Remark", "Scenario", "Num", "Msg"}

	reflectA := reflect.ValueOf(switchA)
	reflectB := reflect.ValueOf(switchB)
	for _, field := range fields {
		valA := reflect.Indirect(reflectA).FieldByName(field).String()
		valB := reflect.Indirect(reflectB).FieldByName(field).String()
		if valA != valB {
			buf.WriteString(fmt.Sprintf("%s [%v]->[%v]; ", util.Msg[field], valA, valB))
		}
	}
	return buf.String()
}

func handleUpdateSwitch(ctx context.Context) {
	id, _ := ctx.Params().GetInt("id")
	appid := util.GetAppID(ctx)

	input := loadSwitchFromContext(ctx)
	if input == nil {
		ctx.StatusCode(iris.StatusBadRequest)
		return
	}

	orig, errCode, err := GetSwitch(appid, id)
	errMsg := ApiError.GetErrorMsg(errCode)
	if errCode != ApiError.SUCCESS {
		ctx.JSON(util.GenRetObj(errCode, err))
		return
	}

	errCode, err = UpdateSwitch(appid, id, input)
	errMsg = ApiError.GetErrorMsg(errCode)
	if errCode != ApiError.SUCCESS {
		ctx.JSON(util.GenRetObj(errCode, err))
		addAudit(ctx, util.AuditOperationEdit, fmt.Sprintf("%s%s code[%s]: %s (%s)",
			util.Msg["Modify"], util.Msg["Error"], input.Code, errMsg, err.Error()), 0)
	} else {
		ctx.JSON(util.GenRetObj(errCode, input))
		diffMsg := diffSwitchInfo(orig, input)
		var msg string

		if diffMsg != "" {
			msg = fmt.Sprintf(
				"%s%s code[%s]\n %s",
				util.Msg["Modify"], util.Msg["Success"], input.Code, diffMsg)
		} else {
			msg = fmt.Sprintf(
				"%s%s code[%s]",
				util.Msg["Modify"], util.Msg["Success"], input.Code)
		}
		addAudit(ctx, util.AuditOperationEdit, msg, 1)
	}

	var ret int
	if orig.Code == "task_engine" {
		ret, err = util.ConsulUpdateTaskEngine(appid, input.Status == 1)
	} else {
		ret, err = util.ConsulUpdateRobotChat(appid)
	}
	if err != nil {
		util.LogInfo.Printf("Update consul result: %d, %s", ret, err.Error())
	} else {
		util.LogInfo.Printf("Update consul result: %d", ret)
	}
}

func handleDeleteSwitch(ctx context.Context) {
	id, _ := ctx.Params().GetInt("id")
	appid := util.GetAppID(ctx)

	errCode, err := DeleteSwitch(appid, id)
	errMsg := ApiError.GetErrorMsg(errCode)
	if errCode != ApiError.SUCCESS {
		ctx.JSON(util.GenRetObj(errCode, err))
		addAudit(ctx, util.AuditOperationDelete, fmt.Sprintf("Delete id %d fail: %s (%s)", id, errMsg, err.Error()), 0)
	} else {
		ctx.JSON(util.GenRetObj(errCode, nil))
		addAudit(ctx, util.AuditOperationDelete, fmt.Sprintf("Delete id %d success", id), 1)
	}
}

func loadSwitchFromContext(ctx context.Context) *SwitchInfo {
	input := &SwitchInfo{}
	err := ctx.ReadJSON(input)
	if err != nil {
		util.LogInfo.Printf("Bad request when loading from input: %s", err.Error())
		return nil
	}
	input.UpdateTime = time.Now()
	return input
}

func addAudit(ctx context.Context, operation string, msg string, result int) {
	userID := util.GetUserID(ctx)
	userIP := util.GetUserIP(ctx)

	util.AddAuditLog(userID, userIP, util.AuditModuleSwitchList, operation, msg, result)
}
