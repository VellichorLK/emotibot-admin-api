package Robot

import (
	"github.com/kataras/iris/context"

	"emotibot.com/emotigo/module/vipshop-admin/util"
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
			util.NewEntryPoint("POST", "function/{name:string}", []string{"edit"}, handleUpdateFunction),

			util.NewEntryPoint("GET", "robotqas", []string{"view"}, handleRobotQAList),
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

func addAudit(ctx context.Context, operation string, msg string, result int) {
	userID := util.GetUserID(ctx)
	userIP := util.GetUserIP(ctx)

	util.AddAuditLog(userID, userIP, util.AuditModuleSwitchList, operation, msg, result)
}
