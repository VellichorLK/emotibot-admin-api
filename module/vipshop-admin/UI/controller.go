package UI

import (
	"fmt"

	"github.com/kataras/iris"

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

			util.NewEntryPoint("POST", "export-log", []string{}, handleExportAuditLog),
		},
	}
}

func handleDumpUISetting(ctx context.Context) {
	envs := getEnvironments()
	ctx.JSON(util.GenRetObj(ApiError.SUCCESS, envs))
}

func handleExportAuditLog(ctx context.Context) {
	module := ctx.FormValue("module")
	fileName := ctx.FormValue("filename")
	userID := util.GetUserID(ctx)
	userIP := util.GetUserIP(ctx)

	moduleID := ""
	switch module {
	case "qalist":
		moduleID = util.AuditModuleQA // = "2" // "问答库"
		break
	case "dictionary":
		moduleID = util.AuditModuleDictionary // = "5" // "词库管理"
		break
	case "statistic-analysis":
		moduleID = util.AuditModuleStatistics // = "6" // "数据管理"
		break
	case "statistic-daily":
		moduleID = util.AuditModuleStatistics // = "6" // "数据管理"
		break
	}

	if moduleID == "" || fileName == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		util.LogInfo.Printf("Bad request: module:[%s] file:[%s]", moduleID, fileName)
		return
	}

	moduleName := util.ModuleName[module]
	log := fmt.Sprintf("%s%s %s", util.Msg["DownloadFile"], moduleName, fileName)
	err := util.AddAuditLog(userID, userIP, moduleID, util.AuditOperationExport, log, 1)
	if err != nil {
		ctx.JSON(util.GenRetObj(ApiError.DB_ERROR, err.Error()))
	} else {
		ctx.JSON(util.GenSimpleRetObj(ApiError.SUCCESS))
	}
}

func getEnvironments() map[string]string {
	return util.GetEnvOf(ModuleInfo.ModuleName)
}
