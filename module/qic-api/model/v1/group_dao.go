package model

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"emotibot.com/emotigo/pkg/logger"
)

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
}

type GroupSQLDao struct {
	conn *sql.DB
}

func NewGroupSQLDao(conn *sql.DB) *GroupSQLDao {
	return &GroupSQLDao{
		conn: conn,
	}
}

type SimpleGroup struct {
	ID   string `json:"group_id"`
	Name string `json:"group_name"`
}

// GroupWCond is Group with Condition struct
type GroupWCond struct {
	ID              int64                     `json:"-"`
	UUID            string                    `json:"group_id,omitempty"`
	Name            *string                   `json:"group_name,omitempty"`
	Enterprise      string                    `json:",omitempty"`
	Enabled         *int8                     `json:"is_enable,omitempty"`
	Speed           *float64                  `json:"limit_speed,omitempty"`
	SlienceDuration *float64                  `json:"limit_silence,omitempty"`
	Rules           *[]SimpleConversationRule `json:"rules"`
	Condition       *GroupCondition           `json:"other,omitempty"`
	CreateTime      int64                     `json:"create_time,omitempty"`
	Description     *string                   `json:"description"`
	RuleCount       int                       `json:"rule_count"`
	Deleted         int8                      `json:"-"`
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
type Group struct {
	ID             int64
	IsDelete       bool
	Name           string
	EnterpriseID   string
	Description    string
	CreatedTime    int64
	UpdatedTime    int64
	IsEnable       bool
	LimitedSpeed   int
	LimitedSilence float32
	Typ            int8
	UUID           string
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
		fldCondLeftChan, fldCondRightChan,
	}
	ruleCols := []string{
		fldID, fldUUID, fldName,
	}
	queryStr = "SELECT rg.`%s`, gc.`%s`, r.`%s`" +
		" FROM (SELECT * FROM `%s` %s) as rg" +
		" LEFT JOIN (SELECT * FROM `%s` %s) as gc on rg.`%s` = gc.`%s`" + // gc group condition table
		" LEFT JOIN  `%s` as rrr ON rg.`%s` = rrr.`%s`" + // rrr Group_Rule relation table
		" %s as rule on rrr.`%s` = rule.`%s`"

	queryStr = fmt.Sprintf(queryStr,
		strings.Join(grpCols, "`, rg.`"), strings.Join(condCols, "`, gc.`"), strings.Join(ruleCols, "`, r.`"),
		tblRuleGroup, groupStr,
		tblRGC, conditionStr, fldID, fldCondGroupID,
		tblRelGrpRule, fldID, RRRGroupID,
		ruleCondition, RRRRuleID, fldID,
	)
	return
}

func (s *GroupSQLDao) CountGroupsBy(filter *GroupFilter, sqlLike SqlLike) (total int64, err error) {
	queryStr, values := getGroupsSQL(filter)
	queryStr = fmt.Sprintf("SELECT count(rg.%s) FROM (%s) as rg", fldRuleGrpID, queryStr)

	rows, err := sqlLike.Query(queryStr, values...)
	if err != nil {
		err = fmt.Errorf("error while count groups in dao.CountGroupsBy, err: %s", err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&total)
	}
	return
}

func (s *GroupSQLDao) GetGroupsBy(filter *GroupFilter, sqlLike SqlLike) (groups []GroupWCond, err error) {
	queryStr, values := getGroupsSQL(filter)
	if filter.Limit > 0 {
		start := filter.Page * filter.Limit
		queryStr = fmt.Sprintf("%s LIMIT %d, %d", queryStr, start, filter.Limit)
	}

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
		var rUUID *string // rule uuid
		var rName *string // rule name
		var rID *int64

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
			&rID,
			&rUUID,
			&rName,
		)

		if err != nil {
			err = fmt.Errorf("error whiel scan rule group in dao.GetBy, err: %s", err.Error())
			return
		}

		if currentGroup == nil || group.ID != currentGroup.ID {
			if currentGroup != nil {
				groups = append(groups, *currentGroup)
			}

			group.Condition = &condition

			currentGroup = &group
			rules := []SimpleConversationRule{}
			currentGroup.Rules = &rules
		}

		if rUUID != nil && rName != nil {
			rule := SimpleConversationRule{
				ID:   *rID,
				UUID: *rUUID,
				Name: *rName,
			}
			rules := append(*currentGroup.Rules, rule)
			currentGroup.Rules = &rules
		}
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
		fldCondLeftChan,
		fldCondRightChan,
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

func (s *GroupSQLDao) Group(delegatee SqlLike, query GroupQuery) ([]Group, error) {
	if delegatee == nil {
		delegatee = s.conn
	}
	groupCols := []string{
		fldRuleGrpID, fldRuleGrpIsDelete, fldRuleGrpName,
		fldRuleGrpEnterpriseID, fldRuleGrpDescription, fldRuleGrpCreateTime,
		fldRuleGrpUpdateTime, fldRuleGrpIsEnable, fldRuleGrpLimitSpeed,
		fldRuleGrpLimitSilence, fldRuleGrpType,
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
		rows.Scan(&g.ID, &isDeleted, &g.Name, &g.EnterpriseID, &g.Description, &g.CreatedTime, &g.UpdatedTime, &isEnabled, &g.LimitedSpeed, &g.LimitedSilence, &g.Typ)
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
