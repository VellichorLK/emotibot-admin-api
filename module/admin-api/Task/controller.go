package Task

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"emotibot.com/emotigo/module/admin-api/ApiError"
	"emotibot.com/emotigo/module/admin-api/Dictionary"
	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/admin-api/util/audit"
	"emotibot.com/emotigo/module/admin-api/util/requestheader"
	"emotibot.com/emotigo/pkg/logger"
)

var (
	// ModuleInfo is needed for module define
	ModuleInfo util.ModuleInfo
)

const taskAppEntry = "task_engine_app"
const taskScenarioEntry = "task_engine_editor"

func init() {
	ModuleInfo = util.ModuleInfo{
		ModuleName: "task",
		EntryPoints: []util.EntryPoint{
			util.NewEntryPoint("GET", "apps", []string{}, handleGetApps),
			util.NewEntryPoint("POST", "apps", []string{}, handleUpdateApp),

			util.NewEntryPoint("POST", "scenarios-upload", []string{}, handleUploadScenarios),
			util.NewEntryPoint("POST", "scenario-upload", []string{}, handleUploadScenario),
			util.NewEntryPoint("GET", "scenarios", []string{}, handleGetScenarios),
			util.NewEntryPoint("PUT", "scenarios", []string{}, handlePutScenarios),
			util.NewEntryPoint("POST", "scenarios", []string{}, handlePostScenarios),

			util.NewEntryPoint("GET", "mapping-tables", []string{}, handleGetMapTableList),
			util.NewEntryPoint("POST", "mapping-table/upload", []string{}, handleUploadMapTable),
			util.NewEntryPoint("POST", "mapping-table/delete", []string{}, handleDeleteMapTable),
			util.NewEntryPoint("GET", "mapping-table/export", []string{}, handleExportMapTable),
			util.NewEntryPoint("GET", "mapping-table/{name}", []string{}, handleGetMapTable),
			util.NewEntryPoint("GET", "mapping-table", []string{}, handleGetMapTable),
			util.NewEntryPoint("POST", "spreadsheet", []string{}, handleUploadSpreadSheet),
			util.NewEntryPoint("POST", "intent", []string{}, handleIntentV1),
			util.NewEntryPointWithVer("GET", "mapping-tables", []string{}, handleGetMapTableListV2, 2),
			util.NewEntryPointWithVer("GET", "mapping-tables/all", []string{}, handleGetMapTableAllV2, 2),

			util.NewEntryPoint("POST", "audit", []string{}, handleAudit),
			util.NewEntryPoint("GET", "config", []string{}, handleGetConfig),
		},
	}
}

func handleGetConfig(w http.ResponseWriter, r *http.Request) {
	teConfig, errno, err := GetTaskEngineConfig()
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(errno, err.Error()), ApiError.GetHttpStatus(errno))
		return
	}
	retString, err := json.Marshal(teConfig)
	if err != nil {
		errno = ApiError.JSON_PARSE_ERROR
		util.WriteJSONWithStatus(w, util.GenRetObj(errno, err.Error()), ApiError.GetHttpStatus(errno))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, string(retString))
	return
}

func handleUploadScenario(w http.ResponseWriter, r *http.Request) {
}

