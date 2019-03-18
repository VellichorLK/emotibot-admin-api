package model

import (
	"database/sql"
	"fmt"
	"strings"
	"emotibot.com/emotigo/pkg/logger"
	"encoding/json"
)

const (
	FldTestedID  = "tested_id"
	FldHitTags   = "hit_tags"
	FldFailTags  = "fail_tags"
	FldMatchText = "match_text"
	FldHit       = "hit"
	FldTotal     = "total"
	FldAccuracy  = "accuracy"

	TblSentenceTest       = "sentence_test"
	TblSentenceTestResult = "sentence_test_result"
)

type SentenceTestDao interface {
	InsertTestSentence(tx SqlLike, s *TestSentence) error
	SoftDeleteTestSentence(tx SqlLike, s *TestSentenceQuery) (int64, error)
	CountTestSentences(tx SqlLike, query *TestSentenceQuery) (int64, error)
	GetTestSentences(tx SqlLike, query *TestSentenceQuery) ([]*TestSentence, error)
	InsertOrUpdateSentenceTestResult(tx SqlLike, result *SentenceTestResult) error
	UpdateSentenceTest(tx SqlLike, s *TestSentence) error
	GetSentenceTestResultByCategory(tx SqlLike, enterpriseID string, categoryID uint64) ([]*SentenceTestResult, error)
}

type TestSentence struct {
	ID         uint64   `json:"id"`
	IsDelete   int      `json:"is_delete"`
	Name       string   `json:"name"`
	Enterprise string   `json:"enterprise"`
	UUID       string   `json:"uuid"`
	CreateTime int64    `json:"create_time"`
	UpdateTime int64    `json:"update_time"`
	CategoryID uint64   `json:"category_id"`
	TestedID   uint64   `json:"tested_id"`
	HitTags    []uint64 `json:"hit_tags"`
	FailTags   []uint64 `json:"fail_tags"`
	MatchText  []string `json:"match_text"`
}

type TestSentenceQuery struct {
	ID         []uint64
	UUID       []string
	Enterprise *string
	IsDelete   *int8
	Name       *string
	TestedID   []uint64
	CategoryID *uint64
}

type SentenceTestResult struct {
	ID         uint64  `json:"id"`
	Name       string  `json:"name"`
	UUID       string  `json:"uuid"`
	Enterprise string  `json:"enterprise"`
	CategoryID uint64  `json:"category_id"`
	Hit        int     `json:"hit"`
	Total      int     `json:"total"`
	Accuracy   float32 `json:"accuracy"`
	CreateTime int64   `json:"create_time"`
	UpdateTime int64   `json:"update_time"`
}

func NewSentenceTestSQLDao(conn *sql.DB) *SentenceTestSQLDao {
	return &SentenceTestSQLDao{
		conn: conn,
	}
}

type SentenceTestSQLDao struct {
	conn *sql.DB
}

func (d *SentenceTestSQLDao) InsertTestSentence(tx SqlLike, s *TestSentence) error {
	if s == nil {
		return fmt.Errorf("invalid param")
	}

	exe, err := genrateExecutor(d.conn, tx)
	if err != nil {
		return err
	}

	sqlStr := fmt.Sprintf("INSERT INTO %s ( %s, %s, %s, %s, %s, %s, %s ) "+
		"VALUES ( ?, ?, ?, ?, ?, ?, ? )",
		TblSentenceTest,
		fldIsDelete, fldName, fldEnterprise, fldUUID, fldCreateTime, fldUpdateTime, FldTestedID)

	res, err := exe.Exec(sqlStr, s.IsDelete, s.Name, s.Enterprise, s.UUID, s.CreateTime, s.UpdateTime, s.TestedID)
	if err != nil {
		return err
	}
	_, err = res.LastInsertId()
	if err != nil {
		return err
	}

	return nil
}

