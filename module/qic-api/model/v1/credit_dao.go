package model

import (
	"fmt"
	"strings"

	"emotibot.com/emotigo/pkg/logger"
)

//CreditDao is the interface used to access the prediction result
type CreditDao interface {
	InsertCredit(conn SqlLike, c *SimpleCredit) (int64, error)
	InsertCredits(SqlLike, []SimpleCredit) error
	Update(conn SqlLike, q *GeneralQuery, d *UpdateCreditSet) (int64, error)
	InsertSegmentMatch(conn SqlLike, s *SegmentMatch) (int64, error)
	GetCallCredit(conn SqlLike, q *CreditQuery) ([]*SimpleCredit, error)
	GetSegmentMatch(conn SqlLike, q *SegmentPredictQuery) ([]*SegmentMatch, error)
}

type UpdateCreditSet struct {
	Score *int
}

//SegmentPredictQuery is the condition used to query the SegmentPredict
type SegmentPredictQuery struct {
	Segs []uint64
	Whos *int
}

func (s *SegmentPredictQuery) whereSQL() (condition string, bindData []interface{}, err error) {
	flds := []string{
		fldSegID,
		fldWhos,
	}
	return makeAndCondition(s, flds)
}

//CreditQuery is the condition used to query the CUPredictReuslt
type CreditQuery struct {
	Calls []uint64
	//	Whos  int
}

func (c *CreditQuery) whereSQL() (condition string, bindData []interface{}, err error) {
	flds := []string{
		fldCallID,
		//	fldWhos,
	}
	return makeAndCondition(c, flds)
}

//CreditSQLDao implements the credit dao
type CreditSQLDao struct {
}

//SimpleCredit is the struture used to store result
type SimpleCredit struct {
	ID         uint64
	CallID     uint64
	Type       int
	ParentID   uint64
	OrgID      uint64
	Valid      int
	Revise     int
	Score      int
	CreateTime int64
	UpdateTime int64
	Whos       int
	Comment    string
}

//SegmentMatch is the structure used to insert the matched segment
type SegmentMatch struct {
	ID          uint64 `json:"id"`
	SegID       uint64 `json:"segment_id"`
	TagID       uint64 `json:"-"`
	Score       int    `json:"score"`
	Match       string `json:"match"`
	MatchedText string `json:"match_text"`
	CreateTime  int64  `json:"-"`
	UpdateTime  int64  `json:"-"`
	Whos        int    `json:"-"`
}

//InsertCredit inserts the muliple credit
func (c *CreditSQLDao) InsertCredit(conn SqlLike, credit *SimpleCredit) (int64, error) {
	if conn == nil {
		return 0, ErroNoConn
	}
	if credit == nil {
		return 0, nil
	}

	table := tblPredictResult
	insertFlds := []string{
		fldCallID,
		fldType,
		fldParentID,
		fldOrgID,
		fldValid,
		fldRevise,
		fldScore,
		fldCreateTime,
		fldUpdateTime,
		fldWhos,
	}

	var params []interface{}

	params = append(params, credit.CallID, credit.Type, credit.ParentID,
		credit.OrgID, credit.Valid, credit.Revise, credit.Score, credit.CreateTime,
		credit.CreateTime, credit.Whos)

	return insertRecord(conn, table, insertFlds, params)
}

// InsertCredits inserts batch of credits
func (c *CreditSQLDao) InsertCredits(conn SqlLike, credits []SimpleCredit) (err error) {
	if conn == nil {
		return ErroNoConn
	}

	if len(credits) == 0 {
		return
	}

	table := tblPredictResult
	insertFlds := []string{
		fldCallID,
		fldType,
		fldParentID,
		fldOrgID,
		fldValid,
		fldRevise,
		fldScore,
		fldCreateTime,
		fldUpdateTime,
		fldWhos,
		fldDescription,
	}

	var params []interface{}
	paramStrTemplate := "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?),"
	paramStr := ""
	for _, credit := range credits {
		params = append(params, credit.CallID, credit.Type, credit.ParentID,
			credit.OrgID, credit.Valid, credit.Revise, credit.Score, credit.CreateTime,
			credit.CreateTime, credit.Whos, credit.Comment)
		paramStr = fmt.Sprintf("%s %s", paramStr, paramStrTemplate)
	}
	paramStr = paramStr[:len(paramStr)-1]
	for i := 0; i < len(insertFlds); i++ {
		insertFlds[i] = "`" + insertFlds[i] + "`"
	}

	insertStr := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES %s",
		table,
		strings.Join(insertFlds, ","),
		paramStr,
	)

	_, err = conn.Exec(insertStr, params...)
	if err != nil {
		logger.Error.Printf("insert credits failed. %s.\n %s, %+v\n", err, insertStr, params)
	}
	return err
}

