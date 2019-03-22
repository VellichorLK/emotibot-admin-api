package qi

import (
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"emotibot.com/emotigo/module/qic-api/model/v1"
	uuid "github.com/satori/go.uuid"
)

var groupResps = func(filter *model.GroupFilter) (total int64, responses []GroupResp, err error) {
	total, groups, err := GetGroupsByFilter(filter)
	if err != nil {
		return 0, nil, err
	}
	valueQuery := model.UserValueQuery{
		Type: []int8{model.UserValueTypGroup},
	}
	responses = make([]GroupResp, 0, len(groups))
	grpIndexes := make(map[int64]int, len(groups))

	for i, grp := range groups {
		g := GroupResp{
			GroupID:     grp.UUID,
			GroupName:   *grp.Name,
			IsEnable:    *grp.Enabled,
			CreateTime:  grp.CreateTime,
			Description: *grp.Description,
			RuleCount:   grp.RuleCount,
			// Other:       toOther(cond, make(map[string][]interface{})),
		}
		valueQuery.ParentID = append(valueQuery.ParentID, grp.ID)
		responses = append(responses, g)
		grpIndexes[grp.ID] = i
	}
	conditions, err := condDao.Conditions(nil, model.ConditionQuery{
		GroupID: valueQuery.ParentID,
	})
	if err != nil {
		return 0, nil, fmt.Errorf("get conds failed, %v", err)
	}
	for _, cond := range conditions {
		idx := grpIndexes[cond.GroupID]
		resp := responses[idx]
		resp.Other = toOther(&cond, make(map[string][]interface{}))
		responses[idx] = resp
	}
	//query all groups custom values by one time
	grpValues, err := valuesKey(nil, valueQuery)
	if err != nil {
		return 0, nil, fmt.Errorf("get user values failed, %v", err)
	}

	for _, val := range grpValues {
		index, found := grpIndexes[val.LinkID]
		if !found { //ignore corrupt ones.
			continue
		}
		resp := responses[index]
		inputName := val.UserKey.InputName
		resp.Other.CustomColumns[inputName] = append(resp.Other.CustomColumns[inputName], val.Value)
		responses[index] = resp

	}
	return total, responses, nil
}
var newGroupWithAllConditions = func(group model.Group, condition model.Condition, customCols map[string][]interface{}) (model.Group, error) {
	tx, err := dbLike.Begin()
	if err != nil {
		return model.Group{}, fmt.Errorf("begin transaction failed, %v", err)
	}
	defer tx.Rollback()
	if group.UUID == "" {
		newUUID, err := uuid.NewV1()
		if err != nil {
			return model.Group{}, fmt.Errorf("generate uuid failed, %v", err)
		}
		group.UUID = hex.EncodeToString(newUUID[:])
	}
	group, err = newGroup(tx, group)
	if err != nil {
		return model.Group{}, fmt.Errorf("new group failed, %v", err)
	}
	err = setGroupRule(tx, group)
	if err != nil {
		return model.Group{}, fmt.Errorf("set rule relation failed, %v", err)
	}

	condition.GroupID = group.ID
	_, err = newCondition(tx, condition)
	if err != nil {
		return model.Group{}, fmt.Errorf("new condition failed, %v", err)
	}
	_, err = newCustomConditions(tx, group, customCols)
	if err != nil {
		return model.Group{}, fmt.Errorf("new custom column condition failed, %v", err)
	}
	tx.Commit()
	return group, nil
}

func simpleConversationRulesOf(group *model.GroupWCond, sql model.SqlLike) (simpleRules []model.SimpleConversationRule, err error) {
	simpleRules = []model.SimpleConversationRule{}
	if group.Rules == nil || len(*group.Rules) == 0 {
		return
	}

	ruleUUID := make([]string, len(*group.Rules))
	for idx, rule := range *group.Rules {
		ruleUUID[idx] = rule.UUID
	}

	rulefilter := &model.ConversationRuleFilter{
		UUID:       ruleUUID,
		Enterprise: group.Enterprise,
	}

	rules, err := conversationRuleDao.GetBy(rulefilter, sql)
	for _, rule := range rules {
		simpleRule := model.SimpleConversationRule{
			ID:   rule.ID,
			UUID: rule.UUID,
			Name: rule.Name,
		}
		simpleRules = append(simpleRules, simpleRule)
	}
	return
}

