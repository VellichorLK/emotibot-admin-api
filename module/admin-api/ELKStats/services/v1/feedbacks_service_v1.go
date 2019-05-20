package v1

import (
	"fmt"

	"emotibot.com/emotigo/pkg/logger"

	"emotibot.com/emotigo/module/admin-api/ELKStats/data"
	dataCommon "emotibot.com/emotigo/module/admin-api/ELKStats/data/common"
	dataV1 "emotibot.com/emotigo/module/admin-api/ELKStats/data/v1"
	"emotibot.com/emotigo/module/admin-api/ELKStats/services"
	"emotibot.com/emotigo/module/admin-api/util/elasticsearch"
	esData "emotibot.com/emotigo/module/admin-api/util/elasticsearch/data"
	elastic "gopkg.in/olivere/elastic.v6"
)

func DailyAvgRating(query dataV1.FeedbacksQuery) (map[int64]*dataV1.RatingAverageInfo, error) {
	ctx, client := elasticsearch.GetClient()
	aggName := "avg_rating"
	countAggName := "count"
	boolQuery := services.CreateBoolQuery(query.CommonQuery)
	rangeQuery := services.CreateRangeQueryUnixTime(query.CommonQuery,
		data.SessionStartTimeFieldName)
	boolQuery.Filter(rangeQuery)

	tagTermQueries := createTagTermQueries(query)
	for _, tagTermQuery := range tagTermQueries {
		boolQuery.Filter(tagTermQuery)
	}

	avgAgg := elastic.NewDateHistogramAggregation().
		Field(data.SessionStartTimeFieldName).
		Interval("1d").
		SubAggregation(
			aggName,
			elastic.
				NewAvgAggregation().
				Field(data.SessionRatingFieldName)).
		SubAggregation(
			countAggName,
			elastic.
				NewTermsAggregation().
				Field(data.SessionRatingFieldName))

	index, indexType, err := getIndexAndType(query)
	if err != nil {
		return nil, err
	}

	result, err := services.CreateSearchService(ctx, client, boolQuery, index, indexType,
		aggName, avgAgg)
	if err != nil {
		return nil, err
	}

	ret := map[int64]*dataV1.RatingAverageInfo{}
	if agg, found := result.Aggregations.Terms(aggName); found {
		for _, aggBucket := range agg.Buckets {
			// key is UTC time in ms
			key, err := aggBucket.KeyNumber.Int64()
			if err != nil {
				return nil, err
			}
			key = key / 1000

			value := float64(0)
			if avgAgg, avgFound := aggBucket.Avg(aggName); avgFound {
				if avgAgg.Value != nil {
					value = *avgAgg.Value
				} else {
					logger.Trace.Println("value is nil")
				}
			}
			ratingCount, err := getRatingCount(aggBucket)
			if err != nil {
				return nil, err
			}
			ret[key] = &dataV1.RatingAverageInfo{
				Average: value,
				Count:   ratingCount,
			}
		}
	} else {
		logger.Trace.Println("agg not found")
	}
	return ret, nil
}

func getRatingCount(bucket *elastic.AggregationBucketKeyItem) (map[int64]int64, error) {
	ret := map[int64]int64{}
	if agg, found := bucket.Aggregations.Terms("count"); found {
		for _, aggBucket := range agg.Buckets {
			key, err := aggBucket.KeyNumber.Int64()
			if err != nil {
				return nil, err
			}
			ret[key] = aggBucket.DocCount
		}
	} else {
		logger.Trace.Println("count bucket not found")
	}
	return ret, nil
}

func AvgRating(query dataV1.FeedbacksQuery) (float64, error) {
	ctx, client := elasticsearch.GetClient()
	aggName := "avg_rating"
	boolQuery := services.CreateBoolQuery(query.CommonQuery)
	rangeQuery := services.CreateRangeQueryUnixTime(query.CommonQuery,
		data.SessionStartTimeFieldName)
	boolQuery.Filter(rangeQuery)

	tagTermQueries := createTagTermQueries(query)
	for _, tagTermQuery := range tagTermQueries {
		boolQuery.Filter(tagTermQuery)
	}

	avgAgg := elastic.NewAvgAggregation().Field(data.SessionRatingFieldName)

	index, indexType, err := getIndexAndType(query)
	if err != nil {
		return 0, err
	}

	result, err := services.CreateSearchService(ctx, client, boolQuery, index, indexType,
		aggName, avgAgg)
	if err != nil {
		return 0, err
	}

	if agg, found := result.Aggregations.Avg(aggName); found {
		if agg.Value != nil {
			return *agg.Value, nil
		}
	}

	return 0, nil
}

