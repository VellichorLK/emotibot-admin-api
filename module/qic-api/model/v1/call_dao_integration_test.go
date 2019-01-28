package model

import (
	"reflect"
	"testing"
)

func TestITCallDaoCalls(t *testing.T) {
	skipIntergartion(t)
	db := newIntegrationTestDB(t)
	dao := CallSQLDao{
		db: db,
	}
	testset := getCallsSeed(t)
	t.Logf("testset: %+v", testset)
	var (
		timestart int64 = 1546598521
		timeEnd   int64 = timestart + 60
	)

	testTable := map[string]struct {
		Input  CallQuery
		Output []Call
	}{
		"query all": {
			CallQuery{}, testset,
		},
		"query id": {
			CallQuery{ID: []int64{1}},
			testset[1:],
		},
		"query uuid": {
			CallQuery{
				UUID: []string{"ec94dfd6e3974671b8a3533c752e51a6"},
			},
			testset[:1],
		},
		"query status": {
			CallQuery{Status: []int8{CallStatusDone}},
			testset[:1],
		},
		"query call time start": {
			CallQuery{CallTimeStart: &timestart},
			testset,
		},
		"query call time end": {
			CallQuery{CallTimeEnd: &timeEnd},
			testset[1:],
		},
		"query call time range": {
			CallQuery{
				CallTimeStart: &timestart,
				CallTimeEnd:   &timeEnd,
			},
			testset[1:],
		},
		"query staff id": {
			CallQuery{StaffID: []string{"1"}},
			testset,
		},
	}
	for name, tc := range testTable {
		t.Run(name, func(tt *testing.T) {
			calls, err := dao.Calls(nil, tc.Input)
			if err != nil {
				tt.Fatal("query calls expect to be ok, but got ", err)
			}
			if !reflect.DeepEqual(calls, tc.Output) {
				tt.Logf("calls:\n%+v\nexpect output:\n%+v", calls, tc.Output)
				tt.Error("expect query result equal to the expect output, but not equal")
			}
		})
	}

}

func TestITCallDaoNewCall(t *testing.T) {
	skipIntergartion(t)
	db := newIntegrationTestDB(t)
	dao := CallSQLDao{
		db: db,
	}
	exampleCall := Call{
		UUID:               "d95c7d0eff8c49169c64a2225696423f",
		DurationMillSecond: 120,
		UploadUnixTime:     1546827856,
		CallUnixTime:       1546827000,
		StaffID:            "12345",
		StaffName:          "tester",
		Ext:                "66810",
		Department:         "backend",
		CustomerID:         "123",
		CustomerName:       "david",
		CustomerPhone:      "123456789",
		EnterpriseID:       "csbot",
		UploadUser:         "Berta",
		Type:               CallTypeWholeFile,
		LeftChanRole:       CallChanStaff,
		RightChanRole:      CallChanCustomer,
		Status:             CallStatusWaiting,
	}
	expectExampleCall := exampleCall
	expectExampleCall.ID = 3
	testtable := []struct {
		Name   string
		Input  []Call
		Query  CallQuery
		Output []Call
	}{
		{"", []Call{exampleCall}, CallQuery{UUID: []string{"d95c7d0eff8c49169c64a2225696423f"}}, []Call{expectExampleCall}},
	}

	for _, tc := range testtable {
		t.Run(tc.Name, func(tt *testing.T) {
			result, err := dao.NewCalls(nil, tc.Input)
			if err != nil {
				tt.Fatal("expect new calls to be ok, but got ", err)
			}
			if !reflect.DeepEqual(result, tc.Output) {
				tt.Logf("compare with expect output failed:\n%+v\n%+v", result, tc.Output)
				tt.Error("expect result to be same with output")
			}
			queryResult, err := dao.Calls(nil, tc.Query)
			if err != nil {
				tt.Fatal("expect call query to be ok, but got ", err)
			}
			if !reflect.DeepEqual(result, queryResult) {
				tt.Logf("compare with query failed:\n%+v\n%+v\n", result, queryResult)
				tt.Error("expect query back to be same ")
			}
		})
	}
}

func TestITCallDaoSetRuleGroupRelations(t *testing.T) {
	skipIntergartion(t)
	db := newIntegrationTestDB(t)
	dao := CallSQLDao{
		db: db,
	}
	set := getCallsSeed(t)
	idGroup, err := dao.SetRuleGroupRelations(nil, set[0], []Group{
		Group{ID: 1}, Group{ID: 2},
	})
	if err != nil {
		t.Fatal("expect set releation to be ok, but got ", err)
	}
	if len(idGroup) != 2 {
		t.Error("expect get two id in result, but got ", len(idGroup))
	}
}
