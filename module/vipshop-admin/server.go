package main

import (
	"fmt"
	"os"
	"strings"

	"emotibot.com/emotigo/module/vipshop-admin/Dictionary"
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
	initAuditDB()

	serverConfig = util.GetEnvOf("server")
	if port, ok := serverConfig["PORT"]; ok {
		app.Run(iris.Addr(":" + port))
	} else {
		app.Run(iris.Addr(":8181"))
	}

}

// checkPrivilege will call auth api to check user's privilege of this API
func checkPrivilege(ctx context.Context) {
	paths := strings.Split(ctx.Path(), "/")
	module := paths[2]

	appid := ctx.GetHeader("Authorization")
	userid := ctx.GetHeader("X-UserID")

	if checkPrivilegeWithAPI(module, appid, userid) {
		ctx.Next()
	}
	ctx.Skip()
}

func checkPrivilegeWithAPI(module string, appid string, userid string) bool {
	if serverConfig == nil {
		return true
	}

	if authURL, ok := serverConfig["AUTH_URL"]; ok {
		util.LogTrace.Printf(authURL)

		req := make(map[string]string)
		req["userid"] = userid
		req["module"] = module
		// resp = util.HTTPGet(authURL)

		return true
	}

	return true
}

func setRoute(app *iris.Application) {
	modules := []interface{}{
		Dictionary.ModuleInfo,
	}

	for _, module := range modules {
		info := module.(util.ModuleInfo)
		for _, entrypoint := range info.EntryPoints {
			// entry will be api/v_/<module>/<entry>
			entryPath := fmt.Sprintf("%s/v%d/%s/%s", constant["API_PREFIX"], constant["API_VERSION"], info.ModuleName, entrypoint.EntryPath)
			app.Handle(entrypoint.AllowMethod, entryPath, checkPrivilege, entrypoint.Callback)
		}
	}
}

func initDB() {
	initFunc := []func(){
		Dictionary.InitDatabase,
	}

	for _, function := range initFunc {
		function()
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
func initAuditDB() {
	url := getServerEnv("AUDIT_MYSQL_URL")
	user := getServerEnv("AUDIT_MYSQL_USER")
	pass := getServerEnv("AUDIT_MYSQL_PASS")
	db := getServerEnv("AUDIT_MYSQL_DB")
	util.AuditDBInit(url, user, pass, db)
}
