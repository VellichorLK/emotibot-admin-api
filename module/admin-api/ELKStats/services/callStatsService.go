package services

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"

	"emotibot.com/emotigo/module/admin-api/ELKStats/data"
	"emotibot.com/emotigo/module/admin-api/util/elasticsearch"
	"github.com/olivere/elastic"
)

func TotalCallCounts(query data.CallStatsQuery) (map[string]interface{}, error) {
	ctx, client := elasticsearch.GetClient()
	aggName := "total_calls"
	boolQuery := elastic.NewBoolQuery()
	rangeQuery := createRangeQuery(query.CommonQuery, data.SessionEndTimeFieldName)
	boolQuery = boolQuery.Filter(rangeQuery)

	dateHistogramAgg := createCallStatsDateHistogramAggregation(query)
	return doCallStatsDateHistogramAggService(ctx, client, query.AppID, boolQuery, aggName, dateHistogramAgg)
}

func CompleteCallCounts(query data.CallStatsQuery) (map[string]interface{}, error) {
	ctx, client := elasticsearch.GetClient()
	aggName := "complete_calls"
	boolQuery := elastic.NewBoolQuery()
	termQuery := elastic.NewTermQuery("status", data.CallStatusComplete)
	rangeQuery := createRangeQuery(query.CommonQuery, data.SessionEndTimeFieldName)
	boolQuery = boolQuery.Must(termQuery)
	boolQuery = boolQuery.Filter(rangeQuery)

	dateHistogramAgg := createCallStatsDateHistogramAggregation(query)
	return doCallStatsDateHistogramAggService(ctx, client, query.AppID, boolQuery, aggName, dateHistogramAgg)
}

func CompleteCallRates(completeCallCounts map[string]interface{},
	totalCallCounts map[string]interface{}) map[string]interface{} {
	return calculateCallRates(totalCallCounts, completeCallCounts)
}

func ToHumanCallCounts(query data.CallStatsQuery) (map[string]interface{}, error) {
	ctx, client := elasticsearch.GetClient()
	aggName := "to_human_calls"
	boolQuery := elastic.NewBoolQuery()
	termQuery := elastic.NewTermQuery("status", data.CallStatusTranserToHuman)
	rangeQuery := createRangeQuery(query.CommonQuery, data.SessionEndTimeFieldName)
	boolQuery = boolQuery.Must(termQuery)
	boolQuery = boolQuery.Filter(rangeQuery)

	dateHistogramAgg := createCallStatsDateHistogramAggregation(query)
	return doCallStatsDateHistogramAggService(ctx, client, query.AppID, boolQuery, aggName, dateHistogramAgg)
}

func ToHumanCallRates(toHumanCounts map[string]interface{},
	totalCallCounts map[string]interface{}) map[string]interface{} {
	return calculateCallRates(totalCallCounts, toHumanCounts)
}

func TimeoutCallCounts(query data.CallStatsQuery) (map[string]interface{}, error) {
	ctx, client := elasticsearch.GetClient()
	aggName := "timeout_calls"
	boolQuery := elastic.NewBoolQuery()
	termQuery := elastic.NewTermQuery("status", data.CallStatusTimeout)
	rangeQuery := createRangeQuery(query.CommonQuery, data.SessionEndTimeFieldName)
	boolQuery = boolQuery.Must(termQuery)
	boolQuery = boolQuery.Filter(rangeQuery)

	dateHistogramAgg := createCallStatsDateHistogramAggregation(query)
	return doCallStatsDateHistogramAggService(ctx, client, query.AppID, boolQuery, aggName, dateHistogramAgg)
}

func TimeoutCallRates(timeoutCallCounts map[string]interface{},
	totalCallCounts map[string]interface{}) map[string]interface{} {
	return calculateCallRates(totalCallCounts, timeoutCallCounts)
}

