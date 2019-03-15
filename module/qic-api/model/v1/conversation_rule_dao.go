package model

import (
	"fmt"
	"strings"

	"emotibot.com/emotigo/pkg/logger"
)

// RuleMethod 正面與反面規則 flag
const (
	RuleMethodNegative int8 = 1
	RuleMethodPositive int8 = -1
)

type RuleCompleteness struct {
	RuleCompleted       int8
	HasDescription      bool
	HasConversationFlow bool
	SetenceCompleted    int8
}

type SimpleConversationRule struct {
	ID   int64  `json:"-"`
	UUID string `json:"rule_id"`
	Name string `json:"rule_name"`
}

type ConversationRule struct {
	ID           int64
	Deleted      int8
	Name         string
	Method       int8
	Score        int
	Description  string
	Enterprise   string
	Min          int
	Max          int
	Severity     int8
	UUID         string
	CreateTime   int64
	UpdateTime   int64
	Flows        []SimpleConversationFlow
	Completeness *RuleCompleteness
}

type ConversationRuleFilter struct {
	ID         []uint64
	UUID       []string
	Name       string
	Method     int8
	Enterprise string
	Severity   int8
	IsDeleted  int8
	CFUUID     []string // filter by conversation flow uuid
}

type ConversationRuleDao interface {
	Create(rule *ConversationRule, sqlLike SqlLike) (*ConversationRule, error)
	CreateMany(rules []ConversationRule, sqlLike SqlLike) error
	CountBy(filter *ConversationRuleFilter, sqlLike SqlLike) (int64, error)
	GetBy(filter *ConversationRuleFilter, sqlLike SqlLike) ([]ConversationRule, error)
	GetByFlowID([]int64, SqlLike) ([]ConversationRule, error)
	Delete(id string, sqlLike SqlLike) error
	DeleteMany(uuid []string, sqlLike SqlLike) error
}

type ConversationRuleSqlDaoImpl struct{}

func getConversationRuleInsertSQL(rules []ConversationRule) (insertStr string, values []interface{}) {
	if len(rules) == 0 {
		return
	}

	fields := []string{
		fldUUID,
		fldName,
		CRMethod,
		fldEnterprise,
		CRScore,
		CRDescription,
		CRMin,
		CRMax,
		CRSeverity,
		fldCreateTime,
		fldUpdateTime,
	}
	fieldStr := strings.Join(fields, ", ")

	values = []interface{}{}

	valueStr := fmt.Sprintf("(?%s)", strings.Repeat(", ?", len(fields)-1))
	allValueStr := ""
	for _, rule := range rules {
		values = append(
			values,
			rule.UUID,
			rule.Name,
			rule.Method,
			rule.Enterprise,
			rule.Score,
			rule.Description,
			rule.Min,
			rule.Max,
			rule.Severity,
			rule.CreateTime,
			rule.UpdateTime,
		)
		allValueStr = fmt.Sprintf("%s%s,", allValueStr, valueStr)
	}
	// remove last comma
	allValueStr = allValueStr[:len(allValueStr)-1]

	insertStr = fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES %s",
		tblConversationRule,
		fieldStr,
		allValueStr,
	)
	return
}

func getConversationRuleRelationInsertSQL(rules []ConversationRule) (insertStr string, values []interface{}) {
	if len(rules) == 0 {
		return
	}

	insertStr = fmt.Sprintf(
		"INSERT INTO %s (%s, %s) VALUES",
		tblRelCRCF,
		CRCFRID,
		CRCFCFID,
	)

	valueStr := ""
	values = []interface{}{}
	for _, rule := range rules {
		for _, flow := range rule.Flows {
			valueStr += "(?, ?),"
			values = append(values, rule.ID, flow.ID)
		}
	}
	if len(values) > 0 {
		// remove last comma
		valueStr = valueStr[:len(valueStr)-1]
	}

	insertStr = fmt.Sprintf("%s %s", insertStr, valueStr)
	return
}

