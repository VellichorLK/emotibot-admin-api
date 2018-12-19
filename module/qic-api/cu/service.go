package cu

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

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
		return nil, err
	}
	predictResult := &V1PredictResult{}
	err = json.Unmarshal([]byte(resp), predictResult)
	if err != nil {
		return nil, err
	}
	return predictResult, nil

}
