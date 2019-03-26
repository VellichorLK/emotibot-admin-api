package qi

import (
	"bytes"
	"encoding/csv"
	"strconv"

	"emotibot.com/emotigo/module/qic-api/model/v1"
)

var exampleCallContent = []byte(`call_id,task_id,status,"call_uuid","file_name","file_path","demo_file_path","description",duration,upload_time,call_time,"staff_id","staff_name","extension","department","customer_id","customer_name","customer_phone","enterprise","uploader",left_silence_time,right_silence_time,left_speed,right_speed,type,left_channel,right_channel
1,1,0,"633f349535eb4d748eba577104776185","test1","/tmp/test1.wav",NULL,"testing case 1",120000,1546598521,1546598521,"1","Taylor","2222","Backend","2","Dean","123456789","csbot","bot",NULL,NULL,NULL,NULL,0,0,1
2,2,2,"ec94dfd6e3974671b8a3533c752e51a6","test2","/tmp/test2.wav","/tmp/test2.mp3","testing case 2",45000,1546598657,1546598657,"1","Taylor","2222","Backend","3","Ken","123456789","csbot","bot",12.31,1.3,30,30,0,1,0
`)

var exampleTaskContent = []byte(`task_id,status,"description",deal,"series",create_time,update_time,"creator","updator"
1,2,NULL,1,"47960",1546937289,1546937289,"",""
2,0,NULL,0,"25858",1546937302,1546937302,"",""
`)

type mockCallDao struct {
	mockdata     []byte
	mockTaskData []byte
}

func (m *mockCallDao) Calls(delegatee model.SqlLike, query model.CallQuery) ([]model.Call, error) {
	return m.readMockCallData(), nil
}
func (m *mockCallDao) NewCalls(delegatee model.SqlLike, calls []model.Call) ([]model.Call, error) {
	return m.readMockCallData()[1:], nil
}
func (m *mockCallDao) SetRuleGroupRelations(delegatee model.SqlLike, call model.Call, rulegroups []model.Group) ([]int64, error) {
	return []int64{1, 2}, nil
}

func (m *mockCallDao) Count(delegatee model.SqlLike, query model.CallQuery) (int64, error) {
	return int64(len(m.readMockCallData())), nil
}

func (m *mockCallDao) SetCall(delegatee model.SqlLike, call model.Call) error {
	return nil
}

func (m *mockCallDao) ExportCalls(delegatee model.SqlLike) (*bytes.Buffer, error) {
	// TODO: Return valid calls
	return nil, nil
}

func (m *mockCallDao) GetCallIDByUUID(delegatee model.SqlLike, callUUID string) (int64, error) {
	return 1, nil
}

func (m *mockCallDao) readMockCallData() []model.Call {
	reader := csv.NewReader(bytes.NewReader(m.mockdata))
	rows, err := reader.ReadAll()
	if err != nil {
		return nil
	}
	var calls = make([]model.Call, len(rows)-1)
	for i, row := range rows[1:] {
		c := calls[i]
		c.ID, _ = strconv.ParseInt(row[0], 10, 64)
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

		calls[i] = c
	}
	return calls
}
