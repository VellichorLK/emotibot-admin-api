package model

import (
	"emotibot.com/emotigo/pkg/logger"
	"fmt"
	"strings"
)

type ConversationRule struct {
	ID          int64
	UUID        string
	Name        string
	Method      int8
	Score       int
	Enterprise  string
	Description string
	Min         int
	Max         int
	Severity    int8
	Type        int8
	Flows       []SimpleConversationFlow
	CreateTime  int64
	UpdateTime  int64
}

type ConversationRuleFilter struct {
	UUID       []string
	Name       string
	Method     int8 // TODO
	Enterprise string
	Severity   int8
	IsDeleted  int8
}

type ConversationRuleDao interface {
	Create(rule *ConversationRule, sql SqlLike) (*ConversationRule, error)
	CountBy(filter *ConversationRuleFilter, sql SqlLike) (int64, error)
	GetBy(filter *ConversationRuleFilter, sql SqlLike) ([]ConversationRule, error)
	Delete(id string, sql SqlLike) error
}

type ConversationRuleSqlDaoImpl struct{}

func (dao *ConversationRuleSqlDaoImpl) Create(rule *ConversationRule, sql SqlLike) (createdRule *ConversationRule, err error) {
	fields := []string{
		fldUUID,
		fldName,
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

	values := []interface{}{
		rule.UUID,
		rule.Name,
		rule.Enterprise,
		rule.Score,
		rule.Description,
		rule.Min,
		rule.Max,
		rule.Severity,
		rule.CreateTime,
		rule.UpdateTime,
	}
	valueStr := fmt.Sprintf("? %s", strings.Repeat(", ?", len(values)-1))

	insertStr := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", tblConversationRule, fieldStr, valueStr)
	result, err := sql.Exec(insertStr, values...)
	if err != nil {
		err = fmt.Errorf("error while insert conversation rule in dao.Create, err: %s", err.Error())
		return
	}

	ruleID, err := result.LastInsertId()
	if err != nil {
		err = fmt.Errorf("error while get rule id in dao.Create, err: %s", err.Error())
		return
	}

	if len(rule.Flows) > 0 {
		fieldStr = fmt.Sprintf("%s, %s", CRCFRID, CRCFCFID)
		valueStr = "(?, ?)"
		values = []interface{}{
			ruleID,
			rule.Flows[0].ID,
		}

		for _, flow := range rule.Flows[1:] {
			valueStr = fmt.Sprintf("%s, %s", valueStr, "(?, ?)")
			values = append(values, ruleID, flow.ID)
		}

		insertStr = fmt.Sprintf("INSERT INTO %s (%s) VALUES %s", tblRelCRCF, fieldStr, valueStr)

		_, err = sql.Exec(insertStr, values...)
		if err != nil {
			err = fmt.Errorf("error while insert rule flow relation in dao.Create, err: %s", err.Error())
			return
		}
	}
	rule.ID = ruleID
	createdRule = rule
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

	queryStr = fmt.Sprintf(
		`SELECT cr.%s, cr.%s, cr.%s, cr.%s, cr.%s, cr.%s, cr.%s, cr.%s, cr.%s, cr.%s, cr.%s, cr.%s,
		cf.%s as cfUUID, cf.%s as cfName
		 FROM (SELECT * FROM %s %s) as cr
		 LEFT JOIN %s as rcrcf ON cr.%s = rcrcf.%s
		 LEFT JOIN %s as cf ON rcrcf.%s = cf.%s`,
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
		fldCreateTime,
		fldUpdateTime,
		fldUUID,
		fldName,
		tblConversationRule,
		conditionStr,
		tblRelCRCF,
		fldID,
		CRCFRID,
		tblConversationflow,
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

func (dao *ConversationRuleSqlDaoImpl) GetBy(filter *ConversationRuleFilter, sql SqlLike) (rules []ConversationRule, err error) {
	queryStr, values := queryConversationRulesSQLBy(filter)

	rows, err := sql.Query(queryStr, values...)
	if err != nil {
		err = fmt.Errorf("error while get rules in dao.GetBy, err: %s", err.Error())
		return
	}
	defer rows.Close()

	rules = []ConversationRule{}
	var cRule *ConversationRule
	for rows.Next() {
		rule := ConversationRule{}
		flow := SimpleConversationFlow{}

		rows.Scan(
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
			&rule.CreateTime,
			&rule.UpdateTime,
			&flow.UUID,
			&flow.Name,
		)

		if cRule == nil || cRule.UUID != rule.UUID {
			if cRule != nil {
				rules = append(rules, *cRule)
			}
			cRule = &rule
		}
		cRule.Flows = append(cRule.Flows, flow)
		logger.Info.Printf("rules.flows: %+v\n", cRule.Flows)
	}

	if cRule != nil {
		rules = append(rules, *cRule)
	}
	return
}

func (dao *ConversationRuleSqlDaoImpl) Delete(id string, sql SqlLike) (err error) {
	deleteStr := fmt.Sprintf("UPDATE %s SET %s=? WHERE %s=?", tblConversationRule, fldIsDelete, fldUUID)

	_, err = sql.Exec(deleteStr, 1, id)
	if err != nil {
		err = fmt.Errorf("error while delete rule in dao.Delete, err: %s", err.Error())
		return
	}
	return
}