package model

import (
	"fmt"

	"emotibot.com/emotigo/pkg/logger"
)

//NavOnTheFlyDao is the interface of navigate on the fly
type NavOnTheFlyDao interface {
	InitConerationResult(conn SqlLike, callID int64, modelID int64, val string) (int64, error)
	//	UpdateFlowResult(conn SqlLike, callID int64, val string) (int64, error)
	//	GetFlowResultFrom(conn SqlLike, callID int64) (*QIFlowResult, error)
}

//NavOnTheFlySQLDao implements the function to access the db to finish the navigation
type NavOnTheFlySQLDao struct {
}

//InitConerationResult inserts the initial conversation result, simply the name
func (n *NavOnTheFlySQLDao) InitConerationResult(conn SqlLike, callID int64, modelID int64, val string) (int64, error) {
	if conn == nil {
		return 0, ErrNilSqlLike
	}
	insertSQL := fmt.Sprintf("INSERT INTO %s (%s,%s,%s) VALUES (?,?,?)", tblCUPredict, fldCallID, fldAppID, fldCUPredict)
	result, err := conn.Exec(insertSQL, callID, modelID, val)
	if err != nil {
		logger.Error.Println("raw sql: ", insertSQL)
		logger.Error.Printf("raw bind-data: [%v,%v]\n", callID, val)
		return 0, fmt.Errorf("sql executed failed, %v", err)
	}
	return result.LastInsertId()
}
