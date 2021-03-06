package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"emotibot.com/emotigo/module/admin-api/util"
)

// constant define all const used in server
var constant = map[string]interface{}{
	"API_PREFIX": "api",
}

var modules = []*util.ModuleInfo{
}

var serverConfig map[string]string
var logChannel chan util.AccessLog

func init() {
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
	serverEnvs := util.GetEnvOf("server")
	logLevel, ok := serverEnvs["LOG_LEVEL"]
	if !ok {
		logLevel = "INFO"
	}
	util.LogInfo.Printf("Set log level %s\n", logLevel)

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
	serverURL := "0.0.0.0:8181"
	if port, ok := serverConfig["PORT"]; ok {
		serverURL = "0.0.0.0:" + port
	}

	go runOnetimeJob()

	util.LogInfo.Println("Start server on", serverURL)
	err := http.ListenAndServe(serverURL, router)
	if err != nil {
		util.LogError.Println("Start server fail: ", err.Error())
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

	if len(userid) == 0 {
		util.LogTrace.Printf("Unauthorized path[%s]: empty user", ep.EntryPath)
		return false
	}
	if ep.CheckAppID && !util.IsValidAppID(appid) {
		util.LogTrace.Printf("Unauthorized path[%s]: appid[%s]", ep.EntryPath, appid)
		return false
	}
	if ep.CheckAuthToken {
		if len(token) == 0 {
			util.LogTrace.Printf("Unauthorized path[%s]: empty token", ep.EntryPath)
			return false
		}
		return checkPrivilegeWithAPI(module, cmd, token)
	}

	return true
}

func checkPrivilegeWithAPI(module string, cmd string, token string) bool {
	if serverConfig == nil {
		return true
	}

	if authURL, ok := serverConfig["AUTH_URL"]; ok {
		req := make(map[string]string)
		req["module"] = module
		req["cmd"] = cmd

		_, err := util.HTTPGetSimple(fmt.Sprintf("%s/%s", authURL, token))
		if err != nil {
			util.LogTrace.Printf("Get content resp:%s\n", err.Error())
		}

		return true
	}

	return true
}

func clientNoStoreCache(w http.ResponseWriter) {
	w.Header().Add("Cache-Control", "no-store, private")
}

func setRoute() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	for idx := range modules {
		info := modules[idx]
		info.SetEnvironments(util.GetEnvOf(strings.ToLower(info.ModuleName)))

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
						r.Header.Set(util.AuditCustomHeader, info.ModuleName)
						entrypoint.Callback(w, r)
					} else {
						http.Error(w, "Unauthorized", http.StatusUnauthorized)
					}
				})
		}
	}
	router.PathPrefix("/Files/").Methods("GET").Handler(http.StripPrefix("/Files/", http.FileServer(http.Dir(util.GetMountDir()))))
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
		code, err := strconv.Atoi(w.Header().Get("X-Status"))
		if err != nil {
			// TODO: use custom responseWriter to get return http status
			// For now, use X-Status in header to do log
			// if header not set X-Status, default is 200
			code = http.StatusOK
		}

		requestIP := r.Header.Get("X-Real-IP")
		if requestIP == "" {
			requestIP = r.RemoteAddr
		}
		// util.LogInfo.Printf("REQ: [%s][%d] [%.3fs][%s@%s]",
		// 	r.RequestURI, code, time.Since(now).Seconds(), util.GetUserID(r), util.GetAppID(r))
		if logChannel != nil {
			logChannel <- util.AccessLog{
				Path:       r.RequestURI,
				Time:       time.Since(now).Seconds(),
				UserID:     util.GetUserID(r),
				UserIP:     requestIP,
				AppID:      util.GetAppID(r),
				StatusCode: code,
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

func runOnetimeJob() {
	for _, module := range modules {
		if module.OneTimeFunc != nil {
			for key, fun := range module.OneTimeFunc {
				util.LogInfo.Printf("Run func %s of module %s\n", key, module.ModuleName)
				func() {
					defer func() {
						if r := recover(); r != nil {
							util.LogError.Printf("Run func %s error: %s\n", key, r)
						}
					}()
					fun()
				}()
			}
		}
	}
}
