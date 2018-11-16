package Switch

import (
	"bytes"
	"fmt"
	"reflect"
	"time"

	"emotibot.com/emotigo/module/vipshop-admin/ApiError"
	"emotibot.com/emotigo/module/vipshop-admin/util"
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
	lastOperation := time.Now()

	list, errCode, err := GetSwitches(appid)
	util.LogInfo.Printf("get switch list in handleList took: %s", time.Since(lastOperation))
	if errCode != ApiError.SUCCESS {
		ctx.JSON(util.GenRetObj(errCode, err))
	} else {
		ctx.JSON(util.GenRetObj(errCode, list))
	}
}

func handleSwitch(ctx context.Context) {
	id, _ := ctx.Params().GetInt("id")
	appid := util.GetAppID(ctx)
	lastOperation := time.Now()

	ret, errCode, err := GetSwitch(appid, id)
	util.LogInfo.Printf("get switch in handleSwitch took: %s", time.Since(lastOperation))
	if errCode != ApiError.SUCCESS {
		ctx.JSON(util.GenRetObj(errCode, err))
	} else {
		ctx.JSON(util.GenRetObj(errCode, ret))
	}
}

func handleNewSwitch(ctx context.Context) {
	appid := util.GetAppID(ctx)
	lastOperation := time.Now()

	input := loadSwitchFromContext(ctx)
	if input == nil {
		ctx.StatusCode(iris.StatusBadRequest)
		return
	}

	errCode, err := InsertSwitch(appid, input)
	util.LogInfo.Printf("create swtich in handleNewSwitch took: %s", time.Since(lastOperation))
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
	lastOperation := time.Now()

	input := loadSwitchFromContext(ctx)
	if input == nil {
		ctx.StatusCode(iris.StatusBadRequest)
		return
	}

	orig, errCode, err := GetSwitch(appid, id)
	util.LogInfo.Printf("get swtich in handleUpdateSwitch took: %s", time.Since(lastOperation))
	lastOperation = time.Now()
	//errMsg := ApiError.GetErrorMsg(errCode)
	if errCode != ApiError.SUCCESS {
		ctx.JSON(util.GenRetObj(errCode, err))
		return
	}

	errCode, err = UpdateSwitch(appid, id, input)
	util.LogInfo.Printf("update swtich in handleUpdateSwitch took: %s", time.Since(lastOperation))
	lastOperation = time.Now()

	//errMsg = ApiError.GetErrorMsg(errCode)
	if errCode != ApiError.SUCCESS {
		ctx.JSON(util.GenRetObj(errCode, err))
		/**addAudit(ctx, util.AuditOperationEdit, fmt.Sprintf("%s%s code[%s]: %s (%s)",
			util.Msg["Modify"], util.Msg["Error"], input.Code, errMsg, err.Error()), 0)*/
			addAudit(ctx, util.AuditOperationEdit, safeLog(input,orig), 0)
	} else {
		ctx.JSON(util.GenRetObj(errCode, input))
		/**diffMsg := diffSwitchInfo(orig, input)
		var msg string

		if diffMsg != "" {
			msg = fmt.Sprintf(
				"%s%s code[%s]\n %s",
				util.Msg["Modify"], util.Msg["Success"], input.Code, diffMsg)
		} else {
			msg = fmt.Sprintf(
				"%s%s code[%s]",
				util.Msg["Modify"], util.Msg["Success"], input.Code)
		}*/

		//addAudit(ctx, util.AuditOperationEdit, msg, 1)

		updateConsul()
		util.LogInfo.Printf("update consul in handleUpdateSwitch took: %s", time.Since(lastOperation))
		lastOperation = time.Now()
		addAudit(ctx, util.AuditOperationEdit, safeLog(input,orig), 1)
	}

	var ret int
	if orig.Code == "task_engine" {
		ret, err = util.ConsulUpdateTaskEngine(appid, input.Status == 1)
	} else {
		ret, err = util.ConsulUpdateRobotChat(appid)
	}
	util.LogInfo.Printf("update consul by original code in handleUpdateSwitch took: %s", time.Since(lastOperation))

	if err != nil {
		util.LogInfo.Printf("Update consul result: %d, %s", ret, err.Error())
	} else {
		util.LogInfo.Printf("Update consul result: %d", ret)
	}
}

func handleDeleteSwitch(ctx context.Context) {
	id, _ := ctx.Params().GetInt("id")
	appid := util.GetAppID(ctx)
	lastOperation := time.Now()

	errCode, err := DeleteSwitch(appid, id)
	util.LogInfo.Printf("delete switch in handleDeleteSwitch took: %s", time.Since(lastOperation))

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

func safeLog(input *SwitchInfo, dbInfo *SwitchInfo) (string){
	var title = ""
	var status = ""
	if (dbInfo.Status != input.Status) {
		var dbMsg = "开"
		var msg = "开"
		if(dbInfo.Status != 1) {
			dbMsg = "关"
		}
		if(input.Status != 1) {
			msg = "关"
		}

		status = fmt.Sprintf("[开关状态]%s => %s", dbMsg, msg)
	}

	var num = ""
	if (dbInfo.Num != input.Num) {
		var s = "；"
		if(status == "") {
			s = ""
		}
		num = fmt.Sprintf(s+"[次数设置]%d => %d", dbInfo.Num, input.Num)
	}
	var msg = ""
	if (dbInfo.Msg != input.Msg) {
		var s = "；"
		if(status == "" && num == "") {
			s = ""
		}
		msg = fmt.Sprintf(s+"[话术设置]%s => %s", dbInfo.Msg, input.Msg)
	}

	switch input.ID {
		case 1:
			title = "[未解决转人工]："+status+num+msg

		case 2:
			title = "[场景转人工]："+status+num+msg
		case 3:
			title = "[TE开关]："+status+num+msg
		case 4:
			title = "[未匹配转人工]："+status+num+msg
		case 5:
			title = "[TE白名单开关]："+status+num+msg
		default:
			title = ""
	}

	return title
	
}

func updateConsul() {
	unixTime := time.Now().UnixNano() / 1000000
	_, err := util.ConsulUpdateVal("te/gray", unixTime)
	if err != nil {
		util.LogError.Println("consul update failed, %v", err)
	}
}
