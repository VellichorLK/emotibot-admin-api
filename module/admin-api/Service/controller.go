package Service

import (
	"fmt"
	"net/http"

	"emotibot.com/emotigo/module/admin-api/util"
)

var (
	// ModuleInfo is needed for module define
	ModuleInfo util.ModuleInfo
)

func init() {
	ModuleInfo = util.ModuleInfo{
		ModuleName: "service",
		EntryPoints: []util.EntryPoint{
			util.NewEntryPoint("GET", "w2v", []string{}, handleGetW2V),
		},
	}
}

func handleGetW2V(w http.ResponseWriter, r *http.Request) {
	src := r.URL.Query().Get("src")
	dst := r.URL.Query().Get("dst")
	if src == "" || dst == "" {
		w.Write([]byte("0"))
		return
	}

	ret := GetW2VResultFromSentence(src, dst)
	w.Write([]byte(fmt.Sprintf("%f", ret)))
}
