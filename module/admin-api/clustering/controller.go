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

	"emotibot.com/emotigo/pkg/api/dal/v1"

	statDataV1 "emotibot.com/emotigo/module/admin-api/ELKStats/data/v1"
	statServiceCommon "emotibot.com/emotigo/module/admin-api/ELKStats/services/common"
	statServiceV1 "emotibot.com/emotigo/module/admin-api/ELKStats/services/v1"
	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/admin-api/util/AdminErrors"
	"emotibot.com/emotigo/module/admin-api/util/requestheader"
	"emotibot.com/emotigo/pkg/api/faqcluster/v1"
	"emotibot.com/emotigo/pkg/logger"

	dac "emotibot.com/emotigo/pkg/api/dac/v1"
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
	db := util.GetMainDB()
	dbfaq := util.GetFaqDB()
	if db == nil {
		return fmt.Errorf("cant not get db of " + moduleName)
	}
	if dbfaq == nil {
		return fmt.Errorf("cant not get db of " + moduleName)
	}
	ss := &sqlService{db: db}
	ssfaq := &sqlService{db: dbfaq}
	httpClient := &http.Client{Timeout: 0}
	toolURL, _ := envs["TOOL_URL"]
	addr, err := url.Parse(toolURL)
	if err != nil {
		return fmt.Errorf("parse env failed, %v", err)
	}
	clusterClient := faqcluster.NewClientWithHTTPClient(addr, httpClient)
	dalURL, found := util.GetEnvOf("server")["DAL_URL"]
	if !found {
		return fmt.Errorf("CAN NOT FOUND SERVER ENV \"DAL_URL\"")
	}

	dalClient, err := dal.NewClientWithHTTPClient(dalURL, &http.Client{Timeout: time.Duration(5) * time.Second})
	if err != nil {
		return fmt.Errorf("new dal client failed, %v", err)
	}

	dacURL, found := util.GetEnvOf("server")["DAC_URL"]
	if !found {
		return fmt.Errorf("CAN NOT FOUND SERVER ENV \"DAC_URL\"")
	}

	dacClient, err := dac.NewClientWithHTTPClient(dacURL, &http.Client{Timeout: time.Duration(5) * time.Second})
	if err != nil {
		return fmt.Errorf("new dac client failed, %v", err)
	}


	ModuleInfo = util.ModuleInfo{
		ModuleName: moduleName,
		EntryPoints: []util.EntryPoint{
			util.NewEntryPoint(http.MethodPut, "reports", []string{}, NewDoReportHandler(ss, ss, ss, ss, clusterClient, dalClient)),
			util.NewEntryPoint(http.MethodGet, "reports/{id}", []string{}, NewGetReportHandler(ss, ss, ss)),

			util.NewEntryPointWithVer(http.MethodPut, "reports", []string{}, NewDoReportHandlerV2(ss, ss, ss, ssfaq, clusterClient, dacClient), 2),
			util.NewEntryPointWithVer(http.MethodGet, "reports/{id}", []string{}, NewGetReportHandler(ss, ss, ss), 2),
		},
	}
	worker = newClusteringWork(ss, ss, ss, clusterClient)
	return nil
}

