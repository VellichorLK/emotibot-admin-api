package v1

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
	dataV1 "emotibot.com/emotigo/module/admin-api/ELKStats/data/v1"
	servicesCommon "emotibot.com/emotigo/module/admin-api/ELKStats/services/common"
	"emotibot.com/emotigo/module/admin-api/util/elasticsearch"
	"emotibot.com/emotigo/pkg/logger"
	"github.com/olivere/elastic"
	"github.com/tealeg/xlsx"
)

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

//RecordResult is the result fetched from Datastore,
//which will include the Hits raw data and aggregation static data
//Aggs always contain total_size, and may contain stats info like total_marked_size ...
type RecordResult struct {
	Hits []*dataV1.VisitRecordsData
	Aggs map[string]interface{}
}

//VisitRecordsQuery will fetch raw and aggregation data from Data store.
//input query decide which raw data to fetchand & aggs will decided Result's aggs
//In the return, result.Aggs may contain isMarked or isIgnored key if correspond ElasticSearchCommand have bee given in aggs
func VisitRecordsQuery(query dataV1.RecordQuery, aggs ...servicesCommon.ElasticSearchCommand) (*RecordResult, error) {
	ctx, client := elasticsearch.GetClient()
	index := fmt.Sprintf("%s-*", data.ESRecordsIndex)
	logger.Trace.Printf("index: %s\n", index)
	boolQuery := newBoolQueryWithRecordQuery(&query)
	traceQuery(boolQuery)
	source := elastic.NewFetchSourceContext(true)
	source.Include(
		dataCommon.VisitRecordsMetricSessionID,
		dataCommon.VisitRecordsMetricUserID,
		dataCommon.VisitRecordsMetricUserQ,
		dataCommon.VisitRecordsMetricScore,
		dataCommon.VisitRecordsMetricStdQ,
		"answer.value",
		dataCommon.VisitRecordsMetricLogTime,
		dataCommon.VisitRecordsMetricEmotion,
		dataCommon.VisitRecordsMetricModule,
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
		Hits: make([]*dataV1.VisitRecordsData, 0),
		Aggs: map[string]interface{}{
			"total_size": result.TotalHits(),
		},
	}

	if markedSize, found := result.Aggregations.Filter("isMarked"); found {
		r.Aggs["isMarked"] = markedSize.DocCount
	} else {
		r.Aggs["isMarked"] = 0
	}

	if ignoredSize, found := result.Aggregations.Filter("isIgnored"); found {
		r.Aggs["isIgnored"] = ignoredSize.DocCount
	} else {
		r.Aggs["isIgnored"] = 0
	}

	for _, hit := range result.Hits.Hits {
		hitResult := dataV1.VisitRecordsHitResult{}
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

		record := &dataV1.VisitRecordsData{
			VisitRecordsDataBase: dataV1.VisitRecordsDataBase{
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

func VisitRecordsExport(query *dataV1.RecordQuery) (exportTaskID string, err error) {
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
func UpdateRecords(query dataV1.RecordQuery, cmd servicesCommon.UpdateCommand) error {
	ctx, client := elasticsearch.GetClient()
	index := fmt.Sprintf("%s-*", data.ESRecordsIndex)
	bq := newBoolQueryWithRecordQuery(&query)
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

// newBoolQueryWithRecordQuery create a *elastic.BoolQuery with the Record condition.
func newBoolQueryWithRecordQuery(query *dataV1.RecordQuery) *elastic.BoolQuery {
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

func extractExportRecordsHitResultHandler(hit *elastic.SearchHit) (recordPtr interface{}, err error) {
	rawRecord := dataV1.VisitRecordsExportRawData{}
	jsonStr, _err := hit.Source.MarshalJSON()
	if _err != nil {
		err = _err
		return
	}

	err = json.Unmarshal(jsonStr, &rawRecord)
	if err != nil {
		return
	}

	// Convert log time from UTC+0 back to local time
	logTime, _err := time.Parse(data.LogTimeFormat, rawRecord.LogTime)
	if _err != nil {
		err = _err
		return
	}
	logTime = logTime.Local()

	answers := make([]string, 0)

	for _, answer := range rawRecord.Answer {
		answers = append(answers, answer.Value)
	}

	customInfoString := ""

	if rawRecord.CustomInfo != nil {
		customInfo, _err := json.Marshal(rawRecord.CustomInfo)
		if _err != nil {
			err = _err
			return
		}

		customInfoString = string(customInfo)
		if customInfoString == "{}" {
			customInfoString = ""
		}
	}

	recordPtr = &dataV1.VisitRecordsExportData{
		VisitRecordsExportBase: dataV1.VisitRecordsExportBase{
			SessionID:    rawRecord.SessionID,
			UserID:       rawRecord.UserID,
			UserQ:        rawRecord.UserQ,
			StdQ:         rawRecord.StdQ,
			Module:       rawRecord.Module,
			Emotion:      rawRecord.Emotion,
			EmotionScore: rawRecord.EmotionScore,
			Intent:       rawRecord.Intent,
			IntentScore:  rawRecord.IntentScore,
			LogTime:      logTime.Format(data.LogTimeFormat),
			Score:        rawRecord.Score,
			Source:       rawRecord.Source,
		},
		Answer:     strings.Join(answers, ", "),
		CustomInfo: customInfoString,
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

	// Header row
	row := sheet.AddRow()
	for _, header := range dataV1.VisitRecordsExportHeader {
		cell := row.AddCell()
		cell.Value = header
	}

	// Data rows
	for _, recordPtr := range recordPtrs {
		record := recordPtr.(*dataV1.VisitRecordsExportData)

		row := sheet.AddRow()
		xlsxData := []string{
			record.SessionID,
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
		}

		for _, d := range xlsxData {
			cell := row.AddCell()
			cell.Value = d
		}
	}

	err = xlsxFile.Save(xlsxFilePath)
	return
}

func createExportRecordsTaskOption(query *dataV1.RecordQuery, exportTaskID string) *data.ExportTaskOption {
	index := fmt.Sprintf("%s-*", data.ESRecordsIndex)

	boolQuery := newBoolQueryWithRecordQuery(query)

	source := elastic.NewFetchSourceContext(true)
	source.Include(
		dataCommon.VisitRecordsMetricSessionID,
		dataCommon.VisitRecordsMetricUserID,
		dataCommon.VisitRecordsMetricUserQ,
		dataCommon.VisitRecordsMetricStdQ,
		"answer.value",
		dataCommon.VisitRecordsMetricModule,
		dataCommon.VisitRecordsMetricEmotion,
		dataCommon.VisitRecordsMetricEmotionScore,
		dataCommon.VisitRecordsMetricIntent,
		dataCommon.VisitRecordsMetricIntentScore,
		dataCommon.VisitRecordsMetricLogTime,
		dataCommon.VisitRecordsMetricScore,
		dataCommon.VisitRecordsMetricSource,
		dataCommon.VisitRecordsMetricCustomInfo,
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

func traceQuery(bq *elastic.BoolQuery) {
	src, _ := bq.Source()
	d, _ := json.Marshal(src)
	logger.Trace.Printf("query: %s\n", d)
}
