package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"emotibot.com/emotigo/module/vipshop-admin/Dictionary"
	"emotibot.com/emotigo/module/vipshop-admin/FAQ"
	"emotibot.com/emotigo/module/vipshop-admin/QA"
	"emotibot.com/emotigo/module/vipshop-admin/Robot"
	"emotibot.com/emotigo/module/vipshop-admin/SelfLearning"
	"emotibot.com/emotigo/module/vipshop-admin/Stats"
	"emotibot.com/emotigo/module/vipshop-admin/Switch"
	"emotibot.com/emotigo/module/vipshop-admin/UI"
	"emotibot.com/emotigo/module/vipshop-admin/imagesManager"
	"emotibot.com/emotigo/module/vipshop-admin/util"
	"emotibot.com/emotigo/module/vipshop-admin/websocket"

	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
)

// constant define all const used in server
var constant = map[string]interface{}{
	"API_PREFIX":  "api",
	"API_VERSION": 1,
}

var serverConfig map[string]string

func init() {
	// util.LogInit(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)
	util.LogInit(os.Stderr, os.Stdout, os.Stdout, os.Stderr)
	if len(os.Args) > 1 {
		err := util.LoadConfigFromFile(os.Args[1])
		if err != nil {
			util.LogError.Printf(err.Error())
			os.Exit(-1)
		}
	}
}

func main() {
	app := iris.New()
	//Init Consul Client
	serverEnvs := util.GetEnvOf("server")
	consulAddr, ok := serverEnvs["CONSUL_URL"]
	if !ok {
		util.LogError.Printf("Can not init without server env:'CONSUL_URL env'")
		os.Exit(-1)
	}
	u, err := url.Parse(consulAddr)
	if err != nil {
		util.LogError.Printf("env parsing as URL failed, %v", err)
	}
	customHTTPClient := &http.Client{
		Timeout: time.Duration(5) * time.Second,
	}
	util.DefaultConsulClient = util.NewConsulClientWithCustomHTTP(u, customHTTPClient)

	setRoute(app)
	initDB()
	app.StaticWeb("/Files", util.GetMountDir())

	// wsEndpoint := fmt.Sprintf("%s/v%d/%s", constant["API_PREFIX"], constant["API_VERSION"], "realtime")
	// setupWebsocket(wsEndpoint, app)

	serverConfig = util.GetEnvOf("server")
	if port, ok := serverConfig["PORT"]; ok {
		app.Run(iris.Addr(":"+port), iris.WithoutVersionChecker)
	} else {
		app.Run(iris.Addr(":8181"), iris.WithoutVersionChecker)
	}

}

// checkPrivilege will call auth api to check user's privilege of this API
func checkPrivilege(ctx context.Context) {
	paths := strings.Split(ctx.Path(), "/")
	module := paths[3]
	cmd := paths[4]

	appid := ctx.GetHeader("Authorization")
	userid := ctx.GetHeader("X-UserID")

	if len(appid) == 0 || len(userid) == 0 {
		util.LogTrace.Printf("Unauthorized appid:[%s] userid:[%s]", appid, userid)
		ctx.StatusCode(iris.StatusUnauthorized)
		ctx.Skip()
	}

	if checkPrivilegeWithAPI(module, cmd, appid, userid) {
		ctx.Next()
	}
	ctx.Skip()
}

func checkPrivilegeWithAPI(module string, cmd string, appid string, userid string) bool {
	if serverConfig == nil {
		return true
	}

	if authURL, ok := serverConfig["AUTH_URL"]; ok {
		util.LogTrace.Printf(authURL)

		req := make(map[string]string)
		req["userid"] = userid
		req["module"] = module
		req["cmd"] = cmd
		util.LogTrace.Printf("Check privilege: %#v", req)
		// resp = util.HTTPGet(authURL)

		return true
	}

	return true
}

func clientNoStoreCache(ctx context.Context) {
	ctx.Header("Cache-Control", "no-store, private")
}

func setRoute(app *iris.Application) {
	modules := []interface{}{
		Dictionary.ModuleInfo,
		Switch.ModuleInfo,
		Robot.ModuleInfo,
		Stats.ModuleInfo,
		QA.ModuleInfo,
		FAQ.ModuleInfo,
		QA.TestModuleInfo,
		SelfLearning.ModuleInfo,
		imagesManager.ModuleInfo,
	}

	for _, module := range modules {
		info := module.(util.ModuleInfo)
		for _, entrypoint := range info.EntryPoints {
			// entry will be api/v_/<module>/<entry>
			entryPath := fmt.Sprintf("%s/v%d/%s/%s", constant["API_PREFIX"], constant["API_VERSION"], info.ModuleName, entrypoint.EntryPath)
			if app.Handle(entrypoint.AllowMethod, entryPath, checkPrivilege, entrypoint.Callback, clientNoStoreCache) == nil {
				util.LogInfo.Printf("Add route for %s (%s) fail", entryPath, entrypoint.AllowMethod)
			} else {
				util.LogInfo.Printf("Add route for %s (%s) success", entryPath, entrypoint.AllowMethod)
			}
		}
	}

	// Entry for ui setting do not has to check privilege
	info := UI.ModuleInfo
	for _, entrypoint := range info.EntryPoints {
		// entry will be api/v_/<module>/<entry>
		entryPath := fmt.Sprintf("%s/v%d/%s/%s", constant["API_PREFIX"], constant["API_VERSION"], info.ModuleName, entrypoint.EntryPath)
		if app.Handle(entrypoint.AllowMethod, entryPath, entrypoint.Callback) == nil {
			util.LogInfo.Printf("Add route for %s (%s) fail", entryPath, entrypoint.AllowMethod)
		} else {
			util.LogInfo.Printf("Add route for %s (%s) success", entryPath, entrypoint.AllowMethod)
		}
	}
}

func getServerEnv(key string) string {
	envs := util.GetEnvOf("server")
	if envs != nil {
		if val, ok := envs[key]; ok {
			return val
		}
	}
	return ""
}

func initDB() {
	url := getServerEnv("MYSQL_URL")
	user := getServerEnv("MYSQL_USER")
	pass := getServerEnv("MYSQL_PASS")
	db := getServerEnv("MYSQL_DB")
	util.InitMainDB(url, user, pass, db)

	url = getServerEnv("AUDIT_MYSQL_URL")
	user = getServerEnv("AUDIT_MYSQL_USER")
	pass = getServerEnv("AUDIT_MYSQL_PASS")
	db = getServerEnv("AUDIT_MYSQL_DB")
	util.InitAuditDB(url, user, pass, db)

	Stats.InitDB()

	SelfLearning.InitDB()
	imagesManager.InitDB()
	imagesManager.InitDaemon()
}

func setupWebsocket(endpoint string, app *iris.Application) {
	ws := websocket.Store()

	setRealtimeEvent(&ws)
	app.Any(endpoint, ws.Handler())
}

func setRealtimeEvent(ws *websocket.Server) {
	modules := []interface{}{
		FAQ.RealtimeEvents,
	}

	for _, module := range modules {
		events := module.([]websocket.RealtimeEvent)
		for _, event := range events {
			ws.On(event.Event, event.Handle)
		}
	}
}

