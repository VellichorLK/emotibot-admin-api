package model

import (
	"reflect"
	"regexp"
	"testing"

	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestGetGroupsSQL(t *testing.T) {
	filter := GroupFilter{
		FileName:  "test.wav",
		Extension: "abcdefg",
	}
	matcher := regexp.MustCompile("SELECT .* FROM .*RuleGroup.* LEFT JOIN .*Relation_RuleGroup_Rule.* LEFT JOIN .*Rule.*")
	queryStr, values := getGroupsSQL(&filter)

	if len(values) != 2 {
		t.Error("expect values length 2")
		return
	}

	if len(matcher.FindAllString(queryStr, -1)) < 1 {
		t.Error("expect a valid querystring, but got ", queryStr)
		return
	}
}

func TestGroupSQLDaoGroup(t *testing.T) {
	var goldenGroups = []Group{
		Group{
			ID:             1,
			Name:           "testing",
			EnterpriseID:   "123456789",
			Description:    "this is an integration test data",
			CreatedTime:    0,
			UpdatedTime:    0,
			IsEnable:       false,
			IsDelete:       false,
			LimitedSpeed:   0,
			LimitedSilence: 0,
			Typ:            0,
		},
		Group{
			ID:             2,
			Name:           "testing2",
			EnterpriseID:   "123456789",
			Description:    "this is another integration test data",
			CreatedTime:    0,
			UpdatedTime:    0,
			IsEnable:       true,
			IsDelete:       false,
			LimitedSpeed:   0,
			LimitedSilence: 0,
			Typ:            1,
		},
	}
	db, mocker, _ := sqlmock.New()
	serviceDao := &GroupSQLDao{
		conn: db,
	}
	rows := exampleGroups(goldenGroups)
	mocker.ExpectQuery("SELECT .+ FROM `" + tblRuleGroup + "`").WillReturnRows(rows)
	groups, err := serviceDao.Group(nil, GroupQuery{})
	if err != nil {
		t.Fatal("expect dao.Group is ok, but got error: ", err)
	}
	if reflect.DeepEqual(goldenGroups, groups) {
		t.Fatal("expect result to be exact with golden group but no")
	}
}

func exampleGroups(examples []Group) *sqlmock.Rows {
	groupCols := []string{
		fldRuleGrpID, fldRuleGrpIsDelete, fldRuleGrpName,
		fldRuleGrpEnterpriseID, fldRuleGrpDescription, fldRuleGrpCreateTime,
		fldRuleGrpUpdateTime, fldRuleGrpIsEnable, fldRuleGrpLimitSpeed,
		fldRuleGrpLimitSilence, fldRuleGrpType,
	}
	rows := sqlmock.NewRows(groupCols)
	for _, e := range examples {
		rows.AddRow(e.ID, e.IsDelete, e.Name,
			e.EnterpriseID, e.Description, e.CreatedTime, e.UpdatedTime,
			e.IsEnable, e.LimitedSpeed, e.LimitedSilence, e.Typ)
	}

	return rows
}