func (dao *ConversationRuleSqlDaoImpl) Create(rule *ConversationRule, sqlLike SqlLike) (createdRule *ConversationRule, err error) {
	insertStr, values := getConversationRuleInsertSQL([]ConversationRule{*rule})

	result, err := sqlLike.Exec(insertStr, values...)
	if err != nil {
		logger.Error.Printf("error while insert conversation rule in dao.Create, sql: %s\n", insertStr)
		logger.Error.Printf("error while insert conversation rule in dao.Create, values: %+v\n", values)
		err = fmt.Errorf("error while insert conversation rule in dao.Create, err: %s", err.Error())
		return
	}

	ruleID, err := result.LastInsertId()
	if err != nil {
		err = fmt.Errorf("error while get rule id in dao.Create, err: %s", err.Error())
		return
	}

	if len(rule.Flows) > 0 {
		fieldStr := fmt.Sprintf("%s, %s", CRCFRID, CRCFCFID)
		valueStr := "(?, ?)"
		values = []interface{}{
			ruleID,
			rule.Flows[0].ID,
		}

		for _, flow := range rule.Flows[1:] {
			valueStr = fmt.Sprintf("%s, %s", valueStr, "(?, ?)")
			values = append(values, ruleID, flow.ID)
		}

		insertStr = fmt.Sprintf("INSERT INTO %s (%s) VALUES %s", tblRelCRCF, fieldStr, valueStr)

		_, err = sqlLike.Exec(insertStr, values...)
		if err != nil {
			err = fmt.Errorf("error while insert rule flow relation in dao.Create, err: %s", err.Error())
			return
		}
	}
	rule.ID = ruleID
	createdRule = rule
	return
}

func (dao *ConversationRuleSqlDaoImpl) CreateMany(rules []ConversationRule, sqlLike SqlLike) (err error) {
	if len(rules) == 0 {
		return
	}
	insertStr, values := getConversationRuleInsertSQL(rules)

	_, err = sqlLike.Exec(insertStr, values...)
	if err != nil {
		logger.Error.Printf("error while insert rules in dao.CreateMany, sql: %s\n", insertStr)
		logger.Error.Printf("error while insert rules in dao.CreateMany, values: %+v\n", values)
		err = fmt.Errorf("error while insert rules in dao.CreateMany, err: %s\n", err.Error())
		return
	}

	ruleUUID := make([]string, len(rules))
	for i := 0; i < len(rules); i++ {
		ruleUUID[i] = rules[i].UUID
	}

	filter := &ConversationRuleFilter{
		UUID: ruleUUID,
	}

	// we need to get new id of rules
	newRules, err := dao.GetBy(filter, sqlLike)
	if err != nil {
		logger.Error.Printf("error while get rules in dao.CreateMany, sql: %s\n", insertStr)
		logger.Error.Printf("error while get rules in dao.CreateMany, values: %+v\n", values)
		err = fmt.Errorf("error while get rules in dao.CreateMany, err: %s\n", err.Error())
		return
	}
	ruleIndex := map[string]int{}
	for idx, rule := range rules {
		ruleIndex[rule.UUID] = idx
	}

	for _, newRule := range newRules {
		idx := ruleIndex[newRule.UUID]
		rules[idx].ID = newRule.ID
	}

	insertStr, values = getConversationRuleRelationInsertSQL(rules)
	if len(values) > 0 {
		_, err = sqlLike.Exec(insertStr, values...)
		if err != nil {
			logger.Error.Printf("error while insert rules to flows relation in dao.CreateMany, sql: %s\n", insertStr)
			logger.Error.Printf("error while get rules to flows relation in dao.CreateMany, values: %+v\n", values)
			err = fmt.Errorf("error while get rules to flows relation in dao.CreateMany, err: %s\n", err.Error())
			return
		}
	}
	return
}

