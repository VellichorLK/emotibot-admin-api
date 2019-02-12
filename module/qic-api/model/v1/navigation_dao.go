package model

import (
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"time"

	"emotibot.com/emotigo/pkg/logger"
)

var (
	ErrGenCondition = errors.New("generate conidtion failed")
)

//NavigationDao is the interface of navigation dao
type NavigationDao interface {
	NewFlow(conn SqlLike, p *NavFlow) (int64, error)
	GetFlows(conn SqlLike, q *NavQuery, l *NavLimit) ([]*NavFlow, error)
	InsertRelation(conn SqlLike, parent int64, child int64) (int64, error)
	DeleteRelation(conn SqlLike, parent int64) (int64, error)
	SoftDeleteFlows(conn SqlLike, q *NavQuery) (int64, error)
	DeleteFlows(conn SqlLike, q *NavQuery) (int64, error)
	CountFlows(conn SqlLike, q *NavQuery) (int64, error)
	CountNodes(conn SqlLike, navs []int64) (map[int64]int64, error)
	GetNodeID(conn SqlLike, nav int64) ([]int64, error)
	UpdateFlows(conn SqlLike, q *NavQuery, d *NavFlowUpdate) (int64, error)
}

//NavigationSQLDao implements the function to access the navigation db
type NavigationSQLDao struct {
}

//NavFlow is the structure of the navigation informtation
type NavFlow struct {
	ID           int64
	UUID         string
	Name         string
	IgnoreIntent int
	IntentName   string
	IntentLinkID int64
	Enterprise   string
	CreateTime   int64
	UpdateTime   int64
}

type NavFlowUpdate struct {
	Name         *string
	IgnoreIntent *int
	IntentName   *string
	IntentLinkID *int64
	UpdateTime   *int64
}

//NavQuery is query condition used in navigation api
type NavQuery struct {
	ID         []int64
	UUID       []string
	Enterprise *string
	IsDelete   *int
}

//NavLimit is the limitation of the returned rows
type NavLimit struct {
	Page  int
	Limit int
}

func (n *NavQuery) whereSQL() (condition string, bindData []interface{}, err error) {
	flds := []string{
		fldID,
		fldUUID,
		fldEnterprise,
		fldIsDelete,
	}
	return makeAndCondition(n, flds)
}

//NewFlow creates the new flow
func (n *NavigationSQLDao) NewFlow(conn SqlLike, p *NavFlow) (int64, error) {
	if conn == nil {
		return 0, ErroNoConn
	}
	if p == nil {
		return 0, ErrNeedRequest
	}

	flds := []string{
		fldName,
		fldUUID,
		fldIgoreIntent,
		fldIntenName,
		fldIntentLink,
		fldEnterprise,
		fldCreateTime,
		fldUpdateTime,
	}

	vals := make([]interface{}, 0, 8)
	vals = append(vals, p.Name, p.UUID, p.IgnoreIntent, p.IntentName, p.IntentLinkID, p.Enterprise, p.CreateTime, p.CreateTime)
	return insertRow(conn, tblNavigation, flds, vals)
}

//GetFlows gets the flow according to the query condition
func (n *NavigationSQLDao) GetFlows(conn SqlLike, q *NavQuery, l *NavLimit) ([]*NavFlow, error) {
	if conn == nil {
		return nil, ErroNoConn
	}
	if q == nil {
		return nil, ErrNeedCondition
	}
	condition, params, err := q.whereSQL()
	if err != nil {
		return nil, ErrGenCondition
	}

	var limitStr string
	if l != nil {
		if l.Page > 0 && l.Limit > 0 {
			limitStr = fmt.Sprintf("LIMIT %d OFFSET %d", l.Limit, (l.Page-1)*l.Limit)
		}
	}

	flds := []string{
		fldID,
		fldName,
		fldUUID,
		fldIgoreIntent,
		fldIntenName,
		fldIntentLink,
		fldEnterprise,
		fldCreateTime,
		fldUpdateTime,
	}

	querySQL := fmt.Sprintf("SELECT %s FROM %s %s %s", strings.Join(flds, ","), tblNavigation, condition, limitStr)
	rows, err := conn.Query(querySQL, params...)
	if err != nil {
		logger.Error.Printf("query failed. %s %+v\n", querySQL, params)
		return nil, err
	}
	defer rows.Close()
	flows := make([]*NavFlow, 0, 18)
	for rows.Next() {
		var n NavFlow
		err = rows.Scan(&n.ID, &n.Name, &n.UUID, &n.IgnoreIntent, &n.IntentName,
			&n.IntentLinkID, &n.Enterprise, &n.CreateTime, &n.UpdateTime)
		if err != nil {
			logger.Error.Printf("scan failed. %s\n", err)
			return nil, err
		}
		flows = append(flows, &n)
	}

	return flows, nil
}

