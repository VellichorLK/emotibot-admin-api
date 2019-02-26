package model

import (
	"emotibot.com/emotigo/pkg/logger"
	"errors"
	"fmt"
	"github.com/mediocregopher/radix"
	"strings"
)

const (
	tblSensitiveWord       = "SensitiveWord"
	tblRelSensitiveWordSen = "Relation_SensitiveWord_Sentence"
	fldRelSWID             = "sw_id"
	redisKey               = "sensitive-words"
	StaffExceptionType     = int8(0)
	CustomerExceptionType  = int8(1)
	SwCategoryType         = int8(1)
)

//error message
var (
	ErrEmptyCategoryName = errors.New("Category name can not be nil")
)

type SensitiveWordFilter struct {
	ID         []int64
	UUID       []string
	Category   *int64
	Enterprise *string
	Keyword    string
	Page       int
	Limit      int
	Deleted    *int8
}

func (filter *SensitiveWordFilter) Where() (string, []interface{}) {
	builder := NewWhereBuilder(andLogic, "")

	if filter.Enterprise != nil {
		builder.Eq(fldEnterprise, *filter.Enterprise)
	}

	if filter.Category != nil {
		builder.Eq(fldCategoryID, *filter.Category)
	}

	if len(filter.UUID) > 0 {
		builder.In(fldUUID, stringToWildCard(filter.UUID...))
	}

	if len(filter.ID) > 0 {
		builder.In(fldID, int64ToWildCard(filter.ID...))
	}

	if filter.Deleted != nil {
		builder.Eq(fldIsDelete, *filter.Deleted)
	}
	return builder.ParseWithWhere()
}

type SensitiveWordDao interface {
	Create(*SensitiveWord, SqlLike) (int64, error)
	CountBy(*SensitiveWordFilter, SqlLike) (int64, error)
	GetBy(*SensitiveWordFilter, SqlLike) ([]SensitiveWord, error)
	GetRel(int64, SqlLike) (map[int8][]uint64, error)
	Delete(*SensitiveWordFilter, SqlLike) (int64, error)
	Move(*SensitiveWordFilter, int64, SqlLike) (int64, error)
	Names(SqlLike, bool) ([]string, error)
}

type SensitiveWord struct {
	ID                int64
	UUID              string
	Name              string
	Score             int
	StaffException    []SimpleSentence
	CustomerException []SimpleSentence
	Enterprise        string
	CategoryID        int64
}

type SensitiveWordSqlDao struct {
	Redis *radix.Cluster
}

func NewDefaultSensitiveWordDao(cluster *radix.Cluster) SensitiveWordDao {
	return &SensitiveWordSqlDao{
		Redis: cluster,
	}
}

func getSensitiveWordInsertSQL(words []SensitiveWord) (insertStr string, values []interface{}) {
	values = []interface{}{}
	if len(words) == 0 {
		return
	}

	fields := []string{
		fldUUID,
		fldName,
		fldEnterprise,
		fldScore,
		fldCategoryID,
	}
	fieldStr := strings.Join(fields, ", ")

	variableStr := fmt.Sprintf(
		"(?%s)",
		strings.Repeat(", ?", len(fields)-1),
	)

	valueStr := ""
	for _, word := range words {
		values = append(
			values,
			word.UUID,
			word.Name,
			word.Enterprise,
			word.Score,
			word.CategoryID,
		)

		valueStr = fmt.Sprintf("%s%s,", valueStr, variableStr)
	}
	// remove last comma
	valueStr = valueStr[:len(valueStr)-1]
	insertStr = fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES %s",
		tblSensitiveWord,
		fieldStr,
		valueStr,
	)
	return
}

func getSensitiveWordRelationInsertSQL(words []SensitiveWord) (insertStr string, values []interface{}) {
	values = []interface{}{}
	if len(words) == 0 {
		return
	}

	fields := []string{
		fldRelSWID,
		fldRelSenID,
		fldType,
	}
	fiedlStr := strings.Join(fields, ", ")

	variableStr := fmt.Sprintf("(?%s)", strings.Repeat(", ?", len(fields)-1))
	valueStr := ""

	for _, word := range words {
		for _, customerException := range word.CustomerException {
			values = append(values, word.ID, customerException.ID, CustomerExceptionType)

			valueStr = fmt.Sprintf("%s%s,", valueStr, variableStr)
		}

		for _, staffException := range word.StaffException {
			values = append(values, word.ID, staffException.ID, StaffExceptionType)

			valueStr = fmt.Sprintf("%s%s,", valueStr, variableStr)
		}
	}
	if len(values) > 0 {
		valueStr = valueStr[:len(valueStr)-1]
	}

	insertStr = fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES %s",
		tblRelSensitiveWordSen,
		fiedlStr,
		valueStr,
	)
	return
}