func queryConversationRulesSQLBy(filter *ConversationRuleFilter) (queryStr string, values []interface{}) {
	values = []interface{}{}
	conditionStr := "WHERE "
	conditions := []string{}

	if len(filter.UUID) > 0 {
		idStr := fmt.Sprintf("? %s", strings.Repeat(", ?", len(filter.UUID)-1))
		conditions = append(conditions, fmt.Sprintf("%s IN(%s)", fldUUID, idStr))

		for _, uuid := range filter.UUID {
			values = append(values, uuid)
		}
	}

	if len(filter.ID) > 0 {
		idStr := fmt.Sprintf("? %s", strings.Repeat(", ?", len(filter.ID)-1))
		conditions = append(conditions, fmt.Sprintf("%s IN(%s)", fldID, idStr))

		for _, id := range filter.ID {
			values = append(values, id)
		}
	}

	if filter.Enterprise != "" {
		conditions = append(conditions, fmt.Sprintf("%s=?", fldEnterprise))
		values = append(values, filter.Enterprise)
	}

	if filter.Name != "" {
		conditions = append(conditions, fmt.Sprintf("%s=?", fldName))
		values = append(values, filter.Name)
	}

	if filter.IsDeleted != -1 {
		conditions = append(conditions, fmt.Sprintf("%s=?", fldIsDelete))
		values = append(values, filter.IsDeleted)
	}

	if filter.Severity != -1 {
		conditions = append(conditions, fmt.Sprintf("%s=?", CRSeverity))
		values = append(values, filter.Severity)
	}

	if len(conditions) > 0 {
		conditionStr = fmt.Sprintf("%s %s", conditionStr, strings.Join(conditions, " and "))
	} else {
		conditionStr = ""
	}

	cfCondition := fmt.Sprintf("LEFT JOIN %s", tblConversationflow)
	if len(filter.CFUUID) > 0 {
		cfCondition = fmt.Sprintf(
			"INNER JOIN (SELECT * FROM %s WHERE %s IN (%s))",
			tblConversationflow,
			fldUUID,
			fmt.Sprintf("?%s", strings.Repeat(", ?", len(filter.CFUUID)-1)),
		)
		for _, cfUUID := range filter.CFUUID {
			values = append(values, cfUUID)
		}
	}

	queryStr = fmt.Sprintf(
		`SELECT cr.%s, cr.%s, cr.%s, cr.%s, cr.%s, cr.%s, cr.%s, cr.%s, cr.%s, cr.%s, cr.%s, cr.%s, cr.%s,
		cf.%s as cfID, cf.%s as cfUUID, cf.%s as cfName
		 FROM (SELECT * FROM %s %s) as cr
		 LEFT JOIN %s as rcrcf ON cr.%s = rcrcf.%s
		 %s as cf ON rcrcf.%s = cf.%s`,
		fldID,
		fldUUID,
		fldName,
		CRMethod,
		CRScore,
		fldEnterprise,
		CRDescription,
		CRMin,
		CRMax,
		CRSeverity,
		fldIsDelete,
		fldCreateTime,
		fldUpdateTime,
		fldID,
		fldUUID,
		fldName,
		tblConversationRule,
		conditionStr,
		tblRelCRCF,
		fldID,
		CRCFRID,
		cfCondition,
		CRCFCFID,
		fldID,
	)
	return
}

func (dao *ConversationRuleSqlDaoImpl) CountBy(filter *ConversationRuleFilter, sql SqlLike) (total int64, err error) {
	queryStr, values := queryConversationRulesSQLBy(filter)
	queryStr = fmt.Sprintf("SELECT COUNT(cr.%s) FROM (%s) as cr", fldID, queryStr)

	rows, err := sql.Query(queryStr, values...)
	if err != nil {
		err = fmt.Errorf("error while count rules in dao.CountBy, err: %s", err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&total)
	}
	return
}

