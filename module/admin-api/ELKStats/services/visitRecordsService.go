package services

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"emotibot.com/emotigo/module/admin-api/ELKStats/dao"
	"emotibot.com/emotigo/module/admin-api/ELKStats/data"
	"emotibot.com/emotigo/module/admin-api/util"
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

const (
	csvExportDir           = "exports"
	csvDirTimestampFormat  = "20060102"
	csvFileTimestampFormat = "20060102_150405"
)

var exportBaseDir string

func VisitRecordsServiceInit() error {
	err := createExportRecordsBaseDir()
	if err != nil {
		return err
	}

	// Unlock all enterprises export task
	// to prevent any enterprise locked due to the crash of admin-api
	err = dao.UnlockAllEnterprisesExportTask()
	if err != nil {
		return err
	}

	// Start a housekeeping goroutine to clean up outdated exported records
	go exportedRecordsHousekeeping()

	return nil
}

// newBoolQueryWithRecordQuery create a *elastic.BoolQuery with the Record condition.
func newBoolQueryWithRecordQuery(query *data.RecordQuery) *elastic.BoolQuery {
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
	boolQuery := newBoolQueryWithRecordQuery(&query)
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
		data.VisitRecordsMetricModule,
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

func VisitRecordsExport(query *data.RecordQuery) (exportTaskID string, err error) {
	// Try to create export task
	exportTaskID, err = dao.TryCreateExportTask(query.EnterpriseID)
	if err != nil {
		return
	}

	// Create a goroutine to exporting records in background
	go exportTask(query, exportTaskID)
	return
}

func VisitRecordsExportDownload(w http.ResponseWriter, exportTaskID string) error {
	exists, err := dao.ExportRecordsExists(exportTaskID)
	if !exists {
		return data.ErrExportTaskNotFound
	} else if err != nil {
		return err
	}

	statusCode, err := dao.GetExportTaskStatus(exportTaskID)
	if err != nil {
		return err
	}

	switch statusCode {
	case data.CodeExportTaskRunning:
		return data.ErrExportTaskInProcess
	case data.CodeExportTaskFailed:
		return data.ErrExportTaskFailed
	case data.CodeExportTaskEmpty:
		return data.ErrExportTaskEmpty
	}

	filePath, err := dao.GetExportRecordsFile(exportTaskID)
	if err != nil {
		return err
	}

	switch path.Ext(filePath) {
	case ".zip":
		w.Header().Set("Content-type", "application/zip")
	case ".csv":
		w.Header().Set("Content-type", "text/csv;charset=utf-8")
	default:
		return fmt.Errorf("Invalid exported records file format: %s", filePath)
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment;filename=%s", path.Base(filePath)))

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}

	_, err = io.Copy(w, file)
	if err != nil {
		return err
	}

	return nil
}

func VisitRecordsExportDelete(exportTaskID string) error {
	exists, err := dao.ExportRecordsExists(exportTaskID)
	if !exists {
		return data.ErrExportTaskNotFound
	} else if err != nil {
		return err
	}

	return dao.DeleteExportTask(exportTaskID)
}

func VisitRecordsExportStatus(exportTaskID string) (status string, err error) {
	exists, _err := dao.ExportRecordsExists(exportTaskID)
	if !exists {
		err = data.ErrExportTaskNotFound
		return
	} else if _err != nil {
		err = _err
		return
	}

	statusCode, _err := dao.GetExportTaskStatus(exportTaskID)
	if _err != nil {
		err = _err
		return
	}

	status = data.ExportTaskCodesMap[statusCode]
	return
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

// exportTask exports records to CSV file in background
// TODO: Cancel the corresponding running exporting records task when:
//		 DELETE /export/{export-id} API is called.
func exportTask(query *data.RecordQuery, exportTaskID string) {
	var err error
	defer func() {
		if err != nil {
			logger.Error.Printf("Something goes wrong with exporting task: %s, error: %s\n",
				exportTaskID, err.Error())
			dao.ExportTaskFailed(exportTaskID, err.Error())
		}
	}()

	ctx, client := elasticsearch.GetClient()

	logger.Info.Printf("Start exporting task: %s\n", exportTaskID)

	boolQuery := newBoolQueryWithRecordQuery(query)

	source := elastic.NewFetchSourceContext(true)
	source.Include(
		data.VisitRecordsMetricSessionID,
		data.VisitRecordsMetricAppID,
		data.VisitRecordsMetricUserID,
		data.VisitRecordsMetricUserQ,
		data.VisitRecordsMetricStdQ,
		"answer.value",
		data.VisitRecordsMetricModule,
		data.VisitRecordsMetricEmotion,
		data.VisitRecordsMetricEmotionScore,
		data.VisitRecordsMetricIntent,
		data.VisitRecordsMetricIntentScore,
		data.VisitRecordsMetricLogTime,
		data.VisitRecordsMetricScore,
		data.VisitRecordsMetricCustomInfo,
		data.VisitRecordsMetricNote,
		data.VisitRecordsMetricQType,
	)

	index := fmt.Sprintf("%s-%s-*", data.ESRecordsIndex, query.AppID)

	service := client.Scroll().
		Index(index).
		Query(boolQuery).
		Size(data.ESScrollSize).
		FetchSourceContext(source).
		Sort(data.LogTimeFieldName, false).
		KeepAlive(data.ESScrollKeepAlive)

	records := make([]*data.VisitRecordsExportData, 0)
	csvFilePaths := make([]string, 0)
	numOfRecords := 0
	numOfRecordsPerCSV := 0
	timestamp := time.Now().Format(csvFileTimestampFormat)
	numOfCSVFiles := 0

	for {
		logger.Trace.Printf("Task %s: Fetching results..\n", exportTaskID)
		results, _err := service.Do(ctx)
		logger.Trace.Printf("Task %s: Fetch results completed!!\n", exportTaskID)
		if _err != nil {
			// No more data, scroll finished
			if _err == io.EOF {
				// Flush the remaining records in the buffer to CSV file, if any
				if len(records) > 0 {
					numOfCSVFiles++
					csvFileName := fmt.Sprintf("%s_%d", timestamp, numOfCSVFiles)
					logger.Info.Printf("Task %s: Create CSV file %s.csv\n", exportTaskID, csvFileName)
					csvFilePath, _err := createExportRecordsCSV(records, csvFileName)
					if _err != nil {
						err = _err
						return
					}
					csvFilePaths = append(csvFilePaths, csvFilePath)
				}

				if len(csvFilePaths) == 0 {
					err = dao.ExportTaskEmpty(exportTaskID)
					if err == nil {
						logger.Info.Printf("Task %s: Exporting completed, no results\n", exportTaskID)
					}
				} else if len(csvFilePaths) == 1 {
					err = dao.ExportTaskCompleted(exportTaskID, csvFilePaths[0])
					if err == nil {
						logger.Info.Printf("Task %s: Exporting completed, total %d records\n",
							exportTaskID, numOfRecords)
					}
				} else {
					// Multiple CSV files, zip the files
					logger.Info.Printf("Task %s: Start compressing %d CSV files\n",
						exportTaskID, len(csvFilePaths))

					zipFilePath, _err := compressExportRecordsCSV(csvFilePaths, timestamp)
					if _err != nil {
						err = _err
						return
					}

					logger.Info.Printf("Task %s: Compresion finished!!\n", exportTaskID)

					err = dao.ExportTaskCompleted(exportTaskID, zipFilePath)
					if err != nil {
						return
					}

					// Delete CSV files
					for _, csvFilePath := range csvFilePaths {
						err = os.Remove(csvFilePath)
						if err != nil {
							return
						}
					}

					logger.Info.Printf("Task %s: Exporting completed, total %d records\n",
						exportTaskID, numOfRecords)
				}

				return
			}

			err = _err
			return
		}

		logger.Trace.Printf("Task %s: Start converting %d hit results\n",
			exportTaskID, len(results.Hits.Hits))

		for _, hit := range results.Hits.Hits {
			rawRecord := data.VisitRecordsExportRawData{}
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

			customInfo, _err := json.Marshal(rawRecord.CustomInfo)
			if _err != nil {
				err = _err
				return
			}

			record := data.VisitRecordsExportData{
				VisitRecordsExportBase: data.VisitRecordsExportBase{
					SessionID:    rawRecord.SessionID,
					AppID:        rawRecord.AppID,
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
					Note:         rawRecord.Note,
				},
				Answer:     strings.Join(answers, ", "),
				CustomInfo: string(customInfo),
			}

			records = append(records, &record)
			numOfRecords++
			numOfRecordsPerCSV++
		}

		logger.Trace.Printf("Task %s: Converting %d hit results finished!!\n",
			exportTaskID, len(results.Hits.Hits))

		if numOfRecordsPerCSV == data.MaxNumRecordsPerCSV {
			numOfCSVFiles++
			csvFileName := fmt.Sprintf("%s_%d", timestamp, numOfCSVFiles)
			logger.Info.Printf("Task %s: Create CSV file %s.csv\n", exportTaskID, csvFileName)
			csvFilePath, _err := createExportRecordsCSV(records, csvFileName)
			if _err != nil {
				err = _err
				return
			}
			csvFilePaths = append(csvFilePaths, csvFilePath)

			records = make([]*data.VisitRecordsExportData, 0)
			numOfRecordsPerCSV = 0
		}
	}
}

func createExportRecordsCSV(records []*data.VisitRecordsExportData,
	csvFileName string) (csvFilePath string, err error) {
	dirPath, _err := getExportRecordsDir()
	if _err != nil {
		err = _err
		return
	}

	csvFilePath = fmt.Sprintf("%s/%s.csv", dirPath, csvFileName)
	csvFile, _err := os.Create(csvFilePath)
	if _err != nil {
		err = _err
		return
	}
	defer csvFile.Close()

	// FIXME: ADD windows byte mark for utf-8 support on old EXCEL
	out := csv.NewWriter(csvFile)
	out.Write(data.VisitRecordsExportHeader)

	for _, record := range records {
		csvData := []string{
			record.SessionID,
			record.UserID,
			record.UserQ,
			record.StdQ,
			record.Answer,
			strconv.FormatFloat(record.Score, 'f', -1, 64),
			record.Module,
			record.LogTime,
			record.Emotion,
			strconv.FormatFloat(record.EmotionScore, 'f', -1, 64),
			record.Intent,
			strconv.FormatFloat(record.IntentScore, 'f', -1, 64),
			record.CustomInfo,
			record.Note,
		}

		err = out.Write(csvData)
	}

	out.Flush()
	err = out.Error()
	return
}

func compressExportRecordsCSV(files []string, timestamp string) (zipFilePath string, err error) {
	dirPath, _err := getExportRecordsDir()
	if _err != nil {
		err = _err
		return
	}

	zipFilePath = fmt.Sprintf("%s/%s.zip", dirPath, timestamp)
	err = util.CompressFiles(files, zipFilePath)
	return
}

func createExportRecordsBaseDir() error {
	curDir, err := util.GetCurDir()
	if err != nil {
		return err
	}

	exportBaseDir = fmt.Sprintf("%s/%s", curDir, csvExportDir)
	return os.MkdirAll(exportBaseDir, 0755)
}

/* getExportRecordsDir returns the directory path of the exported records
 * If the directory not existed, create one
 */
func getExportRecordsDir() (dirPath string, err error) {
	dirName := time.Now().Format(csvDirTimestampFormat)
	dirPath = fmt.Sprintf("%s/%s", exportBaseDir, dirName)
	err = os.MkdirAll(dirPath, 0755)
	return
}

/* exportedRecordsHousekeeping do the housekeeping everyday
 * to removes all the outdated exported records,
 * including files and database rows everyday
 * Only output logs when encountering any errors
 */
func exportedRecordsHousekeeping() {
	for {
		now := time.Now()
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
		todayStr := today.Format("2006-01-02")

		logger.Info.Printf("Start doing housekeeping of %s\n", todayStr)
		logger.Info.Printf("Removing exported records before %s\n", todayStr)

		// Remove all outdated exported records directories
		files, err := filepath.Glob(fmt.Sprintf("%s/*", exportBaseDir))
		if err != nil {
			logger.Error.Println(err.Error())
		} else {
			for _, fPath := range files {
				fInfo, err := os.Stat(fPath)
				if err != nil {
					logger.Error.Println(err.Error())
					continue
				}

				if fInfo.IsDir() {
					dirTimestamp, err := time.Parse(csvDirTimestampFormat, fInfo.Name())
					if err != nil {
						logger.Error.Println(err.Error())
						continue
					}

					if dirTimestamp.Before(today) {
						err = os.RemoveAll(fPath)
						if err != nil {
							logger.Error.Println(err.Error())
							continue
						}
					}
				}
			}
		}

		// Remove all outdated database rows
		err = dao.RemoveAllOutdatedExportTasks(today.Unix())
		if err != nil {
			logger.Error.Println(err.Error())
		}

		logger.Info.Printf("Housekeeping of %s has completed, see you tomorrow!!\n", todayStr)

		// Too tired, sleep for a day for another exhausted housekeeping
		time.Sleep(24 * time.Hour)
	}
}
