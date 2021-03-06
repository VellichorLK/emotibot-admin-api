package v1

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"emotibot.com/emotigo/module/admin-api/ELKStats/data"
	dataCommon "emotibot.com/emotigo/module/admin-api/ELKStats/data/common"
	dataV1 "emotibot.com/emotigo/module/admin-api/ELKStats/data/v1"
	"emotibot.com/emotigo/module/admin-api/ELKStats/services"
	"emotibot.com/emotigo/module/admin-api/ELKStats/services/common"
	"emotibot.com/emotigo/module/admin-api/util/elasticsearch"
	esData "emotibot.com/emotigo/module/admin-api/util/elasticsearch/data"
	elastic "gopkg.in/olivere/elastic.v6"
)

func ConversationCounts(query dataV1.VisitStatsQuery) (map[string]interface{}, error) {
	ctx, client := elasticsearch.GetClient()
	aggName := "conversations"
	boolQuery := services.CreateBoolQuery(query.CommonQuery)
	rangeQuery := services.CreateRangeQuery(query.CommonQuery, data.SessionStartTimeFieldName)
	boolQuery = boolQuery.Filter(rangeQuery)

	switch query.AggBy {
	case data.AggByTime:
		dateHistogramAgg := services.CreateDateHistogramAggregation(query.CommonQuery, data.SessionStartTimeFieldName).
			Interval(query.AggInterval)

		index := fmt.Sprintf("%s-*", esData.ESSessionsIndex)
		result, err := services.CreateSearchService(ctx, client, boolQuery, index, esData.ESSessionsType, aggName, dateHistogramAgg)
		if err != nil {
			return nil, err
		}

		counts := services.ExtractCountsFromAggDateHistogramBuckets(result, aggName)
		return counts, nil
	case data.AggByTag:
		tagExistsQuery := common.CreateStatsTagExistsQuery(query.AggTagType)
		boolQuery = boolQuery.Filter(tagExistsQuery)
		tagTermAgg := common.CreateStatsTagTermsAggregation(query.AggTagType)

		index := fmt.Sprintf("%s-*", esData.ESSessionsIndex)
		result, err := services.CreateSearchService(ctx, client, boolQuery, index, esData.ESSessionsType, aggName, tagTermAgg)
		if err != nil {
			return nil, err
		}

		counts := services.ExtractCountsFromAggTermsBuckets(result, aggName)
		return common.NormalizeTagCounts(counts, query.AppID, query.AggTagType)
	default:
		return nil, data.ErrInvalidAggType
	}
}

func UniqueUserCounts(query dataV1.VisitStatsQuery) (map[string]interface{}, error) {
	ctx, client := elasticsearch.GetClient()
	aggName := "unique_users"
	boolQuery := common.CreateStatsBoolQuery(query.CommonQuery)
	rangeQuery := services.CreateRangeQuery(query.CommonQuery, data.LogTimeFieldName)
	boolQuery = boolQuery.Filter(rangeQuery)

	uniqueUserCountAggName := "unique_user_count"
	uniqueUserCardinalityAgg := elastic.NewCardinalityAggregation()
	uniqueUserCardinalityAgg.Field("user_id")

	var _agg elastic.Aggregation

	switch query.AggBy {
	case data.AggByTime:
		dateHistogramAgg := createVisitStatsDateHistogramAggregation(query)
		dateHistogramAgg.SubAggregation(uniqueUserCountAggName, uniqueUserCardinalityAgg)

		_agg = dateHistogramAgg
	case data.AggByTag:
		boolQuery = boolQuery.Filter(rangeQuery)
		tagExistsQuery := common.CreateStatsTagExistsQuery(query.AggTagType)
		boolQuery = boolQuery.Filter(tagExistsQuery)
		tagTermAgg := common.CreateStatsTagTermsAggregation(query.AggTagType)
		tagTermAgg.SubAggregation(uniqueUserCountAggName, uniqueUserCardinalityAgg)

		_agg = tagTermAgg
	default:
		return nil, data.ErrInvalidAggType
	}

	result, err := createVisitStatsSearchService(ctx, client, boolQuery, aggName, _agg)
	if err != nil {
		return nil, err
	}

	uniqueUserCounts := make(map[string]interface{})

	if agg, found := result.Aggregations.Terms(aggName); found {
		for _, bucket := range agg.Buckets {
			uniqueUserCount, found := bucket.Cardinality(uniqueUserCountAggName)
			if !found {
				return nil, esData.ErrESCardinalityNotFound
			}

			var bucketKey string

			switch query.AggBy {
			case data.AggByTime:
				bucketKey = *bucket.KeyAsString
			case data.AggByTag:
				bucketKey = bucket.Key.(string)
			default:
				return nil, data.ErrInvalidAggType
			}

			uniqueUserCounts[bucketKey] = int64(*uniqueUserCount.Value)
		}
	}

	return uniqueUserCounts, nil
}

