package model

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"emotibot.com/emotigo/pkg/logger"
)

type SilenceRule struct {
	ID              int64  `json:"-"`
	Name            string `json:"name"`
	Score           int    `json:"score"`
	Seconds         int    `json:"seconds"`
	Times           int    `json:"times"`
	ExceptionBefore string `json:"-"`
	ExceptionAfter  string `json:"-"`
	Enterprise      string `json:"-"`
	IsDelete        int    `json:"-"`
	CreateTime      int64  `json:"-"`
	UpdateTime      int64  `json:"-"`
	UUID            string `json:"id"`
}

type SilenceUpdateSet struct {
	Name            *string `json:"name"`
	Score           *int    `json:"score"`
	Seconds         *int    `json:"seconds"`
	Times           *int    `json:"times"`
	ExceptionBefore *string `json:"-"`
	ExceptionAfter  *string `json:"-"`
}

type GeneralQuery struct {
	ID         []int64
	UUID       []string
	Enterprise *string
	IsDelete   *int
}

func (g *GeneralQuery) whereSQL() (condition string, bindData []interface{}, err error) {
	flds := []string{
		fldID,
		fldUUID,
		fldEnterprise,
		fldIsDelete,
	}
	return makeAndCondition(g, flds)
}

type SilenceRuleDao interface {
	Add(conn SqlLike, r *SilenceRule) (int64, error)
	Get(conn SqlLike, q *GeneralQuery, p *Pagination) ([]*SilenceRule, error)
	Count(conn SqlLike, q *GeneralQuery) (int64, error)
	SoftDelete(conn SqlLike, q *GeneralQuery) (int64, error)
	Update(conn SqlLike, q *GeneralQuery, d *SilenceUpdateSet) (int64, error)
	Copy(conn SqlLike, q *GeneralQuery) (int64, error)
}

type SilenceRuleSQLDao struct {
}

//Add inserts a new record
func (s *SilenceRuleSQLDao) Add(conn SqlLike, r *SilenceRule) (int64, error) {
	if conn == nil {
		return 0, ErroNoConn
	}
	if r == nil {
		return 0, ErrNeedRequest
	}
	table := tblSilenceRule
	flds := []string{
		fldID,
		fldName,
		fldScore,
		fldSilSecond,
		fldSilTime,
		fldExcptBefore,
		fldExcptAfter,
		fldEnterprise,
		fldIsDelete,
		fldCreateTime,
		fldUpdateTime,
		fldUUID,
	}
	for i := range flds {
		flds[i] = "`" + flds[i] + "`"
	}

	vals := make([]interface{}, 0, len(flds))
	err := extractSimpleStructureValue(&vals, r)
	if err != nil {
		return 0, err
	}
	//remove the ID
	vals = vals[1:]
	flds = flds[1:]

	return insertRow(conn, table, flds, vals)
}

//Get gets the data under the condition
func (s *SilenceRuleSQLDao) Get(conn SqlLike, q *GeneralQuery, p *Pagination) ([]*SilenceRule, error) {
	flds := []string{
		fldID,
		fldName,
		fldScore,
		fldSilSecond,
		fldSilTime,
		fldExcptBefore,
		fldExcptAfter,
		fldEnterprise,
		fldIsDelete,
		fldCreateTime,
		fldUpdateTime,
		fldUUID,
	}
	for i := range flds {
		flds[i] = "`" + flds[i] + "`"
	}

	var condition string
	var params []interface{}
	var err error
	var offset string
	if q != nil {
		condition, params, err = q.whereSQL()
		if err != nil {
			return nil, ErrGenCondition
		}
	}
	if p != nil {
		offset = p.offsetSQL()
	}
	querySQL := fmt.Sprintf("SELECT %s FROM %s %s %s", strings.Join(flds, ","), tblSilenceRule, condition, offset)
	rows, err := conn.Query(querySQL, params...)
	if err != nil {
		logger.Error.Printf("query failed. %s %+v\n", querySQL, params)
		return nil, err
	}
	defer rows.Close()
	resp := make([]*SilenceRule, 0, 4)
	for rows.Next() {
		var d SilenceRule
		err = rows.Scan(&d.ID, &d.Name, &d.Score, &d.Seconds, &d.Times,
			&d.ExceptionBefore, &d.ExceptionAfter, &d.Enterprise, &d.IsDelete, &d.CreateTime, &d.UpdateTime, &d.UUID)
		if err != nil {
			logger.Error.Printf("scan failed. %s\n", err)
			return nil, err
		}
		resp = append(resp, &d)
	}
	return resp, nil

}

//Count counts number of the rows under the condition
func (s *SilenceRuleSQLDao) Count(conn SqlLike, q *GeneralQuery) (int64, error) {
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
	return countRows(conn, tblSilenceRule, condition, params)
}

//SoftDelete simply set the is_delete to 1
func (s *SilenceRuleSQLDao) SoftDelete(conn SqlLike, q *GeneralQuery) (int64, error) {
	if q == nil || (len(q.ID) == 0 && len(q.UUID) == 0) {
		return 0, ErrNeedCondition
	}
	table := tblSilenceRule
	return softDelete(conn, q, table)
}

