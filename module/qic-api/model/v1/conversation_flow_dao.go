package model

import (
	"fmt"
	"strings"

	_ "emotibot.com/emotigo/pkg/logger"
)

type ConversationFlow struct {
	ID             int64
	UUID           string
	Name           string
	Enterprise     string
	Expression     string
	Type           string
	SentenceGroups []SimpleSentenceGroup
	CreateTime     int64
	UpdateTime     int64
	Min            int
}

type SimpleConversationFlow struct {
	ID   int64  `json:"-"`
	UUID string `json:"flow_id"`
	Name string `json:"flow_name"`
}

type ConversationFlowFilter struct {
	UUID       []string
	ID         []uint64
	Name       string
	Enterprise string
	IsDelete   int8
}

type ConversationFlowDao interface {
	Create(flow *ConversationFlow, sql SqlLike) (*ConversationFlow, error)
	CountBy(filter *ConversationFlowFilter, sql SqlLike) (int64, error)
	GetBy(filter *ConversationFlowFilter, sql SqlLike) ([]ConversationFlow, error)
	Update(id string, flow *ConversationFlow, sql SqlLike) (*ConversationFlow, error)
	Delete(id string, sql SqlLike) error
}

type ConversationFlowSqlDaoImpl struct{}

func (dao *ConversationFlowSqlDaoImpl) Create(flow *ConversationFlow, sql SqlLike) (createdFlow *ConversationFlow, err error) {
	fields := []string{
		fldUUID,
		fldName,
		fldEnterprise,
		CRMin,
		CFExpression,
		fldCreateTime,
		fldUpdateTime,
		fldMin,
	}
	fieldStr := strings.Join(fields, ", ")

	values := []interface{}{
		flow.UUID,
		flow.Name,
		flow.Enterprise,
		flow.Min,
		flow.Expression,
		flow.CreateTime,
		flow.UpdateTime,
		flow.Min,
	}
	valueStr := "?"
	valueStr = fmt.Sprintf("%s %s", valueStr, strings.Repeat(", ?", len(values)-1))

	insertStr := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", tblConversationflow, fieldStr, valueStr)

	result, err := sql.Exec(insertStr, values...)
	if err != nil {
		err = fmt.Errorf("error while insert flow in dao.Create, err: %s", err.Error())
		return
	}

	flowID, err := result.LastInsertId()
	if err != nil {
		err = fmt.Errorf("error while get flow id in dao.Create, err: %s", err.Error())
		return
	}

	// create conversation flow to sentence group relation
	if len(flow.SentenceGroups) > 0 {
		fields = []string{
			RCFSGCFID,
			RCFSGSGID,
		}
		fieldStr = strings.Join(fields, ", ")

		valueStr = "(?, ?)"
		values = []interface{}{
			flowID,
			flow.SentenceGroups[0].ID,
		}

		for idx := range flow.SentenceGroups[1:] {
			valueStr += ", (?, ?)"
			values = append(values, flowID, flow.SentenceGroups[idx].ID)
		}

		insertStr = fmt.Sprintf("INSERT INTO %s (%s) VALUES %s", tblRelCFSG, fieldStr, valueStr)
		_, err = sql.Exec(insertStr, values...)
		if err != nil {
			err = fmt.Errorf("error while insert flow to sentence group relation in dao.Create, err: %s", err.Error())
			return
		}
	}
	flow.ID = flowID
	createdFlow = flow
	return
}

