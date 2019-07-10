package v1

import (
	"emotibot.com/emotigo/module/admin-api/ELKStats/data"
	dataCommon "emotibot.com/emotigo/module/admin-api/ELKStats/data/common"
	dataV1 "emotibot.com/emotigo/module/admin-api/ELKStats/data/v1"
	servicesCommon "emotibot.com/emotigo/module/admin-api/ELKStats/services/common"
	"emotibot.com/emotigo/module/admin-api/util/elasticsearch"
	esData "emotibot.com/emotigo/module/admin-api/util/elasticsearch/data"
	"emotibot.com/emotigo/pkg/logger"
	"encoding/json"
	"fmt"
	elastic "gopkg.in/olivere/elastic.v6"
	"time"
)

type CcsRecordResult struct {
	Hits []*dataV1.VisitCcsRecordsData
	Aggs map[string]interface{}
}

func VisitCcsRecordsQuery(query dataV1.CcsRecordQuery, aggs ...servicesCommon.ElasticSearchCommand) (*CcsRecordResult, error) {
	ctx, client := elasticsearch.GetClient()
	index := fmt.Sprintf("%s-*", esData.ESCcsRecordsIndex)
	logger.Trace.Printf("index: %s\n", index)
	boolQuery := newBoolQueryWithCcsRecordQuery(&query)
	traceQuery(boolQuery)
	source := elastic.NewFetchSourceContext(true)
	source.Include(
		dataCommon.VisitRecordsMetricSessionID,
		dataCommon.VisitRecordsMetricUserID,
		dataCommon.VisitRecordsMetricUserQ,
		dataCommon.VisitRecordsMetricAnswer,
		dataCommon.VisitRecordsMetricLogTime,
		dataCommon.VisitRecordsMetricModule,
		dataCommon.VisitRecordsMetricUnique,
		dataCommon.VisitRecordsMetricRawResponse,
		dataCommon.VisitRecordsMetricRuleIDs,
		dataCommon.VisitRecordsMetricTaskID,
		dataCommon.VisitTagsMetricCaseId,
		dataCommon.VisitTagsMetricAppId,
		dataCommon.VisitTagsMetricDataset,
	)

	ss := client.Search()
	for _, agg := range aggs {
		agg(ss)
	}
	result, err := ss.
		Index(index).
		Type(esData.ESRecordType).
		Query(boolQuery).
		From(int(query.From)).
		Size(query.Limit).
		FetchSourceContext(source).
		Sort(data.LogTimeFieldName, false).
		Do(ctx)
	if err != nil {
		return nil, err
	}

	var r = &CcsRecordResult{
		Hits: make([]*dataV1.VisitCcsRecordsData, 0),
		Aggs: map[string]interface{}{
			"total_size": result.TotalHits(),
		},
	}

	for _, hit := range result.Hits.Hits {
		hitResult := dataV1.VisitCcsRecordsData{}
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
		r.Hits = append(r.Hits, &hitResult)
	}

	return r, nil
}

func newBoolQueryWithCcsRecordQuery(query *dataV1.CcsRecordQuery) *elastic.BoolQuery {
	boolQuery := elastic.NewBoolQuery()

	if query.AppID != "" {
		appTermQuery := elastic.NewTermQuery("app_id", query.AppID)
		boolQuery = boolQuery.Filter(appTermQuery)
	}

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

			case dataCommon.CategoryBusiness:
				terms["faq"] = true
				terms["task_engine"] = true
			case dataCommon.CategoryChat:
				terms["chat"] = true
			case dataCommon.CategoryOther:
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

	tags := make([]data.QueryTags, 0)
	if query.Platforms != nil {
		var tag = data.QueryTags{
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
		tags = append(tags, data.QueryTags{
			Type:  "sex",
			Texts: inputs,
		})
	}
	if len(tags) > 0 {
		tagsShouldBoolQuery := elastic.NewBoolQuery()
		tagsBoolQueries := servicesCommon.CreateTagsBoolQueries(make([]*elastic.BoolQuery, 0),
			tags, 0, make([]data.QueryTag, 0))

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