package common

import (
	"net/http"

	"github.com/olivere/elastic"
)

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

type UpdateCommand func(*elastic.UpdateByQueryService) *elastic.UpdateByQueryService

func VisitRecordsExportDownload(w http.ResponseWriter, exportTaskID string) error {
	return DownloadExportRecords(w, exportTaskID)
}

func VisitRecordsExportDelete(exportTaskID string) error {
	return DeleteExportRecords(exportTaskID)
}

func VisitRecordsExportStatus(exportTaskID string) (status string, err error) {
	status, err = GetRecordsExportStatus(exportTaskID)
	return
}

func UpdateRecordMark(status bool) UpdateCommand {
	return func(qs *elastic.UpdateByQueryService) *elastic.UpdateByQueryService {
		script := elastic.NewScript("ctx._source.isMarked = params.mark")
		script.Param("mark", status)
		qs.Script(script)
		return qs
	}
}

func UpdateRecordIntentMark(intentID int64) UpdateCommand {
	return func(qs *elastic.UpdateByQueryService) *elastic.UpdateByQueryService {
		script := elastic.NewScript("ctx._source.marked_intent = params.mark")
		script.Param("mark", intentID)

		if intentID < 0 {
			script = elastic.NewScript("ctx._source.remove(\"marked_intent\")")
		}
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
