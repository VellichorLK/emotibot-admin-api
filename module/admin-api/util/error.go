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
