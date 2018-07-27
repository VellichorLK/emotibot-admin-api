package v2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/admin-api/util/AdminErrors"
	"emotibot.com/emotigo/module/admin-api/util/LocaleMsg"
)

var (
	// EntryList will be merged in the module controller
	EntryList = []util.EntryPoint{
		util.NewEntryPointWithVer("GET", "intents", []string{"view"}, handleGetIntentsV2, 2),
		util.NewEntryPointWithVer("POST", "intent", []string{"create"}, handleAddIntentV2, 2),
		util.NewEntryPointWithVer("GET", "intent/{intentID}", []string{"view"}, handleGetIntentV2, 2),
		util.NewEntryPointWithVer("PATCH", "intent/{intentID}", []string{"view"}, handleUpdateIntentV2, 2),
		util.NewEntryPointWithVer("DELETE", "intent/{intentID}", []string{"view"}, handleDeleteIntentV2, 2),
	}
)

func init() {
	initV2Dao()
}

func initV2Dao() {
	dao = intentDaoV2{}
}

func handleGetIntentsV2(w http.ResponseWriter, r *http.Request) {
	appid := util.GetAppID(r)
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
	appid := util.GetAppID(r)
	locale := util.GetLocale(r)
	keyword := strings.TrimSpace(r.URL.Query().Get("keyword"))
	intentID, convertErr := util.GetMuxInt64Var(r, "intentID")
	if convertErr != nil {
		util.LogTrace.Println("Transform to int fail: ", convertErr.Error())
		util.Return(w, AdminErrors.New(AdminErrors.ErrnoRequestError, LocaleMsg.Get(locale, "IntentID")), nil)
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
	appid := util.GetAppID(r)
	locale := util.GetLocale(r)
	intentID, convertErr := util.GetMuxInt64Var(r, "intentID")
	var err AdminErrors.AdminError
	var ret interface{}
	var origIntent *IntentV2

	defer func() {
		result := 0
		var auditBuf bytes.Buffer
		auditBuf.WriteString(LocaleMsg.Get(locale, "DeleteIntent"))
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
		util.AddAuditFromRequestAuto(r, auditMsg, result)
		util.Return(w, err, ret)
	}()

	if convertErr != nil {
		util.LogTrace.Println("Transform to int fail: ", convertErr.Error())
		err = AdminErrors.New(AdminErrors.ErrnoRequestError, LocaleMsg.Get(locale, "IntentID"))
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
	appid := util.GetAppID(r)
	locale := util.GetLocale(r)
	var err AdminErrors.AdminError
	var ret interface{}
	var newIntent *IntentV2

	defer func() {
		result := 0
		var auditBuf bytes.Buffer
		auditBuf.WriteString(LocaleMsg.Get(locale, "AddIntent"))
		if newIntent != nil {
			msgTemplate := LocaleMsg.Get(locale, "AddIntentSummaryTpl")
			auditBuf.WriteString(fmt.Sprintf(msgTemplate,
				newIntent.Name, newIntent.PositiveCount, newIntent.NegativeCount))
		}
		if err != nil {
			auditBuf.WriteString(LocaleMsg.Get(locale, "Fail"))
			auditBuf.WriteString(", ")
			auditBuf.WriteString(err.String())
		}
		auditMsg := auditBuf.String()
		if err == nil {
			result = 1
		} else {
			ret = auditMsg
		}
		util.AddAuditFromRequestAuto(r, auditMsg, result)
		util.Return(w, err, ret)
	}()

	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		err = AdminErrors.New(AdminErrors.ErrnoRequestError, LocaleMsg.Get(locale, "IntentName"))
		return
	}

	positiveStr := strings.TrimSpace(r.FormValue("positive"))
	negativeStr := strings.TrimSpace(r.FormValue("negative"))

	positive := []string{}
	negative := []string{}
	jsonErr := json.Unmarshal([]byte(positiveStr), &positive)
	if jsonErr != nil {
		err = AdminErrors.New(AdminErrors.ErrnoRequestError, LocaleMsg.Get(locale, "IntentPositive"))
		return
	}
	jsonErr = json.Unmarshal([]byte(negativeStr), &negative)
	if jsonErr != nil {
		err = AdminErrors.New(AdminErrors.ErrnoRequestError, LocaleMsg.Get(locale, "IntentNegative"))
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
	appid := util.GetAppID(r)
	locale := util.GetLocale(r)
	var err AdminErrors.AdminError
	var ret interface{}
	var newIntent *IntentV2
	updateList := []*SentenceV2WithType{}
	deleteList := []int64{}

	defer func() {
		result := 0
		var auditBuf bytes.Buffer
		auditBuf.WriteString(LocaleMsg.Get(locale, "UpdateIntent"))
		if newIntent != nil {
			msgTemplate := LocaleMsg.Get(locale, "UpdateIntentSummaryTpl")
			auditBuf.WriteString(fmt.Sprintf(msgTemplate,
				newIntent.Name, len(updateList), len(deleteList)))
		}
		if err != nil {
			auditBuf.WriteString(LocaleMsg.Get(locale, "Fail"))
			auditBuf.WriteString(", ")
			auditBuf.WriteString(err.String())
		}
		auditMsg := auditBuf.String()
		if err == nil {
			result = 1
		} else {
			ret = auditMsg
		}
		util.AddAuditFromRequestAuto(r, auditMsg, result)
		util.Return(w, err, ret)
	}()
	intentID, convertErr := util.GetMuxInt64Var(r, "intentID")
	if convertErr != nil {
		util.LogTrace.Println("Transform to int fail: ", convertErr.Error())
		util.Return(w, AdminErrors.New(AdminErrors.ErrnoRequestError, LocaleMsg.Get(locale, "IntentID")), nil)
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		util.LogTrace.Println("Error name param in request")
		err = AdminErrors.New(AdminErrors.ErrnoRequestError, LocaleMsg.Get(locale, "IntentName"))
		return
	}

	updateStr := strings.TrimSpace(r.FormValue("update"))
	deleteStr := strings.TrimSpace(r.FormValue("delete"))

	jsonErr := json.Unmarshal([]byte(updateStr), &updateList)
	if jsonErr != nil {
		err = AdminErrors.New(AdminErrors.ErrnoRequestError, LocaleMsg.Get(locale, "IntentModifyUpdate"))
		return
	}
	jsonErr = json.Unmarshal([]byte(deleteStr), &deleteList)
	if jsonErr != nil {
		err = AdminErrors.New(AdminErrors.ErrnoRequestError, LocaleMsg.Get(locale, "IntentModifyDelete"))
		return
	}

	newIntent, err = ModifyIntent(appid, intentID, name, updateList, deleteList)
	if err != nil {
		return
	}
	ret = newIntent
	return
}
