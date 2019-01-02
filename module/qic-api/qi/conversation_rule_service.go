package qi

import (
	"emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/pkg/logger"
	"fmt"
	"github.com/satori/go.uuid"
	"strings"
	"time"
)

var (
	ErrNilConversationRule = fmt.Errorf("conversation rule can not be nil")
)

var conversationRuleDao model.ConversationRuleDao = &model.ConversationRuleSqlDaoImpl{}

func simpleConversationFlowsOf(rule *model.ConversationRule, sql model.SqlLike) (simpleFlows []model.SimpleConversationFlow, err error) {
	uuids := make([]string, len(rule.Flows))
	for idx, _ := range rule.Flows {
		uuids[idx] = rule.Flows[idx].UUID
	}

	var isDelete int8 = int8(0)
	filter := &model.ConversationFlowFilter{
		Enterprise: rule.Enterprise,
		UUID:       uuids,
		IsDelete:   isDelete,
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

	err = conversationRuleDao.Delete(id, tx)
	if err != nil {
		return
	}

	rule.Flows = flows
	rule.UUID = id
	rule.CreateTime = originRule.CreateTime
	rule.UpdateTime = time.Now().Unix()

	updatedRule, err = conversationRuleDao.Create(rule, tx)
	logger.Info.Printf("updatedRule: %+v\n", updatedRule)
	if err != nil {
		return
	}
	dbLike.Commit(tx)
	return
}

func DeleteConversationRule(id string) (err error) {
	return conversationRuleDao.Delete(id, sqlConn)
}