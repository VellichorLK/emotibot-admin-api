package model

import (
	"encoding/csv"
	"os"
	"reflect"
	"strconv"
	"testing"
)

func getCallsSeed(t *testing.T) []Call {
	f, err := os.Open("./testdata/call.csv")
	if err != nil {
		t.Fatal("can not open call's testdata, ", err)
	}
	reader := csv.NewReader(f)
	rows, err := reader.ReadAll()
	if err != nil {
		t.Fatal("can not read call's testdata ", err)
	}
	var calls = make([]Call, len(rows)-1)
	for i, row := range rows[1:] {
		c := calls[i]
		c.ID, _ = strconv.ParseUint(row[0], 10, 64)
		status, _ := strconv.ParseInt(row[1], 10, 8)
		c.Status = int8(status)
		c.UUID = row[2]
		if row[3] != "NULL" {
			c.FileName = &row[3]
		}
		if row[4] != "NULL" {
			c.FilePath = &row[4]
		}
		if row[5] != "NULL" {
			c.Description = &row[5]
		}
		c.DurationSecond, _ = strconv.Atoi(row[6])
		c.UploadUnixTime, _ = strconv.ParseInt(row[7], 10, 64)
		c.CallUnixTime, _ = strconv.ParseInt(row[8], 10, 64)
		c.StaffID = row[9]
		c.StaffName = row[10]
		c.Ext = row[11]
		c.Department = row[12]
		c.CustomerID = row[13]
		c.CustomerName = row[14]
		c.CustomerPhone = row[15]
		c.EnterpriseID = row[16]
		c.UploadUser = row[17]
		if row[18] != "NULL" {
			lst, _ := strconv.ParseFloat(row[18], 64)
			c.LeftSilenceTime = &lst
		}
		if row[19] != "NULL" {
			rst, _ := strconv.ParseFloat(row[19], 64)
			c.RightSilenceTime = &rst
		}
		if row[20] != "NULL" {
			ls, _ := strconv.ParseFloat(row[20], 64)
			c.LeftSpeed = &ls
		}
		if row[21] != "NULL" {
			rs, _ := strconv.ParseFloat(row[21], 64)
			c.RightSpeed = &rs
		}
		typ, _ := strconv.ParseInt(row[22], 10, 8)
		c.Type = int8(typ)
		lc, _ := strconv.ParseInt(row[23], 10, 8)
		c.LeftChanRole = int8(lc)
		rc, _ := strconv.ParseInt(row[24], 10, 8)
		c.RightChanRole = int8(rc)
		calls[i] = c
	}
	return calls
}

func TestCallDaoCallsIntegrations(t *testing.T) {
	skipIntergartion(t)
	db := newIntegrationTestDB(t)
	dao := CallSQLDao{
		db: db,
	}
	testset := getCallsSeed(t)
	t.Logf("testset: %+v", testset)
	testTable := map[string]struct {
		Input  CallQuery
		Output []Call
	}{
		"query all": {
			CallQuery{}, testset,
		},
		"query id": {
			CallQuery{ID: []uint64{1}},
			testset[:1],
		},
		"query uuid": {
			CallQuery{
				UUID: []string{"ec94dfd6e3974671b8a3533c752e51a6"},
			},
			testset[0:],
		},
		"query status": {
			CallQuery{Status: []int8{CallStatusDone}},
			testset[0:],
		},
	}
	for name, tc := range testTable {
		t.Run(name, func(tt *testing.T) {
			calls, err := dao.Calls(nil, tc.Input)
			if err != nil {
				t.Fatal("query calls expect to be ok, but got ", err)
			}
			if reflect.DeepEqual(calls, tc.Output) {
				t.Error("expect query result equal to the expect output, but not equal")
			}
		})
	}

}
