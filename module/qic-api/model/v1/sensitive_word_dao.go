package model

import (
	"emotibot.com/emotigo/pkg/logger"
	"fmt"
	"strings"
)

const (
	tblSensitiveWord       = "SensitiveWord"
	tblRelSensitiveWordSen = "Relation_SensitiveWord_Sentence"
	fldRelSWID             = "sw_id"
	staffExceptionType     = 0
	customerExceptionType  = 1
)

type SensitiveWordDao interface {
	Create(*SensitiveWord, SqlLike) (int64, error)
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
