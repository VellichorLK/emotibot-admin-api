package qi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"emotibot.com/emotigo/module/admin-api/ApiError"
	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/admin-api/util/requestheader"
	"emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/module/qic-api/util/general"
	"emotibot.com/emotigo/pkg/logger"
)

type silenceRq struct {
	Name    string `json:"name"`
	Score   int    `json:"score"`
	Seconds int    `json:"seconds"`
	Times   int    `json:"times"`
}

//Error msg
var (
	ErrEmptyName    = errors.New("empty name")
	ErrEmptyRequest = errors.New("empty request")
	ErrWrongSecond  = errors.New("invalid seconds")
	ErrorWrongTimes = errors.New("invalid times")
	ErrorWrongScore = errors.New("invalid score")
	ErrNoSuchID     = errors.New("no such id")
)

type exceptionList struct {
	Staff    []string `json:"staff"`
	Customer []string `json:"customer"`
}

func checkSilenceRule(r *model.SilenceRule) error {
	if r == nil {
		return ErrEmptyRequest
	}
	if r.Name == "" {
		return ErrEmptyName
	}
	if r.Score < 0 {
		return ErrorWrongScore
	}
	if r.Seconds <= 0 {
		return ErrWrongSecond
	}
	if r.Times <= 0 {
		return ErrorWrongTimes
	}
	return nil
}

func handleNewRuleSilence(w http.ResponseWriter, r *http.Request) {
	enterprise := requestheader.GetEnterpriseID(r)

	var requestBody model.SilenceRule
	err := util.ReadJSON(r, &requestBody)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
		return
	}

	err = checkSilenceRule(&requestBody)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
		return
	}

	uuid, err := NewRuleSilence(&requestBody, enterprise)
	if err != nil {
		logger.Error.Printf("create rule silence failed. %s\n", err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}

	err = util.WriteJSON(w, struct {
		UUID string `json:"silence_id"`
	}{UUID: uuid})
	if err != nil {
		logger.Error.Printf("%s\n", err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.JSON_PARSE_ERROR, err.Error()), http.StatusInternalServerError)
	}
}

func handleGetRuleSilenceList(w http.ResponseWriter, r *http.Request) {
	enterprise := requestheader.GetEnterpriseID(r)
	page, limit, err := getPageLimit(r)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err), http.StatusBadRequest)
		return
	}

	isDelete := 0
	q := &model.GeneralQuery{Enterprise: &enterprise, IsDelete: &isDelete}
	p := &model.Pagination{Limit: limit, Page: page}

	resp, err := GetRuleSilences(q, p)
	if err != nil {
		logger.Error.Printf("get the silence rule failed. %s\n", err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}

	total, err := CountRuleSilence(q)
	if err != nil {
		logger.Error.Printf("count the flows failed. q: %+v, err: %s\n", *q, err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}

	err = util.WriteJSON(w, struct {
		Page pageResp             `json:"paging"`
		Data []*model.SilenceRule `json:"data"`
	}{
		Page: pageResp{Current: page, Limit: limit, Total: uint64(total)},
		Data: resp,
	})
	if err != nil {
		logger.Error.Printf("%s\n", err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.JSON_PARSE_ERROR, err.Error()), http.StatusInternalServerError)
	}
}

func handleGetRuleSilence(w http.ResponseWriter, r *http.Request) {
	enterprise := requestheader.GetEnterpriseID(r)

	uuid := general.ParseID(r)

	isDelete := 0
	q := &model.GeneralQuery{UUID: []string{uuid}, Enterprise: &enterprise, IsDelete: &isDelete}

	settings, err := GetRuleSilences(q, nil)
	if err != nil {
		logger.Error.Printf("get the silence rule failed. %s\n", err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}

	if len(settings) == 0 {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, "no such id"), http.StatusBadRequest)
		return
	}

	except, _, err := GetRuleSilenceException(settings[0])
	if err != nil {
		logger.Error.Printf("get the exception rule failed. %s\n", err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}

	err = util.WriteJSON(w, struct {
		Setting   model.SilenceRule    `json:"setting"`
		Exception RuleSilenceException `json:"exception"`
	}{
		Setting: *settings[0],

		Exception: *except,
	})

	if err != nil {
		logger.Error.Printf("%s\n", err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.JSON_PARSE_ERROR, err.Error()), http.StatusInternalServerError)
	}

}