func handleUploadScenarios(w http.ResponseWriter, r *http.Request) {
	appid := requestheader.GetAppID(r)
	useNewID := r.FormValue("useNewId") == "true"
	file, info, err := r.FormFile("scenario_json")

	ext := path.Ext(info.Filename)
	if ext != ".json" {
		util.WriteJSONWithStatus(w,
			util.GenRetObj(ApiError.REQUEST_ERROR, "file ext should be json"),
			http.StatusBadRequest)
		return
	}
	content, err := ReadUploadJSON(file)
	if err != nil {
		util.WriteJSONWithStatus(w,
			util.GenRetObj(ApiError.REQUEST_ERROR, "read file fail: "+err.Error()),
			http.StatusBadRequest)
		return
	}
	taskEngineJSON := &map[string]interface{}{}
	multiTaskEngineJSON := &[]interface{}{}
	err = json.Unmarshal([]byte(content), taskEngineJSON)
	err2 := json.Unmarshal([]byte(content), multiTaskEngineJSON)
	if err != nil && err2 != nil {
		util.WriteJSONWithStatus(w,
			util.GenRetObj(ApiError.REQUEST_ERROR, fmt.Sprintf("invalid json: %s, %s", err.Error(), err2.Error())),
			http.StatusBadRequest)
		auditMsg := util.Msg["AuditImportJSONError"]
		addAuditLog(r, audit.AuditOperationImport, auditMsg, true)
		return
	}
	ret := map[string]interface{}{
		"return": 0,
		"error":  "Update success",
	}
	if err == nil {
		ImportScenario(appid, appid, useNewID, taskEngineJSON)
	} else {
		ImportScenarios(appid, appid, useNewID, *multiTaskEngineJSON)
	}
	auditMsg := fmt.Sprintf(util.Msg["AuditImportTpl"], info.Filename)
	addAuditLog(r, audit.AuditOperationImport, auditMsg, true)
	util.WriteJSON(w, ret)
}

func handleGetScenarios(w http.ResponseWriter, r *http.Request) {
	appid := requestheader.GetAppID(r)
	userID := appid
	scenarioid := r.URL.Query().Get("scenarioid")
	public := r.URL.Query().Get("public")

	if scenarioid == "" {
		scenarioid = r.FormValue("scenarioid")
	}
	if public == "" {
		public = r.FormValue("public")
	}
	params := url.Values{
		"appid": []string{appid},
	}
	if public != "" {
		params.Set("public", public)
	} else {
		params.Set("userid", userID)
	}

	if scenarioid == "all" {
		public, _ := strconv.Atoi(public)
		if public > 0 {
			// get public(template) scenarios
			templateScenarioInfoList, errno, err := GetTemplateScenarioInfoList()
			if err != nil {
				util.WriteJSONWithStatus(w, util.GenRetObj(errno, err.Error()), ApiError.GetHttpStatus(errno))
				return
			}
			ret := TemplateScenarioInfoListResponse{
				Result: templateScenarioInfoList,
			}
			retString, err := json.Marshal(ret)
			if err != nil {
				errno = ApiError.JSON_PARSE_ERROR
				util.WriteJSONWithStatus(w, util.GenRetObj(errno, err.Error()), ApiError.GetHttpStatus(errno))
				return
			}
			w.Header().Set("Content-Type", "application/json")
			io.WriteString(w, string(retString))
			return
		}
		// TODO handle scenarioid == all, public == 0 API
		logger.Info.Printf("scenarioid: %s, public: %d", scenarioid, public)
	} else {
		teConfig, errno, err := GetTaskEngineConfig()
		if err != nil {
			util.WriteJSONWithStatus(w, util.GenRetObj(errno, err.Error()), ApiError.GetHttpStatus(errno))
			return
		}
		scenario := &Scenario{}
		if teConfig.TEv2Config.EnableJSCode {
			scenario, errno, err = GetDecryptedScenario(scenarioid)
		} else {
			scenario, errno, err = GetScenario(scenarioid)
		}
		if err != nil {
			util.WriteJSONWithStatus(w, util.GenRetObj(errno, err.Error()), ApiError.GetHttpStatus(errno))
			return
		} else if scenario == nil {
			errMsg := fmt.Sprintf("No scenario found in DB with scenarioID: %s", scenarioid)
			util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, errMsg), http.StatusNotFound)
			return
		} else {
			result := GetScenarioResult{
				Content:        scenario.Content,
				Layout:         scenario.Layout,
				Editing:        scenario.Editing,
				EditingContent: scenario.EditingContent,
				EditingLayout:  scenario.EditingLayout,
			}
			ret := GetScenarioResponse{
				Result: &result,
			}
			retString, err := json.Marshal(ret)
			if err != nil {
				errno = ApiError.JSON_PARSE_ERROR
				util.WriteJSONWithStatus(w, util.GenRetObj(errno, err.Error()), ApiError.GetHttpStatus(errno))
				return
			}
			w.Header().Set("Content-Type", "application/json")
			io.WriteString(w, string(retString))
			return
		}
	}
}

