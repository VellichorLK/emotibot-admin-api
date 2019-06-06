package lqcheck

import (
	robotConfig "emotibot.com/emotigo/module/admin-api/Robot/config.v1"
	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/admin-api/util/requestheader"
	"net/http"
)

var (
	ModuleInfo util.ModuleInfo
)

func init() {
	ModuleInfo = util.ModuleInfo{
		ModuleName: "lq_health",
		EntryPoints: []util.EntryPoint{
			// 获取报告
			util.NewEntryPointWithVer("GET", "report/{appid}", []string{}, handleGetLqCheckReport, 1),
			// 生成报告
			util.NewEntryPointWithVer("POST", "report/{appid}/{locale}", []string{}, handleCreateLqCheckReport, 1),
			// 获取报告状态
			util.NewEntryPointWithVer("GET", "report/status/{appid}", []string{}, handleHealthCheckStatus, 1),
			// 获取标准问语料
			util.NewEntryPointWithVer("GET", "report/sqlq/{appid}", []string{}, handleGetReportSqLq, 1),
			// 获取标准问语料总数
			util.NewEntryPointWithVer("GET", "report/sqlq_count/{appid}", []string{}, handleGetReportSqLqCount, 1),
		},
	}
}

// 获取健康检查报告
func handleGetLqCheckReport(w http.ResponseWriter, r *http.Request) {
	appid := util.GetMuxVar(r, "appid")

	data, err := getHealCheckReport(appid)

	util.Return(w, err, data)
}

// 创建健康检查报告，获取任务id
func handleCreateLqCheckReport(w http.ResponseWriter, r *http.Request) {
	appid := util.GetMuxVar(r, "appid")
	locale := util.GetMuxVar(r, "locale")
	outerUrl := getOuterUrl(r, appid)
	data, err := createReport(appid, locale, outerUrl)

	util.Return(w, err, data)
}

// 获取健康检查状态
func handleHealthCheckStatus(w http.ResponseWriter, r *http.Request) {
	appid := util.GetMuxVar(r, "appid")

	data, err := getHealthCheckResult(appid)

	util.Return(w, err, data)
}

// 获取标准问语料
func handleGetReportSqLq(w http.ResponseWriter, r *http.Request) {
	appid := util.GetMuxVar(r, "appid")

	data, err := getSqLq(appid)

	util.Return(w, err, data)
}

// 获取标准问语料总数
func handleGetReportSqLqCount(w http.ResponseWriter, r *http.Request) {
	appid := util.GetMuxVar(r, "appid")

	data, err := getSqLqCount(appid)

	util.Return(w, err, data)
}

func getOuterUrl(r *http.Request, appid string) string {
	server := requestheader.GetOrigin(r)

	config, err := robotConfig.GetConfig(appid, "uploadimg_server")
	if err == nil && config != nil && config.Value != "" {
		server = config.Value
	}

	return server
}
