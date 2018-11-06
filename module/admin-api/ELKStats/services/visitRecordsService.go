package services

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"emotibot.com/emotigo/module/admin-api/ELKStats/data"
	"emotibot.com/emotigo/module/admin-api/util/elasticsearch"
	"emotibot.com/emotigo/pkg/logger"
	"github.com/olivere/elastic"
)

//RecordResult is the result fetched from Datastore,
//which will include the Hits raw data and aggregation static data
//Aggs always contain total_size, and may contain stats info like total_marked_size ...
type RecordResult struct {
	Hits []*data.VisitRecordsData
	Aggs map[string]interface{}
}

type ElasticSearchCommand func(*elastic.SearchService) *elastic.SearchService

func AggregateFilterMarkedRecord(ss *elastic.SearchService) *elastic.SearchService {
	markFilter := elastic.NewFilterAggregation().Filter(elastic.NewTermQuery("isMarked", true))
	ss = ss.Aggregation("isMarked", markFilter)
	return ss
}

func AggregateFilterIgnoredRecord(ss *elastic.SearchService) *elastic.SearchService {
	ignoreFilter := elastic.NewFilterAggregation().Filter(elastic.NewTermQuery("isIgnored", true))
	ss = ss.Aggregation("isIgnored", ignoreFilter)
	return ss
}

var platformDict = map[string]string{
	"wechat": "微信",
	"app":    "App",
	"web":    "Web",
	"ios":    "ios",
}

var genderDict = map[string]string{
	"0": "男",
	"1": "女",
}
var emotionDict = map[string]string{
	"angry":         "愤怒",
	"not_satisfied": "不满",
	"satisfied":     "满意",
	"neutral":       "中性",
}

//NewBoolQueryWithRecordQuery create a *elastic.BoolQuery with the Record condition.
func NewBoolQueryWithRecordQuery(query data.RecordQuery) *elastic.BoolQuery {
	boolQuery := elastic.NewBoolQuery()
	if query.Records != nil {
		//Executing a Terms Query request with a lot of terms can be quite slow, as each additional term demands extra processing and memory.
		//To safeguard against this, the maximum number of terms that can be used in a Terms Query both directly or through lookup has been limited to 65536.
		boolQuery.Filter(elastic.NewTermsQuery("unique_id", query.Records...))
	}
	if query.StartTime != nil && query.EndTime != nil {
		var rq = elastic.NewRangeQuery("log_time")
		rq.Gte(time.Unix(*query.StartTime, 0).In(time.UTC).Format(data.ESTimeFormat))
		rq.Lte(time.Unix(*query.EndTime, 0).In(time.UTC).Format(data.ESTimeFormat))
		rq.Format("yyyy-MM-dd HH:mm:ss")
		//Since ElasticSearch will transfer the timezone for us, it is no need to transfer the timezone by ourself
		rq.TimeZone("+00:00")
		boolQuery = boolQuery.Filter(rq)
	}

	if query.Keyword != nil {
		userQMatchQuery := elastic.NewMatchQuery("user_q", *query.Keyword)
		answerMatchQuery := elastic.NewMatchQuery("answer.value", *query.Keyword)
		answerNestedQuery := elastic.NewNestedQuery("answer", answerMatchQuery)
		keywordBoolQuery := elastic.NewBoolQuery().Should(userQMatchQuery, answerNestedQuery)
		boolQuery = boolQuery.Filter(keywordBoolQuery)
	}

	if query.UserID != nil {
		userIDTermQuery := elastic.NewTermQuery("user_id", *query.UserID)
		boolQuery = boolQuery.Filter(userIDTermQuery)
	}

	if query.Emotions != nil {
		emotions := make([]interface{}, 0)
		for _, emotion := range query.Emotions {
			e, ok := emotionDict[emotion]
			if ok {
				emotions = append(emotions, e)
			}
		}
		if len(emotions) > 0 {
			emotionQuery := elastic.NewTermsQuery("emotion", emotions...)
			boolQuery = boolQuery.Filter(emotionQuery)
		}
	}

	//the BFBS origin one is a bug, which
	//Optimize skip if qtype is all included, which will be meaningless to use this condition
	if query.QTypes != nil && len(query.QTypes) < 3 {
		terms := make(map[string]bool, 0)
		haveOther := false

		for _, typ := range query.QTypes {
			switch typ {

			case data.CategoryBusiness:
				terms["faq"] = true
				terms["task_engine"] = true
			case data.CategoryChat:
				terms["chat"] = true
			case data.CategoryOther:
				haveOther = true
			}

		}
		if len(terms) > 0 {
			var inputs = make([]interface{}, 0)
			for name := range terms {
				inputs = append(inputs, name)
			}
			boolQuery = boolQuery.Filter(elastic.NewTermsQuery("module", inputs...))
		}

		if haveOther {
			otherTermsQuery := elastic.NewTermsQuery("module", "faq", "task_engine", "chat")
			boolQuery = boolQuery.MustNot(otherTermsQuery)
		}
	}

	//  tags should contain following category info
	//{
	//   "type": "platform",
	// },
	// {
	//   "type": "brand",
	// },
	// {
	//   "type": "sex",
	// },
	// {
	//   "type": "age",
	//   "type_text": "年龄段",
	// },
	// {
	//   "type": "hobbies",
	// }
	tags := make([]data.VisitRecordsQueryTag, 0)
	if query.Platforms != nil {
		var tag = data.VisitRecordsQueryTag{
			Type:  "platform",
			Texts: query.Platforms,
		}
		tags = append(tags, tag)
	}
	if query.Genders != nil {
		inputs := make([]string, len(query.Genders))
		for i, v := range query.Genders {
			inputs[i] = genderDict[v]
		}
		tags = append(tags, data.VisitRecordsQueryTag{
			Type:  "sex",
			Texts: inputs,
		})
	}
	if len(tags) > 0 {
		tagsShouldBoolQuery := elastic.NewBoolQuery()
		tagsBoolQueries, err := createTagsBoolQueries(make([]*elastic.BoolQuery, 0), tags, 0, make([]data.VisitRecordsTag, 0))
		if err != nil {
			//Temp log report
			log.Printf("err %v", err)
			return boolQuery
		}

		for _, tagsBoolQuery := range tagsBoolQueries {
			tagsShouldBoolQuery = tagsShouldBoolQuery.Should(tagsBoolQuery)
		}

		boolQuery.Filter(tagsShouldBoolQuery)
	}
	if query.IsMarked != nil {
		if *query.IsMarked {
			boolQuery.Filter(elastic.NewTermQuery("isMarked", true))
		} else {
			q := elastic.NewBoolQuery()
			q.Should(elastic.NewBoolQuery().MustNot(elastic.NewExistsQuery("isMarked")))
			q.Should(elastic.NewBoolQuery().Filter(elastic.NewTermQuery("isMarked", false)))
			boolQuery.Filter(q)
		}
	}
	if query.IsIgnored != nil {
		if *query.IsIgnored {
			boolQuery.Filter(elastic.NewTermQuery("isIgnored", true))
		} else {
			q := elastic.NewBoolQuery()
			q.Should(elastic.NewBoolQuery().MustNot(elastic.NewExistsQuery("isIgnored")))
			q.Should(elastic.NewBoolQuery().Filter(elastic.NewTermQuery("isIgnored", false)))
			boolQuery.Filter(q)
		}
	}
	return boolQuery
}

