package main

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"emotibot.com/emotigo/module/admin-api/auth"
	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/admin-api/util/audit"
	"emotibot.com/emotigo/module/admin-api/util/requestheader"
	"emotibot.com/emotigo/module/admin-api/util/validate"
	"emotibot.com/emotigo/module/qic-api/cu"
	"emotibot.com/emotigo/pkg/config/v1"
	"emotibot.com/emotigo/pkg/logger"
)

// constant define all const used in server
var constant = map[string]interface{}{
	"API_PREFIX": "api",
}

var initErrors = []error{}

var modules = []*util.ModuleInfo{
	&auth.ModuleInfo,
	&cu.ModuleInfo,
}

var serverConfig map[string]string
var logChannel chan util.AccessLog

func logAvailablePath(router *mux.Router) {
	router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		methods, _ := route.GetMethods()
		pathTemplate, _ := route.GetPathTemplate()
		if len(methods) == 0 {
			methods = []string{"ANY"}
		}
		logger.Info.Printf("%6s ROUTE: %s\n", methods, pathTemplate)

		return nil
	})
}

// serverInitial init a serial steps for api server, including config and GOMAXPROCS
func serverInitial() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	var err error
	if len(os.Args) > 1 {
		err = util.LoadConfigFromFile(os.Args[1])
		if err != nil {
			logger.Error.Printf(err.Error())
			os.Exit(-1)
		}
	} else {
		err = util.LoadConfigFromOSEnv()
		if err != nil {
			logger.Error.Printf(err.Error())
			os.Exit(-1)
		}
	}

	logger.Info.Printf("Set GOMAXPROCS to %d\n", runtime.NumCPU())
	serverEnvs := config.GetEnvOf("server")
	logLevel, ok := serverEnvs["LOG_LEVEL"]
	if !ok {
		logLevel = "INFO"
	}
	logger.SetLevel(logLevel)
	logger.Info.Printf("Set log level %s\n", logLevel)

	accessLog := serverEnvs["ACCESS_LOG"]
	if accessLog == "1" {
		logChannel = make(chan util.AccessLog)
		util.InitAccessLog(logChannel)
	}
}

func main() {
	serverInitial()
	router := setRoute()
	initDB()
	logAvailablePath(router)

	serverConfig = util.GetEnvOf("server")
	serverURL := "0.0.0.0:8181"
	if port, ok := serverConfig["PORT"]; ok {
		serverURL = "0.0.0.0:" + port
	}

	go runOnetimeJob()

	logger.Info.Println("Start server on", serverURL)
	err := http.ListenAndServe(serverURL, router)
	if err != nil {
		logger.Error.Println("Start server fail: ", err.Error())
	}
}

// checkPrivilege will call auth api to check user's privilege of this API
func checkPrivilege(r *http.Request, ep util.EntryPoint) bool {
	paths := strings.Split(r.URL.Path, "/")
	module := paths[3]
	cmd := paths[4]

	if len(ep.Command) == 0 {
		logger.Trace.Printf("Path: %s need no auth check\n", ep.EntryPath)
		return true
	}

	appid := requestheader.GetAppID(r)
	userid := requestheader.GetUserID(r)
	token := requestheader.GetAuthToken(r)

	if len(userid) == 0 {
		logger.Trace.Printf("Unauthorized path[%s]: empty user", ep.EntryPath)
		return false
	}
	if ep.CheckAppID && !validate.IsValidAppID(appid) {
		logger.Trace.Printf("Unauthorized path[%s]: appid[%s]", ep.EntryPath, appid)
		return false
	}
	if ep.CheckAuthToken {
		if len(token) == 0 {
			logger.Trace.Printf("Unauthorized path[%s]: empty token", ep.EntryPath)
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
			logger.Trace.Printf("Get content resp:%s\n", err.Error())
		}

		return true
	}

	return true
}

func clientNoStoreCache(w http.ResponseWriter) {
	w.Header().Add("Cache-Control", "no-store, private")
}