//Update updates the records
func (s *SilenceRuleSQLDao) Update(conn SqlLike, q *GeneralQuery, d *SilenceUpdateSet) (int64, error) {
	if q == nil || (len(q.ID) == 0 && len(q.UUID) == 0) {
		return 0, ErrNeedCondition
	}
	flds := []string{
		fldName,
		fldScore,
		fldSilSecond,
		fldSilTime,
		fldExcptBefore,
		fldExcptAfter,
	}
	table := tblSilenceRule
	return updateSQL(conn, q, d, table, flds)
}

//Copy copys only one record, only use the first ID in q
func (s *SilenceRuleSQLDao) Copy(conn SqlLike, q *GeneralQuery) (int64, error) {
	if q == nil || (len(q.ID) == 0) {
		return 0, ErrNeedCondition
	}

	flds := []string{
		fldName,
		fldScore,
		fldSilSecond,
		fldSilTime,
		fldExcptBefore,
		fldExcptAfter,
		fldEnterprise,
		fldIsDelete,
		fldCreateTime,
		fldUpdateTime,
		fldUUID,
	}

	for i := range flds {
		flds[i] = "`" + flds[i] + "`"
	}

	fieldsSQL := strings.Join(flds, ",")
	table := tblSilenceRule
	copySQL := fmt.Sprintf("INSERT INTO %s (%s) SELECT %s FROM %s WHERE %s=?", table,
		fieldsSQL, fieldsSQL, table, fldID)

	res, err := conn.Exec(copySQL, q.ID[0])
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

type condition interface {
	whereSQL() (condition string, bindData []interface{}, err error)
}

//must has update_time field in the given table
func updateSQL(conn SqlLike, c condition, d interface{}, table string, flds []string) (int64, error) {
	if conn == nil {
		return 0, ErroNoConn
	}
	if c == nil {
		return 0, ErrNeedCondition
	}
	if d == nil {
		return 0, ErrNeedRequest
	}

	condition, cparams, err := c.whereSQL()
	if err != nil {
		return 0, ErrGenCondition
	}

	setStr, sparams, err := makeSets(d, flds)
	if err != nil {
		logger.Error.Printf("make set query failed. %s\n", err)
		return 0, err
	}

	if setStr == "" {
		return 0, nil
	}

	//sets the update_time to now
	setStr += "," + fldUpdateTime + "=?"
	sparams = append(sparams, time.Now().Unix())

	setSQL := fmt.Sprintf("UPDATE %s %s %s", tblSilenceRule, setStr, condition)
	params := append(sparams, cparams...)
	return execSQL(conn, setSQL, params)
}

//the table must has the field is_delete and update_time
func softDelete(conn SqlLike, c condition, table string) (int64, error) {
	if conn == nil {
		return 0, ErroNoConn
	}
	if c == nil {
		return 0, ErrNeedCondition
	}
	condition, params, err := c.whereSQL()
	if err != nil {
		return 0, ErrGenCondition
	}
	updateSQL := fmt.Sprintf("UPDATE %s SET %s=1,%s=%d %s", table, fldIsDelete,
		fldUpdateTime, time.Now().Unix(), condition)
	return execSQL(conn, updateSQL, params)
}

//this function append the value in the strucutre
func extractSimpleStructureValue(p *[]interface{}, d interface{}) error {
	t := reflect.TypeOf(d)
	vals := reflect.ValueOf(d)
	if reflect.Ptr == t.Kind() {
		t = t.Elem()
		vals = vals.Elem()
	}

	if t.Kind() != reflect.Struct {
		return errors.New("Only suppoert structure input")
	}

	for i := 0; i < t.NumField(); i++ {

		f := t.Field(i)
		vx := vals.Field(i)

		switch f.Type.Kind() {
		case reflect.Slice:
			if vx.Len() > 0 {
				switch f.Type.String() {
				case "[]int64":
					elements := vx.Interface().([]int64)
					for _, ele := range elements {
						*p = append(*p, ele)
					}
				case "[]string":
					elements := vx.Interface().([]string)
					for _, ele := range elements {
						*p = append(*p, ele)
					}
				case "[]uint64":
					elements := vx.Interface().([]uint64)
					for _, ele := range elements {
						*p = append(*p, ele)
					}
				default:
					return fmt.Errorf("unsupported type %s", f.Type.String())
				}
			}
		case reflect.Ptr:
			if !vx.IsNil() {
				switch f.Type.String() {
				case "*string":
					val := vx.Interface().(*string)
					*p = append(*p, *val)
				case "*int":
					val := vx.Interface().(*int)
					*p = append(*p, *val)
				case "*int8":
					val := vx.Interface().(*int8)
					*p = append(*p, *val)
				case "*int64":
					val := vx.Interface().(*int64)
					*p = append(*p, *val)
				case "*uint64":
					val := vx.Interface().(*uint64)
					*p = append(*p, *val)
				default:
					return fmt.Errorf("unsupported type %s", f.Type.String())
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
			*p = append(*p, vx.Interface())
		default:
			return fmt.Errorf("unsupported type %s", f.Type.String())
		}
	}
	return nil

}