// Create creates a record to SensitiveWord table and relation records to Relation_SensitiveWord_Sentence
func (dao *SensitiveWordSqlDao) Create(word *SensitiveWord, sqlLike SqlLike) (rowID int64, err error) {
	// insert to sensitive word table
	if word == nil {
		return
	}
	insertStr, values := getSensitiveWordInsertSQL([]SensitiveWord{*word})

	result, err := sqlLike.Exec(insertStr, values...)
	if err != nil {
		logger.Error.Printf("insert sensitive word failed, sql: %s\n", insertStr)
		logger.Error.Printf("values: %+v\n", values)
		err = fmt.Errorf("insert sensitive word failed in dao.Create, err: %s", err.Error())
		return
	}

	// get row id
	rowID, err = result.LastInsertId()
	if err != nil {
		err = fmt.Errorf("get row id failed in dao.Create, err: %s", err.Error())
		logger.Error.Printf(err.Error())
		return
	}
	word.ID = rowID

	// insert relation
	insertStr, values = getSensitiveWordRelationInsertSQL([]SensitiveWord{*word})
	if len(values) > 0 {
		_, err = sqlLike.Exec(insertStr, values...)
		if err != nil {
			logger.Error.Printf("insert sensitive word relation, sql: %s\n", insertStr)
			logger.Error.Printf("values: %+v\n", values)
			err = fmt.Errorf("insert sensitive word sentence relation in dao.Create, err: %s", err.Error())
		}
	}

	// update redis
	names, ierr := dao.Names(sqlLike, true)
	if ierr != nil {
		logger.Error.Printf("get sensitive names failed, err: %s", ierr.Error())
		return
	}

	if dao.Redis != nil {
		ierr = dao.Redis.Do(radix.Cmd(nil, "DEL", redisKey))
		if ierr != nil {
			logger.Error.Print(ierr)
			return
		}

		cmds := append([]string{redisKey}, names...)
		ierr = dao.Redis.Do(radix.Cmd(nil, "LPUSH", cmds...))
		if ierr != nil {
			logger.Error.Print(ierr)
			return
		}
	}
	return
}

func getSensitiveWordQuerySQL(filter *SensitiveWordFilter) (queryStr string, values []interface{}) {
	fields := []string{
		fldID,
		fldUUID,
		fldName,
		fldEnterprise,
		fldScore,
		fldCategoryID,
	}
	fieldStr := strings.Join(fields, ", ")

	conditionStr, values := filter.Where()
	queryStr = fmt.Sprintf(
		"SELECT %s FROM %s %s",
		fieldStr,
		tblSensitiveWord,
		conditionStr,
	)
	return
}

func (dao *SensitiveWordSqlDao) CountBy(filter *SensitiveWordFilter, sqlLike SqlLike) (total int64, err error) {
	queryStr, values := getSensitiveWordQuerySQL(filter)
	queryStr = fmt.Sprintf("SELECT COUNT(sw.%s) FROM (%s) as sw", fldID, queryStr)

	rows, err := sqlLike.Query(queryStr, values...)
	if err != nil {
		logger.Error.Printf("error when count rows in dao.CountBy, sql: %s\n", queryStr)
		logger.Error.Printf("values: %+v\n", values)
		err = fmt.Errorf("count sensitive words failed, err: %s", err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&total)
	}
	return
}

func (dao *SensitiveWordSqlDao) GetBy(filter *SensitiveWordFilter, sqlLike SqlLike) (sensitiveWords []SensitiveWord, err error) {
	queryStr, values := getSensitiveWordQuerySQL(filter)

	rows, err := sqlLike.Query(queryStr, values...)
	if err != nil {
		logger.Error.Printf("error when count rows in dao.GetBy, sql: %s\n", queryStr)
		logger.Error.Printf("values: %+v\n", values)
		err = fmt.Errorf("get sensitive words failed, err: %s", err.Error())
		return
	}
	defer rows.Close()

	sensitiveWords = []SensitiveWord{}
	for rows.Next() {
		word := SensitiveWord{}
		rows.Scan(
			&word.ID,
			&word.UUID,
			&word.Name,
			&word.Enterprise,
			&word.Score,
			&word.CategoryID,
		)
		sensitiveWords = append(sensitiveWords, word)
	}
	return
}

