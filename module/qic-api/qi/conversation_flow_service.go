package qi

import (
	"fmt"
	"strings"
	"time"

	"emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/pkg/logger"
	"github.com/satori/go.uuid"
)

var (
	ErrNilFlow = fmt.Errorf("flow can not be nil")
)

var conversationFlowDao model.ConversationFlowDao = &model.ConversationFlowSqlDaoImpl{}

func simpleSentenceGroupsOf(flow *model.ConversationFlow, sql model.SqlLike) ([]model.SimpleSentenceGroup, error) {
	groups := []model.SimpleSentenceGroup{}
	var err error
	if len(flow.SentenceGroups) > 0 {
		uuids := make([]string, len(flow.SentenceGroups))
		for idx, _ := range flow.SentenceGroups {
			uuids[idx] = flow.SentenceGroups[idx].UUID
		}

		var isDelete int8 = int8(0)
		filter := &model.SentenceGroupFilter{
			Enterprise: flow.Enterprise,
			UUID:       uuids,
			IsDelete:   &isDelete,
		}

		sentenceGroups, err := sentenceGroupDao.GetBy(filter, sql)
		if err != nil {
			return groups, err
		}
		groups = make([]model.SimpleSentenceGroup, len(sentenceGroups))
		for idx := range sentenceGroups {
			simpleGroup := model.SimpleSentenceGroup{
				ID:   sentenceGroups[idx].ID,
				UUID: sentenceGroups[idx].UUID,
				Name: sentenceGroups[idx].Name,
			}
			groups[idx] = simpleGroup
		}
	}
	return groups, err
}

func CreateConversationFlow(flow *model.ConversationFlow) (createdFlow *model.ConversationFlow, err error) {
	if flow == nil {
		err = ErrNilFlow
		return
	}

	// create uuid for the new flow
	uuid, err := uuid.NewV4()
	if err != nil {
		err = fmt.Errorf("error while create uuid in CreateGroup, err: %s", err.Error())
		return
	}
	flow.UUID = uuid.String()
	flow.UUID = strings.Replace(flow.UUID, "-", "", -1)

	// create conversation flow
	tx, err := dbLike.Begin()
	if err != nil {
		return
	}
	defer dbLike.ClearTransition(tx)

	simpleGroups, err := simpleSentenceGroupsOf(flow, tx)
	if err != nil {
		return
	}
	flow.SentenceGroups = simpleGroups

	now := time.Now().Unix()
	flow.CreateTime = now
	flow.UpdateTime = now

	createdFlow, err = conversationFlowDao.Create(flow, tx)
	if err != nil {
		return
	}

	err = dbLike.Commit(tx)
	return
}

func GetConversationFlowsBy(filter *model.ConversationFlowFilter) (total int64, flows []model.ConversationFlow, err error) {
	total, err = conversationFlowDao.CountBy(filter, sqlConn)
	if err != nil {
		return
	}

	flows, err = conversationFlowDao.GetBy(filter, sqlConn)
	if err != nil {
		return
	}
	return
}

func UpdateConversationFlow(id, enterprise string, flow *model.ConversationFlow) (updatedFlow *model.ConversationFlow, err error) {
	if flow == nil {
		err = ErrNilFlow
		return
	}

	tx, err := dbLike.Begin()
	if err != nil {
		return
	}
	defer dbLike.ClearTransition(tx)

	filter := &model.ConversationFlowFilter{
		UUID: []string{
			id,
		},
		Enterprise: enterprise,
		IsDelete:   0,
	}

	flows, err := conversationFlowDao.GetBy(filter, tx)
	if err != nil {
		return
	}

	if len(flows) == 0 {
		return
	}

	originFlow := flows[0]
	rules, err := conversationRuleDao.GetByFlowID([]int64{originFlow.ID}, tx)
	if err != nil {
		return
	}

	err = conversationFlowDao.Delete(id, tx)
	if err != nil {
		return
	}

	simpleGroups, err := simpleSentenceGroupsOf(flow, tx)
	if err != nil {
		return
	}

	flow.UUID = id
	flow.SentenceGroups = simpleGroups
	flow.CreateTime = originFlow.CreateTime
	flow.UpdateTime = time.Now().Unix()

	updatedFlow, err = conversationFlowDao.Create(flow, tx)
	if err != nil {
		return
	}

	ruleUUID := make([]string, len(rules))
	for i := 0; i < len(rules); i++ {
		ruleUUID[i] = rules[i].UUID
	}

	err = propagateUpdateFromRule(rules, []model.ConversationFlow{*updatedFlow}, tx)
	if err != nil {
		return
	}

	err = dbLike.Commit(tx)
	return
}

func DeleteConversationFlow(id string) (err error) {
	return conversationFlowDao.Delete(id, sqlConn)
}

func propagateUpdateFromRule(rules []model.ConversationRule, flows []model.ConversationFlow, sqlLike model.SqlLike) (err error) {
	if len(rules) == 0 || len(flows) == 0 {
		return
	}

	// create flow map
	flowMap := map[string]int64{}
	for _, flow := range flows {
		flowMap[flow.UUID] = flow.ID
	}

	for i := range rules {
		rule := &rules[i]
		for j, flow := range rule.Flows {
			if flowID, ok := flowMap[flow.UUID]; ok {
				rule.Flows[j].ID = flowID
			}
		}
	}

	ruleUUID := make([]string, len(rules))
	for idx := range rules {
		ruleUUID[idx] = rules[idx].UUID
	}

	ruleID := []int64{}
	for _, rule := range rules {
		ruleID = append(ruleID, rule.ID)
	}

	groups, err := serviceDAO.GetGroupsByRuleID(ruleID, sqlLike)
	if err != nil {
		return
	}

	err = conversationRuleDao.DeleteMany(ruleUUID, sqlLike)
	if err != nil {
		return
	}

	err = conversationRuleDao.CreateMany(rules, sqlLike)
	if err != nil {
		return
	}

	ruleFilter := &model.ConversationRuleFilter{
		UUID:      ruleUUID,
		IsDeleted: int8(0),
	}
	rules, err = conversationRuleDao.GetBy(ruleFilter, sqlLike)
	if err != nil {
		return
	}

	err = propagateUpdateFromGroup(groups, rules, sqlLike)
	return
}
