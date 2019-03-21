package model

import (
	"fmt"
	"strings"
	"time"

	"bufio"
	"bytes"
	"reflect"
	"strconv"

	"emotibot.com/emotigo/pkg/logger"
	"github.com/kataras/iris/core/errors"
	"github.com/tealeg/xlsx"
)

// TODO: Refractor Group & GroupWCond; Condition & GroupCondition
type GroupDAO interface {
	CountGroupsBy(filter *GroupFilter, sqlLike SqlLike) (int64, error)
	CreateGroup(group *GroupWCond, sqlLike SqlLike) (*GroupWCond, error)
	Group(delegatee SqlLike, query GroupQuery) ([]Group, error)
	DeleteGroup(id string, sqlLike SqlLike) error
	GetGroupsBy(filter *GroupFilter, sqlLike SqlLike) ([]GroupWCond, error)
	GetGroupsByRuleID(id []int64, sqlLike SqlLike) ([]GroupWCond, error)
	GroupsByCalls(delegatee SqlLike, query CallQuery) (map[int64][]Group, error)
	CreateMany([]GroupWCond, SqlLike) error
	DeleteMany([]string, SqlLike) error
	ExportGroups(sqlLike SqlLike) (*bytes.Buffer, error)
	ImportGroups(sqlLike SqlLike, fileName string) error
}

type GroupSQLDao struct {
	conn SqlLike
}

func NewGroupSQLDao(db DBLike) *GroupSQLDao {
	return &GroupSQLDao{
		conn: db.Conn(),
	}
}

type SimpleGroup struct {
	ID   string `json:"group_id"`
	Name string `json:"group_name"`
}

// GroupWCond is Group with Condition struct
type GroupWCond struct {
	ID              int64               `json:"-"`
	Deleted         int8                `json:"-"`
	Name            *string             `json:"group_name,omitempty"`
	Enterprise      string              `json:",omitempty"`
	Description     *string             `json:"description"`
	CreateTime      int64               `json:"create_time,omitempty"`
	Enabled         *int8               `json:"is_enable,omitempty"`
	Speed           *float64            `json:"limit_speed,omitempty"`
	SlienceDuration *float64            `json:"limit_silence,omitempty"`
	UUID            string              `json:"group_id,omitempty"`
	Rules           *[]ConversationRule `json:"rules"`
	Condition       *GroupCondition     `json:"other,omitempty"`
	RuleCount       int                 `json:"rule_count"`
}

type GroupFilter struct {
	FileName      string
	Deal          *int
	Series        string
	CallStart     int64
	CallEnd       int64
	StaffID       string
	StaffName     string
	Extension     string
	Department    string
	CustomerID    string
	CustomerName  string
	CustomerPhone string
	EnterpriseID  string
	Page          int
	Limit         int
	UUID          []string
	ID            []uint64
	Delete        *int8
	Rules         []string
}

// Group is the one to one represent of rule group table schema
// Rules & Condition & CustomConditions is the virtual column that indicate relations.
type Group struct {
	ID               int64
	IsDelete         bool
	Name             string
	EnterpriseID     string
	Description      string
	CreatedTime      int64
	UpdatedTime      int64
	IsEnable         bool
	LimitedSpeed     int
	LimitedSilence   float32
	Typ              int8
	UUID             string
	Rules            []ConversationRule
	Condition        *Condition
	CustomConditions []UserValue
}

type GroupCondition struct {
	FileName         *string `json:"file_name"`
	CallDuration     *int64  `json:"call_time"`
	CallComment      *string `json:"call_comment"`
	Deal             *int    `json:"transcation"`
	Series           *string `json:"series"`
	StaffID          *string `json:"host_id"`
	StaffName        *string `json:"host_name"`
	Extension        *string `json:"extension"`
	Department       *string `json:"department"`
	ClientID         *string `json:"guest_id"`
	ClientName       *string `json:"guest_name"`
	ClientPhone      *string `json:"guest_phone"`
	LeftChannel      *string `json:"left_channel"`
	LeftChannelCode  *int    `json:"-"`
	RightChannel     *string `json:"right_channel"`
	RightChannelCode *int    `json:"-"`
	CallStart        *int64  `json:"call_from"`
	CallEnd          *int64  `json:"call_end"`
}

// getGroupsSQL generate a complex sql join for each rows contain the group with its condition and rules table.
//
// **Users should create their own select part.**
func getGroupsSQL(filter *GroupFilter) (queryStr string, values []interface{}) {
	values = []interface{}{}
	groupStr := ""
	groupConditions := []string{}
	if len(filter.UUID) > 0 {
		groupStr = fmt.Sprintf("%s IN (?%s)", fldUUID, strings.Repeat(", ?", len(filter.UUID)-1))
		groupConditions = append(groupConditions, groupStr)
		for _, id := range filter.UUID {
			values = append(values, id)
		}
	}

	if filter.EnterpriseID != "" {
		groupStr = fmt.Sprintf("%s = ?", fldRuleGrpEnterpriseID)
		groupConditions = append(groupConditions, groupStr)
		values = append(values, filter.EnterpriseID)
	}

	if len(filter.ID) > 0 {
		groupStr = fmt.Sprintf("%s IN (?%s)", fldID, strings.Repeat(", ?", len(filter.ID)-1))
		groupConditions = append(groupConditions, groupStr)
		for _, id := range filter.ID {
			values = append(values, id)
		}
	}

	if filter.Delete != nil {
		groupStr = fmt.Sprintf("%s = ?", fldIsDelete)
		groupConditions = append(groupConditions, groupStr)
		values = append(values, *filter.Delete)
	}

	if len(groupConditions) > 0 {
		groupStr = fmt.Sprintf("%s %s", "WHERE", strings.Join(groupConditions, " and "))
	} else {
		groupStr = ""
	}

	conditions := []string{}
	conditionStr := "WHERE"
	if filter.FileName != "" {
		conditions = append(conditions, fmt.Sprintf("%s = ?", fldCondFileName))
		values = append(values, filter.FileName)
	}

	if filter.CallEnd != 0 {
		conditions = append(conditions, fmt.Sprintf("%s = ?", fldCondCallEnd))
		values = append(values, filter.CallEnd)
	}

	if filter.CallStart != 0 {
		conditions = append(conditions, fmt.Sprintf("%s = ?", fldCondCallStart))
		values = append(values, filter.CallStart)
	}

	if filter.CustomerID != "" {
		conditions = append(conditions, fmt.Sprintf("%s = ?", fldCondCustomerID))
		values = append(values, filter.CustomerID)
	}

	if filter.CustomerName != "" {
		conditions = append(conditions, fmt.Sprintf("%s = ?", fldCondCustomerName))
		values = append(values, filter.CustomerName)
	}

	if filter.CustomerPhone != "" {
		conditions = append(conditions, fmt.Sprintf("%s = ?", fldCondCustomerPhone))
		values = append(values, filter.CustomerPhone)
	}

	if filter.Deal != nil {
		conditions = append(conditions, fmt.Sprintf("%s = ?", fldCondDeal))
		values = append(values, *filter.Deal)
	}

	if filter.Department != "" {
		conditions = append(conditions, fmt.Sprintf("%s = ?", fldCondDepartment))
		values = append(values, filter.Department)
	}

	if filter.Extension != "" {
		conditions = append(conditions, fmt.Sprintf("%s = ?", fldCondExtension))
		values = append(values, filter.Extension)
	}

	if filter.Series != "" {
		conditions = append(conditions, fmt.Sprintf("%s = ?", fldCondSeries))
		values = append(values, filter.Series)
	}

	if filter.StaffID != "" {
		conditions = append(conditions, fmt.Sprintf("%s = ?", fldCondStaffID))
		values = append(values, filter.StaffID)
	}

	if filter.StaffName != "" {
		conditions = append(conditions, fmt.Sprintf("%s = ?", fldCondStaffName))
		values = append(values, filter.StaffName)
	}

	if len(conditions) == 0 {
		conditionStr = ""
	} else {
		conditionStr = fmt.Sprintf("%s %s", conditionStr, strings.Join(conditions, " and "))
	}

	ruleCondition := fmt.Sprintf("LEFT JOIN %s", tblRule)
	if len(filter.Rules) > 0 {
		ruleCondition = fmt.Sprintf(
			"INNER JOIN (SELECT * FROM %s WHERE %s IN (%s))",
			tblRule,
			fldUUID,
			fmt.Sprintf("?%s", strings.Repeat(", ?", len(filter.Rules)-1)),
		)
		for _, ruleID := range filter.Rules {
			values = append(values, ruleID)
		}
	}
	queryStr = " FROM (SELECT * FROM `%s` %s) as rg" +
		" LEFT JOIN (SELECT * FROM `%s` %s) as gc on rg.`%s` = gc.`%s`" + // gc group condition table
		" LEFT JOIN  `%s` as rrr ON rg.`%s` = rrr.`%s`" + // rrr Group_Rule relation table
		" %s as rule on rrr.`%s` = rule.`%s`"

	queryStr = fmt.Sprintf(queryStr,
		tblRuleGroup, groupStr,
		tblRGC, conditionStr, fldID, fldCondGroupID,
		tblRelGrpRule, fldID, RRRGroupID,
		ruleCondition, RRRRuleID, fldID,
	)
	return
}

