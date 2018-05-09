package FAQ

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"emotibot.com/emotigo/module/admin-api/ApiError"
	"emotibot.com/emotigo/module/admin-api/util"
)

func parseLabelFromRequest(r *http.Request) (*Label, error) {
	ret := &Label{}

	ret.Name = strings.TrimSpace(r.FormValue("name"))
	if ret.Name == "" {
		return nil, errors.New("Invalid name")
	}

	return ret, nil
}

func handleDeleteLabel(w http.ResponseWriter, r *http.Request) {
	var retObj interface{}
	status := http.StatusOK
	retCode := ApiError.SUCCESS
	defer func() {
		if status == http.StatusOK {
			util.WriteJSON(w, util.GenRetObj(retCode, retObj))
		} else {
			util.WriteJSONWithStatus(w, util.GenRetObj(retCode, retObj), status)
		}
	}()

	appid := util.GetAppID(r)
	id, err := util.GetMuxIntVar(r, "id")
	if err != nil {
		status, retCode, retObj = genBadRequestReturn(err)
		return
	}
	retCode, err = DeleteLabel(appid, id)
	if err != nil {
		retObj = err.Error()
		switch retCode {
		case ApiError.REQUEST_ERROR:
			status = http.StatusBadRequest
		default:
			status = http.StatusInternalServerError
		}
		return
	}
	retObj = nil
}

func handleUpdateLabel(w http.ResponseWriter, r *http.Request) {
	var retObj interface{}
	status := http.StatusOK
	retCode := ApiError.SUCCESS
	defer func() {
		if status == http.StatusOK {
			util.WriteJSON(w, util.GenRetObj(retCode, retObj))
		} else {
			util.WriteJSONWithStatus(w, util.GenRetObj(retCode, retObj), status)
		}
	}()

	appid := util.GetAppID(r)
	newLabel, err := parseLabelFromRequest(r)
	if err != nil {
		status, retCode, retObj = genBadRequestReturn(err)
		return
	}
	id, err := util.GetMuxIntVar(r, "id")
	if err != nil {
		status, retCode, retObj = genBadRequestReturn(err)
		return
	}
	newLabel.ID = id

	retCode, err = UpdateLabel(appid, newLabel)
	if err != nil {
		retObj = err.Error()
		status = http.StatusInternalServerError
		return
	}
	retObj = newLabel
}

func handleAddLabel(w http.ResponseWriter, r *http.Request) {
	var retObj interface{}
	status := http.StatusOK
	retCode := ApiError.SUCCESS
	defer func() {
		if status == http.StatusOK {
			util.WriteJSON(w, util.GenRetObj(retCode, retObj))
		} else {
			util.WriteJSONWithStatus(w, util.GenRetObj(retCode, retObj), status)
		}
	}()

	appid := util.GetAppID(r)
	newLabel, err := parseLabelFromRequest(r)
	if err != nil {
		status, retCode, retObj = genBadRequestReturn(err)
		return
	}

	retCode, err = AddNewLabel(appid, newLabel)
	if err != nil {
		retObj = err.Error()
		status = http.StatusInternalServerError
		return
	}
	retObj = newLabel
}

func handleGetLabels(w http.ResponseWriter, r *http.Request) {
	appid := util.GetAppID(r)
	tags, err := GetQuestionLabels(appid)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
	} else {
		util.WriteJSON(w, util.GenRetObj(ApiError.SUCCESS, tags))
	}
}

func genBadRequestReturn(err error) (int, int, interface{}) {
	return http.StatusBadRequest, ApiError.REQUEST_ERROR, err.Error()
}

