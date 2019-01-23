package model

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/pkg/logger"
)

type GroupDAO interface {
	Begin() (*sql.Tx, error)
	Commit(tx *sql.Tx) error
	ClearTranscation(tx *sql.Tx)
	CountGroupsBy(filter *GroupFilter) (int64, error)
	CreateGroup(group *GroupWCond, tx *sql.Tx) (*GroupWCond, error)
	Group(delegatee SqlLike, query GroupQuery) ([]Group, error)
	// UpdateGroup(id int64, group *GroupWCond, tx *sql.Tx) error
	DeleteGroup(id string, tx *sql.Tx) error
	GetGroupsBy(filter *GroupFilter) ([]GroupWCond, error)
	GroupsByCalls(delegatee SqlLike, query CallQuery) (map[int64][]Group, error)
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
}

// Group is the one to one represent of rule group table schema
type Group struct {
	ID             int64
	Name           string
	EnterpriseID   string
	Description    string
	CreatedTime    int64
	UpdatedTime    int64
	IsDelete       bool
	IsEnable       bool
	LimitedSpeed   int
	LimitedSilence float32
	Typ            int8
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

//InitDB is used to get the db in this module
//	deprecated, origin version should not be used anymore for performance and race-condition issues.
//	It is keeped only to minimize code changed for current functions.
//  GroupSQLDao will get the inner conn db at somewhere else.
func (s *GroupSQLDao) initDB() error {
	if s.conn == nil {
		return fmt.Errorf("package db have not initialized yet")
	}
	return nil
}

//Begin is used to start a transaction
func (s *GroupSQLDao) Begin() (*sql.Tx, error) {
	if s.conn == nil {
		err := s.initDB()
		if err != nil {
			return nil, err
		}
	}
	return s.conn.Begin()
}

//Commit commits the data
func (s *GroupSQLDao) Commit(tx *sql.Tx) error {
	if tx != nil {
		return tx.Commit()
	}
	return nil
}

func (s *GroupSQLDao) ClearTranscation(tx *sql.Tx) {
	if tx != nil {
		util.ClearTransition(tx)
	}
}

func (s *GroupSQLDao) GetGroups() (groups []GroupWCond, err error) {
	if s.conn == nil {
		err = s.initDB()
		if err != nil {
			return
		}
	}

	queryStr := fmt.Sprintf(
		"SELECT %s, %s FROM %s where %s=1",
		fldRuleGrpID,
		fldRuleGrpName,
		tblRuleGroup,
		fldRuleGrpIsEnable,
	)

	rows, err := s.conn.Query(queryStr)
	if err != nil {
		err = fmt.Errorf("error while query groups in dao.GetGroups, err: %s", err.Error())
		return
	}
	defer rows.Close()

	groups = make([]GroupWCond, 0)
	for rows.Next() {
		group := GroupWCond{}
		rows.Scan(&group.ID, &group.Name)

		groups = append(groups, group)
	}
	return
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

	if filter.Delete != -1 {
		groupStr = fmt.Sprintf("%s = ?", fldIsDelete)
		groupConditions = append(groupConditions, groupStr)
		values = append(values, filter.Delete)
	}

	if len(groupConditions) > 0 {
		groupStr = fmt.Sprintf("%s %s", "WHERE", strings.Join(groupConditions, " and "))
	} else {
		groupStr = ""
	}

	conditions := []string{}
	conditionStr := "WHERE"
	if filter.FileName != "" {
		conditions = append(conditions, fmt.Sprintf("%s = ?", RGCFileName))
		values = append(values, filter.FileName)
	}

	if filter.CallEnd != 0 {
		conditions = append(conditions, fmt.Sprintf("%s = ?", RGCCallEnd))
		values = append(values, filter.CallEnd)
	}

	if filter.CallStart != 0 {
		conditions = append(conditions, fmt.Sprintf("%s = ?", RGCCallStart))
		values = append(values, filter.CallStart)
	}

	if filter.CustomerID != "" {
		conditions = append(conditions, fmt.Sprintf("%s = ?", RGCCustomerID))
		values = append(values, filter.CustomerID)
	}

	if filter.CustomerName != "" {
		conditions = append(conditions, fmt.Sprintf("%s = ?", RGCCustomerName))
		values = append(values, filter.CustomerName)
	}

	if filter.CustomerPhone != "" {
		conditions = append(conditions, fmt.Sprintf("%s = ?", RGCCustomerPhone))
		values = append(values, filter.CustomerPhone)
	}

	if filter.Deal != -1 {
		conditions = append(conditions, fmt.Sprintf("%s = ?", RGCDeal))
		values = append(values, filter.Deal)
	}

	if filter.Department != "" {
		conditions = append(conditions, fmt.Sprintf("%s = ?", RGCDepartment))
		values = append(values, filter.Department)
	}

	if filter.Extension != "" {
		conditions = append(conditions, fmt.Sprintf("%s = ?", RGCExtension))
		values = append(values, filter.Extension)
	}

	if filter.Series != "" {
		conditions = append(conditions, fmt.Sprintf("%s = ?", RGCSeries))
		values = append(values, filter.Series)
	}

	if filter.StaffID != "" {
		conditions = append(conditions, fmt.Sprintf("%s = ?", RGCStaffID))
		values = append(values, filter.StaffID)
	}

	if filter.StaffName != "" {
		conditions = append(conditions, fmt.Sprintf("%s = ?", RGCStaffName))
		values = append(values, filter.StaffName)
	}

	if len(conditions) == 0 {
		conditionStr = ""
	} else {
		conditionStr = fmt.Sprintf("%s %s", conditionStr, strings.Join(conditions, " and "))
	}

	queryStr = `SELECT rg.%s, rg.%s, rg.%s, rg.%s, rg.%s, rg.%s, rg.%s, rg.%s, rg.%s, 
	gc.%s, gc.%s, gc.%s, gc.%s, gc.%s, gc.%s, gc.%s, 
	gc.%s, gc.%s, gc.%s, gc.%s, gc.%s, gc.%s, gc.%s,
	r.%s as rUUID, r.%s as rName
	FROM (SELECT * FROM %s %s) as rg
	INNER JOIN (SELECT * FROM %s %s) as gc on rg.%s = gc.%s
	LEFT JOIN %s as rrr ON rg.%s = rrr.%s
	LEFT JOIN %s as r on rrr.%s = r.%s
	`

	queryStr = fmt.Sprintf(
		queryStr,
		fldRuleGrpID,
		fldRuleGrpUUID,
		fldRuleGrpName,
		fldDescription,
		fldRuleGrpLimitSpeed,
		fldRuleGrpLimitSilence,
		fldCreateTime,
		fldRuleGrpIsEnable,
		fldEnterprise,
		RGCFileName,
		RGCDeal,
		RGCSeries,
		RGCStaffID,
		RGCStaffName,
		RGCExtension,
		RGCDepartment,
		RGCCustomerID,
		RGCCustomerName,
		RGCCustomerPhone,
		RGCCallStart,
		RGCCallEnd,
		RGCLeftChannel,
		RGCRightChannel,
		fldUUID,
		fldName,
		tblRuleGroup,
		groupStr,
		tblRGC,
		conditionStr,
		fldID,
		RGCGroupID,
		tblRelGrpRule,
		fldID,
		RRRGroupID,
		tblRule,
		RRRRuleID,
		fldID,
	)
	return
}

func (s *GroupSQLDao) CountGroupsBy(filter *GroupFilter) (total int64, err error) {
	if s.conn == nil {
		err = s.initDB()
		if err != nil {
			err = fmt.Errorf("error while init db in dao.CreateGroup, err: %s", err.Error())
			return
		}
	}

	queryStr, values := getGroupsSQL(filter)
	queryStr = fmt.Sprintf("SELECT count(rg.%s) FROM (%s) as rg", fldRuleGrpID, queryStr)

	rows, err := s.conn.Query(queryStr, values...)
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

func (s *GroupSQLDao) GetGroupsBy(filter *GroupFilter) (groups []GroupWCond, err error) {
	if s.conn == nil {
		err = s.initDB()
		if err != nil {
			err = fmt.Errorf("error while init db in dao.CreateGroup, err: %s", err.Error())
			return
		}
	}

	queryStr, values := getGroupsSQL(filter)
	if filter.Limit > 0 {
		start := filter.Page * filter.Limit
		queryStr = fmt.Sprintf("%s LIMIT %d, %d", queryStr, start, filter.Limit)
	}

	rows, err := s.conn.Query(queryStr, values...)
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

		rows.Scan(
			&group.ID,
			&group.UUID,
			&group.Name,
			&group.Description,
			&group.Speed,
			&group.SlienceDuration,
			&group.CreateTime,
			&group.Enabled,
			&group.Enterprise,
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
			&rUUID,
			&rName,
		)

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

func genInsertRelationSQL(id int64, rules *[]SimpleConversationRule) (str string, values []interface{}) {
	str = fmt.Sprintf(
		"INSERT INTO %s (%s, %s) VALUES ",
		tblRelGrpRule,
		RRRGroupID,
		RRRRuleID,
	)
	values = []interface{}{}
	for _, rule := range *rules {
		str = addCommaIfNotFirst(str, len(values) == 0)
		str += " (?, ?)"
		values = append(values, id, rule.ID)
	}
	return
}

func (s *GroupSQLDao) CreateGroup(group *GroupWCond, tx *sql.Tx) (createdGroup *GroupWCond, err error) {
	if s.conn == nil {
		err = s.initDB()
		if err != nil {
			err = fmt.Errorf("error while init db in dao.CreateGroup, err: %s", err.Error())
			return
		}
	}

	now := time.Now().Unix()

	// insert group
	insertStr := fmt.Sprintf(
		"INSERT INTO `%s` (%s, %s, %s, %s, %s, %s, %s, %s, %s) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		tblRuleGroup,
		fldRuleGrpUUID,
		fldRuleGrpName,
		fldRuleGrpEnterpriseID,
		fldDescription,
		fldRuleGrpCreateTime,
		fldRuleGrpUpdateTime,
		fldRuleGrpIsEnable,
		fldRuleGrpLimitSpeed,
		fldRuleGrpLimitSilence,
	)

	values := []interface{}{
		group.UUID,
		group.Name,
		group.Enterprise,
		group.Description,
		now,
		now,
		group.Enabled,
		group.Speed,
		group.SlienceDuration,
	}
	result, err := tx.Exec(insertStr, values...)
	if err != nil {
		err = fmt.Errorf("error while insert group in dao.CreateGroup, err: %s", err.Error())
		return
	}

	groupID, err := result.LastInsertId()
	if err != nil {
		err = fmt.Errorf("error while get group id in dao.CreateGroup, err: %s", err.Error())
		return
	}

	// insert condition
	insertStr = fmt.Sprintf(
		"INSERT INTO %s (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		tblRGC,
		RGCGroupID,
		RGCFileName,
		RGCDeal,
		RGCSeries,
		RGCStaffID,
		RGCStaffName,
		RGCExtension,
		RGCDepartment,
		RGCCustomerID,
		RGCCustomerName,
		RGCCustomerPhone,
		RGCCallStart,
		RGCCallEnd,
		RGCLeftChannel,
		RGCRightChannel,
	)
	values = []interface{}{
		groupID,
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
	}

	_, err = tx.Exec(insertStr, values...)
	if err != nil {
		err = fmt.Errorf("error while insert condition in dao.CreateGroup, err: %s", err.Error())
		return
	}

	// insert into group_rule_map
	if group.Rules != nil && len(*group.Rules) != 0 {
		insertStr, values = genInsertRelationSQL(groupID, group.Rules)

		_, err = tx.Exec(insertStr, values...)
		if err != nil {
			err = fmt.Errorf("error while insert relation_group_rule in dao.CreateGroup, err: %s", err.Error())
			return
		}
	}

	group.ID = groupID
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

func (s *GroupSQLDao) DeleteGroup(id string, tx *sql.Tx) (err error) {
	if tx == nil {
		return
	}

	deleteStr := fmt.Sprintf(
		"UPDATE %s SET %s = 1 WHERE %s = ?",
		tblRuleGroup,
		fldRuleGrpIsDelete,
		fldRuleGrpUUID,
	)
	_, err = tx.Exec(deleteStr, id)
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
