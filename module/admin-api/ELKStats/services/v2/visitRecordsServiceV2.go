package v2

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"emotibot.com/emotigo/module/admin-api/ELKStats/dao"
	"emotibot.com/emotigo/module/admin-api/ELKStats/data"
	dataCommon "emotibot.com/emotigo/module/admin-api/ELKStats/data/common"
	dataV2 "emotibot.com/emotigo/module/admin-api/ELKStats/data/v2"
	"emotibot.com/emotigo/module/admin-api/ELKStats/services"
	servicesCommon "emotibot.com/emotigo/module/admin-api/ELKStats/services/common"
	"emotibot.com/emotigo/module/admin-api/util/elasticsearch"
	"emotibot.com/emotigo/pkg/logger"
	"github.com/olivere/elastic"
	"github.com/tealeg/xlsx"
)

func VisitRecordsQuery(query *dataV2.VisitRecordsQuery,
	aggs ...servicesCommon.ElasticSearchCommand) (queryResult *dataV2.VisitRecordsQueryResult, err error) {
	ctx, client := elasticsearch.GetClient()
	boolQuery := newBoolQueryWithRecordQuery(query)

	source := elastic.NewFetchSourceContext(true)
	source.Include(
		dataCommon.VisitRecordsMetricSessionID,
		dataCommon.VisitRecordsMetricTESessionID,
		dataCommon.VisitRecordsMetricUserID,
		dataCommon.VisitRecordsMetricUserQ,
		dataCommon.VisitRecordsMetricScore,
		dataCommon.VisitRecordsMetricStdQ,
		"answer.value",
		dataCommon.VisitRecordsMetricLogTime,
		dataCommon.VisitRecordsMetricEmotion,
		dataCommon.VisitRecordsMetricIntent,
		dataCommon.VisitRecordsMetricModule,
		"unique_id",
		"isMarked",
		"isIgnored",
		"faq_cat_id",
		"faq_robot_tag_id",
		dataCommon.VisitRecordsMetricFeedback,
		dataCommon.VisitRecordsMetricCustomFeedback,
		dataCommon.VisitRecordsMetricFeedbackTime,
		dataCommon.VisitRecordsMetricThreshold,
	)

	index := fmt.Sprintf("%s-*", data.ESRecordsIndex)
	ss := client.Search()

	for _, agg := range aggs {
		agg(ss)
	}

	result, err := ss.
		Index(index).
		Type(data.ESRecordType).
		Query(boolQuery).
		FetchSourceContext(source).
		From(int(query.From)).
		Size(int(query.Limit)).
		Sort(data.LogTimeFieldName, false).
		Do(ctx)
	if err != nil {
		return nil, err
	}

	queryResult = &dataV2.VisitRecordsQueryResult{
		TotalSize: result.Hits.TotalHits,
	}

	if markedSize, found := result.Aggregations.Filter("isMarked"); found {
		queryResult.MarkedSize = markedSize.DocCount
	} else {
		queryResult.MarkedSize = 0
	}

	if ignoredSize, found := result.Aggregations.Filter("isIgnored"); found {
		queryResult.IgnoredSize = ignoredSize.DocCount
	} else {
		queryResult.IgnoredSize = 0
	}

	records := make([]*dataV2.VisitRecordsData, 0)

	for _, hit := range result.Hits.Hits {
		rawRecord := dataV2.VisitRecordsRawData{}
		jsonStr, err := hit.Source.MarshalJSON()
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(jsonStr, &rawRecord)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}

		recordCommon, err := extractRawRecord(&rawRecord)
		if err != nil {
			return nil, err
		}

		record := &dataV2.VisitRecordsData{
			VisitRecordsCommon: *recordCommon,
			UniqueID:           rawRecord.UniqueID,
			IsMarked:           rawRecord.IsMarked,
			IsIgnored:          rawRecord.IsIgnored,
		}

		records = append(records, record)
	}

	// Convert FAQ category ID to FAQ category names
	err = updateFaqCategoryNames(records)
	if err != nil {
		return nil, err
	}

	// Convert FAQ robot tag IDs to FAQ FAQ robot tag names
	err = updateFaqRobotTags(records)
	if err != nil {
		return nil, err
	}

	queryResult.Data = records
	return
}

