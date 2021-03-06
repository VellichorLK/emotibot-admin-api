package Switch

import (
	"bytes"
	"fmt"
	"net/http"
	"reflect"
	"time"

	"emotibot.com/emotigo/module/admin-api/ApiError"
	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/admin-api/util/audit"
	"emotibot.com/emotigo/module/admin-api/util/requestheader"
	"emotibot.com/emotigo/pkg/logger"
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
			util.NewEntryPoint("GET", "switch/{id}", []string{"view"}, handleSwitch),
			util.NewEntryPoint("POST", "switch/{id}", []string{"edit"}, handleUpdateSwitch),
			util.NewEntryPoint("DELETE", "switch/{id}", []string{"delete"}, handleDeleteSwitch),
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
func handleList(w http.ResponseWriter, r *http.Request) {
	appid := requestheader.GetAppID(r)

	list, errCode, err := GetSwitches(appid)
	if errCode != ApiError.SUCCESS {
		util.WriteJSON(w, util.GenRetObj(errCode, err))
	} else {
		util.WriteJSON(w, util.GenRetObj(errCode, list))
	}
}

func handleSwitch(w http.ResponseWriter, r *http.Request) {
	id, _ := util.GetMuxIntVar(r, "id")
	appid := requestheader.GetAppID(r)

	ret, errCode, err := GetSwitch(appid, id)
	if errCode != ApiError.SUCCESS {
		util.WriteJSON(w, util.GenRetObj(errCode, err))
	} else {
		util.WriteJSON(w, util.GenRetObj(errCode, ret))
	}
}

func handleNewSwitch(w http.ResponseWriter, r *http.Request) {
	appid := requestheader.GetAppID(r)

	input := loadSwitchFromContext(w, r)
	if input == nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	errCode, err := InsertSwitch(appid, input)
	errMsg := ApiError.GetErrorMsg(errCode)
	if errCode != ApiError.SUCCESS {
		util.WriteJSON(w, util.GenRetObj(errCode, err))
		addAudit(r, audit.AuditOperationAdd, fmt.Sprintf("Add fail: %s (%s)", errMsg, err.Error()), 0)
	} else {
		util.WriteJSON(w, util.GenRetObj(errCode, input))
		addAudit(r, audit.AuditOperationAdd, fmt.Sprintf("Add success %#v", input), 1)
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

func handleUpdateSwitch(w http.ResponseWriter, r *http.Request) {
	id, _ := util.GetMuxIntVar(r, "id")
	appid := requestheader.GetAppID(r)

	input := loadSwitchFromContext(w, r)
	if input == nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	orig, errCode, err := GetSwitch(appid, id)
	errMsg := ApiError.GetErrorMsg(errCode)
	if errCode != ApiError.SUCCESS {
		util.WriteJSON(w, util.GenRetObj(errCode, err))
		return
	}

	errCode, err = UpdateSwitch(appid, id, input)
	errMsg = ApiError.GetErrorMsg(errCode)
	if errCode != ApiError.SUCCESS {
		util.WriteJSON(w, util.GenRetObj(errCode, err))
		addAudit(r, audit.AuditOperationEdit, fmt.Sprintf("%s%s code[%s]: %s (%s)",
			util.Msg["Modify"], util.Msg["Error"], input.Code, errMsg, err.Error()), 0)
	} else {
		util.WriteJSON(w, util.GenRetObj(errCode, input))
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
		addAudit(r, audit.AuditOperationEdit, msg, 1)
	}

	var ret int
	if orig.Code == "task_engine" {
		ret, err = util.ConsulUpdateTaskEngine(appid, input.Status == 1)
	} else {
		ret, err = util.ConsulUpdateRobotChat(appid)
	}
	if err != nil {
		logger.Info.Printf("Update consul result: %d, %s", ret, err.Error())
	} else {
		logger.Info.Printf("Update consul result: %d", ret)
	}
}

func handleDeleteSwitch(w http.ResponseWriter, r *http.Request) {
	id, _ := util.GetMuxIntVar(r, "id")
	appid := requestheader.GetAppID(r)

	errCode, err := DeleteSwitch(appid, id)
	errMsg := ApiError.GetErrorMsg(errCode)
	if errCode != ApiError.SUCCESS {
		util.WriteJSON(w, util.GenRetObj(errCode, err))
		addAudit(r, audit.AuditOperationDelete, fmt.Sprintf("Delete id %d fail: %s (%s)", id, errMsg, err.Error()), 0)
	} else {
		util.WriteJSON(w, util.GenRetObj(errCode, nil))
		addAudit(r, audit.AuditOperationDelete, fmt.Sprintf("Delete id %d success", id), 1)
	}
}

func loadSwitchFromContext(w http.ResponseWriter, r *http.Request) *SwitchInfo {
	input := &SwitchInfo{}
	err := util.ReadJSON(r, input)
	if err != nil {
		logger.Info.Printf("Bad request when loading from input: %s", err.Error())
		return nil
	}
	input.UpdateTime = time.Now()
	return input
}

func addAudit(r *http.Request, operation string, msg string, result int) {
	userID := requestheader.GetUserID(r)
	userIP := requestheader.GetUserIP(r)
	appid := requestheader.GetAppID(r)
	enterpriseID := requestheader.GetEnterpriseID(r)

	audit.AddAuditLog(enterpriseID, appid, userID, userIP, audit.AuditModuleRobotFunction, operation, msg, result)
}
