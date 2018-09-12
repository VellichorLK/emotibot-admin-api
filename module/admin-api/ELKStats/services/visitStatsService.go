package services

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"time"

	"emotibot.com/emotigo/module/admin-api/ELKStats/data"
	"github.com/olivere/elastic"
)

const (
	HourInSeconds   = 60 * 60
	MinuteInSeconds = 60
)

const ActiveUsersThreshold = 30

func ConversationCounts(ctx context.Context, client *elastic.Client,
	query data.VisitStatsQuery) (map[string]interface{}, error) {
	aggName := "conversations"
	boolQuery := elastic.NewBoolQuery()
	rangeQuery := createRangeQuery(query.CommonQuery, data.SessionEndTimeFieldName)
	boolQuery = boolQuery.Filter(rangeQuery)

	groupBySessionsTermAgg := elastic.NewTermsAggregation().Field("session_id").Size(data.ESTermAggSize)

	switch query.AggBy {
	case data.AggByTime:
		dateHistogramAgg := createDateHistogramAggregation(query.CommonQuery, data.SessionEndTimeFieldName).
			Interval(query.AggInterval)
		dateHistogramAgg.SubAggregation(aggName, groupBySessionsTermAgg)

		index := fmt.Sprintf("%s-%s-*", data.ESSessionsIndex, query.AppID)
		result, err := createSearchService(ctx, client, boolQuery, index, data.ESSessionsType, aggName, dateHistogramAgg)
		if err != nil {
			return nil, err
		}

		counts := extractCountsFromAggDateHistogramBuckets(result, aggName)
		return counts, nil
	case data.AggByTag:
		tagExistsQuery := createVisitStatsTagExistsQuery(query.AggTagType)
		boolQuery = boolQuery.Filter(tagExistsQuery)
		tagTermAgg := createVisitStatsTagTermsAggregation(query.AggTagType)
		tagTermAgg.SubAggregation(aggName, groupBySessionsTermAgg)

		index := fmt.Sprintf("%s-%s-*", data.ESSessionsIndex, query.AppID)
		result, err := createSearchService(ctx, client, boolQuery, index, data.ESSessionsType, aggName, tagTermAgg)
		if err != nil {
			return nil, err
		}

		counts := extractCountsFromAggTermsBuckets(result, aggName)
		return normalizeTagCounts(counts, query.AggTagType)
	default:
		return nil, data.ErrInvalidAggType
	}
}

