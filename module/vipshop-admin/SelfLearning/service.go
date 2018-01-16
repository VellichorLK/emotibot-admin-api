package SelfLearning

import (
	"runtime/debug"
	"sort"
	"time"

	"emotibot.com/emotigo/module/vipshop-admin/SelfLearning/internal/data"
	"emotibot.com/emotigo/module/vipshop-admin/SelfLearning/internal/model"
	"emotibot.com/emotigo/module/vipshop-admin/util"
)

func doClustering(s time.Time, e time.Time, reportID uint64, store StoreCluster) error {

	status := 1
	//update the status
	defer func() {
		if r := recover(); r != nil {
			util.LogError.Printf("do clustering panic. %s: %s\n ", r, debug.Stack())
			status = -1
		}

		sql := "update " + TableProps.report.name + " set " + TableProps.report.status + "=? where " + TableProps.report.id + "=?"
		sqlExec(sql, status, reportID)
	}()

	feedbackQs, feedbackQID, err := getFeedbackQ(s, e)
	if err != nil {
		status = -1
		util.LogError.Println(err)
		return err
	}

	if len(feedbackQID) > 0 {
		cluster := getClusteringResult(feedbackQs, feedbackQID)

		//no clustering result
		if cluster == nil {
			status = -1
		} else {
			util.LogTrace.Printf("Num of clusters %v,%v\n", cluster.numClustered, len(cluster.clusters))
			cluster.reportID = reportID
			err = storeClusterData(store, cluster)
			if err != nil {
				status = -1
				util.LogError.Println(err)
			}
		}

	} else {
		util.LogInfo.Println("Empty feedback question.")
	}

	return err
}

func createOneReport(s time.Time, e time.Time) (uint64, error) {
	sql := "insert into " + TableProps.report.name + " (" + TableProps.report.startTime + "," + TableProps.report.endTime + ") values (?,?)"
	result, err := sqlExec(sql, s, e)
	if err != nil {
		util.LogError.Println(err)
		return 0, err
	}
	reportID, err := result.LastInsertId()
	if err != nil {
		util.LogError.Println(err)
		return 0, err
	}
	return uint64(reportID), nil
}

func isDuplicate(s time.Time, e time.Time) (bool, uint64, error) {

	sql := "select " + TableProps.report.id + " from " + TableProps.report.name + " where " + TableProps.report.startTime + "=?" +
		" and " + TableProps.report.endTime + "=?"

	rows, err := sqlQuery(sql, s, e)
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
func getFeedbackQ(s time.Time, e time.Time) ([]string, []uint64, error) {

	sql := "select " + TableProps.feedback.id + "," + TableProps.feedback.question + " from " + TableProps.feedback.name +
		" where " + TableProps.feedback.createdTime + ">=FROM_UNIXTIME(?) and " + TableProps.feedback.createdTime + " <=FROM_UNIXTIME(?) group by " + TableProps.feedback.question +
		" limit " + util.GetEnviroment(Envs, "MAX_NUM_TO_CLUSTER")

	feedbackQs := make([]string, 0)
	feedbackQID := make([]uint64, 0)

	rows, err := sqlQuery(sql, s.Unix(), e.Unix())
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
	embeddings := make([]model.Vector, 0)
	var nativeLog data.NativeLog
	nativeLog.Init()

	util.LogTrace.Printf("The size of sentences to cluster is [%v]\n", len(feedbackQs))

	nativeLog.GetWordPos(NluURL, feedbackQs)
	util.LogTrace.Println("Calculate embeddings starts.")
	embeddingStart := time.Now()
	embeddingBatch := time.Now()
	for i := 0; i < len(nativeLog.Logs); i++ {
		dalItem := nativeLog.Logs[i]
		dalItem.Embedding = model.Vector(data.GetSentenceVector(dalItem.KeyWords, dalItem.Tokens))
		embeddings = append(embeddings, dalItem.Embedding)
		if i%500 == 0 {
			util.LogTrace.Printf("Cal: [%v]/[%v] vectors, time consuming for this batch: [%v]s\n",
				i,
				len(nativeLog.Logs),
				time.Since(embeddingBatch).Seconds())
			embeddingBatch = time.Now()
		}
	}
	util.LogTrace.Printf("Calculate embeddings ends. Time consuming: [%v]s\n",
		time.Since(embeddingStart).Seconds())

	if len(embeddings) == 0 {
		util.LogError.Println("This task is going to be closed, because: [effective embedding size is 0]")
		return nil
	}

	clusters := make([]clustering, 0)
	clusterIdxes := doCluster(embeddings, 10)
	for _, idxes := range clusterIdxes {
		feedbackQIDs := make([]uint64, 0)
		clusterTags := make([]string, 0)
		clusterPos := make([]string, 0)
		for _, idx := range idxes {

			feedbackQIDs = append(feedbackQIDs, feedbackQID[idx])
			for _, v := range nativeLog.Logs[idx].KeyWords {
				clusterPos = append(clusterPos, v)
			}
		}
		tags := data.ExtractTags(clusterPos, 2)
		for _, v := range tags {
			clusterTags = append(clusterTags, v.Text())
		}
		clusters = append(clusters, clustering{feedbackID: feedbackQIDs, tags: clusterTags})
	}

	return &clusteringResult{numClustered: len(nativeLog.Logs), clusters: clusters}
}

func doCluster(vectors []model.Vector, topN int) [][]int {
	util.LogTrace.Println("Clustering starts.")

	numClusters := len(vectors) / ClusteringBatch
	clusteredVectors := model.KmeansPP(
		vectors,
		numClusters,
		EarlyStopThreshold)

	sizeClusters := make([]int, numClusters)
	avgDistance := make([]float64, numClusters)
	for _, clusteredVector := range clusteredVectors {
		sizeClusters[clusteredVector.ClusterNumber]++
		avgDistance[clusteredVector.ClusterNumber] += clusteredVector.Distance
	}

	r := ranker{}
	for idx := range sizeClusters {
		if sizeClusters[idx] >= MinSizeCluster {
			avgDistance[idx] = avgDistance[idx] / float64(sizeClusters[idx])
			r = append(r, rankerElm{idx, avgDistance[idx]})
		}
	}

	sort.Sort(r)

	results := make([][]int, 0)
	cnt := 0
	for _, elm := range r {
		if cnt >= numClusters {
			break
		}

		targetLabel := make([]int, 0)
		for idx, vector := range clusteredVectors {
			if vector.ClusterNumber == elm.idx {
				targetLabel = append(targetLabel, idx)
			}
		}
		results = append(results, targetLabel)
		cnt++
	}
	return results
}

func storeClusterData(sc StoreCluster, clusters *clusteringResult) error {
	return sc.Store(clusters)
}
