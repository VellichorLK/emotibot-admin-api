package SelfLearning

import (
	"context"
	"net/http"
	"runtime/debug"
	"sort"
	"strings"
	"time"

	"emotibot.com/emotigo/module/vipshop-admin/SelfLearning/internal/data"
	"emotibot.com/emotigo/module/vipshop-admin/SelfLearning/internal/model"
	"emotibot.com/emotigo/module/vipshop-admin/util"
)

func doClustering(s time.Time, e time.Time, reportID uint64, store StoreCluster, appid string) error {

	status := S_SUCCESS
	//update the status
	defer func() {
		if r := recover(); r != nil {
			util.LogError.Printf("do clustering panic. %s: %s\n ", r, debug.Stack())
			status = S_PANIC
		}
		sql := "update " + TableProps.report.name + " set " + TableProps.report.status + "=? where " + TableProps.report.id + "=?"
		sqlExec(sql, status, reportID)
	}()

	feedbackQs, feedbackQID, err := getFeedbackQ(s, e, appid)
	if err != nil {
		status = S_PANIC
		util.LogError.Println(err)
		return err
	}

	if len(feedbackQID) > 0 {
		cluster := getClusteringResult(feedbackQs, feedbackQID)

		//no clustering result
		if cluster == nil {
			status = S_PANIC
		} else {
			util.LogTrace.Printf("Num of clusters %v,%v\n", cluster.numClustered, len(cluster.clusters))
			cluster.reportID = reportID
			err = storeClusterData(store, cluster)
			if err != nil {
				status = S_PANIC
				util.LogError.Println(err)
			}
		}

	} else {
		util.LogInfo.Println("Empty feedback question.")
	}

	return err
}

func createOneReport(s time.Time, e time.Time, appid string) (uint64, error) {
	sql := "insert into " + TableProps.report.name + " (" + TableProps.report.startTime + "," + TableProps.report.endTime + "," + TableProps.report.appid + ") values (?,?,?)"
	result, err := sqlExec(sql, s, e, appid)
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

func isDuplicate(s time.Time, e time.Time, appid string) (bool, uint64, error) {

	sql := "select " + TableProps.report.id + " from " + TableProps.report.name + " where " + TableProps.report.startTime + "=?" +
		" and " + TableProps.report.endTime + "=?" + " and " + TableProps.report.appid + "=?"

	rows, err := sqlQuery(sql, s, e, appid)
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
func getFeedbackQ(s time.Time, e time.Time, appid string) ([]string, []uint64, error) {

	sql := "select max(" + TableProps.feedback.id + ")," + TableProps.feedback.question + " from " + TableProps.feedback.name +
		" where " + TableProps.feedback.createdTime + ">=FROM_UNIXTIME(?) and " +
		TableProps.feedback.createdTime + " <=FROM_UNIXTIME(?) and " + TableProps.feedback.appid + "=? " +
		"group by " + TableProps.feedback.question + " limit " + util.GetEnviroment(Envs, "MAX_NUM_TO_CLUSTER")

	feedbackQs := make([]string, 0, MaxNumToCluster)
	feedbackQID := make([]uint64, 0, MaxNumToCluster)

	rows, err := sqlQuery(sql, s.Unix(), e.Unix(), appid)
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

	nativeLog.GetWordPos(NluURL, feedbackQs, feedbackQID)
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

	clusters := make([]clustering, 0)

	if len(embeddings) == 0 {
		util.LogError.Println("This task is going to be closed, because: [effective embedding size is 0]")
		clusters = append(clusters, clustering{feedbackID: feedbackQID, tags: make([]string, 0)})
	} else {
		clusterIdxes := doCluster(embeddings, 10)
		clusterMap := make(map[string]clustering, 0)

		for _, idxes := range clusterIdxes {
			feedbackQIDs := make([]uint64, 0)
			clusterTags := make([]string, 0)
			clusterPos := make([]string, 0)

			for _, idx := range idxes {

				feedbackQIDs = append(feedbackQIDs, nativeLog.Logs[idx].ContentID)
				for _, v := range nativeLog.Logs[idx].KeyWords {
					clusterPos = append(clusterPos, v)
				}
			}
			tags := data.ExtractTags(clusterPos, 2)
			for _, v := range tags {
				clusterTags = append(clusterTags, v.Text())
			}

			sort.Strings(clusterTags)

			tagKey := strings.Join(clusterTags, "|")

			if c, ok := clusterMap[tagKey]; ok {
				c.feedbackID = append(c.feedbackID, feedbackQIDs...)
				clusterMap[tagKey] = c
			} else {
				var c clustering
				c.feedbackID = feedbackQIDs
				c.tags = clusterTags
				clusterMap[tagKey] = c
			}

			//clusters = append(clusters, clustering{feedbackID: feedbackQIDs, tags: clusterTags})
		}

		for _, v := range clusterMap {
			clusters = append(clusters, clustering{feedbackID: v.feedbackID, tags: v.tags})
		}
	}
	return &clusteringResult{numClustered: len(nativeLog.Logs), clusters: clusters}
}

func doCluster(vectors []model.Vector, topN int) [][]int {
	util.LogTrace.Println("Clustering starts.")

	numClusters := len(vectors) / ClusteringBatch

	if numClusters == 0 {
		numClusters = 1
	}

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

func getRecommend(sentence []string) ([]*RecommendQ, error) {
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
					r, err := response.Post(s)
					if err != nil {
						util.LogError.Println(err)
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

	questionIDMap, err := GetQuestionIDByContent(sorter.sliceData)
	if err != nil {
		return nil, err
	}

	recommend := make([]*RecommendQ, 0, len(sorter.sliceData))

	for i := 0; i < len(sorter.sliceData); i++ {
		if id, ok := questionIDMap[sorter.sliceData[i].(string)]; ok {
			rQ := &RecommendQ{QID: id, Content: sorter.sliceData[i].(string)}
			recommend = append(recommend, rQ)
		} else {
			util.LogWarn.Printf("[SelfLearn][Recommend] has %s question but doesn't have id\n", sorter.sliceData[i].(string))
		}
	}

	return recommend, nil
}
