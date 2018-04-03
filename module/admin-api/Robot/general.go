package Robot

import (
	"net/http"

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
			util.NewEntryPoint("GET", "functions", []string{"view"}, handleFunctionList),
			util.NewEntryPoint("POST", "functions", []string{"edit"}, handleUpdateAllFunction),
			util.NewEntryPoint("POST", "function/{name}", []string{"edit"}, handleUpdateFunction),

			util.NewEntryPoint("GET", "qas", []string{"view"}, handleRobotQAList),
			util.NewEntryPoint("POST", "qabuild", []string{"edit"}, handleRobotQAModelRebuild),
			util.NewEntryPoint("GET", "qa/{id}", []string{"view"}, handleRobotQA),
			util.NewEntryPoint("POST", "qa/{id}", []string{"edit"}, handleUpdateRobotQA),

			util.NewEntryPoint("GET", "chats", []string{"view"}, handleChatList),
			util.NewEntryPoint("GET", "chat/{id}", []string{"view"}, handleGetChat),
			util.NewEntryPoint("POST", "chats", []string{"edit"}, handleMultiChatModify),
			util.NewEntryPoint("GET", "chat-info", []string{"view"}, handleChatInfoList),
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

	util.AddAuditLog(userID, userIP, module, operation, msg, result)
}
