package model

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"emotibot.com/emotigo/pkg/logger"
)

//TrainedModelDao defines the db access function
type TrainedModelDao interface {
	TrainedModelInfo(conn SqlLike, q *TModelQuery) ([]*TModel, error)
	NewModel(conn SqlLike, q *TModel) (int64, error)
	UpdateModel(conn SqlLike, q *TModel) (int64, error)
	DeleteModel(conn SqlLike, q *TModelQuery) (int64, error)
}

//TrainedModelSQLDao implements the
type TrainedModelSQLDao struct {
}

//TModel is the structure of the trained model information
type TModel struct {
	ID         uint64
	Enterprise string
	CreateTime int64
	UpdateTime int64
	Status     int
}

//error message
var (
	ErrNeedRequest = errors.New("Need parameter")
)

//TModelQuery uses as query structure of model
type TModelQuery struct {
	ID         []uint64
	Status     *int
	Enterprise *string
}

func (q *TModelQuery) whereSQL() (condition string, params []interface{}) {
	flds := make([]string, 0, 4)
	if len(q.ID) > 0 {
		qStr := fldID + " IN " + " (?" + strings.Repeat(",?", len(q.ID)-1) + ")"
		flds = append(flds, qStr)
		for _, id := range q.ID {
			params = append(params, id)
		}

	}
	if q.Status != nil {
		flds = append(flds, fldStatus+"=?")
		params = append(params, *q.Status)
	}
	if q.Enterprise != nil {
		flds = append(flds, fldEnterprise+"=?")
		params = append(params, *q.Enterprise)
	}
	if len(flds) > 0 {
		condition = "WHERE "
		condition = condition + strings.Join(flds, " AND ")
	}
	return
}

//TrainedModelInfo gets the model information
func (m *TrainedModelSQLDao) TrainedModelInfo(conn SqlLike, q *TModelQuery) ([]*TModel, error) {

	if conn == nil {
		return nil, ErrNilSqlLike
	}

	var condition string
	var params []interface{}
	if q != nil {
		condition, params = q.whereSQL()
	}
	flds := []string{
		fldID,
		fldCreateTime,
		fldUpdateTime,
		fldStatus,
		fldEnterprise,
	}
	querySQL := fmt.Sprintf("SELECT %s FROM %s %s",
		strings.Join(flds, ","),
		tblTrainedModel, condition)

	rows, err := conn.Query(querySQL, params...)
	if err != nil {
		logger.Error.Printf("query failed. %s\n query: %s, params:%+v", err, querySQL, params)
		return nil, err
	}
	defer rows.Close()

	models := make([]*TModel, 0, 10)
	for rows.Next() {
		var m TModel
		err = rows.Scan(&m.ID, &m.CreateTime, &m.UpdateTime, &m.Status, &m.Enterprise)
		if err != nil {
			logger.Error.Printf("%s\n", err)
			return nil, err
		}
		models = append(models, &m)
	}
	return models, nil
}

//NewModel inserts a new model information
func (m *TrainedModelSQLDao) NewModel(conn SqlLike, q *TModel) (int64, error) {

	if conn == nil {
		return 0, ErrNilSqlLike
	}
	if q == nil {
		return 0, ErrNeedRequest
	}
	flds := []string{
		fldCreateTime,
		fldUpdateTime,
		fldStatus,
		fldEnterprise,
	}

	insertSQL := fmt.Sprintf("INSERT INTO %s (%s) VALUES (?%s)",
		tblTrainedModel, strings.Join(flds, ","),
		strings.Repeat(",?", len(flds)-1))

	res, err := conn.Exec(insertSQL, q.CreateTime, q.CreateTime, q.Status, q.Enterprise)
	if err != nil {
		logger.Error.Printf("insert model info failed.%s\n sql:%s\n", err, insertSQL)
		return 0, err
	}
	return res.LastInsertId()
}

//DeleteModel deletes the record
func (m *TrainedModelSQLDao) DeleteModel(conn SqlLike, q *TModelQuery) (int64, error) {
	if conn == nil {
		return 0, ErrNilSqlLike
	}
	if q == nil {
		return 0, ErrNeedCondition
	}

	condition, params := q.whereSQL()
	if condition == "" {
		return 0, ErrNeedCondition
	}

	deleteSQL := fmt.Sprintf("DELETE FROM %s %s", tblTrainedModel, condition)
	res, err := conn.Exec(deleteSQL, params...)
	if err != nil {
		logger.Error.Printf("delete model info failed.%s\n sql:%s params:%+v\n", err, deleteSQL, params)
		return 0, err
	}
	return res.RowsAffected()
}

//UpdateModel updates the model information, currently only using status
func (m *TrainedModelSQLDao) UpdateModel(conn SqlLike, q *TModel) (int64, error) {
	if conn == nil {
		return 0, ErrNilSqlLike
	}
	if q == nil {
		return 0, ErrNeedRequest
	}
	now := time.Now().Unix()
	updateSQL := fmt.Sprintf("UPDATE %s SET %s=?,%s=%d WHERE %s=?",
		tblTrainedModel, fldStatus, fldUpdateTime, now, fldID)
	res, err := conn.Exec(updateSQL, q.Status, q.ID)
	if err != nil {
		logger.Error.Printf("update model info failed.%s\n sql:%s params:%d,%d\n", err, updateSQL, q.Status, q.ID)
		return 0, err
	}
	return res.RowsAffected()
}
