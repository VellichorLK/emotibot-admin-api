package util

import "errors"

// ErrorMCLock represent multicustomer return locking status which can not run the command, should try later
var ErrorMCLock = errors.New("Someone else is already locking qa module, please try again later")

// ErrSQLRowNotFound represent SQL query not found error
var ErrSQLRowNotFound = errors.New("Not Found")

// ErrSQLAlreadyOccupied represent rows already have value, should not updated it.
var ErrSQLAlreadyOccupied = errors.New("db row already updated")
