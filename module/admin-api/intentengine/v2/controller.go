package intentenginev2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/admin-api/util/AdminErrors"
	"emotibot.com/emotigo/module/admin-api/util/audit"
	"emotibot.com/emotigo/module/admin-api/util/localemsg"
	"emotibot.com/emotigo/module/admin-api/util/requestheader"
	"emotibot.com/emotigo/module/admin-api/util/validate"
	"emotibot.com/emotigo/pkg/logger"
)

func init() {
	initV2Dao()
}

func initV2Dao() {
	dao = intentDaoV2{}
}

func handleSearchSentence(w http.ResponseWriter, r *http.Request) {
	appid := requestheader.GetAppID(r)
	versionStr := r.URL.Query().Get("version")
	content := r.URL.Query().Get("content")
	sentenceTypeStr := r.URL.Query().Get("type")

	if content == "" {
		util.Return(w, AdminErrors.New(AdminErrors.ErrnoRequestError, "content empty"), nil)
		return
	}

	var version *int
	if versionStr == "" {
		version = nil
	} else {
		val, convErr := strconv.Atoi(versionStr)
		if convErr != nil {
			util.Return(w, AdminErrors.New(AdminErrors.ErrnoRequestError, "version"), convErr.Error())
			return
		}
		version = &val
	}

	var name string
	var err AdminErrors.AdminError
	sentenceType, convErr := strconv.Atoi(sentenceTypeStr)
	if convErr != nil {
		name, sentenceType, err = SearchSentence(appid, version, content)
	} else {
		name, err = SearchSentenceWithType(appid, version, content, sentenceType)
	}

	if err != nil {
		util.Return(w, err, nil)
	} else {
		util.Return(w, nil, map[string]interface{}{
			"name": name,
			"type": sentenceType,
		})
	}
}

func handleGetIntentsV2(w http.ResponseWriter, r *http.Request) {
	appid := requestheader.GetAppID(r)
	versionStr := r.URL.Query().Get("version")
	keyword := strings.TrimSpace(r.URL.Query().Get("keyword"))

	var version *int
	if val, err := strconv.Atoi(versionStr); err == nil {
		version = &val
	} else {
		version = nil
	}

	intents, err := GetIntents(appid, version, keyword)
	if err != nil {
		util.Return(w, err, "")
		return
	}
	util.Return(w, nil, intents)
}

func handleGetIntentV2(w http.ResponseWriter, r *http.Request) {
	appid := requestheader.GetAppID(r)
	locale := requestheader.GetLocale(r)
	keyword := strings.TrimSpace(r.URL.Query().Get("keyword"))
	intentID, convertErr := util.GetMuxInt64Var(r, "intentID")
	if convertErr != nil {
		logger.Trace.Println("Transform to int fail: ", convertErr.Error())
		util.Return(w, AdminErrors.New(AdminErrors.ErrnoRequestError, localemsg.Get(locale, "IntentID")), nil)
		return
	}

	intent, err := GetIntent(appid, intentID, keyword)
	if err != nil {
		util.Return(w, err, nil)
		return
	}
	util.Return(w, nil, intent)
}

func handleDeleteIntentV2(w http.ResponseWriter, r *http.Request) {
	appid := requestheader.GetAppID(r)
	locale := requestheader.GetLocale(r)
	intentID, convertErr := util.GetMuxInt64Var(r, "intentID")
	var err AdminErrors.AdminError
	var ret interface{}
	var origIntent *IntentV2

	defer func() {
		result := 0
		var auditBuf bytes.Buffer
		auditBuf.WriteString(localemsg.Get(locale, "DeleteIntent"))
		if origIntent != nil {
			auditBuf.WriteString(": ")
			auditBuf.WriteString(origIntent.Name)
		}
		auditMsg := auditBuf.String()
		if err == nil || err.Errno() == AdminErrors.ErrnoNotFound {
			result = 1
			err = nil
		} else {
			ret = auditMsg
		}
		audit.AddAuditFromRequestAuto(r, auditMsg, result)
		util.Return(w, err, ret)
	}()

	if convertErr != nil {
		logger.Trace.Println("Transform to int fail: ", convertErr.Error())
		err = AdminErrors.New(AdminErrors.ErrnoRequestError, localemsg.Get(locale, "IntentID"))
		return
	}

	origIntent, err = GetIntent(appid, intentID, "")
	if err != nil {
		return
	}

	err = DeleteIntent(appid, intentID)
	return
}