func (s *GroupSQLDao) CountGroupsBy(filter *GroupFilter, sqlLike SqlLike) (total int64, err error) {
	queryStr, values := getGroupsSQL(filter)
	queryStr = fmt.Sprintf("SELECT count(DISTINCT rg.`%s`) %s", fldRuleGrpID, queryStr)
	err = sqlLike.QueryRow(queryStr, values...).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("query row failed, %s", err.Error())
	}
	return total, nil
}

func (s *GroupSQLDao) GetGroupsBy(filter *GroupFilter, sqlLike SqlLike) (groups []GroupWCond, err error) {
	queryStr, values := getGroupsSQL(filter)
	if filter.Limit > 0 {
		start := filter.Page * filter.Limit
		queryStr = fmt.Sprintf("%s LIMIT %d, %d", queryStr, start, filter.Limit)
	}
	grpCols := []string{
		fldRuleGrpID, fldRuleGrpUUID, fldRuleGrpName,
		fldDescription, fldRuleGrpLimitSpeed, fldRuleGrpLimitSilence,
		fldCreateTime, fldRuleGrpIsEnable, fldEnterprise,
		fldIsDelete,
	}
	condCols := []string{
		fldCondFileName, fldCondDeal, fldCondSeries,
		fldCondStaffID, fldCondStaffName, fldCondExtension,
		fldCondDepartment, fldCondCustomerID, fldCondCustomerName,
		fldCondCustomerPhone, fldCondCallStart, fldCondCallEnd,
		fldCondLeftChanRole, fldCondRightChanRole,
	}
	ruleCols := []string{
		fldID, fldUUID, fldName,
		fldRuleMethod, fldRuleScore, fldEnterprise,
		fldRuleDescription, fldRuleMin, fldRuleMax,
		fldRuleSeverity, fldRuleCreateTime, fldRuleUpdateTime,
		fldIsDelete,
	}
	queryStr = fmt.Sprintf("SELECT rg.`%s`, gc.`%s`, rule.`%s` %s",
		strings.Join(grpCols, "`, rg.`"), strings.Join(condCols, "`, gc.`"), strings.Join(ruleCols, "`, rule.`"), queryStr)
	rows, err := sqlLike.Query(queryStr, values...)
	if err != nil {
		err = fmt.Errorf("error while get groups in dao.GetGroupsBy, err: %s", err.Error())
		return
	}
	defer rows.Close()

	groups = []GroupWCond{}
	var currentGroup *GroupWCond
	for rows.Next() {
		group := GroupWCond{}
		condition := GroupCondition{}
		var (
			rUUID, rName, enterprise, description *string // rule uuid
			rID, rCreatedTime, rUpdatedTime       *int64
			rMethod, severity, deleted            *int8
			score, min, max                       *int
		)
		err = rows.Scan(
			&group.ID,
			&group.UUID,
			&group.Name,
			&group.Description,
			&group.Speed,
			&group.SlienceDuration,
			&group.CreateTime,
			&group.Enabled,
			&group.Enterprise,
			&group.Deleted,
			&condition.FileName,
			&condition.Deal,
			&condition.Series,
			&condition.StaffID,
			&condition.StaffName,
			&condition.Extension,
			&condition.Department,
			&condition.ClientID,
			&condition.ClientName,
			&condition.ClientPhone,
			&condition.CallStart,
			&condition.CallEnd,
			&condition.LeftChannelCode,
			&condition.RightChannelCode,
			&rID, &rUUID, &rName,
			&rMethod, &score, &enterprise,
			&description, &min, &max,
			&severity, &rCreatedTime, &rUpdatedTime,
			&deleted,
		)

		if currentGroup == nil || group.ID != currentGroup.ID {
			if currentGroup != nil {
				groups = append(groups, *currentGroup)
			}

			group.Condition = &condition

			currentGroup = &group
			rules := []ConversationRule{}
			currentGroup.Rules = &rules
		}

		if rUUID != nil && rName != nil {
			rule := ConversationRule{
				ID:          *rID,
				UUID:        *rUUID,
				Name:        *rName,
				Method:      *rMethod,
				Score:       *score,
				Enterprise:  *enterprise,
				Description: *description,
				Min:         *min,
				Max:         *max,
				Severity:    *severity,
				CreateTime:  *rCreatedTime,
				UpdateTime:  *rUpdatedTime,
				Deleted:     *deleted,
			}
			rules := append(*currentGroup.Rules, rule)
			currentGroup.Rules = &rules
		}
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("")
	}
	if currentGroup != nil {
		groups = append(groups, *currentGroup)
	}
	return
}

func (s *GroupSQLDao) GetGroupsByRuleID(id []int64, sqlLike SqlLike) (groups []GroupWCond, err error) {
	groups = []GroupWCond{}
	if len(id) == 0 {
		return
	}

	builder := NewWhereBuilder(andLogic, "")
	builder.In(RRRRuleID, int64ToWildCard(id...))
	conditionStr, values := builder.Parse()

	queryStr := fmt.Sprintf(
		"SELECT %s FROM %s WHERE %s",
		RRRGroupID,
		tblRelGrpRule,
		conditionStr,
	)

	rows, err := sqlLike.Query(queryStr, values...)
	if err != nil {
		err = fmt.Errorf("error while query group id in dao.GetGroupsByRuleID, err: %s", err.Error())
		return
	}
	defer rows.Close()

	groupID := []uint64{}
	for rows.Next() {
		var gid uint64
		err = rows.Scan(&gid)
		if err != nil {
			err = fmt.Errorf("error while scan group id in dao.GetGroupsByRuleID, err: %s", err.Error())
			return
		}
		groupID = append(groupID, gid)
	}

	if len(groupID) == 0 {
		return
	}

	filter := &GroupFilter{
		ID: groupID,
	}

	groups, err = s.GetGroupsBy(filter, sqlLike)
	if err != nil {
		err = fmt.Errorf("error while query groups in dao.GetGroupsByRuleID, err: %s", err.Error())
	}
	return
}

func genInsertRelationSQL(groups []GroupWCond) (insertStr string, values []interface{}) {
	if len(groups) == 0 {
		return
	}

	insertStr = fmt.Sprintf(
		"INSERT INTO %s (%s, %s) VALUES",
		tblRelGrpRule,
		RRRGroupID,
		RRRRuleID,
	)

	variableStr := "(?, ?)"
	valueStr := ""
	values = []interface{}{}
	for _, group := range groups {
		for _, rule := range *group.Rules {
			values = append(values, group.ID, rule.ID)
			valueStr = fmt.Sprintf("%s%s,", valueStr, variableStr)
		}
	}

	if len(values) > 0 {
		// remove last comma
		valueStr = valueStr[:len(valueStr)-1]
	}
	insertStr = fmt.Sprintf("%s %s", insertStr, valueStr)

	return
}

func genInsertGroupSQL(groups []GroupWCond) (insertStr string, values []interface{}) {
	if len(groups) == 0 {
		return
	}
	fields := []string{
		fldRuleGrpUUID,
		fldRuleGrpName,
		fldRuleGrpEnterpriseID,
		fldDescription,
		fldRuleGrpCreateTime,
		fldRuleGrpUpdateTime,
		fldRuleGrpIsEnable,
		fldRuleGrpLimitSpeed,
		fldRuleGrpLimitSilence,
	}
	insertStr = fmt.Sprintf(
		"INSERT INTO `%s` (%s) VALUES",
		tblRuleGroup,
		strings.Join(fields, ", "),
	)

	variableStr := fmt.Sprintf(
		"(?%s)",
		strings.Repeat(", ?", len(fields)-1),
	)

	valueStr := ""
	values = []interface{}{}
	now := time.Now().Unix()
	for _, group := range groups {
		valueStr = fmt.Sprintf("%s%s,", valueStr, variableStr)
		values = append(
			values,
			group.UUID,
			group.Name,
			group.Enterprise,
			group.Description,
			now,
			now,
			group.Enabled,
			group.Speed,
			group.SlienceDuration,
		)
	}
	// remove last comma
	valueStr = valueStr[:len(valueStr)-1]
	insertStr = fmt.Sprintf("%s %s", insertStr, valueStr)

	return
}

