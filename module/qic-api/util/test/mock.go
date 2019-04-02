package test

import (
	"fmt"
	"os"

	"emotibot.com/emotigo/module/qic-api/model/v1"
	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
)

type MockDBLike struct {
	model.DefaultDBLike
}

func (m *MockDBLike) Begin() (model.SQLTx, error) {
	var mock sqlmock.Sqlmock
	var err error
	m.DB, mock, err = sqlmock.New()
	if err != nil {
		fmt.Printf("sqlmock new failed. %s\n", err)
		os.Exit(-1)
	}
	mock.ExpectBegin()
	mock.ExpectCommit()

	return m.DB.Begin()
}

func (m *MockDBLike) ClearTransition(tx model.SQLTx) {
	return
}

func (m *MockDBLike) Commit(tx model.SQLTx) (err error) {
	return tx.Commit()
}

func (m *MockDBLike) Conn() model.SqlLike {
	return m.DB
}