//setRoute create a mux.Router based on modules
func setRoute() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	for idx := range modules {
		info := modules[idx]
		info.SetEnvironments(util.GetEnvOf(strings.ToLower(info.ModuleName)))

		for idx := range info.EntryPoints {
			entrypoint := info.EntryPoints[idx]
			// entry will be api/v_/<module>/<entry>
			entryPath := fmt.Sprintf("/%s/v%d/%s/%s", constant["API_PREFIX"],
				entrypoint.Version, info.ModuleName, entrypoint.EntryPath)
			router.
				Methods(entrypoint.AllowMethod).
				Path(entryPath).
				Name(entrypoint.EntryPath).
				HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					defer logHandleRuntime(w, r)()
					clientNoStoreCache(w)
					defer func() {
						if err := recover(); err != nil {
							errMsg := fmt.Sprintf("%#v", err)
							util.WriteWithStatus(w, errMsg, http.StatusInternalServerError)
							util.PrintRuntimeStack(10)
							logger.Error.Println("Panic error:", errMsg)
						}
					}()

					if checkPrivilege(r, entrypoint) {
						r.Header.Set(audit.AuditCustomHeader, info.ModuleName)
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
		ret := healthCheck()
		util.WriteJSON(w, ret)
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
		// logger.Info.Printf("REQ: [%s][%d] [%.3fs][%s@%s]",
		// 	r.RequestURI, code, time.Since(now).Seconds(), requestheader.GetUserID(r), requestheader.GetAppID(r))
		if logChannel != nil {
			logChannel <- util.AccessLog{
				Path:         r.RequestURI,
				Time:         time.Since(now).Seconds(),
				UserID:       requestheader.GetUserID(r),
				UserIP:       requestIP,
				AppID:        requestheader.GetAppID(r),
				EnterpriseID: requestheader.GetEnterpriseID(r),
				StatusCode:   code,
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
	err := util.InitMainDB(url, user, pass, db)
	if err != nil {
		logger.Error.Println("Init main db fail")
		initErrors = append(initErrors, err)
	}

	url = getServerEnv("AUDIT_MYSQL_URL")
	user = getServerEnv("AUDIT_MYSQL_USER")
	pass = getServerEnv("AUDIT_MYSQL_PASS")
	db = getServerEnv("AUDIT_MYSQL_DB")
	err = util.InitAuditDB(url, user, pass, db)
	if err != nil {
		logger.Error.Println("Init audit db fail")
		initErrors = append(initErrors, err)
	}

	err = auth.InitDB()
	if err != nil {
		logger.Error.Println("Init auth db fail")
		initErrors = append(initErrors, err)
	}

	err = cu.InitDB()
	if err != nil {
		logger.Error.Println("Init cu(qi) db fail")
		initErrors = append(initErrors, err)
	}
}

func runOnetimeJob() {
	for _, module := range modules {
		if module.OneTimeFunc != nil {
			for key, fun := range module.OneTimeFunc {
				logger.Info.Printf("Run func %s of module %s\n", key, module.ModuleName)
				func() {
					defer func() {
						if r := recover(); r != nil {
							logger.Error.Printf("Run func %s error: %s\n", key, r)
						}
					}()
					fun()
				}()
			}
		}
	}
}

type healthResult struct {
	MemoryUsage uint64            `json:"memory"`
	MaxProcess  int               `json:"max_process"`
	NumCPU      int               `json:"num_cpu"`
	InitErrors  []string          `json:"init_errors"`
	DBHealth    map[string]string `json:"db_status"`
}

func healthCheck() *healthResult {
	ret := healthResult{}
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	ret.MemoryUsage = m.Alloc
	ret.MaxProcess = runtime.GOMAXPROCS(0)
	ret.NumCPU = runtime.NumCPU()
	ret.InitErrors = []string{}

	for idx := range initErrors {
		ret.InitErrors = append(ret.InitErrors, initErrors[idx].Error())
	}

	ret.DBHealth = util.GetDBStatus()
	return &ret
}
