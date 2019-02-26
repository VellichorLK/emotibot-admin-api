package qi

import (
	"testing"

	model "emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/module/qic-api/util/test"
)

var mockRel = []map[uint64][]uint64{
	RGContainsR,
	RContainsC,
	CContainsSG,
	SGContainsS,
	SContainsT,
}

var mockRelNULL = []map[uint64][]uint64{
	RGContainsR,
	RContainsCNULL,
	CContainsSGNULL,
	SGContainsSNULL,
	SContainsTNULL,
}

var RGContainsR = map[uint64][]uint64{
	1: []uint64{10, 11, 12},
	2: []uint64{15},
	3: []uint64{12, 13, 14, 15},
}

var RContainsC = map[uint64][]uint64{
	11: []uint64{21, 22},
	12: []uint64{23},
	13: []uint64{21, 23},
	14: []uint64{24, 25},
	15: []uint64{26},
}

var CContainsSG = map[uint64][]uint64{
	22: []uint64{31},
	23: []uint64{32},
	24: []uint64{33},
	25: []uint64{34},
	26: []uint64{35},
}

var SGContainsS = map[uint64][]uint64{
	32: []uint64{41, 42},
	33: []uint64{43},
	34: []uint64{44, 45},
	35: []uint64{41, 45},
}

var SContainsT = map[uint64][]uint64{
	42: []uint64{51},
	43: []uint64{52, 53},
	44: []uint64{54, 55},
	45: []uint64{52, 53},
}

var RContainsCNULL = map[uint64][]uint64{
	10: []uint64{0},
	11: []uint64{21, 22},
	12: []uint64{23},
	13: []uint64{22, 23},
	14: []uint64{24, 25},
	15: []uint64{26},
}
var CContainsSGNULL = map[uint64][]uint64{
	21: []uint64{0},
	22: []uint64{31},
	23: []uint64{32},
	24: []uint64{33},
	25: []uint64{34},
	26: []uint64{35},
}
var SGContainsSNULL = map[uint64][]uint64{
	31: []uint64{0},
	32: []uint64{41, 42},
	33: []uint64{0},
	34: []uint64{44, 45},
	35: []uint64{45},
}

var SContainsTNULL = map[uint64][]uint64{
	41: []uint64{0},
	42: []uint64{51},
	43: []uint64{52, 53},
	44: []uint64{54, 55},
	45: []uint64{52, 53},
}

var expectIntegrity = map[uint64]LevelVaild{
	1: LevelVaild{Valid: false, InValidInfo: []*InvalidUnit{
		&InvalidUnit{InValidLevel: LevRule, InValidID: 10},
		&InvalidUnit{InValidLevel: LevConversation, InValidID: 21},
		&InvalidUnit{InValidLevel: LevSenGroup, InValidID: 31},
		&InvalidUnit{InValidLevel: LevSentence, InValidID: 41},
	}},
	2: LevelVaild{Valid: true},
	3: LevelVaild{Valid: false, InValidInfo: []*InvalidUnit{
		&InvalidUnit{InValidLevel: LevSenGroup, InValidID: 31},
		&InvalidUnit{InValidLevel: LevSenGroup, InValidID: 33},
		&InvalidUnit{InValidLevel: LevSentence, InValidID: 41},
	}},
}

var levelName = map[Levels]string{
	LevRuleGroup:    "rule group",
	LevRule:         "rule",
	LevConversation: "conversation",
	LevSenGroup:     "sentence group",
	LevSentence:     "sentence",
	LevTag:          "tag",
}

type mockSQLRelationDao struct {
}

func (m *mockSQLRelationDao) GetLevelRelationID(sql model.SqlLike, from int, to int, id []uint64, ignoreNULL bool) ([]map[uint64][]uint64, [][]uint64, error) {
	if ignoreNULL {
		return mockRel[from:to], nil, nil
	}
	return mockRelNULL[from:to], nil, nil
}

func mockRelationDoa() {
	relationDao = &mockSQLRelationDao{}
}
func TestGetLevelsRel(t *testing.T) {
	mockRelationDoa()
	mockDBLike := &test.MockDBLike{}
	dbLike = mockDBLike
	var from, to Levels

	from = LevRuleGroup
	to = LevRuleGroup
	id := []uint64{1, 3, 5}
	_, _, err := GetLevelsRel(from, to, id, true)
	if err == nil {
		t.Error("expecting get error, but get no error\n")
	}

	from = LevSentence
	to = LevRuleGroup

	_, _, err = GetLevelsRel(from, to, id, true)
	if err == nil {
		t.Error("expecting get error, but get no error\n")
	}

	from = LevRuleGroup
	to = LevSentence
	_, _, err = GetLevelsRel(from, to, id, true)
	if err != nil {
		t.Errorf("expecting no error, but get error %s\n", err)
	}

}

func TestCheckIntegrity(t *testing.T) {
	mockRelationDoa()
	mockDBLike := &test.MockDBLike{}
	dbLike = mockDBLike
	lev := LevRuleGroup
	ids := []uint64{1, 2, 3}
	levValid, err := CheckIntegrity(lev, ids)
	if err != nil {
		t.Fatalf("expecting no error, but get %s\n", err)
	}

	if len(ids) != len(levValid) {
		t.Fatalf("expecting %d len result, but get %d\n", len(ids), len(levValid))
	}

	for idx, id := range ids {
		expect := expectIntegrity[id]
		receive := levValid[idx]
		if expect.Valid != receive.Valid {
			t.Fatalf("expect %d get vaild %t, but get %t\n", ids[idx], expect.Valid, receive.Valid)
		}
		if !expect.Valid {
			if len(expect.InValidInfo) != len(receive.InValidInfo) {
				t.Errorf("expect get %d invalid info, but get %d\n", len(expect.InValidInfo), len(receive.InValidInfo))
			} else {
				for exIdx, expInfo := range expect.InValidInfo {
					recInfo := receive.InValidInfo[exIdx]
					if expInfo.InValidID != recInfo.InValidID {
						t.Errorf("expect get invalid ID %d, but get %d\n", expInfo.InValidID, recInfo.InValidID)
					}
					if expInfo.InValidLevel != recInfo.InValidLevel {
						t.Errorf("expect get level %d, but get %d\n", expInfo.InValidLevel, recInfo.InValidLevel)
					}
				}
			}
		}
	}

}
