package model

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrNilSqlLike = errors.New("SqlLike can not be nil")
)

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
	UUID       string
	Name       string
	Role       int
	Position   int
	Enterprise string
	CreateTime *time.Time
	UpdateTime *time.Time
	Page       int
	Limit      int
}

type SentenceGroupsSqlDao interface {
	Create(group *SentenceGroup, sql SqlLike) (*SentenceGroup, error)
	CountBy(filter *SentenceGroupFilter, sql SqlLike) (int64, error)
	GetBy(filter *SentenceGroupFilter, sql SqlLike) ([]SentenceGroup, error)
	Update(id string, group *SentenceGroup, sql SqlLike) (*SentenceGroup, error)
	Delete(id string) error
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

	fmt.Printf("insertStr: %s\n", insertStr)

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

func (dao *SentenceGroupsSqlDaoImpl) CountBy(filter *SentenceGroupFilter, sql SqlLike) (total int64, err error) {
	return
}

func (dao *SentenceGroupsSqlDaoImpl) GetBy(filter *SentenceGroupFilter, sql SqlLike) (groups []SentenceGroup, err error) {
	return
}

func (dao *SentenceGroupsSqlDaoImpl) Update(id string, group *SentenceGroup, sql SqlLike) (updatedGroup *SentenceGroup, err error) {
	return
}

func (dao *SentenceGroupsSqlDaoImpl) Delete(id string) (err error) {
	return
}
