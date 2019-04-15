package QADoc

import (
	"net/http"

	"emotibot.com/emotigo/module/admin-api/QADoc/controllers"
	"emotibot.com/emotigo/module/admin-api/util"
)

var (
	// ModuleInfo is needed for module define
	ModuleInfo util.ModuleInfo
)

func init() {
	var moduleName = "qa-docs"

	ModuleInfo = util.ModuleInfo{
		ModuleName: moduleName,
		EntryPoints: []util.EntryPoint{
			// v1 APIs
			util.NewEntryPoint(http.MethodPost, "", []string{"edit"}, controllers.CreateQADocHandler),
			util.NewEntryPoint(http.MethodPost, "bulk", []string{"edit"}, controllers.CreateQADocsHandler),
			util.NewEntryPoint(http.MethodDelete, "bulk", []string{"edit"}, controllers.DeleteQADocsHandler),
			util.NewEntryPoint(http.MethodDelete, "query", []string{"edit"}, controllers.DeleteQADocsByIDsHandler),
		},
	}
}