func handleGetRules(w http.ResponseWriter, r *http.Request) {
	var retObj interface{}
	status := http.StatusOK
	retCode := ApiError.SUCCESS
	defer func() {
		if status == http.StatusOK {
			util.WriteJSON(w, util.GenRetObj(retCode, retObj))
		} else {
			util.WriteJSONWithStatus(w, util.GenRetObj(retCode, retObj), status)
		}
	}()
	appid := util.GetAppID(r)

	rules, err := GetRules(appid)
	if err != nil {
		status = http.StatusInternalServerError
		retCode = ApiError.DB_ERROR
		retObj = err.Error()
	} else {
		retObj = rules
	}
}
func handleGetRule(w http.ResponseWriter, r *http.Request) {
	var retObj interface{}
	status := http.StatusOK
	retCode := ApiError.SUCCESS
	defer func() {
		if status == http.StatusOK {
			util.WriteJSON(w, util.GenRetObj(retCode, retObj))
		} else {
			util.WriteJSONWithStatus(w, util.GenRetObj(retCode, retObj), status)
		}
	}()
	appid := util.GetAppID(r)
	id, err := util.GetMuxIntVar(r, "id")
	if err != nil {
		status = http.StatusBadRequest
		retCode = ApiError.REQUEST_ERROR
		retObj = err.Error()
		return
	}

	rule, err := GetRule(appid, id)
	if err != nil {
		status = http.StatusInternalServerError
		retCode = ApiError.DB_ERROR
		retObj = err.Error()
	} else if rule == nil {
		status = http.StatusNotFound
		retCode = ApiError.REQUEST_ERROR
	} else {
		retObj = rule
	}
}
func handleUpdateRule(w http.ResponseWriter, r *http.Request) {
	var retObj interface{}
	status := http.StatusOK
	retCode := ApiError.SUCCESS
	defer func() {
		if status == http.StatusOK {
			util.WriteJSON(w, util.GenRetObj(retCode, retObj))
		} else {
			util.WriteJSONWithStatus(w, util.GenRetObj(retCode, retObj), status)
		}
	}()

	appid := util.GetAppID(r)
	id, err := util.GetMuxIntVar(r, "id")
	if err != nil {
		status, retCode = http.StatusBadRequest, ApiError.REQUEST_ERROR
		return
	}

	origRule, err := GetRule(appid, id)
	if err != nil {
		status, retCode = http.StatusInternalServerError, ApiError.IO_ERROR
		retObj = err.Error()
		return
	}
	if origRule == nil {
		status, retCode = http.StatusNotFound, ApiError.REQUEST_ERROR
		return
	}

	rule, err := parseRuleFromRequest(r)
	if err != nil {
		status, retCode = http.StatusBadRequest, ApiError.REQUEST_ERROR
		return
	}

	err = UpdateRule(appid, id, rule)
	if err != nil {
		status, retCode = http.StatusInternalServerError, ApiError.IO_ERROR
		retObj = err.Error()
		return
	}
	retObj = rule
}
func handleAddRule(w http.ResponseWriter, r *http.Request) {
	var retObj interface{}
	status := http.StatusOK
	retCode := ApiError.SUCCESS
	defer func() {
		if status == http.StatusOK {
			util.WriteJSON(w, util.GenRetObj(retCode, retObj))
		} else {
			util.WriteJSONWithStatus(w, util.GenRetObj(retCode, retObj), status)
		}
	}()

	appid := util.GetAppID(r)
	rule, err := parseRuleFromRequest(r)
	if err != nil {
		status = http.StatusBadRequest
		retCode = ApiError.REQUEST_ERROR
		retObj = err.Error()
		return
	}

	id, err := AddRule(appid, rule)
	if err != nil {
		status = http.StatusInternalServerError
		retCode = ApiError.IO_ERROR
		retObj = err.Error()
		return
	}
	rule.ID = id
	retObj = rule
}
func handleDeleteRule(w http.ResponseWriter, r *http.Request) {
	var retObj interface{}
	status := http.StatusOK
	retCode := ApiError.SUCCESS
	defer func() {
		if status == http.StatusOK {
			util.WriteJSON(w, util.GenRetObj(retCode, retObj))
		} else {
			util.WriteJSONWithStatus(w, util.GenRetObj(retCode, retObj), status)
		}
	}()
	appid := util.GetAppID(r)
	id, err := util.GetMuxIntVar(r, "id")
	if err != nil {
		status = http.StatusBadRequest
		retCode = ApiError.REQUEST_ERROR
		retObj = err.Error()
		return
	}

	err = DeleteRule(appid, id)
	if err != nil {
		status = http.StatusInternalServerError
		retCode = ApiError.DB_ERROR
	}
}
func handleGetRulesOfLabel(w http.ResponseWriter, r *http.Request) {
	var retObj interface{}
	status, retCode := http.StatusOK, ApiError.SUCCESS
	defer func() {
		if status == http.StatusOK {
			util.WriteJSON(w, util.GenRetObj(retCode, retObj))
		} else {
			util.WriteJSONWithStatus(w, util.GenRetObj(retCode, retObj), status)
		}
	}()
	appid := util.GetAppID(r)
	labelID, err := util.GetMuxIntVar(r, "id")
	if err != nil {
		status, retCode = http.StatusBadRequest, ApiError.REQUEST_ERROR
		return
	}

	rules, err := GetRulesOfLabel(appid, labelID)
	if err != nil {
		status, retCode = http.StatusInternalServerError, ApiError.DB_ERROR
		retObj = err.Error()
	} else {
		retObj = rules
	}
}
func parseRuleFromRequest(r *http.Request) (rule *Rule, err error) {
	err = r.ParseForm()
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			util.LogInfo.Printf("Parse rule fail: %s\n", err.Error())
		}
	}()

	ret := Rule{}
	ret.Name = r.FormValue("name")
	ret.Answer = r.FormValue("answer")
	ret.Status = r.FormValue("status") != "0"
	begin, err := time.Parse(time.RFC3339, r.FormValue("begin_time"))
	if err != nil {
		ret.Begin = nil
	} else {
		ret.Begin = &begin
	}
	end, err := time.Parse(time.RFC3339, r.FormValue("end_time"))
	if err != nil {
		ret.End = nil
	} else {
		ret.End = &end
	}

	target, err := strconv.Atoi(r.FormValue("target"))
	if err != nil {
		err = errors.New("Invalid target")
		return
	}
	rtype, err := strconv.Atoi(r.FormValue("response_type"))
	if err != nil {
		err = errors.New("Invalid response type")
		return
	}

	if target > ret.Target.Max() || target < 0 {
		err = errors.New("Invalid target")
		return
	}
	if rtype > ret.Type.Max() || rtype < 0 {
		err = errors.New("Invalid response type")
		return
	}
	ret.Target = RuleTarget(target)
	ret.Type = ResponseType(rtype)

	ruleStr := r.FormValue("rule")
	ruleContents := []*RuleContent{}
	err = json.Unmarshal([]byte(ruleStr), &ruleContents)
	if err != nil {
		err = fmt.Errorf("Invalid rule content: %s", err.Error())
		return
	}
	for i, r := range ruleContents {
		if !r.IsValid() {
			err = fmt.Errorf("rule content error of rule %d", i+1)
			return
		}
	}
	ret.Rule = ruleContents

	labelsStr := r.FormValue("labels")
	labels := strings.Split(labelsStr, ",")
	existedLabel := map[int]bool{}
	for _, labelStr := range labels {
		str := strings.TrimSpace(labelStr)
		if str == "" {
			continue
		}
		id := 0
		id, err = strconv.Atoi(str)
		if err != nil {
			return
		}
		if _, ok := existedLabel[id]; !ok {
			ret.LinkLabel = append(ret.LinkLabel, id)
			existedLabel[id] = true
		}
	}

	rule = &ret
	return
}

func handleGetLabelsOfRule(w http.ResponseWriter, r *http.Request) {
	var retObj interface{}
	status, retCode := http.StatusOK, ApiError.SUCCESS
	defer func() {
		if status == http.StatusOK {
			util.WriteJSON(w, util.GenRetObj(retCode, retObj))
		} else {
			util.WriteJSONWithStatus(w, util.GenRetObj(retCode, retObj), status)
		}
	}()
	appid := util.GetAppID(r)
	ruleID, err := util.GetMuxIntVar(r, "id")
	if err != nil {
		status, retCode = http.StatusBadRequest, ApiError.REQUEST_ERROR
		return
	}

	labels, err := GetLabelsOfRule(appid, ruleID)
	if err != nil {
		status, retCode = http.StatusInternalServerError, ApiError.DB_ERROR
		retObj = err.Error()
	} else {
		retObj = labels
	}
}

// func handleAddLabelToRule(w http.ResponseWriter, r *http.Request) {

// }
// func handleDeleteLabelFromRule(w http.ResponseWriter, r *http.Request) {

// }
// func handleAddRuleToLabel(w http.ResponseWriter, r *http.Request) {

// }
// func handleDeleteRuleFromLabel(w http.ResponseWriter, r *http.Request) {

// }
