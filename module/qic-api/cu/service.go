package cu

import (
	"encoding/hex"
	"time"

	uuid "github.com/satori/go.uuid"
)

var (
	serviceDao Dao
)

func createFlowConversation(enterprise string, user string, body *apiFlowCreateBody) (string, error) {
	now := time.Now().Unix()
	uuid, err := uuid.NewV4()
	if err != nil {
		return "", err
	}
	uuidStr := hex.EncodeToString(uuid[:])
	daoData := &daoFlowCreate{enterprise: enterprise, typ: Flow, leftChannel: Host, rightChannel: Guest,
		fileName: body.FileName, callTime: body.CreateTime, uploadTime: now, updateTime: now, uuid: uuidStr,
		user: user}

	_, err = serviceDao.CreateFlowConversation(nil, daoData)
	if err != nil {
		return "", err
	}
	return uuidStr, nil
}