func handlePutScenarios(w http.ResponseWriter, r *http.Request) {
	appid := requestheader.GetAppID(r)
	scenarioid := r.FormValue("scenarioid")
	editingContent := r.FormValue("content")
	editingLayout := r.FormValue("layout")
	publish := r.FormValue("publish")
	delete := r.FormValue("delete")

	if delete != "" {
		// delete scenario
		errno, err := DeleteScenario(scenarioid, appid)
		if err != nil {
			util.WriteJSONWithStatus(w, util.GenRetObj(errno, err.Error()), ApiError.GetHttpStatus(errno))
			return
		}
		// delete app-scenario pair in taskengineapp
		errno, err = DeleteAppScenario(scenarioid, appid)
		if err != nil {
			util.WriteJSONWithStatus(w, util.GenRetObj(errno, err.Error()), ApiError.GetHttpStatus(errno))
			return
		}
		errno, err = UpdateAppScenarioPairToConsul(appid)
		if err != nil {
			errno := ApiError.JSON_PARSE_ERROR
			util.WriteJSONWithStatus(w, util.GenRetObj(errno, err.Error()), ApiError.GetHttpStatus(errno))
			return
		}
		auditMsg := fmt.Sprintf(util.Msg["AuditPublishTpl"], scenarioid)
		addAuditLog(r, audit.AuditOperationPublish, auditMsg, err == nil)
	} else if publish != "" {
		// publish scenario
		errno, err := PublishScenario(scenarioid, appid, appid)
		if err != nil {
			util.WriteJSONWithStatus(w, util.GenRetObj(errno, err.Error()), ApiError.GetHttpStatus(errno))
			return
		}
		// update consul to inform TE to reload scenarios
		errno, err = util.ConsulUpdateTaskEngineScenario()
		if err != nil {
			logger.Error.Printf("Failed to update consul key:te/scenario errno: %d, %s", errno, err.Error())
			util.WriteJSONWithStatus(w, util.GenRetObj(errno, err.Error()), ApiError.GetHttpStatus(errno))
			return
		}
		auditMsg := fmt.Sprintf(util.Msg["AuditPublishTpl"], scenarioid)
		addAuditLog(r, audit.AuditOperationPublish, auditMsg, err == nil)
	} else {
		// update scenario
		errno, err := UpdateScenario(scenarioid, appid, appid, editingContent, editingLayout)
		if err != nil {
			util.WriteJSONWithStatus(w, util.GenRetObj(errno, err.Error()), ApiError.GetHttpStatus(errno))
			return
		}
	}
	ret := ResultMsgResponse{
		Msg: "Update success",
	}
	retString, err := json.Marshal(ret)
	if err != nil {
		errno := ApiError.JSON_PARSE_ERROR
		util.WriteJSONWithStatus(w, util.GenRetObj(errno, err.Error()), ApiError.GetHttpStatus(errno))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, string(retString))
	return
}
func handlePostScenarios(w http.ResponseWriter, r *http.Request) {
	method := r.FormValue("method")
	if method == "GET" {
		handleGetScenarios(w, r)
		return
	} else if method == "PUT" {
		handlePutScenarios(w, r)
		return
	}
	appid := requestheader.GetAppID(r)
	userid := appid
	scenarioName := r.FormValue("scenarioName")
	if scenarioName == "" {
		scenarioName = "New Scenario"
	}

	metadata, errno, err := CreateInitialScenario(appid, userid, scenarioName)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(errno, err.Error()), ApiError.GetHttpStatus(errno))
		return
	}

	ret := CreateScenarioResponse{
		Template: &TemplateResult{
			Metadata: metadata,
		},
		ScenarioID: metadata.ScenarioID,
	}
	retString, err := json.Marshal(ret)
	if err != nil {
		errno := ApiError.JSON_PARSE_ERROR
		util.WriteJSONWithStatus(w, util.GenRetObj(errno, err.Error()), ApiError.GetHttpStatus(errno))
		return
	}
	auditMsg := fmt.Sprintf("%s%s: %s", util.Msg["Add"], util.Msg["TaskEngineScenario"], scenarioName)
	addAuditLog(r, audit.AuditOperationAdd, auditMsg, err == nil)
	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, string(retString))
	return
}

