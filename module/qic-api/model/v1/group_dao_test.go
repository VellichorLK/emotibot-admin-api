package model

import (
	"fmt"
	"reflect"
	"testing"

	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestGetGroupsSQL(t *testing.T) {
	filter := GroupFilter{
		Deal:      -1,
		FileName:  "test.wav",
		Extension: "abcdefg",
	}

	targetStr := `SELECT rg.%s, rg.%s, rg.%s, rg.%s, rg.%s, rg.%s, rg.%s, rg.%s, rg.%s, 
	gc.%s, gc.%s, gc.%s, gc.%s, gc.%s, gc.%s, gc.%s, 
	gc.%s, gc.%s, gc.%s, gc.%s, gc.%s, gc.%s, gc.%s,
	r.%s as rUUID, r.%s as rName
	FROM (SELECT * FROM %s WHERE is_delete = ?) as rg
	INNER JOIN (SELECT * FROM %s WHERE file_name = ? and extension = ?) as gc on rg.%s = gc.%s
	LEFT JOIN %s as rrr ON rg.%s = rrr.%s
	LEFT JOIN %s as r on rrr.%s = r.%s
	`
	targetStr = fmt.Sprintf(
		targetStr,
		fldRuleGrpID,
		fldRuleGrpUUID,
		fldRuleGrpName,
		fldDescription,
		fldRuleGrpLimitSpeed,
		fldRuleGrpLimitSilence,
		fldCreateTime,
		fldRuleGrpIsEnable,
		fldEnterprise,
		RGCFileName,
		RGCDeal,
		RGCSeries,
		RGCStaffID,
		RGCStaffName,
		RGCExtension,
		RGCDepartment,
		RGCCustomerID,
		RGCCustomerName,
		RGCCustomerPhone,
		RGCCallStart,
		RGCCallEnd,
		RGCLeftChannel,
		RGCRightChannel,
		fldUUID,
		fldName,
		tblRuleGroup,
		tblRGC,
		fldID,
		RGCGroupID,
		tblRelGrpRule,
		fldID,
		RRRGroupID,
		tblConversationRule,
		RRRRuleID,
		fldID,
	)

	queryStr, values := getGroupsSQL(&filter)

	if len(values) != 3 {
		t.Error("expect values length 3")
		return
	}

	if targetStr != queryStr {
		t.Errorf("exptect %s but got %s", targetStr, queryStr)
		return
	}
}

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

func TestIntegrationGroupSQLDaoGroup(t *testing.T) {
	t.Skip("need to have the csv data to re-implement this test.")
	if !isIntegration {
		t.Skip("skip intergration test, please specify -intergation flag.")
	}
	db := newIntegrationTestDB(t)
	dao := GroupSQLDao{conn: db}
	groups, err := dao.Group(nil, GroupQuery{})
	if err != nil {
		t.Fatal("dao group executed failed, ", err)
	}
	if len(groups) != 2 {
		t.Error("expect groups should be 2, but got", len(groups))
	}
	groups, err = dao.Group(nil, GroupQuery{
		Type: []int8{0},
	})
	if err != nil {
		t.Fatal("dao group with type [1] query failed, ", err)
	}
	if !reflect.DeepEqual(groups, goldenGroups[:1]) {
		t.Error("expect group 0 be equal to goldenGroups 0")
	}
	tx, _ := db.Begin()
	var exampleEnterprise = "123456789"
	groups, err = dao.Group(tx, GroupQuery{
		EnterpriseID: &exampleEnterprise,
	})
	if err != nil {
		t.Fatal("dao group with enterpriseID '12345' query failed, ", err)
	}
	if len(groups) != 2 {
		t.Error("expect groups should be 2, but got ", len(groups))
	}
	if !reflect.DeepEqual(groups, goldenGroups) {
		fmt.Printf("%+v\n%+v\n", groups, goldenGroups)
		t.Error("expect group to be identical with golden group")
	}
}

func TestGroupSQLDaoGroup(t *testing.T) {
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
