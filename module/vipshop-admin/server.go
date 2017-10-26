package main

import (
	"fmt"
	"os"
	"strings"

	"emotibot.com/emotigo/module/vipshop-admin/Dictionary"
	"emotibot.com/emotigo/module/vipshop-admin/Switch"
	"emotibot.com/emotigo/module/vipshop-admin/util"

	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
)

// constant define all const used in server
var constant = map[string]interface{}{
	"API_PREFIX":  "api",
	"API_VERSION": 1,
}

var serverConfig map[string]string

func main() {
	app := iris.New()

	// util.LogInit(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)
	util.LogInit(os.Stderr, os.Stdout, os.Stdout, os.Stderr)
	if len(os.Args) > 1 {
		err := util.LoadConfigFromFile(os.Args[1])
		if err != nil {
			util.LogError.Printf(err.Error())
			os.Exit(-1)
		}
	}

	setRoute(app)
	initDB()

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

func setRoute(app *iris.Application) {
	modules := []interface{}{
		Dictionary.ModuleInfo,
		Switch.ModuleInfo,
	}

	for _, module := range modules {
		info := module.(util.ModuleInfo)
		for _, entrypoint := range info.EntryPoints {
			// entry will be api/v_/<module>/<entry>
			entryPath := fmt.Sprintf("%s/v%d/%s/%s", constant["API_PREFIX"], constant["API_VERSION"], info.ModuleName, entrypoint.EntryPath)
			if app.Handle(entrypoint.AllowMethod, entryPath, checkPrivilege, entrypoint.Callback) == nil {
				util.LogInfo.Printf("Add route for %s (%s) fail", entryPath, entrypoint.AllowMethod)
			} else {
				util.LogInfo.Printf("Add route for %s (%s) success", entryPath, entrypoint.AllowMethod)
			}
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
}
