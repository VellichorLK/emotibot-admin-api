package qi

import (
	"net/http"

	"emotibot.com/emotigo/module/admin-api/ApiError"
	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/admin-api/util/requestheader"
	model "emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/module/qic-api/util/general"
	"emotibot.com/emotigo/pkg/logger"
)

func handleNewRuleInterposal(w http.ResponseWriter, r *http.Request) {
	enterprise := requestheader.GetEnterpriseID(r)

	var requestBody model.InterposalRule
	err := util.ReadJSON(r, &requestBody)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
		return
	}

	err = checkInterposalRule(&requestBody)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
		return
	}

	uuid, err := NewRuleInterposal(&requestBody, enterprise)
	if err != nil {
		logger.Error.Printf("create interposal speed failed. %s\n", err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}

	err = util.WriteJSON(w, struct {
		UUID string `json:"interposal_id"`
	}{UUID: uuid})
	if err != nil {
		logger.Error.Printf("%s\n", err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.JSON_PARSE_ERROR, err.Error()), http.StatusInternalServerError)
	}
}

func checkInterposalRule(r *model.InterposalRule) error {
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

// util.NewEntryPoint(http.MethodPost, "rule/interposal", []string{}, handleNewRuleInterposal),
// util.NewEntryPoint(http.MethodGet, "rule/interposal", []string{}, handleGetRuleInterposalList),
// util.NewEntryPoint(http.MethodGet, "rule/interposal/{id}", []string{}, handleGetRuleInterposal),
// util.NewEntryPoint(http.MethodDelete, "rule/interposal/{id}", []string{}, handleDeleteRuleInterposal),
// util.NewEntryPoint(http.MethodPut, "rule/interposal/{id}", []string{}, handleModifyRuleInterposal),
func handleGetRuleInterposalList(w http.ResponseWriter, r *http.Request) {
	enterprise := requestheader.GetEnterpriseID(r)
	page, limit, err := getPageLimit(r)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err), http.StatusBadRequest)
		return
	}

	isDelete := 0
	q := &model.GeneralQuery{Enterprise: &enterprise, IsDelete: &isDelete}
	p := &model.Pagination{Limit: limit, Page: page}

	resp, err := GetRuleInterposals(q, p)
	if err != nil {
		logger.Error.Printf("get the interposal rule failed. %s\n", err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}

	total, err := CountRuleInterposal(q)
	if err != nil {
		logger.Error.Printf("count the interposal rule failed. q: %+v, err: %s\n", *q, err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}

	err = util.WriteJSON(w, struct {
		Page pageResp                `json:"paging"`
		Data []*model.InterposalRule `json:"data"`
	}{
		Page: pageResp{Current: page, Limit: limit, Total: uint64(total)},
		Data: resp,
	})
	if err != nil {
		logger.Error.Printf("%s\n", err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.JSON_PARSE_ERROR, err.Error()), http.StatusInternalServerError)
	}
}

func handleGetRuleInterposal(w http.ResponseWriter, r *http.Request) {
	enterprise := requestheader.GetEnterpriseID(r)
	uuid := general.ParseID(r)
	isDelete := 0
	q := &model.GeneralQuery{UUID: []string{uuid}, Enterprise: &enterprise, IsDelete: &isDelete}

	settings, err := GetRuleInterposals(q, nil)
	if err != nil {
		logger.Error.Printf("get the interposal rule failed. %s\n", err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}
	if len(settings) == 0 {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, "no such id"), http.StatusBadRequest)
		return
	}

	err = util.WriteJSON(w, struct {
		Setting model.InterposalRule `json:"setting"`
	}{
		Setting: *settings[0],
	})
	if err != nil {
		logger.Error.Printf("%s\n", err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.JSON_PARSE_ERROR, err.Error()), http.StatusInternalServerError)
	}
}

func handleDeleteRuleInterposal(w http.ResponseWriter, r *http.Request) {
	enterprise := requestheader.GetEnterpriseID(r)
	uuid := general.ParseID(r)

	isDelete := 0
	q := &model.GeneralQuery{UUID: []string{uuid}, Enterprise: &enterprise, IsDelete: &isDelete}
	_, err := DeleteRuleInterposal(q)
	if err != nil {
		logger.Error.Printf("delete %s failed. %s\n", uuid, err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}
}

func checkInterposalRuleUpdateSet(r model.InterposalUpdateSet) error {
	sr := model.InterposalRule{Name: "valid name", Score: 99, Seconds: 99, Times: 99}
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
	return checkInterposalRule(&sr)
}

func handleModifyRuleInterposal(w http.ResponseWriter, r *http.Request) {
	enterprise := requestheader.GetEnterpriseID(r)
	uuid := general.ParseID(r)

	var req model.InterposalUpdateSet

	err := util.ReadJSON(r, &req)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
		return
	}

	err = checkInterposalRuleUpdateSet(req)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
		return
	}

	isDelete := 0
	_, err = UpdateRuleInterposal(&model.GeneralQuery{UUID: []string{uuid}, Enterprise: &enterprise, IsDelete: &isDelete}, &req)
	if err != nil {
		if err == ErrNoSuchID {
			util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
		} else {
			logger.Error.Printf("update %s failed. %s\n", uuid, err)
			util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		}
	}
}