func genInsertConditionSQL(groups []GroupWCond) (insertStr string, values []interface{}) {
	if len(groups) == 0 {
		return
	}

	fields := []string{
		fldCondGroupID,
		fldCondFileName,
		fldCondDeal,
		fldCondSeries,
		fldCondStaffID,
		fldCondStaffName,
		fldCondExtension,
		fldCondDepartment,
		fldCondCustomerID,
		fldCondCustomerName,
		fldCondCustomerPhone,
		fldCondCallStart,
		fldCondCallEnd,
		fldCondLeftChanRole,
		fldCondRightChanRole,
	}

	insertStr = fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES",
		tblRGC,
		strings.Join(fields, ", "),
	)

	variableStr := fmt.Sprintf(
		"(?%s)",
		strings.Repeat(", ?", len(fields)-1),
	)

	valueStr := ""
	values = []interface{}{}
	for _, group := range groups {
		values = append(
			values,
			group.ID,
			group.Condition.FileName,
			group.Condition.Deal,
			group.Condition.Series,
			group.Condition.StaffID,
			group.Condition.StaffName,
			group.Condition.Extension,
			group.Condition.Department,
			group.Condition.ClientID,
			group.Condition.ClientName,
			group.Condition.ClientPhone,
			group.Condition.CallStart,
			group.Condition.CallEnd,
			group.Condition.LeftChannelCode,
			group.Condition.RightChannelCode,
		)
		valueStr = fmt.Sprintf("%s%s,", valueStr, variableStr)
	}
	// remove last comma
	valueStr = valueStr[:len(valueStr)-1]
	insertStr = fmt.Sprintf("%s %s", insertStr, valueStr)

	return
}

func (s *GroupSQLDao) CreateGroup(group *GroupWCond, sqlLike SqlLike) (createdGroup *GroupWCond, err error) {
	// insert group
	insertStr, values := genInsertGroupSQL([]GroupWCond{*group})

	result, err := sqlLike.Exec(insertStr, values...)
	if err != nil {
		logger.Error.Printf("error while insert group in dao.CreateGroup, sql: %s\n", insertStr)
		logger.Error.Printf("error while insert group in dao.CreateGroup, values: %+v\n", values)
		err = fmt.Errorf("error while insert group in dao.CreateGroup, err: %s", err.Error())
		return
	}

	groupID, err := result.LastInsertId()
	group.ID = groupID
	if err != nil {
		err = fmt.Errorf("error while get group id in dao.CreateGroup, err: %s", err.Error())
		return
	}

	// insert condition
	insertStr, values = genInsertConditionSQL([]GroupWCond{*group})

	_, err = sqlLike.Exec(insertStr, values...)
	if err != nil {
		logger.Error.Printf("error while insert condition in dao.CreateGroup, sql: %s\n", insertStr)
		logger.Error.Printf("error while insert condition in dao.CreateGroup, values: %+v\n", values)
		err = fmt.Errorf("error while insert condition in dao.CreateGroup, err: %s", err.Error())
		return
	}

	// insert into group_rule_map
	if group.Rules != nil && len(*group.Rules) != 0 {
		insertStr, values = genInsertRelationSQL([]GroupWCond{*group})

		_, err = sqlLike.Exec(insertStr, values...)
		if err != nil {
			logger.Error.Printf("error while insert relation_group_rule in dao.CreateGroup, sql: %s\n", insertStr)
			logger.Error.Printf("error while insert relation_group_rule in dao.CreateGroup, values: %+v\n", values)
			err = fmt.Errorf("error while insert relation_group_rule in dao.CreateGroup, err: %s", err.Error())
			return
		}
	}

	createdGroup = group
	return
}

func addCommaIfNotFirst(sqlStr string, first bool) string {
	if !first {
		sqlStr += ","
		return sqlStr
	}
	return sqlStr
}

func (s *GroupSQLDao) DeleteGroup(id string, sqlLike SqlLike) (err error) {
	deleteStr := fmt.Sprintf(
		"UPDATE %s SET %s = 1 WHERE %s = ?",
		tblRuleGroup,
		fldRuleGrpIsDelete,
		fldRuleGrpUUID,
	)
	_, err = sqlLike.Exec(deleteStr, id)
	if err != nil {
		err = fmt.Errorf("error while delete group in dao.DeleteGroup, err: %s", err.Error())
	}
	return
}

var groupCols = []string{
	fldRuleGrpID, fldRuleGrpIsDelete, fldRuleGrpName,
	fldRuleGrpEnterpriseID, fldRuleGrpDescription, fldRuleGrpCreateTime,
	fldRuleGrpUpdateTime, fldRuleGrpIsEnable, fldRuleGrpLimitSpeed,
	fldRuleGrpLimitSilence, fldRuleGrpType, fldRuleGrpUUID,
}

// Group query simple group with the group query.
// No relation will be queried.
func (s *GroupSQLDao) Group(delegatee SqlLike, query GroupQuery) ([]Group, error) {
	if delegatee == nil {
		delegatee = s.conn
	}

	sqlQuery := fmt.Sprintf("SELECT `%s` FROM `%s`", strings.Join(groupCols, "`, `"), tblRuleGroup)
	wherePart, bindData := query.whereSQL()
	if len(bindData) > 0 {
		sqlQuery += " " + wherePart
	}
	rows, err := delegatee.Query(sqlQuery, bindData...)
	if err != nil {
		logger.Error.Println("raw sql: ", sqlQuery)
		logger.Error.Println("raw bind-data: ", bindData)
		return nil, fmt.Errorf("sql executed failed, %v", err)
	}
	defer rows.Close()
	var groups = make([]Group, 0)
	for rows.Next() {

		var g Group
		var isDeleted, isEnabled int
		rows.Scan(
			&g.ID, &isDeleted, &g.Name,
			&g.EnterpriseID, &g.Description, &g.CreatedTime,
			&g.UpdatedTime, &isEnabled, &g.LimitedSpeed,
			&g.LimitedSilence, &g.Typ, &g.UUID,
		)
		if isDeleted == 1 {
			g.IsDelete = true
		}
		if isEnabled == 1 {
			g.IsEnable = true
		}
		groups = append(groups, g)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("sql scan failed, %v", err)
	}

	return groups, nil
}

func (s *GroupSQLDao) GroupsWithCondition(delegatee SqlLike, query GroupQuery) ([]Group, error) {
	if delegatee == nil {
		delegatee = s.conn
	}
	wherePart, bindData := query.whereSQL()
	sqlQuery := fmt.Sprintf("SELECT g.`%s`, c.`%s` "+
		"FROM `%s` as g LEFT JOIN `%s` as c "+
		"ON g.`%s` = c.`%s` %s",
		strings.Join(groupCols, "`, g.`"), strings.Join(conditionCols, "`, c.`"),
		tblRuleGroup, tblRGC,
		fldRuleGrpID, fldCondGroupID, wherePart,
	)
	rows, err := delegatee.Query(sqlQuery, bindData...)
	if err != nil {
		return nil, fmt.Errorf("sql query failed, %v", err)
	}
	defer rows.Close()
	var groups = make([]Group, 0)
	for rows.Next() {
		var (
			g                    Group
			isDeleted, isEnabled int
			c                    Condition
		)
		rows.Scan(
			&g.ID, &isDeleted, &g.Name,
			&g.EnterpriseID, &g.Description, &g.CreatedTime,
			&g.UpdatedTime, &isEnabled, &g.LimitedSpeed,
			&g.LimitedSilence, &g.Typ, &g.UUID,
			&c.ID, &c.GroupID, &c.Type,
			&c.FileName, &c.Deal, &c.Series,
			&c.UploadTimeStart, &c.UploadTimeEnd, &c.StaffID,
			&c.StaffName, &c.Extension, &c.Department,
			&c.CustomerID, &c.CustomerName, &c.CustomerPhone,
			&c.CallStart, &c.CallEnd, &c.LeftChannel,
			&c.RightChannel,
		)
		if isDeleted == 1 {
			g.IsDelete = true
		}
		if isEnabled == 1 {
			g.IsEnable = true
		}
		g.Condition = &c
		groups = append(groups, g)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("sql scan error, %v", err)
	}
	return groups, nil
}