type UpdateCommand func(*elastic.UpdateByQueryService) *elastic.UpdateByQueryService

func UpdateRecordMark(status bool) UpdateCommand {
	return func(qs *elastic.UpdateByQueryService) *elastic.UpdateByQueryService {
		script := elastic.NewScript("ctx._source.isMarked = params.mark")
		script.Param("mark", status)
		qs.Script(script)
		return qs
	}
}

func UpdateRecordIgnore(status bool) UpdateCommand {
	return func(qs *elastic.UpdateByQueryService) *elastic.UpdateByQueryService {
		script := elastic.NewScript("ctx._source.isIgnored = params.ignore")
		script.Param("ignore", status)
		qs.Script(script)
		return qs
	}
}

//UpdateRecords will update the records based on given query
//It will return error if any problem occur.
func UpdateRecords(query data.RecordQuery, cmd UpdateCommand) error {
	ctx, client := elasticsearch.GetClient()
	index := fmt.Sprintf("%s-%s-*", data.ESRecordsIndex, query.AppID)
	bq := NewBoolQueryWithRecordQuery(query)
	traceQuery(bq)
	s := client.UpdateByQuery(index)
	s.Type(data.ESRecordType)
	s.Query(bq)
	s.ProceedOnVersionConflict()
	s = cmd(s)

	resp, err := s.Do(ctx)
	if err != nil {
		return err
	}
	logger.Trace.Printf("response: %+v\n", resp)
	return nil
}

func traceQuery(bq *elastic.BoolQuery) {
	src, _ := bq.Source()
	d, _ := json.Marshal(src)
	logger.Trace.Printf("query: %s\n", d)
}

