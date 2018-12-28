package model

import (
	"time"
)

type SentenceGroup struct {
	ID         int64
	UUID       string
	Name       string
	Role       int
	Position   int
	Sentences  []string
	Enterprise string
	CreateTime *time.Time
	UpdateTime *time.Time
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