func handleUpdateApp(w http.ResponseWriter, r *http.Request) {
	appid := requestheader.GetAppID(r)
	enable := r.FormValue("enable")
	scenarioid := r.FormValue("scenarioid")

	var errno int
	var err error
	var ret ResultMsgResponse
	if enable == "true" {
		errno, err = CreateAppScenario(scenarioid, appid)
		ret = ResultMsgResponse{
			Msg: "Enable success",
		}
	} else {
		errno, err = DeleteAppScenario(scenarioid, appid)
		ret = ResultMsgResponse{
			Msg: "Disable success",
		}
	}
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(errno, err.Error()), ApiError.GetHttpStatus(errno))
		return
	}
	errno, err = UpdateAppScenarioPairToConsul(appid)
	if err != nil {
		errno := ApiError.JSON_PARSE_ERROR
		util.WriteJSONWithStatus(w, util.GenRetObj(errno, err.Error()), ApiError.GetHttpStatus(errno))
		return
	}
	retString, err := json.Marshal(ret)
	if err != nil {
		errno := ApiError.JSON_PARSE_ERROR
		util.WriteJSONWithStatus(w, util.GenRetObj(errno, err.Error()), ApiError.GetHttpStatus(errno))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, string(retString))

	auditTpl := ""
	target := scenarioid
	operation := audit.AuditOperationActive
	if enable == "true" {
		auditTpl = util.Msg["AuditActiveTpl"]
		if scenarioid == "all" {
			EnableAllScenario(appid)
			target = util.Msg["All"]
		} else {
			EnableScenario(appid, scenarioid)
		}
	} else {
		auditTpl = util.Msg["AuditDeactiveTpl"]
		operation = audit.AuditOperationDeactive
		if scenarioid == "all" {
			DisableAllScenario(appid)
			target = util.Msg["All"]
		} else {
			DisableScenario(appid, scenarioid)
		}
	}
	auditMsg := fmt.Sprintf(auditTpl, target)
	addAuditLog(r, operation, auditMsg, err == nil)
}

func handleGetApps(w http.ResponseWriter, r *http.Request) {
	appid := requestheader.GetAppID(r)
	scenarioInfoList, errno, err := GetScenarioInfoList(appid, appid)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(errno, err.Error()), ApiError.GetHttpStatus(errno))
		return
	}
	ret := ScenarioInfoListResponse{
		Msg: scenarioInfoList,
	}
	retString, err := json.Marshal(ret)
	if err != nil {
		errno = ApiError.JSON_PARSE_ERROR
		util.WriteJSONWithStatus(w, util.GenRetObj(errno, err.Error()), ApiError.GetHttpStatus(errno))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, string(retString))
}

func handleGetMapTableList(w http.ResponseWriter, r *http.Request) {
	appid := requestheader.GetAppID(r)
	userID := requestheader.GetUserID(r)
	userInQuery := r.URL.Query().Get("user")
	if userInQuery != "" {
		userID = userInQuery
	}

	logger.Trace.Printf("Get mapping list of %s, %s\n", appid, userID)
	list, errno, err := GetMapTableList(appid, userID)
	if err != nil {
		w.WriteHeader(ApiError.GetHttpStatus(errno))
		w.Write([]byte(err.Error()))
		return
	}

	buf, err := json.Marshal(list)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Write(buf)
}

