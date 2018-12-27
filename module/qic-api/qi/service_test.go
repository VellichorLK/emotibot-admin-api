package qi

import (
	"database/sql"
	"testing"

	"emotibot.com/emotigo/module/qic-api/model/v1"
)

type mockDAO struct{}

var mockGroups = []model.GroupWCond{
	model.GroupWCond{
		ID:   55688,
		UUID: "ABCDE",
		Name: "test1",
	},
	model.GroupWCond{
		ID:   55699,
		UUID: "CDEFG",
		Name: "test2",
	},
}

func (m *mockDAO) Begin() (*sql.Tx, error) {
	return nil, nil
}

func (m *mockDAO) Commit(tx *sql.Tx) error {
	return nil
}

func (m *mockDAO) ClearTranscation(tx *sql.Tx) {}

func (m *mockDAO) CreateGroup(group *model.GroupWCond, tx *sql.Tx) (*model.GroupWCond, error) {
	createdGroup := &model.GroupWCond{
		ID:              55688,
		UUID:            "abcde",
		Name:            group.Name,
		Enterprise:      group.Enterprise,
		Enabled:         group.Enabled,
		Speed:           group.Speed,
		SlienceDuration: group.SlienceDuration,
		Condition:       group.Condition,
		Rules:           group.Rules,
	}
	return createdGroup, nil
}

func (m *mockDAO) GetGroupBy(id string) (*model.GroupWCond, error) {
	if id == mockGroups[0].UUID {
		mockGroup.ID = mockGroups[0].ID
		return mockGroup, nil
	} else {
		return nil, nil
	}
}

func (m *mockDAO) CountGroupsBy(filter *model.GroupFilter) (int64, error) {
	return int64(len(mockGroups)), nil
}

func (m *mockDAO) GetGroupsBy(filter *model.GroupFilter) ([]model.GroupWCond, error) {
	return mockGroups, nil
}

func (m *mockDAO) DeleteGroup(id string, tx *sql.Tx) (err error) {
	return
}

func restoreDAO(originDAO model.GroupDAO) {
	serviceDAO = originDAO
}

var mockCondition = &model.GroupCondition{
	FileName:     "FileName",
	CallDuration: 55688,
	CallComment:  "comment",
	Deal:         1,
	Series:       "series",
	StaffID:      "staff_id",
	StaffName:    "staff_name",
	Extension:    "extension",
	Department:   "department",
	ClientID:     "client_id",
	ClientName:   "client_name",
	LeftChannel:  "left_channel",
	RightChannel: "right_channel",
	CallStart:    55699,
	CallEnd:      55670,
}

var mockGroup = &model.GroupWCond{
	Name:            "group_name",
	Enterprise:      "enterpries",
	Enabled:         1,
	Speed:           300,
	SlienceDuration: 0.33,
	Condition:       mockCondition,
	Rules:           []int64{1, 2, 3},
}

func TestCreateGroup(t *testing.T) {
	// mock dao
	originDAO := serviceDAO
	m := &mockDAO{}
	serviceDAO = m
	defer restoreDAO(originDAO)

	createdGroup, err := CreateGroup(mockGroup)
	if err != nil {
		t.Error(err)
		return
	}

	if createdGroup == nil {
		t.Error("created group is nil")
		return
	}

	if createdGroup.ID != 55688 || !sameGroup(mockGroup, createdGroup) {
		t.Error("created group fail")
		return
	}
}

func sameGroup(g1, g2 *model.GroupWCond) bool {
	same := true
	if g1.Enabled != g2.Enabled || g1.Enterprise != g2.Enterprise || g1.Name != g2.Name || g1.SlienceDuration != g2.SlienceDuration || g1.Speed != g2.Speed {
		same = false
	}

	if g1.Condition.CallComment != g2.Condition.CallComment || g1.Condition.CallDuration != g2.Condition.CallDuration {
		same = false
	}

	if g1.Condition.CallStart != g2.Condition.CallStart || g1.Condition.Deal != g2.Condition.Deal {
		same = false
	}

	if g1.Condition.Department != g2.Condition.Department || g1.Condition.Extension != g2.Condition.Extension {
		same = false
	}

	if g1.Condition.FileName != g2.Condition.FileName || g1.Condition.LeftChannel != g2.Condition.LeftChannel {
		same = false
	}

	if g1.Condition.RightChannel != g2.Condition.RightChannel || g1.Condition.Series != g2.Condition.Series {
		same = false
	}

	if g1.Condition.StaffID != g2.Condition.StaffID || g1.Condition.StaffName != g2.Condition.StaffName {
		same = false
	}

	if len(g1.Rules) != len(g2.Rules) {
		same = false
	} else {
		for id := range g1.Rules {
			if g1.Rules[id] != g2.Rules[id] {
				same = false
				break
			}
		}
	}

	return same
}

func TestGetSingleGroup(t *testing.T) {
	// mock dao
	originDAO := serviceDAO
	m := &mockDAO{}
	serviceDAO = m
	defer restoreDAO(originDAO)

	group, err := GetGroupBy(mockGroups[0].UUID)
	if err != nil {
		t.Error(err)
		return
	}

	if group == nil {
		t.Error("get nil group")
		return
	}

	if group.ID != mockGroups[0].ID {
		t.Errorf("expect group id: %d, but get %d", mockGroups[0].ID, group.ID)
		return
	}

	group, err = GetGroupBy(mockGroups[1].UUID)
	if err != nil {
		t.Error(err)
		return
	}

	if group != nil {
		t.Errorf("expect nil group but get one")
		return
	}
}