//InsertRelation inserts a relation into Relation_Nav_SenGrp table
func (n *NavigationSQLDao) InsertRelation(conn SqlLike, parent int64, child int64) (int64, error) {
	if conn == nil {
		return 0, ErroNoConn
	}
	return insertRow(conn, tblRelNavSenGrp, []string{fldNavID, fldSenGrpID}, []interface{}{parent, child})
}

//SoftDeleteFlows updates the is_delete to 1
func (n *NavigationSQLDao) SoftDeleteFlows(conn SqlLike, q *NavQuery) (int64, error) {
	if conn == nil {
		return 0, ErroNoConn
	}
	if q == nil {
		return 0, ErrNeedCondition
	}
	condition, params, err := q.whereSQL()
	if err != nil {
		return 0, ErrGenCondition
	}
	updateSQL := fmt.Sprintf("UPDATE %s SET %s=1,%s=%d %s", tblNavigation, fldIsDelete,
		fldUpdateTime, time.Now().Unix(), condition)
	return execSQL(conn, updateSQL, params)
}

//DeleteFlows deletes the flow
func (n *NavigationSQLDao) DeleteFlows(conn SqlLike, q *NavQuery) (int64, error) {
	if conn == nil {
		return 0, ErroNoConn
	}
	if q == nil {
		return 0, ErrNeedCondition
	}
	condition, params, err := q.whereSQL()
	if err != nil {
		return 0, ErrGenCondition
	}
	deleteSQL := fmt.Sprintf("DELETE FROM %s %s", tblNavigation, condition)
	return execSQL(conn, deleteSQL, params)
}

//DeleteRelation deletes the relation
func (n *NavigationSQLDao) DeleteRelation(conn SqlLike, parent int64) (int64, error) {
	if conn == nil {
		return 0, ErroNoConn
	}
	deleteSQL := fmt.Sprintf("DELETE FROM %s WHERE %s=?", tblRelNavSenGrp, fldNavID)
	return execSQL(conn, deleteSQL, []interface{}{parent})
}

//CountFlows counts number of rows that meet the condition
func (n *NavigationSQLDao) CountFlows(conn SqlLike, q *NavQuery) (int64, error) {
	if conn == nil {
		return 0, ErroNoConn
	}
	var condition string
	var params []interface{}
	var err error
	if q != nil {
		condition, params, err = q.whereSQL()
		if err != nil {
			return 0, ErrGenCondition
		}
	}
	return countRows(conn, tblNavigation, condition, params)
}

//CountNodes counts nodes
func (n *NavigationSQLDao) CountNodes(conn SqlLike, navs []int64) (map[int64]int64, error) {
	if conn == nil {
		return nil, ErroNoConn
	}
	var condition string

	numOfNavs := len(navs)
	params := make([]interface{}, 0, numOfNavs)
	for _, v := range navs {
		params = append(params, v)
	}
	if numOfNavs > 0 {
		condition = "WHERE a.`" + fldNavID + "` IN (?" + strings.Repeat(",?", numOfNavs-1) + ")"
	}

	if condition != "" {
		condition += " AND"
	} else {
		condition += "WHERE"
	}
	condition += " c.`" + fldIsDelete + "`=0"

	countSQL := fmt.Sprintf("select a.`%s`,COUNT(*) FROM %s AS a INNER JOIN %s AS b ON a.`%s`=b.`%s` INNER JOIN %s AS c ON b.`%s`=c.`%s` %s  GROUP BY a.`%s`",
		fldNavID, tblRelNavSenGrp, tblSetnenceGroup,
		fldSenGrpID, fldID,
		tblSetnenceGroup, fldUUID, fldUUID,
		condition, fldNavID)

	rows, err := conn.Query(countSQL, params...)
	if err != nil {
		logger.Error.Printf("query sql failed. %s, %+v\n", countSQL, params)
		return nil, err
	}
	defer rows.Close()
	resp := make(map[int64]int64)
	var nav, count int64
	for rows.Next() {
		err = rows.Scan(&nav, &count)
		if err != nil {
			logger.Error.Printf("scan failed. %s\n", err)
			break
		}
		resp[nav] = count
	}
	return resp, err
}

