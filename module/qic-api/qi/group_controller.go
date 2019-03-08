package qi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"emotibot.com/emotigo/module/admin-api/util/AdminErrors"

	"github.com/gorilla/mux"

	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/admin-api/util/requestheader"
	"emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/module/qic-api/util/general"
	"emotibot.com/emotigo/pkg/logger"
)

// NewGroupReq is the request body schema of the POST or PUT group api
type NewGroupReq struct {
	GroupName    string   `json:"group_name"`
	IsEnable     int8     `json:"is_enable"`
	LimitSpeed   int      `json:"limit_speed"`
	LimitSilence float32  `json:"limit_silence"`
	Other        Other    `json:"other"`
	Rules        []string `json:"rules"`
}

//Group transfer NewGroupReq as a model.Group struct, any virtual fields should be handled by the caller.
func (n *NewGroupReq) Group() model.Group {
	return model.Group{
		Name:           n.GroupName,
		IsEnable:       n.IsEnable != 0,
		LimitedSpeed:   n.LimitSpeed,
		LimitedSilence: n.LimitSilence,
	}
}

//Other is the condition's json response including custom conditions.
type Other struct {
	Type          int8                     `json:"type"` // it is the ConditionType
	FileName      string                   `json:"file_name"`
	CallTime      int64                    `json:"call_time"`
	CallComment   string                   `json:"call_comment"`
	Deal          int8                     `json:"deal"`
	Series        string                   `json:"series"`
	StaffID       string                   `json:"staff_id"`
	StaffName     string                   `json:"staff_name"`
	Extension     string                   `json:"extension"`
	Department    string                   `json:"department"`
	CustomerID    string                   `json:"customer_id"`
	CustomerName  string                   `json:"customer_name"`
	CustomerPhone string                   `json:"customer_phone"`
	LeftChannel   string                   `json:"left_channel"`
	RightChannel  string                   `json:"right_channel"`
	CallFrom      int64                    `json:"call_from"`
	CallEnd       int64                    `json:"call_end"`
	CustomColumns map[string][]interface{} `json:"-"`
}

// UnmarshalJSON unmarshal Other with additional custom columns
func (o *Other) UnmarshalJSON(data []byte) error {
	// Check NewCallReq UnmarshalJSON
	type Alias Other
	a := &struct {
		*Alias
	}{
		Alias: (*Alias)(o),
	}
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}

	columns := map[string]interface{}{}
	if err := json.Unmarshal(data, &columns); err != nil {
		return err
	}
	o.CustomColumns = map[string][]interface{}{}
	for col, val := range columns {
		if _, exist := ReservedCustomKeywords[col]; exist {
			continue
		}
		o.CustomColumns[col] = append(o.CustomColumns[col], val)
	}
	return nil
}

// MarshalJSON Other will flatten its CustomColumns map with other fields.
func (o Other) MarshalJSON() ([]byte, error) {
	resp := map[string]interface{}{}
	v := reflect.ValueOf(o)
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		tag := t.Field(i).Tag.Get("json")
		name, opt := getJSONName(tag)
		if name == "-" {
			continue
		}
		if strings.Contains(opt, "omitempty") {
			f := v.Field(i)
			switch f.Kind() {
			case reflect.String:
				if f.String() == "" {
					continue
				}
			case reflect.Float64, reflect.Float32:
				if f.Float() == 0 {
					continue
				}
			case reflect.Int64, reflect.Int32, reflect.Int8:
				if f.Int() == 0 {
					continue
				}
			case reflect.Slice, reflect.Array, reflect.Map:
				if f.IsNil() {
					continue
				}
			}
		}

		resp[name] = v.Field(i).Interface()
	}
	for colName, val := range o.CustomColumns {
		if _, exist := resp[colName]; exist {
			return nil, fmt.Errorf("custom column %s is overlapped with require column", colName)
		}
		resp[colName] = val
	}
	return json.Marshal(resp)
}

var conditionTypDict = map[int8]struct{}{
	model.GroupCondTypOn: struct{}{},
}

// ValidcondType Return Condition type code(int8) by given input name.
// If none is matched then GroupCondTypOn will return.
func IsValidcondType(typ int8) bool {
	_, exist := conditionTypDict[typ]
	return exist
}

func (o *Other) ToCondition() *model.Condition {

	return &model.Condition{
		Type:          o.Type,
		FileName:      o.FileName,
		Deal:          o.Deal,
		Series:        o.Series,
		StaffID:       o.StaffID,
		StaffName:     o.StaffName,
		Extension:     o.Extension,
		Department:    o.Department,
		CustomerID:    o.CustomerID,
		CustomerName:  o.CustomerName,
		CustomerPhone: o.CustomerPhone,
		LeftChannel:   int8(RoleMatcherTyp(o.LeftChannel)),
		RightChannel:  int8(RoleMatcherTyp(o.RightChannel)),
		CallStart:     o.CallFrom,
		CallEnd:       o.CallEnd,
	}
}