// handleGetMapTableListV2 load mapping table list by appid from wordbank
func handleGetMapTableListV2(w http.ResponseWriter, r *http.Request) {
	appID := requestheader.GetAppID(r)
	// if the user in query url is templateadmin, get the template scenario mapping tables
	userInQuery := r.URL.Query().Get("user")
	if userInQuery == "templateadmin" {
		appID = userInQuery
	}
	logger.Trace.Printf("appID: %+v", appID)

	wordbanks, errno, err := Dictionary.GetWordbanksV3(appID)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(errno, err.Error()), ApiError.GetHttpStatus(errno))
		return
	}

	mtList := GetMapTableListV2(wordbanks)

	buf, err := json.Marshal(mtList)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Write(buf)
}

// handleGetMapTableAllV2 load mapping table list for all appid from wordbank
func handleGetMapTableAllV2(w http.ResponseWriter, r *http.Request) {
	rootMap, errno, err := Dictionary.GetWordbanksAllV3()
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(errno, err.Error()), ApiError.GetHttpStatus(errno))
		return
	}

	appidToMtMap := GetMapTableAllV2(rootMap)
	buf, err := json.Marshal(appidToMtMap)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Write(buf)
}

func handleGetMapTable(w http.ResponseWriter, r *http.Request) {
	appid := requestheader.GetAppID(r)
	userID := requestheader.GetUserID(r)
	tableName := util.GetMuxVar(r, "name")
	tableNameInQuery := r.URL.Query().Get("mapping_table_name")

	if tableName == "" {
		tableName = tableNameInQuery
	}
	if tableName == "" {
		w.WriteHeader(ApiError.GetHttpStatus(ApiError.REQUEST_ERROR))
		err := util.GenBadRequestError(util.Msg["MappingTableName"])
		w.Write([]byte(err.Error()))
		return
	}

	content, errno, err := GetMapTableContent(appid, userID, tableName)
	if err != nil {
		w.WriteHeader(ApiError.GetHttpStatus(errno))
		w.Write([]byte(err.Error()))
		return
	}
	w.Write([]byte(content))
}

func handleUploadMapTable(w http.ResponseWriter, r *http.Request) {
	appid := requestheader.GetAppID(r)
	userID := requestheader.GetUserID(r)
	errno := ApiError.SUCCESS
	var auditMsg bytes.Buffer
	var ret string
	defer func() {
		status := ApiError.GetHttpStatus(errno)
		logger.Trace.Printf("Upload mapping table ret: %d, %s\n", errno, ret)
		util.WriteJSONWithStatus(w, map[string]interface{}{
			"error":  ret,
			"return": errno,
		}, status)

		if errno == ApiError.SUCCESS {
			addAuditLog(r, audit.AuditOperationImport, auditMsg.String(), true)
		} else {
			auditMsg.WriteString(fmt.Sprintf(", %s", ret))
			addAuditLog(r, audit.AuditOperationImport, auditMsg.String(), false)
		}
	}()
	auditMsg.WriteString(fmt.Sprintf("%s%s", util.Msg["UploadFile"], util.Msg["MappingTable"]))

	file, info, err := r.FormFile("mapping_table")
	if err != nil {
		errno = ApiError.IO_ERROR
		ret = fmt.Sprintf("%s: %s", util.Msg["ErrorReadFileError"], err.Error())
		return
	}
	defer file.Close()
	logger.Info.Printf("Receive uploaded file: %s", info.Filename)
	auditMsg.WriteString(info.Filename)

	size := info.Size
	if size == 0 {
		errno = ApiError.REQUEST_ERROR
		ret = util.Msg["ErrorUploadEmptyFile"]
		return
	}

	buf := make([]byte, size)
	_, err = file.Read(buf)
	if err != nil {
		errno = ApiError.IO_ERROR
		ret = fmt.Sprintf("%s: %s", util.Msg["ErrorReadFileError"], err.Error())
		return
	}

	mappingTuple, err := ParseUploadMappingTable(buf)
	if err != nil {
		errno = ApiError.REQUEST_ERROR
		ret = fmt.Sprintf("%s: %s", util.Msg["ParseError"], err.Error())
		return
	}

	now := time.Now()
	nowTimeStr := now.Format("2006-01-02 15:04:05")
	fileName := fmt.Sprintf("%s_%d%02d%02d%02d%02d%02d",
		info.Filename, now.Year(), now.Month(), now.Day(),
		now.Hour(), now.Minute(), now.Second())

	mappingObj := MappingTable{
		MappingData: mappingTuple,
		Metadata: &MapMeta{
			UpdateTime:       nowTimeStr,
			UpdateUser:       userID,
			MappingTableName: fileName,
		},
	}

	content, err := json.Marshal(mappingObj)
	if err != nil {
		errno = ApiError.IO_ERROR
		ret = fmt.Sprintf("%s: %s", util.Msg["MarshalError"], err.Error())
		return
	}

	errno, err = SaveMappingTable(userID, appid, fileName, string(content))
	if err != nil {
		ret = fmt.Sprintf("%s: %s", util.Msg["ServerError"], err.Error())
		return
	}
	auditMsg.WriteString(fmt.Sprintf(" -> %s", fileName))

	// inform TE to reload mapping table
	var result int
	result, err = util.ConsulUpdateTaskEngineMappingTable()
	if err != nil {
		logger.Info.Printf("Update consul key:te/mapping_table result: %d, %s", result, err.Error())
	} else {
		logger.Info.Printf("Update consul key:te/mapping_table result: %d", result)
	}
}

