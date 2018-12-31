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
