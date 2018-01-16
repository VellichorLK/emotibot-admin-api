package SelfLearning

import (
	"net/http"
	"strconv"
	"time"

	"emotibot.com/emotigo/module/vipshop-admin/ApiError"
	"emotibot.com/emotigo/module/vipshop-admin/SelfLearning/internal/data"
	"emotibot.com/emotigo/module/vipshop-admin/util"
	"github.com/kataras/iris/context"
)

var (
	// ModuleInfo is needed for module define
	ModuleInfo util.ModuleInfo
	//Envs enviroments of clustering module
	Envs map[string]string

	NluURL             string
	EarlyStopThreshold int
	MinSizeCluster     int
	MaxNumToCluster    int
	ClusteringBatch    int
)

func init() {
	ModuleInfo = util.ModuleInfo{
		ModuleName: "selfLearn",
		EntryPoints: []util.EntryPoint{
			util.NewEntryPoint("PUT", "doClustering", []string{}, handleClustering),
			util.NewEntryPoint("GET", "reports", []string{}, handleGetReports),
			util.NewEntryPoint("GET", "reports/{id:int}", []string{}, handleGetReport),
			util.NewEntryPoint("DELETE", "reports/{id:int}", []string{}, handleDeleteReport),
			util.NewEntryPoint("GET", "reports/{id:int}/clusters", []string{}, handleGetClusters),
			util.NewEntryPoint("GET", "userQuestions", []string{}, handleGetUserQuestions),
			util.NewEntryPoint("POST", "userQuestions", []string{}, handleUpdateUserQuestions),
			util.NewEntryPoint("GET", "userQuestions/{id:int}", []string{}, handleGetUserQuestion),
			util.NewEntryPoint("POST", "userQuestions/{id:int}", []string{}, handleUpdateUserQuestion),
			util.NewEntryPoint("POST", "userQuestions/{id:int}/revoke", []string{}, handleRevokeQuestion),
		},
	}
}

func checkNeedEnvs() {
	var err error

	ClusteringBatch, err = strconv.Atoi(util.GetEnviroment(Envs, "CLUSTER_BATCH"))
	if err != nil {
		util.LogError.Println(err)
		ClusteringBatch = 50
	}

	EarlyStopThreshold, err = strconv.Atoi(util.GetEnviroment(Envs, "EARLY_STOP_THRESHOLD"))
	if err != nil {
		util.LogError.Println(err)
		EarlyStopThreshold = 3
	}

	MaxNumToCluster, err = strconv.Atoi(util.GetEnviroment(Envs, "MAX_NUM_TO_CLUSTER"))
	if err != nil {
		util.LogError.Println(err)
		MaxNumToCluster = 10000
	}

	MinSizeCluster, err = strconv.Atoi(util.GetEnviroment(Envs, "MIN_SIZE_CLUSTER"))
	if err != nil {
		util.LogError.Println(err)
		MinSizeCluster = 10
	}

	NluURL = util.GetEnviroment(Envs, "NLU_URL")
	if NluURL == "" {
		util.LogError.Println(err)
		NluURL = "http://172.17.0.1:13901"
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
		util.LogError.Printf("Cannot init self learning db, [%s:%s@%s:%s]: %s\n", user, pass, url, db, err.Error())
		return
	}
	util.SetDB(ModuleInfo.ModuleName, dao)

	if ok := data.InitializeWord2Vec(util.GetEnviroment(Envs, "RESOURCES_PATH")); !ok {
		util.LogError.Println("Load self learning caches failed!")
		return
	}
}

func handleClustering(ctx context.Context) {
	status := -999

	s, err := strconv.ParseInt(ctx.FormValue(START_TIME), 10, 64)
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(util.GenRetObj(status, "start time is not correct time format"))
		return
	}

	e, err := strconv.ParseInt(ctx.FormValue(END_TIME), 10, 64)
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(util.GenRetObj(status, "end time is not correct time format"))
		return
	}

	if s >= e {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(util.GenRetObj(status, "start time >= end time"))
		return
	}

	type respID struct {
		ReportID uint64 `json:"reportID"`
	}

	respone := &respID{}
	var isDup bool
	var reportID uint64

	st := time.Unix(s, 0)
	et := time.Unix(e, 0)
	ctx.StatusCode(http.StatusOK)

	isDup, reportID, err = isDuplicate(st, et)

	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(util.GenRetObj(status, err.Error()))
		return
	}

	if !isDup {

		reportID, err = createOneReport(st, et)
		if err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(util.GenRetObj(status, err.Error()))
			return
		}

		go doClustering(st, et, reportID, &dbStore{})
	}

	respone.ReportID = reportID
	ctx.JSON(util.GenRetObj(ApiError.SUCCESS, respone))
	return
}

