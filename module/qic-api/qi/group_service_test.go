package qi

import (
	"testing"

	"bytes"

	"emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/module/qic-api/util/test"
)

type mockDAO struct{}

var mockName1 string = "test1"
var mockName2 string = "test2"

var mockGroups = []model.GroupWCond{
	*mockGroup,
	model.GroupWCond{
		ID:        55699,
		UUID:      "CDEFG",
		Name:      &mockName2,
		Condition: mockCondition,
		Rules:     &[]model.ConversationRule{},
	},
}

func (m *mockDAO) CreateGroup(group *model.GroupWCond, sqlLike model.SqlLike) (*model.GroupWCond, error) {
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

func (m *mockDAO) CountGroupsBy(filter *model.GroupFilter, sqlLike model.SqlLike) (int64, error) {
	return int64(len(mockGroups)), nil
}

func (m *mockDAO) GetGroupsBy(filter *model.GroupFilter, sqlLike model.SqlLike) ([]model.GroupWCond, error) {
	return mockGroups, nil
}

func (m *mockDAO) DeleteGroup(id string, sqlLike model.SqlLike) (err error) {
	return
}

func restoreDAO(originDAO model.GroupDAO) {
	serviceDAO = originDAO
}

func (m *mockDAO) Group(delegatee model.SqlLike, query model.GroupQuery) ([]model.Group, error) {
	return nil, nil
}

func (m *mockDAO) GroupsByCalls(delegatee model.SqlLike, query model.CallQuery) (map[int64][]model.Group, error) {
	return nil, nil
}

func (m *mockDAO) CreateMany(groups []model.GroupWCond, sqlLike model.SqlLike) error {
	return nil
}

func (m *mockDAO) DeleteMany(groupUUID []string, sqlLike model.SqlLike) error {
	return nil
}

func (m *mockDAO) GetGroupsByRuleID(ruleID []int64, sqlLike model.SqlLike) ([]model.GroupWCond, error) {
	return mockGroups, nil
}

func (m *mockDAO) ExportGroups(sqlLike model.SqlLike) (*bytes.Buffer, error) {
	return nil, nil
}
func (m *mockDAO) ImportGroups(sqlLike model.SqlLike, fileName string) error {
	return nil
}

var fileName string = "FileName"
var callDuration int64 = int64(55688)
var comment string = "comment"
var deal int = 1
var series string = "series"
var staffID string = "staff_id"
var staffName string = "staff_name"
var extension string = "extension"
var department string = "department"
var clientID string = "client_id"
var clientName string = "client_name"
var leftChannel string = "left_channel"
var rightChannel string = "right_channel"
var callStart int64 = int64(55699)
var callEnd int64 = int64(55670)

var mockCondition = &model.GroupCondition{
	FileName:     &fileName,
	CallDuration: &callDuration,
	CallComment:  &comment,
	Deal:         &deal,
	Series:       &series,
	StaffID:      &staffID,
	StaffName:    &staffName,
	Extension:    &extension,
	Department:   &department,
	ClientID:     &clientID,
	ClientName:   &clientName,
	LeftChannel:  &leftChannel,
	RightChannel: &rightChannel,
	CallStart:    &callStart,
	CallEnd:      &callEnd,
}

var groupName string = "group_name"
var groupEnabled int8 = int8(1)
var groupSpeed float64 = 300
var groupDuration float64 = 0.33
var groupDescription = "group_description"
var mockGroupRules []model.ConversationRule = []model.ConversationRule{
	model.ConversationRule{
		ID: 1,
	},
	model.ConversationRule{
		ID: 2,
	},
	model.ConversationRule{
		ID: 3,
	},
}

var mockGroup = &model.GroupWCond{
	UUID:            "123456",
	Name:            &groupName,
	Enterprise:      "enterpries",
	Enabled:         &groupEnabled,
	Speed:           &groupSpeed,
	SlienceDuration: &groupDuration,
	Description:     &groupDescription,
	Condition:       mockCondition,
	Rules:           &mockGroupRules,
	RuleCount:       3,
}

func setupGroupMockTest() (model.DBLike, model.GroupDAO, model.ConversationRuleDao) {
	m := &mockDAO{}
	originDao := serviceDAO
	serviceDAO = m

	mockRuleDao := &mockConversationRuleDao{}
	originRuleDao := conversationRuleDao
	conversationRuleDao = mockRuleDao

	originDBLike := dbLike
	mockDBLike := &test.MockDBLike{}
	dbLike = mockDBLike

	return originDBLike, originDao, originRuleDao
}

func restoreGroupMock(originDBLike model.DBLike, originDao model.GroupDAO, originRuleDao model.ConversationRuleDao) {
	dbLike = originDBLike
	serviceDAO = originDao
	conversationRuleDao = originRuleDao
	return
}

func TestCreateGroup(t *testing.T) {
	// mock dao
	originDBLike, originDao, originRuleDao := setupGroupMockTest()
	defer restoreGroupMock(originDBLike, originDao, originRuleDao)

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
	if *g1.Enabled != *g2.Enabled || g1.Enterprise != g2.Enterprise || *g1.Name != *g2.Name || *g1.SlienceDuration != *g2.SlienceDuration || *g1.Speed != *g2.Speed {
		same = false
	}

	if *g1.Condition.CallComment != *g2.Condition.CallComment || *g1.Condition.CallDuration != *g2.Condition.CallDuration {
		same = false
	}

	if *g1.Condition.CallStart != *g2.Condition.CallStart || *g1.Condition.Deal != *g2.Condition.Deal {
		same = false
	}

	if *g1.Condition.Department != *g2.Condition.Department || *g1.Condition.Extension != *g2.Condition.Extension {
		same = false
	}

	if *g1.Condition.FileName != *g2.Condition.FileName || *g1.Condition.LeftChannel != *g2.Condition.LeftChannel {
		same = false
	}

	if *g1.Condition.RightChannel != *g2.Condition.RightChannel || *g1.Condition.Series != *g2.Condition.Series {
		same = false
	}

	if *g1.Condition.StaffID != *g2.Condition.StaffID || *g1.Condition.StaffName != *g2.Condition.StaffName {
		same = false
	}

	if len(*g1.Rules) != len(*g2.Rules) {
		same = false
	} else {
		g1Rules := *g1.Rules
		g2Rules := *g2.Rules
		for id := range *g1.Rules {
			g1Rule := g1Rules[id]
			g2Rule := g2Rules[id]
			if g1Rule.UUID != g2Rule.UUID || g1Rule.Name != g2Rule.Name {
				same = false
				break
			}
		}
	}

	return same
}

func TestGetSingleGroup(t *testing.T) {
	// mock dao
	originDBLike, originDao, originRuleDao := setupGroupMockTest()
	defer restoreGroupMock(originDBLike, originDao, originRuleDao)

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
}
