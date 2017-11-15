package UI

import (
	"github.com/kataras/iris/context"

	"emotibot.com/emotigo/module/vipshop-admin/ApiError"
	"emotibot.com/emotigo/module/vipshop-admin/util"
)

var (
	// ModuleInfo is needed for module define
	ModuleInfo util.ModuleInfo
)

func init() {
	ModuleInfo = util.ModuleInfo{
		ModuleName: "ui",
		EntryPoints: []util.EntryPoint{
			util.NewEntryPoint("GET", "envs", []string{}, handleDumpUISetting),
		},
	}
}

func handleDumpUISetting(ctx context.Context) {
	envs := getEnvironments()
	ctx.JSON(util.GenRetObj(ApiError.SUCCESS, envs))
}

func getEnvironments() map[string]string {
	return util.GetEnvOf(ModuleInfo.ModuleName)
}
