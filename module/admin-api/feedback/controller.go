package feedback

import (
	"net/http"

	"emotibot.com/emotigo/module/admin-api/util"
)

var (
	// ModuleInfo is needed for module define
	ModuleInfo util.ModuleInfo
)

func init() {
	ModuleInfo = util.ModuleInfo{
		ModuleName: "feedback",
		EntryPoints: []util.EntryPoint{
			util.NewEntryPoint("GET", "feedback/reasons", []string{}, handleGetFeedbackReasons),
			util.NewEntryPoint("POST", "feedback/reason", []string{"edit"}, handleAddFeedbackReason),
			util.NewEntryPoint("DELTE", "feedback/reason/{id}", []string{"delete"}, handleDeleteFeedbackReason),
		},
	}
}

func handleGetFeedbackReasons(w http.ResponseWriter, r *http.Request) {

}
func handleAddFeedbackReason(w http.ResponseWriter, r *http.Request) {

}
func handleDeleteFeedbackReason(w http.ResponseWriter, r *http.Request) {

}
