 package v1
 
 import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"

	"emotibot.com/emotigo/module/admin-api/ELKStats/data"
	dataCommon "emotibot.com/emotigo/module/admin-api/ELKStats/data/common"
	dataV1 "emotibot.com/emotigo/module/admin-api/ELKStats/data/v1"
	"emotibot.com/emotigo/module/admin-api/ELKStats/services"
	"emotibot.com/emotigo/module/admin-api/util/elasticsearch"
	esData "emotibot.com/emotigo/module/admin-api/util/elasticsearch/data"
	"github.com/olivere/elastic"
)

func createCallStatsDateHistogramAggregation(query dataV1.CallStatsQuery) *elastic.DateHistogramAggregation {
	agg := services.CreateDateHistogramAggregation(query.CommonQuery, data.SessionEndTimeFieldName)
	agg.Interval(query.AggInterval)
	return agg
}

func doCallStatsDateHistogramAggService(ctx context.Context, client *elastic.Client,
	query elastic.Query, aggName string, agg elastic.Aggregation) (map[string]interface{}, error) {
	index := fmt.Sprintf("%s-*", esData.ESSessionsIndex)
	result, err := services.CreateSearchService(ctx, client, query, index, esData.ESSessionsType, aggName, agg)
	if err != nil {
		return nil, err
	}

	counts := services.ExtractCountsFromAggDateHistogramBuckets(result, aggName)
	return counts, nil
}

func calculateCallRates(totalCallCounts map[string]interface{},
	targetCallCounts map[string]interface{}) map[string]interface{} {
	counts := make(map[string]interface{})

	for datetime, totalCallCount := range totalCallCounts {
		if totalCallCount.(int64) == 0 {
			counts[datetime] = "N/A"
		} else if targetCallCount, ok := targetCallCounts[datetime]; ok {
			counts[datetime] = strconv.FormatFloat((float64(targetCallCount.(int64)) /
				float64(totalCallCount.(int64))), 'f', 2, 64)
		}
	}

	return counts
}

func TotalCallCounts(query dataV1.CallStatsQuery) (map[string]interface{}, error) {
	ctx, client := elasticsearch.GetClient()
	aggName := "total_calls"
	boolQuery := services.CreateBoolQuery(query.CommonQuery)
	rangeQuery := services.CreateRangeQuery(query.CommonQuery, data.SessionEndTimeFieldName)
	boolQuery = boolQuery.Filter(rangeQuery)

	dateHistogramAgg := createCallStatsDateHistogramAggregation(query)
	return doCallStatsDateHistogramAggService(ctx, client, boolQuery, aggName, dateHistogramAgg)
}

func CompleteCallCounts(query dataV1.CallStatsQuery) (map[string]interface{}, error) {
	ctx, client := elasticsearch.GetClient()
	aggName := "complete_calls"
	boolQuery := services.CreateBoolQuery(query.CommonQuery)
	termQuery := elastic.NewTermQuery("status", dataCommon.CallStatusComplete)
	rangeQuery := services.CreateRangeQuery(query.CommonQuery, data.SessionEndTimeFieldName)
	boolQuery = boolQuery.Must(termQuery)
	boolQuery = boolQuery.Filter(rangeQuery)

	dateHistogramAgg := createCallStatsDateHistogramAggregation(query)
	return doCallStatsDateHistogramAggService(ctx, client, boolQuery, aggName, dateHistogramAgg)
}

func CompleteCallRates(completeCallCounts map[string]interface{},
	totalCallCounts map[string]interface{}) map[string]interface{} {
	return calculateCallRates(totalCallCounts, completeCallCounts)
}

func ToHumanCallCounts(query dataV1.CallStatsQuery) (map[string]interface{}, error) {
	ctx, client := elasticsearch.GetClient()
	aggName := "to_human_calls"
	boolQuery := services.CreateBoolQuery(query.CommonQuery)
	termQuery := elastic.NewTermQuery("status", dataCommon.CallStatusTranserToHuman)
	rangeQuery := services.CreateRangeQuery(query.CommonQuery, data.SessionEndTimeFieldName)
	boolQuery = boolQuery.Must(termQuery)
	boolQuery = boolQuery.Filter(rangeQuery)

	dateHistogramAgg := createCallStatsDateHistogramAggregation(query)
	return doCallStatsDateHistogramAggService(ctx, client, boolQuery, aggName, dateHistogramAgg)
}

func ToHumanCallRates(toHumanCounts map[string]interface{},
	totalCallCounts map[string]interface{}) map[string]interface{} {
	return calculateCallRates(totalCallCounts, toHumanCounts)
}

func TimeoutCallCounts(query dataV1.CallStatsQuery) (map[string]interface{}, error) {
	ctx, client := elasticsearch.GetClient()
	aggName := "timeout_calls"
	boolQuery := services.CreateBoolQuery(query.CommonQuery)
	termQuery := elastic.NewTermQuery("status", dataCommon.CallStatusTimeout)
	rangeQuery := services.CreateRangeQuery(query.CommonQuery, data.SessionEndTimeFieldName)
	boolQuery = boolQuery.Must(termQuery)
	boolQuery = boolQuery.Filter(rangeQuery)

	dateHistogramAgg := createCallStatsDateHistogramAggregation(query)
	return doCallStatsDateHistogramAggService(ctx, client, boolQuery, aggName, dateHistogramAgg)
}

func TimeoutCallRates(timeoutCallCounts map[string]interface{},
	totalCallCounts map[string]interface{}) map[string]interface{} {
	return calculateCallRates(totalCallCounts, timeoutCallCounts)
}

