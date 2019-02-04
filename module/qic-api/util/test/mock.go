package test

import (
	"emotibot.com/emotigo/module/qic-api/model/v1"
)

type MockDBLike struct{}

func (m *MockDBLike) Begin() (model.SQLTx, error) {
	return nil, nil
}

func (m *MockDBLike) ClearTransition(tx model.SQLTx) {
	return
}

func (m *MockDBLike) Commit(tx model.SQLTx) (err error) {
	return
}

func (m *MockDBLike) Conn() model.SqlLike {
	return nil
}