func UniqueUserCounts(ctx context.Context, client *elastic.Client,
	query data.VisitStatsQuery) (map[string]interface{}, error) {
	aggName := "unique_users"
	boolQuery := createVisitStatsBoolQuery(query.CommonQuery)
	rangeQuery := createRangeQuery(query.CommonQuery, data.LogTimeFieldName)
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
		tagExistsQuery := createVisitStatsTagExistsQuery(query.AggTagType)
		boolQuery = boolQuery.Filter(tagExistsQuery)
		tagTermAgg := createVisitStatsTagTermsAggregation(query.AggTagType)
		tagTermAgg.SubAggregation(uniqueUserCountAggName, uniqueUserCardinalityAgg)

		_agg = tagTermAgg
	default:
		return nil, data.ErrInvalidAggType
	}

	result, err := createVisitStatsSearchService(ctx, client, query.AppID, boolQuery, aggName, _agg)
	if err != nil {
		return nil, err
	}

	uniqueUserCounts := make(map[string]interface{})

	if agg, found := result.Aggregations.Terms(aggName); found {
		for _, bucket := range agg.Buckets {
			uniqueUserCount, found := bucket.Cardinality(uniqueUserCountAggName)
			if !found {
				return nil, data.ErrESCardinalityNotFound
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

func ActiveUserCounts(ctx context.Context, client *elastic.Client,
	query data.VisitStatsQuery) (map[string]interface{}, error) {
	aggName := "active_users"
	boolQuery := createVisitStatsBoolQuery(query.CommonQuery)
	rangeQuery := createRangeQuery(query.CommonQuery, data.LogTimeFieldName)
	boolQuery = boolQuery.Filter(rangeQuery)

	groupByUsersTermAggName := "group_by_users"
	groupByUsersTermAgg := elastic.NewTermsAggregation()
	groupByUsersTermAgg.Field("user_id").Size(data.ESTermAggSize)

	activeUsersFilterAggName := "active_user_filter"
	activeUsersFilterAgg := elastic.NewBucketSelectorAggregation()
	activeUsersFilterAgg.AddBucketsPath("DocCount", "_count")
	activeUsersThresholdScript := fmt.Sprintf("params.DocCount > %d", ActiveUsersThreshold)
	activeUsersFilterScript := elastic.NewScript(activeUsersThresholdScript)
	activeUsersFilterAgg.Script(activeUsersFilterScript)

	groupByUsersTermAgg.SubAggregation(activeUsersFilterAggName, activeUsersFilterAgg)

	var _agg elastic.Aggregation

	switch query.AggBy {
	case data.AggByTime:
		dateHistogramAgg := createVisitStatsDateHistogramAggregation(query)
		dateHistogramAgg.SubAggregation(groupByUsersTermAggName, groupByUsersTermAgg)

		_agg = dateHistogramAgg
	case data.AggByTag:
		boolQuery = boolQuery.Filter(rangeQuery)
		tagExistsQuery := createVisitStatsTagExistsQuery(query.AggTagType)
		boolQuery = boolQuery.Filter(tagExistsQuery)
		tagTermAgg := createVisitStatsTagTermsAggregation(query.AggTagType)
		tagTermAgg.SubAggregation(groupByUsersTermAggName, groupByUsersTermAgg)

		_agg = tagTermAgg
	default:
		return nil, data.ErrInvalidAggType
	}

	result, err := createVisitStatsSearchService(ctx, client, query.AppID, boolQuery, aggName, _agg)
	if err != nil {
		return nil, err
	}

	activeUserCounts := make(map[string]interface{})

	if agg, found := result.Aggregations.Terms(aggName); found {
		for _, bucket := range agg.Buckets {
			var bucketKey string

			switch query.AggBy {
			case data.AggByTime:
				bucketKey = *bucket.KeyAsString
			case data.AggByTag:
				bucketKey = bucket.Key.(string)
			default:
				return nil, data.ErrInvalidAggType
			}

			groupByUsersTermAgg, found := bucket.Terms(groupByUsersTermAggName)
			if !found {
				return nil, data.ErrESTermsNotFound
			}
			activeUserCounts[bucketKey] = int64(len(groupByUsersTermAgg.Buckets))
		}
	}

	return activeUserCounts, nil
}

func NewUserCounts(ctx context.Context, client *elastic.Client,
	query data.VisitStatsQuery) (map[string]interface{}, error) {
	aggName := "new_users"
	boolQuery := createVisitStatsBoolQuery(query.CommonQuery)

	termAgg := elastic.NewTermsAggregation().Field("user_id").Size(data.ESTermAggSize)
	userFirstChatAggName := "user_first_chat"
	userFirstChatAgg := elastic.NewTopHitsAggregation()
	userFirstChatAgg.Size(1).Sort(data.LogTimeFieldName, true)

	extractNewUserCountsByTime := func(result *elastic.SearchResult) (map[string]interface{}, error) {
		newUserCounts, err := createTimeRangeCountsMap(query.CommonQuery, query.AggInterval)
		if err != nil {
			return nil, err
		}

		if agg, found := result.Aggregations.Terms(aggName); found {
			for _, bucket := range agg.Buckets {
				userFirstChat, found := bucket.TopHits(userFirstChatAggName)
				if !found {
					return nil, data.ErrESTopHitNotFound
				}

				record := data.Record{}
				jsonStr, err := userFirstChat.Hits.Hits[0].Source.MarshalJSON()
				if err != nil {
					return nil, err
				}

				err = json.Unmarshal(jsonStr, &record)
				if err != nil {
					return nil, err
				}

				// Note: We have to convert log_time (UTC+0) to local time
				logTime, err := time.Parse(data.LogTimeFormat, record.LogTime)
				if err != nil {
					return nil, err
				}
				logTime = logTime.Local()

				if logTime.Before(query.StartTime) || logTime.After(query.EndTime) {
					continue
				}

				// Find out corresponding countsMap key of the logTime
				var countsMapTime time.Time

				switch query.AggInterval {
				case data.IntervalYear:
					countsMapTime = time.Date(logTime.Year(), 0, 0, 0, 0, 0, 0, logTime.Location())
				case data.IntervalMonth:
					countsMapTime = time.Date(logTime.Year(), logTime.Month(), 0, 0, 0, 0, 0,
						logTime.Location())
				case data.IntervalDay:
					countsMapTime = time.Date(logTime.Year(), logTime.Month(), logTime.Day(),
						0, 0, 0, 0, logTime.Location())
				case data.IntervalHour:
					countsMapTime = time.Date(logTime.Year(), logTime.Month(), logTime.Day(),
						logTime.Hour(), 0, 0, 0, logTime.Location())
				case data.IntervalMinute:
					countsMapTime = time.Date(logTime.Year(), logTime.Month(), logTime.Day(),
						logTime.Hour(), logTime.Minute(), 0, 0, logTime.Location())
				case data.IntervalSecond:
					countsMapTime = logTime
				default:
					return nil, data.ErrInvalidAggTimeInterval
				}

				// Increase the counter
				countsMapKey := countsMapTime.Format(data.ESTimeFormat)
				if newUserCount, ok := newUserCounts[countsMapKey]; ok {
					newUserCounts[countsMapKey] = newUserCount.(int64) + 1
				}
			}
		}

		return newUserCounts, nil
	}

	extractNewUserCountsByTag := func(result *elastic.SearchResult,
		tagTermAggName string) (map[string]interface{}, error) {
		newUserCounts, err := createTagCountsMap(query)
		if err != nil {
			return nil, err
		}

		if agg, found := result.Aggregations.Terms(aggName); found {
			for _, bucket := range agg.Buckets {
				if tagAgg, found := bucket.Aggregations.Terms(tagTermAggName); found {
					for _, tagBucket := range tagAgg.Buckets {
						userFirstChat, found := tagBucket.TopHits(userFirstChatAggName)
						if !found {
							return nil, data.ErrESTopHitNotFound
						}

						record := data.Record{}
						jsonStr, err := userFirstChat.Hits.Hits[0].Source.MarshalJSON()
						if err != nil {
							return nil, err
						}

						err = json.Unmarshal(jsonStr, &record)
						if err != nil {
							return nil, err
						}

						logTime, err := time.Parse(data.LogTimeFormat, record.LogTime)
						if err != nil {
							return nil, err
						}

						// Increase the counter if user's first chat of the tag is within query time range
						if !logTime.Before(query.StartTime) && !logTime.After(query.EndTime) {
							// TagBucket.Key is the returned tag name
							if _, ok := newUserCounts[tagBucket.Key.(string)]; ok {
								counter := newUserCounts[tagBucket.Key.(string)].(int64)
								newUserCounts[tagBucket.Key.(string)] = counter + 1
							}
						}
					}
				}
			}
		}

		return newUserCounts, nil
	}

	switch query.AggBy {
	case data.AggByTime:
		termAgg.SubAggregation(userFirstChatAggName, userFirstChatAgg)

		// Result will be the documents of users' first chat times (i.e. MIN(log_time))
		result, err := createVisitStatsSearchService(ctx, client, query.AppID, boolQuery, aggName, termAgg)
		if err != nil {
			return nil, err
		}

		return extractNewUserCountsByTime(result)
	case data.AggByTag:
		boolQuery := elastic.NewBoolQuery()
		tagExistsQuery := createVisitStatsTagExistsQuery(query.AggTagType)
		boolQuery = boolQuery.Filter(tagExistsQuery)
		tagTermAggName := "group_by_tag"
		tagTermAgg := createVisitStatsTagTermsAggregation(query.AggTagType)
		tagTermAgg.SubAggregation(userFirstChatAggName, userFirstChatAgg)
		termAgg.SubAggregation(tagTermAggName, tagTermAgg)

		// Result will be the documents of users' first chat times (i.e. MIN(log_time)) group by each tag
		result, err := createVisitStatsSearchService(ctx, client, query.AppID, boolQuery, aggName, termAgg)
		if err != nil {
			return nil, err
		}

		return extractNewUserCountsByTag(result, tagTermAggName)
	default:
		return nil, data.ErrInvalidAggType
	}
}

func TotalAskCounts(ctx context.Context, client *elastic.Client,
	query data.VisitStatsQuery) (map[string]interface{}, error) {
	aggName := "total_asks"
	boolQuery := createVisitStatsBoolQuery(query.CommonQuery)
	rangeQuery := createRangeQuery(query.CommonQuery, data.LogTimeFieldName)
	boolQuery = boolQuery.Filter(rangeQuery)

	switch query.AggBy {
	case data.AggByTime:
		dateHistogramAgg := createVisitStatsDateHistogramAggregation(query)
		return doVisitStatsDateHistogramAggService(ctx, client, query.AppID, boolQuery, aggName, dateHistogramAgg)
	case data.AggByTag:
		tagExistsQuery := createVisitStatsTagExistsQuery(query.AggTagType)
		boolQuery = boolQuery.Filter(tagExistsQuery)
		tagTermAgg := createVisitStatsTagTermsAggregation(query.AggTagType)

		counts, err := doVisitStatsTermsAggService(ctx, client, query.AppID, boolQuery, aggName, tagTermAgg)
		if err != nil {
			return nil, err
		}

		return normalizeTagCounts(counts, query.AggTagType)
	default:
		return nil, data.ErrInvalidAggType
	}
}

func NormalResponseCounts(ctx context.Context, client *elastic.Client,
	query data.VisitStatsQuery) (map[string]interface{}, error) {
	aggName := "normal_responses"
	boolQuery := createVisitStatsBoolQuery(query.CommonQuery)
	termsQuery := elastic.NewTermsQuery("module", "faq", "task_engine")
	rangeQuery := createRangeQuery(query.CommonQuery, data.LogTimeFieldName)
	boolQuery = boolQuery.Filter(termsQuery)
	boolQuery = boolQuery.Filter(rangeQuery)

	switch query.AggBy {
	case data.AggByTime:
		dateHistogramAgg := createVisitStatsDateHistogramAggregation(query)
		return doVisitStatsDateHistogramAggService(ctx, client, query.AppID, boolQuery, aggName, dateHistogramAgg)
	case data.AggByTag:
		tagExistsQuery := createVisitStatsTagExistsQuery(query.AggTagType)
		boolQuery = boolQuery.Filter(tagExistsQuery)
		tagTermAgg := createVisitStatsTagTermsAggregation(query.AggTagType)

		counts, err := doVisitStatsTermsAggService(ctx, client, query.AppID, boolQuery, aggName, tagTermAgg)
		if err != nil {
			return nil, err
		}

		return normalizeTagCounts(counts, query.AggTagType)
	default:
		return nil, data.ErrInvalidAggType
	}
}

func ChatCounts(ctx context.Context, client *elastic.Client,
	query data.VisitStatsQuery) (map[string]interface{}, error) {
	aggName := "chats"
	boolQuery := createVisitStatsBoolQuery(query.CommonQuery)
	termQuery := elastic.NewTermQuery("module", "chat")
	rangeQuery := createRangeQuery(query.CommonQuery, data.LogTimeFieldName)
	boolQuery = boolQuery.Filter(termQuery)
	boolQuery = boolQuery.Filter(rangeQuery)

	switch query.AggBy {
	case data.AggByTime:
		dateHistogramAgg := createVisitStatsDateHistogramAggregation(query)
		return doVisitStatsDateHistogramAggService(ctx, client, query.AppID, boolQuery, aggName, dateHistogramAgg)
	case data.AggByTag:
		tagExistsQuery := createVisitStatsTagExistsQuery(query.AggTagType)
		boolQuery = boolQuery.Filter(tagExistsQuery)
		tagTermAgg := createVisitStatsTagTermsAggregation(query.AggTagType)

		counts, err := doVisitStatsTermsAggService(ctx, client, query.AppID, boolQuery, aggName, tagTermAgg)
		if err != nil {
			return nil, err
		}

		return normalizeTagCounts(counts, query.AggTagType)
	default:
		return nil, data.ErrInvalidAggType
	}
}

func OtherCounts(ctx context.Context, client *elastic.Client,
	query data.VisitStatsQuery) (map[string]interface{}, error) {
	aggName := "others"
	boolQuery := createVisitStatsBoolQuery(query.CommonQuery)
	termsQuery := elastic.NewTermsQuery("module", "faq", "task_engine", "chat", "backfill")
	rangeQuery := createRangeQuery(query.CommonQuery, data.LogTimeFieldName)
	boolQuery = boolQuery.MustNot(termsQuery)
	boolQuery = boolQuery.Filter(rangeQuery)

	switch query.AggBy {
	case data.AggByTime:
		dateHistogramAgg := createVisitStatsDateHistogramAggregation(query)
		return doVisitStatsDateHistogramAggService(ctx, client, query.AppID, boolQuery, aggName, dateHistogramAgg)
	case data.AggByTag:
		tagExistsQuery := createVisitStatsTagExistsQuery(query.AggTagType)
		boolQuery = boolQuery.Filter(tagExistsQuery)
		tagTermAgg := createVisitStatsTagTermsAggregation(query.AggTagType)

		counts, err := doVisitStatsTermsAggService(ctx, client, query.AppID, boolQuery, aggName, tagTermAgg)
		if err != nil {
			return nil, err
		}

		return normalizeTagCounts(counts, query.AggTagType)
	default:
		return nil, data.ErrInvalidAggType
	}
}

func UnknownQnACounts(ctx context.Context, client *elastic.Client,
	query data.VisitStatsQuery) (map[string]interface{}, error) {
	aggName := "unknown_qna"
	boolQuery := createVisitStatsBoolQuery(query.CommonQuery)
	termQuery := elastic.NewTermQuery("module", "backfill")
	rangeQuery := createRangeQuery(query.CommonQuery, data.LogTimeFieldName)
	boolQuery = boolQuery.Filter(termQuery)
	boolQuery = boolQuery.Filter(rangeQuery)

	switch query.AggBy {
	case data.AggByTime:
		dateHistogramAgg := createVisitStatsDateHistogramAggregation(query)
		return doVisitStatsDateHistogramAggService(ctx, client, query.AppID, boolQuery, aggName, dateHistogramAgg)
	case data.AggByTag:
		tagExistsQuery := createVisitStatsTagExistsQuery(query.AggTagType)
		boolQuery = boolQuery.Filter(tagExistsQuery)
		tagTermAgg := createVisitStatsTagTermsAggregation(query.AggTagType)

		counts, err := doVisitStatsTermsAggService(ctx, client, query.AppID, boolQuery, aggName, tagTermAgg)
		if err != nil {
			return nil, err
		}

		return normalizeTagCounts(counts, query.AggTagType)
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

func TopQuestions(ctx context.Context, client *elastic.Client,
	query data.VisitStatsQuery, topN int) ([]data.Question, error) {
	aggName := "top_user_q"
	boolQuery := createVisitStatsBoolQuery(query.CommonQuery)
	rangeQuery := createRangeQuery(query.CommonQuery, data.LogTimeFieldName)
	boolQuery = boolQuery.Filter(rangeQuery)

	topQTermAgg := elastic.NewTermsAggregation().Field("user_q.keyword").Size(topN).ShardSize(data.ESTermAggSize)
	result, err := doVisitStatsTermsAggService(ctx, client, query.AppID, boolQuery, aggName, topQTermAgg)
	if err != nil {
		return nil, err
	}

	questions := make(data.Questions, 0)
	for question, count := range result {
		q := data.Question{
			Question: question,
			Count:    count.(int64),
		}

		questions = append(questions, q)
	}

	// Sort by counts
	sort.Sort(sort.Reverse(questions))
	return questions, nil
}

func TopUnmatchQuestions(ctx context.Context, client *elastic.Client,
	query data.VisitStatsQuery, topN int) ([]*data.UnmatchQuestion, error) {
	aggName := "top_unmatch_q"
	boolQuery := createVisitStatsBoolQuery(query.CommonQuery)
	termQuery := elastic.NewTermQuery("module", "backfill")
	rangeQuery := createRangeQuery(query.CommonQuery, data.LogTimeFieldName)
	boolQuery = boolQuery.Filter(termQuery, rangeQuery)

	maxLogTimeAggName := "max_log_time"
	minLogTimeAggName := "min_log_time"
	topUnmatchQTermAgg := elastic.NewTermsAggregation().Field("user_q.keyword").Size(topN).ShardSize(data.ESTermAggShardSize)
	maxLogTimeAgg := elastic.NewMaxAggregation().Field(data.LogTimeFieldName).Format("yyyy-MM-dd HH:mm:ss")
	minLogTimeAgg := elastic.NewMinAggregation().Field(data.LogTimeFieldName).Format("yyyy-MM-dd HH:mm:ss")
	topUnmatchQTermAgg.SubAggregation(maxLogTimeAggName, maxLogTimeAgg)
	topUnmatchQTermAgg.SubAggregation(minLogTimeAggName, minLogTimeAgg)

	result, err := createVisitStatsSearchService(ctx, client, query.AppID, boolQuery, aggName, topUnmatchQTermAgg)
	if err != nil {
		return nil, err
	}

	unmatchQuestions := make([]*data.UnmatchQuestion, 0)

	if agg, found := result.Aggregations.Terms(aggName); found {
		for _, bucket := range agg.Buckets {
			maxLogTimeBucket, found := bucket.MaxBucket(maxLogTimeAggName)
			if !found {
				return nil, data.ErrESMaxBucketNotFound
			}

			// Note: We have to convert maximum log_time (UTC+0) to local time
			maxLogTime, err := time.Parse(data.ESTimeFormat, maxLogTimeBucket.ValueAsString)
			if err != nil {
				return nil, err
			}
			maxLogTime = maxLogTime.Local()

			minLogTimeBucket, found := bucket.MinBucket(minLogTimeAggName)
			if !found {
				return nil, data.ErrESMinBucketNotFound
			}

			// Note: We have to convert minimum log_time (UTC+0) to local time
			minLogTime, err := time.Parse(data.ESTimeFormat, minLogTimeBucket.ValueAsString)
			if err != nil {
				return nil, err
			}
			minLogTime = minLogTime.Local()

			unmatchQuestion := data.UnmatchQuestion{
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

func AnswerCategoryCounts(ctx context.Context, client *elastic.Client,
	query data.VisitStatsQuery) (map[string]interface{}, error) {
	aggName := "answer_categories"
	boolQuery := createVisitStatsBoolQuery(query.CommonQuery)
	rangeQuery := createRangeQuery(query.CommonQuery, data.LogTimeFieldName)
	boolQuery = boolQuery.Filter(rangeQuery)

	filtersAgg := elastic.NewFiltersAggregation()

	businessFilterName := data.CategoryBusiness
	businessTermsQuery := elastic.NewTermsQuery("module", "faq", "task_engine")

	chatFilterName := data.CategoryChat
	chatTermQuery := elastic.NewTermQuery("module", "chat")

	otherFilterName := data.CategoryOther
	otherBoolQuery := elastic.NewBoolQuery()
	otherNotTermsQuery := elastic.NewTermsQuery("module", "faq", "task_engine", "chat")
	otherBoolQuery = otherBoolQuery.MustNot(otherNotTermsQuery)

	filtersAgg = filtersAgg.FilterWithName(businessFilterName, businessTermsQuery)
	filtersAgg = filtersAgg.FilterWithName(chatFilterName, chatTermQuery)
	filtersAgg = filtersAgg.FilterWithName(otherFilterName, otherBoolQuery)

	result, err := createVisitStatsSearchService(ctx, client, query.AppID, boolQuery, aggName, filtersAgg)
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
			return nil, data.ErrESNameBucketNotFound
		}

		chatBucket, ok := agg.NamedBuckets[chatFilterName]
		if !ok {
			return nil, data.ErrESNameBucketNotFound
		}

		otherBucket, ok := agg.NamedBuckets[otherFilterName]
		if !ok {
			return nil, data.ErrESNameBucketNotFound
		}

		answerCategoryCounts[businessFilterName] = businessBucket.DocCount
		answerCategoryCounts[chatFilterName] = chatBucket.DocCount
		answerCategoryCounts[otherFilterName] = otherBucket.DocCount
	}

	return answerCategoryCounts, nil
}

func createVisitStatsBoolQuery(query data.CommonQuery) *elastic.BoolQuery {
	boolQuery := elastic.NewBoolQuery()
	welcomeTagTermQuery := elastic.NewTermQuery("user_q.keyword", "welcome_tag")
	boolQuery = boolQuery.MustNot(welcomeTagTermQuery)
	return boolQuery
}

func createVisitStatsTagExistsQuery(tag string) *elastic.ExistsQuery {
	field := fmt.Sprintf("custom_info.%s.keyword", tag)
	return elastic.NewExistsQuery(field)
}

func createVisitStatsDateHistogramAggregation(query data.VisitStatsQuery) *elastic.DateHistogramAggregation {
	agg := createDateHistogramAggregation(query.CommonQuery, data.LogTimeFieldName)
	agg.Interval(query.AggInterval)
	return agg
}

func createVisitStatsTagTermsAggregation(tag string) *elastic.TermsAggregation {
	field := fmt.Sprintf("custom_info.%s.keyword", tag)
	return elastic.NewTermsAggregation().Field(field).Size(data.ESTermAggSize)
}

func createVisitStatsSearchService(ctx context.Context, client *elastic.Client, appID string,
	query elastic.Query, aggName string, agg elastic.Aggregation) (*elastic.SearchResult, error) {
	index := fmt.Sprintf("%s-%s-*", data.ESRecordsIndex, appID)
	return createSearchService(ctx, client, query, index, data.ESRecordType, aggName, agg)
}

func doVisitStatsDateHistogramAggService(ctx context.Context, client *elastic.Client, appID string, query elastic.Query,
	aggName string, agg elastic.Aggregation) (map[string]interface{}, error) {
	index := fmt.Sprintf("%s-%s-*", data.ESRecordsIndex, appID)
	result, err := createSearchService(ctx, client, query, index, data.ESRecordType, aggName, agg)
	if err != nil {
		return nil, err
	}

	counts := extractCountsFromAggDateHistogramBuckets(result, aggName)
	return counts, nil
}

func doVisitStatsTermsAggService(ctx context.Context, client *elastic.Client, appID string, query elastic.Query,
	aggName string, agg elastic.Aggregation) (map[string]interface{}, error) {
	index := fmt.Sprintf("%s-%s-*", data.ESRecordsIndex, appID)
	return doTermsAggService(ctx, client, query, index, data.ESRecordType, aggName, agg)
}

func createTagCountsMap(query data.VisitStatsQuery) (map[string]interface{}, error) {
	counts := make(map[string]interface{})

	aggTags, ok := tags[query.AggTagType]
	if !ok {
		return nil, data.ErrTagTypeNotFound
	}

	for _, aggTag := range aggTags {
		counts[aggTag.Name] = int64(0)
	}

	return counts, nil
}

// normalizeTagCounts returns a new counts map with missing tags and without redundant tags
func normalizeTagCounts(counts map[string]interface{}, tagType string) (map[string]interface{}, error) {
	tagCounts := make(map[string]interface{})

	if aggTags, ok := tags[tagType]; ok {
		for _, aggTag := range aggTags {
			if count, ok := counts[aggTag.Name]; ok {
				tagCounts[aggTag.Name] = count
			} else {
				tagCounts[aggTag.Name] = int64(0)
			}
		}
	}

	return tagCounts, nil
}
