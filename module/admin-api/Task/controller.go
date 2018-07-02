package Task

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"emotibot.com/emotigo/module/admin-api/ApiError"
	"emotibot.com/emotigo/module/admin-api/util"
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
		},
	}
}

func handleUploadScenario(w http.ResponseWriter, r *http.Request) {
}

func handleUploadScenarios(w http.ResponseWriter, r *http.Request) {
	appid := util.GetAppID(r)
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
		return
	}
	ret := map[string]interface{}{
		"return": 0,
		"error":  "Update success",
	}
	if err == nil {
		ImportScenario(appid, useNewID, taskEngineJSON)
	} else {
		ImportScenarios(appid, useNewID, *multiTaskEngineJSON)
	}
	util.LogInfo.Printf("info: %+v\n", ret)
	util.WriteJSON(w, ret)
}

func handleGetScenarios(w http.ResponseWriter, r *http.Request) {
	appid := util.GetAppID(r)
	userID := appid
	taskURL := getEnvironment("SERVER_URL")
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

	url := fmt.Sprintf("%s/%s/%s?%s", taskURL, taskScenarioEntry, scenarioid, params.Encode())
	util.LogTrace.Printf("Get Scenario URL: %s", url)
	content, err := util.HTTPGetSimple(url)
	if err != nil {
		util.WriteJSON(w, util.GenRetObj(ApiError.WEB_REQUEST_ERROR, err.Error()))
	} else {
		io.WriteString(w, content)
	}
	// FIXME: trick here, task-engine will call update almost every click on UI
	// it will cause too much audit log into database
	// As a result, use get API to audit start edit only.
	if scenarioid != "all" {
		auditMsg := fmt.Sprintf("%s%s%s ID: %s", util.Msg["Start"], util.Msg["Modify"], util.Msg["TaskEngineScenario"], scenarioid)
		addAuditLog(r, util.AuditOperationEdit, auditMsg, err == nil)
	}
}

func handlePutScenarios(w http.ResponseWriter, r *http.Request) {
	appid := util.GetAppID(r)
	userID := appid
	taskURL := getEnvironment("SERVER_URL")
	scenarioid := r.FormValue("scenarioid")
	layout := r.FormValue("layout")
	content := r.FormValue("content")
	delete := r.FormValue("delete")
	publish := r.FormValue("publish")

	params := map[string]interface{}{
		"appid":  appid,
		"userid": userID,
	}

	if content != "" {
		params["content"] = content
		params["layout"] = layout
	} else if delete != "" {
		params["delete"] = true
	} else if publish != "" {
		params["publish"] = true
	}
	url := fmt.Sprintf("%s/%s/%s", taskURL, taskScenarioEntry, scenarioid)
	util.LogTrace.Printf("Put scenarios: %s with params: %#v", url, params)
	content, err := util.HTTPPutForm(url, params, 0)
	if err != nil {
		util.WriteJSON(w, util.GenRetObj(ApiError.WEB_REQUEST_ERROR, err.Error()))
	} else {
		io.WriteString(w, content)
	}
	if delete != "" {
		auditMsg := fmt.Sprintf("%s%s ID: %s", util.Msg["Delete"], util.Msg["TaskEngineScenario"], scenarioid)
		addAuditLog(r, util.AuditOperationDelete, auditMsg, err == nil)
	}
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

	appid := util.GetAppID(r)
	userID := appid
	template := r.FormValue("template")
	scenarioName := r.FormValue("scenarioName")
	if scenarioName == "" {
		scenarioName = "New Scenario"
	}
	taskURL := getEnvironment("SERVER_URL")
	params := map[string]string{
		"appid":        appid,
		"userid":       userID,
		"scenarioname": scenarioName,
	}
	if template != "" {
		params["template"] = template
	}
	url := fmt.Sprintf("%s/%s", taskURL, taskScenarioEntry)
	util.LogTrace.Printf("Post scenarios: %s", url)
	content, err := util.HTTPPostForm(url, params, 0)
	if err != nil {
		util.WriteJSON(w, util.GenRetObj(ApiError.WEB_REQUEST_ERROR, err.Error()))
	} else {
		io.WriteString(w, content)
	}

	auditMsg := fmt.Sprintf("%s%s: %s", util.Msg["Add"], util.Msg["TaskEngineScenario"], scenarioName)
	addAuditLog(r, util.AuditOperationAdd, auditMsg, err == nil)
}

func handleUpdateApp(w http.ResponseWriter, r *http.Request) {
	appid := util.GetAppID(r)
	userID := appid
	enable := r.FormValue("enable")
	scenarioID := r.FormValue("scenarioid")
	taskURL := getEnvironment("SERVER_URL")

	url := fmt.Sprintf("%s/%s", taskURL, taskAppEntry)
	params := map[string]string{
		"userid":     userID,
		"scenarioid": scenarioID,
		"enable":     enable,
		"appid":      appid,
	}

	content, err := util.HTTPPostForm(url, params, 0)
	if err != nil {
		util.WriteJSON(w, util.GenRetObj(ApiError.WEB_REQUEST_ERROR, err.Error()))
	} else {
		r.Header.Set("Content-type", "application/json; charset=utf-8")
		io.WriteString(w, content)
	}

	if scenarioID == "all" {
		if enable == "true" {
			EnableAllScenario(appid)
		} else {
			DisableAllScenario(appid)
		}
	} else {
		if enable == "true" {
			EnableScenario(appid, scenarioID)
		} else {
			DisableScenario(appid, scenarioID)
		}
	}
}