// GroupRelation indicate a relation struct of group.
// columns is the select columns for sql
// joinSQL should be the left join condition to the group table.
// type GroupRelation func(group Group) (columns string, joinSQL string)

//WithCondition return a GroupRelation of Condition.
// func WithCondition() GroupRelation {
// 	return func(group Group) (string, string) {
// 		selectSQL := fmt.Sprintf("c.`%s`", strings.Join(conditionCols, "`, c.`"))
// 		joinSQL := fmt.Sprintf(" LEFT JOIN `%s` as c ON g.`%s` = c.`%s` ", tblRGC, fldRuleGrpID, fldCondGroupID)
// 		return selectSQL, joinSQL
// 	}
// }

// func WithRules() GroupRelation {
// 	return func() (string, string) {

// 		return fmt.Sprintf()
// 	}
// }

// Groups query groups with possible relation tables.
// func (s *GroupSQLDao) Groups(delegatee SqlLike, query GroupQuery, relations ...GroupRelation) ([]Group, error) {
// 	if delegatee == nil {
// 		delegatee = s.conn
// 	}
// 	wherePart, bindData := query.whereSQL()
// 	selectSQL := fmt.Sprintf("SELECT g.`%s` ", strings.Join(groupCols, "`, g.`"))
// 	fromSQL := fmt.Sprintf(" FROM `%s` as g ", tblRuleGroup)
// 	for _, r := range relations {
// 		cols, join := r()
// 		selectSQL += cols
// 		fromSQL += join + " "
// 	}
// 	sqlQuery := selectSQL + " " + fromSQL + " " + wherePart
// 	rows, err := delegatee.Query(sqlQuery, bindData...)
// 	if err != nil {
// 		return nil, fmt.Errorf("sql query failed, %v", err)
// 	}
// 	defer rows.Close()
// 	var groups = make([]Group, 0)
// 	for rows.Next() {
// 		var (
// 			g                    Group
// 			isDeleted, isEnabled int
// 			c                    Condition
// 		)
// 		rows.Scan(
// 			&g.ID, &isDeleted, &g.Name,
// 			&g.EnterpriseID, &g.Description, &g.CreatedTime,
// 			&g.UpdatedTime, &isEnabled, &g.LimitedSpeed,
// 			&g.LimitedSilence, &g.Typ, &g.UUID,
// 			&c.ID, &c.GroupID, &c.Type,
// 			&c.FileName, &c.Deal, &c.Series,
// 			&c.UploadTimeStart, &c.StaffID, &c.StaffName,
// 			&c.Extension, &c.Department, &c.CustomerID,
// 			&c.CustomerName, &c.CustomerPhone, &c.CallStart,
// 			&c.CallEnd, &c.LeftChannel, &c.RightChannel,
// 		)
// 		if isDeleted == 1 {
// 			g.IsDelete = true
// 		}
// 		if isEnabled == 1 {
// 			g.IsEnable = true
// 		}
// 		g.Condition = &c
// 		groups = append(groups, g)
// 	}
// 	if err = rows.Err(); err != nil {
// 		return nil, fmt.Errorf("sql scan error, %v", err)
// 	}
// 	return groups, nil
// }

// NewGroup create a plain group without condition or rules.
func (s *GroupSQLDao) NewGroup(delegatee SqlLike, group Group) (Group, error) {
	if delegatee == nil {
		delegatee = s.conn
	}
	groupCols := []string{
		fldRuleGrpIsDelete, fldRuleGrpName, fldRuleGrpEnterpriseID,
		fldRuleGrpDescription, fldRuleGrpCreateTime, fldRuleGrpUpdateTime,
		fldRuleGrpIsEnable, fldRuleGrpLimitSpeed, fldRuleGrpLimitSilence,
		fldRuleGrpType, fldRuleGrpUUID,
	}
	rawsql := fmt.Sprintf("INSERT INTO `%s` (`%s`) VALUES(?%s)",
		tblRuleGroup, strings.Join(groupCols, "`, `"), strings.Repeat(", ?", len(groupCols)-1),
	)
	var isDelete int8
	if group.IsDelete {
		isDelete = 1
	}
	result, err := delegatee.Exec(rawsql,
		isDelete, group.Name, group.EnterpriseID,
		group.Description, group.CreatedTime, group.UpdatedTime,
		group.IsEnable, group.LimitedSpeed, group.LimitedSilence,
		group.Typ, group.UUID,
	)
	if err != nil {
		return Group{}, fmt.Errorf("sql execute failed, %v", err)
	}
	group.ID, err = result.LastInsertId()
	if err != nil {
		return group, ErrAutoIDDisabled
	}
	return group, nil
}

// SetGroupRules set the group rule relation table with given groupID and rules.
func (s *GroupSQLDao) SetGroupRules(delegatee SqlLike, groupID int64, rules []ConversationRule) ([]int64, error) {

	rawsql := fmt.Sprintf(
		"INSERT INTO `%s` (`%s`, `%s`) VALUES(?, ?)",
		tblRelGrpRule,
		RRRGroupID,
		RRRRuleID,
	)
	stmt, err := delegatee.Prepare(rawsql)
	if err != nil {
		return nil, fmt.Errorf("sql prepare failed, %v", err)
	}
	defer stmt.Close()
	relationIDs := make([]int64, 0)
	for _, r := range rules {
		result, err := stmt.Exec(groupID, r.ID)
		if err != nil {
			return nil, fmt.Errorf("insert group %d with rule %d", groupID, r.ID)
		}
		id, err := result.LastInsertId()
		if err != nil {
			return nil, ErrAutoIDDisabled
		}
		relationIDs = append(relationIDs, id)
	}
	return relationIDs, nil
}

// GroupsByCalls find the groups that asssociated with the call given by the query.
// If success, a map of callID with slice of rule group is returned.
func (s *GroupSQLDao) GroupsByCalls(delegatee SqlLike, query CallQuery) (map[int64][]Group, error) {
	if delegatee == nil {
		delegatee = s.conn
	}
	rawsql := fmt.Sprintf("SELECT c.`%s`, gc.`%s` FROM `%s` AS c INNER JOIN `%s` AS gc ON c.`%s` = gc.`%s` ",
		//Select columns
		fldCallID, fldCRGRelRuleGroupID,
		//FROM table
		tblCall, tblRelCallRuleGrp,
		// ON Condition
		fldCallID, fldCRGRelCallID,
	)
	wherepart, data := query.whereSQL("c")
	if len(data) > 0 {
		rawsql = rawsql + " " + wherepart
	}
	rows, err := delegatee.Query(rawsql, data...)
	if err != nil {
		logger.Error.Println("raw error sql: ", rawsql)
		return nil, fmt.Errorf("query error, %v", err)
	}
	defer rows.Close()
	groupUniqueID := map[int64]struct{}{}
	callRuleGrps := map[int64][]int64{}
	for rows.Next() {
		var (
			callID      int64
			ruleGroupID int64
		)
		err := rows.Scan(&callID, &ruleGroupID)
		if err != nil {
			return nil, fmt.Errorf("scan error, %v", err)
		}
		ruleGrps := callRuleGrps[callID]
		ruleGrps = append(ruleGrps, ruleGroupID)
		callRuleGrps[callID] = ruleGrps
		groupUniqueID[ruleGroupID] = struct{}{}
	}
	gq := GroupQuery{}
	for id := range groupUniqueID {
		gq.ID = append(gq.ID, id)
	}
	// TODO: do not relied on another function.
	// potential bug that can be trigger by time delay between querying relation & group table
	groups, err := s.Group(delegatee, gq)
	if err != nil {
		return nil, fmt.Errorf("query group failed, %v", err)
	}
	ruleGrpDict := map[int64]Group{}
	for _, ruleGrp := range groups {
		ruleGrpDict[ruleGrp.ID] = ruleGrp
	}
	result := map[int64][]Group{}
	for call, groups := range callRuleGrps {
		ruleGrps := []Group{}
		for _, id := range groups {
			grp, found := ruleGrpDict[id]
			if !found {
				return nil, fmt.Errorf("relation table's call %d group %d is incorrect, may have corruption data", call, id)
			}
			ruleGrps = append(ruleGrps, grp)
		}
		result[call] = ruleGrps
	}
	return result, nil
}

