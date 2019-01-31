package model

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"emotibot.com/emotigo/pkg/logger"
)

var (
	ErrNilSqlLike = errors.New("SqlLike can not be nil")
)

type SimpleSentenceGroup struct {
	ID   int64  `json:"-"`
	UUID string `json:"sg_id"`
	Name string `json:"sg_name"`
}

type SentenceGroup struct {
	ID         int64
	UUID       string
	Name       string
	Role       int
	Position   int
	Distance   int
	Sentences  []SimpleSentence
	Enterprise string
	CreateTime int64
	UpdateTime int64
	Deleted    int8
}

type SentenceGroupFilter struct {
	UUID        []string
	ID          []uint64
	Name        string
	Role        *int
	Position    *int
	Distance    int
	Enterprise  string
	CreateTime  *time.Time
	UpdateTime  *time.Time
	Page        int
	Limit       int
	IsDelete    *int8
	SetenceUUID []string
}

type SentenceGroupsSqlDao interface {
	Create(group *SentenceGroup, sqlLike SqlLike) (*SentenceGroup, error)
	CountBy(filter *SentenceGroupFilter, sqlLike SqlLike) (int64, error)
	GetBy(filter *SentenceGroupFilter, sqlLike SqlLike) ([]SentenceGroup, error)
	GetBySentenceID(id []int64, sqlLike SqlLike) ([]SentenceGroup, error)
	Update(id string, group *SentenceGroup, sqlLike SqlLike) (*SentenceGroup, error)
	Delete(id string, sqlLike SqlLike) error
	CreateMany([]SentenceGroup, SqlLike) error
	DeleteMany([]string, SqlLike) error
}

type SentenceGroupsSqlDaoImpl struct{}

func getSentenceGroupInsertSQL(groups []SentenceGroup) (insertStr string, values []interface{}) {
	values = []interface{}{}
	if len(groups) == 0 {
		return
	}

	fields := []string{
		fldIsDelete,
		fldName,
		fldEnterprise,
		SGRole,
		SGPoistion,
		SGRange,
		fldUUID,
		fldCreateTime,
		fldUpdateTime,
	}
	fieldStr := strings.Join(fields, "`, `")
	fieldStr = fmt.Sprintf("`%s`", fieldStr)

	insertStr = fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES",
		tblSetnenceGroup,
		fieldStr,
	)

	variableStr := fmt.Sprintf("(?%s)", strings.Repeat(", ?", len(fields)-1))
	valueStr := ""
	for _, group := range groups {
		values = append(
			values,
			0,
			group.Name,
			group.Enterprise,
			group.Role,
			group.Position,
			group.Distance,
			group.UUID,
			group.CreateTime,
			group.UpdateTime,
		)
		valueStr = fmt.Sprintf("%s%s,", valueStr, variableStr)
	}
	// remove last comma
	valueStr = valueStr[:len(valueStr)-1]

	insertStr = fmt.Sprintf("%s %s", insertStr, valueStr)
	return
}

