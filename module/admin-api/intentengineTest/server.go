package intentengineTest

import (
	"emotibot.com/emotigo/module/admin-api/intentengineTest/controllers"
	"emotibot.com/emotigo/module/admin-api/intentengineTest/services"
	"emotibot.com/emotigo/module/admin-api/util"
)

var (
	// ModuleInfo is needed for module define
	ModuleInfo util.ModuleInfo
)

func init() {
	var moduleName = "intent_tests"

	ModuleInfo = util.ModuleInfo{
		ModuleName: moduleName,
		EntryPoints: []util.EntryPoint{
			// v1 APIs
			util.NewEntryPoint("GET", "", []string{"view"}, controllers.IntentTestsGetHandler),
			util.NewEntryPoint("GET", "status", []string{"view"}, controllers.IntentTestsStatusHandler),
			util.NewEntryPoint("GET", "intents", []string{"view"}, controllers.LatestIntentsGetHandler),
			util.NewEntryPoint("POST", "import", []string{"edit"}, controllers.LatestIntentTestImportHandler),
			util.NewEntryPoint("GET", "export", []string{"view"}, controllers.LatestIntentTestExportHandler),
			util.NewEntryPoint("POST", "test", []string{"view"}, controllers.IntentTestsTestHandler),
			util.NewEntryPoint("GET", "models", []string{"view"}, controllers.UsableModelsGetHandler),
			util.NewEntryPoint("GET", "{intent_test_id}", []string{"view"}, controllers.IntentTestGetHandler),
			util.NewEntryPoint("PATCH", "{intent_test_id}", []string{"view"}, controllers.IntentTestPatchHandler),
			util.NewEntryPoint("POST", "{intent_test_id}/save", []string{"edit"}, controllers.IntentTestSaveHandler),
			util.NewEntryPoint("DELETE", "{intent_test_id}/unsave", []string{"edit"}, controllers.IntentTestUnsaveHandler),
			util.NewEntryPoint("GET", "{intent_test_id}/export", []string{"view"}, controllers.IntentTestExportHandler),
			util.NewEntryPoint("POST", "{intent_test_id}/restore", []string{"view"}, controllers.IntentTestRestoreHandler),
			util.NewEntryPoint("GET", "intents/{intent_id}", []string{"view"}, controllers.IntentGetHandler),
			util.NewEntryPoint("PATCH", "intents/{intent_id}", []string{"edit"}, controllers.IntentUpdateHandler),
		},
		OneTimeFunc: map[string]func(){
			"init dao": func() {
				services.InitDao()
			},
		},
	}
}