func NewUserCounts(query dataV1.VisitStatsQuery) (map[string]interface{}, error) {
	ctx, client := elasticsearch.GetClient()
	aggName := "new_users"
	boolQuery := common.CreateStatsBoolQuery(query.CommonQuery)
	rangeQuery := services.CreateRangeQuery(query.CommonQuery, data.FirstLogTimeFieldName)
	boolQuery = boolQuery.Filter(rangeQuery)

	uniqueUserCountAggName := "unique_user_count"
	uniqueUserCardinalityAgg := elastic.NewCardinalityAggregation()
	uniqueUserCardinalityAgg.Field("user_id")

	var _agg elastic.Aggregation

	switch query.AggBy {
	case data.AggByTime:
		dateHistogramAgg := services.CreateDateHistogramAggregation(query.CommonQuery, data.FirstLogTimeFieldName).
			Interval(query.AggInterval)
		dateHistogramAgg.SubAggregation(uniqueUserCountAggName, uniqueUserCardinalityAgg)

		_agg = dateHistogramAgg
	case data.AggByTag:
		tagExistsQuery := common.CreateStatsTagExistsQuery(query.AggTagType)
		boolQuery = boolQuery.Filter(tagExistsQuery)
		tagTermAgg := common.CreateStatsTagTermsAggregation(query.AggTagType).ShardSize(data.ESTermAggShardSize)
		tagTermAgg.SubAggregation(uniqueUserCountAggName, uniqueUserCardinalityAgg)

		_agg = tagTermAgg
	default:
		return nil, data.ErrInvalidAggType
	}

	index := fmt.Sprintf("%s-*", esData.ESUsersIndex)
	result, err := services.CreateSearchService(ctx, client, boolQuery, index, esData.ESUsersType, aggName, _agg)
	if err != nil {
		return nil, err
	}

	newUserCounts := make(map[string]interface{})

	if agg, found := result.Aggregations.Terms(aggName); found {
		for _, bucket := range agg.Buckets {
			uniqueUserCount, found := bucket.Cardinality(uniqueUserCountAggName)
			if !found {
				return nil, esData.ErrESCardinalityNotFound
			}

			var bucketKey string

			switch query.AggBy {
			case data.AggByTime:
				bucketKey = *bucket.KeyAsString
			case data.AggByTag:
				bucketKey = bucket.Key.(string)
			default:
				return nil, data.ErrInvalidAggType
			}

			newUserCounts[bucketKey] = int64(*uniqueUserCount.Value)
		}
	}

	if query.AggBy == data.AggByTag {
		newUserCounts, err = common.NormalizeTagCounts(newUserCounts, query.AppID, query.AggTagType)
		if err != nil {
			return nil, err
		}
	}

	return newUserCounts, nil
}