func (dao *SensitiveWordSqlDao) GetRel(id int64, sqlLike SqlLike) (rel map[int8][]uint64, err error) {
	// get relations
	queryStr := fmt.Sprintf(
		"SELECT %s, %s from %s WHERE %s = ?",
		fldRelSenID,
		fldType,
		tblRelSensitiveWordSen,
		fldRelSWID,
	)

	rows, err := sqlLike.Query(queryStr, id)
	if err != nil {
		logger.Error.Printf("error while query sensitive words relations, sql: %s", queryStr)
		logger.Error.Printf("values: %d", id)
		err = fmt.Errorf("error while query sensitive words relations, err: %s", err.Error())
		return
	}
	defer rows.Close()

	rel = map[int8][]uint64{}
	for rows.Next() {
		var sid uint64
		var relType int8
		err = rows.Scan(&sid, &relType)
		if err != nil {
			err = fmt.Errorf("error while parse sensitive words relations, err: %s", err.Error())
			return
		}

		if _, ok := rel[relType]; !ok {
			rel[relType] = []uint64{}
		}

		rel[relType] = append(rel[relType], sid)
	}
	return
}

func (dao *SensitiveWordSqlDao) Delete(filter *SensitiveWordFilter, sqlLike SqlLike) (affectedRows int64, err error) {
	conditionStr, values := filter.Where()
	deleteStr := fmt.Sprintf(
		"UPDATE %s SET %s = 1 %s",
		tblSensitiveWord,
		fldIsDelete,
		conditionStr,
	)

	result, err := sqlLike.Exec(deleteStr, values...)
	if err != nil {
		logger.Error.Printf("error while set words deleted, sql: %s", deleteStr)
		logger.Error.Printf("values: %+v", values)
		err = fmt.Errorf("error while set words deleted, err: %s", err.Error())
		return
	}

	affectedRows, err = result.RowsAffected()
	if err != nil {
		err = fmt.Errorf("error while get rows affected, err: %s", err.Error())
	}
	return
}

func (dao *SensitiveWordSqlDao) Move(filter *SensitiveWordFilter, categoryID int64, sqlLike SqlLike) (affectedRows int64, err error) {
	conditionStr, values := filter.Where()
	if len(values) == 0 {
		return
	}

	sqlStr := fmt.Sprintf(
		"UPDATE %s SET %s = %d %s",
		tblSensitiveWord,
		fldCategoryID,
		categoryID,
		conditionStr,
	)

	result, err := sqlLike.Exec(sqlStr, values...)
	if err != nil {
		logger.Error.Printf("error while move sensitive word to another category, sql: %s", sqlStr)
		logger.Error.Printf("values: %+v", values)
		err = fmt.Errorf("error while move sensitive word to another category, err: %s", err.Error())
		return
	}

	affectedRows, err = result.RowsAffected()
	if err != nil {
		err = fmt.Errorf("error while get affcted rows when move sensitive word to another category, err: %s", err.Error())
		return
	}
	return
}

// Names is a sugar function for getting all sensitive word names
func (dao *SensitiveWordSqlDao) Names(sqlLike SqlLike, forceReload bool) (names []string, err error) {
	names = []string{}
	if dao.Redis != nil && !forceReload {
		err = dao.Redis.Do(radix.Cmd(&names, "LRANGE", redisKey, "0", "-1"))
		if err != nil {
			logger.Error.Printf("get sensitive word names in redis failed, err: %s", err.Error())
		}
		return
	}

	// get names through mysql
	queryStr := fmt.Sprintf(
		"SELECT %s FROM %s WHERE %s=0",
		fldName,
		tblSensitiveWord,
		fldIsDelete,
	)
	rows, err := sqlLike.Query(queryStr)
	if err != nil {
		err = fmt.Errorf("get sensitive word names in mysql failed, err: %s", err.Error())
		return
	}

	for rows.Next() {
		var name string
		rows.Scan(&name)
		names = append(names, name)
	}
	return
}