func VisitRecordsExport(query *dataV2.VisitRecordsQuery) (exportTaskID string, err error) {
	// Try to create export task
	exportTaskID, err = dao.TryCreateExportTask(query.EnterpriseID)
	if err != nil {
		return
	}

	// Create a goroutine to exporting records in background
	go func() {
		option := createExportRecordsTaskOption(query, exportTaskID)
		servicesCommon.ExportTask(option)
	}()

	return
}

func VisitRecordsExportDownload(w http.ResponseWriter, exportTaskID string) error {
	return servicesCommon.VisitRecordsExportDownload(w, exportTaskID)
}

func VisitRecordsExportDelete(exportTaskID string) error {
	return servicesCommon.VisitRecordsExportDelete(exportTaskID)
}

func VisitRecordsExportStatus(exportTaskID string) (status string, err error) {
	status, err = servicesCommon.VisitRecordsExportStatus(exportTaskID)
	return
}

func UpdateRecordMark(status bool) servicesCommon.UpdateCommand {
	return servicesCommon.UpdateRecordMark(status)
}

func UpdateRecordIgnore(status bool) servicesCommon.UpdateCommand {
	return servicesCommon.UpdateRecordIgnore(status)
}

// UpdateRecords will update the records based on given query
// It will return error if any problem occur.
func UpdateRecords(query *dataV2.VisitRecordsQuery, cmd servicesCommon.UpdateCommand) error {
	ctx, client := elasticsearch.GetClient()
	boolQuery := newBoolQueryWithRecordQuery(query)

	index := fmt.Sprintf("%s-*", data.ESRecordsIndex)
	s := client.UpdateByQuery(index)
	s.Type(data.ESRecordType)
	s.Query(boolQuery)
	s.ProceedOnVersionConflict()
	s = cmd(s)

	resp, err := s.Do(ctx)
	if err != nil {
		return err
	}

	logger.Trace.Printf("Response: %+v\n", resp)
	return nil
}

func extractRawRecord(rawRecord *dataV2.VisitRecordsRawData) (*dataV2.VisitRecordsCommon, error) {
	// Log time
	// Note: We have to convert log_time (UTC+0) to local time and reformat
	logTime, err := time.Parse(data.LogTimeFormat, rawRecord.LogTime)
	if err != nil {
		return nil, err
	}
	logTime = logTime.Local()

	// Answers
	answers := make([]string, 0)
	for _, answer := range rawRecord.Answer {
		answers = append(answers, answer.Value)
	}

	// Feedback time
	var feedbackTime string
	if rawRecord.FeedbackTime != 0 {
		feedbackTime = time.Unix(rawRecord.FeedbackTime, 0).
			Local().Format(data.StandardTimeFormat)
	}

	record := &dataV2.VisitRecordsCommon{
		VisitRecordsDataBase: dataV2.VisitRecordsDataBase{
			SessionID:   rawRecord.SessionID,
			TESessionID: rawRecord.TESessionID,
			UserID:      rawRecord.UserID,
			UserQ:       rawRecord.UserQ,
			Score:       rawRecord.Score,
			StdQ:        rawRecord.StdQ,
			LogTime:     logTime.Format(data.StandardTimeFormat),
			Emotion:     rawRecord.Emotion,
			Intent:      rawRecord.Intent,
			Module:      rawRecord.Module,
		},
		Answer:         strings.Join(answers, ", "),
		FaqCategoryID:  rawRecord.FaqCategoryID,
		FaqRobotTagIDs: rawRecord.FaqRobotTagIDs,
		Feedback:       rawRecord.Feedback,
		CustomFeedback: rawRecord.CustomFeedback,
		FeedbackTime:   feedbackTime,
		Threshold:      rawRecord.Threshold,
	}

	return record, nil
}

