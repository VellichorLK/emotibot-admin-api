package Task

import (
	"fmt"
	"io"
	"net/http"
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
			util.NewEntryPoint("GET", "scenarios", []string{}, handleGetScenarios),
			util.NewEntryPoint("PUT", "scenarios", []string{"create"}, handlePutScenarios),
			util.NewEntryPoint("POST", "scenarios", []string{"edit"}, handlePostScenarios),
		},
	}
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

	params := map[string]string{
		"appid":  appid,
		"userid": userID,
	}

	if content != "" {
		params["content"] = content
		params["layout"] = layout
	} else if delete != "" {
		params["delete"] = "true"
	} else if publish != "" {
		params["publish"] = "true"
	}
	url := fmt.Sprintf("%s/%s/%s", taskURL, taskScenarioEntry, scenarioid)
	util.LogTrace.Printf("Put scenarios: %s", url)
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
