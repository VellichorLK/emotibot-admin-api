package model

import (
	"database/sql"
	"fmt"
)

const (
	tblTags = "tags"
)

type tagSQLDao struct {
	Tx *sql.Tx
}

//TagDao is tag resource manipulating interface, which itself should support ACID transaction.
type TagDao interface {
	Tags(query TagQuery) ([]Tag, error)
	NewTags(tags []Tag) ([]uint, error)
	SetTags(tag []Tag) ([]uint, error)
	DeleteTags(query TagQuery) error
	Commit() error
}

func newTagDao(db *sql.DB) (TagDao, error) {
	row := db.QueryRow("SHOW TABLEs LIKE '" + tblTags + "'")

	if err := row.Scan(); err != nil {
		return nil, fmt.Errorf("expect to have table %s, but got err: %v", tblTags, err)
	}
	tx, err := db.Begin()
	if err != nil {

	}
	return &tagSQLDao{Tx: tx}, err
}

var beginTagDaoFunc = newTagDao

type Tag struct {
	ID uint64
}

type TagQuery struct {
	ID []uint64
}

func (t *tagSQLDao) Tags(query TagQuery) ([]Tag, error) {
	if t.Tx == nil {
		return nil, fmt.Errorf("dao failed, init it properly")
	}

	return nil, nil
}

func (t *tagSQLDao) NewTags(tags []Tag) ([]uint, error) {
	return nil, nil
}

func (t *tagSQLDao) SetTags(tag []Tag) ([]uint, error) {
	return nil, nil
}

func (t *tagSQLDao) DeleteTags(query TagQuery) error {
	return nil
}

func (t *tagSQLDao) Commit() error {
	return t.Tx.Commit()
}
