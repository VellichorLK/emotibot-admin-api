package util

import (
    "errors"
)

var (
    ErrRoleUsersNotEmpty = errors.New("Cannot remove role having users")
)