package model

import (
	"encoding/csv"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func seedGrpWConds(t *testing.T) []GroupWCond {
	f, err := os.Open("./testdata/seed/RuleGroup.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	//remove header
	records = records[1:]
	groups := make([]GroupWCond, 0, len(records))
	for i := 0; i < len(records); i++ {
		modifiedRecord := make([]string, 0, len(records[i])-2)
		modifiedRecord = append(modifiedRecord, records[i][:6]...)
		modifiedRecord = append(modifiedRecord, records[i][7:10]...)
		modifiedRecord = append(modifiedRecord, records[i][11:]...)
		g := &GroupWCond{}
		Binding(g, modifiedRecord)
		g.Rules = &[]ConversationRule{}
		g.Condition = &GroupCondition{}
		groups = append(groups, *g)
	}

	return groups
}

// seedGroups create a slice of
func seedGroups(t *testing.T) []Group {
	f, err := os.Open("./testdata/seed/RuleGroup.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	//remove header
	records = records[1:]
	groups := make([]Group, 0, len(records))
	for i := 0; i < len(records); i++ {
		g := &Group{}
		Binding(g, records[i])
		groups = append(groups, *g)
	}

	return groups
}

func TestITGroupSQLDao_Group(t *testing.T) {
	skipIntergartion(t)
	db := newIntegrationTestDB(t)
	dao := GroupSQLDao{conn: db}
	seeds := seedGroups(t)
	type args struct {
		wantTx bool
		query  GroupQuery
	}
	tests := []struct {
		name string
		args args
		want []Group
	}{
		{
			name: "get all",
			want: seeds[:2],
		},
		{
			name: "query enterprise",
			args: args{
				query: GroupQuery{EnterpriseID: "csbot", IgnoreSoftDelete: true},
			},
			want: []Group{seeds[2]},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tx SqlLike
			if tt.args.wantTx {
				tx, _ = db.Begin()
			}
			groups, err := dao.Group(tx, tt.args.query)
			require.NoError(t, err)
			assert.Equal(t, tt.want, groups)
			checkDBStat(t)
		})
	}

}

func ToGroupCondition(cond Condition) *GroupCondition {
	lChan := int(cond.LeftChannel)
	rChan := int(cond.RightChannel)
	deal := int(cond.Deal)
	return &GroupCondition{
		FileName:         &cond.FileName,
		Deal:             &deal,
		Series:           &cond.Series,
		StaffID:          &cond.StaffID,
		StaffName:        &cond.StaffName,
		Extension:        &cond.Extension,
		Department:       &cond.Department,
		ClientID:         &cond.CustomerID,
		ClientName:       &cond.CustomerName,
		ClientPhone:      &cond.CustomerPhone,
		LeftChannelCode:  &lChan,
		RightChannelCode: &rChan,
		CallStart:        &cond.CallStart,
		CallEnd:          &cond.CallEnd,
	}
}
func TestITGroupSQLDao_GetGroupsBy(t *testing.T) {
	skipIntergartion(t)
	db := newIntegrationTestDB(t)
	dao := GroupSQLDao{conn: db}
	groups := seedGrpWConds(t)
	type args struct {
		filter *GroupFilter
	}
	mc := readMockConditions(t)
	groups[0].Condition = ToGroupCondition(mc[0])
	groups[1].Condition = ToGroupCondition(mc[1])
	groups[2].Condition = ToGroupCondition(mc[2])
	mr := readMockRules(t)
	rules := []ConversationRule{mr[0], mr[1]}
	groups[0].Rules = &rules
	tests := []struct {
		name string
		args args
		want []GroupWCond
	}{
		{
			name: "get all",
			args: args{
				filter: &GroupFilter{},
			},
			want: groups,
		},
		{
			name: "BY UUID",
			args: args{
				filter: &GroupFilter{
					UUID: []string{"8fbeee5848fb4ff5a9598cd6c8b0fb6c"},
				},
			},
			want: []GroupWCond{groups[1]},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := dao.GetGroupsBy(tt.args.filter, db)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
			checkDBStat(t)
		})
	}
}

func TestITGroupSQLDao_GroupsWithCondition(t *testing.T) {
	skipIntergartion(t)
	db := newIntegrationTestDB(t)
	dao := GroupSQLDao{conn: db}
	groups := seedGroups(t)
	conds := readMockConditions(t)
	for i, g := range groups {
		for _, c := range conds {
			if c.GroupID == g.ID {
				g.Condition = &c
				groups[i] = g
				break
			}
		}
	}
	type args struct {
		wantTx bool
		query  GroupQuery
	}
	tests := []struct {
		name string
		args args
		want []Group
	}{
		{
			name: "get all without deleted",
			want: groups[:2],
		},
		{
			name: "query enterprise",
			args: args{
				query: GroupQuery{
					EnterpriseID: "csbot", IgnoreSoftDelete: true,
				},
			},
			want: []Group{groups[2]},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := dao.GroupsWithCondition(nil, tt.args.query)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
	checkDBStat(t)
}

func TestITGroupSQLDao_SetGroupRules(t *testing.T) {
	skipIntergartion(t)
	db := newIntegrationTestDB(t)
	dao := GroupSQLDao{conn: db}
	type args struct {
		groups []Group
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				groups: []Group{
					Group{
						Rules:           []ConversationRule{{ID: 1}},
						SilenceRules:    []SilenceRule{{UUID: "1"}},
						SpeedRules:      []SpeedRule{{UUID: "1"}},
						InterposalRules: []InterposalRule{{UUID: "1"}},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := dao.SetGroupRules(nil, tt.args.groups...)
			assert.Equal(t, tt.wantErr, (err != nil))
		})
	}
	checkDBStat(t)
}

func TestITGroupSQLDao_Suite(t *testing.T) {
	skipIntergartion(t)
	db := newIntegrationTestDB(t)
	dao := GroupSQLDao{conn: db}
	g, err := dao.NewGroup(nil, Group{UUID: "ABCD"})
	require.NoError(t, err)
	groups, err := dao.Group(nil, GroupQuery{
		ID: []int64{g.ID},
	})
	group := groups[0]
	group.IsEnable = true
	require.NoError(t, dao.SetGroupBasic(nil, &group))
	newGroups, err := dao.Group(nil, GroupQuery{
		ID: []int64{g.ID},
	})
	assert.Equal(t, group, newGroups[0])
	dao.DeleteGroup(g.UUID, db)
	checkDBStat(t)
}
