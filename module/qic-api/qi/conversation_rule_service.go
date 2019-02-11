package qi

import (
	"fmt"
	"strings"
	"time"

	"emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/pkg/logger"
	uuid "github.com/satori/go.uuid"
)

var (
	ErrNilConversationRule = fmt.Errorf("conversation rule can not be nil")
)

var conversationRuleDao model.ConversationRuleDao = &model.ConversationRuleSqlDaoImpl{}

func simpleConversationFlowsOf(rule *model.ConversationRule, sql model.SqlLike) (simpleFlows []model.SimpleConversationFlow, err error) {
	simpleFlows = []model.SimpleConversationFlow{}
	if len(rule.Flows) == 0 {
		return
	}
	uuids := make([]string, len(rule.Flows))
	for idx, _ := range rule.Flows {
		uuids[idx] = rule.Flows[idx].UUID
	}

	var isDelete int8 = int8(0)
	filter := &model.ConversationFlowFilter{
		Enterprise: rule.Enterprise,
		UUID:       uuids,
		IsDelete:   &isDelete,
	}

	flows, err := conversationFlowDao.GetBy(filter, sql)
	if err != nil {
		return
	}

	if len(flows) != len(rule.Flows) {
		logger.Warn.Print("number of input flows does not match number of flows in db")
	}

	simpleFlows = make([]model.SimpleConversationFlow, len(flows))
	for idx := range flows {
		simpleFlow := model.SimpleConversationFlow{
			ID:   flows[idx].ID,
			UUID: flows[idx].UUID,
			Name: flows[idx].Name,
		}
		simpleFlows[idx] = simpleFlow
	}
	logger.Info.Printf("simpleFlows: %+v\n", simpleFlows)
	return
}

func CreateConversationRule(rule *model.ConversationRule) (createdRule *model.ConversationRule, err error) {
	if rule == nil {
		err = ErrNilConversationRule
		return
	}

	// create uuid for the new flow
	uuid, err := uuid.NewV4()
	if err != nil {
		err = fmt.Errorf("error while create uuid in CreateGroup, err: %s", err.Error())
		return
	}
	rule.UUID = uuid.String()
	rule.UUID = strings.Replace(rule.UUID, "-", "", -1)

	tx, err := dbLike.Begin()
	if err != nil {
		return
	}
	defer dbLike.ClearTransition(tx)

	simpleFlows, err := simpleConversationFlowsOf(rule, tx)
	if err != nil {
		return
	}
	rule.Flows = simpleFlows

	now := time.Now().Unix()
	rule.CreateTime = now
	rule.UpdateTime = now

	createdRule, err = conversationRuleDao.Create(rule, tx)
	if err != nil {
		return
	}

	dbLike.Commit(tx)
	return
}

func GetConversationRulesBy(filter *model.ConversationRuleFilter) (total int64, rules []model.ConversationRule, err error) {
	if filter == nil {
		return
	}

	tx, err := dbLike.Begin()
	if err != nil {
		return
	}

	total, err = conversationRuleDao.CountBy(filter, tx)
	if err != nil {
		return
	}

	rules, err = conversationRuleDao.GetBy(filter, tx)
	if err != nil {
		return
	}
	return
}

func UpdateConversationRule(id string, rule *model.ConversationRule) (updatedRule *model.ConversationRule, err error) {
	if rule == nil {
		err = ErrNilConversationRule
		return
	}

	tx, err := dbLike.Begin()
	if err != nil {
		return
	}
	defer dbLike.ClearTransition(tx)

	filter := &model.ConversationRuleFilter{
		UUID: []string{
			id,
		},
		Enterprise: rule.Enterprise,
		Severity:   -1,
	}
	rules, err := conversationRuleDao.GetBy(filter, tx)
	if err != nil {
		return
	}

	if len(rules) == 0 {
		return
	}

	originRule := rules[0]

	flows, err := simpleConversationFlowsOf(rule, tx)
	if err != nil {
		return
	}

	// get groups which include this rule
	groups, err := serviceDAO.GetGroupsByRuleID([]int64{originRule.ID}, tx)
	if err != nil {
		return
	}

	err = conversationRuleDao.Delete(id, tx)
	if err != nil {
		return
	}

	rule.Flows = flows
	rule.UUID = id
	rule.CreateTime = originRule.CreateTime
	rule.UpdateTime = time.Now().Unix()

	updatedRule, err = conversationRuleDao.Create(rule, tx)
	if err != nil {
		return
	}

	if len(groups) > 0 {
		err = propagateUpdateFromGroup(groups, []model.ConversationRule{*updatedRule}, tx)
		if err != nil {
			return
		}
	}
	err = dbLike.Commit(tx)
	return
}

func DeleteConversationRule(id string) (err error) {
	tx, err := dbLike.Begin()
	if err != nil {
		return
	}
	defer dbLike.ClearTransition(tx)

	filter := &model.ConversationRuleFilter{
		UUID: []string{
			id,
		},
		IsDeleted: 0,
	}
	rules, err := conversationRuleDao.GetBy(filter, tx)
	if err != nil {
		return
	}

	if len(rules) == 0 {
		return
	}

	rule := rules[0]

	groups, err := serviceDAO.GetGroupsByRuleID([]int64{rule.ID}, tx)
	if err != nil {
		return
	}

	err = conversationRuleDao.Delete(id, tx)
	if err != nil {
		return
	}

	if len(groups) > 0 {
		// remove the rule from groups which are related to this rule
		for idx := range groups {
			group := &groups[idx]
			rules := *group.Rules
			if len(rules) == 0 {
				newRules := []model.SimpleConversationRule{}
				group.Rules = &newRules
				continue
			}

			for j, rule := range rules {
				if rule.UUID == id {
					var newRules []model.SimpleConversationRule
					if j == len(rules)-1 {
						newRules = rules[:j]
					} else {
						newRules = append(rules[:j], rules[j+1:]...)
					}
					group.Rules = &newRules
				}
			}
		}

		err = propagateUpdateFromGroup(groups, rules, tx)
		if err != nil {
			return
		}
	}

	dbLike.Commit(tx)
	return
}

func propagateUpdateFromGroup(groups []model.GroupWCond, rules []model.ConversationRule, sqlLike model.SqlLike) error {
	logger.Info.Printf("groups: %+v", groups)
	logger.Info.Printf("rules: %+v", rules)
	var err error
	if len(groups) == 0 {
		return err
	}

	// create rule map
	ruleMap := map[string]int64{}
	for _, rule := range rules {
		ruleMap[rule.UUID] = rule.ID
	}

	groupUUID := []string{}
	activeGroups := []model.GroupWCond{}
	for i := range groups {
		group := &groups[i]
		if group.Deleted == 1 {
			continue
		}

		rules := *group.Rules
		for j := range rules {
			rule := &rules[j]
			if ruleID, ok := ruleMap[rule.UUID]; ok {
				rule.ID = ruleID
			}
		}
		groupUUID = append(groupUUID, group.UUID)
		activeGroups = append(activeGroups, *group)
	}

	err = serviceDAO.DeleteMany(groupUUID, sqlLike)
	if err != nil {
		return err
	}

	err = serviceDAO.CreateMany(activeGroups, sqlLike)
	if err != nil {
		return err
	}

	return err
}
