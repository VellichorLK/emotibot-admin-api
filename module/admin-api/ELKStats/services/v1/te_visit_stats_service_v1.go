package v1

import (
	"context"
	"fmt"

	"emotibot.com/emotigo/module/admin-api/ELKStats/data"
	dataV1 "emotibot.com/emotigo/module/admin-api/ELKStats/data/v1"
	"emotibot.com/emotigo/module/admin-api/ELKStats/services"
	"emotibot.com/emotigo/module/admin-api/ELKStats/services/common"
	"emotibot.com/emotigo/module/admin-api/util/elasticsearch"
	"github.com/olivere/elastic"
)

func TriggerCounts(query dataV1.TEVisitStatsQuery) (map[string]interface{}, error) {
	ctx, client := elasticsearch.GetClient()
	aggName := "triggers"
	boolQuery := createTEVisitStatsBoolQuery(query)
	rangeQuery := services.CreateRangeQueryUnixTime(query.StatsQuery.CommonQuery,
		data.TERecordsTriggerTimeFieldName)
	boolQuery.Filter(rangeQuery)

	switch query.AggBy {
	case data.AggByTime:
		dateHistogramAgg := createTEVisitStatsDateHistogramAggregation(query)
		return doTEVisitStatsDateHistogramAggService(ctx, client, boolQuery, aggName, dateHistogramAgg)
	case data.AggByTag:
		tagExistsQuery := common.CreateStatsTagExistsQuery(query.AggTagType)
		boolQuery = boolQuery.Filter(tagExistsQuery)
		tagTermAgg := common.CreateStatsTagTermsAggregation(query.AggTagType)

		counts, err := doTEVisitStatsTermsAggService(ctx, client, boolQuery, aggName, tagTermAgg)
		if err != nil {
			return nil, err
		}

		return common.NormalizeTagCounts(counts, query.AppID, query.AggTagType)
	default:
		return nil, data.ErrInvalidAggType
	}
}

func UnfinishedCounts(query dataV1.TEVisitStatsQuery) (map[string]interface{}, error) {
	ctx, client := elasticsearch.GetClient()
	aggName := "unfinished"
	boolQuery := createTEVisitStatsBoolQuery(query)
	rangeQuery := services.CreateRangeQueryUnixTime(query.CommonQuery, data.TERecordsTriggerTimeFieldName)
	finishTimeExistQuery := elastic.NewExistsQuery("finish_time")
	boolQuery.Filter(rangeQuery)
	boolQuery.MustNot(finishTimeExistQuery)

	switch query.AggBy {
	case data.AggByTime:
		dateHistogramAgg := createTEVisitStatsDateHistogramAggregation(query)
		return doTEVisitStatsDateHistogramAggService(ctx, client, boolQuery, aggName, dateHistogramAgg)
	case data.AggByTag:
		tagExistsQuery := common.CreateStatsTagExistsQuery(query.AggTagType)
		boolQuery = boolQuery.Filter(tagExistsQuery)
		tagTermAgg := common.CreateStatsTagTermsAggregation(query.AggTagType)

		counts, err := doTEVisitStatsTermsAggService(ctx, client, boolQuery, aggName, tagTermAgg)
		if err != nil {
			return nil, err
		}

		return common.NormalizeTagCounts(counts, query.AppID, query.AggTagType)
	default:
		return nil, data.ErrInvalidAggType
	}
}

func createTEVisitStatsBoolQuery(query dataV1.TEVisitStatsQuery) *elastic.BoolQuery {
	boolQuery := services.CreateBoolQuery(query.StatsQuery.CommonQuery)

	if query.ScenarioID != "" {
		scenarioIDTermQuery := elastic.NewTermQuery("scenario_id", query.ScenarioID)
		boolQuery.Filter(scenarioIDTermQuery)
	}

	if query.ScenarioName != "" {
		scenarioNameTermQuery := elastic.NewTermQuery("scenario_name", query.ScenarioName)
		boolQuery.Filter(scenarioNameTermQuery)
	}

	return boolQuery
}

func createTEVisitStatsDateHistogramAggregation(query dataV1.TEVisitStatsQuery) *elastic.DateHistogramAggregation {
	agg := services.CreateDateHistogramAggregation(query.CommonQuery, data.TERecordsTriggerTimeFieldName)
	agg.Interval(query.AggInterval)
	return agg
}

func doTEVisitStatsDateHistogramAggService(ctx context.Context, client *elastic.Client, query elastic.Query,
	aggName string, agg elastic.Aggregation) (map[string]interface{}, error) {
	index := fmt.Sprintf("%s-*", data.ESTERecordsIndex)
	result, err := services.CreateSearchService(ctx, client, query, index, data.ESTERecordsType, aggName, agg)
	if err != nil {
		return nil, err
	}

	counts := services.ExtractCountsFromAggDateHistogramBuckets(result, aggName)
	return counts, nil
}

func doTEVisitStatsTermsAggService(ctx context.Context, client *elastic.Client, query elastic.Query,
	aggName string, agg elastic.Aggregation) (map[string]interface{}, error) {
	index := fmt.Sprintf("%s-*", data.ESTERecordsIndex)
	return services.DoTermsAggService(ctx, client, query, index, data.ESTERecordsType, aggName, agg)
}
