package model

import (
	"emotibot.com/emotigo/pkg/logger"
	"fmt"
	"strings"
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
	Deleted        int8
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
	IsDelete   *int8
	SGUUID     []string
}

type ConversationFlowDao interface {
	Create(flow *ConversationFlow, sqlLike SqlLike) (*ConversationFlow, error)
	CountBy(filter *ConversationFlowFilter, sqlLike SqlLike) (int64, error)
	GetBy(filter *ConversationFlowFilter, sqlLike SqlLike) ([]ConversationFlow, error)
	Update(id string, flow *ConversationFlow, sqlLike SqlLike) (*ConversationFlow, error)
	Delete(id string, sqlLike SqlLike) error
	GetBySentenceGroupID([]int64, SqlLike) ([]ConversationFlow, error)
	CreateMany([]ConversationFlow, SqlLike) error
	DeleteMany([]string, SqlLike) error
}

type ConversationFlowSqlDaoImpl struct{}

func getConversationFlowInsertSQL(flows []ConversationFlow) (insertStr string, values []interface{}) {
	values = []interface{}{}
	if len(flows) == 0 {
		return
	}

	fields := []string{
		fldUUID,
		fldName,
		fldEnterprise,
		fldMin,
		CFExpression,
		fldCreateTime,
		fldUpdateTime,
	}
	fieldStr := strings.Join(fields, ", ")

	insertStr = fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES",
		tblConversationflow,
		fieldStr,
	)

	variableStr := fmt.Sprintf("(?%s)", strings.Repeat(", ?", len(fields)-1))
	valueStr := ""
	for _, flow := range flows {
		values = append(
			values,
			flow.UUID,
			flow.Name,
			flow.Enterprise,
			flow.Min,
			flow.Expression,
			flow.CreateTime,
			flow.UpdateTime,
		)
		valueStr = fmt.Sprintf("%s%s,", valueStr, variableStr)
	}
	// remove last comma
	valueStr = valueStr[:len(valueStr)-1]
	insertStr = fmt.Sprintf("%s %s", insertStr, valueStr)

	return
}

func getConversationFlowRelationInsertSQL(flows []ConversationFlow) (insertStr string, values []interface{}) {
	values = []interface{}{}
	insertStr = fmt.Sprintf(
		"INSERT INTO %s (%s, %s) VALUES",
		tblRelCFSG,
		RCFSGCFID,
		RCFSGSGID,
	)

	variableStr := "(?, ?)"
	valueStr := ""
	for _, flow := range flows {
		for _, sentenceGroup := range flow.SentenceGroups {
			values = append(values, flow.ID, sentenceGroup.ID)
			valueStr = fmt.Sprintf("%s%s,", valueStr, variableStr)
		}
	}

	if len(values) == 0 {
		return
	}
	// remove last comma
	valueStr = valueStr[:len(valueStr)-1]

	insertStr = fmt.Sprintf("%s %s", insertStr, valueStr)
	return
}

