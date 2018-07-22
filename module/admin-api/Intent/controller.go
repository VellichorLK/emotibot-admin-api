package Intent

import (
	"fmt"
	"net/http"
	"strconv"

	"emotibot.com/emotigo/module/admin-api/ApiError"
	"emotibot.com/emotigo/module/admin-api/util"
	"github.com/siongui/gojianfan"
)

var (
	// ModuleInfo is needed for module define
	ModuleInfo util.ModuleInfo
)

const (
	defaultIntentEngineURL = "http://127.0.0.1:15001"
	defaultRuleEngineURL   = "http://127.0.0.1:15002"
)

const (
	NotTrained  = "NOT_TRAINED"
	Training    = "TRAINING"
	Trained     = "TRAINED"
	TrainFailed = "TRAIN_FAILED"
)

func init() {
	ModuleInfo = util.ModuleInfo{
		ModuleName: "intents",
		EntryPoints: []util.EntryPoint{
			util.NewEntryPoint("GET", "", []string{"view"}, handleGetIntents),
			util.NewEntryPoint("POST", "upload", []string{"view"}, handleUploadIntents),
			util.NewEntryPoint("GET", "download", []string{}, handleDownloadIntents),
			util.NewEntryPoint("POST", "train", []string{"view"}, handleTrain),
			util.NewEntryPoint("GET", "status", []string{"view"}, handleGetTrainStatus),
			util.NewEntryPoint("GET", "getData", []string{}, handleGetData),

			// util.NewEntryPoint("POST", "intent", []string{"create"}, handleAddIntent),
			// util.NewEntryPoint("GET", "intent/{id}", []string{"view"}, handleGetIntent),
			// util.NewEntryPoint("PUT", "intent/{id}", []string{"view"}, handleUpdateIntent),
			// util.NewEntryPoint("DELETE", "intent/{id}", []string{"view"}, handleDeleteIntent),

			// util.NewEntryPoint("GET", "intent/{id}/sentence/{sid}", []string{"view"}, handleGetIntentSentence),
			// util.NewEntryPoint("POST", "intent/{id}/sentence", []string{"view"}, handleUpdateIntentSentence),
			// util.NewEntryPoint("PUT", "intent/{id}/sentence/{sid}", []string{"view"}, handleUpdateIntentSentence),
			// util.NewEntryPoint("DELETE", "intent/{id}/sentence/{sid}", []string{"view"}, handleDeleteIntentSentence),
		},
	}
}

func handleGetIntents(w http.ResponseWriter, r *http.Request) {
	appID := util.GetAppID(r)
	v := r.URL.Query().Get("version")
	zhTW := r.URL.Query().Get("zh_tw")

	var version int
	if v == "" {
		version = 0
	} else {
		ver, err := strconv.Atoi(v)
		if err != nil {
			http.Error(w, "Invalid intent dataset version", http.StatusBadRequest)
			return
		}

		version = ver
	}

	intents, retCode, err := GetIntents(appID, version)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if retCode == ApiError.NOT_FOUND_ERROR {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if zhTW != "" {
		tradChinese, err := strconv.ParseBool(zhTW)
		if err == nil && tradChinese {
			// Convert intents back to Traditional Chinese
			for i, intent := range intents {
				intents[i] = gojianfan.S2T(intent)
			}
		}
	}

	util.WriteJSON(w, util.GenRetObj(ApiError.SUCCESS, intents))
}

func handleUploadIntents(w http.ResponseWriter, r *http.Request) {
	auditMsg := util.Msg["UploadIntentEngine"]
	var err error
	defer func() {
		ret := 1
		if err != nil {
			ret = 0
		}
		util.AddAuditFromRequest(r, util.AuditModuleIntentEngine, util.AuditOperationImport,
			auditMsg, ret)
	}()

	appID := util.GetAppID(r)
	file, info, err := r.FormFile("file")
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.INTENT_FORMAT_ERROR,
			err.Error()), http.StatusUnprocessableEntity)
		auditMsg += ": " + util.Msg["GetUploadFileFail"]
		return
	}
	defer file.Close()
	auditMsg += " " + info.Filename

	version, retCode, err := UploadIntents(appID, file, info)
	if err != nil {
		if retCode == ApiError.INTENT_FORMAT_ERROR {
			util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.INTENT_FORMAT_ERROR,
				err.Error()), http.StatusUnprocessableEntity)
			auditMsg += ": " + util.Msg["IntentFormatError"]
			return
		} else if retCode == ApiError.IO_ERROR {
			util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.IO_ERROR,
				err.Error()), http.StatusUnprocessableEntity)
			auditMsg += ": " + util.Msg["IOError"]
			return
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			auditMsg = fmt.Sprintf("%s: %s [%s]", auditMsg, util.Msg["UnknownError"], err.Error())
			return
		}
	}

	resp := UploadIntentsResponse{
		Version: version,
	}

	util.WriteJSON(w, util.GenRetObj(ApiError.SUCCESS, resp))
}

