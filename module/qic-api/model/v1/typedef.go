package model

import (
	"database/sql"

	"emotibot.com/emotigo/pkg/logger"
)

type DBLike interface {
	Begin() (*sql.Tx, error)
	ClearTransition(tx *sql.Tx)
	Commit(tx *sql.Tx) error
	Conn() *sql.DB //return the conn
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

func (dl *DefaultDBLike) Commit(tx *sql.Tx) (err error) {
	if tx != nil {
		err = tx.Commit()
		return
	}
	return
}

func (dl *DefaultDBLike) Conn() *sql.DB {
	return dl.DB
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
	EnterpriseID  string
	Page          int
	Limit         int
	UUID          []string
}
