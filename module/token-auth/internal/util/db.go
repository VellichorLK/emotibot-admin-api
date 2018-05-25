package util

import (
	"database/sql"
	"runtime"
)

func LogDBError(err error) {
	if err != nil {
		_, file, line2, _ := runtime.Caller(2)
		_, _, line1, _ := runtime.Caller(1)
		LogError.Printf(`DB Error:
			[%s:%d]
			[%s:%d]
			Error: %s`, file, line1, file, line2, err.Error())
	}
}

func ClearTransition(tx *sql.Tx) {
	rollbackRet := tx.Rollback()
	if rollbackRet != sql.ErrTxDone && rollbackRet != nil {
		LogError.Printf("Critical db error in rollback: %s\n", rollbackRet.Error())
	}
}