// CreateGroup create a group with condition
// ** DEPRECATED API, Using this will not create custom columns, and no data safety is guaranteed**
func CreateGroup(group *model.GroupWCond) (createdGroup *model.GroupWCond, err error) {
	if group == nil || group.Condition == nil {
		return
	}

	uuid, err := uuid.NewV4()
	if err != nil {
		err = fmt.Errorf("error while create uuid in CreateGroup, err: %s", err.Error())
		return
	}

	tx, err := dbLike.Begin()
	if err != nil {
		return
	}
	defer dbLike.ClearTransition(tx)

	if group.Name == nil {
		name := ""
		group.Name = &name
	}

	if group.Enabled == nil {
		enabled := int8(1)
		group.Enabled = &enabled
	}

	if group.Speed == nil {
		speed := float64(0)
		group.Speed = &speed
	}

	if group.SlienceDuration == nil {
		duration := float64(0)
		group.SlienceDuration = &duration
	}

	if group.Description == nil {
		description := ""
		group.Description = &description
	}

	if group.Condition != nil {
		if group.Condition.FileName == nil {
			fileName := ""
			group.Condition.FileName = &fileName
		}

		if group.Condition.CallDuration == nil {
			duration := int64(0)
			group.Condition.CallDuration = &duration
		}

		if group.Condition.CallComment == nil {
			comment := ""
			group.Condition.CallComment = &comment
		}

		if group.Condition.Deal == nil {
			deal := 0
			group.Condition.Deal = &deal
		}

		if group.Condition.Series == nil {
			series := ""
			group.Condition.Series = &series
		}

		if group.Condition.StaffID == nil {
			staffID := ""
			group.Condition.StaffID = &staffID
		}

		if group.Condition.StaffName == nil {
			staffName := ""
			group.Condition.StaffName = &staffName
		}

		if group.Condition.Extension == nil {
			extension := ""
			group.Condition.Extension = &extension
		}

		if group.Condition.Department == nil {
			department := ""
			group.Condition.Department = &department
		}

		if group.Condition.ClientID == nil {
			clientID := ""
			group.Condition.ClientID = &clientID
		}

		if group.Condition.ClientName == nil {
			clientName := ""
			group.Condition.ClientName = &clientName
		}

		if group.Condition.ClientPhone == nil {
			clientPhone := ""
			group.Condition.ClientPhone = &clientPhone
		}

		// TODO: set code left channel & right channel
		if group.Condition.LeftChannelCode == nil {
			leftChannel := 0
			group.Condition.LeftChannelCode = &leftChannel
		}

		if group.Condition.RightChannelCode == nil {
			rightChannel := 1
			group.Condition.RightChannelCode = &rightChannel
		}

		if group.Condition.CallStart == nil {
			callStart := int64(0)
			group.Condition.CallStart = &callStart
		}

		if group.Condition.CallEnd == nil {
			callEnd := int64(0)
			group.Condition.CallEnd = &callEnd
		}
	}

	group.UUID = uuid.String()
	group.UUID = strings.Replace(group.UUID, "-", "", -1)

	createdGroup, err = serviceDAO.CreateGroup(group, tx)
	if err != nil {
		return
	}

	dbLike.Commit(tx)
	return
}

// GetGroupBy Get old group struct by id.
func GetGroupBy(id string) (group *model.GroupWCond, err error) {
	var isDeleted int8 = 0
	filter := &model.GroupFilter{
		UUID: []string{
			id,
		},
		Delete: &isDeleted,
	}

	sqlConn := dbLike.Conn()
	groups, err := serviceDAO.GetGroupsBy(filter, sqlConn)
	if err != nil {
		return
	}

	if len(groups) == 0 {
		return nil, ErrNotFound
	}

	group = &groups[0]

	// TODO: set channel name by code
	staff := "staff"
	client := "client"
	group.Condition.LeftChannel = &staff
	group.Condition.RightChannel = &client
	return
}

// GetGroupsByFilter get old group struct by given filter.
// TODO: Change to the new group struct and query function.
func GetGroupsByFilter(filter *model.GroupFilter) (total int64, groups []model.GroupWCond, err error) {
	tx, err := dbLike.Begin()
	if err != nil {
		return
	}
	defer dbLike.ClearTransition(tx)
	total, err = serviceDAO.CountGroupsBy(filter, tx)
	if err != nil {
		return
	}

	groups, err = serviceDAO.GetGroupsBy(filter, tx)
	staff := "staff"
	client := "client"
	for idx := range groups {
		group := &groups[idx]
		group.Condition.LeftChannel = &staff
		group.Condition.RightChannel = &client
		if group.Rules != nil {
			group.RuleCount = len(*group.Rules)
		}
	}
	return
}

