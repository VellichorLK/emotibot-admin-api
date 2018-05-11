package UI

import (
	"fmt"
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
		ModuleName: "ui",
		EntryPoints: []util.EntryPoint{
			util.NewEntryPoint("GET", "envs", []string{}, handleDumpUISetting),
			util.NewEntryPoint("POST", "export-log", []string{}, handleExportAuditLog),

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
	util.LogTrace.Println("Run: handleDumpUISetting")
	envs := getEnvironments()
	util.WriteJSON(w, util.GenRetObj(ApiError.SUCCESS, envs))
	return
}

func handleExportAuditLog(w http.ResponseWriter, r *http.Request) {
	util.LogTrace.Println("Run: handleExportAuditLog")
	module := r.FormValue("module")
	fileName := r.FormValue("filename")
	extMsg := r.FormValue("info")
	userID := util.GetUserID(r)
	userIP := util.GetUserIP(r)

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
	}

	if moduleID == "" || fileName == "" {
		http.Error(w, "", http.StatusBadRequest)
		util.LogInfo.Printf("Bad request: module:[%s] file:[%s]", moduleID, fileName)
		return
	}

	moduleName := util.ModuleName[module]
	log := fmt.Sprintf("%s%s %s: %s", util.Msg["DownloadFile"], moduleName, fileName, extMsg)
	err := util.AddAuditLog(userID, userIP, moduleID, util.AuditOperationExport, log, 1)
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