func TotalAskCounts(query dataV1.VisitStatsQuery) (map[string]interface{}, error) {
	ctx, client := elasticsearch.GetClient()
	aggName := "total_asks"
	boolQuery := common.CreateStatsBoolQuery(query.CommonQuery)
	rangeQuery := services.CreateRangeQuery(query.CommonQuery, data.LogTimeFieldName)
	boolQuery = boolQuery.Filter(rangeQuery)

	switch query.AggBy {
	case data.AggByTime:
		dateHistogramAgg := createVisitStatsDateHistogramAggregation(query)
		return doVisitStatsDateHistogramAggService(ctx, client, boolQuery, aggName, dateHistogramAgg)
	case data.AggByTag:
		tagExistsQuery := common.CreateStatsTagExistsQuery(query.AggTagType)
		boolQuery = boolQuery.Filter(tagExistsQuery)
		tagTermAgg := common.CreateStatsTagTermsAggregation(query.AggTagType)

		counts, err := doVisitStatsTermsAggService(ctx, client, boolQuery, aggName, tagTermAgg)
		if err != nil {
			return nil, err
		}

		return common.NormalizeTagCounts(counts, query.AppID, query.AggTagType)
	default:
		return nil, data.ErrInvalidAggType
	}
}

func NormalResponseCounts(query dataV1.VisitStatsQuery) (map[string]interface{}, error) {
	ctx, client := elasticsearch.GetClient()
	aggName := "normal_responses"
	boolQuery := common.CreateStatsBoolQuery(query.CommonQuery)
	termsQuery := elastic.NewTermsQuery("module", "faq", "task_engine", "domain_kg")
	rangeQuery := services.CreateRangeQuery(query.CommonQuery, data.LogTimeFieldName)
	boolQuery = boolQuery.Filter(termsQuery)
	boolQuery = boolQuery.Filter(rangeQuery)

	switch query.AggBy {
	case data.AggByTime:
		dateHistogramAgg := createVisitStatsDateHistogramAggregation(query)
		return doVisitStatsDateHistogramAggService(ctx, client, boolQuery, aggName, dateHistogramAgg)
	case data.AggByTag:
		tagExistsQuery := common.CreateStatsTagExistsQuery(query.AggTagType)
		boolQuery = boolQuery.Filter(tagExistsQuery)
		tagTermAgg := common.CreateStatsTagTermsAggregation(query.AggTagType)

		counts, err := doVisitStatsTermsAggService(ctx, client, boolQuery, aggName, tagTermAgg)
		if err != nil {
			return nil, err
		}

		return common.NormalizeTagCounts(counts, query.AppID, query.AggTagType)
	default:
		return nil, data.ErrInvalidAggType
	}
}

func ChatCounts(query dataV1.VisitStatsQuery) (map[string]interface{}, error) {
	ctx, client := elasticsearch.GetClient()
	aggName := "chats"
	boolQuery := common.CreateStatsBoolQuery(query.CommonQuery)
	termQuery := elastic.NewTermQuery("module", "chat")
	rangeQuery := services.CreateRangeQuery(query.CommonQuery, data.LogTimeFieldName)
	boolQuery = boolQuery.Filter(termQuery)
	boolQuery = boolQuery.Filter(rangeQuery)

	switch query.AggBy {
	case data.AggByTime:
		dateHistogramAgg := createVisitStatsDateHistogramAggregation(query)
		return doVisitStatsDateHistogramAggService(ctx, client, boolQuery, aggName, dateHistogramAgg)
	case data.AggByTag:
		tagExistsQuery := common.CreateStatsTagExistsQuery(query.AggTagType)
		boolQuery = boolQuery.Filter(tagExistsQuery)
		tagTermAgg := common.CreateStatsTagTermsAggregation(query.AggTagType)

		counts, err := doVisitStatsTermsAggService(ctx, client, boolQuery, aggName, tagTermAgg)
		if err != nil {
			return nil, err
		}

		return common.NormalizeTagCounts(counts, query.AppID, query.AggTagType)
	default:
		return nil, data.ErrInvalidAggType
	}
}

