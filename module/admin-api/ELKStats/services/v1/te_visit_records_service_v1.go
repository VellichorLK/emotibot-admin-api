package v1

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"emotibot.com/emotigo/module/admin-api/ELKStats/dao"
	"emotibot.com/emotigo/module/admin-api/ELKStats/data"
	dataCommon "emotibot.com/emotigo/module/admin-api/ELKStats/data/common"
	dataV1 "emotibot.com/emotigo/module/admin-api/ELKStats/data/v1"
	"emotibot.com/emotigo/module/admin-api/ELKStats/services"
	servicesCommon "emotibot.com/emotigo/module/admin-api/ELKStats/services/common"
	"emotibot.com/emotigo/module/admin-api/util/elasticsearch"
	"github.com/tealeg/xlsx"
	elastic "gopkg.in/olivere/elastic.v6"
)

func TEVisitRecordsQuery(query *dataV1.TEVisitRecordsQuery) (teRecords []*dataV1.TEVisitRecordsData, totalSize int64, err error) {
	ctx, client := elasticsearch.GetClient()
	boolQuery := newTEBoolQueryWithRecordQuery(query)

	source := elastic.NewFetchSourceContext(true)
	source.Include(
		dataCommon.TEVisitRecordsMetricTESessionID,
		dataCommon.TEVisitRecordsMetricSessionID,
		dataCommon.TEVisitRecordsMetricUserID,
		dataCommon.TEVisitRecordsMetricScenarioID,
		dataCommon.TEVisitRecordsMetricScenarioName,
		dataCommon.TEVisitRecordsMetricLastNodeID,
		dataCommon.TEVisitRecordsMetricLastNodeName,
		dataCommon.TEVisitRecordsMetricTriggerTime,
		dataCommon.TEVisitRecordsMetricFinishTime,
		dataCommon.TEVisitRecordsMetricFeedback,
		dataCommon.TEVisitRecordsMetricCustomFeedback,
		dataCommon.TEVisitRecordsMetricFeedbackTime,
	)

	index := fmt.Sprintf("%s-*", data.ESTERecordsIndex)

	results, err := client.Search().
		Index(index).
		Type(data.ESTERecordsType).
		Query(boolQuery).
		FetchSourceContext(source).
		From(int(query.From)).
		Size(int(query.Limit)).
		Sort(data.TERecordsTriggerTimeFieldName, false).
		Do(ctx)
	if err != nil {
		return nil, 0, err
	}

	teRecords = make([]*dataV1.TEVisitRecordsData, 0)
	totalSize = results.Hits.TotalHits

	for _, hit := range results.Hits.Hits {
		rawTERecord := dataV1.TEVisitRecordsRawData{}
		jsonStr, err := hit.Source.MarshalJSON()
		if err != nil {
			return nil, 0, err
		}

		err = json.Unmarshal(jsonStr, &rawTERecord)
		if err != nil {
			return nil, 0, err
		}

		teRecord, err := extractRawTERecord(&rawTERecord)
		if err != nil {
			return nil, 0, err
		}

		teRecords = append(teRecords, teRecord)
	}

	return
}

func TEVisitRecordsExport(query *dataV1.TEVisitRecordsQuery, locale string) (exportTaskID string, err error) {
	// Try to create export task
	exportTaskID, err = dao.TryCreateExportTask(query.EnterpriseID)
	if err != nil {
		return
	}

	// Create a goroutine to exporting records in background
	go func() {
		option := createExportTERecordsTaskOption(query, exportTaskID, locale)
		servicesCommon.ExportTask(option, locale)
	}()

	return
}

func TEVisitRecordsExportDownload(w http.ResponseWriter, exportTaskID string) error {
	return servicesCommon.DownloadExportRecords(w, exportTaskID)
}

func TEVisitRecordsExportDelete(exportTaskID string) error {
	return servicesCommon.DeleteExportRecords(exportTaskID)
}

func TEVisitRecordsExportStatus(exportTaskID string) (status string, err error) {
	status, err = servicesCommon.GetRecordsExportStatus(exportTaskID)
	return
}

