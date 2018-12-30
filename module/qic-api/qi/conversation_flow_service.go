package qi

import (
	"emotibot.com/emotigo/module/qic-api/model/v1"
	_ "emotibot.com/emotigo/pkg/logger"
	"fmt"
	"github.com/satori/go.uuid"
	"strings"
	"time"
)

var (
	ErrNilFlow = fmt.Errorf("flow can not be nil")
)

var conversationFlowDao model.ConversationFlowDao = &model.ConversationFlowSqlDaoImpl{}

func simpleSentenceGroupsOf(flow *model.ConversationFlow) (groups []model.SimpleSentenceGroup, err error) {
	uuids := make([]string, len(flow.SentenceGroups))
	for idx, _ := range flow.SentenceGroups {
		uuids[idx] = flow.SentenceGroups[idx].UUID
	}

	var isDelete int8 = int8(0)
	filter := &model.SentenceGroupFilter{
		Enterprise: flow.Enterprise,
		UUID:       uuids,
		IsDelete:   isDelete,
		Role:       -1,
		Position:   -1,
	}

	sentenceGroups, err := sentenceGroupDao.GetBy(filter, sqlConn)
	if err != nil {
		return
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
	return
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

	simpleGroups, err := simpleSentenceGroupsOf(flow)
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

func DeleteConversationFlow(id string) (err error) {
	return conversationFlowDao.Delete(id, sqlConn)
}
