package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"emotibot.com/emotigo/module/admin-api/ELKStats/data"
	"github.com/olivere/elastic"
)

func VisitRecordsQuery(ctx context.Context, client *elastic.Client,
	query data.VisitRecordsQuery) (records []*data.VisitRecordsData, totalSize int64, limit int, err error) {
	boolQuery := createBoolQuery(query.CommonQuery)
	rangeQuery := createRangeQuery(query.CommonQuery, data.LogTimeFieldName)
	boolQuery = boolQuery.Filter(rangeQuery)

	if query.Question != "" {
		userQTermQuery := elastic.NewMultiMatchQuery(query.Question, "user_q", "answer.value")
		boolQuery = boolQuery.Filter(userQTermQuery)
	}

	if query.UserID != "" {
		userIDTermQuery := elastic.NewTermQuery("user_id", query.UserID)
		boolQuery = boolQuery.Filter(userIDTermQuery)
	}

	if query.Emotion != "" {
		emotionQuery := elastic.NewTermQuery("emotion", query.Emotion)
		boolQuery = boolQuery.Filter(emotionQuery)
	}

	if query.QType != "" {
		switch query.QType {
		case data.CategoryBusiness:
			businessTermsQuery := elastic.NewTermsQuery("module", "faq", "task_engine")
			boolQuery = boolQuery.Filter(businessTermsQuery)
		case data.CategoryChat:
			chatTermQuery := elastic.NewTermQuery("module", "chat")
			boolQuery = boolQuery.Filter(chatTermQuery)
		case data.CategoryOther:
			otherTermsQuery := elastic.NewTermsQuery("module", "faq", "task_engine", "chat")
			boolQuery = boolQuery.MustNot(otherTermsQuery)
		}
	}

	if query.Tags != nil && len(query.Tags) > 0 {
		tagsShouldBoolQuery := elastic.NewBoolQuery()
		tagsBoolQueries := make([]*elastic.BoolQuery, 0)
		tags := make([]data.VisitRecordsTag, 0)
		tagsBoolQueries, err = createTagsBoolQueries(tagsBoolQueries, query.Tags, 0, tags)
		if err != nil {
			return
		}

		for _, tagsBoolQuery := range tagsBoolQueries {
			tagsShouldBoolQuery = tagsShouldBoolQuery.Should(tagsBoolQuery)
		}

		boolQuery.Filter(tagsShouldBoolQuery)
	}

	from := (query.Page - 1) * query.PageLimit
	source := elastic.NewFetchSourceContext(true)
	source.Include(
		data.VisitRecordsMetricUserID,
		data.VisitRecordsMetricUserQ,
		data.VisitRecordsMetricScore,
		data.VisitRecordsMetricStdQ,
		"answer.value",
		data.VisitRecordsMetricLogTime,
		data.VisitRecordsMetricEmotion,
		"module",
	)

	result, err := client.Search().
		Index(data.ESRecordsIndex).
		Type(data.ESRecordType).
		Query(boolQuery).
		From(from).
		Size(query.PageLimit).
		FetchSourceContext(source).
		Sort(data.LogTimeFieldName, false).
		Do(ctx)
	if err != nil {
		return
	}

	totalSize = result.TotalHits()
	limit = query.PageLimit

	records = make([]*data.VisitRecordsData, 0)

	for _, hit := range result.Hits.Hits {
		hitResult := data.VisitRecordsHitResult{}
		jsonStr, _err := hit.Source.MarshalJSON()
		if _err != nil {
			err = _err
			return
		}

		err = json.Unmarshal(jsonStr, &hitResult)
		if err != nil {
			return
		}

		// Note: We have to convert log_time (UTC+0) to local time and reformat
		logTime, _err := time.Parse(data.LogTimeFormat, hitResult.LogTime)
		if _err != nil {
			err = _err
			return
		}
		logTime = logTime.Local()
		hitResult.LogTime = logTime.Format("2006-01-02 15:04:05")

		rawRecord := hitResult.VisitRecordsRawData
		answers := make([]string, 0)

		for _, answer := range rawRecord.Answer {
			answers = append(answers, answer.Value)
		}

		if hitResult.Module == "faq" || hitResult.Module == "task_engine" {
			rawRecord.QType = "业务类"
		} else if hitResult.Module == "chat" {
			rawRecord.QType = "聊天类"
		} else {
			rawRecord.QType = "其他类"
		}

		record := data.VisitRecordsData{
			UserID:  rawRecord.UserID,
			UserQ:   rawRecord.UserQ,
			Score:   rawRecord.Score,
			StdQ:    rawRecord.StdQ,
			Answer:  strings.Join(answers, ", "),
			LogTime: rawRecord.LogTime,
			Emotion: rawRecord.Emotion,
			QType:   rawRecord.QType,
		}

		records = append(records, &record)
	}

	return
}