func OtherCounts(query dataV1.VisitStatsQuery) (map[string]interface{}, error) {
	ctx, client := elasticsearch.GetClient()
	aggName := "others"
	boolQuery := common.CreateStatsBoolQuery(query.CommonQuery)
	termsQuery := elastic.NewTermsQuery("module", "faq", "task_engine", "domain_kg", "knowledge", "chat", "backfill")
	rangeQuery := services.CreateRangeQuery(query.CommonQuery, data.LogTimeFieldName)
	boolQuery = boolQuery.MustNot(termsQuery)
	boolQuery = boolQuery.Filter(rangeQuery)

	switch query.AggBy {
	case data.AggByTime:
		dateHistogramAgg := createVisitStatsDateHistogramAggregation(query)
		return doVisitStatsDateHistogramAggService(ctx, client, boolQuery, aggName, dateHistogramAgg)
	case data.AggByTag:
		tagExistsQuery := common.CreateStatsTagExistsQuery(query.AggTagType)
		boolQuery = boolQuery.Filter(tagExistsQuery)
		tagTermAgg := common.CreateStatsTagTermsAggregation(query.AggTagType)

		counts, err := doVisitStatsTermsAggService(ctx, client, boolQuery, aggName, tagTermAgg)
		if err != nil {
			return nil, err
		}

		return common.NormalizeTagCounts(counts, query.AppID, query.AggTagType)
	default:
		return nil, data.ErrInvalidAggType
	}
}

func UnknownQnACounts(query dataV1.VisitStatsQuery) (map[string]interface{}, error) {
	ctx, client := elasticsearch.GetClient()
	aggName := "unknown_qna"
	boolQuery := common.CreateStatsBoolQuery(query.CommonQuery)
	termQuery := elastic.NewTermQuery("module", "backfill")
	rangeQuery := services.CreateRangeQuery(query.CommonQuery, data.LogTimeFieldName)
	boolQuery = boolQuery.Filter(termQuery)
	boolQuery = boolQuery.Filter(rangeQuery)

	switch query.AggBy {
	case data.AggByTime:
		dateHistogramAgg := createVisitStatsDateHistogramAggregation(query)
		return doVisitStatsDateHistogramAggService(ctx, client, boolQuery, aggName, dateHistogramAgg)
	case data.AggByTag:
		tagExistsQuery := common.CreateStatsTagExistsQuery(query.AggTagType)
		boolQuery = boolQuery.Filter(tagExistsQuery)
		tagTermAgg := common.CreateStatsTagTermsAggregation(query.AggTagType)

		counts, err := doVisitStatsTermsAggService(ctx, client, boolQuery, aggName, tagTermAgg)
		if err != nil {
			return nil, err
		}

		return common.NormalizeTagCounts(counts, query.AppID, query.AggTagType)
	default:
		return nil, data.ErrInvalidAggType
	}
}

func SuccessRates(unknownQnACounts map[string]interface{},
	totalAskCounts map[string]interface{}) map[string]interface{} {
	counts := make(map[string]interface{})

	for datetime, totalAskCount := range totalAskCounts {
		if totalAskCount.(int64) == 0 {
			counts[datetime] = "N/A"
		} else if unknownQnACount, ok := unknownQnACounts[datetime]; ok {
			counts[datetime] = strconv.FormatFloat((float64(totalAskCount.(int64)-unknownQnACount.(int64)) /
				float64(totalAskCount.(int64))), 'f', 2, 64)
		}
	}

	return counts
}

func CoversationsPerSessionCounts(conversationCounts map[string]interface{},
	totalAskCounts map[string]interface{}) map[string]interface{} {
	counts := make(map[string]interface{})

	for datetime, conversationCount := range conversationCounts {
		if conversationCount.(int64) == 0 {
			counts[datetime] = "N/A"
		} else if totalAskCount, ok := totalAskCounts[datetime]; ok {
			counts[datetime] = strconv.FormatFloat((float64(totalAskCount.(int64)) /
				float64(conversationCount.(int64))), 'f', 2, 64)
		}
	}

	return counts
}

