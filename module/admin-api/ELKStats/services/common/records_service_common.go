package common

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"emotibot.com/emotigo/module/admin-api/util/localemsg"

	"emotibot.com/emotigo/module/admin-api/ELKStats/dao"
	"emotibot.com/emotigo/module/admin-api/ELKStats/data"
	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/admin-api/util/elasticsearch"
	"emotibot.com/emotigo/pkg/logger"
	elastic "gopkg.in/olivere/elastic.v6"
)

const (
	csvExportDir           = "exports"
	csvDirTimestampFormat  = "20060102"
	csvFileTimestampFormat = "20060102_150405"
)

var exportBaseDir string

func RecordsServiceInit() error {
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

// CreateTagsBoolQueries converts queryTags = [
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
func CreateTagsBoolQueries(tagsBoolQueries []*elastic.BoolQuery,
	queryTags []data.QueryTags, queryTagsIndex int,
	tags []data.QueryTag) []*elastic.BoolQuery {
	if queryTagsIndex != len(queryTags)-1 {
		for _, tagText := range queryTags[queryTagsIndex].Texts {
			tag := data.QueryTag{
				Type: queryTags[queryTagsIndex].Type,
				Text: tagText,
			}
			_tags := make([]data.QueryTag, len(tags)+1)
			copy(_tags, tags)
			_tags[len(tags)] = tag

			boolQueries := CreateTagsBoolQueries(tagsBoolQueries, queryTags, queryTagsIndex+1, _tags)
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

		return boolQueries
	}

	return tagsBoolQueries
}

func compressExportRecordsXLSX(files []string, timestamp string) (zipFilePath string, err error) {
	dirPath, _err := GetExportRecordsDir()
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

	exportBaseDir = fmt.Sprintf("%s/%s", curDir, data.XlsxExportDir)
	return os.MkdirAll(exportBaseDir, 0755)
}

/* getExportRecordsDir returns the directory path of the exported records
 * the directory not existed, create one
 */
func GetExportRecordsDir() (dirPath string, err error) {
	dirName := time.Now().Format(data.XlsxDirTimestampFormat)
	dirPath = fmt.Sprintf("%s/%s", exportBaseDir, dirName)
	err = os.MkdirAll(dirPath, 0755)
	return
}

func DownloadExportRecords(w http.ResponseWriter, exportTaskID string) error {
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
	case ".xlsx":
		w.Header().Set("Content-type",
			"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
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

func DeleteExportRecords(exportTaskID string) error {
	exists, err := dao.ExportRecordsExists(exportTaskID)
	if !exists {
		return data.ErrExportTaskNotFound
	} else if err != nil {
		return err
	}

	return dao.DeleteExportTask(exportTaskID)
}

func GetRecordsExportStatus(exportTaskID string) (status string, err error) {
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

// ExportTask exports records to Excel file in background
// TODO: Cancel the corresponding running exporting records task when:
//		 DELETE /export/{export-id} API is called.
func ExportTask(option *data.ExportTaskOption, locale string) {
	var err error
	defer func() {
		if err != nil {
			var errMessage string
			if rootCauseErrors, ok := elasticsearch.ExtractElasticsearchRootCauseErrors(err); ok {
				errMessage = strings.Join(rootCauseErrors, "; ")
			} else {
				errMessage = err.Error()
			}

			logger.Error.Printf("Something goes wrong with exporting task: %s, error: %s\n",
				option.TaskID, errMessage)
			dao.ExportTaskFailed(option.TaskID, errMessage)
		}
	}()

	ctx, client := elasticsearch.GetClient()

	logger.Info.Printf("Start exporting task: %s...\n", option.TaskID)

	service := client.Scroll().
		Index(option.Index).
		Query(option.BoolQuery).
		Size(data.ESScrollSize).
		FetchSourceContext(option.Source).
		Sort(option.SortField, false).
		KeepAlive(data.ESScrollKeepAlive)

	records := make([]interface{}, 0)
	xlsxFilePaths := make([]string, 0)
	numOfRecords := 0
	numOfRecordsPerXlsx := 0
	timestamp := time.Now().Format(data.XlsxFileTimestampFormat)
	numOfXlsxFiles := 0

	var localeStr string = localemsg.ZhCn
	if locale != "" {
		localeStr = locale
	}

	for {
		logger.Trace.Printf("Task %s: Fetching results..\n", option.TaskID)
		results, _err := service.Do(ctx)
		logger.Trace.Printf("Task %s: Fetch results completed!!\n", option.TaskID)
		if _err != nil {
			// No more data, scroll finished
			if _err == io.EOF {
				// Flush the remaining records in the buffer to Excel file, if any
				if len(records) > 0 {
					numOfXlsxFiles++
					xlsxFileName := fmt.Sprintf("%s_%d", timestamp, numOfXlsxFiles)
					logger.Info.Printf("Task %s: Create Excel file %s.xlsx\n", option.TaskID, xlsxFileName)
					xlsxFilePath, _err := option.XlsxCreateHandler(records, xlsxFileName, localeStr, option.AppID)
					if _err != nil {
						err = _err
						return
					}
					xlsxFilePaths = append(xlsxFilePaths, xlsxFilePath)
				}

				if len(xlsxFilePaths) == 0 {
					err = dao.ExportTaskEmpty(option.TaskID)
					if err == nil {
						logger.Info.Printf("Task %s: Exporting completed, no results\n", option.TaskID)
					}
				} else if len(xlsxFilePaths) == 1 {
					err = dao.ExportTaskCompleted(option.TaskID, xlsxFilePaths[0])
					if err == nil {
						logger.Info.Printf("Task %s: Exporting completed, total %d records\n",
							option.TaskID, numOfRecords)
					}
				} else {
					// Multiple Excel files, zip the files
					logger.Info.Printf("Task %s: Start compressing %d Excel files...\n",
						option.TaskID, len(xlsxFilePaths))

					zipFilePath, _err := compressExportRecordsXLSX(xlsxFilePaths, timestamp)
					if _err != nil {
						err = _err
						return
					}

					logger.Info.Printf("Task %s: Compresion finished!!\n", option.TaskID)

					err = dao.ExportTaskCompleted(option.TaskID, zipFilePath)
					if err != nil {
						return
					}

					// Delete Excel files
					for _, xlsxFilePath := range xlsxFilePaths {
						err = os.Remove(xlsxFilePath)
						if err != nil {
							return
						}
					}

					logger.Info.Printf("Task %s: Exporting completed, total %d records\n",
						option.TaskID, numOfRecords)
				}

				return
			}

			err = _err
			return
		}

		logger.Trace.Printf("Task %s: Start converting %d hit results...\n",
			option.TaskID, len(results.Hits.Hits))

		for _, hit := range results.Hits.Hits {
			record, _err := option.ExtractResultHandler(hit)
			if err != nil {
				err = _err
				return
			}

			records = append(records, record)
			numOfRecords++
			numOfRecordsPerXlsx++
		}

		logger.Trace.Printf("Task %s: Converting %d hit results finished!!\n",
			option.TaskID, len(results.Hits.Hits))

		if numOfRecordsPerXlsx == data.MaxNumRecordsPerXlsx {
			numOfXlsxFiles++
			xlsxFileName := fmt.Sprintf("%s_%d", timestamp, numOfXlsxFiles)
			logger.Info.Printf("Task %s: Create Excel file %s.xlsx\n", option.TaskID, xlsxFileName)
			xlsxFilePath, _err := option.XlsxCreateHandler(records, xlsxFileName, localeStr, option.AppID)
			if _err != nil {
				err = _err
				return
			}
			xlsxFilePaths = append(xlsxFilePaths, xlsxFilePath)

			records = make([]interface{}, 0)
			numOfRecordsPerXlsx = 0
		}
	}
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

		logger.Info.Printf("Start doing housekeeping of %s...\n", todayStr)
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
					dirTimestamp, err := time.Parse(data.XlsxDirTimestampFormat, fInfo.Name())
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