func (s *GroupSQLDao) CreateMany(groups []GroupWCond, sqlLike SqlLike) (err error) {
	if len(groups) == 0 {
		return
	}

	insertStr, values := genInsertGroupSQL(groups)
	_, err = sqlLike.Exec(insertStr, values...)
	if err != nil {
		logger.Error.Printf("error while insert groups in dao.CreateMany, sql: %s", insertStr)
		logger.Error.Printf("error while insert groups in dao.CreateMany, values: %s", values)
		err = fmt.Errorf("error while insert groups in dao.CreateMany, err: %s", err.Error())
		return
	}

	groupUUID := make([]string, len(groups))
	for idx := range groups {
		groupUUID[idx] = groups[idx].UUID
	}

	deleted := int8(0)
	filter := &GroupFilter{
		UUID:   groupUUID,
		Delete: &deleted,
	}
	newGroups, err := s.GetGroupsBy(filter, sqlLike)
	if err != nil {
		return
	}

	groupMap := map[string]int64{}
	for _, group := range newGroups {
		groupMap[group.UUID] = group.ID
	}

	for idx := range groups {
		group := &groups[idx]
		group.ID = groupMap[group.UUID]
	}

	insertStr, values = genInsertConditionSQL(groups)
	_, err = sqlLike.Exec(insertStr, values...)
	if err != nil {
		logger.Error.Printf("error while insert conditions in dao.CreateMany, sql: %s", insertStr)
		logger.Error.Printf("error while insert conditions in dao.CreateMany, values: %s", values)
		err = fmt.Errorf("error while insert conditions in dao.CreateMany, err: %s", err.Error())
		return
	}

	insertStr, values = genInsertRelationSQL(groups)
	if len(values) > 0 {
		_, err = sqlLike.Exec(insertStr, values...)
		if err != nil {
			logger.Error.Printf("error while insert group rule reliation in dao.CreateMany, sql: %s", insertStr)
			logger.Error.Printf("error while insert group rule reliation in dao.CreateMany, values: %s", values)
			err = fmt.Errorf("error while insert group rule reliation in dao.CreateMany, err: %s", err.Error())
		}
	}
	return
}

func (s *GroupSQLDao) DeleteMany(groupUUID []string, sqlLike SqlLike) (err error) {
	if len(groupUUID) == 0 {
		return
	}
	builder := NewWhereBuilder(andLogic, "")
	builder.In(fldUUID, stringToWildCard(groupUUID...))
	conditionStr, values := builder.Parse()

	deleteStr := fmt.Sprintf(
		"UPDATE %s SET %s = 1 WHERE %s",
		tblRuleGroup,
		fldIsDelete,
		conditionStr,
	)

	_, err = sqlLike.Exec(deleteStr, values...)
	if err != nil {
		logger.Error.Printf("error while delete groups in dao.DeleteMany, sql: %s\n", deleteStr)
		logger.Error.Printf("error while delete groups in dao.DeleteMany, values: %+v\n", values)
		err = fmt.Errorf("error while delete groups in dao.DeleteMany, err: %s", err.Error())
	}
	return
}

func (s *GroupSQLDao) ExportGroups(delegatee SqlLike) (*bytes.Buffer, error) {

	xlFile := xlsx.NewFile()

	var queryStr string
	var err error

	logger.Trace.Println("export Tag ...")
	queryStr = "SELECT %s" +
		strings.Repeat(", %s", reflect.TypeOf(ExportTag{}).NumField()-1) +
		" FROM %s"
	queryStr = fmt.Sprintf(
		queryStr,
		fldTagID, fldTagIsDeleted, fldTagName, fldTagType, fldTagPosSen, fldTagNegSen, fldTagCreateTime, fldTagUpdateTime, fldTagEnterprise, fldTagUUID,
		tblTags,
	)
	if err = SaveToExcel(xlFile, queryStr, tblTags, delegatee, ExportTag{}); err != nil {
		return nil, err
	}

	logger.Trace.Println("export Sentence ...")
	queryStr = "SELECT %s" +
		strings.Repeat(", %s", reflect.TypeOf(ExportSentence{}).NumField()-1) +
		" FROM %s"
	queryStr = fmt.Sprintf(
		queryStr,
		"id", "is_delete", "name", "enterprise", "uuid", "create_time", "update_time", "category_id",
		tblSentence,
	)
	if err = SaveToExcel(xlFile, queryStr, tblSentence, delegatee, ExportSentence{}); err != nil {
		return nil, err
	}

	logger.Trace.Println("export SentenceGroup ...")
	queryStr = "SELECT %s" +
		strings.Repeat(", %s", reflect.TypeOf(ExportSentenceGroup{}).NumField()-1) +
		" FROM %s"
	queryStr = fmt.Sprintf(
		queryStr,
		"id", "is_delete", "name", "enterprise", "role", "position", "`range`", "uuid", "create_time", "update_time", "optional", "type",
		"SentenceGroup",
	)
	if err = SaveToExcel(xlFile, queryStr, "SentenceGroup", delegatee, ExportSentenceGroup{}); err != nil {
		return nil, err
	}

	logger.Trace.Println("export Rule ...")
	queryStr = "SELECT %s" +
		strings.Repeat(", %s", reflect.TypeOf(ExportRule{}).NumField()-1) +
		" FROM %s"
	queryStr = fmt.Sprintf(
		queryStr,
		"id", "is_delete", "name", "method", "score", "description", "enterprise", "min", "max", "severity", "uuid", "create_time", "update_time",
		tblRule,
	)
	if err = SaveToExcel(xlFile, queryStr, tblRule, delegatee, ExportRule{}); err != nil {
		return nil, err
	}

	logger.Trace.Println("export RuleGroup ...")
	queryStr = "SELECT %s" +
		strings.Repeat(", %s", reflect.TypeOf(ExportRuleGroup{}).NumField()-1) +
		" FROM %s"
	queryStr = fmt.Sprintf(
		queryStr,
		"id", "is_delete", "name", "enterprise", "description", "create_time", "update_time", "is_enable", "limit_speed", "limit_silence", "type", "uuid",
		tblRuleGroup,
	)
	if err = SaveToExcel(xlFile, queryStr, tblRuleGroup, delegatee, ExportRuleGroup{}); err != nil {
		return nil, err
	}

	logger.Trace.Println("export SensitiveWord ...")
	queryStr = "SELECT %s" +
		strings.Repeat(", %s", reflect.TypeOf(ExportSensitiveWord{}).NumField()-1) +
		" FROM %s"
	queryStr = fmt.Sprintf(
		queryStr,
		"id", "uuid", "name", "enterprise", "score", "category_id", "is_delete",
		tblSensitiveWord,
	)
	if err = SaveToExcel(xlFile, queryStr, tblSensitiveWord, delegatee, ExportSensitiveWord{}); err != nil {
		return nil, err
	}

	logger.Trace.Println("export ConversationFlow ...")
	queryStr = "SELECT %s" +
		strings.Repeat(", %s", reflect.TypeOf(ExportConversationFlow{}).NumField()-1) +
		" FROM %s"
	queryStr = fmt.Sprintf(
		queryStr,
		"id", "is_delete", "name", "enterprise", "expression", "uuid", "create_time", "update_time", "min",
		tblConversationflow,
	)
	if err = SaveToExcel(xlFile, queryStr, tblConversationflow, delegatee, ExportConversationFlow{}); err != nil {
		return nil, err
	}

	logger.Trace.Println("export Relation_Sentence_Tag ...")
	queryStr = "SELECT %s" +
		strings.Repeat(", %s", reflect.TypeOf(ExportSentenceTag{}).NumField()-1) +
		" FROM %s"
	queryStr = fmt.Sprintf(
		queryStr,
		"id", "s_id", "tag_id",
		tbleRelSentenceTag,
	)
	if err = SaveToExcel(xlFile, queryStr, "R_Sen_Tag", delegatee, ExportSentenceTag{}); err != nil {
		return nil, err
	}

	logger.Trace.Println("export Relation_SentenceGroup_Sentence ...")
	queryStr = "SELECT %s" +
		strings.Repeat(", %s", reflect.TypeOf(ExportSentenceGroupSentence{}).NumField()-1) +
		" FROM %s"
	queryStr = fmt.Sprintf(
		queryStr,
		"id", "sg_id", "s_id",
		"Relation_SentenceGroup_Sentence",
	)
	if err = SaveToExcel(xlFile, queryStr, "R_SenGrp_Sen", delegatee, ExportSentenceGroupSentence{}); err != nil {
		return nil, err
	}

	logger.Trace.Println("export Relation_SensitiveWord_Sentence ...")
	queryStr = "SELECT %s" +
		strings.Repeat(", %s", reflect.TypeOf(ExportSensitiveWordSentence{}).NumField()-1) +
		" FROM %s"
	queryStr = fmt.Sprintf(
		queryStr,
		"id", "sw_id", "s_id", "type",
		tblRelSensitiveWordSen,
	)
	if err = SaveToExcel(xlFile, queryStr, "R_SensitiveWord_Sen", delegatee, ExportSensitiveWordSentence{}); err != nil {
		return nil, err
	}

	logger.Trace.Println("export RuleGroupCondition ...")
	queryStr = "SELECT %s" +
		strings.Repeat(", %s", reflect.TypeOf(ExportRuleGroupCondition{}).NumField()-1) +
		" FROM %s"
	queryStr = fmt.Sprintf(
		queryStr,
		"id", "rg_id", "type", "file_name", "deal", "series", "upload_time", "staff_id", "staff_name", "extension", "department", "customer_id", "customer_name", "customer_phone", "category", "call_start", "call_end", "left_channel", "right_channel",
		"RuleGroupCondition",
	)
	if err = SaveToExcel(xlFile, queryStr, "RuleGroupCondition", delegatee, ExportRuleGroupCondition{}); err != nil {
		return nil, err
	}

	logger.Trace.Println("export Relation_Rule_ConversationFlow ...")
	queryStr = "SELECT %s" +
		strings.Repeat(", %s", reflect.TypeOf(ExportRuleConversationFlow{}).NumField()-1) +
		" FROM %s"
	queryStr = fmt.Sprintf(
		queryStr,
		"id", "rule_id", "cf_id",
		"Relation_Rule_ConversationFlow",
	)
	if err = SaveToExcel(xlFile, queryStr, "R_Rule_CF", delegatee, ExportRuleConversationFlow{}); err != nil {
		return nil, err
	}

	logger.Trace.Println("export Relation_RuleGroup_Rule ...")
	queryStr = "SELECT %s" +
		strings.Repeat(", %s", reflect.TypeOf(ExportRuleGroupRule{}).NumField()-1) +
		" FROM %s"
	queryStr = fmt.Sprintf(
		queryStr,
		"id", "rg_id", "rule_id",
		"Relation_RuleGroup_Rule",
	)
	if err = SaveToExcel(xlFile, queryStr, "Rel_RuleGroup_Rule", delegatee, ExportRuleGroupRule{}); err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)
	xlFile.Write(writer)

	return &buf, err
}

