package model

import (
	"fmt"
	"strings"

	"emotibot.com/emotigo/pkg/logger"
)

//TrainedModelDao defines the db access function
type TrainedModelDao interface {
	TrainedModelInfo(conn SqlLike, q *TModelQuery) ([]*TModel, error)
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

//TModelQuery uses as query structure of model
type TModelQuery struct {
	Status     *int
	Enterprise *string
}

func (q *TModelQuery) whereSQL() (condition string, params []interface{}) {
	flds := make([]string, 0, 4)
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
