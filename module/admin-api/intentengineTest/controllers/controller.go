package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"emotibot.com/emotigo/module/admin-api/intentengineTest/data"
	"emotibot.com/emotigo/module/admin-api/intentengineTest/services"
	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/admin-api/util/AdminErrors"
	"emotibot.com/emotigo/module/admin-api/util/requestheader"
)

func IntentTestsGetHandler(w http.ResponseWriter, r *http.Request) {
	appID := requestheader.GetAppID(r)
	results, err := services.GetIntentTests(appID)
	if err != nil {
		util.Return(w, err, nil)
		return
	}

	util.Return(w, nil, results)
}

func IntentTestsStatusHandler(w http.ResponseWriter, r *http.Request) {
	appID := requestheader.GetAppID(r)
	version, status, sentencesCount, progress,
		err := services.GetIntentTestStatus(appID)
	if err != nil {
		util.Return(w, err, nil)
		return
	}

	resp := data.IntentTestStatusResp{
		Version:        version,
		Status:         status,
		SentencesCount: sentencesCount,
		Progress:       progress,
	}

	util.Return(w, nil, resp)
}

func IntentTestGetHandler(w http.ResponseWriter, r *http.Request) {
	appID := requestheader.GetAppID(r)
	locale := requestheader.GetLocale(r)
	intentTestID, convErr := util.GetMuxInt64Var(r, "intent_test_id")
	if convErr != nil {
		util.Return(w, AdminErrors.New(AdminErrors.ErrnoRequestError,
			convErr.Error()), nil)
		return
	}
	keyword := strings.TrimSpace(r.URL.Query().Get("keyword"))

	result, err := services.GetIntentTest(appID, intentTestID, keyword, locale)
	if err != nil {
		util.Return(w, err, nil)
		return
	}

	util.Return(w, nil, result)
}