func (s *GroupSQLDao) ImportGroups(sqlLike SqlLike, fileName string) error {

	return fmt.Errorf("sorry, rd is developing ... \n")

	//xlFile, err := xlsx.OpenFile(fileName)
	//
	//if err != nil {
	//	logger.Error.Printf("fail to open %s \n", fileName)
	//	return err
	//}
	//
	//sqlStr := ""
	//
	//for _, sheet := range xlFile.Sheets {
	//	switch sheet.Name {
	//	case "Tag":
	//		tableName := "Tag"
	//		num := reflect.TypeOf(ExportTag{}).NumField() - 1
	//		sqlStr = fmt.Sprintf("INSERT INTO "+tableName+" ( %s"+strings.Repeat(", %s", num)+" ) VALUES ( ?"+strings.Repeat(", ?", num)+" )",
	//			fldTagID, fldTagIsDeleted, fldTagName, fldTagType, fldTagPosSen,
	//			fldTagNegSen, fldTagCreateTime, fldTagUpdateTime, fldTagEnterprise, fldTagUUID)
	//
	//		if sheet, ok := xlFile.Sheet[sheet.Name]; ok {
	//			if err := LoadFromExcel(sheet, sqlLike, sqlStr, tableName, ExportTag{}); err != nil {
	//				logger.Error.Printf("fail to insert into table %s \n", tableName)
	//				logger.Error.Println(err)
	//				return err
	//			}
	//		} else {
	//			logger.Error.Printf("can not find sheet %s \n", sheet.Name)
	//			return fmt.Errorf("can not find sheet %s \n", sheet.Name)
	//		}
	//	case "Sentence":
	//		tableName := "Sentence"
	//		num := reflect.TypeOf(ExportSentence{}).NumField() - 1
	//		sqlStr = fmt.Sprintf("INSERT INTO "+tableName+" ( %s"+strings.Repeat(", %s", num)+" ) VALUES ( ?"+strings.Repeat(", ?", num)+" )",
	//			"id", "is_delete", "name", "enterprise", "uuid", "create_time", "update_time", "category_id")
	//
	//		if sheet, ok := xlFile.Sheet[sheet.Name]; ok {
	//			if err := LoadFromExcel(sheet, sqlLike, sqlStr, tableName, ExportSentence{}); err != nil {
	//				logger.Error.Printf("fail to insert into table %s \n", tableName)
	//				logger.Error.Println(err)
	//				return err
	//			}
	//		} else {
	//			logger.Error.Printf("can not find sheet %s \n", sheet.Name)
	//			return fmt.Errorf("can not find sheet %s \n", sheet.Name)
	//		}
	//
	//	case "SentenceGroup":
	//		tableName := "SentenceGroup"
	//		num := reflect.TypeOf(ExportSentenceGroup{}).NumField() - 1
	//		sqlStr = fmt.Sprintf("INSERT INTO "+tableName+" ( %s"+strings.Repeat(", %s", num)+" ) VALUES ( ?"+strings.Repeat(", ?", num)+" )",
	//			"id", "is_delete", "name", "enterprise", "role", "position", "`range`", "uuid", "create_time", "update_time", "optional", "type")
	//
	//		if sheet, ok := xlFile.Sheet[sheet.Name]; ok {
	//			if err := LoadFromExcel(sheet, sqlLike, sqlStr, tableName, ExportSentenceGroup{}); err != nil {
	//				logger.Error.Printf("fail to insert into table %s \n", tableName)
	//				logger.Error.Println(err)
	//				return err
	//			}
	//		} else {
	//			logger.Error.Printf("can not find sheet %s \n", sheet.Name)
	//			return fmt.Errorf("can not find sheet %s \n", sheet.Name)
	//		}
	//
	//	case "Rule":
	//		tableName := "Rule"
	//		num := reflect.TypeOf(ExportRule{}).NumField() - 1
	//		sqlStr = fmt.Sprintf("INSERT INTO "+tableName+" ( %s"+strings.Repeat(", %s", num)+" ) VALUES ( ?"+strings.Repeat(", ?", num)+" )",
	//			"id", "is_delete", "name", "method", "score", "description", "enterprise", "min", "max", "severity", "uuid", "create_time", "update_time")
	//
	//		if sheet, ok := xlFile.Sheet[sheet.Name]; ok {
	//			if err := LoadFromExcel(sheet, sqlLike, sqlStr, tableName, ExportRule{}); err != nil {
	//				logger.Error.Printf("fail to insert into table %s \n", tableName)
	//				logger.Error.Println(err)
	//				return err
	//			}
	//		} else {
	//			logger.Error.Printf("can not find sheet %s \n", sheet.Name)
	//			return fmt.Errorf("can not find sheet %s \n", sheet.Name)
	//		}
	//
	//	case "RuleGroup":
	//		tableName := "RuleGroup"
	//		num := reflect.TypeOf(ExportRuleGroup{}).NumField() - 1
	//		sqlStr = fmt.Sprintf("INSERT INTO "+tableName+" ( %s"+strings.Repeat(", %s", num)+" ) VALUES ( ?"+strings.Repeat(", ?", num)+" )",
	//			"id", "is_delete", "name", "enterprise", "description", "create_time", "update_time", "is_enable", "limit_speed", "limit_silence", "type", "uuid")
	//
	//		if sheet, ok := xlFile.Sheet[sheet.Name]; ok {
	//			if err := LoadFromExcel(sheet, sqlLike, sqlStr, tableName, ExportRuleGroup{}); err != nil {
	//				logger.Error.Printf("fail to insert into table %s \n", tableName)
	//				logger.Error.Println(err)
	//				return err
	//			}
	//		} else {
	//			logger.Error.Printf("can not find sheet %s \n", sheet.Name)
	//			return fmt.Errorf("can not find sheet %s \n", sheet.Name)
	//		}
	//
	//	case "SensitiveWord":
	//		tableName := "SensitiveWord"
	//		num := reflect.TypeOf(ExportSensitiveWord{}).NumField() - 1
	//		sqlStr = fmt.Sprintf("INSERT INTO "+tableName+" ( %s"+strings.Repeat(", %s", num)+" ) VALUES ( ?"+strings.Repeat(", ?", num)+" )",
	//			"id", "uuid", "name", "enterprise", "score", "category_id", "is_delete")
	//
	//		if sheet, ok := xlFile.Sheet[sheet.Name]; ok {
	//			if err := LoadFromExcel(sheet, sqlLike, sqlStr, tableName, ExportSensitiveWord{}); err != nil {
	//				logger.Error.Printf("fail to insert into table %s \n", tableName)
	//				logger.Error.Println(err)
	//				return err
	//			}
	//		} else {
	//			logger.Error.Printf("can not find sheet %s \n", sheet.Name)
	//			return fmt.Errorf("can not find sheet %s \n", sheet.Name)
	//		}
	//
	//	case "ConversationFlow":
	//		tableName := "ConversationFlow"
	//		num := reflect.TypeOf(ExportConversationFlow{}).NumField() - 1
	//		sqlStr = fmt.Sprintf("INSERT INTO "+tableName+" ( %s"+strings.Repeat(", %s", num)+" ) VALUES ( ?"+strings.Repeat(", ?", num)+" )",
	//			"id", "is_delete", "name", "enterprise", "expression", "uuid", "create_time", "update_time", "min")
	//
	//		if sheet, ok := xlFile.Sheet[sheet.Name]; ok {
	//			if err := LoadFromExcel(sheet, sqlLike, sqlStr, tableName, ExportConversationFlow{}); err != nil {
	//				logger.Error.Printf("fail to insert into table %s \n", tableName)
	//				logger.Error.Println(err)
	//				return err
	//			}
	//		} else {
	//			logger.Error.Printf("can not find sheet %s \n", sheet.Name)
	//			return fmt.Errorf("can not find sheet %s \n", sheet.Name)
	//		}
	//
	//	case "R_Sen_Tag":
	//		tableName := "Relation_Sentence_Tag"
	//		num := reflect.TypeOf(ExportSentenceTag{}).NumField() - 1
	//		sqlStr = fmt.Sprintf("INSERT INTO "+tableName+" ( %s"+strings.Repeat(", %s", num)+" ) VALUES ( ?"+strings.Repeat(", ?", num)+" )",
	//			"id", "s_id", "tag_id")
	//
	//		if sheet, ok := xlFile.Sheet[sheet.Name]; ok {
	//			if err := LoadFromExcel(sheet, sqlLike, sqlStr, tableName, ExportSentenceTag{}); err != nil {
	//				logger.Error.Printf("fail to insert into table %s \n", tableName)
	//				logger.Error.Println(err)
	//				return err
	//			}
	//		} else {
	//			logger.Error.Printf("can not find sheet %s \n", sheet.Name)
	//			return fmt.Errorf("can not find sheet %s \n", sheet.Name)
	//		}
	//
	//	case "R_SenGrp_Sen":
	//		tableName := "Relation_SentenceGroup_Sentence"
	//		num := reflect.TypeOf(ExportSentenceGroupSentence{}).NumField() - 1
	//		sqlStr = fmt.Sprintf("INSERT INTO "+tableName+" ( %s"+strings.Repeat(", %s", num)+" ) VALUES ( ?"+strings.Repeat(", ?", num)+" )",
	//			"id", "sg_id", "s_id")
	//
	//		if sheet, ok := xlFile.Sheet[sheet.Name]; ok {
	//			if err := LoadFromExcel(sheet, sqlLike, sqlStr, tableName, ExportSentenceGroupSentence{}); err != nil {
	//				logger.Error.Printf("fail to insert into table %s \n", tableName)
	//				logger.Error.Println(err)
	//				return err
	//			}
	//		} else {
	//			logger.Error.Printf("can not find sheet %s \n", sheet.Name)
	//			return fmt.Errorf("can not find sheet %s \n", sheet.Name)
	//		}
	//
	//	case "R_SensitiveWord_Sen":
	//		tableName := "Relation_SensitiveWord_Sentence"
	//		num := reflect.TypeOf(ExportSensitiveWordSentence{}).NumField() - 1
	//		sqlStr = fmt.Sprintf("INSERT INTO "+tableName+" ( %s"+strings.Repeat(", %s", num)+" ) VALUES ( ?"+strings.Repeat(", ?", num)+" )",
	//			"id", "sw_id", "s_id", "type")
	//
	//		if sheet, ok := xlFile.Sheet[sheet.Name]; ok {
	//			if err := LoadFromExcel(sheet, sqlLike, sqlStr, tableName, ExportSensitiveWordSentence{}); err != nil {
	//				logger.Error.Printf("fail to insert into table %s \n", tableName)
	//				logger.Error.Println(err)
	//				return err
	//			}
	//		} else {
	//			logger.Error.Printf("can not find sheet %s \n", sheet.Name)
	//			return fmt.Errorf("can not find sheet %s \n", sheet.Name)
	//		}
	//
	//	case "RuleGroupCondition":
	//		tableName := "RuleGroupCondition"
	//		num := reflect.TypeOf(ExportRuleGroupCondition{}).NumField() - 1
	//		sqlStr = fmt.Sprintf("INSERT INTO "+tableName+" ( %s"+strings.Repeat(", %s", num)+" ) VALUES ( ?"+strings.Repeat(", ?", num)+" )",
	//			"id", "rg_id", "type", "file_name", "deal", "series", "upload_time", "staff_id", "staff_name", "extension", "department", "customer_id", "customer_name", "customer_phone", "category", "call_start", "call_end", "left_channel", "right_channel")
	//
	//		if sheet, ok := xlFile.Sheet[sheet.Name]; ok {
	//			if err := LoadFromExcel(sheet, sqlLike, sqlStr, tableName, ExportRuleGroupCondition{}); err != nil {
	//				logger.Error.Printf("fail to insert into table %s \n", tableName)
	//				logger.Error.Println(err)
	//				return err
	//			}
	//
	//		} else {
	//			logger.Error.Printf("can not find sheet %s \n", sheet.Name)
	//			return fmt.Errorf("can not find sheet %s \n", sheet.Name)
	//		}
	//
	//	case "R_Rule_CF":
	//		tableName := "Relation_Rule_ConversationFlow"
	//		num := reflect.TypeOf(ExportRuleConversationFlow{}).NumField() - 1
	//		sqlStr = fmt.Sprintf("INSERT INTO "+tableName+" ( %s"+strings.Repeat(", %s", num)+" ) VALUES ( ?"+strings.Repeat(", ?", num)+" )",
	//			"id", "rule_id", "cf_id")
	//
	//		if sheet, ok := xlFile.Sheet[sheet.Name]; ok {
	//			if err := LoadFromExcel(sheet, sqlLike, sqlStr, tableName, ExportRuleConversationFlow{}); err != nil {
	//				logger.Error.Printf("fail to insert into table %s \n", tableName)
	//				logger.Error.Println(err)
	//				return err
	//			}
	//		} else {
	//			logger.Error.Printf("can not find sheet %s \n", sheet.Name)
	//			return fmt.Errorf("can not find sheet %s \n", sheet.Name)
	//		}
	//
	//	case "Rel_RuleGroup_Rule":
	//		tableName := "Relation_RuleGroup_Rule"
	//		num := reflect.TypeOf(ExportRuleGroupRule{}).NumField() - 1
	//		sqlStr = fmt.Sprintf("INSERT INTO "+tableName+" ( %s"+strings.Repeat(", %s", num)+" ) VALUES ( ?"+strings.Repeat(", ?", num)+" )",
	//			"id", "rg_id", "rule_id")
	//
	//		if sheet, ok := xlFile.Sheet[sheet.Name]; ok {
	//			if err := LoadFromExcel(sheet, sqlLike, sqlStr, tableName, ExportRuleGroupRule{}); err != nil {
	//				logger.Error.Printf("fail to insert into table %s \n", tableName)
	//				logger.Error.Println(err)
	//				return err
	//			}
	//		} else {
	//			logger.Error.Printf("can not find sheet %s \n", sheet.Name)
	//			return fmt.Errorf("can not find sheet %s \n", sheet.Name)
	//		}
	//
	//	}
	//
	//}
	//return nil
}

