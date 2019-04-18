package v1

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"emotibot.com/emotigo/module/admin-api/ELKStats/dao"
	"emotibot.com/emotigo/module/admin-api/ELKStats/data"
	dataCommon "emotibot.com/emotigo/module/admin-api/ELKStats/data/common"
	dataV1 "emotibot.com/emotigo/module/admin-api/ELKStats/data/v1"
	servicesCommon "emotibot.com/emotigo/module/admin-api/ELKStats/services/common"
	"emotibot.com/emotigo/module/admin-api/util/elasticsearch"
	"github.com/olivere/elastic"
	"github.com/tealeg/xlsx"
)

func SessionsQuery(query *dataV1.SessionsQuery) (sessions []*dataV1.SessionsData, totalSize int64, err error) {
	ctx, client := elasticsearch.GetClient()
	boolQuery := newSessionBoolQuery(query)

	source := elastic.NewFetchSourceContext(true)
	source.Include(
		dataCommon.SessionsMetricSessionID,
		dataCommon.SessionsMetricStartTime,
		dataCommon.SessionsMetricEndTime,
		dataCommon.SessionsMetricUserID,
		dataCommon.SessionsMetricRating,
		dataCommon.SessionsMetricCustomInfo,
		dataCommon.SessionsMetricFeedback,
		dataCommon.SessionsMetricCustomFeedback,
		dataCommon.SessionsMetricFeedbackTime,
	)

	index := fmt.Sprintf("%s-*", data.ESSessionsIndex)

	results, err := client.Search().
		Index(index).
		Type(data.ESSessionsType).
		Query(boolQuery).
		FetchSourceContext(source).
		From(int(query.From)).
		Size(int(query.Limit)).
		Sort(data.SessionStartTimeFieldName, false).
		Do(ctx)
	if err != nil {
		return nil, 0, err
	}

	sessions = make([]*dataV1.SessionsData, 0)
	totalSize = results.Hits.TotalHits

	for _, hit := range results.Hits.Hits {
		rawSession := dataV1.SessionsRawData{}
		jsonStr, err := hit.Source.MarshalJSON()
		if err != nil {
			return nil, 0, err
		}

		err = json.Unmarshal(jsonStr, &rawSession)
		if err != nil {
			return nil, 0, err
		}

		sessionCommon, err := extractRawSession(&rawSession)
		if err != nil {
			return nil, 0, err
		}

		session := &dataV1.SessionsData{
			SessionsCommon: *sessionCommon,
			CustomInfo:     rawSession.CustomInfo,
		}

		sessions = append(sessions, session)
	}

	return
}

func SessionsExport(query *dataV1.SessionsQuery, locale string) (exportTaskID string, err error) {
	// Try to create export task
	exportTaskID, err = dao.TryCreateExportTask(query.EnterpriseID)
	if err != nil {
		return
	}

	// Create a goroutine to exporting records in background
	go func() {
		option := createExportSessionsTaskOption(query, exportTaskID, locale)
		servicesCommon.ExportTask(option, locale)
	}()

	return
}

func SessionsExportDownload(w http.ResponseWriter, exportTaskID string) error {
	return servicesCommon.DownloadExportRecords(w, exportTaskID)
}

func SessionsExportDelete(exportTaskID string) error {
	return servicesCommon.DeleteExportRecords(exportTaskID)
}

func SessionsExportStatus(exportTaskID string) (status string, err error) {
	status, err = servicesCommon.GetRecordsExportStatus(exportTaskID)
	return
}

