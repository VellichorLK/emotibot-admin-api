package Task

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
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
			util.NewEntryPoint("GET", "mapping-table/{name}", []string{}, handleGetMapTable),
			util.NewEntryPoint("POST", "mapping-table/upload", []string{}, handleUploadMapTable),
			util.NewEntryPoint("POST", "mapping-table/delete", []string{}, handleDeleteMapTable),
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
		"error":  "",
	}
	// var ret map[string]interface{}
	if err == nil {
		ImportScenario(appid, useNewID, taskEngineJSON)
		// ret = map[string]interface{}{
		// 	"a":    appid,
		// 	"new":  fmt.Sprintf("%t", useNewID),
		// 	"file": taskEngineJSON,
		// }
	} else {
		ImportScenarios(appid, useNewID, *multiTaskEngineJSON)
		// ret = map[string]interface{}{
		// 	"a":    appid,
		// 	"new":  fmt.Sprintf("%t", useNewID),
		// 	"file": multiTaskEngineJSON,
		// }
	}
	// util.LogInfo.Printf("info: %+v\n", ret)
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
	if tableName == "" {
		w.WriteHeader(ApiError.GetHttpStatus(ApiError.REQUEST_ERROR))
		err := ApiError.GenBadRequestError("Table name")
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
	var ret string
	defer func() {
		status := ApiError.GetHttpStatus(errno)
		util.LogTrace.Printf("Upload mapping table ret: %d, %s\n", errno, ret)
		util.WriteJSONWithStatus(w, map[string]interface{}{
			"error":  ret,
			"return": errno,
		}, status)
	}()

	file, info, err := r.FormFile("mapping_table")
	if err != nil {
		errno = ApiError.IO_ERROR
		ret = fmt.Sprintf("%s: %s", util.Msg["ErrorReadFileError"], err.Error())
		return
	}
	defer file.Close()
	util.LogInfo.Printf("Receive uploaded file: %s", info.Filename)

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
		ret = fmt.Sprintf("Parse error: %s", err.Error())
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
		ret = fmt.Sprintf("JSON to string error: %s", err.Error())
		return
	}

	errno, err = SaveMappingTable(userID, appid, fileName, string(content))
	if err != nil {
		ret = err.Error()
	}
}

func handleDeleteMapTable(w http.ResponseWriter, r *http.Request) {
	appid := util.GetAppID(r)
	userID := util.GetUserID(r)
	tableName := r.FormValue("table_name")
	errno := ApiError.SUCCESS
	var ret string
	defer func() {
		status := ApiError.GetHttpStatus(errno)
		util.LogTrace.Printf("Upload mapping table ret: %d, %s\n", errno, ret)
		util.WriteJSONWithStatus(w, map[string]interface{}{
			"error":  ret,
			"return": errno,
		}, status)
	}()

	if tableName == "" {
		errno = ApiError.REQUEST_ERROR
		ret = ApiError.GenBadRequestError("Table name").Error()
		return
	}

	err := DeleteMappingTable(appid, userID, tableName)
	if err != nil {
		errno = ApiError.DB_ERROR
		ret = err.Error()
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
