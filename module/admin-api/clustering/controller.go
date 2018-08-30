package clustering

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	statData "emotibot.com/emotigo/module/admin-api/ELKStats/data"
	statService "emotibot.com/emotigo/module/admin-api/ELKStats/services"
	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/admin-api/util/requestheader"
	"emotibot.com/emotigo/pkg/logger"
)

var (
	// ModuleInfo is needed for module define
	ModuleInfo util.ModuleInfo
	//Envs enviroments of clustering module
	Envs map[string]string

	NluURL             string
	responseURL        string
	EarlyStopThreshold int
	MinSizeCluster     int
	MaxNumToCluster    int
	ClusteringBatch    int
)

func init() {
	ModuleInfo = util.ModuleInfo{
		ModuleName: "clusters",
		EntryPoints: []util.EntryPoint{
			util.NewEntryPoint(http.MethodPut, "reports", []string{}, handleNewReport),
			util.NewEntryPoint(http.MethodGet, "reports", []string{}, handleGetReports),
			util.NewEntryPoint(http.MethodGet, "reports/{id}", []string{}, handleGetReport),
		},
	}
}

func checkNeedEnvs() {
	var err error

	ClusteringBatch, err = strconv.Atoi(util.GetEnviroment(Envs, "CLUSTER_BATCH"))
	if err != nil {
		logger.Error.Println(err)
		ClusteringBatch = 50
	}

	EarlyStopThreshold, err = strconv.Atoi(util.GetEnviroment(Envs, "EARLY_STOP_THRESHOLD"))
	if err != nil {
		logger.Error.Println(err)
		EarlyStopThreshold = 3
	}

	MaxNumToCluster, err = strconv.Atoi(util.GetEnviroment(Envs, "MAX_NUM_TO_CLUSTER"))
	if err != nil {
		logger.Error.Println(err)
		MaxNumToCluster = 10000
	}

	MinSizeCluster, err = strconv.Atoi(util.GetEnviroment(Envs, "MIN_SIZE_CLUSTER"))
	if err != nil {
		logger.Error.Println(err)
		MinSizeCluster = 10
	}

	NluURL = util.GetEnviroment(Envs, "NLU_URL")
	if NluURL == "" {
		logger.Error.Println("cant found NLU_URL, use local NLU URL")
		NluURL = "http://172.17.0.1:13901"
	}

	responseURL = util.GetEnviroment(Envs, "RESPONSE_URL")
	if responseURL == "" {
		logger.Error.Println("cant found RESPONSE_URL")

	}
}

//InitDB init the database connection
func InitDB() {
	Envs = util.GetModuleEnvironments(ModuleInfo.ModuleName)
	checkNeedEnvs()

	url := util.GetEnviroment(Envs, "MYSQL_URL")
	user := util.GetEnviroment(Envs, "MYSQL_USER")
	pass := util.GetEnviroment(Envs, "MYSQL_PASS")
	db := util.GetEnviroment(Envs, "MYSQL_DB")

	dao, err := initSelfLearningDB(url, user, pass, db)
	if err != nil {
		logger.Error.Printf("Cannot init self learning db, [%s:%s@%s:%s]: %s\n", user, pass, url, db, err.Error())
		return
	}
	util.SetDB(ModuleInfo.ModuleName, dao)
}

func handleNewReport(w http.ResponseWriter, r *http.Request) {
	appid := requestheader.GetAppID(r)
	_ = appid
	var query statData.RecordQuery
	err := json.NewDecoder(r.Body).Decode(&query)
	if err != nil {
		logger.Warn.Println("PUT /reports: request can not decode")
		http.Error(w, "input format error", http.StatusBadRequest)
		return
	}
	result, err := statService.VisitRecordsQuery(query)
	if err != nil {
		logger.Error.Printf("get records failed, %v", err)
	}
	var report Report
	report, err = service.NewReport(result.Hits)
	if err != nil {
		logger.Error.Printf("PUT /reports: new request error, %v\n", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(report)
}

func handleGetReports(w http.ResponseWriter, r *http.Request) {
	appid := requestheader.GetAppID(r)
	reports := []Report{}
	limit, err := strconv.Atoi(r.FormValue("limit"))
	if err != nil || limit > 10 {
		limit = 10
	}
	query := ReportQuery{}
	_ = appid
	reports, err = service.Reports(query)
	if err == ErrNotAvailable {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Error.Printf("%s\n", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	util.WriteJSON(w, reports)
}

func handleGetReport(w http.ResponseWriter, r *http.Request) {
	reportID := util.GetMuxVar(r, "id")
	if reportID == "" {
		util.WriteWithStatus(w, fmt.Sprintf("id should not be empty"), http.StatusBadRequest)
		return
	}

	status := -999
	appid := requestheader.GetAppID(r)
	if appid == "" {
		util.WriteJSONWithStatus(w,
			util.GenRetObj(status, "No permission"),
			http.StatusBadRequest)
		return
	}

	reports, err := GetReports(reportID, 1, appid, -1)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Error.Printf("%s\n", err)
		return
	}
	if len(reports) == 0 {
		util.WriteWithStatus(w, "", http.StatusNotFound)
		return
	}
	util.WriteJSON(w, reports[0])
}