func handleDeleteRuleSilence(w http.ResponseWriter, r *http.Request) {
	enterprise := requestheader.GetEnterpriseID(r)
	uuid := general.ParseID(r)

	isDelete := 0
	q := &model.GeneralQuery{UUID: []string{uuid}, Enterprise: &enterprise, IsDelete: &isDelete}
	_, err := DeleteRuleSilence(q)
	if err != nil {
		logger.Error.Printf("delete %s failed. %s\n", uuid, err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}
}

func checkUpdateSet(r model.SilenceUpdateSet) error {
	sr := model.SilenceRule{Name: "valid name", Score: 99, Seconds: 99, Times: 99}
	if r.Name != nil {
		sr.Name = *r.Name
	}
	if r.Score != nil {
		sr.Score = *r.Score
	}
	if r.Seconds != nil {
		sr.Seconds = *r.Seconds
	}
	if r.Times != nil {
		sr.Times = *r.Times
	}
	return checkSilenceRule(&sr)

}

func handleModifyRuleSilence(w http.ResponseWriter, r *http.Request) {
	enterprise := requestheader.GetEnterpriseID(r)
	uuid := general.ParseID(r)

	var req model.SilenceUpdateSet

	err := util.ReadJSON(r, &req)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
		return
	}

	err = checkUpdateSet(req)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
		return
	}

	//just in case
	req.ExceptionBefore = nil
	req.ExceptionAfter = nil

	isDelete := 0
	_, err = UpdateRuleSilence(&model.GeneralQuery{UUID: []string{uuid}, Enterprise: &enterprise, IsDelete: &isDelete}, &req)
	if err != nil {
		if err == ErrNoSuchID {
			util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
		} else {
			logger.Error.Printf("update %s failed. %s\n", uuid, err)
			util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		}
	}
}

func handleExceptionRuleSilenceBefore(w http.ResponseWriter, r *http.Request) {
	enterprise := requestheader.GetEnterpriseID(r)
	uuid := general.ParseID(r)
	var req RuleExceptionInteral

	err := util.ReadJSON(r, &req)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
		return
	}

	var updateSet model.SilenceUpdateSet

	except, err := json.Marshal(req)
	if err != nil {
		logger.Error.Printf("marshal %+v failed. %s\n", req.Customer, err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.JSON_PARSE_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}
	exceptStr := string(except)

	updateSet.ExceptionBefore = &exceptStr

	isDelete := 0
	_, err = UpdateRuleSilence(&model.GeneralQuery{UUID: []string{uuid}, Enterprise: &enterprise, IsDelete: &isDelete}, &updateSet)
	if err != nil {
		if err == ErrNoSuchID {
			util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
		} else {
			logger.Error.Printf("update %s failed. %s\n", uuid, err)
			util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		}
	}

}

func handleExceptionRuleSilenceAfter(w http.ResponseWriter, r *http.Request) {
	enterprise := requestheader.GetEnterpriseID(r)
	uuid := general.ParseID(r)

	var req RuleExceptionInteral

	err := util.ReadJSON(r, &req)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
		return
	}

	var updateSet model.SilenceUpdateSet
	req.Customer = nil

	except, err := json.Marshal(req)
	if err != nil {
		logger.Error.Printf("marshal %+v failed. %s\n", req.Customer, err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.JSON_PARSE_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}
	exceptStr := string(except)

	updateSet.ExceptionAfter = &exceptStr

	isDelete := 0
	_, err = UpdateRuleSilence(&model.GeneralQuery{UUID: []string{uuid}, Enterprise: &enterprise, IsDelete: &isDelete}, &updateSet)
	if err != nil {
		if err == ErrNoSuchID {
			util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
		} else {
			logger.Error.Printf("update %s failed. %s\n", uuid, err)
			util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		}
	}
}

//ReqUsrData is the user request
type ReqUsrData string

//Key to put into context in the middleware
const (
	IDKey         = ReqUsrData("id")
	EnterpriseKey = ReqUsrData("enterprise")
)

//WithIntIDCheck checks the id with int64 type
func WithIntIDCheck(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := general.ParseID(r)
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
			return
		}
		c := context.WithValue(r.Context(), IDKey, id)
		r = r.WithContext(c)
		next.ServeHTTP(w, r)
	}
}
