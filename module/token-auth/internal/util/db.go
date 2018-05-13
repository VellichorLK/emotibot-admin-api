package util

import (
	"database/sql"
	"runtime"
)

func LogDBError(err error) {
	_, file, line, _ := runtime.Caller(1)
	LogError.Printf("Error in [%s:%d] [%s]\n", file, line, err.Error())
}

func ClearTransition(tx *sql.Tx) {
	rollbackRet := tx.Rollback()
	if rollbackRet != sql.ErrTxDone && rollbackRet != nil {
		LogError.Printf("Critical db error in rollback: %s\n", rollbackRet.Error())
	}
}
