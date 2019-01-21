package qi

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"emotibot.com/emotigo/module/qic-api/util/general"
	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"

	"emotibot.com/emotigo/module/qic-api/model/v1"
)

type testDao struct {
	output []interface{}
}

func (t *testDao) popOutput() (interface{}, error) {
	if len(t.output) == 0 {
		return nil, errors.New("all output is depleted")
	}
	o := t.output[0]
	t.output = t.output[1:]

	return o, nil
}
func (t *testDao) Begin() (*sql.Tx, error) {
	o, err := t.popOutput()
	oo, ok := o.(*sql.Tx)
	if !ok {
		return nil, fmt.Errorf("mockOutput %T is not expected type", o)
	}
	return oo, err
}

func (t *testDao) Tags(tx *sql.Tx, query model.TagQuery) ([]model.Tag, error) {
	o, err := t.popOutput()
	oo, ok := o.([]model.Tag)
	if !ok {
		return nil, fmt.Errorf("mockOutput %T is not expected type", o)
	}
	return oo, err
}

func (t *testDao) NewTags(tx *sql.Tx, tags []model.Tag) ([]model.Tag, error) {
	o, err := t.popOutput()
	oo, ok := o.([]model.Tag)
	if !ok {
		return nil, fmt.Errorf("mockOutput %T is not expected type", o)
	}
	return oo, err
}

func (t *testDao) DeleteTags(tx *sql.Tx, query model.TagQuery) (int64, error) {
	o, err := t.popOutput()
	oo, ok := o.(int64)
	if !ok {
		return 0, fmt.Errorf("mockOutput %T is not expected type", o)
	}
	return oo, err
}
func (t *testDao) CountTags(tx *sql.Tx, query model.TagQuery) (uint, error) {
	o, err := t.popOutput()
	oo, ok := o.(uint)
	if !ok {
		return 0, fmt.Errorf("mockOutput %T is not expected type", o)
	}
	return oo, err
}
func (t *testDao) Commit(tx *sql.Tx) error {
	return tx.Commit()
}

func (t *testDao) GetSentences(tx *sql.Tx, q *model.SentenceQuery) ([]*model.Sentence, error) {
	o, err := t.popOutput()
	oo, ok := o.([]*model.Sentence)
	if !ok {
		return nil, fmt.Errorf("mockOutput %T is not expected type", o)
	}
	return oo, err
}
func (t *testDao) InsertSentence(tx *sql.Tx, s *model.Sentence) (int64, error) {
	o, err := t.popOutput()
	oo, ok := o.(int64)
	if !ok {
		return 0, fmt.Errorf("mockOutput %T is not expected type", o)
	}
	return oo, err
}
func (t *testDao) SoftDeleteSentence(tx *sql.Tx, q *model.SentenceQuery) (int64, error) {
	o, err := t.popOutput()
	oo, ok := o.(int64)
	if !ok {
		return 0, fmt.Errorf("mockOutput %T is not expected type", o)
	}
	return oo, err
}
func (t *testDao) CountSentences(tx *sql.Tx, q *model.SentenceQuery) (uint64, error) {
	o, err := t.popOutput()
	oo, ok := o.(uint64)
	if !ok {
		return 0, fmt.Errorf("mockOutput %T is not expected type", o)
	}
	return oo, err
}
func (t *testDao) MoveCategories(x *sql.Tx, q *model.SentenceQuery, category uint64) (int64, error) {
	o, err := t.popOutput()
	oo, ok := o.(int64)
	if !ok {
		return 0, fmt.Errorf("mockOutput %T is not expected type", o)
	}
	return oo, err
}
func (t *testDao) InsertSenTagRelation(tx *sql.Tx, s *model.Sentence) error {
	o, err := t.popOutput()
	oo, ok := o.(error)
	if !ok {
		return nil
	}
	if err != nil {
		return err
	}
	return oo
}
func (t *testDao) GetRelSentenceIDByTagIDs(tx *sql.Tx, tagIDs []uint64) (map[uint64][]uint64, error) {
	o, err := t.popOutput()
	oo, ok := o.(map[uint64][]uint64)
	if !ok {
		return nil, fmt.Errorf("mockOutput %T is not expected type", o)
	}
	return oo, err
}

