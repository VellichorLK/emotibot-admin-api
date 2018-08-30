package SelfLearning

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"emotibot.com/emotigo/module/admin-api/ApiError"
	ed "emotibot.com/emotigo/module/admin-api/ELKStats/data"
	"emotibot.com/emotigo/module/admin-api/SelfLearning/data"
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
		ModuleName: "selfLearn",
		EntryPoints: []util.EntryPoint{
			util.NewEntryPoint(http.MethodPost, "reports", []string{}, handleClustering),
			util.NewEntryPoint("GET", "reports", []string{}, handleGetReports),
			util.NewEntryPoint("GET", "reports/{id}", []string{}, handleGetReport),
			util.NewEntryPoint("DELETE", "reports/{id}", []string{}, handleDeleteReport),
			util.NewEntryPoint("GET", "reports/{id}/clusters", []string{}, handleGetClusters),
			// util.NewEntryPoint("GET", "userQuestions", []string{}, handleGetUserQuestions),
			// util.NewEntryPoint("POST", "userQuestions", []string{}, handleUpdateUserQuestions),
			// util.NewEntryPoint("GET", "userQuestions/{id}", []string{}, handleGetUserQuestion),
			// util.NewEntryPoint("POST", "userQuestions/{id}", []string{}, handleUpdateUserQuestion),
			// util.NewEntryPoint("POST", "userQuestions/{id}/revoke", []string{}, handleRevokeQuestion),
			// util.NewEntryPoint("POST", "recommend", []string{}, handleRecommend),
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

	if ok := data.InitializeWord2Vec(util.GetEnviroment(Envs, "RESOURCES_PATH")); !ok {
		logger.Error.Println("Load self learning caches failed!")
		return
	}
}

func handleClustering(w http.ResponseWriter, r *http.Request) {
	var query = ed.RecordQuery{}
	status := -999
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logger.Error.Println("request body io error, " + err.Error())
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(data, &query)
	if err != nil {
		logger.Warn.Println("json format failed, " + err.Error())
		http.Error(w, "json format failed", http.StatusBadRequest)
		return
	}
	if query.StartTime != nil && query.EndTime != nil {
		if *query.StartTime > *query.EndTime {
			http.Error(w, "start time should be less than end time", http.StatusBadRequest)
		}
	}

	appid := requestheader.GetAppID(r)
	if appid == "" {
		util.WriteJSONWithStatus(w,
			util.GenRetObj(status, "No permission"),
			http.StatusBadRequest)
		return
	}

	type respID struct {
		ReportID uint64 `json:"report_id"`
	}

	respone := &respID{}
	var isDup bool
	var reportID uint64

	st := time.Unix(*query.StartTime, 0)
	et := time.Unix(*query.EndTime, 0)

	pType, err := getQuestionType(r)
	if err != nil {
		util.WriteJSONWithStatus(w,
			util.GenRetObj(status, err.Error()),
			http.StatusBadRequest)
		return
	}

	isDup, reportID, err = isDuplicate(st, et, appid, pType)

	if err != nil {
		util.WriteJSONWithStatus(w,
			util.GenRetObj(status, err.Error()),
			http.StatusInternalServerError)
		return
	}

	if !isDup {
		reportID, err = createOneReport(st, et, appid, pType)
		if err != nil {
			util.WriteJSONWithStatus(w,
				util.GenRetObj(status, err.Error()),
				http.StatusInternalServerError)
			return
		}

		go doClustering(st, et, reportID, &dbStore{}, appid, pType)
	}

	respone.ReportID = reportID
	util.WriteJSON(w, util.GenRetObj(ApiError.SUCCESS, respone))
	return
}

func getQuestionType(r *http.Request) (int, error) {
	typeStr := r.URL.Query().Get(PType)
	pType, err := strconv.Atoi(typeStr)
	if err != nil {
		return 0, err
	}
	if !isValidType(pType) {
		return 0, errors.New("wrong value of type")
	}
	return pType, nil
}

func handleGetReports(w http.ResponseWriter, r *http.Request) {
	status := -999
	appid := requestheader.GetAppID(r)
	if appid == "" {
		util.WriteJSONWithStatus(w,
			util.GenRetObj(status, "No permission"),
			http.StatusBadRequest)
		return
	}

	pType, err := getQuestionType(r)
	if err != nil {
		util.WriteJSONWithStatus(w,
			util.GenRetObj(status, err.Error()),
			http.StatusBadRequest)
		return
	}

	reports := []Report{}
	limit, err := strconv.Atoi(r.FormValue("limit"))
	if err != nil || limit > 10 {
		limit = 10
	}

	reports, err = GetReports("", limit, appid, pType)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Error.Printf("%s\n", err)
		return
	}
	util.WriteJSON(w, reports)
}

func handleGetReport(w http.ResponseWriter, r *http.Request) {
	// TODO: reportID is string?!
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

func handleGetClusters(w http.ResponseWriter, r *http.Request) {
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
		log.Println(err)
		return
	}
	if len(reports) == 0 {
		util.WriteWithStatus(w, "", http.StatusNotFound)
		return
	}
	clusters, err := GetClusters(reports[0])
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}
	util.WriteJSON(w, clusters)
}