func handleGetApps(w http.ResponseWriter, r *http.Request) {
	appid := util.GetAppID(r)
	// userID := util.GetUserID(ctx)
	taskURL := getEnvironment("SERVER_URL")

	// Hack in task-engine, use appid as userid
	url := fmt.Sprintf("%s/%s/%s?userid=%s", taskURL, taskAppEntry, appid, appid)
	util.LogTrace.Printf("Get apps: %s", url)
	content, err := util.HTTPGetSimple(url)
	if err != nil {
		util.WriteJSON(w, util.GenRetObj(ApiError.WEB_REQUEST_ERROR, err.Error()))
	} else {
		// Cannot use json, or ui will has error...
		// r.Header.Set("Content-type", "application/json; charset=utf-8")
		r.Header.Set("Content-type", "text/plain; charset=utf-8")
		io.WriteString(w, content)
	}
}

func handleGetMapTableList(w http.ResponseWriter, r *http.Request) {
	appid := util.GetAppID(r)
	userID := util.GetUserID(r)
	userInQuery := r.URL.Query().Get("user")
	if userInQuery != "" {
		userID = userInQuery
	}

	util.LogTrace.Printf("Get mapping list of %s, %s\n", appid, userID)
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

func handleGetMapTable(w http.ResponseWriter, r *http.Request) {
	appid := util.GetAppID(r)
	userID := util.GetUserID(r)
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
	appid := util.GetAppID(r)
	userID := util.GetUserID(r)
	errno := ApiError.SUCCESS
	var auditMsg bytes.Buffer
	var ret string
	defer func() {
		status := ApiError.GetHttpStatus(errno)
		util.LogTrace.Printf("Upload mapping table ret: %d, %s\n", errno, ret)
		util.WriteJSONWithStatus(w, map[string]interface{}{
			"error":  ret,
			"return": errno,
		}, status)

		if errno == ApiError.SUCCESS {
			addAuditLog(r, util.AuditOperationImport, auditMsg.String(), true)
		} else {
			auditMsg.WriteString(fmt.Sprintf(", %s", ret))
			addAuditLog(r, util.AuditOperationImport, auditMsg.String(), false)
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
	util.LogInfo.Printf("Receive uploaded file: %s", info.Filename)
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
		util.LogInfo.Printf("Update consul key:te/mapping_table result: %d, %s", result, err.Error())
	} else {
		util.LogInfo.Printf("Update consul key:te/mapping_table result: %d", result)
	}
}

func handleDeleteMapTable(w http.ResponseWriter, r *http.Request) {
	appid := util.GetAppID(r)
	userID := util.GetUserID(r)
	tableName := r.FormValue("table_name")
	errno := ApiError.SUCCESS
	var auditMsg bytes.Buffer
	var ret string
	defer func() {
		status := ApiError.GetHttpStatus(errno)
		util.LogTrace.Printf("Upload mapping table ret: %d, %s\n", errno, ret)
		util.WriteJSONWithStatus(w, map[string]interface{}{
			"error":  ret,
			"return": errno,
		}, status)

		if errno == ApiError.SUCCESS {
			addAuditLog(r, util.AuditOperationDelete, auditMsg.String(), true)
		} else {
			auditMsg.WriteString(fmt.Sprintf(", %s", ret))
			addAuditLog(r, util.AuditOperationDelete, auditMsg.String(), false)
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
		util.LogInfo.Printf("Update consul key:te/mapping_table_all result: %d, %s", result, err.Error())
	} else {
		util.LogInfo.Printf("Update consul key:te/mapping_table_all result: %d", result)
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
	appid := util.GetAppID(r)
	user := util.GetUserID(r)
	ip := util.GetUserIP(r)
	retVal := 0
	if ret {
		retVal = 1
	}
	util.AddAuditLog(appid, user, ip, util.AuditModuleTaskEngine, op, msg, retVal)
}

func handleExportMapTable(w http.ResponseWriter, r *http.Request) {
	appid := util.GetAppID(r)
	userID := util.GetUserID(r)
	tableName := util.GetMuxVar(r, "name")
	tableNameInQuery := r.URL.Query().Get("mapping_table_name")
	var auditMsg bytes.Buffer
	var outputBuf bytes.Buffer

	errno := ApiError.SUCCESS
	defer func() {
		status := ApiError.GetHttpStatus(errno)
		w.WriteHeader(status)
		w.Write(outputBuf.Bytes())

		addAuditLog(r, util.AuditOperationDelete, auditMsg.String(), errno == ApiError.SUCCESS)
	}()

	util.LogTrace.Printf("Get mapping table: %s of %s, %s", tableName, userID, appid)
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
