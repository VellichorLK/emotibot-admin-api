package qi

import (
	"emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/pkg/logger"
	"fmt"
	"github.com/satori/go.uuid"
	"strings"
	"time"
)

var sentenceGroupDao model.SentenceGroupsSqlDao = &model.SentenceGroupsSqlDaoImpl{}

func CreateSentenceGroup(group *model.SentenceGroup) (createdGroup *model.SentenceGroup, err error) {
	if group == nil {
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

	uuids := []string{}
	for _, s := range group.Sentences {
		uuids = append(uuids, s.UUID)
	}
	var isDelete int8 = int8(0)
	sentenceQuery := &model.SentenceQuery{
		Enterprise: &group.Enterprise,
		UUID:       uuids,
		IsDelete:   &isDelete,
	}

	sentences, err := sentenceDao.GetSentences(tx, sentenceQuery)
	if err != nil {
		return
	}

	if len(sentences) != len(group.Sentences) {
		logger.Warn.Printf("user input sentences does not match sentences in db")
	}

	simpleSentences := []model.SimpleSentence{}
	for _, s := range sentences {
		simpleSentence := model.SimpleSentence{
			ID:   s.ID,
			UUID: s.UUID,
			Name: s.Name,
		}
		simpleSentences = append(simpleSentences, simpleSentence)
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
