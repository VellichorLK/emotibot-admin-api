package ELKStats

import (
	"fmt"
	"net/http"
	"time"

	"emotibot.com/emotigo/module/admin-api/ELKStats/controllers"
	"emotibot.com/emotigo/module/admin-api/ELKStats/services"
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
	envs := util.GetModuleEnvironments(moduleName)
	dalAddress, ok := envs["DAL_URL"]
	if !ok {
		return fmt.Errorf("Require Module Env DAL_URL")
	}
	var httpClient = &http.Client{
		Timeout: time.Duration(5) * time.Second,
	}
	dalClient, err := dal.NewClientWithHTTPClient(dalAddress, httpClient)
	if err != nil {
		return fmt.Errorf("init dal client failed, %v", err)
	}
	ModuleInfo = util.ModuleInfo{
		ModuleName: moduleName,
		EntryPoints: []util.EntryPoint{
			util.NewEntryPoint("GET", "visit", []string{}, controllers.VisitStatsGetHandler),
			util.NewEntryPoint("GET", "question", []string{}, controllers.QuestionStatsGetHandler),
			util.NewEntryPoint("POST", "records/query", []string{}, controllers.VisitRecordsGetHandler),
			util.NewEntryPoint("POST", "records/download", []string{}, controllers.RecordsDownloadHandler),
			util.NewEntryPoint("POST", "records/mark", []string{}, controllers.NewRecordsMarkUpdateHandler(dalClient)),
			util.NewEntryPoint("POST", "records/ignore", []string{}, controllers.RecordsIgnoredUpdateHandler),
			util.NewEntryPoint("GET", "call", []string{}, controllers.CallStatsGetHandler),
		},
	}
	return services.InitTags()
}