//NewDoReportHandler create a DoReport Handler with given reportSerivce & faqClient.
func NewDoReportHandler(reportService ReportsService, recordsService ReportRecordsService, clusterService ReportClustersService, simpleFTService SimpleFTService, faqClient *faqcluster.Client, dalClient *dal.Client) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		appid := requestheader.GetAppID(r)
		var query statDataV1.RecordQuery
		requestBody, _ := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		err := json.Unmarshal(requestBody, &query)
		if err != nil {
			logger.Warn.Printf("PUT /reports: request body can not be decoded, %s", requestBody)
			util.Return(w, AdminErrors.New(AdminErrors.ErrnoRequestError, "input format error"), nil)
			return
		}
		rawRequestQuery, _ := json.Marshal(query)
		query.AppID = appid
		query.Limit = 0
		result, err := statServiceV1.VisitRecordsQuery(query,
			statServiceCommon.AggregateFilterMarkedRecord, statServiceCommon.AggregateFilterIgnoredRecord)
		if err != nil {
			logger.Error.Printf("get records failed, %v", err)
		}
		markedSize, _ := result.Aggs["isMarked"].(int64)
		ignoredSize, _ := result.Aggs["isIgnored"].(int64)
		logger.Trace.Printf("agg info: %+v\n", result.Aggs)
		//Because we need to query a user query first to see if aggregation result
		//than we change query condition to match the spec of new report(do not include marked & ignored records)
		//limit 10000 is the limitation of elastic search.
		query.Limit = 10000
		query.IsIgnored = new(bool)
		query.IsMarked = new(bool)
		result, err = statServiceV1.VisitRecordsQuery(query)
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
			util.Return(w, AdminErrors.New(AdminErrors.ErrnoDBError, "internal server error"), nil)
			return
		}
		if len(reports) > 0 {
			http.Error(w, "conflicted", http.StatusConflict)
			return
		}

		newReport := Report{
			CreatedTime: time.Now().Unix(),
			UpdatedTime: time.Now().Unix(),
			Condition:   string(rawRequestQuery),
			UserID:      requestheader.GetUserID(r),
			AppID:       appid,
			IgnoredSize: ignoredSize,
			MarkedSize:  markedSize,
			SkippedSize: 0,
			Status:      ReportStatusRunning,
		}
		model, err := simpleFTService.GetFTModel(appid)
		if err == sql.ErrNoRows {
			newReportError(reportService, "bad request, need train model before use", 0)
			util.Return(w, AdminErrors.New(AdminErrors.ErrnoRequestError, "找不到Model，请先透过SSM训练"), nil)
			return
		}
		if err != nil {
			newReportError(reportService, "Failed to get simpleFT model, "+err.Error(), 0)
			util.Return(w, AdminErrors.New(AdminErrors.ErrnoDBError, "internal server error"), nil)
			return
		}
		var inputs = []interface{}{}
		for _, h := range result.Hits {
			found, err := dalClient.IsStandardQuestion(appid, h.UserQ)
			if err != nil {
				newReportError(reportService, "dal client error "+err.Error(), 0)
				util.Return(w, AdminErrors.New(AdminErrors.ErrnoAPIError, "internal server error"), nil)
				return
			}
			if found {
				newReport.SkippedSize++
				continue
			}
			var input = map[string]string{
				"id":    h.UniqueID,
				"value": h.UserQ,
			}
			inputs = append(inputs, input)
		}
		id, err := reportService.NewReport(newReport)
		if err != nil {
			logger.Error.Printf("Failed to create new report, %v", err)
			util.Return(w, AdminErrors.New(AdminErrors.ErrnoDBError, "internal server error"), nil)
			return
		}
		var paramas = map[string]interface{}{
			"model_version": model,
		}
		go worker(id, paramas, inputs)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(struct {
			ID uint64 `json:"report_id"`
		}{ID: id})
	})
}