// func groupInReqToGroupWCond(inreq *GroupInReq) *model.GroupWCond {
// 	group := &model.GroupWCond{
// 		UUID:            inreq.UUID,
// 		Name:            inreq.Name,
// 		Enabled:         inreq.Enabled,
// 		Speed:           inreq.Speed,
// 		SlienceDuration: inreq.SlienceDuration,
// 		Condition:       inreq.Condition.ToGroupCondition(),
// 		CreateTime:      inreq.CreateTime,
// 		Description:     inreq.Description,
// 		RuleCount:       inreq.RuleCount,
// 	}

// 	simpleRules := []model.SimpleConversationRule{}
// 	for _, ruleID := range inreq.Rules {
// 		simpleRule := model.SimpleConversationRule{
// 			UUID: ruleID,
// 		}
// 		simpleRules = append(simpleRules, simpleRule)
// 	}
// 	group.Rules = &simpleRules
// 	return group
// }

func handleCreateGroup(w http.ResponseWriter, r *http.Request) {
	var (
		reqBody   NewGroupReq
		ruleTotal int64
	)
	err := util.ReadJSON(r, &reqBody)
	if err != nil {
		util.ReturnError(w, AdminErrors.ErrnoRequestError, fmt.Sprintf("Bad Request Body, %v", err))
		return
	}
	if !IsValidcondType(reqBody.Other.Type) {
		util.ReturnError(w, AdminErrors.ErrnoRequestError, "require Other Condition")
		return
	}

	group := reqBody.Group()
	group.EnterpriseID = requestheader.GetEnterpriseID(r)
	ruleTotal, group.Rules, err = GetConversationRulesBy(&model.ConversationRuleFilter{
		UUID: reqBody.Rules,
	})
	if err != nil {
		util.ReturnError(w, AdminErrors.ErrnoRequestError, fmt.Sprintf("get rules failed, %v", err))
		return
	}
	if int(ruleTotal) != len(reqBody.Rules) {
		util.ReturnError(w, AdminErrors.ErrnoRequestError, fmt.Sprintf("some input rules invalid"))
		return
	}

	condition := reqBody.Other.ToCondition()
	customConditions := reqBody.Other.CustomColumns
	group, err = NewGroupWithAllConditions(group, *condition, customConditions)
	if err != nil {
		util.ReturnError(w, AdminErrors.ErrnoDBError, fmt.Sprintf("internal server error, %v", err))
		return
	}
	response := struct {
		UUID string `json:"group_id"`
	}{
		UUID: group.UUID,
	}

	util.WriteJSON(w, response)
}

func handleGetGroups(w http.ResponseWriter, r *http.Request) {
	type GroupsResponse struct {
		Paging general.Paging `json:"paging"`
		Data   []GroupResp    `json:"data"`
	}
	values := r.URL.Query()
	filter, err := parseGroupFilter(&values)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	total, groups, err := GroupResps(filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := GroupsResponse{
		Paging: general.Paging{
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

func handleUpdateGroup(w http.ResponseWriter, r *http.Request, group *model.Group) {
	var reqBody NewGroupReq
	err := util.ReadJSON(r, &reqBody)
	if err != nil {
		util.ReturnError(w, AdminErrors.ErrnoRequestError, fmt.Sprintf("Bad Request Body, %v", err))
		return
	}
	if !IsValidcondType(reqBody.Other.Type) {
		util.ReturnError(w, AdminErrors.ErrnoRequestError, "require Other Condition")
		return
	}
	newGroup := reqBody.Group()
	newGroup.EnterpriseID = group.EnterpriseID
	newGroup.ID = group.ID
	newGroup.Condition = reqBody.Other.ToCondition()
	customConditions := reqBody.Other.CustomColumns
	total, rules, err := GetConversationRulesBy(&model.ConversationRuleFilter{
		UUID: reqBody.Rules,
	})
	if int(total) != len(rules) {
		util.ReturnError(w, AdminErrors.ErrnoRequestError, fmt.Sprintf("some rules does not exist."))
		return
	}
	err = UpdateGroup(newGroup, customConditions)
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

	total, groups, err := GroupResps(filter)
	if err != nil {
		logger.Error.Printf("error while get groups by filter in handleGetGroupsByFilter, reason: %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	type GroupsResponse struct {
		Paging general.Paging `json:"paging"`
		Data   []GroupResp    `json:"data"`
	}
	response := GroupsResponse{
		Paging: general.Paging{
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

func groupRequest(next func(w http.ResponseWriter, r *http.Request, group *model.Group)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		const GroupIDKey = "group_id"
		groupUUID := mux.Vars(r)[GroupIDKey]
		if GroupIDKey == "" {
			util.ReturnError(w, AdminErrors.ErrnoRequestError, fmt.Sprintf("require %s in path", GroupIDKey))
			return
		}
		enterprise := requestheader.GetEnterpriseID(r)
		if enterprise == "" {
			util.ReturnError(w, AdminErrors.ErrnoRequestError, fmt.Sprintf("empty enterprise ID"))
			return
		}
		groups, err := serviceDAO.Group(nil, model.GroupQuery{UUID: []string{groupUUID}})
		if err != nil {
			util.ReturnError(w, AdminErrors.ErrnoRequestError, fmt.Sprintf("get group failed, %v", err))
		}
		if len(groups) == 0 {
			util.ReturnError(w, AdminErrors.ErrnoRequestError, fmt.Sprintf("id '%s' is not exist", groupUUID))
			return
		}
		next(w, r, &groups[0])
	}
}
