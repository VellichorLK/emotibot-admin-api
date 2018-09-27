package data

import (
	"errors"
)

var (
	// ErrUnsupportedMethod is the error of unsupported HTTP method
	ErrUnsupportedMethod = errors.New("Unsupported method")
	// ErrAppIDNotSpecified is the error of app_id not specified
	ErrAppIDNotSpecified = errors.New("App ID not specified")
	// ErrUserIDNotSpecified is the error of user_id not specified
	ErrUserIDNotSpecified = errors.New("User ID not specified")
)