func (dao *ConversationFlowSqlDaoImpl) Create(flow *ConversationFlow, sqlLike SqlLike) (createdFlow *ConversationFlow, err error) {
	insertStr, values := getConversationFlowInsertSQL([]ConversationFlow{*flow})

	result, err := sqlLike.Exec(insertStr, values...)
	if err != nil {
		err = fmt.Errorf("error while insert flow in dao.Create, err: %s", err.Error())
		return
	}

	flowID, err := result.LastInsertId()
	if err != nil {
		err = fmt.Errorf("error while get flow id in dao.Create, err: %s", err.Error())
		return
	}
	flow.ID = flowID

	// create conversation flow to sentence group relation
	if len(flow.SentenceGroups) > 0 {
		insertStr, values = getConversationFlowRelationInsertSQL([]ConversationFlow{*flow})
		_, err = sqlLike.Exec(insertStr, values...)
		if err != nil {
			err = fmt.Errorf("error while insert flow to sentence group relation in dao.Create, err: %s", err.Error())
			return
		}
	}

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

	if filter.IsDelete != nil {
		conditions = append(conditions, fmt.Sprintf("%s=?", fldIsDelete))
		values = append(values, *filter.IsDelete)
	}

	if len(conditions) > 0 {
		conditionStr = fmt.Sprintf("%s %s", conditionStr, strings.Join(conditions, " and "))
	} else {
		conditionStr = ""
	}

	sgCondition := fmt.Sprintf("LEFT JOIN %s", tblSetnenceGroup)
	if len(filter.SGUUID) > 0 {
		sgCondition = fmt.Sprintf(
			"INNER JOIN (SELECT * FROM %s WHERE %s IN (?%s))",
			tblSetnenceGroup,
			fldUUID,
			strings.Repeat(", ?", len(filter.SGUUID)-1),
		)
		for _, sgUUID := range filter.SGUUID {
			values = append(values, sgUUID)
		}
	}

	queryStr = fmt.Sprintf(
		`SELECT cf.%s, cf.%s, cf.%s, cf.%s, cf.%s, cf.%s, cf.%s, cf.%s, cf.%s, sg.%s as sgID, sg.%s as sgUUID, sg.%s as sgName
		 FROM (SELECT * FROM %s %s) as cf
		 LEFT JOIN %s as rcfsg ON cf.%s = rcfsg.%s
		 %s as sg ON rcfsg.%s = sg.%s`,
		fldID,
		fldUUID,
		fldName,
		CRMin,
		fldEnterprise,
		CFExpression,
		fldCreateTime,
		fldUpdateTime,
		fldIsDelete,
		fldID,
		fldUUID,
		fldName,
		tblConversationflow,
		conditionStr,
		tblRelCFSG,
		fldID,
		RCFSGCFID,
		sgCondition,
		RCFSGSGID,
		fldID,
	)
	return
}

