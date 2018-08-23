package ELKStats

import (
	"emotibot.com/emotigo/module/admin-api/ELKStats/controllers"
	"emotibot.com/emotigo/module/admin-api/ELKStats/services"
	"emotibot.com/emotigo/module/admin-api/util"
)

var (
	// ModuleInfo is needed for module define
	ModuleInfo util.ModuleInfo
)

func init() {
	ModuleInfo = util.ModuleInfo{
		ModuleName: "stats",
		EntryPoints: []util.EntryPoint{
			util.NewEntryPoint("GET", "visit", []string{}, controllers.VisitStatsGetHandler),
			util.NewEntryPoint("GET", "question", []string{}, controllers.QuestionStatsGetHandler),
			util.NewEntryPoint("POST", "records/query", []string{}, controllers.VisitRecordsGetHandler),
			util.NewEntryPoint("POST", "records/download", []string{}, controllers.RecordsDownloadHandler),
			util.NewEntryPoint("POST", "records/mark", []string{}, controllers.RecordsMarkUpdateHandler),
			util.NewEntryPoint("POST", "records/ignore", []string{}, controllers.RecordsIgnoredUpdateHandler),
			util.NewEntryPoint("GET", "call", []string{}, controllers.CallStatsGetHandler),
		},
	}
}

func InitTags() error {
	return services.InitTags()
}
