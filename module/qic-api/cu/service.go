package cu

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"emotibot.com/emotigo/pkg/logger"

	"emotibot.com/emotigo/module/admin-api/util"

	"emotibot.com/emotigo/module/qic-api/util/timecache"
	uuid "github.com/satori/go.uuid"
)

var (
	serviceDao Dao
	cache      timecache.TimeCache
)

//Error msg
var (
	ErrSpeaker = errors.New("Wrong speaker")
)

func createFlowConversation(enterprise string, user string, body *apiFlowCreateBody) (string, error) {
	now := time.Now().Unix()
	uuid, err := uuid.NewV4()
	if err != nil {
		return "", err
	}
	uuidStr := hex.EncodeToString(uuid[:])
	daoData := &daoFlowCreate{enterprise: enterprise, typ: Flow, leftChannel: ChannelHost, rightChannel: ChannelGuest,
		fileName: body.FileName, callTime: body.CreateTime, uploadTime: now, updateTime: now, uuid: uuidStr,
		user: user}

	_, err = serviceDao.CreateFlowConversation(nil, daoData)
	if err != nil {
		return "", err
	}
	return uuidStr, nil
}

func insertSegmentByUUID(uuid string, asr []*apiFlowAddBody) error {

	if asr == nil || len(uuid) != 32 {
		return nil
	}

	id, err := getIDByUUID(uuid)
	if err != nil {
		return err
	}

	//begin a transaction
	tx, err := serviceDao.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	//create the segment structure and insert into db
	for i := 0; i < len(asr); i++ {
		seg := &Segment{callID: id, asr: asr[i], creatTime: time.Now().Unix()}
		switch asr[i].Speaker {
		case WordHost:
			seg.channel = ChannelHost
		case WordGuest:
			seg.channel = ChannelGuest
		case WordSilence:
			seg.channel = ChannelSilence
		default:
			return ErrSpeaker
		}
		serviceDao.InsertSegment(tx, seg)
	}
	return serviceDao.Commit(tx)

}

func getFlowSentences(uuid string) ([]*Segment, error) {
	id, err := getIDByUUID(uuid)
	if err != nil {
		return nil, err
	}
	return serviceDao.GetSegmentByCallID(nil, id)
}

func segmentToV1PredictRequest(segments []*Segment) []*V1PredictRequestData {
	num := len(segments)
	V1PredictRequestDatas := make([]*V1PredictRequestData, num, num)
	for i := 0; i < num; i++ {
		data := &V1PredictRequestData{ID: i + 1, Sentence: segments[i].asr.Text}
		V1PredictRequestDatas[i] = data
	}
	return V1PredictRequestDatas
}

func getConversation(uuid string) (*ConversationInfo, error) {
	return serviceDao.GetConversationByUUID(nil, uuid)
}

func getIDByUUID(uuid string) (uint64, error) {
	var id uint64
	idCachKey := uuid + "id"
	//get the id from Conversation table by uuid
	v, hasData := cache.GetCache(idCachKey)
	if !hasData {
		info, err := serviceDao.GetConversationByUUID(nil, uuid)
		if err != nil {
			return 0, err
		}
		cache.SetCache(idCachKey, info.CallID)
		id = info.CallID
	} else {
		id = v.(uint64)
	}
	return id, nil
}

func predictByV1CuModule(context *V1PredictContext) (*V1PredictResult, error) {

	url := ModuleInfo.Environments["LOGIC_PREDICT_URL"]
	resp, err := util.HTTPPostJSONWithHeader(url, context, 2, map[string]string{"Content-Type": "application/json"})
	if err != nil {
		logger.Error.Printf("%s\n", err)
		return nil, err
	}
	predictResult := &V1PredictResult{}

	err = json.Unmarshal([]byte(resp), predictResult)

	if err != nil {
		logger.Error.Printf("%s\n", err)
		return nil, err
	}

	return predictResult, nil

}

//GetFlowGroup gets the group for flow usage in enterprise
func GetFlowGroup(enterprise string) ([]Group, error) {
	if enterprise == "" {
		return nil, nil
	}
	queryCondition := GroupQuery{EnterpriseID: &enterprise, Type: []int{Flow, AudioFile}}
	return serviceDao.Group(nil, queryCondition)
}

//GetRuleLogic gets the logic in the rule, the rule in the group information
func GetRuleLogic(groupID uint64) ([]*QIResult, error) {
	ruleLogicIDs, ruleOrder, err := serviceDao.GetGroupToLogicID(nil, groupID)
	if err != nil {
		return nil, err
	}

	//get the rule information
	ruleCondition := RuleQuery{ID: ruleOrder}
	rules, err := serviceDao.GetRule(nil, ruleCondition)
	if err != nil {
		return nil, err
	}

	//transform rule slice to map[rule_id] rule
	rulesMap := make(map[uint64]*Rule, len(rules))
	for i := 0; i < len(rules); i++ {
		rulesMap[rules[i].RuleID] = rules[i]
	}

	//collect all logic id
	logicIDs := make([]uint64, 0, 16)
	for _, v := range ruleLogicIDs {
		logicIDs = append(logicIDs, v...)
	}

	//get all logic information
	logicCondition := LogicQuery{ID: logicIDs}
	logics, err := serviceDao.GetLogic(nil, logicCondition)
	if err != nil {
		return nil, err
	}

	//transform logics to map[logic_id] logic
	logicsMap := make(map[uint64]*Logic, len(logics))
	for i := 0; i < len(logics); i++ {
		logicsMap[logics[i].LogicID] = logics[i]
	}

	numOfRule := len(ruleOrder)
	ruleRes := make([]*QIResult, 0, numOfRule)
	for i := 0; i < numOfRule; i++ {
		ruleID := ruleOrder[i]
		if rule, ok := rulesMap[ruleID]; ok {
			result := &QIResult{Name: rule.Name, ID: ruleID}

			localLogicIDs := ruleLogicIDs[ruleID]
			for i := 0; i < len(localLogicIDs); i++ {

				logicID := localLogicIDs[i]
				if logic, ok := logicsMap[logicID]; ok {
					logicRes := &LogicResult{Name: logic.Name, ID: logicID, Recommend: make([]string, 0)}
					result.LogicResult = append(result.LogicResult, logicRes)
				}

			}
			ruleRes = append(ruleRes, result)
		}
	}
	return ruleRes, nil
}

//FillCUCheckResult fills the result
func FillCUCheckResult(predict *V1PredictResult, result []*QIResult) error {
	if predict == nil || result == nil {
		return nil
	}

	//transform predict result to map
	predictLogicMap := make(map[string]bool, len(predict.LogicResult))
	for i := 0; i < len(predict.LogicResult); i++ {
		logicName := predict.LogicResult[i].LogicRule.Name
		predictLogicMap[logicName] = true
	}

	for i := 0; i < len(result); i++ {
		logics := result[i].LogicResult
		valid := true
		for j := 0; j < len(logics); j++ {
			if _, ok := predictLogicMap[logics[j].Name]; ok {
				logics[i].Valid = true
			} else {
				valid = false
			}
			result[i].Valid = valid
		}
	}

	return nil
}
