package model

import (
	"fmt"
	"strings"

	"emotibot.com/emotigo/pkg/logger"
)

//CreditDao is the interface used to access the prediction result
type CreditDao interface {
	InsertCredit(conn SqlLike, c *SimpleCredit) (int64, error)
	InsertSegmentMatch(conn SqlLike, s *SegmentMatch) (int64, error)
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
}

//SegmentMatch is the structure used to insert the matched segment
type SegmentMatch struct {
	ID          uint64
	SegID       uint64
	TagID       uint64
	Score       int
	Match       string
	MatchedText string
	CreateTime  int64
	UpdateTime  int64
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
	}

	var params []interface{}

	params = append(params, credit.CallID, credit.Type, credit.ParentID,
		credit.OrgID, credit.Valid, credit.Revise, credit.Score, credit.CreateTime,
		credit.CreateTime)

	return insertRecord(conn, table, insertFlds, params)
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
	}

	var params []interface{}
	params = append(params, s.SegID, s.TagID, s.Score, s.Match, s.MatchedText, s.CreateTime, s.CreateTime)

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