func TopQuestions(query dataV1.VisitStatsQuery, topN int) (dataV1.Questions, error) {
	ctx, client := elasticsearch.GetClient()
	aggName := "top_questions"
	boolQuery := common.CreateStatsBoolQuery(query.CommonQuery)
	termQuery := elastic.NewTermQuery("module", "faq")
	rangeQuery := services.CreateRangeQuery(query.CommonQuery, data.LogTimeFieldName)
	stdQEmptyQuery := elastic.NewTermQuery("std_q.keyword", "")
	boolQuery = boolQuery.Filter(termQuery)
	boolQuery = boolQuery.Filter(rangeQuery)
	boolQuery = boolQuery.MustNot(stdQEmptyQuery)

	topQTermAgg := elastic.NewTermsAggregation().Field("std_q.keyword").Size(topN).ShardSize(10000).OrderByCount(false)
	result, err := doVisitStatsTermsAggService(ctx, client, boolQuery, aggName, topQTermAgg)
	if err != nil {
		return nil, err
	}

	questions := make(dataV1.Questions, 0)
	for question, count := range result {
		q := dataV1.Question{
			Question: question,
			Count:    count.(int64),
		}

		questions = append(questions, q)
	}

	return questions, nil
}

func TopUnmatchQuestions(query dataV1.VisitStatsQuery, topN int) ([]*dataV1.UnmatchQuestion, error) {
	ctx, client := elasticsearch.GetClient()
	aggName := "top_unmatch_q"
	boolQuery := common.CreateStatsBoolQuery(query.CommonQuery)
	termQuery := elastic.NewTermQuery("module", "backfill")
	rangeQuery := services.CreateRangeQuery(query.CommonQuery, data.LogTimeFieldName)
	boolQuery = boolQuery.Filter(termQuery, rangeQuery)

	maxLogTimeAggName := "max_log_time"
	minLogTimeAggName := "min_log_time"
	topUnmatchQTermAgg := elastic.NewTermsAggregation().Field("user_q.keyword").Size(topN).ShardSize(10000).OrderByCount(false)
	maxLogTimeAgg := elastic.NewMaxAggregation().Field(data.LogTimeFieldName).Format("yyyy-MM-dd HH:mm:ss")
	minLogTimeAgg := elastic.NewMinAggregation().Field(data.LogTimeFieldName).Format("yyyy-MM-dd HH:mm:ss")
	topUnmatchQTermAgg.SubAggregation(maxLogTimeAggName, maxLogTimeAgg)
	topUnmatchQTermAgg.SubAggregation(minLogTimeAggName, minLogTimeAgg)

	result, err := createVisitStatsSearchService(ctx, client, boolQuery, aggName, topUnmatchQTermAgg)
	if err != nil {
		return nil, err
	}

	unmatchQuestions := make([]*dataV1.UnmatchQuestion, 0)

	if agg, found := result.Aggregations.Terms(aggName); found {
		for _, bucket := range agg.Buckets {
			maxLogTimeBucket, found := bucket.MaxBucket(maxLogTimeAggName)
			if !found {
				return nil, esData.ErrESMaxBucketNotFound
			}

			// Note: We have to convert maximum log_time (UTC+0) to local time
			maxLogTime, err := time.Parse(data.ESTimeFormat, maxLogTimeBucket.ValueAsString)
			if err != nil {
				return nil, err
			}
			maxLogTime = maxLogTime.Local()

			minLogTimeBucket, found := bucket.MinBucket(minLogTimeAggName)
			if !found {
				return nil, esData.ErrESMinBucketNotFound
			}

			// Note: We have to convert minimum log_time (UTC+0) to local time
			minLogTime, err := time.Parse(data.ESTimeFormat, minLogTimeBucket.ValueAsString)
			if err != nil {
				return nil, err
			}
			minLogTime = minLogTime.Local()

			unmatchQuestion := dataV1.UnmatchQuestion{
				Question:   bucket.Key.(string),
				Count:      bucket.DocCount,
				MaxLogTime: maxLogTime.Format(data.ESTimeFormat),
				MinLogTime: minLogTime.Format(data.ESTimeFormat),
			}
			unmatchQuestions = append(unmatchQuestions, &unmatchQuestion)
		}
	}

	return unmatchQuestions, nil
}

