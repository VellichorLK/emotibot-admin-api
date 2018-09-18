package data

import (
	"errors"
)

var (
	// ErrUnsupportedMethod is the error of unsupported HTTP method
	ErrUnsupportedMethod = errors.New("Unsupported method")
)