func updateFaqCategoryNames(records []*dataV2.VisitRecordsData) error {
	categoryPaths, err := dao.GetAllFaqCategoryPaths()
	if err != nil {
		return err
	}

	for _, record := range records {
		if record.FaqCategoryID != 0 {
			if categoryPath, ok := categoryPaths[record.FaqCategoryID]; ok {
				record.FaqCategoryName = categoryPath.Path
			}
		}
	}

	return nil
}

func updateFaqRobotTags(records []*dataV2.VisitRecordsData) error {
	robotTags, err := dao.GetAllFaqRobotTags()
	if err != nil {
		return err
	}

	for _, record := range records {
		if len(record.FaqRobotTagIDs) > 0 {
			robotTagNames := make([]string, 0)

			for _, robotTagID := range record.FaqRobotTagIDs {
				if robotTag, ok := robotTags[robotTagID]; ok {
					robotTagNames = append(robotTagNames, robotTag.Tag)
				}
			}

			record.FaqRobotTagNames = strings.Join(robotTagNames, ", ")
		}
	}

	return nil
}

func extractExportRecordsHitResultHandler(hit *elastic.SearchHit) (recordPtr interface{}, err error) {
	rawRecord := dataV2.VisitRecordsRawData{}
	jsonStr, err := hit.Source.MarshalJSON()
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(jsonStr, &rawRecord)
	if err != nil {
		return
	}

	recordCommon, err := extractRawRecord(&rawRecord)
	if err != nil {
		return nil, err
	}

	// Custom info
	var customInfo string
	if rawRecord.CustomInfo != nil {
		json, err := json.Marshal(rawRecord.CustomInfo)
		if err != nil {
			return nil, err
		}

		customInfo = string(json)
	}

	recordPtr = &dataV2.VisitRecordsExportData{
		VisitRecordsCommon: *recordCommon,
		Source:             rawRecord.Source,
		EmotionScore:       rawRecord.EmotionScore,
		IntentScore:        rawRecord.IntentScore,
		CustomInfo:         customInfo,
	}

	return
}

func createExportRecordsXlsx(recordPtrs []interface{}, xlsxFileName string) (xlsxFilePath string, err error) {
	dirPath, _err := servicesCommon.GetExportRecordsDir()
	if _err != nil {
		err = _err
		return
	}

	xlsxFilePath = fmt.Sprintf("%s/%s.xlsx", dirPath, xlsxFileName)
	xlsxFile := xlsx.NewFile()
	sheet, _err := xlsxFile.AddSheet("日志导出数据")
	if err != nil {
		err = _err
		return
	}

	faqCategoryPaths, err := dao.GetAllFaqCategoryPaths()
	if err != nil {
		return "", err
	}

	faqRobotTags, err := dao.GetAllFaqRobotTags()
	if err != nil {
		return "", err
	}

	// Header row
	row := sheet.AddRow()
	for _, header := range dataV2.VisitRecordsExportHeader {
		cell := row.AddCell()
		cell.Value = header
	}

	// Data rows
	for _, recordPtr := range recordPtrs {
		record := recordPtr.(*dataV2.VisitRecordsExportData)

		// Convert FAQ category ID to FAQ category names
		var faqCategoryName string
		if faqCategoryPath, ok := faqCategoryPaths[record.FaqCategoryID]; ok {
			faqCategoryName = faqCategoryPath.Path
		}

		// Convert FAQ robot tag IDs to FAQ FAQ robot tag names
		var faqRobotTagNames []string
		for _, faqRobotTagID := range record.FaqRobotTagIDs {
			if faqRobotTag, ok := faqRobotTags[faqRobotTagID]; ok {
				faqRobotTagNames = append(faqRobotTagNames, faqRobotTag.Tag)
			}
		}

		row := sheet.AddRow()
		xlsxData := []string{
			record.SessionID,
			record.TESessionID,
			record.UserID,
			record.UserQ,
			record.StdQ,
			record.Answer,
			strconv.FormatFloat(record.Score, 'f', -1, 64),
			record.Module,
			record.Source,
			record.LogTime,
			record.Emotion,
			strconv.FormatFloat(record.EmotionScore, 'f', -1, 64),
			record.Intent,
			strconv.FormatFloat(record.IntentScore, 'f', -1, 64),
			record.CustomInfo,
			faqCategoryName,
			strings.Join(faqRobotTagNames, ", "),
			record.Feedback,
			record.CustomFeedback,
			record.FeedbackTime,
			strconv.FormatInt(record.Threshold, 10),
		}

		for _, d := range xlsxData {
			cell := row.AddCell()
			cell.Value = d
		}
	}

	err = xlsxFile.Save(xlsxFilePath)
	return
}

