package qi

import (
	"encoding/json"
	"fmt"
	"time"

	"emotibot.com/emotigo/module/admin-api/util/AdminErrors"
	"emotibot.com/emotigo/pkg/logger"

	"github.com/satori/go.uuid"

	"emotibot.com/emotigo/module/qic-api/util/general"

	"emotibot.com/emotigo/module/qic-api/model/v1"
)

var tagTypeDict = map[int8]string{
	0: "default",
	1: "keyword",
	2: "dialogue_act",
	3: "user_response",
}

// Tags is the service for getting the tags json response.
// If the query.Paging is nil, response.paging.Limit & paging will be 0, 0.
// If the query.Enterprise is nil, an error will be returned.
func Tags(query model.TagQuery) (resp *TagResponse, err error) {
	if query.Enterprise == nil {
		return nil, fmt.Errorf("query must contain enterpriseID")
	}
	counts, err := tagDao.CountTags(nil, query)
	if err != nil {
		return nil, fmt.Errorf("get tag count from dao failed, %v", err)
	}
	tags, err := TagsByQuery(query)
	if err != nil {
		return nil, fmt.Errorf("call tags by query failed, %v", err)
	}
	resp = &TagResponse{
		Paging: general.Paging{
			Total: int64(counts),
		},
		Data: tags,
	}
	if query.Paging != nil {
		resp.Paging.Limit = query.Paging.Limit
		resp.Paging.Page = query.Paging.Page
	}
	return
}

// toTag transform model.Tag to the presentive tag for tag controller
func toTag(result ...model.Tag) ([]tag, error) {
	var err error
	var tags = make([]tag, 0, len(result))
	for _, t := range result {
		typ, found := tagTypeDict[t.Typ]
		if !found {
			typ = "default"
		}
		var posSentences, negSentences []string
		err = json.Unmarshal([]byte(t.PositiveSentence), &posSentences)
		if err != nil {
			return nil, fmt.Errorf("tag %d positive sentence payload is not a valid string array, %v", t.ID, err)
		}
		err = json.Unmarshal([]byte(t.NegativeSentence), &negSentences)
		if err != nil {
			return nil, fmt.Errorf("tag %d negative sentence payload is not a valid string array, %v", t.ID, err)
		}

		tags = append(tags, tag{
			TagUUID:      t.UUID,
			TagName:      t.Name,
			TagType:      typ,
			PosSentences: posSentences,
			NegSentences: negSentences,
		})
	}
	return tags, nil
}
func TagsByQuery(query model.TagQuery) ([]tag, error) {

	result, err := tagDao.Tags(nil, query)
	if err != nil {
		return nil, fmt.Errorf("get tags from dao failed, %v", err)
	}

	return toTag(result...)
}

// NewTag create a tag from t.
// incremental id will be returned, if the dao supported it.
// If t is not valid(etc uuid or positive sentence is empty...) then an adminError will returned.
func NewTag(t model.Tag) (id uint64, err error) {

	if _, err := uuid.FromString(t.UUID); t.UUID == "" && err != nil {
		return 0, &controllerError{
			errNo: AdminErrors.ErrnoRequestError,
			error: fmt.Errorf("tag UUID '%s' format is not correct", t.UUID),
		}
	}
	var ps []string
	json.Unmarshal([]byte(t.PositiveSentence), &ps)
	// If positive sentence is empty, cu model can not be trained.
	if len(ps) < 1 {
		return 0, &controllerError{
			errNo: AdminErrors.ErrnoRequestError,
			error: fmt.Errorf("must have at least one positive tag"),
		}
	}
	input := []model.Tag{t}
	createdTags, err := tagDao.NewTags(nil, input)
	if err != nil {
		return 0, fmt.Errorf("create tag from dao failed, %v", err)
	}
	if len(createdTags) != len(input) {
		return 0, fmt.Errorf("unexpected dao internal error, %d have been returned instead of input %d", len(createdTags), len(input))
	}
	return createdTags[0].ID, nil
}

// UpdateTag update the origin t, since tag need to keep the origin value.
// update will try to delete the old one with the same uuid, and create new one.
// multiple update called on the same t will try it best to resolve to one state, but not guarantee success.
// if conflicted can not be resolved, id will be 0 and err will be nil.
func UpdateTag(t model.Tag) (id uint64, err error) {
	tx, err := dbLike.Begin()
	if err != nil {
		return 0, fmt.Errorf("dao init transaction failed, %v", err)
	}
	defer func() {
		tx.Rollback()
	}()

	tagsSentences, err := sentenceDao.GetRelSentenceIDByTagIDs(tx, []uint64{t.ID})
	if err != nil {
		return 0, fmt.Errorf("dao get sentence id failed, %v", err)
	}
	rowsCount, err := tagDao.DeleteTags(tx, model.TagQuery{
		ID: []uint64{t.ID},
	})
	if err != nil {
		return 0, fmt.Errorf("dao delete failed, %v", err)
	}
	if rowsCount != 1 {
		return 0, fmt.Errorf("delete tag failed, no affected rows")
	}
	t.UpdateTime = time.Now().Unix()
	tags, err := tagDao.NewTags(tx, []model.Tag{t})
	// TODO: support elegant handle for sql driver not support return incremental id.
	// if err == model.ErrAutoIDDisabled {
	//	tagDao.Tags()
	// 	tx.Commit()
	// }
	if err != nil {
		return 0, fmt.Errorf("dao create new tags failed, %v", err)
	}
	newTag := tags[0]

	sentences := tagsSentences[t.ID]
	// to avoid nil panic if tag have no sentences
	if sentences == nil {
		return newTag.ID, nil
	}

	sentenceGrp, err := sentenceDao.GetSentences(tx, &model.SentenceQuery{
		ID: sentences,
	})

	err = propagateUpdateFromSentence(sentenceGrp, newTag.ID, t.ID, t.Enterprise, tx)
	if err != nil {
		logger.Error.Printf("propage update  failed. %s", err)
		return 0, err
	}

	err = tx.Commit()
	if err != nil {
		logger.Error.Printf("commit transcation failed. %s", err)
		return 0, err
	}

	return newTag.ID, nil
}

