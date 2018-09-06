package clustering

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	statData "emotibot.com/emotigo/module/admin-api/ELKStats/data"
	statService "emotibot.com/emotigo/module/admin-api/ELKStats/services"
	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/admin-api/util/requestheader"
	"emotibot.com/emotigo/pkg/api/faqcluster/v1"
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

func Init() error {

	moduleName := "clustering"
	envs := util.GetModuleEnvironments(moduleName)
	// dbName, _ := envs["MYSQL_DB"]
	// dbUser, _ = envs["MYSQL_USER"]
	// dbPass, _ := envs["MYSQL_PASS"]
	// db := util.InitDB(envs[""], envs[""])
	db := util.GetMainDB()
	if db == nil {
		return fmt.Errorf("cant not get db of " + moduleName)
	}
	ss := &sqlService{db: db}
	httpClient := &http.Client{Timeout: 0}
	toolURL, _ := envs["TOOL_URL"]
	addr, err := url.Parse(toolURL)
	if err != nil {
		return fmt.Errorf("parse env failed, %v", err)
	}
	clusterClient := faqcluster.NewClientWithHTTPClient(addr, httpClient)
	ModuleInfo = util.ModuleInfo{
		ModuleName: moduleName,
		EntryPoints: []util.EntryPoint{
			util.NewEntryPoint(http.MethodPut, "reports", []string{}, NewDoReportHandler(ss, ss, ss, ss, clusterClient)),
			util.NewEntryPoint(http.MethodGet, "reports/{id}", []string{}, NewGetReportHandler(ss, ss, ss)),
		},
	}
	return nil
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

func initSelfLearningDB(url string, user string, pass string, db string) (*sql.DB, error) {
	return util.InitDB(url, user, pass, db)
}

//NewDoReportHandler create a DoReport Handler with given reportSerivce & faqClient.
func NewDoReportHandler(reportService ReportsService, recordsService ReportRecordsService, clusterService ReportClustersService, simpleFTService SimpleFTService, faqClient *faqcluster.Client) http.HandlerFunc {
	type request struct {
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		appid := requestheader.GetAppID(r)
		var query statData.RecordQuery
		requestBody, _ := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		err := json.Unmarshal(requestBody, &query)
		if err != nil {
			logger.Warn.Printf("PUT /reports: request body can not be decoded, %s", requestBody)
			http.Error(w, "input format error", http.StatusBadRequest)
			return
		}
		query.AppID = appid
		query.Limit = 10000
		result, err := statService.VisitRecordsQuery(query, statService.ElasticFilterMarkedRecord, statService.ElasticFilterMarkedRecord)
		if err != nil {
			logger.Error.Printf("get records failed, %v", err)
		}
		now := time.Now().Unix()
		thirtyMinAgo := now - 1800
		s := int(ReportStatusRunning)
		rQuery := ReportQuery{
			AppID: appid,
			UpdatedTime: &searchPeriod{
				StartTime: &thirtyMinAgo,
				EndTime:   &now,
			},
			Status: &s,
		}

		reports, err := reportService.QueryReports(rQuery)
		if err != nil {
			logger.Error.Printf("PUT /reports: new request error, %v\n", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		if len(reports) > 0 {
			http.Error(w, "conflicted", http.StatusConflict)
			return
		}

		ignoredSize, _ := result.Aggs["isMarked"].(int64)
		markedSize, _ := result.Aggs["isIgnored"].(int64)
		newReport := Report{
			CreatedTime: time.Now().Unix(),
			UpdatedTime: time.Now().Unix(),
			Condition:   string(requestBody),
			UserID:      requestheader.GetUserID(r),
			AppID:       appid,
			IgnoredSize: ignoredSize,
			MarkedSize:  markedSize,
			Status:      ReportStatusRunning,
		}
		id, err := reportService.NewReport(newReport)
		if err != nil {
			logger.Error.Printf("Failed to create new report, %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		model, err := simpleFTService.GetFTModel(appid)
		if err == sql.ErrNoRows {
			http.Error(w, "bad request, need train model before use", http.StatusBadRequest)
			return
		}
		if err != nil {
			logger.Error.Printf("Failed to get simpleFT model, %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		var inputs = []interface{}{}
		for _, h := range result.Hits {
			var input = map[string]string{
				"id":    h.UniqueID,
				"value": h.UserQ,
			}
			inputs = append(inputs, input)
		}
		var paramas = map[string]interface{}{
			"model_version": model,
		}
		go func() {
			logger.Trace.Println("total sentences for request: ", len(inputs))
			clusterResult, err := faqClient.Clustering(context.Background(), paramas, inputs)
			if err != nil {
				rawError, ok := err.(*faqcluster.RawError)
				if ok {
					logger.Error.Printf("raw http error, %s", rawError.Body)
				}
				reportError(reportService, "faq clustering failed, "+err.Error(), id)
				reportService.UpdateReportStatus(id, ReportStatusError)
				return
			}
			for _, c := range clusterResult.Clusters {
				fmt.Printf("test %+v", c)
				var records = []ReportRecord{}
				tags, _ := json.Marshal(c)
				cID, err := clusterService.NewCluster(Cluster{
					ReportID:    id,
					Tags:        string(tags),
					CreatedTime: time.Now().Unix(),
				})
				if err != nil {
					reportError(reportService, "new cluster failed, "+err.Error(), id)
					return
				}
				for _, d := range c.Data {
					chID, ok := d.Others["id"].(string)
					if !ok {
						reportError(reportService, "data id is not string", id)
						return
					}
					r := ReportRecord{
						ClusterID:    cID,
						ReportID:     id,
						ChatRecordID: chID,
						Content:      d.Value,
						CreatedTime:  time.Now().Unix(),
					}
					records = append(records, r)
				}
				err = recordsService.NewRecords(records...)
				if err != nil {
					reportError(reportService, "new report_records failed, "+err.Error(), id)
					return
				}
			}
			fmt.Printf("task %d finished, with clusters %+v", id, clusterResult.Clusters)
			reportService.UpdateReportStatus(id, ReportStatusCompleted)
		}()
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(struct {
			ID uint64 `json:"report_id"`
		}{ID: id})
	})
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
	_ = query
	// reports, err = service.Reports(query)
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

func NewGetReportHandler(rs ReportsService, cs ReportClustersService, rrs ReportRecordsService) http.HandlerFunc {
	type record struct {
		ID        string `json:"id"`
		Value     string `json:"value"`
		IsCenterQ bool   `json:"-"`
	}
	type cluster struct {
		ID      uint64   `json:"id"`
		CenterQ []string `json:"center_questions"`
		Tags    []string `json:"tags"`
		Records []record `json:"records"`
	}
	type result struct {
		TotalSize int64       `json:"total_size"`
		Cluster   []cluster   `json:"clusters"`
		Filtered  interface{} `json:"filtered"`
	}
	type response struct {
		ID          uint64       `json:"id"`
		StartTime   int64        `json:"start_time"`
		EndTime     int64        `json:"end_time"`
		Conditions  string       `json:"search_query"`
		Result      *result      `json:"results,omitempty"`
		IgnoredSize int64        `json:"ignored_size"`
		MarkedSize  int64        `json:"marked_size"`
		Status      ReportStatus `json:"status"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		reportID, err := strconv.ParseUint(util.GetMuxVar(r, "id"), 10, 64)
		if err != nil {
			http.Error(w, "id is invalid", http.StatusBadRequest)
			return
		}
		appid := requestheader.GetAppID(r)
		report, err := rs.GetReport(reportID)
		if err != nil {
			logger.Error.Printf("get report failed, %v\n", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		if report.AppID != appid {
			http.Error(w, "permission denied", http.StatusBadRequest)
			return
		}
		var response = response{
			ID:          report.ID,
			StartTime:   report.CreatedTime,
			EndTime:     report.UpdatedTime,
			Conditions:  report.Condition,
			IgnoredSize: report.IgnoredSize,
			MarkedSize:  report.MarkedSize,
			Status:      report.Status,
		}
		if report.Status == ReportStatusCompleted {
			records, err := rrs.GetRecords(report.ID)
			if err != nil {
				http.Error(w, "get records failed, "+err.Error(), http.StatusInternalServerError)
				return
			}
			clusters := map[uint64]Cluster{}
			clusterRecords := map[uint64][]record{}
			filtered := []record{}
			for _, rc := range records {
				respRecord := record{
					ID:        rc.ChatRecordID,
					Value:     rc.Content,
					IsCenterQ: rc.IsCentralNode,
				}
				if rc.ClusterID == nonClusterID {
					filtered = append(filtered, respRecord)
					continue
				}
				c, found := clusters[rc.ClusterID]
				if !found {
					c, err = cs.GetCluster(rc.ClusterID)
					if err != nil {
						return
					}
					clusters[rc.ClusterID] = c
					clusterRecords[rc.ClusterID] = []record{}
				}
				//since we already checked and init at the other map, so we can
				clusterRecords[rc.ClusterID] = append(clusterRecords[rc.ClusterID], respRecord)
			}
			response.Result = &result{
				TotalSize: int64(len(records)),
				Cluster:   []cluster{},
				Filtered:  filtered,
			}
			//transform Cluster into response's cluster
			for id, c := range clusters {
				var tags []string
				json.Unmarshal([]byte(c.Tags), &tags)
				respCluster := cluster{
					ID:   c.ID,
					Tags: tags,
				}

				for _, record := range clusterRecords[id] {
					if record.IsCenterQ {
						respCluster.CenterQ = append(respCluster.CenterQ, record.Value)
					}
					respCluster.Records = append(respCluster.Records, record)
				}
				response.Result.Cluster = append(response.Result.Cluster, respCluster)
			}

		}
		util.WriteJSON(w, response)
	}

}
