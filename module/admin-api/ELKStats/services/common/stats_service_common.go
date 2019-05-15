package common

import (
	"fmt"

	"emotibot.com/emotigo/module/admin-api/ELKStats/data"
	"emotibot.com/emotigo/module/admin-api/ELKStats/services"
	elastic "gopkg.in/olivere/elastic.v6"
)

func CreateStatsBoolQuery(query data.CommonQuery) *elastic.BoolQuery {
	boolQuery := services.CreateBoolQuery(query)
	welcomeTagTermQuery := elastic.NewTermQuery("user_q.keyword", "welcome_tag")
	boolQuery = boolQuery.MustNot(welcomeTagTermQuery)
	return boolQuery
}

func CreateStatsTagExistsQuery(tag string) *elastic.ExistsQuery {
	field := fmt.Sprintf("custom_info.%s.keyword", tag)
	return elastic.NewExistsQuery(field)
}

func CreateStatsTagTermsAggregation(tag string) *elastic.TermsAggregation {
	field := fmt.Sprintf("custom_info.%s.keyword", tag)
	return elastic.NewTermsAggregation().Field(field).Size(data.ESTermAggShardSize)
}

func CreateTagCountsMap(query data.StatsQuery) (map[string]interface{}, error) {
	counts := make(map[string]interface{})
	availableTags := services.GetAvailableTags(query.AppID)

	aggTags, ok := availableTags[query.AggTagType]
	if !ok {
		return nil, data.ErrTagTypeNotFound
	}

	for _, aggTag := range aggTags {
		counts[aggTag.Name] = int64(0)
	}

	return counts, nil
}

// NormalizeTagCounts returns a new counts map with missing tags and without redundant tags
func NormalizeTagCounts(counts map[string]interface{}, appID string, tagType string) (map[string]interface{}, error) {
	tagCounts := make(map[string]interface{})
	availableTags := services.GetAvailableTags(appID)

	if aggTags, ok := availableTags[tagType]; ok {
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
