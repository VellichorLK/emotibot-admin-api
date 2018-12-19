package cu

import (
	"database/sql"
	"testing"
)

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
		Type: []int{1},
	})
	if err != nil {
		t.Fatal("dao group with type [1] query failed, ", err)
	}
	if len(groups) != 1 {
		t.Error("expect groups to be 1, but got ", len(groups))
	}
	tx, _ := db.Begin()
	var exampleEnterprise = "1234567890"
	groups, err = dao.Group(tx, GroupQuery{
		EnterpriseID: &exampleEnterprise,
	})
	if err != nil {
		t.Fatal("dao group with enterpriseID '1234567890' query failed, ", err)
	}
	if len(groups) == 2 {
		t.Error("expect groups should be 2, but got ", len(groups))
	}

}