func (d *SentenceTestSQLDao) SoftDeleteTestSentence(tx SqlLike, query *TestSentenceQuery) (int64, error) {
	if query == nil || query.Enterprise == nil || len(query.UUID) == 0 {
		return 0, fmt.Errorf("invalid param")
	}

	exe, err := genrateExecutor(d.conn, tx)
	if err != nil {
		return 0, err
	}

	condition, params := query.whereSQL()
	sqlStr := fmt.Sprintf("UPDATE %s SET %s = '1' %s", TblSentenceTest, fldIsDelete, condition)
	result, err := exe.Exec(sqlStr, params...)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

func (d *SentenceTestSQLDao) CountTestSentences(tx SqlLike, query *TestSentenceQuery) (int64, error) {
	q, err := genrateExecutor(d.conn, tx)
	if err != nil {
		return 0, nil
	}

	var condition string
	var params []interface{}
	if query != nil {
		condition, params = query.whereSQL()
	}

	queryStr := fmt.Sprintf("SELECT COUNT(*) FROM %s %s", "sentence_test", condition)
	var num int64
	if err = q.QueryRow(queryStr, params...).Scan(&num); err != nil {
		logger.Error.Printf("fail to query. %s \n", err.Error())
		return 0, err
	}
	return num, nil
}

func (d *SentenceTestSQLDao) GetTestSentences(tx SqlLike, query *TestSentenceQuery) ([]*TestSentence, error) {
	q, err := genrateExecutor(d.conn, tx)
	if err != nil {
		return nil, err
	}

	if query == nil {
		return nil, fmt.Errorf("invalid param")
	}
	condition, params := query.whereSQL()

	queryStr := fmt.Sprintf("SELECT %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s FROM %s %s",
		fldID, fldIsDelete, fldName, fldEnterprise, fldUUID, fldCreateTime, fldUpdateTime, fldCategoryID, FldTestedID, FldHitTags, FldFailTags, FldMatchText,
		TblSentenceTest, condition)
	rows, err := q.Query(queryStr, params...)
	if err != nil {
		logger.Error.Printf("Query: %s, Params:%+v, failed. %s\n", queryStr, params, err)
		return nil, err
	}
	defer rows.Close()

	testSentences := make([]*TestSentence, 0)
	for rows.Next() {
		var s TestSentence
		var hitTagsStr, failTagsStr, matchTextStr string
		err = rows.Scan(&s.ID, &s.IsDelete, &s.Name, &s.Enterprise, &s.UUID, &s.CreateTime, &s.UpdateTime, &s.CategoryID, &s.TestedID, &hitTagsStr, &failTagsStr, &matchTextStr)
		if err != nil {
			logger.Error.Printf("fail to query: %s \n", err.Error())
			return nil, err
		}

		if err = json.Unmarshal([]byte(hitTagsStr), &s.HitTags); err != nil {
			return nil, err
		}

		if err = json.Unmarshal([]byte(failTagsStr), &s.FailTags); err != nil {
			return nil, err
		}

		if err = json.Unmarshal([]byte(matchTextStr), &s.MatchText); err != nil {
			return nil, err
		}

		testSentences = append(testSentences, &s)
	}

	return testSentences, nil
}

func (query TestSentenceQuery) whereSQL() (string, []interface{}) {
	numOfUUID := len(query.UUID)
	params := make([]interface{}, 0, numOfUUID+1)
	conditions := make([]string, 0)

	if numOfUUID > 0 {
		condition := fldUUID + " IN ( ?" + strings.Repeat(", ?", numOfUUID-1) + " )"
		conditions = append(conditions, condition)
		for i := 0; i < numOfUUID; i++ {
			params = append(params, query.UUID[i])
		}
	}

	numOfID := len(query.ID)
	if numOfID > 0 {
		condition := fldID + " IN ( ?" + strings.Repeat(", ?", numOfID-1) + " )"
		conditions = append(conditions, condition)
		for i := 0; i < numOfID; i++ {
			params = append(params, query.ID[i])
		}
	}

	numOfTestedID := len(query.TestedID)
	if numOfTestedID > 0 {
		condition := "tested_id" + " IN ( ?" + strings.Repeat(", ?", numOfTestedID-1) + " )"
		conditions = append(conditions, condition)
		for i := 0; i < numOfTestedID; i++ {
			params = append(params, query.TestedID[i])
		}
	}

	if query.Enterprise != nil {
		condition := fldEnterprise + " = ?"
		conditions = append(conditions, condition)
		params = append(params, *query.Enterprise)
	}

	if query.IsDelete != nil {
		condition := fldIsDelete + " = ?"
		conditions = append(conditions, condition)
		params = append(params, *query.IsDelete)
	}

	if query.Name != nil {
		condition := fldName + " = ?"
		conditions = append(conditions, condition)
		params = append(params, *query.Name)
	}

	if query.CategoryID != nil {
		condition := fldCategoryID + " = ?"
		conditions = append(conditions, condition)
		params = append(params, *query.CategoryID)
	}

	var whereSQL string
	if len(conditions) > 0 {
		whereSQL = "WHERE " + strings.Join(conditions, " AND ")
	}
	whereSQL += " ORDER BY " + fldID + " DESC "

	return whereSQL, params
}

func (d *SentenceTestSQLDao) InsertOrUpdateSentenceTestResult(tx SqlLike, result *SentenceTestResult) error {
	if result == nil {
		return fmt.Errorf("invalid param")
	}
	exe, err := genrateExecutor(d.conn, tx)
	if err != nil {
		return nil
	}

	queryStr := fmt.Sprintf("SELECT COUNT(*), %s FROM %s WHERE %s = ? AND %s = ? AND %s = ? GROUP BY %s",
		fldID, TblSentenceTestResult, fldEnterprise, fldIsDelete, fldName, fldID)
	var num int
	var id uint64
	err = exe.QueryRow(queryStr, result.Enterprise, 0, result.Name).Scan(&num, &id)

	var sqlStr string
	if err != nil {
		if err == sql.ErrNoRows {
			sqlStr = fmt.Sprintf("INSERT INTO %s ( %s, %s, %s, %s, %s, %s, %s, %s, %s, %s ) "+"VALUES ( ?, ?, ?, ?, ?, ?, ?, ?, ?, ? )",
				TblSentenceTestResult, fldIsDelete, fldName, fldEnterprise, fldUUID, fldCreateTime, fldUpdateTime, fldCategoryID, FldHit, FldTotal, FldAccuracy)
			_, err = exe.Exec(sqlStr, 0, result.Name, result.Enterprise, result.UUID, result.CreateTime, result.UpdateTime, result.CategoryID, result.Hit, result.Total, result.Accuracy)
			return err
		}
		return err
	}

	if num == 1 {
		sqlStr = fmt.Sprintf("UPDATE %s SET %s = ?, %s = ?, %s = ?, %s = ? WHERE %s = ?",
			TblSentenceTestResult, FldHit, FldTotal, FldAccuracy, fldUpdateTime, fldID)
		_, err = exe.Exec(sqlStr, result.Hit, result.Total, result.Accuracy, result.UpdateTime, id)
	} else {
		return fmt.Errorf("find more than one existing sentenceTestResult")
	}

	return nil
}

func (d *SentenceTestSQLDao) UpdateSentenceTest(tx SqlLike, s *TestSentence) error {
	if s == nil {
		return fmt.Errorf("invalid param")
	}
	exe, err := genrateExecutor(d.conn, tx)
	if err != nil {
		return err
	}

	// TODO add more columns
	sqlStr := fmt.Sprintf("UPDATE %s SET %s = ?, %s = ?, %s = ? WHERE %s = ?",
		TblSentenceTest, FldHitTags, FldFailTags, FldMatchText, fldID)

	hitTagsStr, err := json.Marshal(s.HitTags)
	if err != nil {
		return err
	}
	failTagsStr, err := json.Marshal(s.FailTags)
	if err != nil {
		return err
	}
	matchTextStr, err := json.Marshal(s.MatchText)
	if err != nil {
		return err
	}

	_, err = exe.Exec(sqlStr, hitTagsStr, failTagsStr, matchTextStr, s.ID)

	return err
}

func (d *SentenceTestSQLDao) GetSentenceTestResultByCategory(tx SqlLike, enterpriseID string, categoryID uint64) ([]*SentenceTestResult, error) {
	q, err := genrateExecutor(d.conn, tx)
	if err != nil {
		return nil, err
	}

	queryStr := fmt.Sprintf("SELECT %s, %s, %s, %s, %s, %s, %s, %s, %s, %s FROM %s WHERE %s = ? AND %s = ?",
		fldID, fldName, fldEnterprise, fldUUID, fldCreateTime, fldUpdateTime, fldCategoryID, FldHit, FldTotal, FldAccuracy, TblSentenceTestResult, fldEnterprise, fldCategoryID)
	rows, err := q.Query(queryStr, enterpriseID, categoryID)
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	results := make([]*SentenceTestResult, 0)
	for rows.Next() {
		var result SentenceTestResult
		rows.Scan(&result.ID, &result.Name, &result.Enterprise, &result.UUID, &result.CreateTime, &result.UpdateTime, &result.CategoryID, &result.Hit, &result.Total, &result.Accuracy)
		results = append(results, &result)

	}
	return results, nil
}
