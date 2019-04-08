package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
	"time"

	"emotibot.com/emotigo/module/admin-api/util/localemsg"

	"emotibot.com/emotigo/pkg/misc/emotijwt"

	"github.com/gorilla/mux"
	"github.com/robfig/cron"

	"emotibot.com/emotigo/module/admin-api/BF"
	"emotibot.com/emotigo/module/admin-api/ELKStats"
	"emotibot.com/emotigo/module/admin-api/FAQ"
	"emotibot.com/emotigo/module/admin-api/QA"
	"emotibot.com/emotigo/module/admin-api/Robot"
	"emotibot.com/emotigo/module/admin-api/Service"
	"emotibot.com/emotigo/module/admin-api/Stats"
	"emotibot.com/emotigo/module/admin-api/Switch"
	"emotibot.com/emotigo/module/admin-api/System"
	"emotibot.com/emotigo/module/admin-api/Task"
	"emotibot.com/emotigo/module/admin-api/UI"
	"emotibot.com/emotigo/module/admin-api/auth"
	"emotibot.com/emotigo/module/admin-api/autofill"
	"emotibot.com/emotigo/module/admin-api/clustering"
	"emotibot.com/emotigo/module/admin-api/dictionary"
	"emotibot.com/emotigo/module/admin-api/feedback"
	"emotibot.com/emotigo/module/admin-api/integration"
	"emotibot.com/emotigo/module/admin-api/intentengine"
	"emotibot.com/emotigo/module/admin-api/intentengineTest"
	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/admin-api/util/audit"
	"emotibot.com/emotigo/module/admin-api/util/elasticsearch"
	"emotibot.com/emotigo/module/admin-api/util/requestheader"
	"emotibot.com/emotigo/module/admin-api/util/solr"
	"emotibot.com/emotigo/module/admin-api/util/validate"
	"emotibot.com/emotigo/pkg/logger"
	"emotibot.com/emotigo/pkg/misc/emotibothttpwriter"
	"emotibot.com/emotigo/module/admin-api/CustomChat"
)

// constant define all const used in server
var constant = map[string]interface{}{
	"API_PREFIX": "api",
}

var (
	VERSION    string
	BUILD_TIME string
	GO_VERSION string
)

var initErrors = []error{}

var modules = []*util.ModuleInfo{
	&dictionary.ModuleInfo,
	&Switch.ModuleInfo,
	&Robot.ModuleInfo,
	&QA.ModuleInfo,
	&FAQ.ModuleInfo,
	&QA.TestModuleInfo,
	&Task.ModuleInfo,
	&Stats.ModuleInfo,
	&UI.ModuleInfo,
	&clustering.ModuleInfo,
	&System.ModuleInfo,
	&BF.ModuleInfo,
	&intentengine.ModuleInfo,
	&intentengineTest.ModuleInfo,
	&ELKStats.ModuleInfo,
	&Service.ModuleInfo,
	&integration.ModuleInfo,
	&feedback.ModuleInfo,
	&CustomChat.ModuleInfo,
}

var serverConfig map[string]string
var logChannel chan util.AccessLog

func initConfig() {
	if len(os.Args) > 1 {
		err := util.LoadConfigFromFile(os.Args[1])
		if err != nil {
			logger.Error.Printf(err.Error())
			os.Exit(-1)
		}
	} else {
		err := util.LoadConfigFromOSEnv()
		if err != nil {
			logger.Error.Printf(err.Error())
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
		logger.Info.Printf("%6s ROUTE: %s\n", methods, pathTemplate)

		return nil
	})
}

