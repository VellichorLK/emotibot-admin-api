package model

import (
	_ "emotibot.com/emotigo/pkg/logger"
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
	Method     int8
	Enterprise string
	Severity   int8
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
		fieldStr = fmt.Sprintf("%s, %s", CRCFCFID, CRCFRID)
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

func (dao *ConversationRuleSqlDaoImpl) CountBy(filter *ConversationRuleFilter, sql SqlLike) (total int64, err error) {
	return
}

func (dao *ConversationRuleSqlDaoImpl) GetBy(filter *ConversationRuleFilter, sql SqlLike) (rules []ConversationRule, err error) {
	return
}

func (dao *ConversationRuleSqlDaoImpl) Delete(id string, sql SqlLike) (err error) {
	return
}