//UpdateFlows updates the flow
func (n *NavigationSQLDao) UpdateFlows(conn SqlLike, q *NavQuery, d *NavFlowUpdate) (int64, error) {
	if conn == nil {
		return 0, ErroNoConn
	}
	if q == nil {
		return 0, ErrNeedCondition
	}
	if d == nil {
		return 0, ErrNeedRequest
	}

	condition, cparams, err := q.whereSQL()
	if err != nil {
		return 0, ErrGenCondition
	}

	if d.UpdateTime == nil {
		now := time.Now().Unix()
		d.UpdateTime = &now
	}

	flds := []string{
		fldName,
		fldIgoreIntent,
		fldIntenName,
		fldIntentLink,
		fldUpdateTime,
	}

	setStr, sparams, err := makeSets(d, flds)
	if err != nil {
		logger.Error.Printf("make set query failed. %s\n", err)
		return 0, err
	}
	setSQL := fmt.Sprintf("UPDATE %s %s %s", tblNavigation, setStr, condition)
	params := append(sparams, cparams...)
	return execSQL(conn, setSQL, params)
}

//GetNodeID gets the id in sentenceGroup
func (n *NavigationSQLDao) GetNodeID(conn SqlLike, nav int64) ([]int64, error) {
	if conn == nil {
		return nil, ErroNoConn
	}
	querySQL := fmt.Sprintf("SELECT `%s` FROM %s WHERE `%s`=?", fldSenGrpID, tblRelNavSenGrp, fldNavID)
	rows, err := conn.Query(querySQL, nav)
	if err != nil {
		logger.Error.Printf("query failed. %s,%d. %s\n", querySQL, nav, err)
		return nil, err
	}
	defer rows.Close()

	senGrps := make([]int64, 0, 4)
	var senGrp int64
	for rows.Next() {
		err = rows.Scan(&senGrp)
		if err != nil {
			logger.Error.Printf("scan failed. %s\n", err)
			return nil, err
		}
		senGrps = append(senGrps, senGrp)
	}
	return senGrps, nil

}

//this function doesn't check input, only used internally.
func insertRow(conn SqlLike, table string, flds []string, vals []interface{}) (int64, error) {
	insertSQL := fmt.Sprintf("INSERT INTO %s (%s) VALUES (?%s)",
		table, strings.Join(flds, ","), strings.Repeat(",?", len(vals)-1))
	res, err := conn.Exec(insertSQL, vals...)
	if err != nil {
		pc, _, line, _ := runtime.Caller(1)
		details := runtime.FuncForPC(pc)
		logger.Error.Printf("[%s]%d: execute failed. sql: %s, params: %+v\n", details.Name(), line, insertSQL, vals)
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		pc, _, line, _ := runtime.Caller(1)
		details := runtime.FuncForPC(pc)
		logger.Error.Printf("[%s]%d: get the last insert id failed. sql: %s, params: %+v\n", details.Name(), line, insertSQL, vals)
		return 0, err
	}
	return id, nil
}

//execSQL executes the update and delete sql
func execSQL(conn SqlLike, exeSQL string, params []interface{}) (int64, error) {
	res, err := conn.Exec(exeSQL, params...)
	if err != nil {
		pc, _, line, _ := runtime.Caller(1)
		details := runtime.FuncForPC(pc)
		logger.Error.Printf("[%s]%d: execute failed. sql: %s, params: %+v\n", details.Name(), line, exeSQL, params)
		return 0, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		pc, _, line, _ := runtime.Caller(1)
		details := runtime.FuncForPC(pc)
		logger.Error.Printf("[%s]%d: get the affected rows failed. sql: %s, params: %+v\n", details.Name(), line, exeSQL, params)
		return 0, err
	}
	return affected, nil
}