func getSentenceGroupRelationInsertSQL(groups []SentenceGroup) (insertStr string, values []interface{}) {
	values = []interface{}{}
	if len(groups) == 0 {
		return
	}

	insertStr = fmt.Sprintf("INSERT INTO %s (%s, %s) VALUES", tblRelSGS, RSGSSGID, RSGSSID)

	variableStr := "(?, ?)"
	valueStr := ""
	for _, group := range groups {
		for _, sentence := range group.Sentences {
			values = append(values, group.ID, sentence.ID)
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

func (dao *SentenceGroupsSqlDaoImpl) Create(group *SentenceGroup, sqlLike SqlLike) (createdGroup *SentenceGroup, err error) {
	if sqlLike == nil {
		err = ErrNilSqlLike
		return
	}

	insertStr, values := getSentenceGroupInsertSQL([]SentenceGroup{*group})
	logger.Info.Printf("6\n")

	result, err := sqlLike.Exec(insertStr, values...)
	logger.Info.Printf("7\n")
	if err != nil {
		logger.Error.Printf("error while insert sentence group in dao.Create, sql: %s", insertStr)
		logger.Error.Printf("error while insert sentence group in dao.Create, values: %s", values)
		logger.Error.Printf("error while insert sentence group in dao.Create, err: %s", err.Error())
		err = fmt.Errorf("error while insert sentence group in dao.Create, err: %s", err.Error())
		return
	}

	groupID, err := result.LastInsertId()
	logger.Info.Printf("8\n")
	if err != nil {
		err = fmt.Errorf("error while get group id in dao.Create, err: %s", err.Error())
		return

	}
	group.ID = groupID

	if len(group.Sentences) > 0 {
		// create sentence group to sentence relation
		insertStr, values = getSentenceGroupRelationInsertSQL([]SentenceGroup{*group})

		_, err = sqlLike.Exec(insertStr, values...)
		if err != nil {
			err = fmt.Errorf("error while insert sentence group to sentence relationin dao.Create, err: %s", err.Error())
			return
		}
	}
	createdGroup = group
	return
}

func querySentenceGroupsSQLBy(filter *SentenceGroupFilter) (queryStr string, values []interface{}) {
	values = []interface{}{}

	conditionStr := ""
	conditions := []string{}

	if len(filter.UUID) > 0 {
		uuidStr := fmt.Sprintf("? %s", strings.Repeat(", ?", len(filter.UUID)-1))
		for _, uuid := range filter.UUID {
			values = append(values, uuid)
		}
		conditionStr = fmt.Sprintf("%s IN (%s)", fldUUID, uuidStr)
	}

	if len(filter.ID) > 0 {
		idStr := fmt.Sprintf("? %s", strings.Repeat(", ?", len(filter.ID)-1))
		for _, uuid := range filter.ID {
			values = append(values, uuid)
		}
		if conditionStr != "" {
			conditionStr += " AND "
		}
		conditionStr = fmt.Sprintf("%s %s IN (%s)", conditionStr, fldID, idStr)
	}

	if filter.Name != "" {
		conditions = append(conditions, fldName)
		values = append(values, filter.Name)
	}

	if filter.Enterprise != "" {
		conditions = append(conditions, fldEnterprise)
		values = append(values, filter.Enterprise)
	}

	if filter.Role != nil {
		conditions = append(conditions, SGRole)
		values = append(values, *filter.Role)
	}

	if filter.Position != nil {
		conditions = append(conditions, SGPoistion)
		values = append(values, *filter.Position)
	}

	if filter.Distance != 0 {
		conditions = append(conditions, SGRange)
		values = append(values, filter.Distance)
	}

	if filter.IsDelete != nil {
		conditions = append(conditions, fldIsDelete)
		values = append(values, *filter.IsDelete)
	}

	for _, condition := range conditions {
		if conditionStr != "" {
			conditionStr = fmt.Sprintf("%s and %s=?", conditionStr, condition)
		} else {
			conditionStr = fmt.Sprintf("%s=?", condition)
		}
	}

	if conditionStr != "" {
		conditionStr = "WHERE " + conditionStr
	}

	sentenceCondition := fmt.Sprintf("LEFT JOIN %s", tblSentence)
	if len(filter.SetenceUUID) > 0 {
		sentenceCondition = fmt.Sprintf(
			"INNER JOIN (SELECT * FROM %s WHERE %s IN (%s))",
			tblSentence,
			fldUUID,
			fmt.Sprintf("?%s", strings.Repeat(", ?", len(filter.SetenceUUID)-1)),
		)
		for _, sUUID := range filter.SetenceUUID {
			values = append(values, sUUID)
		}
	}

	queryStr = fmt.Sprintf(
		`SELECT sg.%s, sg.%s, sg.%s, sg.%s, sg.%s, sg.%s, sg.%s, sg.%s, sg.%s, sg.%s, s.%s as sID, s.%s as sUUID, s.%s as sName, s.%s FROM (SELECT * FROM %s %s) as sg
		LEFT JOIN %s as rsgs ON sg.%s = rsgs.%s
		%s as s ON rsgs.%s = s.%s`,
		fldID,
		fldUUID,
		fldName,
		SGRole,
		SGPoistion,
		SGRange,
		fldCreateTime,
		fldUpdateTime,
		fldEnterprise,
		fldIsDelete,
		fldID,
		fldUUID,
		fldName,
		fldCategoryID,
		tblSetnenceGroup,
		conditionStr,
		tblRelSGS,
		fldID,
		RSGSSGID,
		sentenceCondition,
		RSGSSID,
		fldID,
	)
	return
}

func (dao *SentenceGroupsSqlDaoImpl) CountBy(filter *SentenceGroupFilter, sqlLike SqlLike) (total int64, err error) {
	if sqlLike == nil {
		err = ErrNilSqlLike
		return
	}

	queryStr, values := querySentenceGroupsSQLBy(filter)
	queryStr = fmt.Sprintf("SELECT COUNT(sg.%s) FROM (%s) as sg", fldUUID, queryStr)

	rows, err := sqlLike.Query(queryStr, values...)
	if err != nil {
		err = fmt.Errorf("error while count sentence groups in dao.CountBy, err: %s", err.Error())
		return
	}
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&total)
	}
	return
}

func (dao *SentenceGroupsSqlDaoImpl) GetBy(filter *SentenceGroupFilter, sqlLike SqlLike) (groups []SentenceGroup, err error) {
	if sqlLike == nil {
		err = ErrNilSqlLike
		return
	}

	queryStr, values := querySentenceGroupsSQLBy(filter)
	if filter.Limit != 0 {
		start := filter.Page * filter.Limit
		queryStr = fmt.Sprintf("%s LIMIT %d, %d", queryStr, start, filter.Limit)
	}

	rows, err := sqlLike.Query(queryStr, values...)
	if err != nil {
		err = fmt.Errorf("error while get sentence groups in dao.GetBy, err: %s", err.Error())
		return
	}

	groups = []SentenceGroup{}
	var cGroup *SentenceGroup // current group pointer
	for rows.Next() {
		group := SentenceGroup{}
		var sentenceUUID *string
		var sentenceName *string
		var sentenceCID *uint64
		var sentenceID *uint64

		err = rows.Scan(
			&group.ID,
			&group.UUID,
			&group.Name,
			&group.Role,
			&group.Position,
			&group.Distance,
			&group.CreateTime,
			&group.UpdateTime,
			&group.Enterprise,
			&group.Deleted,
			&sentenceID,
			&sentenceUUID,
			&sentenceName,
			&sentenceCID,
		)

		if err != nil {
			err = fmt.Errorf("error while scan setence groups in dao.GetBy, err: %s", err.Error())
			return
		}

		if cGroup == nil || group.UUID != cGroup.UUID {
			if cGroup != nil {
				groups = append(groups, *cGroup)
			}
			cGroup = &group
			cGroup.Sentences = []SimpleSentence{}
		}

		if sentenceUUID != nil && sentenceName != nil && sentenceCID != nil && sentenceID != nil {
			simpleSentence := SimpleSentence{
				ID:         *sentenceID,
				UUID:       *sentenceUUID,
				Name:       *sentenceName,
				CategoryID: *sentenceCID,
			}
			cGroup.Sentences = append(cGroup.Sentences, simpleSentence)
		}
	}

	if cGroup != nil {
		groups = append(groups, *cGroup)
	}
	return
}

func (dao *SentenceGroupsSqlDaoImpl) GetBySentenceID(id []int64, sqlLike SqlLike) (groups []SentenceGroup, err error) {
	groups = []SentenceGroup{}
	if len(id) == 0 {
		return
	}

	builder := NewWhereBuilder(andLogic, "")

	builder.In(RSGSSID, int64ToWildCard(id...))
	conditionStr, values := builder.Parse()

	queryStr := fmt.Sprintf(
		`SELECT %s FROM %s WHERE %s`,
		RSGSSGID,
		tblRelSGS,
		conditionStr,
	)

	rows, err := sqlLike.Query(queryStr, values...)
	if err != nil {
		err = fmt.Errorf("error while query relation in dao.GetBySentenceID, err: %s\n", err.Error())
		return
	}
	defer rows.Close()

	sgID := []uint64{}
	for rows.Next() {
		var sgid uint64
		err = rows.Scan(&sgid)
		if err != nil {
			err = fmt.Errorf("error while scan sentence group id in dao.GetBySentenceID, err: %s\n", err.Error())
			return
		}
		sgID = append(sgID, sgid)
	}

	if len(sgID) > 0 {
		filter := &SentenceGroupFilter{
			ID: sgID,
		}
		groups, err = dao.GetBy(filter, sqlLike)
		if err != nil {
			err = fmt.Errorf("error while get setence group in dao.GetBySentenceID, err: %s\n", err.Error())
			return
		}
	}
	return
}

func (dao *SentenceGroupsSqlDaoImpl) Update(id string, group *SentenceGroup, sqlLike SqlLike) (updatedGroup *SentenceGroup, err error) {
	return
}

func (dao *SentenceGroupsSqlDaoImpl) Delete(id string, sqlLike SqlLike) (err error) {
	deleteStr := fmt.Sprintf("UPDATE %s SET %s=1 WHERE %s=?", tblSetnenceGroup, fldIsDelete, fldUUID)

	_, err = sqlLike.Exec(deleteStr, id)
	if err != nil {
		err = fmt.Errorf("error while delete sentenct group in dao.Delete, err: %s", err.Error())
		return
	}
	return
}

func (dao *SentenceGroupsSqlDaoImpl) CreateMany(groups []SentenceGroup, sqlLike SqlLike) (err error) {
	if len(groups) == 0 {
		return
	}

	insertStr, values := getSentenceGroupInsertSQL(groups)

	_, err = sqlLike.Exec(insertStr, values...)
	if err != nil {
		logger.Error.Printf("error while insert sentence groups in dao.CreateMany, sql: %s", insertStr)
		logger.Error.Printf("error while insert sentence groups in dao.CreateMany, values: %s", values)
		err = fmt.Errorf("error while insert sentence groups in dao.CreateMany, error: %s", err.Error())
		return
	}

	groupUUID := make([]string, len(groups))
	for idx, group := range groups {
		groupUUID[idx] = group.UUID
	}

	deleted := int8(0)
	filter := &SentenceGroupFilter{
		UUID:     groupUUID,
		IsDelete: &deleted,
	}

	newGroups, err := dao.GetBy(filter, sqlLike)
	if err != nil {
		return
	}

	// update groups id
	groupMap := map[string]int64{}
	for _, group := range newGroups {
		groupMap[group.UUID] = group.ID
	}

	for i := range groups {
		group := &groups[i]
		group.ID = groupMap[group.UUID]
	}

	insertStr, values = getSentenceGroupRelationInsertSQL(groups)
	if len(values) == 0 {
		return
	}

	_, err = sqlLike.Exec(insertStr, values...)
	if err != nil {
		logger.Error.Printf("error while insert sentence group sentence relation in dao.CreateMany, sql: %s", insertStr)
		logger.Error.Printf("error while insert sentence group sentence relation in dao.CreateMany, values: %+v", values)
		err = fmt.Errorf("error while insert sentence group sentence relation in dao.CreateMany, err: %s", err.Error())
	}
	return
}

func (dao *SentenceGroupsSqlDaoImpl) DeleteMany(groupUUID []string, sqlLike SqlLike) (err error) {
	if len(groupUUID) == 0 {
		return
	}

	builder := NewWhereBuilder(andLogic, "")
	builder.In(fldUUID, stringToWildCard(groupUUID...))

	conditionStr, values := builder.Parse()

	deleteStr := fmt.Sprintf(
		"UPDATE %s SET %s = 1 WHERE %s",
		tblSetnenceGroup,
		fldIsDelete,
		conditionStr,
	)

	_, err = sqlLike.Exec(deleteStr, values...)
	if err != nil {
		logger.Error.Printf("error while delete sentence groups in dao.DeleteMany, sql: %s", deleteStr)
		logger.Error.Printf("error while delete sentence groups in dao.DeleteMany, values: %+v", values)
		err = fmt.Errorf("error while delete sentence groups in dao.DeleteMany, err: %s", err.Error())
	}
	return
}
