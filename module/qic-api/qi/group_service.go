package qi

import (
	"fmt"
	"strings"

	"emotibot.com/emotigo/module/qic-api/model/v1"
	"github.com/satori/go.uuid"
)

var (
	serviceDAO model.GroupDAO = &model.GroupSQLDao{}
)

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

	simpleRules, err := simpleConversationRulesOf(group, tx)
	if err != nil {
		return
	}
	group.Rules = &simpleRules

	createdGroup, err = serviceDAO.CreateGroup(group, tx)
	if err != nil {
		return
	}

	dbLike.Commit(tx)
	return
}

func GetGroupBy(id string) (group *model.GroupWCond, err error) {
	filter := &model.GroupFilter{
		UUID: []string{
			id,
		},
	}

	sqlConn := dbLike.Conn()
	groups, err := serviceDAO.GetGroupsBy(filter, sqlConn)
	if err != nil {
		return
	}

	if len(groups) == 0 {
		return
	}

	group = &groups[0]

	// TODO: set channel name by code
	staff := "staff"
	client := "client"
	group.Condition.LeftChannel = &staff
	group.Condition.RightChannel = &client
	return
}

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

func UpdateGroup(id string, group *model.GroupWCond) (err error) {
	tx, err := dbLike.Begin()
	if err != nil {
		return
	}
	defer dbLike.ClearTransition(tx)

	// get original group to compare which fileds are need to be updated
	filter := &model.GroupFilter{
		UUID: []string{
			id,
		},
		EnterpriseID: group.Enterprise,
	}

	groups, err := serviceDAO.GetGroupsBy(filter, tx)
	if err != nil {
		return
	}

	if len(groups) == 0 {
		return
	}
	originGroup := &groups[0]

	err = serviceDAO.DeleteGroup(id, tx)
	if err != nil {
		return
	}

	if group.Name == nil {
		group.Name = originGroup.Name
	}

	if group.Enabled == nil {
		group.Enabled = originGroup.Enabled
	}

	if group.Speed == nil {
		group.Speed = originGroup.Speed
	}

	if group.SlienceDuration == nil {
		group.SlienceDuration = originGroup.SlienceDuration
	}

	if group.Description == nil {
		group.Description = originGroup.Description
	}

	if group.Condition != nil {
		if group.Condition.FileName == nil {
			group.Condition.FileName = originGroup.Condition.FileName
		}

		if group.Condition.CallDuration == nil {
			group.Condition.CallDuration = originGroup.Condition.CallDuration
		}

		if group.Condition.CallComment == nil {
			group.Condition.CallComment = originGroup.Condition.CallComment
		}

		if group.Condition.Deal == nil {
			group.Condition.Deal = originGroup.Condition.Deal
		}

		if group.Condition.Series == nil {
			group.Condition.Series = originGroup.Condition.Series
		}

		if group.Condition.StaffID == nil {
			group.Condition.StaffID = originGroup.Condition.StaffID
		}

		if group.Condition.StaffName == nil {
			group.Condition.StaffName = originGroup.Condition.StaffName
		}

		if group.Condition.Extension == nil {
			group.Condition.Extension = originGroup.Condition.Extension
		}

		if group.Condition.Department == nil {
			group.Condition.Department = originGroup.Condition.Department
		}

		if group.Condition.ClientID == nil {
			group.Condition.ClientID = originGroup.Condition.ClientID
		}

		if group.Condition.ClientName == nil {
			group.Condition.ClientName = originGroup.Condition.ClientName
		}

		if group.Condition.ClientPhone == nil {
			group.Condition.ClientPhone = originGroup.Condition.ClientPhone
		}

		if group.Condition.LeftChannelCode == nil {
			group.Condition.LeftChannelCode = originGroup.Condition.LeftChannelCode
		}

		if group.Condition.RightChannelCode == nil {
			group.Condition.RightChannelCode = originGroup.Condition.RightChannelCode
		}

		if group.Condition.CallStart == nil {
			group.Condition.CallStart = originGroup.Condition.CallStart
		}

		if group.Condition.CallEnd == nil {
			group.Condition.CallEnd = originGroup.Condition.CallEnd
		}
	} else {
		group.Condition = originGroup.Condition
	}

	simpleRules, err := simpleConversationRulesOf(group, tx)
	if err != nil {
		return
	}

	group.Rules = &simpleRules
	group.CreateTime = originGroup.CreateTime
	group.UUID = id
	group.Enterprise = group.Enterprise
	_, err = serviceDAO.CreateGroup(group, tx)
	if err != nil {
		return
	}

	err = dbLike.Commit(tx)
	return
}

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
