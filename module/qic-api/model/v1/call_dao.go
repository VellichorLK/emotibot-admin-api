package model

import (
	"database/sql"
	"fmt"
	"strings"

	"bufio"
	"bytes"
	"reflect"

	"emotibot.com/emotigo/pkg/logger"
	"github.com/tealeg/xlsx"
)

type CallDao interface {
	Calls(delegatee SqlLike, query CallQuery) ([]Call, error)
	NewCalls(delegatee SqlLike, calls []Call) ([]Call, error)
	SetRuleGroupRelations(delegatee SqlLike, call Call, rulegroups []Group) ([]int64, error)
	SetCall(delegatee SqlLike, call Call) error
	Count(delegatee SqlLike, query CallQuery) (int64, error)
	ExportCalls(delegatee SqlLike) (*bytes.Buffer, error)
	GetCallIDByUUID(delegatee SqlLike, callUUID string) (int64, error)
}

//CallSQLDao is the sql implements of the call table
type CallSQLDao struct {
	db *sql.DB
}

func NewCallSQLDao(db *sql.DB) *CallSQLDao {
	return &CallSQLDao{
		db: db,
	}
}

//CallQuery is the query to get call table
type CallQuery struct {
	ID            []int64
	UUID          []string
	Status        []int8
	CallTime      RangeCondition
	Typ           []int8
	StaffID       []string
	EnterpriseID  *string
	CustomerPhone *string
	DealStatus    *int8
	Ext           *string
	Department    *string
	Paging        *Pagination
}

func (c *CallQuery) whereSQL(prefix string) (string, []interface{}) {
	var (
		rawSQL  string
		builder = NewWhereBuilder(andLogic, prefix)
	)
	builder.In(fldCallID, int64ToWildCard(c.ID...))
	builder.In(fldCallUUID, stringToWildCard(c.UUID...))
	builder.In(fldCallStatus, int8ToWildCard(c.Status...))
	builder.In(fldCallType, int8ToWildCard(c.Typ...))
	builder.Between(fldCallCallTime, c.CallTime)
	builder.In(fldCallStaffID, stringToWildCard(c.StaffID...))
	if c.EnterpriseID != nil {
		builder.Eq(fldCallEnterprise, *c.EnterpriseID)
	}
	if c.CustomerPhone != nil {
		builder.Eq(fldCallCustomerPhone, *c.CustomerPhone)
	}
	// deal status need to query the task, we will implement this later
	// if c.IsDealStatus != nil {
	// 	cond := fmt.Sprintf("`%s`=?", f)
	// 	conditions = append(conditions, cond)
	// 	bindData = append(bindData, c.CustomerPhone)
	// }

	if c.Ext != nil {
		builder.Eq(fldCallExt, *c.Ext)
	}
	if c.Department != nil {
		builder.Eq(fldCallDepartment, *c.Department)
	}
	rawSQL, bindData := builder.Parse()
	if len(bindData) > 0 {
		rawSQL = " WHERE " + rawSQL
	}
	return rawSQL, bindData
}

// Call represent the call table of the QISYS database.
// Any pointer field is nullable in the schema.Call
// Ext(分機號碼) is the receiver(staff) extension number.
type Call struct {
	ID                 int64
	Status             int8
	UUID               string
	FileName           *string
	FilePath           *string
	DemoFilePath       *string
	Description        *string
	DurationMillSecond int
	UploadUnixTime     int64
	CallUnixTime       int64
	StaffID            string
	StaffName          string
	Ext                string
	Department         string
	CustomerID         string
	CustomerName       string
	CustomerPhone      string
	EnterpriseID       string
	UploadUser         string
	LeftSilenceTime    *float64
	RightSilenceTime   *float64
	LeftSpeed          *float64
	RightSpeed         *float64
	Type               int8
	LeftChanRole       int8
	RightChanRole      int8
	IsDeal             int8
	RemoteFile         *string
}

// the type of the call is created, different type indicate different incoming source of call.
// 	- 0: whole audio upload api (Default)
// 	- 1: realtime audio upload api
const (
	CallTypeWholeFile int8 = 0
	CallTypeRealTime  int8 = 1
)

// Channel type of the call
//	- 0: staff(客服)
//	- 1: customer(客戶)
// 	- 9: default
const (
	CallChanStaff int8 = iota
	CallChanCustomer
	CallChanDefault = 9
)

