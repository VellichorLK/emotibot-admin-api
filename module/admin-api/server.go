package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"emotibot.com/emotigo/module/admin-api/Dictionary"
	"emotibot.com/emotigo/module/admin-api/FAQ"
	"emotibot.com/emotigo/module/admin-api/QA"
	"emotibot.com/emotigo/module/admin-api/Robot"
	"emotibot.com/emotigo/module/admin-api/Stats"
	"emotibot.com/emotigo/module/admin-api/Switch"
	"emotibot.com/emotigo/module/admin-api/Task"
	"emotibot.com/emotigo/module/admin-api/UI"
	"emotibot.com/emotigo/module/admin-api/util"
)

// constant define all const used in server
var constant = map[string]interface{}{
	"API_PREFIX":  "api",
	"API_VERSION": 1,
}

var serverConfig map[string]string
var logChannel chan util.AccessLog

func init() {
	util.LogInit("ADMIN", os.Stdout, os.Stdout, os.Stdout, os.Stdout)
	if len(os.Args) > 1 {
		err := util.LoadConfigFromFile(os.Args[1])
		if err != nil {
			util.LogError.Printf(err.Error())
			os.Exit(-1)
		}
	}
}

func logAvailablePath(router *mux.Router) {
	router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		methods, _ := route.GetMethods()
		pathTemplate, _ := route.GetPathTemplate()
		if len(methods) == 0 {
			methods = []string{"ANY"}
		}
		util.LogInfo.Printf("%6s ROUTE: %s\n", methods, pathTemplate)

		return nil
	})
}

func main() {
	//Init Consul Client
	serverEnvs := util.GetEnvOf("server")
	consulAddr, ok := serverEnvs["CONSUL_URL"]
	if !ok {
		util.LogError.Printf("Can not init without server env:'CONSUL_URL env\n'")
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
	util.LogTrace.Printf("Init consul client with url: %#v\n", u)

	logLevel, ok := serverEnvs["LOG_LEVEL"]
	if !ok {
		logLevel = "INFO"
	}

	accessLog := serverEnvs["ACCESS_LOG"]
	if accessLog == "1" {
		logChannel = make(chan util.AccessLog)
		util.InitAccessLog(logChannel)
	}

	util.SetLogLevel(logLevel)
	router := setRoute()
	initDB()
	logAvailablePath(router)

	serverConfig = util.GetEnvOf("server")
	if port, ok := serverConfig["PORT"]; ok {
		http.ListenAndServe(":"+port, router)
	} else {
		http.ListenAndServe(":8181", router)
	}

}

// checkPrivilege will call auth api to check user's privilege of this API
func checkPrivilege(r *http.Request, ep util.EntryPoint) bool {
	paths := strings.Split(r.URL.Path, "/")
	module := paths[3]
	cmd := paths[4]

	if len(ep.Command) == 0 {
		util.LogTrace.Printf("Path: %s need no auth check\n", ep.EntryPath)
		return true
	}

	appid := util.GetAppID(r)
	userid := util.GetUserID(r)
	token := util.GetAuthToken(r)

	util.LogInfo.Printf("appid: %s, userid: %s\n", appid, userid)
	if len(userid) == 0 || !util.IsValidAppID(appid) || len(token) == 0 {
		util.LogTrace.Printf("Unauthorized path[%s] appid:[%s] userid:[%s] token:[%s]", ep.EntryPath, appid, userid, token)
		return false
	}

	return checkPrivilegeWithAPI(module, cmd, token)
}

func checkPrivilegeWithAPI(module string, cmd string, token string) bool {
	if serverConfig == nil {
		return true
	}

	if authURL, ok := serverConfig["AUTH_URL"]; ok {
		req := make(map[string]string)
		req["module"] = module
		req["cmd"] = cmd

		resp, err := util.HTTPGetSimple(fmt.Sprintf("%s/%s", authURL, token))
		if err != nil {
			util.LogTrace.Printf("Get content resp:%s\n", err.Error())
		}
		util.LogTrace.Printf("%s", resp)

		return true
	}

	return true
}

func clientNoStoreCache(w http.ResponseWriter) {
	w.Header().Add("Cache-Control", "no-store, private")
}

func setRoute() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	modules := []interface{}{
		Dictionary.ModuleInfo,
		Switch.ModuleInfo,
		Robot.ModuleInfo,
		QA.ModuleInfo,
		FAQ.ModuleInfo,
		QA.TestModuleInfo,
		Task.ModuleInfo,
		Stats.ModuleInfo,
	}

	for _, module := range modules {
		info := module.(util.ModuleInfo)
		for idx := range info.EntryPoints {
			entrypoint := info.EntryPoints[idx]
			// entry will be api/v_/<module>/<entry>
			entryPath := fmt.Sprintf("/%s/v%d/%s/%s", constant["API_PREFIX"], entrypoint.Version, info.ModuleName, entrypoint.EntryPath)
			router.
				Methods(entrypoint.AllowMethod).
				Path(entryPath).
				Name(entrypoint.EntryPath).
				HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					defer logHandleRuntime(w, r)()
					clientNoStoreCache(w)
					if checkPrivilege(r, entrypoint) {
						entrypoint.Callback(w, r)
					} else {
						http.Error(w, "Unauthorized", http.StatusUnauthorized)
					}
				})
		}
	}

	// Entry for routes has not to check privilege
	info := UI.ModuleInfo
	for idx := range info.EntryPoints {
		entrypoint := info.EntryPoints[idx]
		// entry will be api/v_/<module>/<entry>
		entryPath := fmt.Sprintf("/%s/v%d/%s/%s", constant["API_PREFIX"], constant["API_VERSION"], info.ModuleName, entrypoint.EntryPath)
		router.
			Path(entryPath).
			Methods(entrypoint.AllowMethod).
			Name(entrypoint.EntryPath).
			HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				clientNoStoreCache(w)
				entrypoint.Callback(w, r)
			})
	}
	router.PathPrefix("/Files/").Handler(http.StripPrefix("/Files/", http.FileServer(http.Dir(util.GetMountDir()))))
	router.HandleFunc("/_health_check", func(w http.ResponseWriter, r *http.Request) {
		// A very simple health check.
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, "ok")
	})
	return router
}

func logHandleRuntime(w http.ResponseWriter, r *http.Request) func() {
	now := time.Now()
	return func() {
		util.LogInfo.Printf("REQ: [%s] [%.3fs][%s@%s]",
			r.RequestURI, time.Since(now).Seconds(), util.GetUserID(r), util.GetAppID(r))
		if logChannel != nil {
			logChannel <- util.AccessLog{
				Path:   r.RequestURI,
				Time:   time.Since(now).Seconds(),
				UserID: util.GetUserID(r),
				AppID:  util.GetAppID(r),
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

	// Stats.InitDB()
}
