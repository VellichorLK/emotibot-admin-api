package lqcheck

import (
	"emotibot.com/emotigo/module/admin-api/util"
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
			util.NewEntryPointWithVer("POST", "report/{appid}", []string{}, handleCreateLqCheckReport, 1),
			// 获取报告状态
			util.NewEntryPointWithVer("GET", "report/status/{taskid}", []string{}, handleHealthCheckStatus, 1),
			// 获取标准问语料
			util.NewEntryPointWithVer("GET", "report/sqlq/{appid}", []string{}, handleGetReportSqLq, 1),
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

	data, err := createReport(appid)

	util.Return(w, err, data)
}

// 获取健康检查状态
func handleHealthCheckStatus(w http.ResponseWriter, r *http.Request) {
	taskid := util.GetMuxVar(r, "taskid")

	data, err := getHealthCheckStatus(taskid)

	util.Return(w, err, data)
}

// 获取标准问语料
func handleGetReportSqLq(w http.ResponseWriter, r *http.Request) {
	appid := util.GetMuxVar(r, "appid")

	data, err := getSqLq(appid)

	util.Return(w, err, data)
}