func LoadFromExcel(sheet *xlsx.Sheet, sqlLike SqlLike, sqlStr string, tableName string, bean interface{}) error {

	logger.Trace.Printf("truncate table %s \n", tableName)

	trunStr := fmt.Sprintf("TRUNCATE TABLE %s", tableName)
	_, err := sqlLike.Exec(trunStr)
	if err != nil {
		logger.Error.Printf("failt to truncate table %s \n", tableName)
		return err
	}

	logger.Trace.Printf("import data to table %s \n", tableName)
	for i, row := range sheet.Rows {
		if i == 0 {
			continue
		}
		var sqlParams = make([]interface{}, 0)
		t := reflect.TypeOf(bean)
		for j, cell := range row.Cells {
			fieldType := t.Field(j).Type.Kind()

			switch fieldType {
			case reflect.Int:
				columnValue, err := cell.Int()
				if err != nil {
					if cell.String() == "" {
						sqlParams = append(sqlParams, nil)
						continue
					}
					return err
				}
				sqlParams = append(sqlParams, columnValue)
			case reflect.Int64:
				columnValue, err := cell.Int64()
				if err != nil {
					if cell.String() == "" {
						sqlParams = append(sqlParams, nil)
						continue
					}
					return err
				}
				sqlParams = append(sqlParams, columnValue)
			case reflect.Uint64:
				columnValue, err := cell.Int64()
				if err != nil {
					if cell.String() == "" {
						sqlParams = append(sqlParams, nil)
						continue
					}
					return err
				}
				sqlParams = append(sqlParams, columnValue)
			case reflect.Float32:
				columnValue, err := cell.Float()
				if err != nil {
					if cell.String() == "" {
						sqlParams = append(sqlParams, nil)
						continue
					}
					return err
				}
				sqlParams = append(sqlParams, columnValue)
			default:
				sqlParams = append(sqlParams, cell.String())
			}
		}
		_, err := sqlLike.Exec(sqlStr, sqlParams...)
		if err != nil {
			logger.Error.Printf("fail to insert into %s \n", tableName)
			logger.Error.Print("params: ")
			for _, para := range sqlParams {
				logger.Error.Print(para, " ")
			}
			logger.Error.Println()
			return err
		}
	}
	return nil
}