func init() {
	initConfig()
	runtime.GOMAXPROCS(runtime.NumCPU())
	logger.Info.Printf("Set GOMAXPROCS to %d\n", runtime.NumCPU())

	serverEnvs := util.GetEnvOf("server")
	logLevel, ok := serverEnvs["LOG_LEVEL"]
	if !ok {
		logLevel = "INFO"
	}
	logger.SetLevel(logLevel)
	logger.Info.Printf("Set log level %s\n", logLevel)
	initConsul()
	initDB()

	accessLog := serverEnvs["ACCESS_LOG"]
	if accessLog == "1" {
		logChannel = make(chan util.AccessLog)
		util.InitAccessLog(logChannel)
	}

	err := initElasticsearch()
	if err != nil {
		logger.Error.Println("Init elastic search failed:", err.Error())
		initErrors = append(initErrors, err)
	}

	err = ELKStats.Init()
	if err != nil {
		logger.Error.Println("Init ELKStats module failed: ", err.Error())
		initErrors = append(initErrors, err)
	}

	err = clustering.Init()
	if err != nil {
		logger.Error.Println("Init Clustering module failed: ", err.Error())
		initErrors = append(initErrors, err)
	}

	err = initSolr()
	if err != nil {
		logger.Error.Println("Init solr failed:", err.Error())
		initErrors = append(initErrors, err)
	}

	err = Service.Init()
	if err != nil {
		logger.Error.Println("Init service module failed: ", err.Error())
		initErrors = append(initErrors, err)
	}

	autofill.Init()
	feedback.SetupDB(util.GetMainDB())
}

func main() {
	router := setRoute()
	logAvailablePath(router)

	serverConfig = util.GetEnvOf("server")
	serverURL := "0.0.0.0:8181"
	if port, ok := serverConfig["PORT"]; ok {
		serverURL = "0.0.0.0:" + port
	}

	go runOnetimeJob()
	go setupCron()

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

	fillUserAppIDFromHeader(r)

	token := requestheader.GetAuthToken(r)
	appid := requestheader.GetAppID(r)
	userid := requestheader.GetUserID(r)

	logger.Trace.Printf("requester: [%s]@[%s]\n", userid, appid)

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
		return checkAuthHeader(appid, token) && checkPrivilegeWithAPI(module, cmd, token)
	}
	return true
}

func fillUserAppIDFromHeader(r *http.Request) error {
	token := requestheader.GetAuthToken(r)
	// Hardcode for debug
	if token == "Bearer EMOTIBOTDEBUGGER" {
		requestheader.SetUserID(r, "DEBUGGER")
		return nil
	}
	params := strings.Split(token, " ")
	if len(params) != 2 {
		return fmt.Errorf("Token format error: %s", token)
	}

	switch params[0] {
	case "Bearer":
		jwt := params[1]
		userInfo, err := emotijwt.ResolveJWTToken(jwt)
		if err != nil {
			logger.Error.Println("Resolve jwt fail:", err.Error())
			return err
		}
		b, err := json.Marshal(userInfo)
		if err != nil {
			logger.Error.Println("Resolve jwt custom to json fail:", err.Error())
			return err
		}
		userObj := JWTUser{}
		err = json.Unmarshal(b, &userObj)
		if err != nil {
			logger.Error.Println("Resolve jwt custom from json fail:", err.Error())
			return err
		}
		requestheader.SetUserID(r, userObj.ID)
	case "Api":
		apiKeyEnterprise, apiKeyAppid, err := auth.GetAppOwner(params[1])
		if err != nil {
			logger.Error.Println("Get appid from apikey fail:", err.Error())
			return err
		}
		requestheader.SetEnterprise(r, apiKeyEnterprise)
		requestheader.SetAppID(r, apiKeyAppid)
		if apiKeyAppid != "" {
			requestheader.SetUserID(r, fmt.Sprintf("%s API", apiKeyAppid))
		} else if apiKeyEnterprise != "" {
			requestheader.SetUserID(r, fmt.Sprintf("%s API", apiKeyEnterprise))
		} else {
			requestheader.SetUserID(r, "System API")

		}
	}
	return nil
}