func handleGetUserQuestions(w http.ResponseWriter, r *http.Request) {
	reportID := r.FormValue("reportID")
	if reportID == "" {
		util.WriteWithStatus(w, "", http.StatusBadRequest)
		return
	}
	clusterID := r.FormValue("clusterID")
	var limit, page int
	var err error
	if l := r.FormValue("limit"); l == "" {
		limit = 10
	} else {
		limit, err = strconv.Atoi(l)
	}
	if err != nil {
		logger.Error.Printf("input [limit] can't parseInt. %s\n", err)
		util.WriteWithStatus(w, fmt.Sprintf("input [limit] can't parseInt. %s\n", err), http.StatusBadRequest)
		return
	}

	if p := r.FormValue("page"); p == "" {
		page = 0
	} else {
		page, err = strconv.Atoi(p)
	}
	if err != nil {
		logger.Error.Printf("input [page] can't parseInt. %s\n", err)
		util.WriteWithStatus(w, fmt.Sprintf("input [page] can't parseInt. %s\n", err), http.StatusBadRequest)
		return
	}

	questions, err := GetUserQuestions(reportID, clusterID, page, limit)
	if err != nil {
		logger.Error.Printf("Can't get report: [%s] of cluster:[%s]'s userQuestions. limit=%d, page=%d. err: %s\n", reportID, clusterID, limit, page, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	util.WriteJSON(w, questions)
}

func handleGetUserQuestion(w http.ResponseWriter, r *http.Request) {
	uID, err := util.GetMuxIntVar(r, "id")
	if err != nil {
		util.WriteWithStatus(w, "", http.StatusBadRequest)
		return
	}

	uQuestion, err := GetUserQuestion(uID)
	if err == ErrRowNotFound {
		util.WriteWithStatus(w, "", http.StatusNotFound)
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	util.WriteJSON(w, uQuestion)
}

func handleUpdateUserQuestion(w http.ResponseWriter, r *http.Request) {
	var request struct {
		StdQuestion string `json:"std_question"`
	}

	err := util.ReadJSON(r, &request)
	if err != nil {
		util.WriteWithStatus(w, "", http.StatusBadRequest)
		logger.Error.Printf("Request's body cant parse as JSON Data. %s\n", err)
		return
	}
	if request.StdQuestion == "" {
		util.WriteWithStatus(w, fmt.Sprintf("request input invalid"), http.StatusBadRequest)
		return
	}
	stdQuestion := request.StdQuestion
	qid, err := util.GetMuxIntVar(r, "id")
	if err != nil {
		util.WriteWithStatus(w, "", http.StatusBadRequest)
		return
	}

	err = UpdateStdQuestions([]int{qid}, stdQuestion)
	if err == ErrRowNotFound {
		util.WriteWithStatus(w, fmt.Sprintf("Can't found one of the id"), http.StatusNotFound)
		return
	} else if err == ErrAlreadyOccupied {
		util.WriteWithStatus(w, fmt.Sprintf("Can't updated one of the id, because it already has value"), http.StatusBadRequest)
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Error.Printf("Update id [%d] failed, %s\n", qid, err)
		return
	}

}

func handleUpdateUserQuestions(w http.ResponseWriter, r *http.Request) {
	var request struct {
		StdQuestion string `json:"std_question"`
		Feedbacks   []int  `json:"feedbacks"`
	}

	err := util.ReadJSON(r, &request)
	if err != nil {
		util.WriteWithStatus(w, fmt.Sprintf("Request's body cant parse as JSON Data. %s\n", err), http.StatusBadRequest)
		return
	}
	if request.Feedbacks == nil || len(request.Feedbacks) == 0 || request.StdQuestion == "" {
		util.WriteWithStatus(w, fmt.Sprintf("request input invalid"), http.StatusBadRequest)
		return
	}
	err = UpdateStdQuestions(request.Feedbacks, request.StdQuestion)
	if err == ErrRowNotFound {
		util.WriteWithStatus(w, fmt.Sprintf("Can't found one of the id"), http.StatusNotFound)
		return
	} else if err == ErrAlreadyOccupied {
		util.WriteWithStatus(w, fmt.Sprintf("Can't updated one of the id, because it already has value"), http.StatusBadRequest)
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Error.Printf("update User Question failed. %s\n", err)
		return
	}
}

func handleRevokeQuestion(w http.ResponseWriter, r *http.Request) {
	var (
		id  int
		err error
	)
	if id, err = util.GetMuxIntVar(r, "id"); err != nil {
		util.WriteWithStatus(w, "", http.StatusBadRequest)
		return
	}

	err = RevokeUserQuestion(id)
	if err == ErrRowNotFound {
		util.WriteWithStatus(w, "", http.StatusNotFound)
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

}

func handleDeleteReport(w http.ResponseWriter, r *http.Request) {
	var (
		id  int
		err error
	)
	if id, err = util.GetMuxIntVar(r, "id"); err != nil {
		util.WriteWithStatus(w, "", http.StatusBadRequest)
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

	err = DeleteReport(id, appid)
	if err == ErrRowNotFound {
		util.WriteWithStatus(w, "", http.StatusNotFound)
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Error.Printf("delete report [%d] failed, %v", id, err)
		return
	}

}

func handleRecommend(w http.ResponseWriter, r *http.Request) {
	sentence := make([]string, 0)
	appid := requestheader.GetAppID(r)
	err := util.ReadJSON(r, &sentence)
	if err != nil {
		util.WriteWithStatus(w, fmt.Sprintf("%s\n", err), http.StatusBadRequest)
		return
	}

	num := len(sentence)

	if num > 20 {
		util.WriteWithStatus(w, fmt.Sprintf("assigned sentence is over limit 30\n%s", err.Error()), http.StatusBadRequest)
		return
	}

	if num > 0 {
		recommend, err := getRecommend(appid, sentence)
		if err != nil {
			util.WriteWithStatus(w, fmt.Sprintf("%s\n", err), http.StatusInternalServerError)
			return
		}

		util.WriteJSON(w, recommend)
	}
}
