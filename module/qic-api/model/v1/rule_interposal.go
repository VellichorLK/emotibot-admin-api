package model

import (
	"fmt"
	"strings"

	"emotibot.com/emotigo/pkg/logger"
)

type InterposalRule struct {
	ID         int64  `json:"_"`
	Name       string `json:"name"`
	Enterprise string `json:"_"`
	Score      int    `json:"score"`
	Seconds    int    `json:"seconds"`
	Times      int    `json:"times"`
	IsDelete   int    `json:"_"`
	CreateTime int64  `json:"_"`
	UpdateTime int64  `json:"_"`
	UUID       string `json:"interposal_id"`
}

type InterposalUpdateSet struct {
	Name    *string `json:"name"`
	Score   *int    `json:"score"`
	Seconds *int    `json:"seconds"`
	Times   *int    `json:"times"`
}

type InterposalRuleDao interface {
	Add(conn SqlLike, r *InterposalRule) (int64, error)
	Get(conn SqlLike, q *GeneralQuery, p *Pagination) ([]*InterposalRule, error)
	Count(conn SqlLike, q *GeneralQuery) (int64, error)
	SoftDelete(conn SqlLike, q *GeneralQuery) (int64, error)
	Update(conn SqlLike, q *GeneralQuery, d *InterposalUpdateSet) (int64, error)
	Copy(conn SqlLike, q *GeneralQuery) (int64, error)
	GetByRuleGroup(conn SqlLike, q *GeneralQuery) ([]*InterposalRule, error)
}

type InterposalRuleSQLDao struct {
}

//Add inserts a new record
func (s *InterposalRuleSQLDao) Add(conn SqlLike, r *InterposalRule) (int64, error) {
	if conn == nil {
		return 0, ErroNoConn
	}
	if conn == nil {
		return 0, ErroNoConn
	}
	if r == nil {
		return 0, ErrNeedRequest
	}
	table := tblInterposalRule
	flds := []string{
		fldID,
		fldName,
		fldEnterprise,
		fldScore,
		fldOverLappedSec,
		fldOverLappedTimes,
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
func (s *InterposalRuleSQLDao) Get(conn SqlLike, q *GeneralQuery, p *Pagination) ([]*InterposalRule, error) {
	if conn == nil {
		return nil, ErroNoConn
	}
	flds := []string{
		fldID,
		fldName,
		fldEnterprise,
		fldScore,
		fldOverLappedSec,
		fldOverLappedTimes,
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
	querySQL := fmt.Sprintf("SELECT %s FROM %s %s %s", strings.Join(flds, ","), tblInterposalRule, condition, offset)
	rows, err := conn.Query(querySQL, params...)
	if err != nil {
		logger.Error.Printf("query failed. %s %+v\n", querySQL, params)
		return nil, err
	}
	defer rows.Close()
	resp := make([]*InterposalRule, 0, 4)
	for rows.Next() {
		var d InterposalRule
		err = rows.Scan(&d.ID, &d.Name, &d.Enterprise,
			&d.Score, &d.Seconds, &d.Times,
			&d.IsDelete, &d.CreateTime, &d.UpdateTime,
			&d.UUID)
		if err != nil {
			logger.Error.Printf("scan failed. %s\n", err)
			return nil, err
		}
		resp = append(resp, &d)
	}
	return resp, nil

}

//Count counts number of the rows under the condition
func (s *InterposalRuleSQLDao) Count(conn SqlLike, q *GeneralQuery) (int64, error) {
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
	return countRows(conn, tblInterposalRule, condition, params)
}

//SoftDelete simply set the is_delete to 1
func (s *InterposalRuleSQLDao) SoftDelete(conn SqlLike, q *GeneralQuery) (int64, error) {
	if conn == nil {
		return 0, ErroNoConn
	}
	if q == nil || (len(q.ID) == 0 && len(q.UUID) == 0) {
		return 0, ErrNeedCondition
	}
	table := tblInterposalRule
	return softDelete(conn, q, table)
}

//Update updates the records
func (s *InterposalRuleSQLDao) Update(conn SqlLike, q *GeneralQuery, d *InterposalUpdateSet) (int64, error) {
	if conn == nil {
		return 0, ErroNoConn
	}
	if q == nil || (len(q.ID) == 0 && len(q.UUID) == 0) {
		return 0, ErrNeedCondition
	}
	flds := []string{
		fldName,
		fldScore,
		fldOverLappedSec,
		fldOverLappedTimes,
	}
	table := tblInterposalRule
	return updateSQL(conn, q, d, table, flds)
}

//Copy copys only one record, only use the first ID in q
func (s *InterposalRuleSQLDao) Copy(conn SqlLike, q *GeneralQuery) (int64, error) {
	if conn == nil {
		return 0, ErroNoConn
	}
	if q == nil || (len(q.ID) == 0) {
		return 0, ErrNeedCondition
	}
	flds := []string{
		fldName,
		fldEnterprise,
		fldScore,
		fldOverLappedSec,
		fldOverLappedTimes,
		fldIsDelete,
		fldCreateTime,
		fldUpdateTime,
		fldUUID,
	}

	for i := range flds {
		flds[i] = "`" + flds[i] + "`"
	}

	fieldsSQL := strings.Join(flds, ",")
	table := tblInterposalRule
	copySQL := fmt.Sprintf("INSERT INTO %s (%s) SELECT %s FROM %s WHERE %s=?", table,
		fieldsSQL, fieldsSQL, table, fldID)

	res, err := conn.Exec(copySQL, q.ID[0])
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

//GetByRuleGroup gets the rule under the conditon of RuleGroup.
//Hence, q is the condition for getting RuleGroup
func (s *InterposalRuleSQLDao) GetByRuleGroup(conn SqlLike, q *GeneralQuery) ([]*InterposalRule, error) {
	if conn == nil {
		return nil, ErroNoConn
	}
	if q == nil || len(q.UUID) == 0 {
		return nil, ErrNeedCondition
	}
	flds := []string{
		fldID,
		fldName,
		fldEnterprise,
		fldScore,
		fldOverLappedSec,
		fldOverLappedTimes,
		fldIsDelete,
		fldCreateTime,
		fldUpdateTime,
		fldUUID,
	}
	for i := range flds {
		flds[i] = "b.`" + flds[i] + "`"
	}

	params := make([]interface{}, 0, len(q.UUID))
	for _, v := range q.UUID {
		params = append(params, v)
	}

	condition := "WHERE a.`" + fldRGUUID + "` IN (?" + strings.Repeat(",?", len(q.UUID)-1) + ")"

	query := fmt.Sprintf("SELECT %s FROM %s AS a INNER JOIN %s AS b %s %s",
		strings.Join(flds, ","),
		tblRelRGInterposal, tblInterposalRule,
		condition)

	return getInterposalRules(conn, query, params)
}

func getInterposalRules(conn SqlLike, querySQL string, params []interface{}) ([]*InterposalRule, error) {
	rows, err := conn.Query(querySQL, params...)
	if err != nil {
		logger.Error.Printf("query failed. %s %+v\n", querySQL, params)
		return nil, err
	}
	defer rows.Close()
	resp := make([]*InterposalRule, 0, 4)
	for rows.Next() {
		var d InterposalRule
		err = rows.Scan(&d.ID, &d.Name, &d.Enterprise,
			&d.Score, &d.Seconds, &d.Times,
			&d.IsDelete, &d.CreateTime, &d.UpdateTime,
			&d.UUID)
		if err != nil {
			logger.Error.Printf("scan failed. %s\n", err)
			return nil, err
		}
		resp = append(resp, &d)
	}
	return resp, nil
}
