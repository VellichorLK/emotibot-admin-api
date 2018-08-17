package UI

import (
	"fmt"
	"net/http"

	"emotibot.com/emotigo/module/admin-api/ApiError"
	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/pkg/logger"
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
			util.NewEntryPoint("GET", "versions", []string{}, handleDumpVersionInfo),

			util.NewEntryPoint("GET", "encrypt", []string{}, handleEncrypt),
			util.NewEntryPoint("GET", "decrypt", []string{}, handleDecrypt),
		},
	}
}

func handleEncrypt(w http.ResponseWriter, r *http.Request) {
	text := r.URL.Query().Get("text")
	if text == "" {
		util.WriteWithStatus(w, text, http.StatusBadRequest)
		return
	}

	encrypt, err := util.DesEncrypt([]byte(text), []byte(util.DesEncryptKey))
	if err != nil {
		util.WriteWithStatus(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte(encrypt))
}

func handleDecrypt(w http.ResponseWriter, r *http.Request) {
	text := r.URL.Query().Get("text")
	if text == "" {
		util.WriteWithStatus(w, text, http.StatusBadRequest)
		return
	}

	decrypt, err := util.DesDecrypt(text, []byte(util.DesEncryptKey))
	if err != nil {
		util.WriteWithStatus(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte(decrypt))
}

func handleDumpUISetting(w http.ResponseWriter, r *http.Request) {
	logger.Trace.Println("Run: handleDumpUISetting")
	envs := getEnvironments()
	util.WriteJSON(w, util.GenRetObj(ApiError.SUCCESS, envs))
	return
}

func handleExportAuditLog(w http.ResponseWriter, r *http.Request) {
	logger.Trace.Println("Run: handleExportAuditLog")
	module := r.FormValue("module")
	fileName := r.FormValue("filename")
	extMsg := r.FormValue("info")
	userID := util.GetUserID(r)
	userIP := util.GetUserIP(r)
	appid := util.GetAppID(r)

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
	case "statistic-audit":
		moduleID = util.AuditModuleStatistics // = "6" // "数据管理"
		break
	case "task-engine":
		moduleID = util.AuditModuleTaskEngine // 任務引擎
		break
	}

	if moduleID == "" || fileName == "" {
		http.Error(w, "", http.StatusBadRequest)
		logger.Info.Printf("Bad request: module:[%s] file:[%s]", moduleID, fileName)
		return
	}

	moduleName := util.ModuleName[module]
	log := fmt.Sprintf("%s%s %s", util.Msg["DownloadFile"], moduleName, fileName)
	if extMsg != "" {
		log = fmt.Sprintf("%s: %s", log, extMsg)
	}
	err := util.AddAuditLog(appid, userID, userIP, moduleID, util.AuditOperationExport, log, 1)
	if err != nil {
		util.WriteJSON(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()))
	} else {
		util.WriteJSON(w, util.GenSimpleRetObj(ApiError.SUCCESS))
	}
	return
}

func getEnvironments() map[string]string {
	return util.GetEnvOf(ModuleInfo.ModuleName)
}

func handleDumpVersionInfo(w http.ResponseWriter, r *http.Request) {
	content, errno, err := util.ConsulGetReleaseSetting()
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(errno, err.Error()), ApiError.GetHttpStatus(errno))
		return
	}
	util.WriteJSON(w, util.GenRetObj(ApiError.SUCCESS, content))
}