func AnswerCategoryCounts(query dataV1.VisitStatsQuery) (map[string]interface{}, error) {
	ctx, client := elasticsearch.GetClient()
	aggName := "answer_categories"
	boolQuery := common.CreateStatsBoolQuery(query.CommonQuery)
	rangeQuery := services.CreateRangeQuery(query.CommonQuery, data.LogTimeFieldName)
	boolQuery = boolQuery.Filter(rangeQuery)

	filtersAgg := elastic.NewFiltersAggregation()

	businessFilterName := dataCommon.CategoryBusiness
	businessTermsQuery := elastic.NewTermsQuery("module", "faq", "task_engine", "domain_kg")

	chatFilterName := dataCommon.CategoryChat
	chatTermQuery := elastic.NewTermQuery("module", "chat")

	otherFilterName := dataCommon.CategoryOther
	otherBoolQuery := elastic.NewBoolQuery()
	otherNotTermsQuery := elastic.NewTermsQuery("module", "faq", "task_engine", "chat")
	otherBoolQuery = otherBoolQuery.MustNot(otherNotTermsQuery)

	filtersAgg = filtersAgg.FilterWithName(businessFilterName, businessTermsQuery)
	filtersAgg = filtersAgg.FilterWithName(chatFilterName, chatTermQuery)
	filtersAgg = filtersAgg.FilterWithName(otherFilterName, otherBoolQuery)

	result, err := createVisitStatsSearchService(ctx, client, boolQuery, aggName, filtersAgg)
	if err != nil {
		return nil, err
	}

	answerCategoryCounts := make(map[string]interface{})
	answerCategoryCounts[businessFilterName] = int64(0)
	answerCategoryCounts[chatFilterName] = int64(0)
	answerCategoryCounts[otherFilterName] = int64(0)

	if agg, found := result.Aggregations.Filters(aggName); found {
		businessBucket, ok := agg.NamedBuckets[businessFilterName]
		if !ok {
			return nil, esData.ErrESNameBucketNotFound
		}

		chatBucket, ok := agg.NamedBuckets[chatFilterName]
		if !ok {
			return nil, esData.ErrESNameBucketNotFound
		}

		otherBucket, ok := agg.NamedBuckets[otherFilterName]
		if !ok {
			return nil, esData.ErrESNameBucketNotFound
		}

		answerCategoryCounts[businessFilterName] = businessBucket.DocCount
		answerCategoryCounts[chatFilterName] = chatBucket.DocCount
		answerCategoryCounts[otherFilterName] = otherBucket.DocCount
	}

	return answerCategoryCounts, nil
}

func createVisitStatsDateHistogramAggregation(query dataV1.VisitStatsQuery) *elastic.DateHistogramAggregation {
	agg := services.CreateDateHistogramAggregation(query.CommonQuery, data.LogTimeFieldName)
	agg.Interval(query.AggInterval)
	return agg
}

func createVisitStatsSearchService(ctx context.Context, client *elastic.Client,
	query elastic.Query, aggName string, agg elastic.Aggregation) (*elastic.SearchResult, error) {
	index := fmt.Sprintf("%s-*", esData.ESRecordsIndex)
	return services.CreateSearchService(ctx, client, query, index, esData.ESRecordType, aggName, agg)
}

func doVisitStatsDateHistogramAggService(ctx context.Context, client *elastic.Client, query elastic.Query,
	aggName string, agg elastic.Aggregation) (map[string]interface{}, error) {
	index := fmt.Sprintf("%s-*", esData.ESRecordsIndex)
	return services.DoDateHistogramAggService(ctx, client, query,
		index, esData.ESRecordType, aggName, agg)
}

func doVisitStatsTermsAggService(ctx context.Context, client *elastic.Client, query elastic.Query,
	aggName string, agg elastic.Aggregation) (map[string]interface{}, error) {
	index := fmt.Sprintf("%s-*", esData.ESRecordsIndex)
	return services.DoTermsAggService(ctx, client, query, index, esData.ESRecordType, aggName, agg)
}
