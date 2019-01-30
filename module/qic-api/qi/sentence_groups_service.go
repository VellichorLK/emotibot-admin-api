package qi

import (
	"database/sql"
	"emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/pkg/logger"
	"fmt"
	"github.com/satori/go.uuid"
	"strings"
	"time"
)

var (
	ErrNilSentenceGroup = fmt.Errorf("Sentence can not be nil")
)

var sentenceGroupDao model.SentenceGroupsSqlDao = &model.SentenceGroupsSqlDaoImpl{}

func simpleSentencesOf(group *model.SentenceGroup, tx *sql.Tx) (simpleSentences []model.SimpleSentence, err error) {
	simpleSentences = []model.SimpleSentence{}
	if len(group.Sentences) == 0 {
		return
	}

	simpleSentences = make([]model.SimpleSentence, len(group.Sentences))
	uuids := make([]string, len(group.Sentences))
	for idx, s := range group.Sentences {
		uuids[idx] = s.UUID
	}

	var isDelete int8 = int8(0)
	query := &model.SentenceQuery{
		Enterprise: &group.Enterprise,
		UUID:       uuids,
		IsDelete:   &isDelete,
	}

	sentences, err := sentenceDao.GetSentences(tx, query)
	if err != nil {
		return
	}

	if len(sentences) != len(group.Sentences) {
		logger.Warn.Printf("user input sentences does not match sentences in db")
	}

	for idx, s := range sentences {
		simpleSentence := model.SimpleSentence{
			ID:   s.ID,
			UUID: s.UUID,
			Name: s.Name,
		}
		simpleSentences[idx] = simpleSentence
	}
	return

}

func CreateSentenceGroup(group *model.SentenceGroup) (createdGroup *model.SentenceGroup, err error) {
	if group == nil {
		err = ErrNilSentenceGroup
		return
	}

	// create uuid for the new group
	uuid, err := uuid.NewV4()
	if err != nil {
		err = fmt.Errorf("error while create uuid in CreateGroup, err: %s", err.Error())
		return
	}
	group.UUID = uuid.String()
	group.UUID = strings.Replace(group.UUID, "-", "", -1)

	// create group
	tx, err := dbLike.Begin()
	if err != nil {
		return
	}
	defer dbLike.ClearTransition(tx)

	simpleSentences, err := simpleSentencesOf(group, tx)
	if err != nil {
		return
	}
	group.Sentences = simpleSentences

	now := time.Now().Unix()
	group.CreateTime = now
	group.UpdateTime = now

	createdGroup, err = sentenceGroupDao.Create(group, tx)
	err = dbLike.Commit(tx)
	return
}

func GetSentenceGroupsBy(filter *model.SentenceGroupFilter) (total int64, groups []model.SentenceGroup, err error) {
	total, err = sentenceGroupDao.CountBy(filter, sqlConn)
	if err != nil {
		return
	}

	groups, err = sentenceGroupDao.GetBy(filter, sqlConn)
	return
}

func UpdateSentenceGroup(uuid string, group *model.SentenceGroup) (updatedGroup *model.SentenceGroup, err error) {
	tx, err := dbLike.Begin()
	if err != nil {
		return
	}
	defer dbLike.ClearTransition(tx)

	deleted := int8(0)
	filter := &model.SentenceGroupFilter{
		UUID: []string{
			uuid,
		},
		Enterprise: group.Enterprise,
		IsDelete:   &deleted,
	}

	groups, err := sentenceGroupDao.GetBy(filter, tx)
	if err != nil {
		return
	}

	if len(groups) == 0 {
		return
	}

	oldGroup := groups[0]

	// fetch flow before disable old sentence group
	flows, err := conversationFlowDao.GetBySentenceGroupID([]int64{oldGroup.ID}, tx)
	if err != nil {
		return
	}

	err = sentenceGroupDao.Delete(uuid, tx)
	if err != nil {
		return
	}

	simpleSentences, err := simpleSentencesOf(group, tx)
	if err != nil {
		return
	}

	group.Sentences = simpleSentences
	group.UUID = uuid
	group.CreateTime = oldGroup.CreateTime
	group.UpdateTime = time.Now().Unix()

	updatedGroup, err = sentenceGroupDao.Create(group, tx)
	if err != nil {
		return
	}

	err = propagateUpdateFromFlow(flows, []model.SentenceGroup{*updatedGroup}, tx)
	if err != nil {
		return
	}

	err = dbLike.Commit(tx)
	return
}

func DeleteSentenceGroup(uuid string) error {
	return sentenceGroupDao.Delete(uuid, sqlConn)
}

func propagateUpdateFromFlow(flows []model.ConversationFlow, sgs []model.SentenceGroup, sqlLike model.SqlLike) (err error) {
	if len(flows) == 0 || len(sgs) == 0 {
		return
	}

	logger.Info.Printf("flows: %+v\n", flows)
	logger.Info.Printf("sgs: %+v\n", sgs)

	// update sg id in flows
	sgMap := map[string]int64{}
	for _, sg := range sgs {
		sgMap[sg.UUID] = sg.ID
	}

	flowUUID := make([]string, len(flows))
	flowID := make([]int64, len(flows))
	for i := range flows {
		flow := &flows[i]
		for j := range flow.SentenceGroups {
			sentenceGroup := &flow.SentenceGroups[j]
			if sgID, ok := sgMap[sentenceGroup.UUID]; ok {
				sentenceGroup.ID = sgID
			}
		}
		flowUUID[i] = flow.UUID
		flowID[i] = flow.ID
	}

	err = conversationFlowDao.DeleteMany(flowUUID, sqlLike)
	if err != nil {
		return
	}

	err = conversationFlowDao.CreateMany(flows, sqlLike)
	if err != nil {
		return
	}

	flowFilter := &model.ConversationFlowFilter{
		UUID: flowUUID,
	}
	newFlows, err := conversationFlowDao.GetBy(flowFilter, sqlLike)
	if err != nil {
		return
	}

	rules, err := conversationRuleDao.GetByFlowID(flowID, sqlLike)
	if err != nil {
		return
	}

	return propagateUpdateFromRule(rules, newFlows, sqlLike)
}
