package clustering

import (
	"context"
	"errors"
	"net/http"
	"runtime/debug"
	"sort"
	"time"

	statData "emotibot.com/emotigo/module/admin-api/ELKStats/data"
	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/pkg/logger"
)

//Service define the operations
type Service interface {
	NewReport([]*statData.VisitRecordsData) (Report, error)
	Report(id string) (Report, error)
	Reports(query ReportQuery) ([]Report, error)
	CancelReport(appID string, id string) error
}

var service Service = sqlService{}

// ErrNotAvailable is used for indicating out of resource, since clustering is a resource intensive operation.
var ErrNotAvailable = errors.New("clustering error: resource is not available yet. please try again")

func doClustering(s time.Time, e time.Time, reportID uint64, store StoreCluster, appid string, qType int) error {

	status := S_SUCCESS
	//update the status
	defer func() {
		if r := recover(); r != nil {
			logger.Error.Printf("do clustering panic. %s: %s\n ", r, debug.Stack())
			status = S_PANIC
		}
		sql := "update " + TableProps.report.name + " set " + TableProps.report.status + "=? where " + TableProps.report.id + "=?"
		sqlExec(sql, status, reportID)
	}()

	feedbackQs, feedbackQID, err := getFeedbackQ(s, e, appid, qType)
	if err != nil {
		status = S_PANIC
		logger.Error.Println(err)
		return err
	}

	if len(feedbackQID) > 0 {
		cluster := getClusteringResult(feedbackQs, feedbackQID)

		//no clustering result
		if cluster == nil {
			status = S_PANIC
		} else {
			logger.Trace.Printf("Num of clusters %v,%v\n", cluster.numClustered, len(cluster.clusters))
			cluster.reportID = reportID
			err = storeClusterData(store, cluster)
			if err != nil {
				status = S_PANIC
				logger.Error.Println(err)
			}
		}

	} else {
		logger.Info.Println("Empty feedback question.")
	}

	return err
}

func createOneReport(s time.Time, e time.Time, appid string, rType int) (uint64, error) {
	sql := "insert into " + TableProps.report.name + " (" + TableProps.report.startTime + "," + TableProps.report.endTime + "," + TableProps.report.appid + "," + TableProps.report.rType + ") values (?,?,?,?)"
	result, err := sqlExec(sql, s, e, appid, rType)
	if err != nil {
		logger.Error.Println(err)
		return 0, err
	}
	reportID, err := result.LastInsertId()
	if err != nil {
		logger.Error.Println(err)
		return 0, err
	}
	return uint64(reportID), nil
}

func isDuplicate(s time.Time, e time.Time, appid string, pType int) (bool, uint64, error) {

	sql := "select " + TableProps.report.id + " from " + TableProps.report.name + " where " + TableProps.report.startTime + "=?" +
		" and " + TableProps.report.endTime + "=?" + " and " + TableProps.report.appid + "=?" + " and " + TableProps.report.rType + "=?"

	rows, err := sqlQuery(sql, s, e, appid, pType)
	if err != nil {
		return false, 0, err
	}
	defer rows.Close()

	var count int
	var reportID uint64
	hasDup := false

	if rows.Next() {
		count++
		rows.Scan(&reportID)
		hasDup = true
	}

	return hasDup, reportID, nil
}

//return parameters, questions, id of questions in database, error
func getFeedbackQ(s time.Time, e time.Time, appid string, qType int) ([]string, []uint64, error) {

	sql := "select max(" + TableProps.feedback.id + ")," + TableProps.feedback.question + " from " + TableProps.feedback.name +
		" where " + TableProps.feedback.createdTime + ">=FROM_UNIXTIME(?) and " +
		TableProps.feedback.createdTime + " <=FROM_UNIXTIME(?) and " + TableProps.feedback.appid + "=? " + " and " + TableProps.feedback.qType + "=? " +
		"group by " + TableProps.feedback.question + " limit " + util.GetEnviroment(Envs, "MAX_NUM_TO_CLUSTER")

	feedbackQs := make([]string, 0, MaxNumToCluster)
	feedbackQID := make([]uint64, 0, MaxNumToCluster)

	rows, err := sqlQuery(sql, s.Unix(), e.Unix(), appid, qType)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var q string
		var id uint64
		err := rows.Scan(&id, &q)
		if err != nil {
			return nil, nil, err
		}
		feedbackQID = append(feedbackQID, id)
		feedbackQs = append(feedbackQs, q)
	}

	return feedbackQs, feedbackQID, rows.Err()
}

func getClusteringResult(feedbackQs []string, feedbackQID []uint64) *clusteringResult {
	return &clusteringResult{numClustered: 0, clusters: nil}
}

func storeClusterData(sc StoreCluster, clusters *clusteringResult) error {
	return sc.Store(clusters)
}

func getRecommend(appid string, sentence []string) ([]*RecommendQ, error) {
	pool := 4
	num := len(sentence)
	if num < pool {
		pool = num
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // cancel when we are finished consuming integers
	stringChannel := make(chan string)
	responseChannel := make(chan *rresponse, pool)

	for i := 0; i < num; i++ {
		go func(ctx context.Context) {
			var s string
			timeout := time.Duration(2 * time.Second)
			response := &responseClient{URL: responseURL, client: &http.Client{Timeout: timeout}}

			for {
				select {
				case <-ctx.Done():
					return
				case s = <-stringChannel:
					r, err := response.Post(appid, s)
					if err != nil {
						logger.Error.Println(err)
					}

					select {
					case responseChannel <- r:
					case <-ctx.Done():
						return
					}

				}
			}
		}(ctx)
	}

	var receiveCount int
	stdQs := make(map[string]float64)

	go func() {
		for i := 0; i < len(sentence); i++ {
			stringChannel <- sentence[i]
		}
	}()

	for {
		r := <-responseChannel
		receiveCount++
		if r != nil && r.OtherInfo != nil &&
			r.OtherInfo.Custom != nil && r.OtherInfo.Custom.RelatedQ != nil {
			for i := 0; i < len(r.OtherInfo.Custom.RelatedQ); i++ {
				if s, ok := stdQs[r.OtherInfo.Custom.RelatedQ[i].StdQ]; ok {
					if s < r.OtherInfo.Custom.RelatedQ[i].Score {
						stdQs[r.OtherInfo.Custom.RelatedQ[i].StdQ] = r.OtherInfo.Custom.RelatedQ[i].Score
					}
				} else {
					stdQs[r.OtherInfo.Custom.RelatedQ[i].StdQ] = r.OtherInfo.Custom.RelatedQ[i].Score
				}
			}
		}

		if receiveCount >= num {
			break
		}
	}

	sorter := &sortMapKey{mapData: stdQs}
	sorter.keyToSlice()
	sort.Sort(sorter)

	questionIDMap, err := GetQuestionIDByContent(appid, sorter.sliceData)
	if err != nil {
		return nil, err
	}

	recommend := make([]*RecommendQ, 0, len(sorter.sliceData))

	for i := 0; i < len(sorter.sliceData); i++ {
		if id, ok := questionIDMap[sorter.sliceData[i].(string)]; ok {
			rQ := &RecommendQ{QID: id, Content: sorter.sliceData[i].(string)}
			recommend = append(recommend, rQ)
		} else {
			logger.Warn.Printf("[SelfLearn][Recommend] has %s question but doesn't have id\n", sorter.sliceData[i].(string))
		}
	}

	return recommend, nil
}
