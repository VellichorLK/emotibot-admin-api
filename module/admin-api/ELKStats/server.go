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
			util.NewEntryPoint("POST", "record", []string{}, controllers.VisitRecordsGetHandler),
			util.NewEntryPoint("GET", "call", []string{}, controllers.CallStatsGetHandler),
		},
	}
}

func InitTags() error {
	return services.InitTags()
}