// newTEBoolQueryWithRecordQuery create a *elastic.BoolQuery with the Record condition.
func newTEBoolQueryWithRecordQuery(query *dataV1.TEVisitRecordsQuery) *elastic.BoolQuery {
	boolQuery := elastic.NewBoolQuery()

	if query.AppID != "" {
		appTermQuery := elastic.NewTermQuery("app_id", query.AppID)
		boolQuery.Filter(appTermQuery)
	}

	// Start time & End time
	rangeQuery := services.CreateRangeQueryUnixTime(query.CommonQuery,
		data.TERecordsTriggerTimeFieldName)
	boolQuery.Filter(rangeQuery)

	// Scenario name
	if query.ScenarioName != nil && *query.ScenarioName != "" {
		boolQuery.Filter(elastic.NewTermQuery("scenario_name", *query.ScenarioName))
	}

	// User ID
	if query.UserID != nil && *query.UserID != "" {
		boolQuery.Filter(elastic.NewTermQuery("user_id", *query.UserID))
	}

	// Feedback
	if query.Feedback != nil && *query.Feedback != "" {
		feedbackTermQuery := elastic.NewTermQuery("feedback", *query.Feedback)
		customFeedbackTermQuery := elastic.NewTermQuery("custom_feedback.keyword",
			*query.Feedback)

		boolQuery.Filter(elastic.NewBoolQuery().Should(feedbackTermQuery, customFeedbackTermQuery))
	}

	// Platforms & Genders
	tags := make([]data.QueryTags, 0)

	if len(query.Platforms) > 0 {
		tags = append(tags, data.QueryTags{
			Type:  "platform",
			Texts: query.Platforms,
		})
	}

	if len(query.Genders) > 0 {
		tags = append(tags, data.QueryTags{
			Type:  "sex",
			Texts: query.Genders,
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

	return boolQuery
}

func extractRawTERecord(rawTERecord *dataV1.TEVisitRecordsRawData) (*dataV1.TEVisitRecordsData, error) {
	teRecord := dataV1.TEVisitRecordsData{
		TEVisitRecordsDataBase: dataV1.TEVisitRecordsDataBase{
			TESessionID:    rawTERecord.TESessionID,
			SessionID:      rawTERecord.SessionID,
			UserID:         rawTERecord.UserID,
			ScenarioID:     rawTERecord.ScenarioID,
			ScenarioName:   rawTERecord.ScenarioName,
			LastNodeID:     rawTERecord.LastNodeID,
			LastNodeName:   rawTERecord.LastNodeName,
			Feedback:       rawTERecord.Feedback,
			CustomFeedback: rawTERecord.CustomFeedback,
		},
	}

	// Convert trigger time, finish time and feedback time
	// from UTC+0 back to local time
	var triggerTime string
	if rawTERecord.TriggerTime != 0 {
		triggerTime = time.Unix(rawTERecord.TriggerTime, 0).Local().
			Format(data.StandardTimeFormat)
	}

	var finishTime string
	if rawTERecord.FinishTime != 0 {
		finishTime = time.Unix(rawTERecord.FinishTime, 0).Local().
			Format(data.StandardTimeFormat)
	}

	var feedbackTime string
	if rawTERecord.FeedbackTime != 0 {
		feedbackTime = time.Unix(rawTERecord.FeedbackTime, 0).Local().
			Format(data.StandardTimeFormat)
	}

	teRecord.TriggerTime = triggerTime
	teRecord.FinishTime = finishTime
	teRecord.FeedbackTime = feedbackTime

	return &teRecord, nil
}

func extractExportTERecordsHitResultHandler(hit *elastic.SearchHit) (teRecordPtr interface{}, err error) {
	rawTERecord := dataV1.TEVisitRecordsRawData{}
	jsonStr, err := hit.Source.MarshalJSON()
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(jsonStr, &rawTERecord)
	if err != nil {
		return
	}

	teRecord, err := extractRawTERecord(&rawTERecord)
	if err != nil {
		return nil, err
	}

	teRecordPtr = teRecord
	return
}

func createExportTERecordsXlsx(teRecordPtrs []interface{}, xlsxFileName string, locale string,
	params ...interface{}) (xlsxFilePath string, err error) {
	dirPath, _err := servicesCommon.GetExportRecordsDir()
	if _err != nil {
		err = _err
		return
	}

	xlsxFilePath = fmt.Sprintf("%s/%s.xlsx", dirPath, xlsxFileName)
	xlsxFile := xlsx.NewFile()
	sheet, _err := xlsxFile.AddSheet("多轮场景日志导出数据")
	if err != nil {
		err = _err
		return
	}

	// Header row
	row := sheet.AddRow()
	for _, header := range dataV1.TEVisitRecordsExportHeader {
		cell := row.AddCell()
		cell.Value = header
	}

	// Data rows
	for _, teRecordPtr := range teRecordPtrs {
		teRecord := teRecordPtr.(*dataV1.TEVisitRecordsData)

		row := sheet.AddRow()
		xlsxData := []string{
			teRecord.TESessionID,
			teRecord.SessionID,
			teRecord.UserID,
			teRecord.ScenarioID,
			teRecord.ScenarioName,
			teRecord.LastNodeID,
			teRecord.LastNodeName,
			teRecord.TriggerTime,
			teRecord.FinishTime,
			teRecord.Feedback,
			teRecord.CustomFeedback,
			teRecord.FeedbackTime,
		}

		for _, d := range xlsxData {
			cell := row.AddCell()
			cell.Value = d
		}
	}

	err = xlsxFile.Save(xlsxFilePath)
	return
}

func createExportTERecordsTaskOption(query *dataV1.TEVisitRecordsQuery,
	exportTaskID string, locale string) *data.ExportTaskOption {
	index := fmt.Sprintf("%s-*", data.ESTERecordsIndex)

	boolQuery := newTEBoolQueryWithRecordQuery(query)

	source := elastic.NewFetchSourceContext(true)
	source.Include(
		dataCommon.TEVisitRecordsMetricTESessionID,
		dataCommon.TEVisitRecordsMetricSessionID,
		dataCommon.TEVisitRecordsMetricUserID,
		dataCommon.TEVisitRecordsMetricScenarioID,
		dataCommon.TEVisitRecordsMetricScenarioName,
		dataCommon.TEVisitRecordsMetricLastNodeID,
		dataCommon.TEVisitRecordsMetricLastNodeName,
		dataCommon.TEVisitRecordsMetricTriggerTime,
		dataCommon.TEVisitRecordsMetricFinishTime,
		dataCommon.TEVisitRecordsMetricFeedback,
		dataCommon.TEVisitRecordsMetricCustomFeedback,
		dataCommon.TEVisitRecordsMetricFeedbackTime,
	)

	return &data.ExportTaskOption{
		TaskID:               exportTaskID,
		Index:                index,
		BoolQuery:            boolQuery,
		Source:               source,
		SortField:            data.TERecordsTriggerTimeFieldName,
		ExtractResultHandler: extractExportTERecordsHitResultHandler,
		XlsxCreateHandler:    createExportTERecordsXlsx,
	}
}