// asr status types of the call
//	- 0: waiting
//	- 1: running
//	- 2: done
//	- 9: failed
const (
	CallStatusWaiting int8 = iota
	CallStatusRunning
	CallStatusDone
	CallStatusFailed = 9
)

// Realtime call source type
// - 0: Remote .wav file
// - 1: text
const (
	CallSourceRemoteWav int8 = iota
	CallSourceText
)

func ValidCallStatus(status int8) bool {
	switch status {
	case CallStatusWaiting:
		fallthrough
	case CallStatusRunning:
		fallthrough
	case CallStatusDone:
		fallthrough
	case CallStatusFailed:
		return true
	default:
		return false
	}
}

// Calls get the query result of call resource.
func (c *CallSQLDao) Calls(delegatee SqlLike, query CallQuery) ([]Call, error) {
	if delegatee == nil {
		delegatee = c.db
	}
	selectCols := []string{fldCallID, fldCallUUID, fldCallFileName,
		fldCallFilePath, fldCallDescription, fldCallDuration,
		fldCallUploadTime, fldCallCallTime, fldCallStaffID, fldCallStaffName,
		fldCallExt, fldCallDepartment, fldCallCustomerID,
		fldCallCustomerName, fldCallCustomerPhone, fldCallEnterprise,
		fldCallUploadedUser, fldCallLeftSilenceTime, fldCallRightSilenceTime,
		fldCallLeftSpeed, fldCallRightSpeed, fldCallType,
		fldCallLeftChan, fldCallRightChan, fldCallStatus,
		fldCallDeal, fldCallDemoFilePath, fldCallRemoteFile,
	}
	wheresql, data := query.whereSQL("")
	limitsql := ""
	if query.Paging != nil {
		query.Paging.offsetSQL()
	}
	rawquery := "SELECT `" + strings.Join(selectCols, "`,`") + "` FROM `" + tblCall + "` " + wheresql + " " + limitsql + " ORDER BY `" + fldCallID + "` DESC"
	rows, err := delegatee.Query(rawquery, data...)
	if err != nil {
		logger.Error.Println("error raw sql", rawquery)
		return nil, fmt.Errorf("select call query failed, %v", err)
	}
	defer rows.Close()
	var calls []Call
	for rows.Next() {
		var (
			c           Call
			fileName    sql.NullString
			filePath    sql.NullString
			description sql.NullString
			leftSTime   sql.NullFloat64
			rightSTime  sql.NullFloat64
			lSpeed      sql.NullFloat64
			rSpeed      sql.NullFloat64
			demoFp      sql.NullString
			remoteFp    sql.NullString
		)
		err := rows.Scan(&c.ID, &c.UUID, &fileName,
			&filePath, &description, &c.DurationMillSecond,
			&c.UploadUnixTime, &c.CallUnixTime, &c.StaffID, &c.StaffName,
			&c.Ext, &c.Department, &c.CustomerID,
			&c.CustomerName, &c.CustomerPhone, &c.EnterpriseID,
			&c.UploadUser, &leftSTime, &rightSTime,
			&lSpeed, &rSpeed, &c.Type,
			&c.LeftChanRole, &c.RightChanRole, &c.Status,
			&c.IsDeal, &demoFp, &remoteFp,
		)
		if err != nil {
			return nil, fmt.Errorf("scan error: %v", err)
		}
		if fileName.Valid {
			c.FileName = &fileName.String
		}
		if filePath.Valid {
			c.FilePath = &filePath.String
		}
		if description.Valid {
			c.Description = &description.String
		}
		if leftSTime.Valid {
			c.LeftSilenceTime = &leftSTime.Float64
		}
		if rightSTime.Valid {
			c.RightSilenceTime = &rightSTime.Float64
		}
		if rSpeed.Valid {
			c.RightSpeed = &rSpeed.Float64
		}
		if lSpeed.Valid {
			c.LeftSpeed = &lSpeed.Float64
		}
		if demoFp.Valid {
			c.DemoFilePath = &demoFp.String
		}
		if remoteFp.Valid {
			c.RemoteFile = &remoteFp.String
		}

		calls = append(calls, c)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("scan error %v", err)
	}
	return calls, nil
}

