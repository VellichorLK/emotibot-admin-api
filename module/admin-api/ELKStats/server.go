package ELKStats

import (
	"fmt"
	"net/http"
	"time"

	controllersV1 "emotibot.com/emotigo/module/admin-api/ELKStats/controllers/v1"
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
			util.NewEntryPoint("GET", "call", []string{"view"}, controllersV1.CallStatsGetHandler),
		},
	}

	err = services.InitTags()
	if err != nil {
		return err
	}

	return common.RecordsServiceInit()
}
