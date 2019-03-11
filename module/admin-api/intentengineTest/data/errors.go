package data

import (
	"errors"
)

var (
	ErrTestNotFound                 = errors.New("Test not found")
	ErrTestIntentNotFound           = errors.New("Test Intent not found")
	ErrRestoredIntentNotFound       = errors.New("Restored intent not found")
	ErrTestTaskExpired              = errors.New("Intents test task has expired")
	ErrTestTaskInProcess            = errors.New("Intents test task is in process")
	ErrTestImportSheetFormat        = errors.New("Intent test import file format error")
	ErrTestImportSheetNoHeader      = errors.New("Intent test import file sheet header not found")
	ErrTestImportSheetEmptySentence = errors.New("Intent test import file with empty sentence")
)

type ErrorResponse struct {
	Message string `json:"message"`
}

func NewErrorResponse(message string) ErrorResponse {
	return ErrorResponse{
		Message: message,
	}
}