func IntentTestPatchHandler(w http.ResponseWriter, r *http.Request) {
	locale := requestheader.GetLocale(r)
	intentTestID, convErr := util.GetMuxInt64Var(r, "intent_test_id")
	if convErr != nil {
		util.Return(w, AdminErrors.New(AdminErrors.ErrnoRequestError,
			convErr.Error()), nil)
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))

	err := services.PatchIntentTest(intentTestID, name, locale)
	if err != nil {
		util.Return(w, err, nil)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func IntentTestSaveHandler(w http.ResponseWriter, r *http.Request) {
	locale := requestheader.GetLocale(r)
	intentTestID, convErr := util.GetMuxInt64Var(r, "intent_test_id")
	if convErr != nil {
		util.Return(w, AdminErrors.New(AdminErrors.ErrnoRequestError,
			convErr.Error()), nil)
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))

	err := services.SaveIntentTest(intentTestID, name, locale)
	if err != nil {
		util.Return(w, err, nil)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func IntentTestUnsaveHandler(w http.ResponseWriter, r *http.Request) {
	locale := requestheader.GetLocale(r)
	intentTestID, convErr := util.GetMuxInt64Var(r, "intent_test_id")
	if convErr != nil {
		util.Return(w, AdminErrors.New(AdminErrors.ErrnoRequestError,
			convErr.Error()), nil)
		return
	}

	err := services.UnsaveIntentTest(intentTestID, locale)
	if err != nil {
		util.Return(w, err, nil)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func IntentTestExportHandler(w http.ResponseWriter, r *http.Request) {
	appID := requestheader.GetAppID(r)
	locale := requestheader.GetLocale(r)
	intentTestID, convErr := util.GetMuxInt64Var(r, "intent_test_id")
	if convErr != nil {
		util.Return(w, AdminErrors.New(AdminErrors.ErrnoRequestError,
			convErr.Error()), nil)
		return
	}

	buf, err := services.ExportIntentTest(appID, &intentTestID, locale)
	if err != nil {
		util.Return(w, err, nil)
		return
	}

	filename := createExportFilename()
	util.ReturnFile(w, filename, buf)
}

func IntentTestRestoreHandler(w http.ResponseWriter, r *http.Request) {
	appID := requestheader.GetAppID(r)
	locale := requestheader.GetLocale(r)
	intentTestID, convErr := util.GetMuxInt64Var(r, "intent_test_id")
	if convErr != nil {
		util.Return(w, AdminErrors.New(AdminErrors.ErrnoRequestError,
			convErr.Error()), nil)
		return
	}

	err := services.RestoreIntentTest(appID, intentTestID, locale)
	if err != nil {
		util.Return(w, err, nil)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func LatestIntentsGetHandler(w http.ResponseWriter, r *http.Request) {
	appID := requestheader.GetAppID(r)
	locale := requestheader.GetLocale(r)
	keyword := strings.TrimSpace(r.URL.Query().Get("keyword"))

	results, err := services.GetLatestIntents(appID, keyword, locale)
	if err != nil {
		util.Return(w, err, nil)
		return
	}

	util.Return(w, nil, results)
}

func LatestIntentTestImportHandler(w http.ResponseWriter, r *http.Request) {
	appID := requestheader.GetAppID(r)
	locale := requestheader.GetLocale(r)

	file, _, ioErr := r.FormFile("file")
	if ioErr != nil {
		util.Return(w, AdminErrors.New(AdminErrors.ErrnoIOError,
			ioErr.Error()), nil)
		return
	}

	var buf bytes.Buffer
	_, ioErr = io.Copy(&buf, file)
	if ioErr != nil {
		util.Return(w, AdminErrors.New(AdminErrors.ErrnoIOError,
			ioErr.Error()), nil)
		return
	}

	err := services.ImportLatestIntentTest(appID, buf.Bytes(), locale)
	if err != nil {
		util.Return(w, err, nil)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func LatestIntentTestExportHandler(w http.ResponseWriter, r *http.Request) {
	appID := requestheader.GetAppID(r)
	locale := requestheader.GetLocale(r)

	buf, err := services.ExportIntentTest(appID, nil, locale)
	if err != nil {
		util.Return(w, err, nil)
		return
	}

	filename := createExportFilename()
	util.ReturnFile(w, filename, buf)
}

func IntentTestsTestHandler(w http.ResponseWriter, r *http.Request) {
	appID := requestheader.GetAppID(r)
	userID := requestheader.GetUserID(r)
	ieModelID := strings.TrimSpace(r.FormValue("ie_model_id"))
	locale := requestheader.GetLocale(r)

	if ieModelID == "" {
		util.Return(w, AdminErrors.New(AdminErrors.ErrnoRequestError,
			"ie_model_id"), nil)
		return
	}

	err := services.TestIntents(appID, userID, ieModelID, locale)
	if err != nil {
		util.Return(w, err, nil)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func UsableModelsGetHandler(w http.ResponseWriter, r *http.Request) {
	appID := requestheader.GetAppID(r)
	models, err := services.GetUsableModels(appID)
	if err != nil {
		util.Return(w, err, nil)
		return
	}
	util.Return(w, nil, models)
}

func IntentGetHandler(w http.ResponseWriter, r *http.Request) {
	locale := requestheader.GetLocale(r)
	testIntentID, convErr := util.GetMuxInt64Var(r, "intent_id")
	if convErr != nil {
		util.Return(w, AdminErrors.New(AdminErrors.ErrnoRequestError,
			"intent_id"), convErr.Error())
		return
	}
	keyword := strings.TrimSpace(r.URL.Query().Get("keyword"))

	results, err := services.GetIntent(testIntentID, keyword, locale)
	if err != nil {
		util.Return(w, err, nil)
		return
	}

	util.Return(w, nil, results)
}

func IntentUpdateHandler(w http.ResponseWriter, r *http.Request) {
	locale := requestheader.GetLocale(r)
	testIntentID, convErr := util.GetMuxInt64Var(r, "intent_id")
	if convErr != nil {
		util.Return(w, AdminErrors.New(AdminErrors.ErrnoRequestError,
			"intent_id"), convErr.Error())
		return
	}
	updateList := make([]*data.UpdateCmd, 0)
	deleteList := make([]int64, 0)

	updateStr := strings.TrimSpace(r.FormValue("update"))
	deleteStr := strings.TrimSpace(r.FormValue("delete"))

	err := json.Unmarshal([]byte(updateStr), &updateList)
	if err != nil {
		util.Return(w, AdminErrors.New(AdminErrors.ErrnoRequestError,
			err.Error()), nil)
		return
	}

	err = json.Unmarshal([]byte(deleteStr), &deleteList)
	if err != nil {
		util.Return(w, AdminErrors.New(AdminErrors.ErrnoRequestError,
			err.Error()), nil)
		return
	}

	updateErr := services.UpdateIntent(testIntentID, updateList,
		deleteList, locale)
	if updateErr != nil {
		util.Return(w, updateErr, nil)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func createExportFilename() string {
	now := time.Now()
	return fmt.Sprintf("intent_test_%s.xlsx", now.Format("20060102_150405"))
}