//NewDoReportHandler create a DoReport Handler with given reportSerivce & faqClient.
func NewDoReportHandlerV2(reportService ReportsService, recordsService ReportRecordsService, clusterService ReportClustersService, simpleFTService SimpleFTService, faqClient *faqcluster.Client, dacClient *dac.Client) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		appid := requestheader.GetAppID(r)
		var query statDataV1.RecordQuery
		requestBody, _ := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		err := json.Unmarshal(requestBody, &query)
		if err != nil {
			logger.Warn.Printf("PUT /reports: request body can not be decoded, %s", requestBody)
			util.Return(w, AdminErrors.New(AdminErrors.ErrnoRequestError, "input format error"), nil)
			return
		}
		rawRequestQuery, _ := json.Marshal(query)
		query.AppID = appid
		query.Limit = 0
		result, err := statServiceV1.VisitRecordsQuery(query,
			statServiceCommon.AggregateFilterMarkedRecord, statServiceCommon.AggregateFilterIgnoredRecord)
		if err != nil {
			logger.Error.Printf("get records failed, %v", err)
		}
		markedSize, _ := result.Aggs["isMarked"].(int64)
		ignoredSize, _ := result.Aggs["isIgnored"].(int64)
		logger.Trace.Printf("agg info: %+v\n", result.Aggs)
		//Because we need to query a user query first to see if aggregation result
		//than we change query condition to match the spec of new report(do not include marked & ignored records)
		//limit 10000 is the limitation of elastic search.
		query.Limit = 10000
		query.IsIgnored = new(bool)
		query.IsMarked = new(bool)
		result, err = statServiceV1.VisitRecordsQuery(query)
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
			util.Return(w, AdminErrors.New(AdminErrors.ErrnoDBError, "internal server error"), nil)
			return
		}
		if len(reports) > 0 {
			http.Error(w, "conflicted", http.StatusConflict)
			return
		}

		newReport := Report{
			CreatedTime: time.Now().Unix(),
			UpdatedTime: time.Now().Unix(),
			Condition:   string(rawRequestQuery),
			UserID:      requestheader.GetUserID(r),
			AppID:       appid,
			IgnoredSize: ignoredSize,
			MarkedSize:  markedSize,
			SkippedSize: 0,
			Status:      ReportStatusRunning,
		}
		model, err := simpleFTService.GetFTModel(appid)
		if err == sql.ErrNoRows {
			newReportError(reportService, "bad request, need train model before use", 0)
			util.Return(w, AdminErrors.New(AdminErrors.ErrnoRequestError, "找不到Model，请先透过SSM训练"), nil)
			return
		}
		if err != nil {
			newReportError(reportService, "Failed to get simpleFT model, "+err.Error(), 0)
			util.Return(w, AdminErrors.New(AdminErrors.ErrnoDBError, "internal server error"), nil)
			return
		}
		var inputs = []interface{}{}
		for _, h := range result.Hits {
			found, err := dacClient.IsStandardQuestion(appid, h.UserQ)
			if err != nil {
				newReportError(reportService, "dac client error "+err.Error(), 0)
				util.Return(w, AdminErrors.New(AdminErrors.ErrnoAPIError, "internal server error"), nil)
				return
			}
			if found {
				newReport.SkippedSize++
				continue
			}
			var input = map[string]string{
				"id":    h.UniqueID,
				"value": h.UserQ,
			}
			inputs = append(inputs, input)
		}
		id, err := reportService.NewReport(newReport)
		if err != nil {
			logger.Error.Printf("Failed to create new report, %v", err)
			util.Return(w, AdminErrors.New(AdminErrors.ErrnoDBError, "internal server error"), nil)
			return
		}
		var paramas = map[string]interface{}{
			"model_version": model,
		}
		go worker(id, paramas, inputs)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(struct {
			ID uint64 `json:"report_id"`
		}{ID: id})
	})
}



