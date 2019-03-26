package qi

import (
	"net/http"
	"fmt"
	"time"
	"emotibot.com/emotigo/pkg/logger"
	"os"
	"io"
	"emotibot.com/emotigo/module/admin-api/util/requestheader"
)

func handleImportTags(w http.ResponseWriter, r *http.Request) {
	var err error
	//appID := requestheader.GetAppID(r)
	enterpriseID := requestheader.GetEnterpriseID(r)

	fileName := fmt.Sprintf("tags_%s.xlsx", time.Now().Format("20060102150405"))

	if err = getUploadedFile(r, fileName); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	if err = BatchAddTags(fileName, enterpriseID); err != nil {
		logger.Error.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = removeUploadedFile(fileName); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleImportSentences(w http.ResponseWriter, r *http.Request) {
	var err error
	//appID := requestheader.GetAppID(r)
	enterpriseID := requestheader.GetEnterpriseID(r)

	fileName := fmt.Sprintf("sentences_%s.xlsx", time.Now().Format("20060102150405"))

	if err = getUploadedFile(r, fileName); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	if err = BatchAddSentences(fileName, enterpriseID); err != nil {
		logger.Error.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = removeUploadedFile(fileName); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleImportRules(w http.ResponseWriter, r *http.Request) {

	// TODO use general.UUID() to simply code

	var err error
	//appID := requestheader.GetAppID(r)
	enterpriseID := requestheader.GetEnterpriseID(r)

	fileName := fmt.Sprintf("rules_%s.xlsx", time.Now().Format("20060102150405"))

	if err = getUploadedFile(r, fileName); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	if err = BatchAddRules(fileName, enterpriseID); err != nil {
		logger.Error.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = removeUploadedFile(fileName); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleImportCallIn(w http.ResponseWriter, r *http.Request) {
	var err error
	enterpriseID := requestheader.GetEnterpriseID(r)

	fileName := fmt.Sprintf("flow_%s.xlsx", time.Now().Format("20060102150405"))

	if err = getUploadedFile(r, fileName); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	if err = BatchAddFlows(fileName, enterpriseID); err != nil {
		logger.Error.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = removeUploadedFile(fileName); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleExportGroups(w http.ResponseWriter, r *http.Request) {

	buf, err := ExportGroups()

	if err != nil {
		logger.Error.Printf("error while export groups in handleExportGroups, reason: %s \n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=export_rules_%s.xlsx", time.Now().Format("20060102150405")))
	w.Header().Set("Content-Type", "application/vnd.ms-excel")
	w.Write(buf.Bytes())
}

func handleImportGroups(w http.ResponseWriter, r *http.Request) {
	var err error
	//appID := requestheader.GetAppID(r)
	fileName := fmt.Sprintf("rule_group_%s.xlsx", time.Now().Format("20060102150405"))

	if err = getUploadedFile(r, fileName); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	if err = ImportGroups(fileName); err != nil {
		logger.Error.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = removeUploadedFile(fileName); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleExportCalls(w http.ResponseWriter, r *http.Request) {

	buf, err := ExportCalls()

	if err != nil {
		logger.Error.Printf("error while export calls in handleExportCalls, reason: %s", err.Error())
	}
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=export_calls_%s.xlsx", time.Now().Format("20060102150405")))
	w.Header().Set("Content-Type", "application/vnd.ms-excel")
	w.Write(buf.Bytes())
}

func getUploadedFile(r *http.Request, fileName string) error {
	r.ParseMultipartForm(32 << 20)
	file, info, err := r.FormFile("file")
	if err != nil {
		logger.Error.Println("fail to receive file")
		return fmt.Errorf("fail to receive file \n")
	}
	defer file.Close()
	logger.Trace.Printf("receive uploaded file: %s \n", info.Filename)

	// parse file
	size := info.Size
	if size == 0 {
		logger.Error.Println("file size is 0")
		return fmt.Errorf("file size is 0 \n")
	}

	f, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err = io.Copy(f, file); err != nil {
		return err
	}

	logger.Trace.Printf("save uploaded file %s \n", fileName)
	return nil
}

func removeUploadedFile(fileName string) error {
	if _, err := os.Stat(fileName); err == nil {
		os.Remove(fileName)
	} else {
		return err
	}
	logger.Trace.Printf("delete uploaded file %s \n", fileName)
	return nil
}