func handleDeleteMapTable(w http.ResponseWriter, r *http.Request) {
	appid := requestheader.GetAppID(r)
	userID := requestheader.GetUserID(r)
	tableName := r.FormValue("table_name")
	errno := ApiError.SUCCESS
	var auditMsg bytes.Buffer
	var ret string
	defer func() {
		status := ApiError.GetHttpStatus(errno)
		logger.Trace.Printf("Upload mapping table ret: %d, %s\n", errno, ret)
		util.WriteJSONWithStatus(w, map[string]interface{}{
			"error":  ret,
			"return": errno,
		}, status)

		if errno == ApiError.SUCCESS {
			addAuditLog(r, audit.AuditOperationDelete, auditMsg.String(), true)
		} else {
			auditMsg.WriteString(fmt.Sprintf(", %s", ret))
			addAuditLog(r, audit.AuditOperationDelete, auditMsg.String(), false)
		}
	}()
	auditMsg.WriteString(fmt.Sprintf("%s%s", util.Msg["Delete"], util.Msg["MappingTable"]))

	if tableName == "" {
		errno = ApiError.REQUEST_ERROR
		ret = util.GenBadRequestError(util.Msg["MappingTableName"]).Error()
		return
	}
	auditMsg.WriteString(tableName)

	err := DeleteMappingTable(appid, userID, tableName)
	if err != nil {
		errno = ApiError.DB_ERROR
		ret = err.Error()
		return
	}

	// inform TE to reload all mapping table
	var result int
	result, err = util.ConsulUpdateTaskEngineMappingTableAll()
	if err != nil {
		logger.Info.Printf("Update consul key:te/mapping_table_all result: %d, %s", result, err.Error())
	} else {
		logger.Info.Printf("Update consul key:te/mapping_table_all result: %d", result)
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

func addAuditLog(r *http.Request, op string, msg string, ret bool) {
	appid := requestheader.GetAppID(r)
	user := requestheader.GetUserID(r)
	ip := requestheader.GetUserIP(r)
	enterprise := requestheader.GetEnterpriseID(r)
	retVal := 0
	if ret {
		retVal = 1
	}
	audit.AddAuditLog(enterprise, appid, user, ip, audit.AuditModuleTaskEngine, op, msg, retVal)
}

func handleExportMapTable(w http.ResponseWriter, r *http.Request) {
	appid := requestheader.GetAppID(r)
	userID := requestheader.GetUserID(r)
	tableName := util.GetMuxVar(r, "name")
	tableNameInQuery := r.URL.Query().Get("mapping_table_name")
	var auditMsg bytes.Buffer
	var outputBuf bytes.Buffer

	errno := ApiError.SUCCESS
	defer func() {
		status := ApiError.GetHttpStatus(errno)
		w.WriteHeader(status)
		w.Write(outputBuf.Bytes())

		addAuditLog(r, audit.AuditOperationDelete, auditMsg.String(), errno == ApiError.SUCCESS)
	}()

	logger.Trace.Printf("Get mapping table: %s of %s, %s", tableName, userID, appid)
	if tableName == "" {
		tableName = tableNameInQuery
	}
	if tableName == "" {
		w.WriteHeader(ApiError.GetHttpStatus(ApiError.REQUEST_ERROR))
		err := util.GenBadRequestError(util.Msg["MappingTableName"])
		auditMsg.WriteString(fmt.Sprintf("%s", err.Error()))
		return
	}

	content, errno, err := GetMapTableContent(appid, userID, tableName)
	if err != nil {
		auditMsg.WriteString(fmt.Sprintf("%s: %s", util.Msg["ServerError"], err.Error()))
		return
	}

	mappingTable := MappingTable{}
	err = json.Unmarshal([]byte(content), &mappingTable)
	if err != nil {
		errno = ApiError.IO_ERROR
		auditMsg.WriteString(fmt.Sprintf("%s: %s", util.Msg["MarshalError"], err.Error()))
		return
	}
	auditMsg.WriteString(fmt.Sprintf("%s%s %s",
		util.Msg["DownloadFile"], util.Msg["MappingTable"], tableName))

	for _, tuple := range mappingTable.MappingData {
		if tuple == nil {
			continue
		}
		key := strings.Replace(tuple.Key, "\"", "\"\"\"", -1)
		value := strings.Replace(tuple.Value, "\"", "\"\"\"", -1)
		outputBuf.WriteString(fmt.Sprintf("%s,%s\n", key, value))
	}
}

func handleIntentV1(w http.ResponseWriter, r *http.Request) {
	reqType := r.FormValue("type")
	// For now, cu_intent type will never used
	// appid := r.FormValue("app_id")
	// cuIntent := r.FormValue("cu_intent")
	if reqType == "" {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, "empty type"), http.StatusBadRequest)
		return
	}
	intentURL := ""
	switch reqType {
	case "delete":
		fallthrough
	case "search":
		fallthrough
	case "check":
		util.Redirect(fmt.Sprintf("%s/%s", intentURL, reqType), w, r, 0)
	case "update":
		util.Redirect(fmt.Sprintf("%s/%s", intentURL, reqType), w, r, 3)
	// case "cu_intent":
	// 	if cuIntent == "" {
	// 		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, "empty cu_intent"), http.StatusBadRequest)
	// 		return
	// 	}
	//  cuIntentURL := ""
	// 	url := fmt.Sprintf("%s/%s", cuIntentURL, appid)
	// 	r.URL.Query().Set("value", cuIntent)
	// 	r.URL.Query().Set("key", "ccu")
	// 	util.Redirect(url, w, r, 0)
	default:
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, "no match type"), http.StatusBadRequest)
	}
}

func handleAudit(w http.ResponseWriter, r *http.Request) {
	logger.Trace.Println("Run: handleAudit")
	action := r.FormValue("action")
	msg := r.FormValue("msg")
	userID := requestheader.GetUserID(r)
	userIP := requestheader.GetUserIP(r)
	appid := requestheader.GetAppID(r)

	auditOp := ""
	switch action {
	case "edit":
		auditOp = audit.AuditOperationEdit
	case "export":
		auditOp = audit.AuditOperationExport
	default:
		util.WriteJSON(w, util.GenRetObj(ApiError.REQUEST_ERROR, "Unknown action"))
		return
	}

	enterprise := requestheader.GetEnterpriseID(r)
	err := audit.AddAuditLog(enterprise, appid, userID, userIP, audit.AuditModuleTaskEngine, auditOp, msg, 1)
	if err != nil {
		util.WriteJSON(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()))
	} else {
		util.WriteJSON(w, util.GenSimpleRetObj(ApiError.SUCCESS))
	}
	return
}