//NewGetReportHandler create a http.Handler for handling request of
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
		SkippedSize int64        `json:"skipped_size"`
		Status      ReportStatus `json:"status"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("20")
		reportID, err := strconv.ParseUint(util.GetMuxVar(r, "id"), 10, 64)
		if err != nil {
			util.Return(w, AdminErrors.New(AdminErrors.ErrnoRequestError, "id is invalid"), nil)
			return
		}
		appid := requestheader.GetAppID(r)
		report, err := rs.GetReport(reportID)
		if err != nil {
			logger.Error.Printf("get report failed, %v\n", err)
			util.Return(w, AdminErrors.New(AdminErrors.ErrnoDBError, "internal server error"), nil)
			return
		}
		if report.AppID != appid {
			util.Return(w, AdminErrors.New(AdminErrors.ErrnoRequestError, "permission denied"), nil)
			return
		}
		var response = response{
			ID:          report.ID,
			StartTime:   report.CreatedTime,
			EndTime:     report.UpdatedTime,
			Conditions:  report.Condition,
			IgnoredSize: report.IgnoredSize,
			MarkedSize:  report.MarkedSize,
			SkippedSize: report.SkippedSize,
			Status:      report.Status,
		}
		if report.Status == ReportStatusCompleted {
			records, err := rrs.GetRecords(report.ID)
			if err != nil {
				logger.Error.Println("get records failed, " + err.Error())
				util.Return(w, AdminErrors.New(AdminErrors.ErrnoDBError, "internal server error"), nil)
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
						logger.Error.Println(err.Error())
						util.Return(w, AdminErrors.New(AdminErrors.ErrnoDBError, "internal server error"), nil)
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
						var alreadyCenterQ = false
						for _, q := range respCluster.CenterQ {
							if record.Value == q {
								alreadyCenterQ = true
								break
							}
						}
						if !alreadyCenterQ {
							respCluster.CenterQ = append(respCluster.CenterQ, record.Value)
						}
					}
					respCluster.Records = append(respCluster.Records, record)
				}
				response.Result.Cluster = append(response.Result.Cluster, respCluster)
			}

		}
		util.WriteJSON(w, response)
	}

}

//cluster use inputs data to do the cluster, and record it into different resources.
type cluster func(reportID uint64, paramas map[string]interface{}, inputs []interface{}) error

//worker is used for forking cluster work
var worker cluster

//newClusteringWork create a cluster func by given rs, rrs, cs, faqClient.
func newClusteringWork(rs ReportsService, rrs ReportRecordsService, cs ReportClustersService, faqClient *faqcluster.Client) cluster {
	return func(reportID uint64, paramas map[string]interface{}, inputs []interface{}) error {
		logger.Trace.Println("total sentences for clustering request: ", len(inputs))
		clusterResult, err := faqClient.Clustering(context.Background(), paramas, inputs)
		if err != nil {
			rawError, ok := err.(*faqcluster.RawError)
			if ok {
				logger.Error.Printf("raw http error, %s", rawError.Body)
			}
			if rawError.StatusCode == http.StatusBadRequest {
				newReportErrorWithStatus(rs, "faq clustering failed, "+err.Error(), reportID, ReportInvalidInput)
			} else {
				newReportError(rs, "faq clustering failed, "+err.Error(), reportID)
			}
			return err
		}
		for _, c := range clusterResult.Clusters {
			var records = []ReportRecord{}
			tags, _ := json.Marshal(c.Tags)
			cID, err := cs.NewCluster(Cluster{
				ReportID:    reportID,
				Tags:        string(tags),
				CreatedTime: time.Now().Unix(),
			})
			if err != nil {
				newReportError(rs, "new cluster failed, "+err.Error(), reportID)
				return err
			}
			for _, data := range c.Data {
				chID, ok := data.Others["id"].(string)
				if !ok {
					newReportError(rs, "data id is not string", reportID)
					return err
				}
				_, isCenterQ := c.CenterQuestions[data.Value]
				r := ReportRecord{
					ClusterID:     cID,
					ReportID:      reportID,
					ChatRecordID:  chID,
					Content:       data.Value,
					IsCentralNode: isCenterQ,
					CreatedTime:   time.Now().Unix(),
				}
				records = append(records, r)
			}
			err = rrs.NewRecords(records...)
			if err != nil {
				newReportError(rs, "new report_records failed, "+err.Error(), reportID)
				return err
			}
		}
		var records = []ReportRecord{}
		for _, data := range clusterResult.Filtered {
			chID := data.Others["id"].(string)
			r := ReportRecord{
				ClusterID:     0,
				ReportID:      reportID,
				ChatRecordID:  chID,
				Content:       data.Value,
				IsCentralNode: false,
				CreatedTime:   time.Now().Unix(),
			}
			records = append(records, r)
		}
		err = rrs.NewRecords(records...)
		if err != nil {
			newReportError(rs, "new report_records for filtered sentences failed, "+err.Error(), reportID)
			return err
		}
		logger.Trace.Printf("task %d finished", reportID)
		rs.UpdateReportStatus(reportID, ReportStatusCompleted)
		return nil
	}

}