func handleGetReports(ctx context.Context) {
	reports := []Report{}
	limit, err := strconv.Atoi(ctx.FormValue("limit"))
	if err != nil || limit > 10 {
		limit = 10
	}

	reports, err = GetReports("", limit)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		util.LogError.Printf("%s\n", err)
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(reports)
}

func handleGetReport(ctx context.Context) {
	reportID := ctx.Params().GetEscape("id")
	if reportID == "" {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.Writef("id should not be empty")
		return
	}

	reports, err := GetReports(reportID, 1)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		util.LogError.Printf("%s\n", err)
		return
	}
	if len(reports) == 0 {
		ctx.StatusCode(http.StatusNotFound)
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(reports[0])
}

func handleGetClusters(ctx context.Context) {
	reportID := ctx.Params().GetEscape("id")
	if reportID == "" {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.Writef("id should not be empty")
		return
	}
	reports, err := GetReports(reportID, 1)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		return
	}
	if len(reports) == 0 {
		ctx.StatusCode(http.StatusNotFound)
		return
	}
	clusters, err := GetClusters(reports[0])
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(clusters)
}

func handleGetUserQuestions(ctx context.Context) {
	reportID := ctx.FormValue("reportID")
	if reportID == "" {
		ctx.StatusCode(http.StatusBadRequest)
		return
	}
	clusterID := ctx.FormValue("clusterID")
	var limit, page int
	var err error
	if l := ctx.FormValue("limit"); l == "" {
		limit = 10
	} else {
		limit, err = strconv.Atoi(l)
	}
	if err != nil {
		util.LogError.Printf("input [limit] can't parseInt. %s\n", err)
	}

	if p := ctx.FormValue("page"); p == "" {
		page = 0
	} else {
		page, err = strconv.Atoi(p)
	}
	if err != nil {
		util.LogError.Printf("input [limit] can't parseInt. %s\n", err)
	}

	questions, err := GetUserQuestions(reportID, clusterID, page, limit)
	if err != nil {
		util.LogError.Printf("Can't get report: [%s] of cluster:[%s]'s userQuestions. limit=%d, page=%d. err: %s\n", reportID, clusterID, limit, page, err)
		ctx.StatusCode(http.StatusInternalServerError)
		return
	}
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(questions)
}

func handleGetUserQuestion(ctx context.Context) {
	uID, err := strconv.Atoi(ctx.Params().GetEscape("id"))
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		return
	}

	uQuestion, err := GetUserQuestion(uID)
	if err == ErrRowNotFound {
		ctx.StatusCode(http.StatusNotFound)
		return
	} else if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		return
	}

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(uQuestion)
}

func handleUpdateUserQuestion(ctx context.Context) {
	var request struct {
		StdQuestion string `json:"std_question"`
	}

	err := ctx.ReadJSON(&request)
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		util.LogError.Printf("Request's body cant parse as JSON Data. %s\n", err)
		return
	}
	stdQuestion := request.StdQuestion
	qid, err := strconv.Atoi(ctx.Params().GetEscape("id"))
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		return
	}

	err = UpdateStdQuestion([]int{qid}, stdQuestion)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		util.LogError.Printf("Update id [%d] failed, %s\n", qid, err)
		return
	}

	ctx.StatusCode(http.StatusOK)
}

func handleUpdateUserQuestions(ctx context.Context) {
	var request struct {
		StdQuestion string `json:"std_question"`
		Feedbacks   []int  `json:"feedbacks"`
	}

	err := ctx.ReadJSON(&request)
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		util.LogError.Printf("Request's body cant parse as JSON Data. %s\n", err)
		return
	}

	err = UpdateStdQuestion(request.Feedbacks, request.StdQuestion)
	if err == ErrRowNotFound {
		ctx.StatusCode(http.StatusNotFound)
		ctx.Writef("Can't found one of the id")
		return
	} else if err == ErrAlreadyOccupied {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.Writef("Can't updated one of the id, because it already has value")
		return
	} else if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		util.LogError.Printf("update User Question failed. %s\n", err)
		return
	}
	ctx.StatusCode(http.StatusOK)
}

func handleRevokeQuestion(ctx context.Context) {
	var (
		id  int
		err error
	)
	if id, err = strconv.Atoi(ctx.Params().GetEscape("id")); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		return
	}

	err = RevokeUserQuestion(id)
	if err == ErrRowNotFound {
		ctx.StatusCode(http.StatusNotFound)
		return
	} else if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		return
	}

}

func handleDeleteReport(ctx context.Context) {
	var (
		id  int
		err error
	)
	if id, err = strconv.Atoi(ctx.Params().GetEscape("id")); err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		return
	}
	err = DeleteReport(id)
	if err == ErrRowNotFound {
		ctx.StatusCode(http.StatusNotFound)
		return
	} else if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		util.LogError.Printf("delete report [%d] failed, %v", id, err)
		return
	}

	ctx.StatusCode(http.StatusOK)
}