// newSessionBoolQuery create a *elastic.BoolQuery
func newSessionBoolQuery(query *dataV1.SessionsQuery) *elastic.BoolQuery {
	boolQuery := elastic.NewBoolQuery()

	if query.AppID != "" {
		appTermQuery := elastic.NewTermQuery("app_id", query.AppID)
		boolQuery.Filter(appTermQuery)
	}

	// Start time & End time
	startTimeQuery := elastic.NewRangeQuery(data.SessionStartTimeFieldName).
		Gte(query.StartTime.Unix()).
		Format("epoch_second")
	endTimeQuery := elastic.NewRangeQuery(data.SessionEndTimeFieldName).
		Lte(query.EndTime.Unix()).
		Format("epoch_second")
	boolQuery.Filter(startTimeQuery)
	boolQuery.Filter(endTimeQuery)

	// Feedback start time
	if query.FeedbackStartTime != nil {
		feedbackStartTime := query.FeedbackStartTime

		feedbackStartTimeRangeQuery := elastic.NewRangeQuery(data.SessionFeedbackTimeFieldName).
			Gte(feedbackStartTime.Unix()).
			Format("epoch_second")
		boolQuery.Filter(feedbackStartTimeRangeQuery)
	}

	// Feedback end time
	if query.FeedbackEndTime != nil {
		feedbackEndTime := query.FeedbackEndTime

		feedbackEndTimeRangeQuery := elastic.NewRangeQuery(data.SessionFeedbackTimeFieldName).
			Lte(feedbackEndTime.Unix()).
			Format("epoch_second")
		boolQuery.Filter(feedbackEndTimeRangeQuery)
	}

	// User ID
	if query.UserID != nil && *query.UserID != "" {
		boolQuery.Filter(elastic.NewTermQuery("user_id", *query.UserID))
	}

	// Session ID
	if query.SessionID != nil && *query.SessionID != "" {
		boolQuery.Filter(elastic.NewTermQuery("session_id", *query.SessionID))
	}

	// Rating max & Rating min
	if query.RatingMax != nil && query.RatingMin != nil {
		rateRangeQuery := elastic.NewRangeQuery(data.SessionRatingFieldName).
			Gte(*query.RatingMin).
			Lte(*query.RatingMax)
		boolQuery.Filter(rateRangeQuery)
	}

	// Platforms & Sex
	tags := make([]data.QueryTags, 0)

	if len(query.Platforms) > 0 {
		tags = append(tags, data.QueryTags{
			Type:  "platform",
			Texts: query.Platforms,
		})
	}

	if len(query.Sex) > 0 {
		tags = append(tags, data.QueryTags{
			Type:  "sex",
			Texts: query.Sex,
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

	// Feedback
	if query.Feedback != nil && *query.Feedback != "" {
		feedbackTermQuery := elastic.NewTermQuery("feedback", *query.Feedback)
		customFeedbackTermQuery := elastic.NewTermQuery("custom_feedback.keyword",
			*query.Feedback)

		boolQuery.Filter(elastic.NewBoolQuery().Should(feedbackTermQuery, customFeedbackTermQuery))
	}

	return boolQuery
}

func extractRawSession(rawSession *dataV1.SessionsRawData) (*dataV1.SessionsCommon, error) {
	session := dataV1.SessionsCommon{
		SessionsDataBase: dataV1.SessionsDataBase{
			SessionID:      rawSession.SessionID,
			UserID:         rawSession.UserID,
			Rating:         rawSession.Rating,
			Feedback:       rawSession.Feedback,
			CustomFeedback: rawSession.CustomFeedback,
		},
	}

	// Convert start time, end time and feedback time
	// from UTC+0 back to local time
	var startTime string
	if rawSession.StartTime != 0 {
		startTime = time.Unix(rawSession.StartTime, 0).Local().
			Format(data.StandardTimeFormat)
	}

	var endTime string
	if rawSession.EndTime != 0 {
		endTime = time.Unix(rawSession.EndTime, 0).Local().
			Format(data.StandardTimeFormat)
	}

	var feedbackTime string
	if rawSession.FeedbackTime != 0 {
		feedbackTime = time.Unix(rawSession.FeedbackTime, 0).Local().
			Format(data.StandardTimeFormat)
	}

	session.StartTime = startTime
	session.EndTime = endTime
	session.FeedbackTime = feedbackTime

	return &session, nil
}

func extractExportSessionsHitResultHandler(hit *elastic.SearchHit) (sessionPtr interface{}, err error) {
	rawSession := dataV1.SessionsRawData{}
	jsonStr, err := hit.Source.MarshalJSON()
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(jsonStr, &rawSession)
	if err != nil {
		return
	}

	sessionCommon, err := extractRawSession(&rawSession)
	if err != nil {
		return nil, err
	}

	// Custom info
	var customInfo string
	if rawSession.CustomInfo != nil {
		json, err := json.Marshal(rawSession.CustomInfo)
		if err != nil {
			return nil, err
		}

		customInfo = string(json)
	}

	sessionPtr = &dataV1.SessionsExportData{
		SessionsCommon: *sessionCommon,
		CustomInfo:     customInfo,
	}

	return
}

func createExportSessionsXlsx(sessionPtrs []interface{}, xlsxFileName string, locale string,
	params ...interface{}) (xlsxFilePath string, err error) {
	dirPath, _err := servicesCommon.GetExportRecordsDir()
	if _err != nil {
		err = _err
		return
	}

	xlsxFilePath = fmt.Sprintf("%s/%s.xlsx", dirPath, xlsxFileName)
	xlsxFile := xlsx.NewFile()
	sheet, _err := xlsxFile.AddSheet("会话日志导出数据")
	if err != nil {
		err = _err
		return
	}

	// Header row
	row := sheet.AddRow()
	for _, header := range dataV1.GetSessionsExportHeader(locale) {
		cell := row.AddCell()
		cell.Value = header
	}

	// Data rows
	for _, sessionPtr := range sessionPtrs {
		session := sessionPtr.(*dataV1.SessionsExportData)

		row := sheet.AddRow()
		ratingStr := ""
		if session.Rating > 0 {
			ratingStr = strconv.FormatInt(session.Rating, 10)
		}
		xlsxData := []string{
			session.SessionID,
			session.StartTime,
			session.EndTime,
			session.UserID,
			ratingStr,
			session.CustomInfo,
			session.Feedback,
			session.CustomFeedback,
			session.FeedbackTime,
		}

		for _, d := range xlsxData {
			cell := row.AddCell()
			cell.Value = d
		}
	}

	err = xlsxFile.Save(xlsxFilePath)
	return
}

func createExportSessionsTaskOption(query *dataV1.SessionsQuery,
	exportTaskID string, locale string) *data.ExportTaskOption {
	index := fmt.Sprintf("%s-*", data.ESSessionsIndex)

	boolQuery := newSessionBoolQuery(query)

	source := elastic.NewFetchSourceContext(true)
	source.Include(
		dataCommon.SessionsMetricSessionID,
		dataCommon.SessionsMetricStartTime,
		dataCommon.SessionsMetricEndTime,
		dataCommon.SessionsMetricUserID,
		dataCommon.SessionsMetricRating,
		dataCommon.SessionsMetricCustomInfo,
		dataCommon.SessionsMetricFeedback,
		dataCommon.SessionsMetricCustomFeedback,
		dataCommon.SessionsMetricFeedbackTime,
	)

	return &data.ExportTaskOption{
		TaskID:               exportTaskID,
		Index:                index,
		BoolQuery:            boolQuery,
		Source:               source,
		SortField:            data.SessionStartTimeFieldName,
		ExtractResultHandler: extractExportSessionsHitResultHandler,
		XlsxCreateHandler:    createExportSessionsXlsx,
	}
}
