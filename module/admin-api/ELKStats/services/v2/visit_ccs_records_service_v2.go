package v2

import (
	"emotibot.com/emotigo/module/admin-api/ELKStats/data"
	dataCommon "emotibot.com/emotigo/module/admin-api/ELKStats/data/common"
	dataV2 "emotibot.com/emotigo/module/admin-api/ELKStats/data/v2"
	"emotibot.com/emotigo/module/admin-api/ELKStats/services"
	servicesCommon "emotibot.com/emotigo/module/admin-api/ELKStats/services/common"
	"emotibot.com/emotigo/module/admin-api/util/elasticsearch"
	esData "emotibot.com/emotigo/module/admin-api/util/elasticsearch/data"
	"encoding/json"
	"fmt"
	elastic "gopkg.in/olivere/elastic.v6"
	"strings"
)

func VisitCcsRecordsQuery(query *dataV2.VisitCcsRecordsQuery,
	aggs ...servicesCommon.ElasticSearchCommand) (queryResult *dataV2.VisitCcsRecordsQueryResult, err error) {
	ctx, client := elasticsearch.GetClient()
	boolQuery := NewBoolQueryWithCcsRecordQuery(query)

	source := elastic.NewFetchSourceContext(true)
	source.Include(
		dataCommon.VisitRecordsMetricSessionID,
		dataCommon.VisitRecordsMetricUserID,
		dataCommon.VisitRecordsMetricUserQ,
		dataCommon.VisitRecordsMetricLogTime,
		dataCommon.VisitRecordsMetricRawResponse,
		dataCommon.VisitTagsMetricAnswers,
		dataCommon.VisitTagsMetricAIModule,
	)

	index := fmt.Sprintf("%s-*", esData.ESCcsRecordsIndex)
	ss := client.Search()

	for _, agg := range aggs {
		agg(ss)
	}

	result, err := ss.
		Index(index).
		Type(esData.ESRecordType).
		Query(boolQuery).
		FetchSourceContext(source).
		From(int(query.From)).
		Size(int(query.Limit)).
		Sort(data.LogTimeFieldName, false).
		Do(ctx)
	if err != nil {
		return nil, err
	}

	queryResult = &dataV2.VisitCcsRecordsQueryResult{
		TotalSize: result.Hits.TotalHits,
	}

	records := make([]*dataV2.VisitCcsRecordsData, 0)

	for _, hit := range result.Hits.Hits {
		rawRecord := dataV2.VisitCcsRecordsRawData{}
		jsonStr, err := hit.Source.MarshalJSON()
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(jsonStr, &rawRecord)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}

		logTime := rawRecord.LogTime
		logTime = logTime[:len(logTime)-5]
		logTime = strings.Replace(logTime, "T", " ", 1)

		record := &dataV2.VisitCcsRecordsData{
			SessionID:	rawRecord.SessionID,
			UserID:		rawRecord.UserID,
			UserQ:		rawRecord.UserQ,
			LogTime:	rawRecord.LogTime,
			RawResponse:rawRecord.RawResponse,
			Answers:	rawRecord.Answers,
			AIModule:	rawRecord.AIModule,
		}

		records = append(records, record)
	}

	queryResult.Data = records
	return
}

func NewBoolQueryWithCcsRecordQuery(query *dataV2.VisitCcsRecordsQuery) *elastic.BoolQuery {
	boolQuery := elastic.NewBoolQuery()

	if query.AppID != "" {
		appTermQuery := elastic.NewTermQuery("app_id", query.AppID)
		boolQuery.Filter(appTermQuery)
	}

	// Session ID
	if query.SessionID != nil && *query.SessionID != "" {
		boolQuery.Filter(elastic.NewTermQuery("session_id", *query.SessionID))
	}

	// Start time & End time
	rangeQuery := services.CreateRangeQueryUnixTime(query.CommonQuery,
		data.LogTimeFieldName)
	boolQuery.Filter(rangeQuery)

	return boolQuery
}