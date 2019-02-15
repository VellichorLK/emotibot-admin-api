package model

import (
	"emotibot.com/emotigo/pkg/logger"
	"errors"
	"fmt"
	"strings"
)

const (
	tblSensitiveWord       = "SensitiveWord"
	tblRelSensitiveWordSen = "Relation_SensitiveWord_Sentence"
	fldRelSWID             = "sw_id"
	staffExceptionType     = 0
	customerExceptionType  = 1
	swCategoryType         = 1
)

//error message
var (
	ErrEmptyCategoryName = errors.New("Category name can not be nil")
)

type SensitiveWordFilter struct {
	UUID       []string
	Category   *int64
	Enterprise *string
	Name       string
	Page       int
	Limit      int
}

type SensitiveWordCategoryFilter struct {
	ID         []int64
	Enterprise *string
}

type SensitiveWordDao interface {
	Create(*SensitiveWord, SqlLike) (int64, error)
	CountBy(*SensitiveWordFilter, SqlLike) (int64, error)
	GetBy(*SensitiveWordFilter, SqlLike) ([]SensitiveWord, error)

	CreateCateogry(*SensitiveWordCategory, SqlLike) (int64, error)
	GetCategories(*SensitiveWordCategoryFilter, SqlLike) ([]SensitiveWordCategory, error)
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

type SensitiveWordCategory struct {
	ID         int64  `json:"category_id,string"`
	Name       string `json:"name"`
	Enterprise string `json:"-"`
}

type SensitiveWordSqlDao struct {
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
			values = append(values, word.ID, customerException.ID, customerExceptionType)

			valueStr = fmt.Sprintf("%s%s,", valueStr, variableStr)
		}

		for _, staffException := range word.StaffException {
			values = append(values, word.ID, staffException.ID, staffExceptionType)

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
	return
}

func getSensitiveWordQuerySQL(filter *SensitiveWordFilter) (queryStr string, values []interface{}) {
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

	fields := []string{
		fldID,
		fldUUID,
		fldName,
		fldEnterprise,
		fldScore,
		fldCategoryID,
	}
	fieldStr := strings.Join(fields, ", ")

	conditionStr, values := builder.ParseWithWhere()
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

func (dao *SensitiveWordSqlDao) CreateCateogry(category *SensitiveWordCategory, sqlLike SqlLike) (rowID int64, err error) {
	if category.Name == "" {
		err = ErrEmptyCategoryName
		return
	}

	fields := []string{
		fldName,
		fldEnterprise,
		fldType,
	}
	fieldStr := strings.Join(fields, ", ")

	insertStr := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (?%s)",
		tblCategory,
		fieldStr,
		strings.Repeat(", ?", len(fields)-1),
	)

	values := []interface{}{
		category.Name,
		category.Enterprise,
		swCategoryType,
	}

	result, err := sqlLike.Exec(insertStr, values...)
	if err != nil {
		logger.Error.Printf("insert sensitive word category, sql: %s\n", insertStr)
		logger.Error.Printf("values: %+v\n", values)
		err = fmt.Errorf("insert sensitive word category failed in dao.CreateCateogry, err: %s", err.Error())
		return
	}

	rowID, err = result.LastInsertId()
	if err != nil {
		err = fmt.Errorf("failed in get row id in dao.CreateCategory, err: %s", err.Error())
	}
	return
}

func (dao *SensitiveWordSqlDao) GetCategories(filter *SensitiveWordCategoryFilter, sqlLike SqlLike) (categories []SensitiveWordCategory, err error) {
	builder := NewWhereBuilder(andLogic, "")

	if len(filter.ID) > 0 {
		builder.In(fldID, int64ToWildCard(filter.ID...))
	}

	if filter.Enterprise != nil {
		builder.Eq(fldEnterprise, *filter.Enterprise)
	}

	builder.Eq(fldType, swCategoryType)

	fields := []string{
		fldID,
		fldName,
		fldEnterprise,
	}
	fieldStr := strings.Join(fields, ", ")

	conditionStr, values := builder.ParseWithWhere()
	queryStr := fmt.Sprintf(
		"SELECT %s FROM %s %s",
		fieldStr,
		tblCategory,
		conditionStr,
	)

	rows, err := sqlLike.Query(queryStr, values...)
	if err != nil {
		logger.Error.Printf("error while query sensitive word categories, sql: %s", queryStr)
		logger.Error.Printf("values: %+v\n", values)
		err = fmt.Errorf("error while query sensitive word categories in dao.GetCategories, err: %s", err.Error())
		return
	}
	defer rows.Close()

	categories = []SensitiveWordCategory{}
	for rows.Next() {
		cate := SensitiveWordCategory{}
		rows.Scan(&cate.ID, &cate.Name, &cate.Enterprise)
		categories = append(categories, cate)
	}
	return
}
