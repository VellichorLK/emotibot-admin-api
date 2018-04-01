package Task

import (
	"fmt"
	"strings"

	"emotibot.com/emotigo/module/admin-api/ApiError"
	"emotibot.com/emotigo/module/admin-api/util"
	"github.com/kataras/iris/context"
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

func handleGetScenarios(ctx context.Context) {
	appid := util.GetAppID(ctx)
	userID := appid
	taskURL := getEnvironment("SERVER_URL")
	scenarioid := ctx.Params().GetEscape("scenarioid")
	public := ctx.Params().GetEscape("public")

	if scenarioid == "" {
		scenarioid = ctx.FormValue("scenarioid")
	}
	if public == "" {
		public = ctx.FormValue("public")
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
		ctx.JSON(util.GenRetObj(ApiError.WEB_REQUEST_ERROR, err.Error()))
		ctx.Writef(err.Error())
	} else {
		ctx.Writef(content)
	}
}

func handlePutScenarios(ctx context.Context) {
	appid := util.GetAppID(ctx)
	userID := appid
	taskURL := getEnvironment("SERVER_URL")
	scenarioid := ctx.FormValue("scenarioid")
	layout := ctx.FormValue("layout")
	content := ctx.FormValue("content")
	delete := ctx.FormValue("delete")
	publish := ctx.FormValue("publish")

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
		ctx.JSON(util.GenRetObj(ApiError.WEB_REQUEST_ERROR, err.Error()))
		ctx.Writef(err.Error())
	} else {
		ctx.Writef(content)
	}
}
func handlePostScenarios(ctx context.Context) {
	method := ctx.FormValue("method")

	if method == "GET" {
		handleGetScenarios(ctx)
		return
	} else if method == "PUT" {
		handlePutScenarios(ctx)
		return
	}

	appid := util.GetAppID(ctx)
	userID := appid
	template := ctx.FormValue("template")
	scenarioName := ctx.FormValue("scenarioName")
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
		ctx.JSON(util.GenRetObj(ApiError.WEB_REQUEST_ERROR, err.Error()))
		ctx.Writef(err.Error())
	} else {
		ctx.Writef(content)
	}
}

func handleUpdateApp(ctx context.Context) {
	appid := util.GetAppID(ctx)
	userID := appid
	enable := ctx.FormValue("enable")
	scenarioID := ctx.FormValue("scenarioid")
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
		ctx.JSON(util.GenRetObj(ApiError.WEB_REQUEST_ERROR, err.Error()))
	} else {
		ctx.Header("Content-type", "application/json; charset=utf-8")
		// ctx.Header("Content-type", "text/plain; charset=utf-8")
		ctx.Writef(content)
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

func handleGetApps(ctx context.Context) {
	appid := util.GetAppID(ctx)
	// userID := util.GetUserID(ctx)
	taskURL := getEnvironment("SERVER_URL")

	// Hack in task-engine, use appid as userid
	url := fmt.Sprintf("%s/%s/%s?userid=%s", taskURL, taskAppEntry, appid, appid)
	util.LogTrace.Printf("Get apps: %s", url)
	content, err := util.HTTPGetSimple(url)
	if err != nil {
		ctx.JSON(util.GenRetObj(ApiError.WEB_REQUEST_ERROR, err.Error()))
	} else {
		// Cannot use json, or ui will has error...
		// ctx.Header("Content-type", "application/json; charset=utf-8")
		ctx.Header("Content-type", "text/plain; charset=utf-8")
		ctx.Writef(content)
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
