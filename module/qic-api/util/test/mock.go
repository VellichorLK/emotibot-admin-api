package test

import (
	"database/sql"
)

type MockDBLike struct{}

func (m *MockDBLike) Begin() (*sql.Tx, error) {
	return nil, nil
}

func (m *MockDBLike) ClearTransition(tx *sql.Tx) {
	return
}

func (m *MockDBLike) Commit(tx *sql.Tx) (err error) {
	return
}

func (m *MockDBLike) Conn() *sql.DB {
	return nil
}
