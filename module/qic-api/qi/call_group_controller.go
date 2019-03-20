package qi

import (
	"errors"
	"net/http"
	"strconv"

	"emotibot.com/emotigo/module/admin-api/ApiError"
	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/admin-api/util/requestheader"
	model "emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/module/qic-api/util/general"
	"emotibot.com/emotigo/pkg/logger"
)

func handleCreateCallGroupCondition(w http.ResponseWriter, r *http.Request) {
	enterprise := requestheader.GetEnterpriseID(r)

	var reqModel model.CallGroupCondition
	err := util.ReadJSON(r, &reqModel)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
		return
	}

	err = checkCallGroupCondition(&reqModel)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
		return
	}

	id, err := CreateCallGroupCondition(&reqModel, enterprise)
	if err != nil {
		logger.Error.Printf("create call group condition failed. %s\n", err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}

	err = util.WriteJSON(w, struct {
		ID int64 `json:"cg_condition_id"`
	}{ID: id})
	if err != nil {
		logger.Error.Printf("%s\n", err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.JSON_PARSE_ERROR, err.Error()), http.StatusInternalServerError)
	}
}

func checkCallGroupCondition(cond *model.CallGroupCondition) error {
	if cond == nil {
		return ErrEmptyRequest
	}
	if cond.Name == "" {
		return ErrEmptyName
	}
	if cond.DayRange <= 0 {
		return errors.New("invalid day_range")
	}
	if cond.DurationMin < 0 {
		return errors.New("invalid duration_min")
	}
	if cond.DurationMax <= 0 {
		return errors.New("invalid duration_max")
	}
	return nil
}

func handleGetCallGroupConditionList(w http.ResponseWriter, r *http.Request) {
	enterprise := requestheader.GetEnterpriseID(r)
	page, limit, err := getPageLimit(r)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err), http.StatusBadRequest)
		return
	}

	isDelete := 0
	query := &model.GeneralQuery{Enterprise: &enterprise, IsDelete: &isDelete}
	pagination := &model.Pagination{Limit: limit, Page: page}

	condList, err := GetCallGroupConditionList(query, pagination)
	if err != nil {
		logger.Error.Printf("get call group list failed. %s\n", err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}
	respList := make([]*model.CallGroupConditionListResponseItem, 0)
	for idx := range condList {
		cond := condList[idx]
		isEnable := cond.IsEnable == 1
		resp := model.CallGroupConditionListResponseItem{
			ID:          cond.ID,
			Name:        cond.Name,
			Description: cond.Description,
			IsEnable:    isEnable,
		}
		respList = append(respList, &resp)
	}

	total, err := CountCallGroupCondition(query)
	if err != nil {
		logger.Error.Printf("count call group list failed. query: %+v, err: %s\n", *query, err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}

	err = util.WriteJSON(w, struct {
		Page pageResp                                    `json:"paging"`
		Data []*model.CallGroupConditionListResponseItem `json:"data"`
	}{
		Page: pageResp{Current: page, Limit: limit, Total: uint64(total)},
		Data: respList,
	})
	if err != nil {
		logger.Error.Printf("%s\n", err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.JSON_PARSE_ERROR, err.Error()), http.StatusInternalServerError)
	}
}

func handleGetCallGroupCondition(w http.ResponseWriter, r *http.Request) {
	enterprise := requestheader.GetEnterpriseID(r)
	isDelete := 0
	idStr := general.ParseID(r)
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		logger.Error.Printf("invalid cg_condition_id: %s\n", idStr)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
		return
	}
	query := &model.GeneralQuery{ID: []int64{id}, Enterprise: &enterprise, IsDelete: &isDelete}

	condList, err := GetCallGroupConditionList(query, nil)
	if err != nil {
		logger.Error.Printf("get call group list failed. %s\n", err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}
	if len(condList) == 0 {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, "no such id"), http.StatusBadRequest)
		return
	}
	cond := condList[0]
	isEnable := cond.IsEnable == 1
	resp := model.CallGroupConditionResponse{
		Name:          cond.Name,
		Description:   cond.Description,
		IsEnable:      isEnable,
		DayRange:      cond.DayRange,
		DurationMin:   cond.DurationMin,
		DurationMax:   cond.DurationMax,
		FilterByValue: make([]interface{}, 0),
		GroupByValue:  make([]string, 0),
	}
	err = util.WriteJSON(w, resp)
	if err != nil {
		logger.Error.Printf("%s\n", err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.JSON_PARSE_ERROR, err.Error()), http.StatusInternalServerError)
	}
}

func handleUpdateCallGroupCondition(w http.ResponseWriter, r *http.Request) {
	enterprise := requestheader.GetEnterpriseID(r)
	isDelete := 0
	idStr := general.ParseID(r)
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		logger.Error.Printf("invalid cg_condition_id: %s\n", idStr)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
		return
	}
	query := &model.GeneralQuery{ID: []int64{id}, Enterprise: &enterprise, IsDelete: &isDelete}

	var reqModel model.CallGroupConditionUpdateSet
	err = util.ReadJSON(r, &reqModel)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
		return
	}
	err = checkCallGroupConditionUpdateSet(reqModel)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
		return
	}

	_, err = UpdateCallGroupCondition(query, &reqModel)
	if err != nil {
		if err == ErrNoSuchID {
			util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
		} else {
			logger.Error.Printf("update call group condition:%s failed. %s\n", id, err)
			util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		}
	}
}

func checkCallGroupConditionUpdateSet(data model.CallGroupConditionUpdateSet) error {
	cond := model.CallGroupCondition{Name: "valid name", DayRange: 99, DurationMin: 99, DurationMax: 99}
	if data.Name != nil {
		cond.Name = *data.Name
	}
	if data.Description != nil {
		cond.Description = *data.Description
	}
	return checkCallGroupCondition(&cond)
}

func handleDeleteCallGroupCondition(w http.ResponseWriter, r *http.Request) {
	enterprise := requestheader.GetEnterpriseID(r)
	isDelete := 0
	idStr := general.ParseID(r)
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		logger.Error.Printf("invalid cg_condition_id: %s\n", idStr)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
		return
	}
	query := &model.GeneralQuery{ID: []int64{id}, Enterprise: &enterprise, IsDelete: &isDelete}

	_, err = DeleteCallGroupCondition(query)
	if err != nil {
		logger.Error.Printf("delete CallGroupCondition:%s failed. %s\n", idStr, err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}
}

func handleCreateCallGroups(w http.ResponseWriter, r *http.Request) {
	enterprise := requestheader.GetEnterpriseID(r)
	idStr := general.ParseID(r)
	conditionID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		logger.Error.Printf("invalid cg_condition_id: %s\n", idStr)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
		return
	}

	var reqModel model.CallGroupCreateList
	err = util.ReadJSON(r, &reqModel)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
		return
	}

	err = CreateCallGroups(enterprise, conditionID, &reqModel)
	if err != nil {
		logger.Error.Printf("create call group failed. %s\n", err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}
}
