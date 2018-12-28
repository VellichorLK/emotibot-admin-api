package model

import (
	"database/sql"

	"emotibot.com/emotigo/pkg/logger"
)

type DBLike interface {
	Begin() (*sql.Tx, error)
	ClearTransition(tx *sql.Tx)
}

type DefaultDBLike struct {
	DB *sql.DB
}

func (dl *DefaultDBLike) Begin() (*sql.Tx, error) {
	return dl.DB.Begin()
}

func (dl *DefaultDBLike) ClearTransition(tx *sql.Tx) {
	rollbackRet := tx.Rollback()
	if rollbackRet != sql.ErrTxDone && rollbackRet != nil {
		logger.Error.Printf("Critical db error in rollback: %s", rollbackRet.Error())
	}
}

type SqlLike interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Prepare(query string) (*sql.Stmt, error)
}

type GroupFilter struct {
	FileName      string
	Deal          int
	Series        string
	CallStart     int64
	CallEnd       int64
	StaffID       string
	StaffName     string
	Extension     string
	Department    string
	CustomerID    string
	CustomerName  string
	CustomerPhone string
	Page          int
	Limit         int
}
