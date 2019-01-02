package qi

import (
	"encoding/json"
	"fmt"
	"time"

	"emotibot.com/emotigo/module/qic-api/util/general"

	"emotibot.com/emotigo/module/qic-api/model/v1"
)

var tagTypeDict = map[int8]string{
	0: "default",
	1: "keyword",
	2: "dialogue_act",
	3: "user_response",
}

//Tags is the service for getting the tags json response.
func Tags(entID string, limit, page int) (resp *TagResponse, err error) {
	enterprise := entID
	query := model.TagQuery{
		Enterprise: &enterprise,
		Paging: &model.Pagination{
			Limit: limit,
			Page:  page,
		},
	}
	result, err := tagDao.Tags(nil, query)
	if err != nil {
		return nil, fmt.Errorf("get tag from dao failed, %v", err)
	}
	counts, err := tagDao.CountTags(nil, query)
	if err != nil {
		return nil, fmt.Errorf("get tag count from dao failed, %v", err)
	}
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
	resp = &TagResponse{
		Paging: general.Paging{
			Total: int64(counts),
			Limit: limit,
			Page:  page,
		},
		Data: tags,
	}
	return
}

// NewTag create a tag from t.
// incremental id will be returned, if the dao supported it.
func NewTag(t model.Tag) (id uint64, err error) {
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
	tx, err := tagDao.Begin()
	if err != nil {
		return 0, fmt.Errorf("dao init transaction failed, %v", err)
	}
	defer func() {
		tx.Rollback()
	}()
	var (
		tags       []model.Tag
		maxRetries = 3
		rowsCount  int64
	)

	// we will try at least maxRetries time for delete operation
	for i := 0; i <= 3; i++ {
		if i == maxRetries {
			return 0, fmt.Errorf("unexpected affected rows count")
		}
		query := model.TagQuery{
			UUID: []string{t.UUID},
		}
		tags, err = tagDao.Tags(tx, query)
		if len(tags) != 1 {
			return 0, fmt.Errorf("dao found %d tag with the query %+v", len(tags), t.UUID)
		}
		query.ID = []uint64{tags[0].ID}
		rowsCount, err = tagDao.DeleteTags(tx, query)
		if err != nil {
			return 0, fmt.Errorf("dao delete failed, %v", err)
		}
		if rowsCount == 1 {
			//we got the success signal, continue normal flow.
			break
		}
	}

	t.CreateTime = tags[0].CreateTime
	t.UpdateTime = time.Now().Unix()
	tags, err = tagDao.NewTags(tx, []model.Tag{t})
	if err != nil {
		return 0, fmt.Errorf("dao create new tags failed, %v", err)
	}
	err = tx.Commit()
	if err != nil {
		return 0, fmt.Errorf("tx commit failed, %v", err)
	}
	return tags[0].ID, nil
}

// DeleteTag delete the tag by dao.
// all operation will be included in one transaction.
// Which ensure each uuid is delete or nothing is deleted.
func DeleteTag(uuid ...string) error {
	if len(uuid) == 0 {
		return nil
	}
	tx, err := tagDao.Begin()
	if err != nil {
		return fmt.Errorf("dao init transaction failed, %v", err)
	}
	defer func() {
		tx.Commit()
	}()
	query := model.TagQuery{
		UUID: uuid,
	}
	_, err = tagDao.DeleteTags(tx, query)
	if err != nil && err != model.ErrAutoIDDisabled {
		return fmt.Errorf("dao delete failed, %v", err)
	}
	tx.Commit()
	return nil
}
