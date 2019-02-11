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
	ErrNilSentenceGroup = fmt.Errorf("Sentence can not be nil")
)

var sentenceGroupDao model.SentenceGroupsSqlDao = &model.SentenceGroupsSqlDaoImpl{}

func simpleSentencesOf(group *model.SentenceGroup, tx model.SQLTx) (simpleSentences []model.SimpleSentence, err error) {
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

func DeleteSentenceGroup(uuid string) (err error) {
	tx, err := dbLike.Begin()
	if err != nil {
		return
	}
	defer dbLike.ClearTransition(tx)

	deleted := int8(0)
	filter := &model.SentenceGroupFilter{
		UUID:     []string{uuid},
		IsDelete: &deleted,
	}

	groups, err := sentenceGroupDao.GetBy(filter, tx)
	if err != nil {
		return
	}

	if len(groups) == 0 {
		return
	}

	group := groups[0]

	flows, err := conversationFlowDao.GetBySentenceGroupID([]int64{group.ID}, tx)
	if err != nil {
		return
	}

	err = sentenceGroupDao.Delete(uuid, sqlConn)
	if err != nil {
		return
	}

	// remove sentence group from these flows
	for i := range flows {
		flow := &flows[i]
		if len(flow.SentenceGroups) == 1 {
			flow.SentenceGroups = []model.SimpleSentenceGroup{}
			continue
		}

		for j, sg := range flow.SentenceGroups {
			if sg.ID == group.ID {
				if j == len(flow.SentenceGroups)-1 {
					flow.SentenceGroups = flow.SentenceGroups[:j]
				} else {
					flow.SentenceGroups = append(flow.SentenceGroups[:j], flow.SentenceGroups[j+1:]...)
				}
			}
		}
	}

	err = propagateUpdateFromFlow(flows, groups, tx)
	if err != nil {
		return
	}
	return dbLike.Commit(tx)
}

func propagateUpdateFromFlow(flows []model.ConversationFlow, sgs []model.SentenceGroup, sqlLike model.SqlLike) (err error) {
	logger.Info.Printf("flows: %+v", flows)
	logger.Info.Printf("sgs: %+v", sgs)
	if len(flows) == 0 || len(sgs) == 0 {
		return
	}

	// update sg id in flows
	sgMap := map[string]int64{}
	for _, sg := range sgs {
		sgMap[sg.UUID] = sg.ID
	}

	flowUUID := []string{}
	flowID := []int64{}
	activeFlows := []model.ConversationFlow{}
	for i := range flows {
		flow := &flows[i]
		if flow.Deleted == 1 {
			// ingore deleted flows
			continue
		}

		for j := range flow.SentenceGroups {
			sentenceGroup := &flow.SentenceGroups[j]
			if sgID, ok := sgMap[sentenceGroup.UUID]; ok {
				sentenceGroup.ID = sgID
			}
		}

		flowUUID = append(flowUUID, flow.UUID)
		flowID = append(flowID, flow.ID)
		activeFlows = append(activeFlows, *flow)
	}

	logger.Info.Printf("flowUUID: %v\n", flowUUID)

	err = conversationFlowDao.DeleteMany(flowUUID, sqlLike)
	if err != nil {
		return
	}

	err = conversationFlowDao.CreateMany(activeFlows, sqlLike)
	if err != nil {
		return
	}

	rules, err := conversationRuleDao.GetByFlowID(flowID, sqlLike)
	if err != nil {
		return
	}

	return propagateUpdateFromRule(rules, activeFlows, sqlLike)
}
