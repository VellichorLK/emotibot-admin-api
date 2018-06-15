package util

import (
	"errors"
)

var (
	ErrUserNameExists       = errors.New("Conflict user name")
	ErrUserEmailExists      = errors.New("Conflict user email")
	ErrEnterpriseInfoExists = errors.New("Conflict enterprise info")
	ErrAppInfoExists        = errors.New("Conflict app info")
	ErrGroupInfoExists      = errors.New("Conflict group info")
	ErrRoleInfoExists       = errors.New("Conflict role info")
	ErrRobotGroupNotExist   = errors.New("Robot group does not exist")
	ErrRobotNotExist        = errors.New("Robot does not exist")
	ErrRoleNotExist         = errors.New("Role does not exist")
)
