package Task

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"

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
			util.NewEntryPoint("POST", "apps", []string{"edit"}, handleUpdateApp),

			util.NewEntryPoint("POST", "scenarios-upload", []string{"import"}, handleUploadScenarios),
			util.NewEntryPoint("POST", "scenario-upload", []string{"import"}, handleUploadScenario),
			util.NewEntryPoint("GET", "scenarios", []string{}, handleGetScenarios),
			util.NewEntryPoint("PUT", "scenarios", []string{"create"}, handlePutScenarios),
			util.NewEntryPoint("POST", "scenarios", []string{"edit"}, handlePostScenarios),
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

	params := []string{fmt.Sprintf("appid=%s", appid)}
	if public != "" {
		params = append(params, fmt.Sprintf("public=%s", public))
	} else {
		params = append(params, fmt.Sprintf("userid=%s", userID))
	}

	url := fmt.Sprintf("%s/%s/%s?%s", taskURL, taskScenarioEntry, scenarioid, strings.Join(params, "&"))
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