func SaveToExcel(xlFile *xlsx.File, queryStr string, sheetName string, delegatee SqlLike, bean interface{}) error {

	rows, err := delegatee.Query(queryStr)
	if err != nil {
		logger.Error.Println("raw sql: ", queryStr)
		logger.Error.Println("raw bind-data: ", "")
		return fmt.Errorf("sql executed failed, %v", err)
	}
	defer rows.Close()

	sheet, err := xlFile.AddSheet(sheetName)
	row := sheet.AddRow()

	t := reflect.TypeOf(bean)
	if t.Kind() != reflect.Struct {
		return errors.New("kind error")
	}

	// Add title
	for i := 0; i < t.NumField(); i++ {
		cell := row.AddCell()
		cell.Value = t.Field(i).Name
	}

	scanResults := make([]interface{}, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		var column interface{}
		scanResults[i] = &column
	}

	for rows.Next() {
		row = sheet.AddRow()

		if err = rows.Scan(scanResults...); err != nil {
			return err
		}

		for j := 0; j < t.NumField(); j++ {
			cell := row.AddCell()

			rawValue := reflect.Indirect(reflect.ValueOf(scanResults[j]))
			// if rwo is null then ignore
			if rawValue.Interface() == nil {
				continue
			}

			rawValueType := reflect.TypeOf(rawValue.Interface())
			vv := reflect.ValueOf(rawValue.Interface())

			hasAssigned := false

			switch t.Field(j).Type.Kind() {
			case reflect.String:
				if rawValueType.Kind() == reflect.String {
					hasAssigned = true
					cell.Value = vv.String()
					continue
				}
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				switch rawValueType.Kind() {
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					hasAssigned = true
					cell.Value = strconv.FormatInt(vv.Int(), 10)
					continue
				}

			case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
				switch rawValueType.Kind() {
				case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
					hasAssigned = true
					cell.Value = strconv.FormatInt(vv.Int(), 10)
					continue
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					hasAssigned = true
					cell.Value = strconv.FormatInt(vv.Int(), 10)
					continue
				}

			case reflect.Float32, reflect.Float64:
				switch rawValueType.Kind() {
				case reflect.Float32, reflect.Float64:
					hasAssigned = true
					cell.Value = strconv.FormatFloat(vv.Float(), 'e', 5, 32)
					continue
				}
			}

			if !hasAssigned {
				data, err := value2Bytes(&rawValue)
				if err != nil {
					return err
				}
				cell.Value = data

			}

		}

	}
	return nil
}

func value2Bytes(rawValue *reflect.Value) (string, error) {
	str, err := value2String(rawValue)
	if err != nil {
		return "", err
	}
	return str, nil
}

func value2String(rawValue *reflect.Value) (str string, err error) {
	aa := reflect.TypeOf((*rawValue).Interface())
	vv := reflect.ValueOf((*rawValue).Interface())
	var cTIME time.Time
	timeType := reflect.TypeOf(cTIME)
	switch aa.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		str = strconv.FormatInt(vv.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		str = strconv.FormatUint(vv.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		str = strconv.FormatFloat(vv.Float(), 'f', -1, 64)
	case reflect.String:
		str = vv.String()
	case reflect.Array, reflect.Slice:
		switch aa.Elem().Kind() {
		case reflect.Uint8:
			data := rawValue.Interface().([]byte)
			str = string(data)
			if str == "\x00" {
				str = "0"
			}
		default:
			err = fmt.Errorf("Unsupported struct type %v", vv.Type().Name())
		}
		// time type
	case reflect.Struct:
		if aa.ConvertibleTo(timeType) {
			str = vv.Convert(timeType).Interface().(time.Time).Format(time.RFC3339Nano)
		} else {
			err = fmt.Errorf("Unsupported struct type %v", vv.Type().Name())
		}
	case reflect.Bool:
		str = strconv.FormatBool(vv.Bool())
	case reflect.Complex128, reflect.Complex64:
		str = fmt.Sprintf("%v", vv.Complex())
		/* TODO: unsupported types below
		   case reflect.Map:
		   case reflect.Ptr:
		   case reflect.Uintptr:
		   case reflect.UnsafePointer:
		   case reflect.Chan, reflect.Func, reflect.Interface:
		*/
	default:
		err = fmt.Errorf("Unsupported struct type %v", vv.Type().Name())
	}
	return
}

type ExportTag struct {
	ID           uint64
	IsDelete     int
	Name         string
	Type         int
	PosSentences string
	NegSentences string
	CreateTime   int64
	UpdateTime   int64
	Enterprise   string
	UUID         string
}

type ExportSentence struct {
	ID         uint64
	IsDelete   int
	Name       string
	Enterprise string
	UUID       string
	CreateTime int64
	UpdateTime int64
	CategoryID uint64
}

type ExportSentenceGroup struct {
	ID         uint64
	IsDelete   int
	Name       string
	Enterprise string
	Role       int
	Position   int
	Range      int
	UUID       string
	CreateTime uint64
	UpdateTime uint64
	Optional   int
	Type       int
}

type ExportRule struct {
	ID          uint64
	IsDelete    int
	Name        string
	Method      int
	Score       int
	Description string
	Enterprise  string
	Min         int
	Max         int
	Severity    int
	UUID        string
	CreateTime  int64
	UpdateTime  int64
}

type ExportRuleGroup struct {
	ID           uint64
	IsDelete     bool
	Name         string
	Enterprise   string
	Description  string
	CreateTime   uint64
	UpdateTime   uint64
	IsEnable     int
	LimitSpeed   int
	LimitSilence float32
	Type         int
	UUID         string
}

type ExportSensitiveWord struct {
	ID         uint64
	UUID       string
	Name       string
	Enterprise string
	Score      int
	CategoryId uint64
	IsDelete   int
}

type ExportConversationFlow struct {
	ID         uint64
	IsDelete   int
	Name       string
	Enterprise string
	Expression string
	UUID       string
	CreateTime int64
	UpdateTime int64
	Min        int
}

type ExportSentenceTag struct {
	ID    uint64
	SID   uint64
	TagID uint64
}

type ExportSentenceGroupSentence struct {
	ID   uint64
	SgID uint64
	SID  uint64
}

type ExportSensitiveWordSentence struct {
	ID   uint64
	SwID uint64
	SID  uint64
	Type int
}

type ExportRuleGroupCondition struct {
	ID            uint64
	RgID          uint64
	Type          int
	FileName      string
	Deal          int
	Series        string
	UploadTime    uint64
	StaffID       string
	StaffName     string
	Extension     string
	Department    string
	CustomerID    string
	CustomerName  string
	CustomerPhone string
	Category      string
	CallStart     uint64
	CallEnd       uint64
	LeftChannel   string
	RightChannel  string
}

type ExportRuleConversationFlow struct {
	ID     uint64
	RuleID uint64
	CfID   uint64
}

type ExportRuleGroupRule struct {
	ID     uint64
	RgID   uint64
	RuleID uint64
}
