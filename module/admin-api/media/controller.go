package media

import (
	"fmt"
	"net/http"
	"strconv"

	"emotibot.com/emotigo/module/admin-api/util/AdminErrors"

	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/admin-api/util/requestheader"
	"emotibot.com/emotigo/pkg/logger"
)

var (
	// ModuleInfo is needed for module define
	ModuleInfo util.ModuleInfo
)

func init() {
	ModuleInfo = util.ModuleInfo{
		ModuleName: "media",
		EntryPoints: []util.EntryPoint{
			util.NewEntryPoint("POST", "image", []string{"edit"}, handleUploadImage),
			util.NewEntryPoint("GET", "image/{id}", []string{}, handleGetImage),
		},
	}
}

func handleUploadImage(w http.ResponseWriter, r *http.Request) {
	appid := requestheader.GetAppID(r)

	file, info, formErr := r.FormFile("file")
	defer file.Close()
	logger.Info.Printf("Receive uploaded file: %s", info.Filename)
	logger.Trace.Printf("Uploaded file info %#v", info.Header)

	if formErr != nil {
		util.ReturnError(w, AdminErrors.ErrnoRequestError, fmt.Sprintf("Cannot get upload file, %s", formErr.Error()))
		return
	}

	id, err := AddFile(appid, file)
	util.Return(w, err, id)
}

func handleGetImage(w http.ResponseWriter, r *http.Request) {
	appid := requestheader.GetAppID(r)
	id := util.GetMuxVar(r, "id")
	if id == "" {
		w.Write([]byte("Id is invalid"))
		return
	}

	buf, err := GetFile(appid, id)
	if err != nil {
		w.Write([]byte(fmt.Sprintf("Get file fail, %s", err.Error())))
		return
	}

	contentType := http.DetectContentType(buf)

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", strconv.Itoa(len(buf)))
	w.Write(buf)
}
