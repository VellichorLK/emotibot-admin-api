package services

import (
	"context"
	"fmt"
	"time"

	"emotibot.com/emotigo/module/admin-api/ELKStats/dao"
	"emotibot.com/emotigo/module/admin-api/ELKStats/data"

	"github.com/olivere/elastic"
)

var tags map[string]map[string][]data.Tag
var timezone = getSystemTimezone()

func CreateBoolQuery(query data.CommonQuery) *elastic.BoolQuery {
	boolQuery := elastic.NewBoolQuery()

	if query.AppID != "" {
		appTermQuery := elastic.NewTermQuery("app_id", query.AppID)
		boolQuery = boolQuery.Filter(appTermQuery)
	}

	return boolQuery
}

func CreateRangeQuery(query data.CommonQuery, queryField string) *elastic.RangeQuery {
	return elastic.NewRangeQuery(queryField).
		Gte(query.StartTime.Format(data.ESTimeFormat)).
		Lte(query.EndTime.Format(data.ESTimeFormat)).
		Format("yyyy-MM-dd HH:mm:ss").
		TimeZone(timezone)
}

func CreateRangeQueryUnixTime(query data.CommonQuery, queryField string) *elastic.RangeQuery {
	return elastic.NewRangeQuery(queryField).
		Gte(query.StartTime.Unix()).
		Lte(query.EndTime.Unix()).
		Format("epoch_second")
}

func CreateDateHistogramAggregation(query data.CommonQuery, queryField string) *elastic.DateHistogramAggregation {
	return elastic.NewDateHistogramAggregation().
		Field(queryField).
		Format("yyyy-MM-dd HH:mm:ss").
		TimeZone(timezone).
		MinDocCount(0).
		ExtendedBounds(query.StartTime.Format(data.ESTimeFormat), query.EndTime.Format(data.ESTimeFormat))
}

func CreateSearchService(ctx context.Context, client *elastic.Client,
	query elastic.Query, index string, indexType string,
	aggName string, agg elastic.Aggregation) (*elastic.SearchResult, error) {
	if client == nil {
		return nil, data.ErrNotInit
	}

	return client.Search().
		Index(index).
		Type(indexType).
		Query(query).
		Aggregation(aggName, agg).
		Size(0).
		Do(ctx)
}

func ExtractCountsFromAggDateHistogramBuckets(result *elastic.SearchResult, aggName string) map[string]interface{} {
	counts := make(map[string]interface{})

	if agg, found := result.Aggregations.DateHistogram(aggName); found {
		for _, bucket := range agg.Buckets {
			counts[*bucket.KeyAsString] = bucket.DocCount
		}
	}

	return counts
}

func ExtractCountsFromAggTermsBuckets(result *elastic.SearchResult, aggName string) map[string]interface{} {
	counts := make(map[string]interface{})

	if agg, found := result.Aggregations.Terms(aggName); found {
		for _, bucket := range agg.Buckets {
			counts[bucket.Key.(string)] = bucket.DocCount
		}
	}

	return counts
}

func DoDateHistogramAggService(ctx context.Context, client *elastic.Client, query elastic.Query,
	index string, indexType string, aggName string, agg elastic.Aggregation) (map[string]interface{}, error) {
	result, err := CreateSearchService(ctx, client, query, index, indexType, aggName, agg)
	if err != nil {
		return nil, err
	}

	counts := ExtractCountsFromAggDateHistogramBuckets(result, aggName)
	return counts, nil
}

func DoTermsAggService(ctx context.Context, client *elastic.Client, query elastic.Query,
	index string, indexType string, aggName string, agg elastic.Aggregation) (map[string]interface{}, error) {
	result, err := CreateSearchService(ctx, client, query, index, indexType, aggName, agg)
	if err != nil {
		return nil, err
	}

	counts := ExtractCountsFromAggTermsBuckets(result, aggName)
	return counts, nil
}

func getSystemTimezone() string {
	_, offset := time.Now().Zone()

	hour := offset / (data.HourInSeconds)
	minute := ((offset % (data.HourInSeconds)) / data.MinuteInSeconds)
	if minute < 0 {
		minute = -minute
	}

	systemTimezone := fmt.Sprintf("%+03d:%02d", hour, minute)
	return systemTimezone
}

func createTimeRangeCountsMap(query data.CommonQuery,
	queryInterval string) (map[string]interface{}, error) {
	var currentTime = query.StartTime
	countsMap := make(map[string]interface{})

	for !currentTime.After(query.EndTime) {
		var mapTime time.Time

		switch queryInterval {
		case data.IntervalYear:
			mapTime = time.Date(currentTime.Year(), 0, 0, 0, 0, 0, 0, currentTime.Location())
		case data.IntervalMonth:
			mapTime = time.Date(currentTime.Year(), currentTime.Month(),
				0, 0, 0, 0, 0, currentTime.Location())
		case data.IntervalDay:
			mapTime = time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(),
				0, 0, 0, 0, currentTime.Location())
		case data.IntervalHour:
			mapTime = time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(),
				currentTime.Hour(), 0, 0, 0, currentTime.Location())
		case data.IntervalMinute:
			mapTime = time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(),
				currentTime.Hour(), currentTime.Minute(), 0, 0, currentTime.Location())
		case data.IntervalSecond:
			mapTime = currentTime
		default:
			return nil, data.ErrInvalidAggTimeInterval
		}

		mapKey := mapTime.Format(data.ESTimeFormat)
		countsMap[mapKey] = int64(0)

		switch queryInterval {
		case data.IntervalYear:
			currentTime = currentTime.AddDate(1, 0, 0)
		case data.IntervalMonth:
			currentTime = currentTime.AddDate(0, 1, 0)
		case data.IntervalDay:
			currentTime = currentTime.AddDate(0, 0, 1)
		case data.IntervalHour:
			currentTime = currentTime.Add(time.Hour)
		case data.IntervalMinute:
			currentTime = currentTime.Add(time.Minute)
		case data.IntervalSecond:
			currentTime = currentTime.Add(time.Second)
		default:
			return nil, data.ErrInvalidAggTimeInterval
		}
	}

	return countsMap, nil
}

func InitTags() (err error) {
	tags, err = dao.GetTags()
	if err != nil {
		return
	}

	return
}

func GetTagIDByName(appID string, tagType string, tagName string) (tagID string, found bool) {
	availableTags := GetAvailableTags(appID)

	_availableTags := availableTags[tagType]
	for _, tag := range _availableTags {
		if tag.Name == tagName {
			tagID = tag.Code
			found = true
			return
		}
	}

	found = false
	return
}

func GetTagNameByID(appID string, tagType string, tagID string) (tagName string, found bool) {
	availableTags := GetAvailableTags(appID)

	_availableTags := availableTags[tagType]
	for _, tag := range _availableTags {
		if tag.Code == tagID {
			tagName = tag.Name
			found = true
			return
		}
	}

	found = false
	return
}

func GetAvailableTags(appID string) map[string][]data.Tag {
	availableTags := make(map[string][]data.Tag, 0)

	_tags, ok := tags["system"]
	if ok {
		for _tagType, _tag := range _tags {
			availableTags[_tagType] = _tag
		}
	}

	_tags, ok = tags[appID]
	if ok {
		for _tagType, _tag := range _tags {
			availableTags[_tagType] = append(availableTags[_tagType], _tag...)
		}
	}

	return availableTags
}
