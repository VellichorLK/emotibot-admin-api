package util

import (
	"errors"
	"fmt"
)

var (
	// ErrDBNotInit used when cannot get default db
	ErrDBNotInit  = errors.New("DB not init")
	ErrDuplicated = errors.New("資源已存在")
	ErrNotFound   = errors.New("資源不存在")

	// ErrorMCLock represent multicustomer return locking status which can not run the command, should try later
	ErrorMCLock = errors.New("Someone else is already locking qa module, please try again later")

	// ErrSQLRowNotFound represent SQL query not found error
	ErrSQLRowNotFound = errors.New("Not Found")

	// ErrSQLAlreadyOccupied represent rows already have value, should not updated it.
	ErrSQLAlreadyOccupied = errors.New("db row already updated")

	ErrParameter = errors.New("Invalid parameter")
)

func GenNotFoundError(name string) error {
	return fmt.Errorf(Msg["NotExistTemplate"], name)
}

func GenDuplicatedError(column string, resource string) error {
	return fmt.Errorf(Msg["DuplicateTemplate"], column, resource)
}

func GenBadRequestError(column string) error {
	return fmt.Errorf(Msg["BadRequestTemplate"], column)
}