func handleAddIntentV2(w http.ResponseWriter, r *http.Request) {
	appid := requestheader.GetAppID(r)
	locale := requestheader.GetLocale(r)
	var err AdminErrors.AdminError
	var ret interface{}
	var newIntent *IntentV2

	defer func() {
		result := 0
		var auditBuf bytes.Buffer
		auditBuf.WriteString(localemsg.Get(locale, "AddIntent"))
		if newIntent != nil {
			msgTemplate := localemsg.Get(locale, "AddIntentSummaryTpl")
			auditBuf.WriteString(fmt.Sprintf(msgTemplate,
				newIntent.Name, newIntent.PositiveCount, newIntent.NegativeCount))
		}
		if err != nil {
			auditBuf.WriteString(localemsg.Get(locale, "Fail"))
			auditBuf.WriteString(", ")
			auditBuf.WriteString(err.String())
		}
		auditMsg := auditBuf.String()
		if err == nil {
			result = 1
		} else {
			ret = auditMsg
		}
		audit.AddAuditFromRequestAuto(r, auditMsg, result)
		util.Return(w, err, ret)
	}()

	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		err = AdminErrors.New(AdminErrors.ErrnoRequestError, localemsg.Get(locale, "IntentName"))
		return
	}

	positiveStr := strings.TrimSpace(r.FormValue("positive"))
	negativeStr := strings.TrimSpace(r.FormValue("negative"))

	positive := []string{}
	negative := []string{}
	jsonErr := json.Unmarshal([]byte(positiveStr), &positive)
	if jsonErr != nil {
		err = AdminErrors.New(AdminErrors.ErrnoRequestError, localemsg.Get(locale, "IntentPositive"))
		return
	}
	jsonErr = json.Unmarshal([]byte(negativeStr), &negative)
	if jsonErr != nil {
		err = AdminErrors.New(AdminErrors.ErrnoRequestError, localemsg.Get(locale, "IntentNegative"))
		return
	}

	newIntent, err = AddIntent(appid, name, positive, negative)
	if err != nil {
		return
	}
	ret = newIntent
	return
}

func handleUpdateIntentV2(w http.ResponseWriter, r *http.Request) {
	appid := requestheader.GetAppID(r)
	locale := requestheader.GetLocale(r)
	var err AdminErrors.AdminError
	var ret interface{}
	var newIntent *IntentV2
	updateList := []*SentenceV2WithType{}
	deleteList := []int64{}

	defer func() {
		result := 0
		var auditBuf bytes.Buffer
		auditBuf.WriteString(localemsg.Get(locale, "UpdateIntent"))
		if newIntent != nil {
			msgTemplate := localemsg.Get(locale, "UpdateIntentSummaryTpl")
			auditBuf.WriteString(fmt.Sprintf(msgTemplate,
				newIntent.Name, len(updateList), len(deleteList)))
		}
		if err != nil {
			auditBuf.WriteString(localemsg.Get(locale, "Fail"))
			auditBuf.WriteString(", ")
			auditBuf.WriteString(err.String())
		}
		auditMsg := auditBuf.String()
		if err == nil {
			result = 1
		} else {
			ret = auditMsg
		}
		audit.AddAuditFromRequestAuto(r, auditMsg, result)
		util.Return(w, err, ret)
	}()
	intentID, convertErr := util.GetMuxInt64Var(r, "intentID")
	if convertErr != nil {
		logger.Trace.Println("Transform to int fail: ", convertErr.Error())
		util.Return(w, AdminErrors.New(AdminErrors.ErrnoRequestError, localemsg.Get(locale, "IntentID")), nil)
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		logger.Trace.Println("Error name param in request")
		err = AdminErrors.New(AdminErrors.ErrnoRequestError, localemsg.Get(locale, "IntentName"))
		return
	}

	updateStr := strings.TrimSpace(r.FormValue("update"))
	deleteStr := strings.TrimSpace(r.FormValue("delete"))

	jsonErr := json.Unmarshal([]byte(updateStr), &updateList)
	if jsonErr != nil {
		err = AdminErrors.New(AdminErrors.ErrnoRequestError, localemsg.Get(locale, "IntentModifyUpdate"))
		return
	}
	jsonErr = json.Unmarshal([]byte(deleteStr), &deleteList)
	if jsonErr != nil {
		err = AdminErrors.New(AdminErrors.ErrnoRequestError, localemsg.Get(locale, "IntentModifyDelete"))
		return
	}

	newIntent, err = ModifyIntent(appid, intentID, name, updateList, deleteList)
	if err != nil {
		return
	}
	ret = newIntent
	return
}

