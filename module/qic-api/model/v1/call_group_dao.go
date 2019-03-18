package model

import (
	"fmt"
	"strings"

	"emotibot.com/emotigo/pkg/logger"
)

// CallGroupDao defines dao interface for CallGroupCondition
type CallGroupDao interface {
	CreateCondition(conn SqlLike, model *CallGroupCondition) (int64, error)
	GetConditionList(conn SqlLike, query *GeneralQuery, pagination *Pagination) ([]*CallGroupCondition, error)
	CountCondition(conn SqlLike, query *GeneralQuery) (int64, error)
	UpdateCondition(conn SqlLike, query *GeneralQuery, model *CallGroupConditionUpdateSet) (int64, error)
	SoftDeleteCondition(conn SqlLike, query *GeneralQuery) (int64, error)
}

// CallGroupSQLDao defines SQL implementation of CallGroupDao
type CallGroupSQLDao struct {
}

// CallGroupCondition defines the initial call group condition model
type CallGroupCondition struct {
	ID          int64  `json:"cg_condition_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Enterprise  string `json:"-"`
	IsEnable    int    `json:"-"`
	IsDelete    int    `json:"-"`
	DayRange    int    `json:"day_range"`
	DurationMin int    `json:"duration_min"`
	DurationMax int    `json:"duration_max"`
	CreateTime  int64  `json:"-"`
	UpdateTime  int64  `json:"-"`
}

// CallGroupConditionUpdateSet defines the json body of handleUpdateCallGroup request
type CallGroupConditionUpdateSet struct {
	Name        *string `json:name`
	Description *string `json:"description"`
}

// CallGroupConditionListResponseItem defines the item in the response data list of handleGetCallGroupList
type CallGroupConditionListResponseItem struct {
	ID          int64  `json:"cg_condition_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsEnable    bool   `json:"is_enable"`
}

// CallGroupConditionResponse defines the response data of handleGetCallGroup
type CallGroupConditionResponse struct {
	Name          string        `json:"name"`
	Description   string        `json:"description"`
	IsEnable      bool          `json:"is_enable"`
	DayRange      int           `json:"day_range"`
	DurationMin   int           `json:"duration_min"`
	DurationMax   int           `json:"duration_max"`
	FilterByValue []interface{} `json:"filter_by_value"`
	GroupByValue  []string      `json:"group_by_value"`
}

var (
	callGroupConditionFlds = []string{
		fldID,
		fldName,
		fldDescription,
		fldEnterprise,
		fldIsEnable,
		fldIsDelete,
		fldDayRange,
		fldDurationMin,
		fldDurationMax,
		fldCreateTime,
		fldUpdateTime,
	}
)

//CreateCondition careate a new CallGroupCondition
func (*CallGroupSQLDao) CreateCondition(conn SqlLike, model *CallGroupCondition) (int64, error) {
	if conn == nil {
		return 0, ErroNoConn
	}
	if model == nil {
		return 0, ErrNeedRequest
	}
	flds := callGroupConditionFlds
	for i := range flds {
		flds[i] = "`" + flds[i] + "`"
	}
	vals := make([]interface{}, 0, len(flds))
	err := extractSimpleStructureValue(&vals, model)
	if err != nil {
		return 0, err
	}
	//remove the ID
	vals = vals[1:]
	flds = flds[1:]
	return insertRow(conn, tblCallGroupCondition, flds, vals)
}

//GetConditionList return the requested CallGroupCondition list
func (*CallGroupSQLDao) GetConditionList(conn SqlLike, query *GeneralQuery, pagination *Pagination) ([]*CallGroupCondition, error) {
	flds := callGroupConditionFlds
	for i := range flds {
		flds[i] = "`" + flds[i] + "`"
	}

	var condition string
	var params []interface{}
	var err error
	var offset string
	if query != nil {
		condition, params, err = query.whereSQL()
		if err != nil {
			return nil, ErrGenCondition
		}
	}
	if pagination != nil {
		offset = pagination.offsetSQL()
	}
	querySQL := fmt.Sprintf("SELECT %s FROM %s %s %s", strings.Join(flds, ","), tblCallGroupCondition, condition, offset)

	rows, err := conn.Query(querySQL, params...)
	if err != nil {
		logger.Error.Printf("query failed. %s %+v\n", querySQL, params)
		return nil, err
	}
	defer rows.Close()

	condList := make([]*CallGroupCondition, 0)
	for rows.Next() {
		var data CallGroupCondition
		err = rows.Scan(&data.ID, &data.Name, &data.Description, &data.IsEnable, &data.IsDelete,
			&data.DayRange, &data.DurationMin, &data.DurationMax, &data.CreateTime, &data.UpdateTime)
		if err != nil {
			logger.Error.Printf("scan failed. %s\n", err)
			return nil, err
		}
		condList = append(condList, &data)
	}
	return condList, nil
}

//CountCondition counts the number of the requested CallGroupConditions
func (*CallGroupSQLDao) CountCondition(conn SqlLike, query *GeneralQuery) (int64, error) {
	if conn == nil {
		return 0, ErroNoConn
	}
	var condition string
	var params []interface{}
	var err error
	if query != nil {
		condition, params, err = query.whereSQL()
		if err != nil {
			return 0, ErrGenCondition
		}
	}
	return countRows(conn, tblCallGroupCondition, condition, params)
}

//UpdateCondition updates the name or description of the CallGroupCondition
func (*CallGroupSQLDao) UpdateCondition(conn SqlLike, query *GeneralQuery, data *CallGroupConditionUpdateSet) (int64, error) {
	if query == nil || len(query.ID) == 0 {
		return 0, ErrNeedCondition
	}
	flds := []string{
		fldName,
		fldDescription,
	}
	return updateSQL(conn, query, data, tblCallGroupCondition, flds)
}

//SoftDeleteCondition simply set the is_delete of CallGroupCondition to 1
func (*CallGroupSQLDao) SoftDeleteCondition(conn SqlLike, query *GeneralQuery) (int64, error) {
	if query == nil || len(query.ID) == 0 {
		return 0, ErrNeedCondition
	}
	return softDelete(conn, query, tblCallGroupCondition)
}
