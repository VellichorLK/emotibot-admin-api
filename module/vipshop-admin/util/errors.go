package util

import "errors"

// ErrorMCLock represent multicustomer return locking status which can not run the command, should try later
var ErrorMCLock = errors.New("Someone else is already locking qa module, please try again later")