//Update updates the records
func (c *CreditSQLDao) Update(conn SqlLike, q *GeneralQuery, d *UpdateCreditSet) (int64, error) {
	if conn == nil {
		return 0, ErroNoConn
	}
	if q == nil || (len(q.ID) == 0 && len(q.UUID) == 0) {
		return 0, ErrNeedCondition
	}
	flds := []string{
		fldScore,
	}
	table := tblPredictResult
	return updateSQL(conn, q, d, table, flds)
}

//InsertSegmentMatch inserts the matched context data
func (c *CreditSQLDao) InsertSegmentMatch(conn SqlLike, s *SegmentMatch) (int64, error) {
	if conn == nil {
		return 0, ErroNoConn
	}
	if s == nil {
		return 0, nil
	}

	table := tblSegmentPredict
	insertFlds := []string{
		fldSegID,
		fldRelTagID,
		fldScore,
		fldMatch,
		fldMatchText,
		fldCreateTime,
		fldUpdateTime,
		fldWhos,
	}

	var params []interface{}
	params = append(params, s.SegID, s.TagID, s.Score, s.Match, s.MatchedText, s.CreateTime, s.CreateTime, s.Whos)

	return insertRecord(conn, table, insertFlds, params)
}

func insertRecord(conn SqlLike, table string, fields []string, params []interface{}) (int64, error) {
	for i := 0; i < len(fields); i++ {
		fields[i] = "`" + fields[i] + "`"
	}
	fldStr := strings.Join(fields, ",")
	numOfFlds := len(fields)
	blockStr := "(?" + strings.Repeat(",?", numOfFlds-1) + ")"

	insertSQL := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s", table, fldStr, blockStr)
	res, err := conn.Exec(insertSQL, params...)
	if err != nil {
		logger.Error.Printf("insert sql failed. %s.\n %s, %+v\n", err, insertSQL, params)
		return 0, err
	}
	return res.LastInsertId()
}

//GetCallCredit gets the credit for given call
func (c *CreditSQLDao) GetCallCredit(conn SqlLike, q *CreditQuery) ([]*SimpleCredit, error) {
	if conn == nil {
		return nil, ErroNoConn
	}

	flds := []string{
		fldID,
		fldCallID,
		fldType,
		fldParentID,
		fldOrgID,
		fldValid,
		fldRevise,
		fldScore,
		fldCreateTime,
		fldUpdateTime,
		fldWhos,
		fldDescription,
	}
	for i, v := range flds {
		flds[i] = "`" + v + "`"
	}

	table := tblPredictResult
	selectFldsStr := strings.Join(flds, ",")

	condition, params, err := q.whereSQL()
	if err != nil {
		return nil, ErrGenCondition
	}

	querySQL := fmt.Sprintf("SELECT %s FROM %s %s ORDER BY %s ASC", selectFldsStr, table, condition, fldType)
	rows, err := conn.Query(querySQL, params...)
	if err != nil {
		logger.Error.Printf("get rows failed. %s. sql:%s %+v\n", err, querySQL, params)
		return nil, err
	}
	defer rows.Close()

	resp := make([]*SimpleCredit, 0, 10)
	for rows.Next() {
		var s SimpleCredit
		err = rows.Scan(&s.ID, &s.CallID, &s.Type, &s.ParentID, &s.OrgID, &s.Valid, &s.Revise, &s.Score, &s.CreateTime, &s.UpdateTime, &s.Whos, &s.Comment)
		if err != nil {
			logger.Error.Printf("Scan failed. %s\n", err)
			return nil, err
		}
		resp = append(resp, &s)
	}

	return resp, nil
}

//GetSegmentMatch gets the matched words in segment
//not allowed retreive the whole matched words
//must give the segments id
func (c *CreditSQLDao) GetSegmentMatch(conn SqlLike, q *SegmentPredictQuery) ([]*SegmentMatch, error) {
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

	table := tblSegmentPredict
	flds := []string{
		fldID,
		fldSegID,
		fldRelTagID,
		fldScore,
		fldMatch,
		fldMatchText,
		fldCreateTime,
		fldUpdateTime,
		fldWhos,
	}
	for i := range flds {
		flds[i] = "`" + flds[i] + "`"
	}

	fldStr := strings.Join(flds, ",")

	querySQL := fmt.Sprintf("SELECT %s FROM %s %s",
		fldStr, table, condition)

	rows, err := conn.Query(querySQL, params...)
	if err != nil {
		logger.Error.Printf("query failed. %s %+v\n %s\n", querySQL, params, err)
		return nil, err
	}
	defer rows.Close()

	resp := make([]*SegmentMatch, 0)
	for rows.Next() {
		var seg SegmentMatch
		err = rows.Scan(&seg.ID, &seg.SegID, &seg.TagID, &seg.Score, &seg.Match,
			&seg.MatchedText, &seg.CreateTime, &seg.UpdateTime, &seg.Whos)
		if err != nil {
			logger.Error.Printf("scan error. %s\n", err)
			return nil, err
		}
		resp = append(resp, &seg)
	}
	return resp, nil
}