func handleGetIntentStatusV2(w http.ResponseWriter, r *http.Request) {
	appid := requestheader.GetAppID(r)

	status, err := GetIntentEngineStatus(appid)
	if err != nil {
		util.Return(w, err, nil)
		return
	}
	util.Return(w, nil, status)
}

func handleStartTrain(w http.ResponseWriter, r *http.Request) {
	appid := requestheader.GetAppID(r)

	version, err := StartTrain(appid)
	if err != nil {
		util.Return(w, err, nil)
		return
	}
	util.Return(w, nil, version)
}

func handleGetTrainDataV2(w http.ResponseWriter, r *http.Request) {
	appid := r.URL.Query().Get("app_id")
	if appid == "" {
		util.Return(w, AdminErrors.New(AdminErrors.ErrnoRequestError,
			"app_id not specified"), nil)
		return
	}
	rsp, err := GetTrainData(appid)
	if err != nil {
		util.Return(w, err, nil)
		return
	}
	js, jsonErr := json.Marshal(rsp)
	if jsonErr != nil {
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(js)
	return
}

func handleImportIntentV2(w http.ResponseWriter, r *http.Request) {
	appid := requestheader.GetAppID(r)
	locale := requestheader.GetLocale(r)
	var err AdminErrors.AdminError
	var auditMsg bytes.Buffer

	defer func() {
		retVal := 0
		if err == nil {
			retVal = 1
		} else {
			auditMsg.WriteString(":")
			auditMsg.WriteString(err.Error())
		}

		audit.AddAuditFromRequestAuto(r, auditMsg.String(), retVal)
		util.Return(w, err, auditMsg.String())
	}()
	auditMsg.WriteString(util.Msg["UploadIntentEngine"])

	file, info, ioErr := r.FormFile("file")
	if ioErr != nil {
		err = AdminErrors.New(AdminErrors.ErrnoIOError, "")
		return
	}

	var buffer bytes.Buffer
	_, ioErr = io.Copy(&buffer, file)
	if ioErr != nil {
		err = AdminErrors.New(AdminErrors.ErrnoIOError, "")
		return
	}
	auditMsg.WriteString(info.Filename)

	intents, parseErr := ParseImportIntentFile(buffer.Bytes(), locale)
	if parseErr != nil {
		err = AdminErrors.New(AdminErrors.ErrnoRequestError, parseErr.Error())
		return
	}
	auditMsg.WriteString(fmt.Sprintf(util.Msg["UploadIntentInfoTpl"], len(intents)))

	err = UpdateLatestIntents(appid, intents)
	if err != nil {
		return
	}
	return
}

func handleExportIntentV2(w http.ResponseWriter, r *http.Request) {
	appid := r.URL.Query().Get("appid")
	locale := requestheader.GetLocale(r)
	var err AdminErrors.AdminError
	var ret []byte
	var auditMsg bytes.Buffer

	defer func() {
		retVal := 0
		if err == nil {
			retVal = 1
			now := time.Now()
			filename := fmt.Sprintf("intent_%d%02d%02d_%02d%02d%02d.xlsx",
				now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second())
			util.ReturnFile(w, filename, ret)
			auditMsg.WriteString(":")
			auditMsg.WriteString(filename)
		} else {
			auditMsg.WriteString(":")
			auditMsg.WriteString(err.Error())
			util.Return(w, err, nil)
		}
		audit.AddAuditFromRequestAuto(r, auditMsg.String(), retVal)
	}()
	auditMsg.WriteString(localemsg.Get(locale, "IntentExport"))

	if !validate.IsValidAppID(appid) {
		err = AdminErrors.New(AdminErrors.ErrnoRequestError, "APPID")
		return
	}

	ret, err = GetExportIntentsBFFormat(appid, locale)
	return
}

func getEnvironments() map[string]string {
	return util.GetEnvOf(moduleName)
}

func getEnvironment(key string) string {
	envs := util.GetEnvOf(moduleName)
	if envs != nil {
		if val, ok := envs[key]; ok {
			return val
		}
	}
	return ""
}
