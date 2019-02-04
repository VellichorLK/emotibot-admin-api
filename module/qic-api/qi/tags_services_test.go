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

func (t *testDao) Tags(tx model.SqlLike, query model.TagQuery) ([]model.Tag, error) {
	o, err := t.popOutput()
	oo, ok := o.([]model.Tag)
	if !ok {
		return nil, fmt.Errorf("mockOutput %T is not expected type", o)
	}
	return oo, err
}

func (t *testDao) NewTags(tx model.SqlLike, tags []model.Tag) ([]model.Tag, error) {
	o, err := t.popOutput()
	oo, ok := o.([]model.Tag)
	if !ok {
		return nil, fmt.Errorf("mockOutput %T is not expected type", o)
	}
	return oo, err
}

func (t *testDao) DeleteTags(tx model.SqlLike, query model.TagQuery) (int64, error) {
	o, err := t.popOutput()
	oo, ok := o.(int64)
	if !ok {
		return 0, fmt.Errorf("mockOutput %T is not expected type", o)
	}
	return oo, err
}
func (t *testDao) CountTags(tx model.SqlLike, query model.TagQuery) (uint, error) {
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

func (t *testDao) GetSentences(tx model.SqlLike, q *model.SentenceQuery) ([]*model.Sentence, error) {
	o, err := t.popOutput()
	oo, ok := o.([]*model.Sentence)
	if !ok {
		return nil, fmt.Errorf("mockOutput %T is not expected type", o)
	}
	return oo, err
}
func (t *testDao) InsertSentence(tx model.SqlLike, s *model.Sentence) (int64, error) {
	o, err := t.popOutput()
	oo, ok := o.(int64)
	if !ok {
		return 0, fmt.Errorf("mockOutput %T is not expected type", o)
	}
	return oo, err
}
func (t *testDao) SoftDeleteSentence(tx model.SqlLike, q *model.SentenceQuery) (int64, error) {
	o, err := t.popOutput()
	oo, ok := o.(int64)
	if !ok {
		return 0, fmt.Errorf("mockOutput %T is not expected type", o)
	}
	return oo, err
}
func (t *testDao) CountSentences(tx model.SqlLike, q *model.SentenceQuery) (uint64, error) {
	o, err := t.popOutput()
	oo, ok := o.(uint64)
	if !ok {
		return 0, fmt.Errorf("mockOutput %T is not expected type", o)
	}
	return oo, err
}
func (t *testDao) MoveCategories(x model.SqlLike, q *model.SentenceQuery, category uint64) (int64, error) {
	o, err := t.popOutput()
	oo, ok := o.(int64)
	if !ok {
		return 0, fmt.Errorf("mockOutput %T is not expected type", o)
	}
	return oo, err
}
func (t *testDao) InsertSenTagRelation(tx model.SqlLike, s *model.Sentence) error {
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
func (t *testDao) GetRelSentenceIDByTagIDs(tx model.SqlLike, tagIDs []uint64) (map[uint64][]uint64, error) {
	o, err := t.popOutput()
	oo, ok := o.(map[uint64][]uint64)
	if !ok {
		return nil, fmt.Errorf("mockOutput %T is not expected type", o)
	}
	return oo, err
}

func (t *testDao) InsertSentences(tx model.SqlLike, sentences []model.Sentence) error {
	return nil
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
	tags, _ := toTag(expectTags...)
	var enterprise = "csbot"
	type args struct {
		Query model.TagQuery
	}
	testcases := []struct {
		Name   string
		Args   args
		Output TagResponse
	}{
		{
			Name: "normal",
			Args: args{
				Query: model.TagQuery{
					Enterprise: &enterprise,
					Paging: &model.Pagination{
						Page:  1,
						Limit: 1,
					},
				},
			},
			Output: TagResponse{
				Paging: general.Paging{
					Total: 10,
					Page:  1,
					Limit: 1,
				},
				Data: tags,
			},
		},
		{
			Name: "unlimited tags",
			Args: args{
				Query: model.TagQuery{
					Enterprise: &enterprise,
				},
			},
			Output: TagResponse{
				Paging: general.Paging{
					Total: 10,
				},
				Data: tags,
			},
		},
	}
	for _, tt := range testcases {
		t.Run(tt.Name, func(t *testing.T) {
			tagDao = &testDao{
				output: []interface{}{
					uint(10),
					expectTags,
				},
			}
			resp, err := Tags(tt.Args.Query)
			if err != nil {
				t.Fatal("expect tags called ok, but got error: ", err)
			}
			if !reflect.DeepEqual(*resp, tt.Output) {
				t.Logf("expected: %+v\nreal: %+v\n", tt.Output, resp)
				t.Error("expect paging data is correct, but got unexpected data.")
			}

		})
	}
}

func TestNewTag(t *testing.T) {
	testcases := []struct {
		name        string
		arg         model.Tag
		mock        *testDao
		expectID    uint64
		expectError bool
	}{
		{
			name: "normal",
			arg: model.Tag{
				UUID:             "4c9c82fb7b5845369a520b757ab03f8b",
				PositiveSentence: `["1","2"]`,
			},
			mock: &testDao{
				output: []interface{}{
					[]model.Tag{
						model.Tag{ID: 1},
					},
				},
			},
			expectID:    1,
			expectError: false,
		},
		{
			name: "db error",
			arg: model.Tag{
				UUID:             "4c9c82fb7b5845369a520b757ab03f8b",
				PositiveSentence: `["1","2"]`,
			},
			mock: &testDao{
				output: []interface{}{
					errors.New("test failed"),
				},
			},
			expectError: true,
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			tagDao = tt.mock
			id, err := NewTag(tt.arg)
			if !tt.expectError && err != nil {
				t.Fatal("expect new tag to be ok, but got ", err)
			}
			if tt.expectError {
				if err == nil {
					t.Fatal("expect new tag has error, but got nil")
				}
				return
			}
			if id != tt.expectID {
				t.Error("expect new tag id to be ", tt.expectID, ", but got ", id)
			}
		})
	}

}

func TestUpdateTags(t *testing.T) {
	t.Skip("any level higher than tag need to mock now, skip it for a test refractor")
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