func CancelCallCounts(query dataV1.CallStatsQuery) (map[string]interface{}, error) {
	ctx, client := elasticsearch.GetClient()
	aggName := "cancel_calls"
	boolQuery := services.CreateBoolQuery(query.CommonQuery)
	termQuery := elastic.NewTermQuery("status", dataCommon.CallStatusCancel)
	rangeQuery := services.CreateRangeQuery(query.CommonQuery, data.SessionEndTimeFieldName)
	boolQuery = boolQuery.Must(termQuery)
	boolQuery = boolQuery.Filter(rangeQuery)

	dateHistogramAgg := createCallStatsDateHistogramAggregation(query)
	return doCallStatsDateHistogramAggService(ctx, client, boolQuery, aggName, dateHistogramAgg)
}

func CancelCallRates(cancelCallCounts map[string]interface{},
	totalCallCounts map[string]interface{}) map[string]interface{} {
	return calculateCallRates(totalCallCounts, cancelCallCounts)
}

func Unknowns(query dataV1.CallStatsQuery) (map[string]interface{}, error) {
	ctx, client := elasticsearch.GetClient()
	// TODO: Metric definition not defined yet, return maps with all zero values now
	aggName := "unknowns"
	boolQuery := services.CreateBoolQuery(query.CommonQuery)
	termQuery := elastic.NewTermQuery("status", -10)
	rangeQuery := services.CreateRangeQuery(query.CommonQuery, data.SessionEndTimeFieldName)
	boolQuery = boolQuery.Must(termQuery)
	boolQuery = boolQuery.Filter(rangeQuery)

	dateHistogramAgg := createCallStatsDateHistogramAggregation(query)
	return doCallStatsDateHistogramAggService(ctx, client, boolQuery, aggName, dateHistogramAgg)
}

func TopToHumanAnswers(query dataV1.CallStatsQuery, topN int) (dataV1.ToHumanAnswers, error) {
	ctx, client := elasticsearch.GetClient()
	// Query all the session IDs in the query date range
	groupBySessionsAggName := "group_by_sessions"
	boolQuery := services.CreateBoolQuery(query.CommonQuery)
	rangeQuery := services.CreateRangeQuery(query.CommonQuery, data.SessionEndTimeFieldName)
	toHumanTermQuery := elastic.NewTermQuery("status", dataCommon.CallStatusTranserToHuman)
	boolQuery = boolQuery.Filter(rangeQuery)
	boolQuery = boolQuery.Filter(toHumanTermQuery)

	groupBySessionsAgg := elastic.NewTermsAggregation().
		Field("session_id").
		Size(data.ESTermAggSize)

	index := fmt.Sprintf("%s-*", esData.ESSessionsIndex)

	result, err := client.Search().
		Index(index).
		Type(esData.ESSessionsType).
		Query(boolQuery).
		Aggregation(groupBySessionsAggName, groupBySessionsAgg).
		Size(0).
		Do(ctx)
	if err != nil {
		return nil, err
	}

	vaildRecordCount := 2
	sessionIDs := make([]interface{}, 0)
	toHumanAnswersMap := make(map[string]int64)
	toHumanAnswers := make(dataV1.ToHumanAnswers, 0)

	if agg, found := result.Aggregations.Terms(groupBySessionsAggName); found {
		for _, bucket := range agg.Buckets {
			sessionIDs = append(sessionIDs, bucket.Key.(string))
		}
	}

	if len(sessionIDs) == 0 {
		return toHumanAnswers, nil
	}

	// Query the latest 'two' records in the quried sessions
	latestRecordsAggName := "latest_records"
	termsQuery := elastic.NewTermsQuery("session_id", sessionIDs...)

	latestRecordsTopHitAgg := elastic.NewTopHitsAggregation().Sort(data.LogTimeFieldName, false).Size(vaildRecordCount)
	groupBySessionsAgg.SubAggregation(latestRecordsAggName, latestRecordsTopHitAgg)

	index = fmt.Sprintf("%s-*", esData.ESRecordsIndex)

	result, err = client.Search().
		Index(index).
		Type(esData.ESRecordType).
		Query(termsQuery).
		Aggregation(groupBySessionsAggName, groupBySessionsAgg).
		Size(0).
		Do(ctx)
	if err != nil {
		return nil, err
	}

	if agg, found := result.Aggregations.Terms(groupBySessionsAggName); found {
		for _, bucket := range agg.Buckets {
			recordBuckets, found := bucket.TopHits(latestRecordsAggName)
			if !found {
				return nil, esData.ErrESTopHitNotFound
			}

			if recordBuckets.Hits.TotalHits < int64(vaildRecordCount) {
				continue
			}

			record := dataV1.Record{}
			jsonStr, err := recordBuckets.Hits.Hits[1].Source.MarshalJSON()
			if err != nil {
				return nil, err
			}

			err = json.Unmarshal(jsonStr, &record)
			if err != nil {
				return nil, err
			}

			if record.Answer != nil && len(record.Answer) > 0 {
				count, ok := toHumanAnswersMap[record.Answer[0].Value]
				if !ok {
					toHumanAnswersMap[record.Answer[0].Value] = 1
				} else {
					toHumanAnswersMap[record.Answer[0].Value] = count + 1
				}
			}
		}
	}

	for answer, count := range toHumanAnswersMap {
		answer := dataV1.ToHumanAnswer{
			Answer: answer,
			Count:  count,
		}
		toHumanAnswers = append(toHumanAnswers, answer)
	}

	// Sort by counts
	sort.Sort(sort.Reverse(toHumanAnswers))
	return toHumanAnswers, nil
}
