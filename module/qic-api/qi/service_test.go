package qi

import (
	"database/sql"
	"testing"
)

type mockDAO struct{}


var mockGroups []Group = []Group{
	Group{
		ID: 55688,
		Name: "test1",
	},
	Group{
		ID: 55699,
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

func (m *mockDAO) GetGroups() ([]Group, error) {
	return mockGroups, nil
}

func (m *mockDAO) CreateGroup(group *Group, tx *sql.Tx) (*Group, error) {
	createdGroup := &Group{
		ID: 55688,
		Name: group.Name,
		Enterprise: group.Enterprise,
		Enabled: group.Enabled,
		Speed: group.Speed,
		SlienceDuration: group.SlienceDuration,
		Condition: group.Condition,
	}
	return createdGroup, nil
}

func restoreDAO (originDAO DAO) {
	serviceDAO = originDAO
}

var mockCondition *GroupCondition = &GroupCondition{
	FileName: "FileName",
	CallDuration: 55688,
	CallComment: "comment",
	Deal: 1,
	Series: "series",
	StaffID: "staff_id",
	StaffName: "staff_name",
	Extension: "extension",
	Department: "department",
	ClientID: "client_id",
	ClientName: "client_name",
	LeftChannel: "left_channel",
	RightChannel: "right_channel",
	CallStart: 55699,
	CallEnd: 55670,
}

var mockGroup *Group = &Group {
	Name: "group_name",
	Enterprise: "enterpries",
	Enabled: 1,
	Speed: 300,
	SlienceDuration: 0.33,
	Condition: mockCondition,
}

func TestGetGroups(t *testing.T) {
	// mock dao
	originDAO := serviceDAO
	m := &mockDAO{}
	serviceDAO = m
	defer restoreDAO(originDAO)

	groups, _ := GetGroups()
	if len(groups) != 2 {
		t.Error("expect 2 groups but got ", len(groups))
	}

	for idx := range groups {
		g := groups[idx]
		targetG := mockGroups[idx]
		if g.ID != targetG.ID || g.Name != targetG.Name {
			t.Error("expect ", targetG.ID, " ", targetG.Name, "but got ", g.ID, " ", g.Name)
		}
	}
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

func sameGroup(g1, g2 *Group) bool {
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

	return same
}