func checkAuthHeader(appid, token string) bool {
	// Hardcode for debug
	if token == "Bearer EMOTIBOTDEBUGGER" {
		return true
	}
	params := strings.Split(token, " ")
	if len(params) != 2 {
		logger.Error.Println("Token format error:", token)
		return false
	}
	switch params[0] {
	case "Bearer":
		jwt := params[1]
		_, err := emotijwt.ResolveJWTToken(jwt)
		if err != nil {
			logger.Error.Println("Resolve jwt fail:", err.Error())
			return false
		}
		return true
	case "Api":
		_, apiKeyAppid, err := auth.GetAppOwner(params[1])
		if err != nil {
			logger.Error.Println("Get appid of apikey err:", err.Error())
			return false
		}
		// TODO: check if appid is in apiKeyEnterprise
		if apiKeyAppid == "" {
			return true
		}
		return apiKeyAppid == appid
	}
	return false
}

func checkPrivilegeWithAPI(module string, cmd string, token string) bool {
	if serverConfig == nil {
		return true
	}

	params := strings.Split(token, " ")
	if len(params) != 2 {
		logger.Error.Println("Token format error:", token)
		return false
	}
	if params[0] == "Api" {
		return true
	}

	if authURL, ok := serverConfig["AUTH_URL"]; ok {
		req := make(map[string]string)
		req["module"] = module
		req["cmd"] = cmd

		_, err := util.HTTPGetSimple(fmt.Sprintf("%s/%s", authURL, params[1]))
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
				HandlerFunc(func(origWriter http.ResponseWriter, r *http.Request) {
					w := emotibothttpwriter.New(origWriter)

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
		if info.EntryPrefix == nil {
			continue
		}
		for idx := range info.EntryPrefix {
			entrypoint := info.EntryPrefix[idx]
			// entry will be api/v_/<module>/<entry>
			entryPath := fmt.Sprintf("/%s/v%d/%s/%s", constant["API_PREFIX"],
				entrypoint.Version, info.ModuleName, entrypoint.EntryPath)
			router.
				Methods(entrypoint.AllowMethod).
				PathPrefix(entryPath).
				Name(entrypoint.EntryPath).
				HandlerFunc(func(origWriter http.ResponseWriter, r *http.Request) {
					w := emotibothttpwriter.New(origWriter)

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
	router.PathPrefix("/Files/").Methods("GET").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		locale := requestheader.GetLocale(r)
		if locale == "" {
			locale = localemsg.ZhCn
		}
		handler := http.FileServer(http.Dir(util.GetMountDir()))

		if newPath := strings.TrimPrefix(r.URL.Path, "/Files/"); len(newPath) < len(r.URL.Path) {
			newRequest := new(http.Request)
			*newRequest = *r
			newRequest.URL = new(url.URL)
			*newRequest.URL = *r.URL
			newRequest.URL.Path = fmt.Sprintf("%s/%s", locale, newPath)
			logger.Trace.Printf("Get file from %s\n", newRequest.URL.Path)

			if strings.HasSuffix(newPath, ".xlsx") {
				w.Header().Set("Content-Type", "application/octet-stream")
			}

			handler.ServeHTTP(w, newRequest)
		} else {
			http.NotFound(w, r)
		}
	})
	router.HandleFunc(fmt.Sprintf("/%s/_health_check", constant["API_PREFIX"]), func(w http.ResponseWriter, r *http.Request) {
		// A very simple health check.
		ret := healthCheck()
		util.WriteJSON(w, ret)
	})
	router.HandleFunc(fmt.Sprintf("/%s/_info", constant["API_PREFIX"]), func(w http.ResponseWriter, r *http.Request) {
		ret := map[string]string{
			"VERSION":    VERSION,
			"BUILD_TIME": BUILD_TIME,
			"GO_VERSION": GO_VERSION,
		}
		util.WriteJSON(w, ret)
	})
	return router
}

func logHandleRuntime(w *emotibothttpwriter.EmotibotHTTPWriter, r *http.Request) func() {
	now := time.Now()
	return func() {
		code := w.GetStatusCode()
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

	err = Stats.InitDB()
	if err != nil {
		logger.Error.Println("Init stats db fail")
		initErrors = append(initErrors, err)
	}
	err = auth.InitDB()
	if err != nil {
		logger.Error.Println("Init auth db fail")
		initErrors = append(initErrors, err)
	}
}

func initElasticsearch() (err error) {
	host := getServerEnv("ELASTICSEARCH_HOST")
	port := getServerEnv("ELASTICSEARCH_PORT")
	basicAuthUsername := getServerEnv("ELASTICSEARCH_BASIC_AUTH_USERNAME")
	basicAuthPassword := getServerEnv("ELASTICSEARCH_BASIC_AUTH_PASSWORD")

	if host == "" {
		return errors.New("ELASTICSEARCH_HOST env missing")
	}

	if port == "" {
		return errors.New("ELASTICSEARCH_PORT env missing")
	}

	if basicAuthUsername == "" {
		return errors.New("ELASTICSEARCH_BASIC_AUTH_USERNAME env missing")
	}

	if basicAuthPassword == "" {
		return errors.New("ELASTICSEARCH_BASIC_AUTH_PASSWORD env missing")
	}

	return elasticsearch.Setup(host, port, basicAuthUsername, basicAuthPassword)
}

func initSolr() (err error) {
	host := getServerEnv("SOLR_HOST")
	port := getServerEnv("SOLR_PORT")

	if host == "" {
		return errors.New("SOLR_HOST env missing")
	}

	if port == "" {
		return errors.New("SOLR_PORT env missing")
	}

	solr.Setup(host, port)
	return nil
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

func setupCron() {
	c := cron.New()
	for _, module := range modules {
		if module.Cronjobs != nil {
			for key, job := range module.Cronjobs {
				funName := fmt.Sprintf("[%s@%s]", key, module.ModuleName)
				err := c.AddFunc(job.Period, func() {
					defer func() {
						if r := recover(); r != nil {
							logger.Error.Printf("Run period func %s error: %s\n", funName, r)
						}
					}()
					now := time.Now()
					logger.Trace.Printf("Exec %s at [%s]\n", funName, now.UTC())
					job.Handler()
				})
				if err != nil {
					logger.Error.Printf("Add cron job %s fail: %s\n", funName, err.Error())
				}
			}
		}
	}
}

func initConsul() {
	//Init Consul Client
	serverEnvs := util.GetEnvOf("server")
	consulURL, ok := serverEnvs["CONSUL_URL"]
	if !ok {
		logger.Error.Printf("Can not init without server env:'CONSUL_URL env\n'")
		os.Exit(-1)
	}
	consulPrefix := serverEnvs["CONSUL_PREFIX"]

	consulAddr := fmt.Sprintf("%s/v1/kv/%s/", strings.TrimRight(consulURL, "/"), consulPrefix)
	consulRootAddr := fmt.Sprintf("%s/v1/kv/", strings.TrimRight(consulURL, "/"))

	u, err := url.Parse(consulAddr)
	if err != nil {
		logger.Error.Printf("env parsing as URL failed, %v", err)
	}
	rootURL, err := url.Parse(consulRootAddr)
	if err != nil {
		logger.Error.Printf("env parsing as URL failed, %v", err)
	}

	customHTTPClient := &http.Client{
		Timeout: time.Duration(5) * time.Second,
	}
	rootHTTPClient := &http.Client{
		Timeout: time.Duration(5) * time.Second,
	}

	util.DefaultConsulClient = util.NewConsulClientWithCustomHTTP(u, customHTTPClient)
	util.RootConsulClient = util.NewConsulClientWithCustomHTTP(rootURL, rootHTTPClient)
	logger.Info.Printf("Init consul client with url: %s\n", u.String())
	logger.Info.Printf("Init root consul client with url: %s\n", rootURL.String())
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

type JWTUser struct {
	ID       string `json:"id"`
	UserName string `json:"user_name"`
	Type     int    `json:"type"`
}