// NewCalls create the Call based on the Call struct.
// Call ID will be ignored, and assigned to new one if sql-driver support it.
func (c *CallSQLDao) NewCalls(delegatee SqlLike, calls []Call) ([]Call, error) {
	if delegatee == nil {
		delegatee = c.db
	}
	insertCols := []string{fldCallUUID, fldCallFileName,
		fldCallFilePath, fldCallDescription, fldCallDuration,
		fldCallUploadTime, fldCallCallTime, fldCallStaffID, fldCallStaffName,
		fldCallExt, fldCallDepartment, fldCallCustomerID,
		fldCallCustomerName, fldCallCustomerPhone, fldCallEnterprise,
		fldCallUploadedUser, fldCallLeftSilenceTime, fldCallRightSilenceTime,
		fldCallLeftSpeed, fldCallRightSpeed, fldCallType,
		fldCallLeftChan, fldCallRightChan, fldCallDeal, fldCallRemoteFile,
	}

	rawquery := "INSERT INTO `" + tblCall + "` (`" + strings.Join(insertCols, "`, `") + "`) VALUE(?" + strings.Repeat(",? ", len(insertCols)-1) + ")"
	stmt, err := delegatee.Prepare(rawquery)
	if err != nil {
		return nil, fmt.Errorf("statement prepare failed")
	}
	defer stmt.Close()
	var (
		hasSupportID = true
	)
	for i, c := range calls {

		r, err := stmt.Exec(c.UUID, c.FileName, c.FilePath,
			c.Description, c.DurationMillSecond, c.UploadUnixTime,
			c.CallUnixTime, c.StaffID, c.StaffName,
			c.Ext, c.Department, c.CustomerID,
			c.CustomerName, c.CustomerPhone, c.EnterpriseID,
			c.UploadUser, c.LeftSilenceTime, c.RightSilenceTime,
			c.LeftSpeed, c.RightSpeed, c.Type,
			c.LeftChanRole, c.RightChanRole, c.IsDeal, c.RemoteFile)
		if err != nil {
			return nil, fmt.Errorf("create new call failed, %v", err)
		}
		id, err := r.LastInsertId()
		if err != nil {
			hasSupportID = false
			calls[i] = c
			continue
		}
		c.ID = id
		calls[i] = c
	}
	if !hasSupportID {
		err = ErrAutoIDDisabled
	}
	return calls, err
}

// SetRuleGroupRelations insert relations between input call and rulegroups.
// id is the created relation table's id group
// err, if LastInsertId is not support for current driver, it will return ErrAutoIDDisabled.
// Which user need to seem id as unreliable source and re-query again.
func (c *CallSQLDao) SetRuleGroupRelations(delegatee SqlLike, call Call, rulegroups []Group) (id []int64, err error) {
	if delegatee == nil {
		delegatee = c.db
	}
	insertCols := []string{fldCRGRelCallID, fldCRGRelRuleGroupID}
	rawQuery := "INSERT INTO `" + tblRelCallRuleGrp + "` (`" + strings.Join(insertCols, "`, `") + "`) VALUE (?" + strings.Repeat(",?", len(insertCols)-1) + ")"
	stmt, err := delegatee.Prepare(rawQuery)
	if err != nil {
		return nil, fmt.Errorf("sql prepare failed, %v", err)
	}
	defer stmt.Close()
	idGroup := make([]int64, 0, len(rulegroups))
	hasSupportID := true
	for _, r := range rulegroups {
		result, err := stmt.Exec(call.ID, r.ID)
		if err != nil {
			return nil, fmt.Errorf("create " + tblRelCallRuleGrp + " failed")
		}
		id, err := result.LastInsertId()
		if err != nil {
			hasSupportID = false
			continue
		}
		idGroup = append(idGroup, id)
	}

	if !hasSupportID {
		return idGroup, ErrAutoIDDisabled
	}

	return idGroup, nil
}

func (c *CallSQLDao) SetCall(delegatee SqlLike, call Call) error {
	if delegatee == nil {
		delegatee = c.db
	}
	updatepart, data := createCallUpdateSQL(call)
	rawquery := "UPDATE `" + tblCall + "` SET " + updatepart + " WHERE `" + fldCallID + "` = ?"

	data = append(data, call.ID)
	_, err := delegatee.Exec(rawquery, data...)

	if err != nil {
		return fmt.Errorf("update execute failed, %v", err)
	}
	return nil

}