func (dao *ConversationFlowSqlDaoImpl) CountBy(filter *ConversationFlowFilter, sqlLike SqlLike) (total int64, err error) {
	queryStr, values := queryConversationFlowsSQLBy(filter)
	queryStr = fmt.Sprintf("SELECT COUNT(cf.%s) FROM (%s) as cf", fldID, queryStr)

	rows, err := sqlLike.Query(queryStr, values...)
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

func (dao *ConversationFlowSqlDaoImpl) GetBy(filter *ConversationFlowFilter, sqlLike SqlLike) (flows []ConversationFlow, err error) {
	queryStr, values := queryConversationFlowsSQLBy(filter)

	rows, err := sqlLike.Query(queryStr, values...)
	if err != nil {
		err = fmt.Errorf("error while get conversation flows in dao.GetBy, err: %s", err.Error())
		return
	}
	defer rows.Close()

	flows = []ConversationFlow{}
	var cFlow *ConversationFlow // current conversatin flow
	for rows.Next() {
		flow := ConversationFlow{}
		var sgID *int64
		var sgUUID *string
		var sgName *string

		err = rows.Scan(
			&flow.ID,
			&flow.UUID,
			&flow.Name,
			&flow.Min,
			&flow.Enterprise,
			&flow.Expression,
			&flow.CreateTime,
			&flow.UpdateTime,
			&flow.Deleted,
			&sgID,
			&sgUUID,
			&sgName,
		)

		if err != nil {
			err = fmt.Errorf("error while scan flow in dao.GetBy, err: %s", err.Error())
			return
		}

		if cFlow == nil || cFlow.UUID != flow.UUID {
			if cFlow != nil {
				flows = append(flows, *cFlow)
			}
			cFlow = &flow
		}
		if sgUUID != nil && sgName != nil {
			sg := SimpleSentenceGroup{
				ID:   *sgID,
				UUID: *sgUUID,
				Name: *sgName,
			}
			cFlow.SentenceGroups = append(cFlow.SentenceGroups, sg)
		}
	}

	if cFlow != nil {
		flows = append(flows, *cFlow)
	}
	return
}

func (dao *ConversationFlowSqlDaoImpl) Update(id string, flow *ConversationFlow, sqlLike SqlLike) (updatedFlow *ConversationFlow, err error) {
	return
}

func (dao *ConversationFlowSqlDaoImpl) Delete(id string, sqlLike SqlLike) (err error) {
	deleteStr := fmt.Sprintf("UPDATE %s SET %s=? WHERE %s=?", tblConversationflow, fldIsDelete, fldUUID)

	_, err = sqlLike.Exec(deleteStr, 1, id)
	if err != nil {
		err = fmt.Errorf("error while delete conversation flow in dao.Delete, reason: %s", err.Error())
		return
	}
	return
}

func (dao *ConversationFlowSqlDaoImpl) GetBySentenceGroupID(sgid []int64, sqlLike SqlLike) (flows []ConversationFlow, err error) {
	flows = []ConversationFlow{}
	if len(sgid) == 0 {
		return
	}

	builder := NewWhereBuilder(andLogic, "")
	builder.In(RCFSGSGID, int64ToWildCard(sgid...))

	conditionStr, values := builder.Parse()
	queryStr := fmt.Sprintf(
		"SELECT %s FROM %s WHERE %s",
		RCFSGCFID,
		tblRelCFSG,
		conditionStr,
	)

	logger.Info.Printf("queryStr: %s\n", queryStr)
	logger.Info.Printf("values: %+v\n", values)
	rows, err := sqlLike.Query(queryStr, values...)
	if err != nil {
		logger.Error.Printf("error while get flow id in dao.GetBySentenceGroupID, sql: %s", queryStr)
		logger.Error.Printf("error while get flow id in dao.GetBySentenceGroupID, values: %+v", values)
		err = fmt.Errorf("error while get flow id in dao.GetBySentenceGroupID, err: %s", err.Error())
		return
	}

	cfid := []uint64{}
	for rows.Next() {
		var id uint64
		rows.Scan(&id)
		cfid = append(cfid, id)
	}

	if len(cfid) == 0 {
		return
	}

	filter := &ConversationFlowFilter{
		ID: cfid,
	}

	flows, err = dao.GetBy(filter, sqlLike)
	return
}

func (dao *ConversationFlowSqlDaoImpl) CreateMany(flows []ConversationFlow, sqlLike SqlLike) (err error) {
	if len(flows) == 0 {
		return
	}

	insertStr, values := getConversationFlowInsertSQL(flows)

	_, err = sqlLike.Exec(insertStr, values...)
	if err != nil {
		logger.Error.Printf("error while insert flows in dao.CreateMany, sql: %s", insertStr)
		logger.Error.Printf("error while insert flows in dao.CreateMany, values: %s", values)
		err = fmt.Errorf("error while insert flows in dao.CreateMany, err: %s", err.Error())
		return
	}

	// update flow ids
	flowUUID := []string{}
	for _, flow := range flows {
		flowUUID = append(flowUUID, flow.UUID)
	}

	deleted := int8(0)
	filter := &ConversationFlowFilter{
		UUID:     flowUUID,
		IsDelete: &deleted,
	}

	newFlows, err := dao.GetBy(filter, sqlLike)
	if err != nil {
		return
	}

	flowMap := map[string]int64{}
	for _, newFlow := range newFlows {
		flowMap[newFlow.UUID] = newFlow.ID
	}

	for idx := range flows {
		flow := &flows[idx]
		flow.ID = flowMap[flow.UUID]
	}

	// create flow sentence group relation
	insertStr, values = getConversationFlowRelationInsertSQL(flows)
	if len(values) == 0 {
		return
	}

	_, err = sqlLike.Exec(insertStr, values...)
	if err != nil {
		logger.Error.Printf("error while insert flow sentence group relation in dao.CreateMany, sql: %s", insertStr)
		logger.Error.Printf("error while insert flow sentence group relation in dao.CreateMany, values: %s", values)
		err = fmt.Errorf("error while insert flow sentence group relation in dao.CreateMany, err: %s", err.Error())
	}
	return
}

func (dao *ConversationFlowSqlDaoImpl) DeleteMany(uuids []string, sqlLike SqlLike) (err error) {
	builder := NewWhereBuilder(andLogic, "")

	builder.In(fldUUID, stringToWildCard(uuids...))

	conditionStr, values := builder.Parse()

	deleteStr := fmt.Sprintf(
		"UPDATE %s SET %s = 1 WHERE %s",
		tblConversationflow,
		fldIsDelete,
		conditionStr,
	)

	_, err = sqlLike.Exec(deleteStr, values...)
	if err != nil {
		logger.Error.Printf("error while delete flows in dao.DeleteMany, sql: %s", deleteStr)
		logger.Error.Printf("error while delete flows in dao.DeleteMany, values: %s", values)
		err = fmt.Errorf("error while delete flows in dao.DeleteMany, err: %s", err.Error())
	}
	return
}
