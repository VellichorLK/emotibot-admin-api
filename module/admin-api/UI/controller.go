package UI

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

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

const (
	// MaxIconSize is the max image size can be update to server
	MaxIconSize = 128 * 1024
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

			util.NewEntryPoint("GET", "logo", []string{}, handleGetLogo),
			util.NewEntryPointWithConfig("PUT", "logo", []string{"edit"}, handleUploadLogo, util.EntryConfig{
				Version:     1,
				IgnoreAppID: true,
			}),
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
	userID := requestheader.GetUserID(r)
	userIP := requestheader.GetUserIP(r)
	appid := requestheader.GetAppID(r)
	enterprise := requestheader.GetEnterpriseID(r)

	moduleID := ""
	switch module {
	case "qalist":
		moduleID = audit.AuditModuleSSM // = "2" // "问答库"
		break
	case "dictionary":
		moduleID = audit.AuditModuleWordbank // = "5" // "词库管理"
		break
	case "statistic-analysis":
		moduleID = audit.AuditModuleStatisticAnalysis // = "6" // "数据管理"
		break
	case "statistic-daily":
		moduleID = audit.AuditModuleStatisticDaily // = "6" // "数据管理"
		break
	case "statistic-audit":
		moduleID = audit.AuditModuleAudit // = "6" // "数据管理"
		break
	case "task-engine":
		moduleID = audit.AuditModuleTaskEngine // 任務引擎
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
	err := audit.AddAuditLog(enterprise, appid, userID, userIP, moduleID, audit.AuditOperationExport, log, 1)
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

func handleGetLogo(w http.ResponseWriter, r *http.Request) {
	enterprise := r.URL.Query().Get("enterprise")
	iconType := r.URL.Query().Get("type")
	if iconType == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Get icon of enterprise first, if fail, get system icon
	data, err := util.ConsulGetLogo(enterprise, iconType)
	if err != nil {
		data, err = util.ConsulGetLogo("", iconType)
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	if len(data) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	contentType := http.DetectContentType(data)
	if strings.Index(contentType, "text/xml") >= 0 {
		w.Header().Set("Content-Type", "image/svg+xml")
	} else {
		w.Header().Set("Content-Type", contentType)
	}
	w.Write(data)
}

func handleUploadLogo(w http.ResponseWriter, r *http.Request) {
	enterprise := r.FormValue("enterprise")
	iconType := r.FormValue("type")
	if iconType == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	file, info, err := r.FormFile("file")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	if info.Size > MaxIconSize {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("Max file size: %d, get %d", MaxIconSize, info.Size)))
		return
	}

	content, err := ioutil.ReadAll(file)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	err = util.ConsulUpdateLogo(enterprise, iconType, content)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	util.Return(w, nil, nil)
}
