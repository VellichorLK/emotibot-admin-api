package test

import (
	"database/sql"
	"fmt"
	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
	"os"
)

type MockDBLike struct{}

func (m *MockDBLike) Begin() (*sql.Tx, error) {
	db, mock, err := sqlmock.New()
	if err != nil {
		fmt.Printf("sqlmock new failed. %s\n", err)
		os.Exit(-1)
	}
	mock.ExpectBegin()
	mock.ExpectCommit()

	return db.Begin()
}

func (m *MockDBLike) ClearTransition(tx *sql.Tx) {
	return
}

func (m *MockDBLike) Commit(tx *sql.Tx) error {
	return tx.Commit()
}

func (m *MockDBLike) Conn() *sql.DB {
	return nil
}