func GetGroupRules(group model.Group) (*model.Group, error) {
	var isDeleted = 0
	rules, otherRules, err := groupRules(nil, group)
	if err != nil {
		return nil, fmt.Errorf("query group rules failed, %v", err)
	}
	filter := &model.ConversationRuleFilter{
		IsDeleted: 0,
	}
	for _, rule := range rules {
		filter.ID = append(filter.ID, uint64(rule))
	}
	group.Rules = make([]model.ConversationRule, 0)
	if len(filter.ID) > 0 {
		_, group.Rules, err = GetConversationRulesBy(filter)
		if err != nil {
			return nil, fmt.Errorf("get rule failed, %v", err)
		}
	}

	group.SilenceRules = make([]model.SilenceRule, 0)
	if len(otherRules[model.GroupRuleTypeSilence]) > 0 {
		rs, err := GetRuleSilences(&model.GeneralQuery{
			UUID:     otherRules[model.GroupRuleTypeSilence],
			IsDelete: &isDeleted,
		}, nil)
		if err != nil {
			return nil, fmt.Errorf("get silence rules failed, %v", err)
		}
		for _, r := range rs {
			group.SilenceRules = append(group.SilenceRules, *r)
		}
	}
	group.SpeedRules = make([]model.SpeedRule, 0)
	if len(otherRules[model.GroupRuleTypeSpeed]) > 0 {
		ruleSpeeds, err := GetRuleSpeeds(&model.GeneralQuery{
			UUID:     otherRules[model.GroupRuleTypeSpeed],
			IsDelete: &isDeleted,
		}, nil)
		if err != nil {
			return nil, fmt.Errorf("get speed rules failed, %v", err)
		}
		for _, r := range ruleSpeeds {
			group.SpeedRules = append(group.SpeedRules, *r)
		}
	}
	group.InterposalRules = make([]model.InterposalRule, 0)
	if len(otherRules[model.GroupRuleTypeInterposal]) > 0 {
		ruleInterposal, err := GetRuleInterposals(&model.GeneralQuery{
			UUID:     otherRules[model.GroupRuleTypeInterposal],
			IsDelete: &isDeleted,
		}, nil)
		if err != nil {
			return nil, fmt.Errorf("get interposal failed, %v", err)
		}
		for _, r := range ruleInterposal {
			group.InterposalRules = append(group.InterposalRules, *r)
		}
	}

	return &group, nil
}

// func UpdateGroupBasic(group model.Group) error {
// 	updateGroup()
// }

// UpdateGroup soft delete the group and create new group & conditions & custom conditions
func UpdateGroup(group model.Group, customcols map[string][]interface{}) (err error) {
	if group.Condition == nil {
		return fmt.Errorf("group require a condition")
	}
	tx, err := dbLike.Begin()
	if err != nil {
		return
	}
	defer tx.Rollback()

	err = serviceDAO.DeleteGroup(group.UUID, tx)
	if err != nil {
		return fmt.Errorf("delete group failed, %v", err)
	}
	group.UpdatedTime = time.Now().Unix()
	group, err = newGroup(tx, group)
	if err != nil {
		return
	}
	err = resetGroupRules(tx, group)
	if err != nil {
		return fmt.Errorf("reset group rule failed, %v", err)
	}
	err = setGroupRule(tx, group)
	if err != nil {
		return fmt.Errorf("update group rule failed, %v", err)
	}
	group.Condition.GroupID = group.ID
	_, err = newCondition(tx, *group.Condition)
	if err != nil {
		return
	}
	_, err = newCustomConditions(tx, group, customcols)
	if err != nil {
		return
	}
	// Set Rules
	// simpleRules
	tx.Commit()
	return
}

// DeleteGroup soft delete the specify group by UUID,
// It is caller decision to delete group's Conditions or uservalues or not.
// Since these will become nonreachable by group.
func DeleteGroup(id string) (err error) {
	tx, err := dbLike.Begin()
	if err != nil {
		return
	}
	defer dbLike.ClearTransition(tx)

	err = serviceDAO.DeleteGroup(id, tx)
	if err != nil {
		return
	}

	err = dbLike.Commit(tx)
	return
}

// NewGroupWithAllConditions create a group with condition and all sort of custom columns by UserValue
func NewGroupWithAllConditions(group model.Group, condition model.Condition, customCols map[string][]interface{}) (model.Group, error) {
	return newGroupWithAllConditions(group, condition, customCols)
}

// GroupResp is the response schema of get group
type GroupResp struct {
	GroupID     string `json:"group_id"`
	GroupName   string `json:"group_name"`
	IsEnable    int8   `json:"is_enable"`
	Other       Other  `json:"other"`
	CreateTime  int64  `json:"create_time"`
	Description string `json:"description"`
	RuleCount   int    `json:"rule_count"`
}

func GroupResps(filter *model.GroupFilter) (total int64, responses []GroupResp, err error) {
	return groupResps(filter)
}