// createTagsBoolQueries converts queryTags = [
//     {
//         Type: "platform"
//         Texts: ["android", "ios"]
//     },
//     {
//         Type: "sex"
//         Texts: ["男", "女"]	
//     }
// ]
//
// to the form:
// {
//   "bool": {
//     "filter": [
//       {
//         "term": {
//           "custom_info.platform": "android"
//         }
//       },
//       {
//         "term": {
//       	"custom_info.sex": "男"
//         }
//       }
//     ]
//   }
// },
// {
//   "bool": {
//     "filter": [
//       {
//         "term": {
//           "custom_info.platform": "android"
//         }
//       },
//       {
//         "term": {
//           "custom_info.sex": "女"
//         }
//       }
//     ]
//   }
// },
// {
//   "bool": {
//     "filter": [
//       {
//         "term": {
//           "custom_info.platform": "ios"
//         }
//       },
//       {
//         "term": {
//           "custom_info.sex": "男"
//         }
//       }
//     ]
//   }
// },
// {
//   "bool": {
//     "filter": [
//       {
//         "term": {
//           "custom_info.platform": "ios"
//         }
//       },
//       {
//         "term": {
//           "custom_info.sex": "女"
//         }
//       }
//     ]
//   }
// }
func createTagsBoolQueries(tagsBoolQueries []*elastic.BoolQuery,
	queryTags []data.VisitRecordsQueryTag, queryTagsIndex int, tags []data.VisitRecordsTag) ([]*elastic.BoolQuery, error) {
	if queryTagsIndex != len(queryTags)-1 {
		for _, tagText := range queryTags[queryTagsIndex].Texts {
			tag := data.VisitRecordsTag{
				Type: queryTags[queryTagsIndex].Type,
				Text: tagText,
			}
			_tags := make([]data.VisitRecordsTag, len(tags)+1)
			copy(_tags, tags)
			_tags[len(tags)] = tag

			boolQueries, err := createTagsBoolQueries(tagsBoolQueries, queryTags, queryTagsIndex+1, _tags)
			if err != nil {
				return nil, err
			}

			tagsBoolQueries = append(tagsBoolQueries, boolQueries...)
		}
	} else {
		boolQueries := make([]*elastic.BoolQuery, 0)

		for _, text := range queryTags[queryTagsIndex].Texts {
			boolQuery := elastic.NewBoolQuery()

			for _, tag := range tags {
				field := fmt.Sprintf("custom_info.%s", tag.Type)
				termQuery := elastic.NewTermQuery(field, tag.Text)
				boolQuery = boolQuery.Filter(termQuery)
			}

			field := fmt.Sprintf("custom_info.%s", queryTags[queryTagsIndex].Type)
			termQuery := elastic.NewTermQuery(field, text)
			boolQuery = boolQuery.Filter(termQuery)

			boolQueries = append(boolQueries, boolQuery)
		}

		return boolQueries, nil
	}

	return tagsBoolQueries, nil
}
