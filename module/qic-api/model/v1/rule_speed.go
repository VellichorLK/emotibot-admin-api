package model

import (
	"fmt"
	"strings"

	"emotibot.com/emotigo/pkg/logger"
)

type SpeedRule struct {
	ID             int64  `json:"-"`
	Name           string `json:"name"`
	Score          int    `json:"score"`
	Min            int    `json:"min"`
	Max            int    `json:"max"`
	ExceptionUnder string `json:"-"`
	ExceptionOver  string `json:"-"`
	Enterprise     string `json:"-"`
	IsDelete       int    `json:"-"`
	CreateTime     int64  `json:"-"`
	UpdateTime     int64  `json:"-"`
	UUID           string `json:"speed_id"`
}

type SpeedUpdateSet struct {
	Name           *string `json:"name"`
	Score          *int    `json:"score"`
	Min            *int    `json:"min"`
	Max            *int    `json:"max"`
	ExceptionUnder *string `json:"_"`
	ExceptionOver  *string `json:"_"`
}

type SpeedRuleDao interface {
	Add(conn SqlLike, r *SpeedRule) (int64, error)
	Get(conn SqlLike, q *GeneralQuery, p *Pagination) ([]*SpeedRule, error)
	Count(conn SqlLike, q *GeneralQuery) (int64, error)
	SoftDelete(conn SqlLike, q *GeneralQuery) (int64, error)
	Update(conn SqlLike, q *GeneralQuery, d *SpeedUpdateSet) (int64, error)
	Copy(conn SqlLike, q *GeneralQuery) (int64, error)
}

type SpeedRuleSQLDao struct {
}

//Add inserts a new record
func (s *SpeedRuleSQLDao) Add(conn SqlLike, r *SpeedRule) (int64, error) {
	if conn == nil {
		return 0, ErroNoConn
	}
	if r == nil {
		return 0, ErrNeedRequest
	}
	table := tblSpeedRule
	flds := []string{
		fldID,
		fldName,
		fldScore,
		fldMin,
		fldMax,
		fldExcptUnder,
		fldExcptOver,
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
func (s *SpeedRuleSQLDao) Get(conn SqlLike, q *GeneralQuery, p *Pagination) ([]*SpeedRule, error) {
	flds := []string{
		fldID,
		fldName,
		fldScore,
		fldMin,
		fldMax,
		fldExcptUnder,
		fldExcptOver,
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
	querySQL := fmt.Sprintf("SELECT %s FROM %s %s %s", strings.Join(flds, ","), tblSpeedRule, condition, offset)
	rows, err := conn.Query(querySQL, params...)
	if err != nil {
		logger.Error.Printf("query failed. %s %+v\n", querySQL, params)
		return nil, err
	}
	defer rows.Close()
	resp := make([]*SpeedRule, 0, 4)
	for rows.Next() {
		var d SpeedRule
		err = rows.Scan(&d.ID, &d.Name, &d.Score, &d.Min, &d.Max,
			&d.ExceptionUnder, &d.ExceptionOver, &d.Enterprise, &d.IsDelete, &d.CreateTime, &d.UpdateTime, &d.UUID)
		if err != nil {
			logger.Error.Printf("scan failed. %s\n", err)
			return nil, err
		}
		resp = append(resp, &d)
	}
	return resp, nil

}

//Count counts number of the rows under the condition
func (s *SpeedRuleSQLDao) Count(conn SqlLike, q *GeneralQuery) (int64, error) {
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
	return countRows(conn, tblSpeedRule, condition, params)
}

//SoftDelete simply set the is_delete to 1
func (s *SpeedRuleSQLDao) SoftDelete(conn SqlLike, q *GeneralQuery) (int64, error) {
	if q == nil || (len(q.ID) == 0 && len(q.UUID) == 0) {
		return 0, ErrNeedCondition
	}
	table := tblSpeedRule
	return softDelete(conn, q, table)
}

//Update updates the records
func (s *SpeedRuleSQLDao) Update(conn SqlLike, q *GeneralQuery, d *SpeedUpdateSet) (int64, error) {
	if q == nil || (len(q.ID) == 0 && len(q.UUID) == 0) {
		return 0, ErrNeedCondition
	}
	flds := []string{
		fldName,
		fldScore,
		fldMin,
		fldMax,
		fldExcptUnder,
		fldExcptOver,
	}
	table := tblSpeedRule
	return updateSQL(conn, q, d, table, flds)
}

//Copy copys only one record, only use the first ID in q
func (s *SpeedRuleSQLDao) Copy(conn SqlLike, q *GeneralQuery) (int64, error) {
	if q == nil || (len(q.ID) == 0) {
		return 0, ErrNeedCondition
	}

	flds := []string{
		fldName,
		fldScore,
		fldMin,
		fldMax,
		fldExcptUnder,
		fldExcptOver,
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
	table := tblSpeedRule
	copySQL := fmt.Sprintf("INSERT INTO %s (%s) SELECT %s FROM %s WHERE %s=?", table,
		fieldsSQL, fieldsSQL, table, fldID)

	res, err := conn.Exec(copySQL, q.ID[0])
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}