func countRows(conn SqlLike, table string, condition string, params []interface{}) (int64, error) {
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM %s %s", table, condition)
	var count int64
	err := conn.QueryRow(countSQL, params...).Scan(&count)
	if err != nil {
		pc, _, line, _ := runtime.Caller(1)
		details := runtime.FuncForPC(pc)
		logger.Error.Printf("[%s]%d: count rows failed. sql: %s, params: %+v\n", details.Name(), line, countSQL, params)
		return 0, err
	}
	return count, nil
}

//only used internally. make the condition based on input strcut d
//if the attribute is slice, conidtion uses in, others uses =
func makeAndCondition(d interface{}, flds []string) (condition string, bindData []interface{}, err error) {
	var states []string
	states, bindData, err = makePreprare(d, flds)
	if err != nil {
		return "", nil, err
	}
	if len(states) > 0 {
		condition = "WHERE " + strings.Join(states, " AND ")
	}
	return
}

func makeSets(d interface{}, flds []string) (sets string, bindData []interface{}, err error) {
	var states []string
	states, bindData, err = makePreprare(d, flds)
	if err != nil {
		return "", nil, err
	}
	if len(states) > 0 {
		sets = "SET " + strings.Join(states, ",")
	}
	return
}

func makePreprare(d interface{}, flds []string) (states []string, bindData []interface{}, err error) {
	t := reflect.TypeOf(d)
	vals := reflect.ValueOf(d)
	if reflect.Ptr == t.Kind() {
		t = t.Elem()
		vals = vals.Elem()
	}

	if t.Kind() != reflect.Struct {
		err = errors.New("Only suppoert structure input")
		return
	}

	numOfFlds := len(flds)
	for i := 0; i < t.NumField(); i++ {
		if i >= numOfFlds {
			err = errors.New("# struct field > # input field")
			return
		}
		f := t.Field(i)
		vx := vals.Field(i)

		switch f.Type.Kind() {
		case reflect.Slice:
			if vx.Len() > 0 {
				p := flds[i] + " IN (?" + strings.Repeat(",?", vx.Len()-1) + ")"
				states = append(states, p)
			}
			switch f.Type.String() {
			case "[]int64":
				elements := vx.Interface().([]int64)
				for _, ele := range elements {
					bindData = append(bindData, ele)
				}
			case "[]string":
				elements := vx.Interface().([]string)
				for _, ele := range elements {
					bindData = append(bindData, ele)
				}
			case "[]uint64":
				elements := vx.Interface().([]uint64)
				for _, ele := range elements {
					bindData = append(bindData, ele)
				}
			default:
				err = fmt.Errorf("unsupported type %s", f.Type.String())
				return
			}
		case reflect.Ptr:
			if !vx.IsNil() {

				p := flds[i] + "=?"
				states = append(states, p)
				switch f.Type.String() {
				case "*string":
					val := vx.Interface().(*string)
					bindData = append(bindData, *val)
				case "*int":
					val := vx.Interface().(*int)
					bindData = append(bindData, *val)
				case "*int8":
					val := vx.Interface().(*int8)
					bindData = append(bindData, *val)
				case "*int64":
					val := vx.Interface().(*int64)
					bindData = append(bindData, *val)
				case "*uint64":
					val := vx.Interface().(*uint64)
					bindData = append(bindData, *val)
				default:
					err = fmt.Errorf("unsupported type %s", f.Type.String())
					return
				}
			}
		case reflect.Int:
			fallthrough
		case reflect.Int32:
			fallthrough
		case reflect.Int64:
			fallthrough
		case reflect.Int8:
			fallthrough
		case reflect.Uint:
			fallthrough
		case reflect.Uint8:
			fallthrough
		case reflect.Uint16:
			fallthrough
		case reflect.Uint32:
			fallthrough
		case reflect.Uint64:
			fallthrough
		case reflect.String:
			p := flds[i] + "=?"
			states = append(states, p)
			bindData = append(bindData, vx.Interface())
		default:
			err = fmt.Errorf("unsupported type %s", f.Type.String())
			return
		}
	}
	return
}