func assertTag(te *testing.T, mt model.Tag, t tag) {
	if t.TagUUID != mt.UUID {
		te.Error("expect tag id to be ", mt.ID, ", but got ", t.TagUUID)
		return
	}
	if t.TagName != mt.Name {
		te.Error("expect tag name to be ", mt.Name, ", but got ", t.TagName)
		return
	}
	resolvedTyp, ok := tagTypeDict[mt.Typ]
	if !ok {
		resolvedTyp = "default"
	}
	if resolvedTyp != t.TagType {
		te.Error("expect tag type to be ", mt.Typ, ", but got ", t.TagType)
		return
	}
	var negSentences, posSentences []string
	err := json.Unmarshal([]byte(mt.NegativeSentence), &negSentences)
	if err != nil {
		te.Fatal("unmarshal failed with model tag negative sentences payload ", mt.NegativeSentence, ": ", err)
	}
	err = json.Unmarshal([]byte(mt.PositiveSentence), &posSentences)
	if err != nil {
		te.Fatal("unmarshal failed with model tag positive sentences payload ", mt.PositiveSentence, ": ", err)
	}
	if !reflect.DeepEqual(t.NegSentences, negSentences) {
		te.Errorf("expect tag negative sentence to be the same, but got %#v and %#v ", negSentences, t.NegSentences)
		return
	}
	if !reflect.DeepEqual(t.PosSentences, posSentences) {
		te.Errorf("expect tag positive sentence to be the same, but got %#v and %#v", posSentences, t.PosSentences)
		return
	}
}

func TestTags(t *testing.T) {
	var expectTags = []model.Tag{
		model.Tag{
			UUID:             "abc",
			PositiveSentence: "[]",
			NegativeSentence: "[]",
		},
	}
	tagDao = &testDao{
		output: []interface{}{
			uint(10),
			expectTags,
		},
	}
	resp, err := Tags("csbot", 1, 1)
	if err != nil {
		t.Fatal("expect tags called ok, but got error: ", err)
	}
	if len(resp.Data) != len(expectTags) {
		t.Error("expect data to be ", len(expectTags), ", but got ", len(resp.Data))
	}
	for i, d := range resp.Data {
		t.Log("data index: ", i)
		assertTag(t, expectTags[i], d)
	}
	var expectedPaging = general.Paging{
		Total: 10,
		Page:  1,
		Limit: 1,
	}
	if !reflect.DeepEqual(expectedPaging, resp.Paging) {
		t.Logf("expected: %v\nreal: %v\n", expectedPaging, resp.Paging)
		t.Error("expect paging data is correct, but got unexpected data.")
	}
}

func TestNewTag(t *testing.T) {
	var expectedTags = []model.Tag{
		model.Tag{ID: 1},
	}
	tagDao = &testDao{
		output: []interface{}{
			expectedTags,
			errors.New("test failed"),
		},
	}

	id, err := NewTag(model.Tag{ID: 0, UUID: "4c9c82fb7b5845369a520b757ab03f8b"})
	if err != nil {
		t.Fatal("expect new tag to be ok, but got ", err)
	}
	if id != expectedTags[0].ID {
		t.Error("expect new tag id to be ", expectedTags[0].ID, ", but got ", id)
	}
	_, err = NewTag(model.Tag{})
	if err == nil {
		t.Fatal("expect new tag to handle error, but no error has returned")
	}
}
func TestUpdateTags(t *testing.T) {
	d := &testDao{
		output: []interface{}{
			//Begin
			mockTx(t),
			//Query
			[]model.Tag{
				model.Tag{ID: 1},
			},
			//GetRelSentenceIDByTagIDs
			map[uint64][]uint64{
				1: []uint64{1, 2},
			},
			//Delete
			int64(1),
			//New
			[]model.Tag{
				model.Tag{ID: 2},
			},
			//InsertSenTagRelation
			nil,
			nil,
		},
	}
	tagDao = d
	sentenceDao = d
	id, err := UpdateTag(model.Tag{})
	if err != nil {
		t.Fatal("expect update to be ok, but got ", err)
	}
	if id != 2 {
		t.Error("expect new id to be 2, but got ", id)
	}
}
func TestDeleteTag(t *testing.T) {
	tagDao = &testDao{
		output: []interface{}{
			mockTx(t),
			int64(3),
			mockTx(t),
			int64(2),
		},
	}
	err := DeleteTag("1", "2", "3")
	if err != nil {
		t.Fatal("expect delete tag to be ok, but got ", err)
	}
	err = DeleteTag("1", "2", "3")
	if err == nil {
		t.Fatal("expect delete tag to has error, but got none")
	}

}

func mockTx(t *testing.T) *sql.Tx {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal("init sqlmock failed, ", err)
	}
	mock.ExpectBegin()
	mock.ExpectCommit()
	tx, _ := db.Begin()
	return tx
}