// DeleteTag delete the tag by dao.
// all operation will be included in one transaction.
// Which ensure each uuid is delete or nothing is deleted.
func DeleteTag(uuid ...string) error {
	if len(uuid) == 0 {
		return nil
	}
	tx, err := dbLike.Begin()
	if err != nil {
		return fmt.Errorf("dao init transaction failed, %v", err)
	}
	defer func() {
		tx.Rollback()
	}()

	query := model.TagQuery{
		UUID:             uuid,
		IgnoreSoftDelete: false,
	}
	tags, err := tagDao.Tags(tx, query)
	if err != nil {
		return fmt.Errorf("dao get tags failed. %v", err)
	}

	affectedrows, err := tagDao.DeleteTags(tx, query)
	if err == model.ErrAutoIDDisabled {
		logger.Warn.Println("tag table does not support affectedrow, we will continue to work, but we can not detect conflict now.")
		tx.Commit()
		return nil
	}
	if err != nil {
		return fmt.Errorf("dao delete failed, %v", err)
	}
	if err != model.ErrAutoIDDisabled && int(affectedrows) != len(uuid) {
		return fmt.Errorf("dao delete should delete %d of rows, but only %d. possible conflict operation at the same time", len(uuid), affectedrows)
	}

	tagMap := map[uint64]bool{}
	tagID := []uint64{}
	var enterprise string
	for _, tag := range tags {
		tagMap[tag.ID] = true
		tagID = append(tagID, tag.ID)
		enterprise = tag.Enterprise
	}

	sentences, err := sentenceDao.GetRelSentenceIDByTagIDs(tx, tagID)
	if err != nil {
		return fmt.Errorf("get sentences failed. %v", err)
	}

	sentenceID := []uint64{}
	for _, v := range sentences {
		sentenceID = append(sentenceID, v...)
	}

	sq := &model.SentenceQuery{
		ID:         sentenceID,
		Enterprise: &enterprise,
	}

	sentenceGrp, err := sentenceDao.GetSentences(tx, sq)
	if err != nil {
		return err
	}

	for i := range sentenceGrp {
		s := sentenceGrp[i]
		if len(s.TagIDs) == 1 {
			s.TagIDs = []uint64{}
			continue
		}

		for j, tag := range s.TagIDs {
			if _, ok := tagMap[tag]; ok {
				if j == len(s.TagIDs)-1 {
					s.TagIDs = s.TagIDs[:j]
				} else {
					s.TagIDs = append(s.TagIDs[:j], s.TagIDs[j+1:]...)
				}
			}
		}
	}

	err = propagateUpdateFromSentence(sentenceGrp, 0, 0, enterprise, tx)
	if err != nil {
		return err
	}

	err = tx.Commit()
	return err
}

func propagateUpdateFromSentence(sentences []*model.Sentence, newTag, oldTag uint64, enterprise string, tx model.SQLTx) (err error) {
	logger.Info.Printf("tags %+v\n", newTag)
	logger.Info.Printf("sentences: %+v\n", sentences)
	if len(sentences) == 0 {
		return
	}

	sUUID := []string{}
	sID := []int64{}
	activeSentences := []model.Sentence{}
	for i := range sentences {
		s := sentences[i]
		if s.IsDelete == 1 {
			continue
		}

		for j, tagID := range s.TagIDs {
			if tagID == oldTag {
				s.TagIDs[j] = newTag
			}
		}
		sID = append(sID, int64(s.ID))
		sUUID = append(sUUID, s.UUID)
		activeSentences = append(activeSentences, *s)
	}

	// delete old sentences
	var deleted int8
	sentenceQuery := &model.SentenceQuery{
		UUID:       sUUID,
		IsDelete:   &deleted,
		Enterprise: &enterprise,
	}

	logger.Info.Printf("sq: %+v\n", sentenceQuery)

	_, err = sentenceDao.SoftDeleteSentence(tx, sentenceQuery)
	if err != nil {
		logger.Error.Printf("delete sentence failed. %s", err)
		return
	}

	err = sentenceDao.InsertSentences(tx, activeSentences)
	if err != nil {
		logger.Error.Printf("insert sentences failed. %s", err)
		return
	}

	sgs, err := sentenceGroupDao.GetBySentenceID(sID, tx)
	if err != nil {
		logger.Error.Printf("get sentence groups failed. %s", err)
		return
	}

	return propagateUpdateFromSentenceGroup(sgs, activeSentences, tx)
}
