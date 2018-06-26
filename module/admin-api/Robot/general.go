package Robot

import (
	"net/http"

	"emotibot.com/emotigo/module/admin-api/ApiError"
	"emotibot.com/emotigo/module/admin-api/util"
)

var (
	// ModuleInfo is needed for module define
	ModuleInfo util.ModuleInfo
)

func init() {
	ModuleInfo = util.ModuleInfo{
		ModuleName: "robot",
		EntryPoints: []util.EntryPoint{
			util.NewEntryPoint("GET", "functions", []string{"view"}, handleDBFunctionList),
			util.NewEntryPoint("POST", "functions", []string{"edit"}, handleUpdateAllDBFunction),
			util.NewEntryPoint("POST", "function/{name}", []string{"edit"}, handleUpdateDBFunction),
			util.NewEntryPointWithVer("GET", "functions", []string{"view"}, handleDBFunctionListV2, 2),
			util.NewEntryPointWithVer("POST", "functions", []string{"edit"}, handleUpdateAllDBFunctionV2, 2),
			util.NewEntryPointWithVer("POST", "function/{name}", []string{"edit"}, handleUpdateDBFunctionV2, 2),

			util.NewEntryPoint("GET", "qas", []string{"view"}, handleRobotQAList),
			util.NewEntryPoint("POST", "qabuild", []string{"edit"}, handleRobotQAModelRebuild),
			util.NewEntryPoint("GET", "qa/{id}", []string{"view"}, handleRobotQA),
			util.NewEntryPoint("POST", "qa/{id}", []string{"edit"}, handleUpdateRobotQA),
			util.NewEntryPointWithVer("GET", "qas", []string{"view"}, handleRobotQAListV2, 2),
			util.NewEntryPointWithVer("GET", "qa/{id}", []string{"view"}, handleRobotQAV2, 2),
			util.NewEntryPointWithVer("POST", "qa/{id}", []string{"edit"}, handleUpdateRobotQAV2, 2),

			util.NewEntryPoint("GET", "chats", []string{"view"}, handleChatList),
			util.NewEntryPoint("GET", "chat/{id}", []string{"view"}, handleGetChat),
			util.NewEntryPoint("POST", "chats", []string{"edit"}, handleMultiChatModify),
			util.NewEntryPoint("GET", "chat-info", []string{"view"}, handleChatInfoList),

			util.NewEntryPoint("GET", "chatQAList", []string{"view"}, handleChatQAList),

			util.NewEntryPointWithVer("GET", "chats", []string{"view"}, handleGetRobotWords, 2),
			util.NewEntryPointWithVer("GET", "chat/{id}", []string{"view"}, handleGetRobotWord, 2),
			util.NewEntryPointWithVer("PUT", "chat/{id}", []string{"edit"}, handleUpdateRobotWord, 2),
			util.NewEntryPointWithVer("POST", "chat/{id}/content", []string{"edit"}, handleAddRobotWordContent, 2),
			util.NewEntryPointWithVer("PUT", "chat/{id}/content/{cid}", []string{"edit"}, handleUpdateRobotWordContent, 2),
			util.NewEntryPointWithVer("DELETE", "chat/{id}/content/{cid}", []string{"delete"}, handleDeleteRobotWordContent, 2),

			util.NewEntryPointWithCustom("POST", "data", []string{"edit"}, handleInitRobotData, 2, false),
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

func addAudit(r *http.Request, module string, operation string, msg string, result int) {
	userID := util.GetUserID(r)
	userIP := util.GetUserIP(r)
	appid := util.GetAppID(r)

	util.AddAuditLog(appid, userID, userIP, module, operation, msg, result)
}

func handleInitRobotData(w http.ResponseWriter, r *http.Request) {
	appid := r.FormValue("appid")

	errRobot := InitRobotFunction(appid)
	errQA := InitRobotQAData(appid)
	if errRobot != nil {
		util.WriteJSON(w, util.GenRetObj(ApiError.DB_ERROR, errRobot.Error()))
		return
	}
	if errQA != nil {
		util.WriteJSON(w, util.GenRetObj(ApiError.DB_ERROR, errQA.Error()))
		return
	}
	// after init data, update consul to notify controller to init data
	go util.ConsulUpdateRobotChat(appid)
	go util.ConsulUpdateFunctionStatus(appid)

	// call multicustomer to handle robot QA
	go util.McRebuildRobotQA(appid)

	util.WriteJSON(w, util.GenRetObj(ApiError.SUCCESS, nil))
}
