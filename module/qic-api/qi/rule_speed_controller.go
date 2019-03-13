package qi

import (
	"encoding/json"
	"errors"
	"net/http"

	"emotibot.com/emotigo/module/admin-api/ApiError"
	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/admin-api/util/requestheader"
	model "emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/module/qic-api/util/general"
	"emotibot.com/emotigo/pkg/logger"
)

//Error msg
var (
	ErrorWrongMin = errors.New("invalid speed min")
	ErrorWrongMax = errors.New("invalid speed max")
)

func handleNewRuleSpeed(w http.ResponseWriter, r *http.Request) {
	enterprise := requestheader.GetEnterpriseID(r)

	var requestBody model.SpeedRule
	err := util.ReadJSON(r, &requestBody)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
		return
	}

	err = checkSpeedRule(&requestBody)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
		return
	}

	uuid, err := NewRuleSpeed(&requestBody, enterprise)
	if err != nil {
		logger.Error.Printf("create rule speed failed. %s\n", err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}

	err = util.WriteJSON(w, struct {
		UUID string `json:"speed_id"`
	}{UUID: uuid})
	if err != nil {
		logger.Error.Printf("%s\n", err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.JSON_PARSE_ERROR, err.Error()), http.StatusInternalServerError)
	}
}

func checkSpeedRule(r *model.SpeedRule) error {
	if r == nil {
		return ErrEmptyRequest
	}
	if r.Name == "" {
		return ErrEmptyName
	}
	if r.Score < 0 {
		return ErrorWrongScore
	}
	if r.Min <= 0 {
		return ErrorWrongMin
	}
	if r.Max <= 0 {
		return ErrorWrongMax
	}
	return nil
}

func handleGetRuleSpeedList(w http.ResponseWriter, r *http.Request) {
	enterprise := requestheader.GetEnterpriseID(r)
	page, limit, err := getPageLimit(r)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err), http.StatusBadRequest)
		return
	}

	isDelete := 0
	q := &model.GeneralQuery{Enterprise: &enterprise, IsDelete: &isDelete}
	p := &model.Pagination{Limit: limit, Page: page}

	resp, err := GetRuleSpeeds(q, p)
	if err != nil {
		logger.Error.Printf("get the speed rule failed. %s\n", err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}

	total, err := CountRuleSpeed(q)
	if err != nil {
		logger.Error.Printf("count the speed rule failed. q: %+v, err: %s\n", *q, err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}

	err = util.WriteJSON(w, struct {
		Page pageResp           `json:"paging"`
		Data []*model.SpeedRule `json:"data"`
	}{
		Page: pageResp{Current: page, Limit: limit, Total: uint64(total)},
		Data: resp,
	})
	if err != nil {
		logger.Error.Printf("%s\n", err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.JSON_PARSE_ERROR, err.Error()), http.StatusInternalServerError)
	}
}

func handleGetRuleSpeed(w http.ResponseWriter, r *http.Request) {
	enterprise := requestheader.GetEnterpriseID(r)
	uuid := general.ParseID(r)
	isDelete := 0
	q := &model.GeneralQuery{UUID: []string{uuid}, Enterprise: &enterprise, IsDelete: &isDelete}

	settings, err := GetRuleSpeeds(q, nil)
	if err != nil {
		logger.Error.Printf("get the speed rule failed. %s\n", err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}
	if len(settings) == 0 {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, "no such id"), http.StatusBadRequest)
		return
	}

	except, err := GetRuleSpeedException(settings[0])
	if err != nil {
		logger.Error.Printf("get the exception of speed rule failed. %s\n", err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}

	err = util.WriteJSON(w, struct {
		Setting   model.SpeedRule    `json:"setting"`
		Exception RuleSpeedException `json:"exception"`
	}{
		Setting:   *settings[0],
		Exception: *except,
	})
	if err != nil {
		logger.Error.Printf("%s\n", err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.JSON_PARSE_ERROR, err.Error()), http.StatusInternalServerError)
	}
}

func handleDeleteRuleSpeed(w http.ResponseWriter, r *http.Request) {
	enterprise := requestheader.GetEnterpriseID(r)
	uuid := general.ParseID(r)

	isDelete := 0
	q := &model.GeneralQuery{UUID: []string{uuid}, Enterprise: &enterprise, IsDelete: &isDelete}
	_, err := DeleteRuleSpeed(q)
	if err != nil {
		logger.Error.Printf("delete %s failed. %s\n", uuid, err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}
}

func checkSpeedRuleUpdateSet(r model.SpeedUpdateSet) error {
	sr := model.SpeedRule{Name: "valid name", Score: 99, Min: 99, Max: 99}
	if r.Name != nil {
		sr.Name = *r.Name
	}
	if r.Score != nil {
		sr.Score = *r.Score
	}
	if r.Min != nil {
		sr.Min = *r.Min
	}
	if r.Max != nil {
		sr.Max = *r.Max
	}
	return checkSpeedRule(&sr)
}

func handleModifyRuleSpeed(w http.ResponseWriter, r *http.Request) {
	enterprise := requestheader.GetEnterpriseID(r)
	uuid := general.ParseID(r)

	var req model.SpeedUpdateSet

	err := util.ReadJSON(r, &req)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
		return
	}

	err = checkSpeedRuleUpdateSet(req)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
		return
	}

	//just in case
	req.ExceptionOver = nil
	req.ExceptionUnder = nil

	isDelete := 0
	_, err = UpdateRuleSpeed(&model.GeneralQuery{UUID: []string{uuid}, Enterprise: &enterprise, IsDelete: &isDelete}, &req)
	if err != nil {
		if err == ErrNoSuchID {
			util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
		} else {
			logger.Error.Printf("update %s failed. %s\n", uuid, err)
			util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		}
	}
}

func handleExceptionRuleSpeedUnder(w http.ResponseWriter, r *http.Request) {
	handleExceptionRuleSpeed(w, r, "under")
}

func handleExceptionRuleSpeedOver(w http.ResponseWriter, r *http.Request) {
	handleExceptionRuleSpeed(w, r, "over")
}

func handleExceptionRuleSpeed(w http.ResponseWriter, r *http.Request, exceptType string) {
	enterprise := requestheader.GetEnterpriseID(r)
	uuid := general.ParseID(r)
	var req RuleExceptionInteral

	err := util.ReadJSON(r, &req)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
		return
	}

	var updateSet model.SpeedUpdateSet

	except, err := json.Marshal(req)
	if err != nil {
		logger.Error.Printf("marshal %+v failed. %s\n", req.Customer, err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.JSON_PARSE_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}
	exceptStr := string(except)
	if exceptType == "under" {
		updateSet.ExceptionUnder = &exceptStr
	} else if exceptType == "over" {
		updateSet.ExceptionOver = &exceptStr
	}

	isDelete := 0
	_, err = UpdateRuleSpeed(&model.GeneralQuery{UUID: []string{uuid}, Enterprise: &enterprise, IsDelete: &isDelete}, &updateSet)
	if err != nil {
		if err == ErrNoSuchID {
			util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
		} else {
			logger.Error.Printf("update %s failed. %s\n", uuid, err)
			util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		}
	}
}
