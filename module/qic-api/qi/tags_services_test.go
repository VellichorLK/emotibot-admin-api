package qi

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"emotibot.com/emotigo/module/qic-api/util/general"

	"emotibot.com/emotigo/module/qic-api/model/v1"
)

type testTagDao struct {
	output []interface{}
}

func (t *testTagDao) popOutput() (interface{}, error) {
	if len(t.output) == 0 {
		return nil, errors.New("all output is depleted")
	}
	o := t.output[0]
	t.output = t.output[1:]

	return o, nil
}
func (t *testTagDao) Begin() (*sql.Tx, error) {
	//not mocking begin yet
	return nil, nil
}

func (t *testTagDao) Tags(tx *sql.Tx, query model.TagQuery) ([]model.Tag, error) {
	o, err := t.popOutput()
	oo, ok := o.([]model.Tag)
	if !ok {
		return nil, fmt.Errorf("mockOutput is not expected type")
	}
	return oo, err
}

func (t *testTagDao) NewTags(tx *sql.Tx, tags []model.Tag) ([]model.Tag, error) {
	o, err := t.popOutput()
	oo, ok := o.([]model.Tag)
	if !ok {
		return nil, fmt.Errorf("mockOutput is not expected type")
	}
	return oo, err
}

func (t *testTagDao) DeleteTags(tx *sql.Tx, query model.TagQuery) (int64, error) {
	o, err := t.popOutput()
	oo, ok := o.(int64)
	if !ok {
		return 0, fmt.Errorf("mockOutput is not expected type")
	}
	return oo, err
}
func (t *testTagDao) CountTags(tx *sql.Tx, query model.TagQuery) (uint, error) {
	o, err := t.popOutput()
	oo, ok := o.(uint)
	if !ok {
		return 0, fmt.Errorf("mockOutput is not expected type")
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
		te.Error("expect tag negative sentence to be the same, but got ", mt.NegativeSentence)
		return
	}
	if !reflect.DeepEqual(t.PosSentences, posSentences) {
		te.Error("expect tag positive sentence to be the same, but got ", mt.PositiveSentence)
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
	tagDao = &testTagDao{
		output: []interface{}{
			expectTags,
			uint(10),
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
	tagDao = &testTagDao{
		output: []interface{}{
			expectedTags,
			errors.New("test failed"),
		},
	}

	id, err := NewTag(model.Tag{ID: 0})
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
