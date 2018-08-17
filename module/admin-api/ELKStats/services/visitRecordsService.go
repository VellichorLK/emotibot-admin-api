package services

import (
	"context"
	"encoding/json"
	"time"

	"emotibot.com/emotigo/module/admin-api/ELKStats/data"
	"github.com/olivere/elastic"
)

const RecordsPageLimit = 20

func VisitRecordsQuery(ctx context.Context, client *elastic.Client,
	query data.VisitRecordsQuery) (records []*data.VisitRecordsData, totalSize int64, limit int, err error) {
	boolQuery := createBoolQuery(query.CommonQuery)
	rangeQuery := createRangeQuery(query.CommonQuery, data.LogTimeFieldName)
	boolQuery = boolQuery.Filter(rangeQuery)

	if query.Question != "" {
		userQTermQuery := elastic.NewTermQuery("user_q", query.Question)
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

	from := (query.Page - 1) * RecordsPageLimit
	source := elastic.NewFetchSourceContext(true)
	source.Include(
		data.VisitRecordsMetricUserID,
		data.VisitRecordsMetricUserQ,
		data.VisitRecordsMetricScore,
		data.VisitRecordsMetricStdQ,
		data.VisitRecordsMetricLogTime,
		data.VisitRecordsMetricEmotion,
		"module",
	)

	result, err := client.Search().
		Index(data.ESRecordsIndex).
		Type(data.ESRecordType).
		Query(boolQuery).
		From(from).
		Size(RecordsPageLimit).
		FetchSourceContext(source).
		Sort(data.LogTimeFieldName, false).
		Do(ctx)
	if err != nil {
		return
	}

	totalSize = result.TotalHits()
	limit = RecordsPageLimit

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

		record := hitResult.VisitRecordsData

		if hitResult.Module == "faq" || hitResult.Module == "task_engine" {
			record.QType = "业务类"
		} else if hitResult.Module == "chat" {
			record.QType = "聊天类"
		} else {
			record.QType = "其他类"
		}

		records = append(records, &record)
	}

	return
}