func handleDownloadIntents(w http.ResponseWriter, r *http.Request) {
	appID := util.GetAppID(r)
	v := r.URL.Query().Get("version")
	auditMsg := util.Msg["ExportIntentEngine"]

	var version int
	if v == "" {
		version = 0
	} else {
		ver, err := strconv.Atoi(v)
		if err != nil {
			http.Error(w, "Invalid intent dataset version", http.StatusBadRequest)
			util.AddAuditFromRequest(r, util.AuditModuleIntentEngine, util.AuditOperationExport,
				fmt.Sprintf("%s: %s", auditMsg, util.Msg["IntentVersionError"]), 0)
			return
		}

		version = ver
	}

	content, filename, errno, err := GetDownloadIntents(appID, version)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(errno, err.Error()), ApiError.GetHttpStatus(errno))
		util.AddAuditFromRequest(r, util.AuditModuleIntentEngine, util.AuditOperationExport,
			auditMsg, 0)
		return
	}

	fileNameHeader := fmt.Sprintf("attachment; filename=%s", filename)
	w.Header().Set("Content-Disposition", fileNameHeader)
	w.Header().Set("Content-Type",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(content)))
	w.Write(content)
	util.AddAuditFromRequest(r, util.AuditModuleIntentEngine, util.AuditOperationExport,
		auditMsg, 1)
}

func handleTrain(w http.ResponseWriter, r *http.Request) {
	appID := util.GetAppID(r)
	v := r.URL.Query().Get("version")
	auto := r.URL.Query().Get("auto_reload")
	engine := r.URL.Query().Get("engine")

	var version int
	if v == "" {
		version = 0
	} else {
		ver, err := strconv.Atoi(v)
		if err != nil {
			http.Error(w, "Invalid intent dataset version", http.StatusBadRequest)
			return
		}

		version = ver
	}

	var autoReload bool
	if auto == "" {
		autoReload = true
	} else {
		_auto, err := strconv.ParseBool(auto)
		if err != nil {
			http.Error(w, "Invalid auto_reload parameter", http.StatusBadRequest)
			return
		}

		autoReload = _auto
	}

	var trainEngine int
	if engine == "" {
		trainEngine = TrainBothEngines
	} else if engine == "intent_engine" {
		trainEngine = TrainIntentEngine
	} else if engine == "rule_engine" {
		trainEngine = TrainRuleEngine
	} else {
		http.Error(w, "Invalid engine parameter", http.StatusBadRequest)
		return
	}

	retCode, err := Train(appID, version, autoReload, trainEngine)
	if err != nil {
		if retCode == ApiError.REQUEST_ERROR {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	} else if retCode == ApiError.NOT_FOUND_ERROR {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func handleGetTrainStatus(w http.ResponseWriter, r *http.Request) {
	appID := util.GetAppID(r)
	v := r.URL.Query().Get("version")

	var version int
	if v == "" {
		version = 0
	} else {
		ver, err := strconv.Atoi(v)
		if err != nil {
			http.Error(w, "Invalid intent dataset version", http.StatusBadRequest)
			return
		}

		version = ver
	}

	resp, retCode, err := GetTrainStatus(appID, version)
	if err != nil {
		if retCode == ApiError.REQUEST_ERROR {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else if retCode == ApiError.NOT_FOUND_ERROR {
			statusResp := StatusResponse{
				IntentEngineStatus: NotTrained,
				RuleEngineStatus:   NotTrained,
			}
			util.WriteJSON(w, util.GenRetObj(ApiError.SUCCESS, statusResp))
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	} else if retCode == ApiError.NOT_FOUND_ERROR {
		statusResp := StatusResponse{
			IntentEngineStatus: NotTrained,
			RuleEngineStatus:   NotTrained,
		}
		util.WriteJSON(w, util.GenRetObj(ApiError.SUCCESS, statusResp))
		return
	}

	util.WriteJSON(w, util.GenRetObj(ApiError.SUCCESS, resp))
}

func handleGetData(w http.ResponseWriter, r *http.Request) {
	appID := r.URL.Query().Get("app_id")
	if appID == "" {
		http.Error(w, "app_id not specified", http.StatusBadRequest)
		return
	}

	flag := r.URL.Query().Get("flag")
	if flag == "" {
		http.Error(w, "flag not specified", http.StatusBadRequest)
		return
	}

	trainingData, retCode, err := GetTrainingData(appID, flag)
	if err != nil {
		if retCode == ApiError.REQUEST_ERROR {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	} else if retCode == ApiError.NOT_FOUND_ERROR {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	util.WriteJSON(w, trainingData)
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
