package model

import (
	"errors"
	"fmt"
	"strings"
	"time"

	_ "emotibot.com/emotigo/pkg/logger"
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
}

type SentenceGroupFilter struct {
	UUID       []string
	ID         []uint64
	Name       string
	Role       int
	Position   int
	Distance   int
	Enterprise string
	CreateTime *time.Time
	UpdateTime *time.Time
	Page       int
	Limit      int
	IsDelete   int8
}

type SentenceGroupsSqlDao interface {
	Create(group *SentenceGroup, sql SqlLike) (*SentenceGroup, error)
	CountBy(filter *SentenceGroupFilter, sql SqlLike) (int64, error)
	GetBy(filter *SentenceGroupFilter, sql SqlLike) ([]SentenceGroup, error)
	Update(id string, group *SentenceGroup, sql SqlLike) (*SentenceGroup, error)
	Delete(id string, sqllike SqlLike) error
}

type SentenceGroupsSqlDaoImpl struct{}

func (dao *SentenceGroupsSqlDaoImpl) Create(group *SentenceGroup, sql SqlLike) (createdGroup *SentenceGroup, err error) {
	if sql == nil {
		err = ErrNilSqlLike
		return
	}

	// insert sentence group
	fileds := []string{
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

	values := []interface{}{
		0,
		group.Name,
		group.Enterprise,
		group.Role,
		group.Position,
		group.Distance,
		group.UUID,
		group.CreateTime,
		group.UpdateTime,
	}

	fieldStr := strings.Join(fileds, "`, `")
	fieldStr = fmt.Sprintf("`%s`", fieldStr)
	valueStr := fmt.Sprintf("?%s", strings.Repeat(", ?", len(fileds)-1))
	insertStr := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", tblSetnenceGroup, fieldStr, valueStr)

	result, err := sql.Exec(insertStr, values...)
	if err != nil {
		err = fmt.Errorf("error while insert sentence group in dao.Create, err: %s", err.Error())
		return
	}

	groupID, err := result.LastInsertId()
	if err != nil {
		err = fmt.Errorf("error while get group id in dao.Create, err: %s", err.Error())
		return

	}

	if len(group.Sentences) > 0 {
		// create sentence group to sentence relation
		values = []interface{}{}
		for _, ss := range group.Sentences {
			values = append(values, groupID, ss.ID)
		}

		valueStr = fmt.Sprintf("(?, ?) %s", strings.Repeat(", (?, ?)", len(group.Sentences)-1))
		insertStr = fmt.Sprintf("INSERT INTO %s (%s, %s) VALUES %s", tblRelSGS, RSGSSGID, RSGSSID, valueStr)

		_, err = sql.Exec(insertStr, values...)
		if err != nil {
			err = fmt.Errorf("error while insert sentence group to sentence relationin dao.Create, err: %s", err.Error())
			return
		}
	}
	group.ID = groupID
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

	if filter.Role != -1 {
		conditions = append(conditions, SGRole)
		values = append(values, filter.Role)
	}

	if filter.Position != -1 {
		conditions = append(conditions, SGPoistion)
		values = append(values, filter.Position)
	}

	if filter.Distance != 0 {
		conditions = append(conditions, SGRange)
		values = append(values, filter.Distance)
	}

	if filter.IsDelete != -1 {
		conditions = append(conditions, fldIsDelete)
		values = append(values, filter.IsDelete)
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

	queryStr = fmt.Sprintf(
		`SELECT sg.%s, sg.%s, sg.%s, sg.%s, sg.%s, sg.%s, sg.%s, sg.%s, s.%s as sUUID, s.%s as sName, s.%s FROM (SELECT * FROM %s %s) as sg
		LEFT JOIN %s as rsgs ON sg.%s = rsgs.%s
		LEFT JOIN %s as s ON rsgs.%s = s.%s`,
		fldID,
		fldUUID,
		fldName,
		SGRole,
		SGPoistion,
		SGRange,
		fldCreateTime,
		fldUpdateTime,
		fldUUID,
		fldName,
		fldCategoryID,
		tblSetnenceGroup,
		conditionStr,
		tblRelSGS,
		fldID,
		RSGSSGID,
		tblSentence,
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

func (dao *SentenceGroupsSqlDaoImpl) GetBy(filter *SentenceGroupFilter, sql SqlLike) (groups []SentenceGroup, err error) {
	if sql == nil {
		err = ErrNilSqlLike
		return
	}

	queryStr, values := querySentenceGroupsSQLBy(filter)
	if filter.Limit != 0 {
		start := filter.Page * filter.Limit
		queryStr = fmt.Sprintf("%s LIMIT %d, %d", queryStr, start, filter.Limit)
	}

	rows, err := sql.Query(queryStr, values...)
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

		err = rows.Scan(
			&group.ID,
			&group.UUID,
			&group.Name,
			&group.Role,
			&group.Position,
			&group.Distance,
			&group.CreateTime,
			&group.UpdateTime,
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

		if sentenceUUID != nil && sentenceName != nil && sentenceCID != nil {
			simpleSentence := SimpleSentence{
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

func (dao *SentenceGroupsSqlDaoImpl) Update(id string, group *SentenceGroup, sql SqlLike) (updatedGroup *SentenceGroup, err error) {
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
