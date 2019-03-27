package main

import (
	"fmt"
	"os"

	"emotibot.com/emotigo/module/vipshop-auth/CAS"

	"emotibot.com/emotigo/module/vipshop-admin/util"
	"emotibot.com/emotigo/module/vipshop-auth/CAuth"

	"github.com/kataras/iris"
)

// constant define all const used in server
var constant = map[string]interface{}{
	"API_PREFIX":  "cauth",
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

	initDB()
	setRoute(app)

	serverConfig = util.GetEnvOf("server")
	if port, ok := serverConfig["PORT"]; ok {
		app.Run(iris.Addr(":" + port))
	} else {
		app.Run(iris.Addr(":8786"))
	}

}

func setRoute(app *iris.Application) {
	modules := []interface{}{
		CAuth.ModuleInfo,
		CAS.ModuleInfo,
	}

	for _, module := range modules {
		info := module.(util.ModuleInfo)
		for _, entrypoint := range info.EntryPoints {
			// entry will be api/v_/<module>/<entry>

			entryPath := fmt.Sprintf("%s/v%d/%s/%s", constant["API_PREFIX"], constant["API_VERSION"], info.ModuleName, entrypoint.EntryPath)
			if info.ModuleName == "cas" {
				entryPath = "login"
			}
			if app.Handle(entrypoint.AllowMethod, entryPath, entrypoint.Callback) == nil {
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
	url := getServerEnv("AUDIT_MYSQL_URL")
	user := getServerEnv("AUDIT_MYSQL_USER")
	pass := getServerEnv("AUDIT_MYSQL_PASS")
	db := getServerEnv("AUDIT_MYSQL_DB")
	util.InitAuditDB(url, user, pass, db)
}