func Ratings(query dataV1.FeedbacksQuery, topN int) (dataV1.FeedbackRatings, error) {
	ctx, client := elasticsearch.GetClient()
	aggName := "ratings"
	boolQuery := services.CreateBoolQuery(query.CommonQuery)
	rangeQuery := services.CreateRangeQueryUnixTime(query.CommonQuery,
		data.SessionStartTimeFieldName)
	boolQuery.Filter(rangeQuery)

	tagTermQueries := createTagTermQueries(query)
	for _, tagTermQuery := range tagTermQueries {
		boolQuery.Filter(tagTermQuery)
	}

	ratingsAgg := elastic.NewTermsAggregation().Field(data.SessionRatingFieldName).
		Size(topN).ShardSize(10000).OrderByCount(false)

	index, indexType, err := getIndexAndType(query)
	if err != nil {
		return nil, err
	}

	result, err := services.DoTermsAggService(ctx, client, boolQuery, index, indexType,
		aggName, ratingsAgg)
	if err != nil {
		return nil, err
	}

	ratings := make(dataV1.FeedbackRatings, 0)
	for rating, count := range result {
		r := dataV1.FeedbackRating{
			Rating: rating,
			Count:  count.(int64),
		}

		ratings = append(ratings, r)
	}

	return ratings, nil
}

func Feedbacks(query dataV1.FeedbacksQuery, topN int) (dataV1.FeedbackCounts, error) {
	ctx, client := elasticsearch.GetClient()
	aggName := "feedbacks"
	boolQuery := services.CreateBoolQuery(query.CommonQuery)
	rangeQuery := services.CreateRangeQueryUnixTime(query.CommonQuery,
		data.SessionStartTimeFieldName)
	boolQuery.Filter(rangeQuery)

	tagTermQueries := createTagTermQueries(query)
	for _, tagTermQuery := range tagTermQueries {
		boolQuery.Filter(tagTermQuery)
	}

	feedbacksAgg := elastic.NewTermsAggregation().Field(data.SessionFeedbackFieldName).
		Size(topN).ShardSize(10000).OrderByCount(false)

	index, indexType, err := getIndexAndType(query)
	if err != nil {
		return nil, err
	}

	result, err := services.DoTermsAggService(ctx, client, boolQuery, index, indexType,
		aggName, feedbacksAgg)
	if err != nil {
		return nil, err
	}

	feedbacks := make(dataV1.FeedbackCounts, 0)
	for feedback, count := range result {
		f := dataV1.FeedbackCount{
			Feedback: feedback,
			Count:    count.(int64),
		}

		feedbacks = append(feedbacks, f)
	}

	return feedbacks, nil
}

func getIndexAndType(query dataV1.FeedbacksQuery) (index string, indexType string, err error) {
	switch query.Type {
	case dataCommon.FeedbacksStatsTypeSessions:
		index = fmt.Sprintf("%s-*", esData.ESSessionsIndex)
		indexType = esData.ESSessionsType
	case dataCommon.FeedbacksStatsTypeRecords:
		index = fmt.Sprintf("%s-*", esData.ESRecordsIndex)
		indexType = esData.ESRecordType
	case dataCommon.FeedbacksStatsTypeTERecords:
		index = fmt.Sprintf("%s-*", esData.ESTERecordsIndex)
		indexType = esData.ESTERecordsType
	default:
		err = data.ErrInvalidFeedbacksType
	}

	return
}

func createTagTermQueries(query dataV1.FeedbacksQuery) []*elastic.TermQuery {
	tagsTermQueries := make([]*elastic.TermQuery, 0)

	if query.Platform != "" {
		field := "custom_info.platform.keyword"
		platformTermQuery := elastic.NewTermQuery(field, query.Platform)
		tagsTermQueries = append(tagsTermQueries, platformTermQuery)
	}

	if query.Gender != "" {
		field := "custom_info.sex.keyword"
		genderTermQuery := elastic.NewTermQuery(field, query.Gender)
		tagsTermQueries = append(tagsTermQueries, genderTermQuery)
	}

	return tagsTermQueries
}
