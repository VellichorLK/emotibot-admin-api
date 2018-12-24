package cu

import (
	"database/sql"
	"fmt"
	"reflect"
	"testing"

	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
)

var goldenGroups = []Group{
	Group{
		AppID:          1,
		Name:           "testing",
		EnterpriseID:   "123456789",
		Description:    "this is an integration test data",
		CreatedTime:    0,
		UpdatedTime:    0,
		IsEnable:       false,
		IsDelete:       false,
		LimitedSpeed:   0,
		LimitedSilence: 0,
		typ:            0,
	},
	Group{
		AppID:          2,
		Name:           "testing2",
		EnterpriseID:   "123456789",
		Description:    "this is another integration test data",
		CreatedTime:    0,
		UpdatedTime:    0,
		IsEnable:       true,
		IsDelete:       false,
		LimitedSpeed:   0,
		LimitedSilence: 0,
		typ:            1,
	},
}

func TestIntegrationSQLDaoGroup(t *testing.T) {
	if !isIntegration {
		t.Skip("skip intergration test, please specify -intergation flag.")
	}
	db, err := sql.Open("mysql", "root:password@tcp(127.0.0.1)/QISYS?parseTime=true&loc=Asia%2FTaipei")
	if err != nil {
		t.Fatal("can not open mysql ", err)
	}
	dao := SQLDao{conn: db}
	groups, err := dao.Group(nil, GroupQuery{})
	if err != nil {
		t.Fatal("dao group executed failed, ", err)
	}
	if len(groups) != 2 {
		t.Error("expect groups should be 2, but got", len(groups))
	}
	groups, err = dao.Group(nil, GroupQuery{
		Type: []int{0},
	})
	if err != nil {
		t.Fatal("dao group with type [1] query failed, ", err)
	}
	if len(groups) != 1 {
		t.Error("expect groups to be 1, but got ", len(groups))
	}
	if !reflect.DeepEqual(groups[0], goldenGroups[0]) {
		fmt.Printf("%v\n%v\n", groups[0], goldenGroups[0])
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

func TestSQLDaoGroup(t *testing.T) {
	db, mocker, _ := sqlmock.New()
	serviceDao = SQLDao{
		conn: db,
	}
	rows := exampleGroups(goldenGroups)
	mocker.ExpectQuery("SELECT .+ FROM `" + tblGroup + "`").WillReturnRows(rows)
	groups, err := serviceDao.Group(nil, GroupQuery{})
	if err != nil {
		t.Fatal("expect dao.Group is ok, but got error: ", err)
	}
	if reflect.DeepEqual(goldenGroups, groups) {
		t.Fatal("expect result to be exact with golden group but no")
	}
}

func exampleGroups(examples []Group) *sqlmock.Rows {
	rows := sqlmock.NewRows([]string{fldGroupAppID, fldGroupIsDeleted, fldGroupName,
		fldGroupEnterprise, fldGroupDescription, fldGroupCreatedTime, fldGroupUpdatedTime,
		fldGroupIsEnabled, fldGroupLimitedSpeed, fldGroupLimitedSilence, fldGroupType})
	for _, e := range examples {
		rows.AddRow(e.AppID, e.IsDelete, e.Name,
			e.EnterpriseID, e.Description, e.CreatedTime, e.UpdatedTime,
			e.IsEnable, e.LimitedSpeed, e.LimitedSilence, e.typ)
	}

	return rows
}