func (dao *ConversationRuleSqlDaoImpl) GetBy(filter *ConversationRuleFilter, sqlLike SqlLike) (rules []ConversationRule, err error) {
	queryStr, values := queryConversationRulesSQLBy(filter)

	rows, err := sqlLike.Query(queryStr, values...)
	if err != nil {
		err = fmt.Errorf("error while get rules in dao.GetBy, err: %s", err.Error())
		return
	}
	defer rows.Close()

	rules = []ConversationRule{}
	var cRule *ConversationRule
	for rows.Next() {
		rule := ConversationRule{}
		var flowUUID *string
		var flowName *string
		var flowID *int64
		err = rows.Scan(
			&rule.ID,
			&rule.UUID,
			&rule.Name,
			&rule.Method,
			&rule.Score,
			&rule.Enterprise,
			&rule.Description,
			&rule.Min,
			&rule.Max,
			&rule.Severity,
			&rule.Deleted,
			&rule.CreateTime,
			&rule.UpdateTime,
			&flowID,
			&flowUUID,
			&flowName,
		)

		if err != nil {
			err = fmt.Errorf("error while scan rule in dao.GetBy, err: %s", err.Error())
			return
		}

		if cRule == nil || cRule.UUID != rule.UUID {
			if cRule != nil {
				completeness := RuleCompleteness{
					RuleCompleted:       int8(0),
					HasDescription:      cRule.Description != "",
					HasConversationFlow: len(cRule.Flows) > 0,
					SetenceCompleted:    int8(0),
				}
				cRule.Completeness = &completeness

				rules = append(rules, *cRule)
			}
			cRule = &rule
			cRule.Flows = []SimpleConversationFlow{}
		}
		if flowUUID != nil && flowName != nil {
			flow := SimpleConversationFlow{
				ID:   *flowID,
				UUID: *flowUUID,
				Name: *flowName,
			}
			cRule.Flows = append(cRule.Flows, flow)
		}
	}

	if cRule != nil {
		completeness := RuleCompleteness{
			RuleCompleted:       int8(0),
			HasDescription:      cRule.Description != "",
			HasConversationFlow: len(cRule.Flows) > 0,
			SetenceCompleted:    int8(0),
		}
		cRule.Completeness = &completeness

		rules = append(rules, *cRule)
	}
	return
}

func (dao *ConversationRuleSqlDaoImpl) GetByFlowID(flowID []int64, sqlLike SqlLike) (rules []ConversationRule, err error) {
	rules = []ConversationRule{}
	if len(flowID) == 0 {
		return
	}
	builder := NewWhereBuilder(andLogic, "")
	builder.In(CRCFCFID, int64ToWildCard(flowID...))

	conditionStr, values := builder.Parse()

	queryStr := fmt.Sprintf(
		"SELECT %s FROM %s WHERE %s",
		CRCFRID,
		tblRelCRCF,
		conditionStr,
	)

	rows, err := sqlLike.Query(queryStr, values...)
	if err != nil {
		err = fmt.Errorf("error while query rule id in dao.GetByFlowID, err: %s", err.Error())
		return
	}
	defer rows.Close()

	ruleID := []uint64{}
	for rows.Next() {
		var rid uint64
		err = rows.Scan(&rid)
		if err != nil {
			err = fmt.Errorf("error while scan rule id in dao.GetByFlowID, err: %s", err.Error())
			return
		}
		ruleID = append(ruleID, rid)
	}

	if len(ruleID) == 0 {
		return
	}

	filter := &ConversationRuleFilter{
		ID: ruleID,
	}
	rules, err = dao.GetBy(filter, sqlLike)
	return
}

func (dao *ConversationRuleSqlDaoImpl) Delete(id string, sqlLike SqlLike) (err error) {
	deleteStr := fmt.Sprintf("UPDATE %s SET %s=? WHERE %s=?", tblConversationRule, fldIsDelete, fldUUID)

	_, err = sqlLike.Exec(deleteStr, 1, id)
	if err != nil {
		err = fmt.Errorf("error while delete rule in dao.Delete, err: %s", err.Error())
		return
	}
	return
}

func (dao *ConversationRuleSqlDaoImpl) DeleteMany(uuid []string, sqlLike SqlLike) (err error) {
	if len(uuid) == 0 {
		return
	}
	builder := whereBuilder{
		ConcatLogic: andLogic,
		conditions:  []string{},
		data:        []interface{}{},
	}

	builder.In(fldUUID, stringToWildCard(uuid...))
	conditionStr, values := builder.Parse()

	deleteStr := fmt.Sprintf(
		"UPDATE %s SET %s=1 WHERE %s",
		tblConversationRule,
		fldIsDelete,
		conditionStr,
	)

	_, err = sqlLike.Exec(deleteStr, values...)
	if err != nil {
		logger.Error.Printf("error while delete rules in dao.DeleteMany, sql: %s\n", deleteStr)
		logger.Error.Printf("error while delete rules in dao.DeleteMany, values: %+v\n", values)
		err = fmt.Errorf("error while delete rules in dao.DeleteMany, err: %s", err.Error())
	}
	return
}
