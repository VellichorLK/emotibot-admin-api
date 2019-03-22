package model

import (
	"fmt"
	"strings"

	"github.com/go-sql-driver/mysql"

	"emotibot.com/emotigo/pkg/logger"
)

//GroupCondTypOn Off determine condition apply to
//  * 0 on
//  * 1 off, all files will import
const (
	GroupCondTypOn int8 = iota
	GroupCondTypOff
)

// GroupCondRole is the matching rule for condition
// IT IS NOT IMPLEMENT YET.
// * 0: staff only
// * 1: Customer Only
// * 2: Both Staff And Customer
const (
	GroupCondRoleStaff int8 = iota
	GroupCondRoleCustomer
	GroupCondRoleAny
)

//GroupConditionDao is the db operation to Group Condition Table
type GroupConditionDao struct {
	db DBLike
}

// NewConditionDao create a GroupConditionDao with the given db
func NewConditionDao(db DBLike) *GroupConditionDao {
	return &GroupConditionDao{db: db}
}

type ConditionQuery struct {
	ID         []int64
	GroupID    []int64
	Type       []int8
	Pagination *Pagination
}

func (c *ConditionQuery) whereSQL() (string, []interface{}) {
	builder := NewWhereBuilder(andLogic, "")
	builder.In(fldCondID, int64ToWildCard(c.ID...))
	builder.In(fldCondGroupID, int64ToWildCard(c.GroupID...))
	builder.In(fldCondType, int8ToWildCard(c.Type...))

	return builder.ParseWithWhere()
}

// Condition represent the same table `RuleGroupCondition` as struct GroupCondition,
// but without those json tagging and with its ID & group ID & missing fields.
type Condition struct {
	ID              int64
	GroupID         int64
	Type            int8
	FileName        string
	Deal            int8
	Series          string
	UploadTimeStart int64
	UploadTimeEnd   int64
	StaffID         string
	StaffName       string
	Extension       string
	Department      string
	CustomerID      string
	CustomerName    string
	CustomerPhone   string
	CallStart       int64
	CallEnd         int64
	LeftChannel     int8
	RightChannel    int8
}

var conditionCols = []string{
	fldCondID, fldCondGroupID, fldCondType,
	fldCondFileName, fldCondDeal, fldCondSeries,
	fldCondUploadTimeStart, fldCondUploadTimeEnd, fldCondStaffID,
	fldCondStaffName, fldCondExtension, fldCondDepartment,
	fldCondCustomerID, fldCondCustomerName, fldCondCustomerPhone,
	fldCondCallStart, fldCondCallEnd, fldCondLeftChanRole,
	fldCondRightChanRole,
}

func (g *GroupConditionDao) Conditions(delegatee SqlLike, query ConditionQuery) ([]Condition, error) {
	if delegatee == nil {
		delegatee = g.db.Conn()
	}

	var (
		offsetSQL string
	)
	if query.Pagination != nil {
		offsetSQL = query.Pagination.offsetSQL()
	}
	wherePart, data := query.whereSQL()
	rawsql := fmt.Sprintf("SELECT `%s`FROM `%s`"+
		" %s %s ORDER BY `%s` ASC",
		strings.Join(conditionCols, "`, `"), tblRGC,
		wherePart, offsetSQL, fldCondID,
	)
	rows, err := delegatee.Query(rawsql, data...)
	if err != nil {
		mysqlErr, ok := err.(*mysql.MySQLError)
		if ok && mysqlErr.Number == 1064 {
			logger.Error.Println("raw error sql: ", rawsql)
		}
		return nil, fmt.Errorf("query sql failed, %v", err)
	}
	defer rows.Close()
	var scanned = []Condition{}
	for rows.Next() {
		var cond Condition
		rows.Scan(
			&cond.ID, &cond.GroupID, &cond.Type,
			&cond.FileName, &cond.Deal, &cond.Series,
			&cond.UploadTimeStart, &cond.UploadTimeEnd, &cond.StaffID,
			&cond.StaffName, &cond.Extension, &cond.Department,
			&cond.CustomerID, &cond.CustomerName, &cond.CustomerPhone,
			&cond.CallStart, &cond.CallEnd, &cond.LeftChannel,
			&cond.RightChannel,
		)
		scanned = append(scanned, cond)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("sql scan error, %v", err)
	}
	return scanned, nil
}

func (g *GroupConditionDao) NewCondition(delegatee SqlLike, cond Condition) (Condition, error) {
	if delegatee == nil {
		delegatee = g.db.Conn()
	}
	condCols := []string{
		fldCondGroupID, fldCondType, fldCondFileName,
		fldCondDeal, fldCondSeries, fldCondUploadTimeStart,
		fldCondUploadTimeEnd, fldCondStaffID, fldCondStaffName,
		fldCondExtension, fldCondDepartment, fldCondCustomerID,
		fldCondCustomerName, fldCondCustomerPhone, fldCondCallStart,
		fldCondCallEnd, fldCondLeftChanRole, fldCondRightChanRole,
	}
	rawsql := fmt.Sprintf("INSERT INTO `%s` (`%s`) VALUE(?%s)",
		tblRGC, strings.Join(condCols, "`,`"), strings.Repeat(", ?", len(condCols)-1))
	result, err := delegatee.Exec(rawsql,
		cond.GroupID, cond.Type, cond.FileName,
		cond.Deal, cond.Series, cond.UploadTimeStart,
		cond.UploadTimeEnd, cond.StaffID, cond.StaffName,
		cond.Extension, cond.Department, cond.CustomerID,
		cond.CustomerName, cond.CustomerPhone, cond.CallStart,
		cond.CallEnd, cond.LeftChannel, cond.RightChannel,
	)
	if err != nil {
		return Condition{}, fmt.Errorf("sql execute failed, %v", err)
	}
	cond.ID, err = result.LastInsertId()
	if err != nil {
		return cond, ErrAutoIDDisabled
	}
	return cond, nil
}