func queryConversationFlowsSQLBy(filter *ConversationFlowFilter) (queryStr string, values []interface{}) {
	values = []interface{}{}
	conditionStr := "WHERE "
	conditions := []string{}

	if len(filter.UUID) > 0 {
		idStr := fmt.Sprintf("? %s", strings.Repeat(", ?", len(filter.UUID)-1))
		// conditionStr = fmt.Sprintf("%s IN(%s)", fldUUID, idStr)
		conditions = append(conditions, fmt.Sprintf("%s IN(%s)", fldUUID, idStr))

		for _, uuid := range filter.UUID {
			values = append(values, uuid)
		}
	}

	if len(filter.ID) > 0 {
		idStr := fmt.Sprintf("? %s", strings.Repeat(", ?", len(filter.ID)-1))
		conditions = append(conditions, fmt.Sprintf("%s IN(%s)", fldID, idStr))
		for _, uuid := range filter.ID {
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

	if filter.IsDelete != -1 {
		conditions = append(conditions, fmt.Sprintf("%s=?", fldIsDelete))
		values = append(values, filter.IsDelete)
	}

	if len(conditions) > 0 {
		conditionStr = fmt.Sprintf("%s %s", conditionStr, strings.Join(conditions, " and "))
	} else {
		conditionStr = ""
	}

	queryStr = fmt.Sprintf(
		`SELECT cf.%s, cf.%s, cf.%s, cf.%s, cf.%s, cf.%s, cf.%s, cf.%s, sg.%s as sgUUID, sg.%s as sgName
		 FROM (SELECT * FROM %s %s) as cf
		 LEFT JOIN %s as rcfsg ON cf.%s = rcfsg.%s
		 LEFT JOIN %s as sg ON rcfsg.%s = sg.%s`,
		fldID,
		fldUUID,
		fldName,
		CRMin,
		fldEnterprise,
		CFExpression,
		fldCreateTime,
		fldUpdateTime,
		fldMin,
		fldUUID,
		fldName,
		tblConversationflow,
		conditionStr,
		tblRelCFSG,
		fldID,
		RCFSGCFID,
		tblSetnenceGroup,
		RCFSGSGID,
		fldID,
	)
	return
}

func (dao *ConversationFlowSqlDaoImpl) CountBy(filter *ConversationFlowFilter, sql SqlLike) (total int64, err error) {
	queryStr, values := queryConversationFlowsSQLBy(filter)
	queryStr = fmt.Sprintf("SELECT COUNT(cf.%s) FROM (%s) as cf", fldID, queryStr)

	rows, err := sql.Query(queryStr, values...)
	if err != nil {
		err = fmt.Errorf("error while count conversation flows in dao.CountBy, err: %s", err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&total)
	}
	return
}

func (dao *ConversationFlowSqlDaoImpl) GetBy(filter *ConversationFlowFilter, sql SqlLike) (flows []ConversationFlow, err error) {
	queryStr, values := queryConversationFlowsSQLBy(filter)

	rows, err := sql.Query(queryStr, values...)
	if err != nil {
		err = fmt.Errorf("error while get conversation flows in dao.GetBy, err: %s", err.Error())
		return
	}
	defer rows.Close()

	flows = []ConversationFlow{}
	var cFlow *ConversationFlow // current conversatin flow
	for rows.Next() {
		flow := ConversationFlow{}
		sg := SimpleSentenceGroup{}

		rows.Scan(
			&flow.ID,
			&flow.UUID,
			&flow.Name,
			&flow.Min,
			&flow.Enterprise,
			&flow.Expression,
			&flow.CreateTime,
			&flow.UpdateTime,
			&flow.Min,
			&sg.UUID,
			&sg.Name,
		)

		if cFlow == nil || cFlow.UUID != flow.UUID {
			if cFlow != nil {
				flows = append(flows, *cFlow)
			}
			cFlow = &flow
		}
		if sg.UUID != "" && sg.Name != "" {
			cFlow.SentenceGroups = append(cFlow.SentenceGroups, sg)
		}
	}

	if cFlow != nil {
		flows = append(flows, *cFlow)
	}
	return
}

func (dao *ConversationFlowSqlDaoImpl) Update(id string, flow *ConversationFlow, sql SqlLike) (updatedFlow *ConversationFlow, err error) {
	return
}

func (dao *ConversationFlowSqlDaoImpl) Delete(id string, sql SqlLike) (err error) {
	deleteStr := fmt.Sprintf("UPDATE %s SET %s=? WHERE %s=?", tblConversationflow, fldIsDelete, fldUUID)

	_, err = sql.Exec(deleteStr, 1, id)
	if err != nil {
		err = fmt.Errorf("error while delete conversation flow in dao.Delete, reason: %s", err.Error())
		return
	}
	return
}
