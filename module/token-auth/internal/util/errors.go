package util

import (
	"errors"
)

var (
	ErrInvalidParameter     = errors.New("Invalid Paramter")
	ErrOperationForbidden   = errors.New("Operation forbidden")
	ErrResourceNotFound     = errors.New("Resource not found")
	ErrUserNameExists       = errors.New("Conflict user name")
	ErrUserEmailExists      = errors.New("Conflict user email")
	ErrEnterpriseInfoExists = errors.New("Conflict enterprise info")
	ErrAppInfoExists        = errors.New("Conflict app info")
	ErrGroupInfoExists      = errors.New("Conflict group info")
	ErrRoleInfoExists       = errors.New("Conflict role info")
	ErrRobotGroupNotExist   = errors.New("Robot group does not exist")
	ErrRobotNotExist        = errors.New("Robot does not exist")
	ErrRoleNotExist         = errors.New("Role does not exist")
	ErrInteralServer        = errors.New("Internal server error")
)
