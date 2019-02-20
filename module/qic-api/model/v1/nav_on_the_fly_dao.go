package model

import (
	"database/sql"
	"fmt"
	"strings"

	"emotibot.com/emotigo/pkg/logger"
)

type StreamingPredict struct {
	ID      int64
	CallID  int64
	AppID   int64
	Predict string
}

//NavOnTheFlyDao is the interface of navigate on the fly
type NavOnTheFlyDao interface {
	InitConversationResult(conn SqlLike, callID int64, modelID int64, val string) (int64, error)
	UpdateFlowResult(conn SqlLike, callID int64, val string) (int64, error)
	GetStreamingPredictResult(conn SqlLike, callID int64) ([]*StreamingPredict, error)
	//	GetFlowResultFrom(conn SqlLike, callID int64) (*QIFlowResult, error)
}

//NavOnTheFlySQLDao implements the function to access the db to finish the navigation
type NavOnTheFlySQLDao struct {
}

//InitConversationResult inserts the initial conversation result, simply the name
func (n *NavOnTheFlySQLDao) InitConversationResult(conn SqlLike, callID int64, modelID int64, val string) (int64, error) {
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

//UpdateFlowResult updates the record the CUPredict for now
func (n *NavOnTheFlySQLDao) UpdateFlowResult(conn SqlLike, callID int64, val string) (int64, error) {
	if conn == nil {
		return 0, ErrNilSqlLike
	}
	updateSQL := fmt.Sprintf("UPDATE %s SET %s=? WHERE %s=?", tblCUPredict, fldCUPredict, fldCallID)
	result, err := conn.Exec(updateSQL, val, callID)
	if err != nil {
		logger.Error.Println("raw sql: ", updateSQL)
		logger.Error.Printf("raw bind-data: [%v,%v]\n", callID, val)
		return 0, fmt.Errorf("sql executed failed, %v", err)
	}
	return result.RowsAffected()
}

//GetStreamingPredictResult gets the streaming result from CUPredict table
func (n *NavOnTheFlySQLDao) GetStreamingPredictResult(conn SqlLike, callID int64) ([]*StreamingPredict, error) {
	if conn == nil {
		return nil, ErrNilSqlLike
	}
	flds := []string{
		fldID,
		fldCallID,
		fldAppID,
		fldCUPredict,
	}
	selectedStr := strings.Join(flds, ",")

	querySQL := fmt.Sprintf("SELECT %s FROM %s WHERE %s=?", selectedStr, tblCUPredict, fldCallID)
	rows, err := conn.Query(querySQL, callID)
	if err != nil {
		logger.Error.Printf("query failed. raw sql:%s. %s\n ", querySQL, err)
		return nil, err
	}
	defer rows.Close()

	predicts := make([]*StreamingPredict, 0)

	var id int64
	var call, appID sql.NullInt64
	var predict sql.NullString
	for rows.Next() {

		err = rows.Scan(&id, &call, &appID, &predict)
		if err != nil {
			logger.Error.Printf("scan failed. %s\n", err)
			return nil, err
		}

		record := &StreamingPredict{ID: id, CallID: call.Int64, AppID: appID.Int64, Predict: predict.String}
		predicts = append(predicts, record)
	}

	return predicts, nil

}