func createExportRecordsTaskOption(query *dataV2.VisitRecordsQuery, exportTaskID string) *data.ExportTaskOption {
	index := fmt.Sprintf("%s-*", data.ESRecordsIndex)

	boolQuery := newBoolQueryWithRecordQuery(query)

	source := elastic.NewFetchSourceContext(true)
	source.Include(
		dataCommon.VisitRecordsMetricSessionID,
		dataCommon.VisitRecordsMetricTESessionID,
		dataCommon.VisitRecordsMetricUserID,
		dataCommon.VisitRecordsMetricUserQ,
		dataCommon.VisitRecordsMetricStdQ,
		"answer.value",
		dataCommon.VisitRecordsMetricScore,
		dataCommon.VisitRecordsMetricModule,
		dataCommon.VisitRecordsMetricLogTime,
		dataCommon.VisitRecordsMetricEmotion,
		dataCommon.VisitRecordsMetricEmotionScore,
		dataCommon.VisitRecordsMetricIntent,
		dataCommon.VisitRecordsMetricIntentScore,
		dataCommon.VisitRecordsMetricCustomInfo,
		dataCommon.VisitRecordsMetricSource,
		"faq_cat_id",
		"fat_robot_tag_id",
		dataCommon.VisitRecordsMetricFeedback,
		dataCommon.VisitRecordsMetricCustomFeedback,
		dataCommon.VisitRecordsMetricFeedbackTime,
		dataCommon.VisitRecordsMetricThreshold,
	)

	return &data.ExportTaskOption{
		TaskID:               exportTaskID,
		Index:                index,
		BoolQuery:            boolQuery,
		Source:               source,
		SortField:            data.LogTimeFieldName,
		ExtractResultHandler: extractExportRecordsHitResultHandler,
		XlsxCreateHandler:    createExportRecordsXlsx,
	}
}