//VisitRecordsQuery will fetch raw and aggregation data from Data store.
//input query decide which raw data to fetchand & aggs will decided Result's aggs
//In the return, result.Aggs may contain isMarked or isIgnored key if correspond ElasticSearchCommand have bee given in aggs
func VisitRecordsQuery(query data.RecordQuery, aggs ...ElasticSearchCommand) (*RecordResult, error) {
	ctx, client := elasticsearch.GetClient()
	index := fmt.Sprintf("%s-%s-*", data.ESRecordsIndex, query.AppID)
	logger.Trace.Printf("index: %s\n", index)
	boolQuery := NewBoolQueryWithRecordQuery(query)
	traceQuery(boolQuery)
	source := elastic.NewFetchSourceContext(true)
	source.Include(
		data.VisitRecordsMetricSessionID,
		data.VisitRecordsMetricUserID,
		data.VisitRecordsMetricUserQ,
		data.VisitRecordsMetricScore,
		data.VisitRecordsMetricStdQ,
		"answer.value",
		data.VisitRecordsMetricLogTime,
		data.VisitRecordsMetricEmotion,
		"module",
		"unique_id",
		"isMarked",
		"isIgnored",
	)
	ss := client.Search()
	for _, agg := range aggs {
		agg(ss)
	}
	result, err := ss.
		Index(index).
		Type(data.ESRecordType).
		Query(boolQuery).
		From(int(query.From)).
		Size(query.Limit).
		FetchSourceContext(source).
		Sort(data.LogTimeFieldName, false).
		Do(ctx)
	if err != nil {
		return nil, err
	}

	var r = &RecordResult{
		Hits: make([]*data.VisitRecordsData, 0),
		Aggs: map[string]interface{}{
			"total_size": result.TotalHits(),
		},
	}
	if markedSize, found := result.Aggregations.Filter("isMarked"); found {
		r.Aggs["isMarked"] = markedSize.DocCount
	}
	if ignoredSize, found := result.Aggregations.Filter("isIgnored"); found {
		r.Aggs["isIgnored"] = ignoredSize.DocCount
	}

	for _, hit := range result.Hits.Hits {
		hitResult := data.VisitRecordsHitResult{}
		jsonStr, _ := hit.Source.MarshalJSON()
		err = json.Unmarshal(jsonStr, &hitResult)
		if err != nil {
			return nil, fmt.Errorf("hit result unmarshal failed, %v", err)
		}

		// Note: We have to convert log_time (UTC+0) to local time and reformat
		logTime, _err := time.Parse(data.LogTimeFormat, hitResult.LogTime)
		if _err != nil {
			err = fmt.Errorf("Result LogTime cant not parse into time, reason: %v", _err)
			return nil, err
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

		record := &data.VisitRecordsData{
			VisitRecordsDataBase: data.VisitRecordsDataBase{
				SessionID: rawRecord.SessionID,
				UserID:    rawRecord.UserID,
				UserQ:     rawRecord.UserQ,
				Score:     rawRecord.Score,
				StdQ:      rawRecord.StdQ,
				LogTime:   rawRecord.LogTime,
				Emotion:   rawRecord.Emotion,
				QType:     rawRecord.QType,
				IsMarked:  rawRecord.IsMarked,
				IsIgnored: rawRecord.IsIgnored,
			},
			UniqueID: rawRecord.UniqueID,
			Answer:   strings.Join(answers, ", "),
		}
		r.Hits = append(r.Hits, record)
	}

	return r, nil
}

// createTagsBoolQueries converts queryTags = [
//     {
//         Type: "platform",
//         Texts: ["android", "ios"]
//     },
//     {
//         Type: "sex",
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
//           "custom_info.platform.keyword": "android"
//         }
//       },
//       {
//         "term": {
//       	"custom_info.sex.keyword": "男"
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
//           "custom_info.platform.keyword": "android"
//         }
//       },
//       {
//         "term": {
//           "custom_info.sex.keyword": "女"
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
//           "custom_info.platform.keyword": "ios"
//         }
//       },
//       {
//         "term": {
//           "custom_info.sex.keyword": "男"
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
//           "custom_info.platform.keyword": "ios"
//         }
//       },
//       {
//         "term": {
//           "custom_info.sex.keyword": "女"
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
				field := fmt.Sprintf("custom_info.%s.keyword", tag.Type)
				termQuery := elastic.NewTermQuery(field, tag.Text)
				boolQuery = boolQuery.Filter(termQuery)
			}

			field := fmt.Sprintf("custom_info.%s.keyword", queryTags[queryTagsIndex].Type)
			termQuery := elastic.NewTermQuery(field, text)
			boolQuery = boolQuery.Filter(termQuery)

			boolQueries = append(boolQueries, boolQuery)
		}

		return boolQueries, nil
	}

	return tagsBoolQueries, nil
}
