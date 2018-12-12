package ELKStats

import (
	"fmt"
	"net/http"
	"time"

	controllersV1 "emotibot.com/emotigo/module/admin-api/ELKStats/controllers/v1"
	controllersV2 "emotibot.com/emotigo/module/admin-api/ELKStats/controllers/v2"
	"emotibot.com/emotigo/module/admin-api/ELKStats/services"
	"emotibot.com/emotigo/module/admin-api/ELKStats/services/common"
	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/pkg/api/dal/v1"
)

var (
	// ModuleInfo is needed for module define
	ModuleInfo util.ModuleInfo
)

// Init init the package ModuleInfo & other essential data
func Init() error {
	var moduleName = "stats"
	var dalClient *dal.Client
	var err error

	dalAddress, ok := util.GetEnvOf("server")["DAL_URL"]
	if ok {
		var httpClient = &http.Client{
			Timeout: time.Duration(5) * time.Second,
		}
		dalClient, err = dal.NewClientWithHTTPClient(dalAddress, httpClient)
		if err != nil {
			err = fmt.Errorf("init dal client failed, %v", err)
		}
	} else {
		err = fmt.Errorf("Require Module Env DAL_URL")
	}
	if err != nil {
		return err
	}

	ModuleInfo = util.ModuleInfo{
		ModuleName: moduleName,
		EntryPoints: []util.EntryPoint{
			// v1 APIs
			util.NewEntryPoint("GET", "visit", []string{"view"}, controllersV1.VisitStatsGetHandler),
			util.NewEntryPoint("GET", "question", []string{"view"}, controllersV1.QuestionStatsGetHandler),
			util.NewEntryPoint("POST", "records/query", []string{"view"}, controllersV1.VisitRecordsGetHandler),
			util.NewEntryPoint("POST", "records/export", []string{"view", "export"}, controllersV1.VisitRecordsExportHandler),
			util.NewEntryPoint("GET", "records/export/{export_id}",
				[]string{"view", "export"}, controllersV1.VisitRecordsExportDownloadHandler),
			util.NewEntryPoint("DELETE", "records/export/{export_id}",
				[]string{"view", "export"}, controllersV1.VisitRecordsExportDeleteHandler),
			util.NewEntryPoint("GET", "records/export/{export_id}/status",
				[]string{"view", "export"}, controllersV1.VisitRecordsExportStatusHandler),
			util.NewEntryPoint("POST", "records/mark", []string{"view", "export"}, controllersV1.NewRecordsMarkUpdateHandler(dalClient)),
			util.NewEntryPoint("POST", "records/ignore", []string{"view", "export"}, controllersV1.RecordsIgnoredUpdateHandler),
			util.NewEntryPoint("GET", "records/{id}/marked", []string{"view", "export"}, controllersV1.NewRecordSSMHandler(dalClient)),
			util.NewEntryPoint("GET", "teVisit", []string{"view"}, controllersV1.TEVisitStatsGetHandler),
			util.NewEntryPoint("POSt", "teRecords/query", []string{"view"}, controllersV1.TEVisitRecordsGetHandler),
			util.NewEntryPoint("POST", "teRecords/export", []string{"view", "export"}, controllersV1.TEVisitRecordsExportHandler),
			util.NewEntryPoint("GET", "teRecords/export/{export_id}",
				[]string{"view", "export"}, controllersV1.TEVisitRecordsExportDownloadHandler),
			util.NewEntryPoint("DELETE", "teRecords/export/{export_id}",
				[]string{"view", "export"}, controllersV1.TEVisitRecordsExportDeleteHandler),
			util.NewEntryPoint("GET", "teRecords/export/{export_id}/status",
				[]string{"view", "export"}, controllersV1.TEVisitRecordsExportStatusHandler),
			util.NewEntryPoint("GET", "feedbacks", []string{"view"}, controllersV1.FeedbacksGetHandler),
			util.NewEntryPoint("GET", "call", []string{"view"}, controllersV1.CallStatsGetHandler),
			// v2 APIs
			util.NewEntryPointWithVer("POST", "records/query", []string{"view"}, controllersV2.VisitRecordsGetHandler, 2),
			util.NewEntryPointWithVer("POST", "records/export", []string{"view", "export"}, controllersV2.VisitRecordsExportHandler, 2),
			util.NewEntryPointWithVer("GET", "records/export/{export_id}",
				[]string{"view", "export"}, controllersV2.VisitRecordsExportDownloadHandler, 2),
			util.NewEntryPointWithVer("DELETE", "records/export/{export_id}",
				[]string{"view", "export"}, controllersV2.VisitRecordsExportDeleteHandler, 2),
			util.NewEntryPointWithVer("GET", "records/export/{export_id}/status",
				[]string{"view", "export"}, controllersV2.VisitRecordsExportStatusHandler, 2),
			util.NewEntryPointWithVer("POST", "records/mark", []string{"view", "export"}, controllersV2.NewRecordsMarkUpdateHandler(dalClient), 2),
			util.NewEntryPointWithVer("POST", "records/ignore", []string{"view", "export"}, controllersV2.RecordsIgnoredUpdateHandler, 2),
			util.NewEntryPointWithVer("GET", "records/{id}/marked", []string{"view", "export"}, controllersV2.NewRecordSSMHandler(dalClient), 2),
		},
	}

	err = services.InitTags()
	if err != nil {
		return err
	}

	return common.RecordsServiceInit()
}