func (c *CallSQLDao) Count(delegatee SqlLike, query CallQuery) (int64, error) {
	if delegatee == nil {
		delegatee = c.db
	}
	wheresql, data := query.whereSQL("")
	rawquery := "SELECT count(*) FROM `" + tblCall + "` " + wheresql
	var count int64
	err := delegatee.QueryRow(rawquery, data...).Scan(&count)
	if err != nil {
		logger.Error.Println("raw error sql: ", rawquery)
		return 0, fmt.Errorf("query failed, %v", err)
	}
	return count, nil
}
func createCallUpdateSQL(c Call) (string, []interface{}) {
	parts := []string{}
	updateCols := []string{
		fldCallUUID, fldCallDuration, fldCallUploadTime,
		fldCallFileName, fldCallFilePath, fldCallDescription,
		fldCallCallTime, fldCallStaffID, fldCallStaffName,
		fldCallExt, fldCallDepartment, fldCallCustomerID,
		fldCallCustomerName, fldCallCustomerPhone, fldCallEnterprise,
		fldCallUploadedUser, fldCallType, fldCallLeftChan,
		fldCallRightChan, fldCallStatus, fldCallDemoFilePath,
		fldCallDeal, fldCallLeftSilenceTime, fldCallRightSilenceTime,
		fldCallLeftSpeed, fldCallRightSpeed, fldCallRemoteFile,
	}

	data := []interface{}{
		c.UUID, c.DurationMillSecond, c.UploadUnixTime,
		c.FileName, c.FilePath, c.Description,
		c.CallUnixTime, c.StaffID, c.StaffName,
		c.Ext, c.Department, c.CustomerID,
		c.CustomerName, c.CustomerPhone, c.EnterpriseID,
		c.UploadUser, c.Type, c.LeftChanRole,
		c.RightChanRole, c.Status, c.DemoFilePath,
		c.IsDeal, c.LeftSilenceTime, c.RightSilenceTime,
		c.LeftSpeed, c.RightSpeed, c.RemoteFile,
	}

	for _, colName := range updateCols {
		p := fmt.Sprintf("`%s` = ?", colName)
		parts = append(parts, p)
	}

	rawsql := strings.Join(parts, " , ")
	return rawsql, data

}

func (c *CallSQLDao) ExportCalls(delegatee SqlLike) (*bytes.Buffer, error) {

	xlFile := xlsx.NewFile()
	fmt.Println(xlFile)

	var queryStr string
	var err error

	logger.Trace.Println("export calls ... ")
	queryStr = "SELECT %s" +
		strings.Repeat(", %s", reflect.TypeOf(ExportCall{}).NumField()-1) +
		" FROM %s"
	queryStr = fmt.Sprintf(
		queryStr,
		"call_id", "task_id", "status", "call_uuid", "file_name", "file_path", "demo_file_path", "description", "duration", "upload_time", "call_time", "staff_id", "staff_name", "extension", "department", "customer_id", "customer_name", "customer_phone", "enterprise", "uploader", "left_silence_time", "right_silence_time", "left_speed", "right_speed", "`type`", "left_channel", "right_channel",
		"`call`",
	)
	if err = SaveToExcel(xlFile, queryStr, "call", delegatee, ExportCall{}); err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)
	xlFile.Write(writer)

	return &buf, err
}

func (c *CallSQLDao) GetCallIDByUUID(delegatee SqlLike, callUUID string) (int64, error) {
	if delegatee == nil {
		delegatee = c.db
	}

	var callID int64
	queryStr := "SELECT `" + fldCallID + "` FROM `" + tblCall + "` WHERE `" + fldCallUUID + "` = ?"
	err := delegatee.QueryRow(queryStr, callUUID).Scan(&callID)
	if err != nil {
		return 0, err
	}

	return callID, nil
}

type ExportCall struct {
	CallID           uint64
	TaskId           uint64
	Status           int
	CallUUID         string
	FileName         string
	FilePath         string
	DemoFilePath     string
	Description      string
	Duration         int
	UploadTime       uint64
	CallTime         uint64
	StaffID          string
	StaffName        string
	Extension        string
	Department       string
	CustomerID       string
	CustomerName     string
	CustomerPhone    string
	Enterprise       string
	Uploader         string
	LeftSilenceTime  float32
	RightSilenceTime float32
	LeftSpeed        float32
	RightSpeed       float32
	Type             int
	LeftChannel      int
	RightChannel     int
}
