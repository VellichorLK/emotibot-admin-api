package Service

import (
	"fmt"
	"net/http"
	"strings"

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
			util.NewEntryPoint("GET", "w2v_seg", []string{}, handleGetW2VSeg),
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

func handleGetW2VSeg(w http.ResponseWriter, r *http.Request) {
	src_seg := strings.Split(r.URL.Query().Get("src_seg"), ",")
	src_kw := strings.Split(r.URL.Query().Get("src_kw"), ",")
	tar_seg := strings.Split(r.URL.Query().Get("tar_seg"), ",")
	tar_kw := strings.Split(r.URL.Query().Get("tar_kw"), ",")

	ret := GetW2VResultFromSeg(src_seg, src_kw, tar_seg, tar_kw)
	w.Write([]byte(fmt.Sprintf("%f", ret)))
}
