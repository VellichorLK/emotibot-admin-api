package model

import (
	"encoding/csv"
	"os"
	"strconv"
	"testing"
)

func getCallsSeed(t *testing.T) []Call {
	f, err := os.Open("./testdata/seed/call.csv")
	if err != nil {
		t.Fatal("can not open call's testdata, ", err)
	}
	reader := csv.NewReader(f)
	rows, err := reader.ReadAll()
	if err != nil {
		t.Fatal("can not read call's testdata ", err)
	}
	var calls = make([]Call, 0)
	for i := len(rows[1:]); i >= 1; i-- {
		row := rows[i]
		var c Call
		c.ID, _ = strconv.ParseInt(row[0], 10, 64)
		taskID, _ := strconv.ParseInt(row[1], 10, 64)
		c.TaskID = taskID

		status, _ := strconv.ParseInt(row[2], 10, 8)
		c.Status = int8(status)
		c.UUID = row[3]
		if row[4] != "NULL" {
			c.FileName = &row[4]
		}

		if row[5] != "NULL" {
			c.FilePath = &row[5]
		}
		if row[6] != "NULL" {
			c.DemoFilePath = &row[6]
		}
		if row[7] != "NULL" {
			c.Description = &row[7]
		}
		c.DurationMillSecond, _ = strconv.Atoi(row[8])
		c.UploadUnixTime, _ = strconv.ParseInt(row[9], 10, 64)
		c.CallUnixTime, _ = strconv.ParseInt(row[10], 10, 64)
		c.StaffID = row[11]
		c.StaffName = row[12]
		c.Ext = row[13]
		c.Department = row[14]
		c.CustomerID = row[15]
		c.CustomerName = row[16]
		c.CustomerPhone = row[17]
		c.EnterpriseID = row[18]
		c.UploadUser = row[19]
		if row[20] != "NULL" {
			lst, _ := strconv.ParseFloat(row[20], 64)
			c.LeftSilenceTime = &lst
		}
		if row[21] != "NULL" {
			rst, _ := strconv.ParseFloat(row[21], 64)
			c.RightSilenceTime = &rst
		}
		if row[22] != "NULL" {
			ls, _ := strconv.ParseFloat(row[22], 64)
			c.LeftSpeed = &ls
		}
		if row[23] != "NULL" {
			rs, _ := strconv.ParseFloat(row[23], 64)
			c.RightSpeed = &rs
		}
		typ, _ := strconv.ParseInt(row[24], 10, 8)
		c.Type = int8(typ)
		lc, _ := strconv.ParseInt(row[25], 10, 8)
		c.LeftChanRole = int8(lc)
		rc, _ := strconv.ParseInt(row[26], 10, 8)
		c.RightChanRole = int8(rc)

		calls = append(calls, c)
	}
	return calls
}
