package model

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func TestITCallDaoCalls(t *testing.T) {
	skipIntergartion(t)
	db := newIntegrationTestDB(t)
	dao := CallSQLDao{
		db: db,
	}
	//testset is the reverted csv
	testset := getCallsSeed(t)
	var (
		timestart int64 = 1546598521
		timeEnd   int64 = timestart + 60
	)
	testTable := []struct {
		name string
		args CallQuery
		want []Call
	}{
		{
			"query all",
			CallQuery{},
			testset,
		},
		{
			"query id",
			CallQuery{ID: []int64{1}},
			[]Call{testset[2]},
		},
		{
			"query uuid",
			CallQuery{
				UUID: []string{"ec94dfd6e3974671b8a3533c752e51a6"},
			},
			[]Call{testset[1]},
		},
		{
			"query status",
			CallQuery{Status: []int8{CallStatusDone}},
			[]Call{testset[0], testset[1]},
		},
		{
			"query call time start",
			CallQuery{CallTimeStart: &timestart},
			testset,
		},
		{
			"query call time end",
			CallQuery{CallTimeEnd: &timeEnd},
			[]Call{testset[2]},
		},
		{
			"query call time range",
			CallQuery{
				CallTimeStart: &timestart,
				CallTimeEnd:   &timeEnd,
			},
			[]Call{testset[2]},
		},
		{
			"query staff id",
			CallQuery{StaffID: []string{"1"}},
			testset,
		},
		{
			"query call type",
			CallQuery{Typ: []int8{CallTypeWholeFile}},
			[]Call{testset[1], testset[2]},
		},
	}
	for _, tt := range testTable {
		t.Run(tt.name, func(t *testing.T) {
			calls, err := dao.Calls(nil, tt.args)
			if err != nil {
				t.Fatal("query calls expect to be ok, but got ", err)
			}
			assert.Equal(t, tt.want, calls)
		})
	}

}

// TestITCallDao_Suite test from insert & select result.
// TODO: implement DELETE for clean up to be repeatable testsuite.
func TestITCallDao_Suite(t *testing.T) {
	skipIntergartion(t)
	db := newIntegrationTestDB(t)
	dao := CallSQLDao{
		db: db,
	}
	input := []Call{Call{
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
	}}
	result, err := dao.NewCalls(nil, input)
	require.NoError(t, err, "expect insert to be ok")
	query := CallQuery{UUID: []string{input[0].UUID}}
	queryResult, err := dao.Calls(nil, query)
	require.NoError(t, err, "expect call query to be ok")
	assert.Equal(t, result, queryResult)
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