func CancelCallCounts(query data.CallStatsQuery) (map[string]interface{}, error) {
	ctx, client := elasticsearch.GetClient()
	aggName := "cancel_calls"
	boolQuery := elastic.NewBoolQuery()
	termQuery := elastic.NewTermQuery("status", data.CallStatusCancel)
	rangeQuery := createRangeQuery(query.CommonQuery, data.SessionEndTimeFieldName)
	boolQuery = boolQuery.Must(termQuery)
	boolQuery = boolQuery.Filter(rangeQuery)

	dateHistogramAgg := createCallStatsDateHistogramAggregation(query)
	return doCallStatsDateHistogramAggService(ctx, client, query.AppID, boolQuery, aggName, dateHistogramAgg)
}

func CancelCallRates(cancelCallCounts map[string]interface{},
	totalCallCounts map[string]interface{}) map[string]interface{} {
	return calculateCallRates(totalCallCounts, cancelCallCounts)
}

func Unknowns(query data.CallStatsQuery) (map[string]interface{}, error) {
	ctx, client := elasticsearch.GetClient()
	// TODO: Metric definition not defined yet, return maps with all zero values now
	aggName := "unknowns"
	boolQuery := elastic.NewBoolQuery()
	termQuery := elastic.NewTermQuery("status", -10)
	rangeQuery := createRangeQuery(query.CommonQuery, data.SessionEndTimeFieldName)
	boolQuery = boolQuery.Must(termQuery)
	boolQuery = boolQuery.Filter(rangeQuery)

	dateHistogramAgg := createCallStatsDateHistogramAggregation(query)
	return doCallStatsDateHistogramAggService(ctx, client, query.AppID, boolQuery, aggName, dateHistogramAgg)
}

func TopToHumanAnswers(query data.CallStatsQuery, topN int) (data.ToHumanAnswers, error) {
	ctx, client := elasticsearch.GetClient()
	// Query all the session IDs in the query date range
	groupBySessionsAggName := "group_by_sessions"
	boolQuery := elastic.NewBoolQuery()
	rangeQuery := createRangeQuery(query.CommonQuery, data.SessionEndTimeFieldName)
	toHumanTermQuery := elastic.NewTermQuery("status", data.CallStatusTranserToHuman)
	boolQuery = boolQuery.Filter(rangeQuery)
	boolQuery = boolQuery.Filter(toHumanTermQuery)

	groupBySessionsAgg := elastic.NewTermsAggregation().
		Field("session_id").
		Size(data.ESTermAggSize)

	index := fmt.Sprintf("%s-%s-*", data.ESSessionsIndex, query.AppID)

	result, err := client.Search().
		Index(index).
		Type(data.ESSessionsType).
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
	toHumanAnswers := make(data.ToHumanAnswers, 0)

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

	index = fmt.Sprintf("%s-%s-*", data.ESRecordsIndex, query.AppID)

	result, err = client.Search().
		Index(index).
		Type(data.ESRecordType).
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
				return nil, data.ErrESTopHitNotFound
			}

			if recordBuckets.Hits.TotalHits < int64(vaildRecordCount) {
				continue
			}

			record := data.Record{}
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
		answer := data.ToHumanAnswer{
			Answer: answer,
			Count:  count,
		}
		toHumanAnswers = append(toHumanAnswers, answer)
	}

	// Sort by counts
	sort.Sort(sort.Reverse(toHumanAnswers))
	return toHumanAnswers, nil
}

func createCallStatsDateHistogramAggregation(query data.CallStatsQuery) *elastic.DateHistogramAggregation {
	agg := createDateHistogramAggregation(query.CommonQuery, data.SessionEndTimeFieldName)
	agg.Interval(query.AggInterval)
	return agg
}

func doCallStatsDateHistogramAggService(ctx context.Context, client *elastic.Client, appID string,
	query elastic.Query, aggName string, agg elastic.Aggregation) (map[string]interface{}, error) {
	index := fmt.Sprintf("%s-%s-*", data.ESSessionsIndex, appID)
	result, err := createSearchService(ctx, client, query, index, data.ESSessionsType, aggName, agg)
	if err != nil {
		return nil, err
	}

	counts := extractCountsFromAggDateHistogramBuckets(result, aggName)
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
