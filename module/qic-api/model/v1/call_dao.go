package model

import (
	"database/sql"
	"fmt"
	"strings"

	"emotibot.com/emotigo/pkg/logger"
)

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
	CallTimeStart *int64
	CallTimeEnd   *int64
	StaffID       []string
	EnterpriseID  *string
	CustomerPhone *string
	DealStatus    *int8
	Ext           *string
	Department    *string
	Paging        *Pagination
}

func (c *CallQuery) whereSQL() (string, []interface{}) {
	var (
		rawSQL     string
		bindData   []interface{}
		conditions []string
	)
	if len(c.ID) > 0 {
		cond := fmt.Sprintf("`%s` IN (? %s)", fldCallID, strings.Repeat(",? ", len(c.ID)-1))
		conditions = append(conditions, cond)
		for _, id := range c.ID {
			bindData = append(bindData, id)
		}
	}
	if len(c.UUID) > 0 {
		cond := fmt.Sprintf("`%s` IN (? %s)", fldCallUUID, strings.Repeat(",? ", len(c.UUID)-1))
		conditions = append(conditions, cond)
		for _, uuid := range c.UUID {
			bindData = append(bindData, uuid)
		}
	}
	if len(c.Status) > 0 {
		cond := fmt.Sprintf("`%s` IN (? %s)", fldCallStatus, strings.Repeat(",? ", len(c.Status)-1))
		conditions = append(conditions, cond)
		for _, s := range c.Status {
			bindData = append(bindData, s)
		}
	}
	if c.CallTimeStart != nil {
		cond := fmt.Sprintf("`%s` >= ?", fldCallCallTime)
		conditions = append(conditions, cond)
		bindData = append(bindData, *c.CallTimeStart)
	}
	if c.CallTimeEnd != nil {
		cond := fmt.Sprintf("`%s` <= ?", fldCallCallTime)
		conditions = append(conditions, cond)
		bindData = append(bindData, *c.CallTimeEnd)
	}
	if len(c.StaffID) > 0 {
		cond := fmt.Sprintf("`%s` IN (? %s)", fldCallStaffID, strings.Repeat(",? ", len(c.StaffID)-1))
		conditions = append(conditions, cond)
		for _, s := range c.StaffID {
			bindData = append(bindData, s)
		}
	}
	if c.EnterpriseID != nil {
		cond := fmt.Sprintf("`%s`=?", fldCallEnterprise)
		conditions = append(conditions, cond)
		bindData = append(bindData, *c.EnterpriseID)
	}
	if c.CustomerPhone != nil {
		cond := fmt.Sprintf("`%s`=?", fldCallCustomerPhone)
		conditions = append(conditions, cond)
		bindData = append(bindData, *c.CustomerPhone)
	}
	// deal status need to query the task, we will implement this later
	// if c.DealStatus != nil {
	// 	cond := fmt.Sprintf("`%s`=?", f)
	// 	conditions = append(conditions, cond)
	// 	bindData = append(bindData, c.CustomerPhone)
	// }

	if c.Ext != nil {
		cond := fmt.Sprintf("`%s`=?", fldCallExt)
		conditions = append(conditions, cond)
		bindData = append(bindData, *c.Ext)
	}
	if c.Department != nil {
		cond := fmt.Sprintf("`%s`=?", fldCallDepartment)
		conditions = append(conditions, cond)
		bindData = append(bindData, *c.Department)
	}
	if len(conditions) > 0 {
		rawSQL = " WHERE " + strings.Join(conditions, " AND ")
	}
	return rawSQL, bindData
}

// Call represent the call table of the QISYS database.
// Any pointer field is nullable in the schema.Call
// Ext(分機號碼) is the receiver(staff) extension number.
type Call struct {
	ID                 int64
	UUID               string
	FileName           *string
	FilePath           *string
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
	Status             int8
	DemoFilePath       *string
	TaskID             int64
}

// the type of the call is created, different type indicate different incoming source of call.
// 	- 0: whole audio upload api (Default)
// 	- 1: realtime audio upload api
const (
	CallTypeWholeFile int8 = 0
	CallTypeRealTime  int8 = 1
)

// Channel type of the call
// 	- 0: default,
//	- 1: staff(客服)
//	- 2: customer(客戶)
const (
	CallChanDefault int8 = iota
	CallChanStaff
	CallChanCustomer
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
		fldCallTaskID, fldCallDemoFilePath,
	}
	wheresql, data := query.whereSQL()
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
		)
		err := rows.Scan(&c.ID, &c.UUID, &fileName,
			&filePath, &description, &c.DurationMillSecond,
			&c.UploadUnixTime, &c.CallUnixTime, &c.StaffID, &c.StaffName,
			&c.Ext, &c.Department, &c.CustomerID,
			&c.CustomerName, &c.CustomerPhone, &c.EnterpriseID,
			&c.UploadUser, &leftSTime, &rightSTime,
			&lSpeed, &rSpeed, &c.Type,
			&c.LeftChanRole, &c.RightChanRole, &c.Status,
			&c.TaskID, &demoFp,
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
		fldCallLeftChan, fldCallRightChan, fldCallTaskID,
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
			c.LeftChanRole, c.RightChanRole, c.TaskID)
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

func (c *CallSQLDao) SetRuleGroupRelations(delegatee SqlLike, call Call, rulegroups []uint64) ([]int64, error) {
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
		result, err := stmt.Exec(call.ID, r)
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
	wheresql, data := query.whereSQL()
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
		fldCallTaskID, fldCallLeftSilenceTime, fldCallRightSilenceTime,
		fldCallLeftSpeed, fldCallRightSpeed,
	}

	data := []interface{}{
		c.UUID, c.DurationMillSecond, c.UploadUnixTime,
		*c.FileName, *c.FilePath, *c.Description,
		c.CallUnixTime, c.StaffID, c.StaffName,
		c.Ext, c.Department, c.CustomerID,
		c.CustomerName, c.CustomerPhone, c.EnterpriseID,
		c.UploadUser, c.Type, c.LeftChanRole,
		c.RightChanRole, c.Status, c.DemoFilePath,
		c.TaskID, c.LeftSilenceTime, c.RightSilenceTime,
		c.LeftSpeed, c.RightSpeed,
	}

	for _, colName := range updateCols {
		p := fmt.Sprintf("`%s` = ?", colName)
		parts = append(parts, p)
	}

	rawsql := strings.Join(parts, " , ")
	return rawsql, data

}
