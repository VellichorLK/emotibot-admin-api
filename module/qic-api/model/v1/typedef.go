package model

import (
	"database/sql"

	"emotibot.com/emotigo/pkg/logger"
)

// DBLike is the abstract form of sql.DB
// Which should be able to get a connection and get transaction
type DBLike interface {
	Begin() (SQLTx, error)
	ClearTransition(SQLTx) //a way to compatible with old code
	Commit(SQLTx) error    //a way to compatible with old code
	Conn() SqlLike
}

// SQLTx is a Transaction form of SqlLike
type SQLTx interface {
	SqlLike
	Commit() error
	Rollback() error
}

type DefaultDBLike struct {
	DB *sql.DB
}

func (dl *DefaultDBLike) Begin() (SQLTx, error) {
	return dl.DB.Begin()
}

func (dl *DefaultDBLike) ClearTransition(tx SQLTx) {
	rollbackRet := tx.Rollback()
	if rollbackRet != sql.ErrTxDone && rollbackRet != nil {
		logger.Error.Printf("Critical db error in rollback: %s", rollbackRet.Error())
	}
}

func (dl *DefaultDBLike) Commit(tx SQLTx) (err error) {
	if tx != nil {
		err = tx.Commit()
		return
	}
	return
}

func (dl *DefaultDBLike) Conn() SqlLike {
	return dl.DB
}

type SqlLike interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Prepare(query string) (*sql.Stmt, error)
}
