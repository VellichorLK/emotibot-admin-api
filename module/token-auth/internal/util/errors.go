package util

import (
	"errors"
)

var (
	ErrRoleUsersNotEmpty  = errors.New("Cannot remove role having users")
	ErrRobotGroupNotExist = errors.New("Robot group does not exist")
	ErrRobotNotExist      = errors.New("Robot does not exist")
	ErrRoleNotExist       = errors.New("Role does not exist")
)