// newBoolQueryWithRecordQuery create a *elastic.BoolQuery with the Record condition.
func newBoolQueryWithRecordQuery(query *dataV2.VisitRecordsQuery) *elastic.BoolQuery {
	boolQuery := elastic.NewBoolQuery()

	if query.AppID != "" {
		appTermQuery := elastic.NewTermQuery("app_id", query.AppID)
		boolQuery.Filter(appTermQuery)
	}

	// Specified record IDs
	if len(query.RecordIDs) > 0 {
		// Executing a Terms Query request with a lot of terms can be quite slow,
		// as each additional term demands extra processing and memory.
		// To safeguard against this, the maximum number of terms that can be used in
		// a Terms Query both directly or through lookup has been limited to 65536.
		boolQuery.Filter(elastic.NewTermsQuery("unique_id", query.RecordIDs...))
		return boolQuery
	}

	// Start time & End time
	rangeQuery := services.CreateRangeQueryUnixTime(query.CommonQuery,
		data.LogTimeFieldName)
	boolQuery.Filter(rangeQuery)

	// Modules
	if len(query.Modules) > 0 {
		modules := make([]interface{}, 0)
		for _, m := range query.Modules {
			modules = append(modules, m)
		}
		boolQuery.Filter(elastic.NewTermsQuery("module", modules...))
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

	// Emotions
	if len(query.Emotions) > 0 {
		emotions := make([]interface{}, 0)
		for _, e := range query.Emotions {
			emotions = append(emotions, e)
		}
		boolQuery.Filter(elastic.NewTermsQuery("emotion", emotions...))
	}

	// IsIgnored
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

	// IsMarked
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

	// Keyword
	if query.Keyword != nil && *query.Keyword != "" {
		userQMatchQuery := elastic.NewMatchQuery("user_q", *query.Keyword)
		answerMatchQuery := elastic.NewMatchQuery("answer.value", *query.Keyword)
		answerNestedQuery := elastic.NewNestedQuery("answer", answerMatchQuery)
		keywordBoolQuery := elastic.NewBoolQuery().Should(userQMatchQuery, answerNestedQuery)
		boolQuery = boolQuery.Filter(keywordBoolQuery)
	}

	// User ID
	if query.UserID != nil && *query.UserID != "" {
		boolQuery.Filter(elastic.NewTermQuery("user_id", *query.UserID))
	}

	// Session ID
	if query.SessionID != nil && *query.SessionID != "" {
		boolQuery.Filter(elastic.NewTermQuery("session_id", *query.SessionID))
	}

	// Task Engine Session ID
	if query.TESessionID != nil && *query.TESessionID != "" {
		boolQuery.Filter(elastic.NewTermQuery("taskengine_session_id",
			*query.TESessionID))
	}

	// Intent
	if query.Intent != nil && *query.Intent != "" {
		boolQuery.Filter(elastic.NewTermQuery("intent", *query.Intent))
	}

	// Min score & Max score
	if query.MinScore != nil && query.MaxScore != nil {
		minScore := *query.MinScore
		maxScore := *query.MaxScore

		scoreRangeQuery := elastic.NewRangeQuery("score").Gte(minScore).Lte(maxScore)
		boolQuery.Filter(scoreRangeQuery)
	}

	// Low confidence threshold
	if query.LowConfidence != nil {
		scoreExistsQuery := elastic.NewExistsQuery("score")
		thresholdExistsQuery := elastic.NewExistsQuery("threshold")

		paramName := "low_confidence"
		scriptSource := fmt.Sprintf("doc['score'].value - doc['threshold'].value <= params.%s",
			paramName)
		script := elastic.NewScript(scriptSource).Param(paramName, query.LowConfidence)
		lowConfidenceScriptQuery := elastic.NewScriptQuery(script)

		boolQuery.Must(scoreExistsQuery, thresholdExistsQuery)
		boolQuery.Filter(lowConfidenceScriptQuery)
	}

	// FAQ categories
	if len(query.FaqCategories) > 0 {
		faqCategories := make([]interface{}, 0)
		for _, c := range query.FaqCategories {
			faqCategories = append(faqCategories, c)
		}
		boolQuery.Filter(elastic.NewTermsQuery("faq_cat_id", faqCategories...))
	}

	// FAQ robot tags
	if len(query.FaqRobotTags) > 0 {
		faqRobotTags := make([]interface{}, 0)
		for _, t := range query.FaqRobotTags {
			faqRobotTags = append(faqRobotTags, t)
		}
		boolQuery.Filter(elastic.NewTermsQuery("faq_robot_tag_id", faqRobotTags...))
	}

	// Feedback
	if query.Feedback != nil && *query.Feedback != "" {
		feedbackTermQuery := elastic.NewTermQuery("feedback", *query.Feedback)
		customFeedbackTermQuery := elastic.NewTermQuery("custom_feedback.keyword",
			*query.Feedback)

		boolQuery.Filter(elastic.NewBoolQuery().Should(feedbackTermQuery, customFeedbackTermQuery))
	}

	return boolQuery
}
