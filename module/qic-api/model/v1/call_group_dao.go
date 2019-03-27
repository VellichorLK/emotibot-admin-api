package model

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"emotibot.com/emotigo/pkg/logger"
)

// CallGroupDao defines dao interface call group
type CallGroupDao interface {
	CreateCondition(conn SqlLike, model *CallGroupCondition) (int64, error)
	GetConditionList(conn SqlLike, query *GeneralQuery, pagination *Pagination) ([]*CallGroupCondition, error)
	CountCondition(conn SqlLike, query *GeneralQuery) (int64, error)
	UpdateCondition(conn SqlLike, query *GeneralQuery, model *CallGroupConditionUpdateSet) (int64, error)
	SoftDeleteCondition(conn SqlLike, query *GeneralQuery) (int64, error)
	// GetCGCondition(conn SqlLike, enterprise string) ([]*CGCondition, error)
	GetCGCondition(conn SqlLike, query *CGConditionQuery) ([]*CGCondition, error)
	GetCallIDsToGroup(conn SqlLike, query *CallsToGroupQuery) ([]int64, error)
	GetCallGroups(conn SqlLike, query *CallGroupQuery) ([]*CallGroup, error)
	SoftDeleteCallGroup(conn SqlLike, query *GeneralQuery) error
	CreateCallGroup(conn SqlLike, model *CallGroup) (int64, error)
	CreateCallGroupRelation(conn SqlLike, model *CallGroupRelation) (int64, error)
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

// CallGroupConditionUpdateSet defines the json body of handleUpdateCallGroupCondition request
type CallGroupConditionUpdateSet struct {
	Name        *string `json:name`
	Description *string `json:"description"`
	IsEnable    *int    `json:"is_enable"`
}

// CallGroupConditionListResponseItem defines the item in the response data list of handleGetCallGroupConditionList
type CallGroupConditionListResponseItem struct {
	ID          int64  `json:"cg_condition_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsEnable    bool   `json:"is_enable"`
}

// CallGroupConditionResponse defines the response data of handleGetCallGroupCondition
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

// CallGroupCreateList defines the CallGroup list structrue to be created
type CallGroupCreateList struct {
	CallGroups [][]int64 `json:"call_groups"`
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
	if conn == nil {
		return nil, ErroNoConn
	}
	flds := callGroupConditionFlds
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
		err = rows.Scan(&data.ID, &data.Name, &data.Description, &data.Enterprise, &data.IsEnable, &data.IsDelete,
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
	if conn == nil {
		return 0, ErroNoConn
	}
	if query == nil || len(query.ID) == 0 {
		return 0, ErrNeedCondition
	}
	flds := []string{
		fldName,
		fldDescription,
		fldIsEnable,
	}
	return updateSQL(conn, query, data, tblCallGroupCondition, flds)
}

//SoftDeleteCondition simply set the is_delete of CallGroupCondition to 1
func (*CallGroupSQLDao) SoftDeleteCondition(conn SqlLike, query *GeneralQuery) (int64, error) {
	if conn == nil {
		return 0, ErroNoConn
	}
	if query == nil || len(query.ID) == 0 {
		return 0, ErrNeedCondition
	}
	return softDelete(conn, query, tblCallGroupCondition)
}

// CreateCallGroups create relation between CallGroups and Calls
func (*CallGroupSQLDao) CreateCallGroups(model *CallGroupCondition) error {
	return nil
}

// CGCondition defines the condition data structure used to group calls
type CGCondition struct {
	CallGroupCondition
	FilterBy map[string]string
	GroupBy  []*groupByUserKey
}

type groupByUserKey struct {
	Inputname string
	ID        int64
}

//Query parameter
const (
	GlobalEnterprise = "global"
)

//CallGroupConditionKey type value
const (
	CGCKeyGroupByKey       = 1
	CGCKeyFilterByKeyValue = 2
)

// CGConditionQuery defines the query parameters to get CGCondition
type CGConditionQuery struct {
	Enterprise      *string
	IsEnable        *int
	IsDelete        *int
	UserKeyIsDelete *int
}

func (q *CGConditionQuery) whereSQL() (condition string, bindData []interface{}, err error) {
	conds := []string{}
	if q.Enterprise != nil {
		cond := fmt.Sprintf("c.%s = ?", fldEnterprise)
		conds = append(conds, cond)
		bindData = append(bindData, *q.Enterprise)
	}
	if q.IsEnable != nil {
		cond := fmt.Sprintf("c.%s = ?", fldIsEnable)
		conds = append(conds, cond)
		bindData = append(bindData, *q.IsEnable)
	}
	if q.IsDelete != nil {
		cond := fmt.Sprintf("c.%s = ?", fldIsDelete)
		conds = append(conds, cond)
		bindData = append(bindData, *q.IsDelete)
	}
	if q.UserKeyIsDelete != nil {
		cond := fmt.Sprintf("uk.%s = ?", fldIsDelete)
		conds = append(conds, cond)
		bindData = append(bindData, *q.UserKeyIsDelete)
	}
	condition = "WHERE " + strings.Join(conds, " AND ")
	return
}

//GetCGCondition return requested CGCondition list
func (*CallGroupSQLDao) GetCGCondition(conn SqlLike, query *CGConditionQuery) ([]*CGCondition, error) {
	if conn == nil {
		return nil, ErroNoConn
	}
	whereSQL, params, err := query.whereSQL()
	if err != nil {
		return nil, ErrGenCondition
	}
	querySQL := fmt.Sprintf(
		`SELECT c.*, ck.type, uk.id, uk.inputname
		FROM %s c
		LEFT JOIN %s ck
		ON c.id = ck.cg_condition_id
		LEFT JOIN %s uk
		ON ck.group_by = uk.id
		%s`,
		tblCallGroupCondition, tblCallGroupConditionKey, tblUserKey, whereSQL)
	rows, err := conn.Query(querySQL, params...)
	if err != nil {
		logger.Error.Printf("query failed. %s %+v\n", querySQL, params)
		return nil, err
	}
	defer rows.Close()

	condMap := make(map[int64]*CGCondition)
	var conds = []*CGCondition{}
	for rows.Next() {
		var data CGCondition
		var inputname string
		var groupByType int
		var userKeyID int64
		var dayRange, durationMin, durationMax sql.NullInt64
		err = rows.Scan(
			&data.ID, &data.Name, &data.Description, &data.Enterprise, &data.IsEnable,
			&data.IsDelete, &dayRange, &durationMin, &durationMax, &data.CreateTime,
			&data.UpdateTime, &groupByType, &userKeyID, &inputname)
		if err != nil {
			logger.Error.Printf("scan failed. %s\n", err)
			return nil, err
		}
		data.DayRange = int(dayRange.Int64)
		data.DurationMin = int(durationMin.Int64)
		data.DurationMax = int(durationMax.Int64)
		cond, ok := condMap[data.ID]
		if !ok {
			condMap[data.ID] = &data
			cond = &data
			cond.FilterBy = make(map[string]string)
			conds = append(conds, cond)
		}
		if groupByType == CGCKeyGroupByKey {
			cond.GroupBy = append(cond.GroupBy, &groupByUserKey{
				Inputname: inputname,
				ID:        userKeyID,
			})
		}
		if groupByType == CGCKeyFilterByKeyValue {
			cond.FilterBy[inputname] = ""
		}
	}

	querySQL = fmt.Sprintf(
		`SELECT uk.id, uk.inputname, uv.link_id, uv.value
		FROM %s uk
		LEFT JOIN %s uv
		ON uk.id = uv.userkey_id
		WHERE uk.is_delete = ? AND uv.is_delete = ? AND uv.type = ?`,
		tblUserKey, tblUserValue)
	params = []interface{}{0, 0, UserValueTypCallGroupCondition}
	valueRows, err := conn.Query(querySQL, params...)
	if err != nil {
		logger.Error.Printf("query failed. %s %+v\n", querySQL, params)
		return nil, err
	}
	defer valueRows.Close()

	for valueRows.Next() {
		var userKeyID, condID int64
		var inputname, value string
		err = valueRows.Scan(&userKeyID, &inputname, &condID, &value)
		if err != nil {
			logger.Error.Printf("scan failed. %s\n", err)
			return nil, err
		}
		if cond, ok := condMap[condID]; ok {
			if _, ok := cond.FilterBy[inputname]; ok {
				cond.FilterBy[inputname] = value
			}
		}
	}
	return conds, nil
}

// CallsToGroupQuery defines the query parameters to get calls to gorup
type CallsToGroupQuery struct {
	UserValueType *int8
	UserKeyID     *int64
	UserValue     *[]string
	StartTime     *int64
	EndTime       *int64
}

func (q *CallsToGroupQuery) whereSQL() (condition string, bindData []interface{}, err error) {
	conds := []string{}
	if q.UserValueType != nil {
		cond := fmt.Sprintf("`%s`.%s = ?", tblUserValue, fldUserValueType)
		conds = append(conds, cond)
		bindData = append(bindData, *q.UserValueType)
	}
	if q.UserKeyID != nil {
		cond := fmt.Sprintf("`%s`.%s = ?", tblUserValue, fldUserValueUserKey)
		conds = append(conds, cond)
		bindData = append(bindData, *q.UserKeyID)
	}
	if q.UserValue != nil && len(*q.UserValue) > 0 {
		cond := fmt.Sprintf("`%s`.%s IN %s", tblUserValue, fldUserValueVal, "(?"+strings.Repeat(",?", len(*q.UserValue)-1)+")")
		conds = append(conds, cond)
		for _, value := range *q.UserValue {
			bindData = append(bindData, value)
		}
	}
	if q.StartTime != nil && q.EndTime != nil {
		cond := fmt.Sprintf("`%s`.%s BETWEEN ? AND ?", tblCall, fldCallUploadTime)
		conds = append(conds, cond)
		bindData = append(bindData, *q.StartTime, *q.EndTime)
	}
	condition = "WHERE " + strings.Join(conds, " AND ")
	return
}

//GetCallIDsToGroup return call ID list to group
func (*CallGroupSQLDao) GetCallIDsToGroup(conn SqlLike, query *CallsToGroupQuery) ([]int64, error) {
	if conn == nil {
		return nil, ErroNoConn
	}
	whereSQL, params, err := query.whereSQL()
	if err != nil {
		return nil, ErrGenCondition
	}
	querySQL := fmt.Sprintf(
		"SELECT `%s`.link_id FROM `%s` LEFT JOIN `%s` ON `%s`.link_id = `%s`.call_id %s",
		tblUserValue, tblUserValue, tblCall, tblUserValue, tblCall, whereSQL)
	rows, err := conn.Query(querySQL, params...)
	if err != nil {
		logger.Error.Printf("query failed. %s %+v\n", querySQL, params)
		return nil, err
	}
	defer rows.Close()

	var callIDs []int64
	for rows.Next() {
		var id int64
		err = rows.Scan(&id)
		if err != nil {
			logger.Error.Printf("scan failed. %s\n", err)
			return nil, err
		}
		callIDs = append(callIDs, id)
	}
	return callIDs, nil
}

// CallGroup defines the CallGroup model
type CallGroup struct {
	ID                   int64
	IsDelete             int
	CallGroupConditionID int64
	Enterprise           string
	FirstCallID          int64
	CreateTime           int64
	UpdateTime           int64
	Calls                []int64
}

// CallGroupQuery defines the query parameters to get CallGroup list
type CallGroupQuery struct {
	IDs        *[]int64
	Enterprise *string
	IsDelete   *int
	CallIDs    *[]int64
	StartTime  *int64
	EndTime    *int64
}

func (q *CallGroupQuery) whereSQL() (condition string, bindData []interface{}, err error) {
	conds := []string{}
	if q.IDs != nil {
		cond := fmt.Sprintf("cg.%s IN %s", fldID, "(?"+strings.Repeat(",?", len(*q.IDs)-1)+")")
		conds = append(conds, cond)
		for _, id := range *q.IDs {
			bindData = append(bindData, id)
		}
	}
	if q.Enterprise != nil {
		cond := fmt.Sprintf("cg.%s = ?", fldEnterprise)
		conds = append(conds, cond)
		bindData = append(bindData, *q.Enterprise)
	}
	if q.IsDelete != nil {
		cond := fmt.Sprintf("cg.%s = ?", fldIsDelete)
		conds = append(conds, cond)
		bindData = append(bindData, *q.IsDelete)
	}
	if q.CallIDs != nil && len(*q.CallIDs) > 0 {
		cond := fmt.Sprintf("relcg.%s IN (SELECT %s FROM %s WHERE %s IN %s)",
			fldCallGroupID, fldCallGroupID, tblRelCallGroupCall, fldCallID, "(?"+strings.Repeat(",?", len(*q.CallIDs)-1)+")")
		conds = append(conds, cond)
		for _, id := range *q.CallIDs {
			bindData = append(bindData, id)
		}
	}
	if q.StartTime != nil && q.EndTime != nil {
		cond := fmt.Sprintf("cg.%s BETWEEN ? AND ?", fldCreateTime)
		conds = append(conds, cond)
		bindData = append(bindData, *q.StartTime, *q.EndTime)
	}
	condition = "WHERE " + strings.Join(conds, " AND ")
	return
}

// GetCallGroups return the queried CallGroup list
func (*CallGroupSQLDao) GetCallGroups(conn SqlLike, query *CallGroupQuery) ([]*CallGroup, error) {
	if conn == nil {
		return nil, ErroNoConn
	}
	whereSQL, params, err := query.whereSQL()
	if err != nil {
		return nil, ErrGenCondition
	}
	querySQL := fmt.Sprintf(
		`SELECT cg.*, relcg.call_id
		FROM %s as cg
		LEFT JOIN %s as relcg
		ON cg.id = relcg.cg_id
		%s`,
		tblCallGroup, tblRelCallGroupCall, whereSQL)
	rows, err := conn.Query(querySQL, params...)
	if err != nil {
		logger.Error.Printf("query failed. %s %+v\n", querySQL, params)
		return nil, err
	}
	defer rows.Close()

	callGroupMap := make(map[int64]*CallGroup)
	for rows.Next() {
		var data CallGroup
		var callID int64
		err = rows.Scan(&data.ID, &data.IsDelete, &data.CallGroupConditionID, &data.Enterprise,
			&data.FirstCallID, &data.CreateTime, &data.UpdateTime, &callID)
		if err != nil {
			logger.Error.Printf("scan failed. %s\n", err)
			return nil, err
		}
		callGroup, ok := callGroupMap[data.ID]
		if !ok {
			callGroupMap[data.ID] = &data
			callGroup = &data
		}
		callGroup.Calls = append(callGroup.Calls, callID)
	}
	callGroupList := make([]*CallGroup, 0)
	for _, group := range callGroupMap {
		callGroupList = append(callGroupList, group)
	}
	return callGroupList, nil
}

// SoftDeleteCallGroup delete requested CallGroup
func (*CallGroupSQLDao) SoftDeleteCallGroup(conn SqlLike, query *GeneralQuery) error {
	if conn == nil {
		return ErroNoConn
	}
	if len(query.ID) == 0 {
		return nil
	}
	whereSQL, params, err := query.whereSQL()
	if err != nil {
		return ErrGenCondition
	}
	querySQL := fmt.Sprintf("UPDATE `%s` SET `%s` = 1 %s", tblCallGroup, fldIsDelete, whereSQL)
	_, err = conn.Exec(querySQL, params...)
	if err != nil {
		logger.Error.Printf("sql execution failed. %s %+v\n", querySQL, params)
		return err
	}

	params = []interface{}{}
	for _, id := range query.ID {
		params = append(params, id)
	}
	querySQL = fmt.Sprintf("DELETE FROM `%s` WHERE `%s` IN %s",
		tblRelCallGroupCall, fldCallGroupID, "(?"+strings.Repeat(",?", len(query.ID)-1)+")")
	_, err = conn.Exec(querySQL, params...)
	if err != nil {
		logger.Error.Printf("sql execution failed. %s %+v\n", querySQL, params)
		return err
	}
	return nil
}

const (
	fldCallGroupConditionID = "cg_cond_id"
	fldFirstCallID          = "first_call_id"
	fldCallGroupID          = "cg_id"
)

// error message
var (
	ErroNoCalls = errors.New("no calls to group")
)

//CreateCallGroup careate a new CallGroup record
func (*CallGroupSQLDao) CreateCallGroup(conn SqlLike, model *CallGroup) (int64, error) {
	if conn == nil {
		return 0, ErroNoConn
	}
	if model == nil {
		return 0, ErrNeedRequest
	}
	if len(model.Calls) == 0 {
		return 0, ErroNoCalls
	}
	insertCols := []string{
		fldIsDelete, fldCallGroupConditionID, fldEnterprise, fldFirstCallID, fldCreateTime,
		fldUpdateTime,
	}
	params := []interface{}{
		model.IsDelete, model.CallGroupConditionID, model.Enterprise, model.FirstCallID, model.CreateTime,
		model.UpdateTime,
	}
	querySQL := fmt.Sprintf(
		"INSERT INTO `%s` (`%s`) VALUES (%s)",
		tblCallGroup,
		strings.Join(insertCols, "`, `"),
		"?"+strings.Repeat(",?", len(insertCols)-1),
	)

	result, err := conn.Exec(querySQL, params...)
	if err != nil {
		logger.Error.Printf("sql execution failed. %s %+v\n", querySQL, params)
		return 0, err
	}
	model.ID, err = result.LastInsertId()
	if err != nil {
		return 0, ErrAutoIDDisabled
	}
	return model.ID, nil
}

// CallGroupRelation defines the Relation_CallGroup_Call model
type CallGroupRelation struct {
	ID          int64
	CallGroupID int64
	CallID      int64
}

//CreateCallGroupRelation careate a new Relation_CallGroup_Call record
func (*CallGroupSQLDao) CreateCallGroupRelation(conn SqlLike, model *CallGroupRelation) (int64, error) {
	if conn == nil {
		return 0, ErroNoConn
	}
	if model == nil {
		return 0, ErrNeedRequest
	}
	insertCols := []string{
		fldCallGroupID, fldCallID,
	}
	params := []interface{}{
		model.CallGroupID, model.CallID,
	}
	querySQL := fmt.Sprintf(
		"INSERT INTO `%s` (`%s`) VALUES (%s)",
		tblRelCallGroupCall,
		strings.Join(insertCols, "`, `"),
		"?"+strings.Repeat(",?", len(insertCols)-1),
	)

	result, err := conn.Exec(querySQL, params...)
	if err != nil {
		logger.Error.Printf("sql execution failed. %s %+v\n", querySQL, params)
		return 0, err
	}
	model.ID, err = result.LastInsertId()
	if err != nil {
		return 0, ErrAutoIDDisabled
	}
	return model.ID, nil
}
