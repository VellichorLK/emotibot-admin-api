package qi

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/gorilla/mux"

	"emotibot.com/emotigo/module/admin-api/ApiError"
	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/admin-api/util/requestheader"
	"emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/module/qic-api/util/general"
	"emotibot.com/emotigo/pkg/logger"
)

type GroupInReq struct {
	UUID            string                `json:"group_id"`
	Name            *string               `json:"group_name"`
	Enabled         *int8                 `json:"is_enable"`
	Speed           *float64              `json:"limit_speed"`
	SlienceDuration *float64              `json:"limit_silence"`
	Rules           []string              `json:"rules"`
	Condition       *model.GroupCondition `json:"other"`
	CreateTime      int64                 `json:"create_time"`
	Description     *string               `json:"description"`
	RuleCount       int                   `json:"rule_count"`
}

func groupInReqToGroupWCond(inreq *GroupInReq) *model.GroupWCond {
	group := &model.GroupWCond{
		UUID:            inreq.UUID,
		Name:            inreq.Name,
		Enabled:         inreq.Enabled,
		Speed:           inreq.Speed,
		SlienceDuration: inreq.SlienceDuration,
		Condition:       inreq.Condition,
		CreateTime:      inreq.CreateTime,
		Description:     inreq.Description,
		RuleCount:       inreq.RuleCount,
	}

	simpleRules := []model.SimpleConversationRule{}
	for _, ruleID := range inreq.Rules {
		simpleRule := model.SimpleConversationRule{
			UUID: ruleID,
		}
		simpleRules = append(simpleRules, simpleRule)
	}
	group.Rules = &simpleRules
	return group
}

func handleCreateGroup(w http.ResponseWriter, r *http.Request) {
	enterprise := requestheader.GetEnterpriseID(r)
	group := model.GroupWCond{}
	err := util.ReadJSON(r, &group)

	group.Enterprise = enterprise

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	createdGroup, err := CreateGroup(&group)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	type GroupResponse struct {
		UUID string `json:"group_id"`
	}

	response := GroupResponse{
		UUID: createdGroup.UUID,
	}

	util.WriteJSON(w, response)
}

func handleGetGroups(w http.ResponseWriter, r *http.Request) {
	values := r.URL.Query()
	filter, err := parseGroupFilter(&values)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	total, groups, err := GetGroupsByFilter(filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := GroupsResponse{
		Paging: &general.Paging{
			Page:  filter.Page,
			Limit: filter.Limit,
			Total: total,
		},
		Data: groups,
	}

	util.WriteJSON(w, response)
}

func parseID(r *http.Request) (id string) {
	vars := mux.Vars(r)
	return vars["id"]
}

func handleGetGroup(w http.ResponseWriter, r *http.Request) {
	id := parseID(r)

	group, err := GetGroupBy(id)
	if err != nil {
		logger.Error.Printf("error while get group in handleGetGroup, reason: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if group == nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	util.WriteJSON(w, group)

}

func handleUpdateGroup(w http.ResponseWriter, r *http.Request) {
	id := parseID(r)
	enterprise := requestheader.GetEnterpriseID(r)

	groupInReq := GroupInReq{}
	err := util.ReadJSON(r, &groupInReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(groupInReq.Rules) > 0 {
		filter := &model.ConversationRuleFilter{
			Enterprise: enterprise,
			Severity:   -1,
			UUID:       groupInReq.Rules,
		}

		total, rules, err := GetConversationRulesBy(filter)
		if err != nil {
			logger.Error.Printf("error while get rules in handleGetConversationRules, reason: %s\n", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if len(rules) != len(groupInReq.Rules) {
			util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, "contains invalid rule"), http.StatusBadRequest)
			return
		}

		ruleIDs := make([]uint64, 0, total)
		for _, v := range rules {
			ruleIDs = append(ruleIDs, uint64(v.ID))
		}

		levValid, err := CheckIntegrity(LevRuleGroup, ruleIDs)
		if err != nil {
			logger.Error.Printf("call check integrity failed. %s\n", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for idx, lev := range levValid {
			if !lev.Valid {
				util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, rules[idx].Name+"("+rules[idx].UUID+") is not complete"), http.StatusBadRequest)
				return
			}
		}
	}

	group := groupInReqToGroupWCond(&groupInReq)
	group.Enterprise = enterprise

	err = UpdateGroup(id, group)
	if err != nil {
		logger.Error.Printf("error while update group in handleUpdateGroup, reason: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func handleDeleteGroup(w http.ResponseWriter, r *http.Request) {
	id := parseID(r)

	err := DeleteGroup(id)

	if err != nil {
		logger.Error.Printf("error while delete group in handleDeleteGroup, reason: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func handleGetGroupsByFilter(w http.ResponseWriter, r *http.Request) {
	values := r.URL.Query()
	filter, err := parseGroupFilter(&values)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	total, groups, err := GetGroupsByFilter(filter)
	if err != nil {
		logger.Error.Printf("error while get groups by filter in handleGetGroupsByFilter, reason: %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := GroupsResponse{
		Paging: &general.Paging{
			Page:  filter.Page,
			Total: total,
			Limit: filter.Limit,
		},
		Data: groups,
	}

	util.WriteJSON(w, response)
}

func parseGroupFilter(values *url.Values) (filter *model.GroupFilter, err error) {
	filter = &model.GroupFilter{}
	filter.FileName = values.Get("file_name")
	filter.Series = values.Get("series")
	filter.StaffID = values.Get("staff_id")
	filter.StaffName = values.Get("staff_name")
	filter.Extension = values.Get("extension")
	filter.Department = values.Get("department")
	filter.CustomerID = values.Get("customer_id")
	filter.CustomerName = values.Get("customer_name")
	filter.CustomerPhone = values.Get("customer_phone")

	deleted := int8(0)
	filter.Delete = &deleted

	dealStr := values.Get("deal")
	if dealStr != "" {
		deal, ierr := strconv.Atoi(dealStr)
		filter.Deal = &deal
		if err != nil {
			return filter, ierr
		}
	}

	callStartStr := values.Get("call_start")
	if callStartStr != "" {
		filter.CallStart, err = strconv.ParseInt(callStartStr, 10, 64)
		if err != nil {
			return
		}
	}

	callEndStr := values.Get("call_end")
	if callEndStr != "" {
		filter.CallEnd, err = strconv.ParseInt(callEndStr, 10, 64)
		if err != nil {
			return
		}
	}

	pageStr := values.Get("page")
	if pageStr != "" {
		filter.Page, err = strconv.Atoi(pageStr)
		if err != nil {
			return
		}
	}

	limitStr := values.Get("limit")
	if limitStr != "" {
		filter.Limit, err = strconv.Atoi(limitStr)
		if err != nil {
			return
		}
	} else {
		filter.Limit = 0
	}
	return
}
