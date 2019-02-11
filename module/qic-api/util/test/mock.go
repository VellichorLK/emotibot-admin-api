package test

import (
	"emotibot.com/emotigo/module/qic-api/model/v1"
	"fmt"
	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
	"os"
)

type MockDBLike struct{}

func (m *MockDBLike) Begin() (model.SQLTx, error) {
	db, mock, err := sqlmock.New()
	if err != nil {
		fmt.Printf("sqlmock new failed. %s\n", err)
		os.Exit(-1)
	}
	mock.ExpectBegin()
	mock.ExpectCommit()

	return db.Begin()
}

func (m *MockDBLike) ClearTransition(tx model.SQLTx) {
	return
}

func (m *MockDBLike) Commit(tx model.SQLTx) (err error) {
	return tx.Commit()
}

func (m *MockDBLike) Conn() model.SqlLike {
	return nil
}
