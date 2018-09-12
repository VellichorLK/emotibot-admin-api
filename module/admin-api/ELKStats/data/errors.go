package data

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidRequestBody     = errors.New("Invalid request body")
	ErrTagTypeNotFound        = errors.New("Tag type not found")
	ErrInvalidAggType         = errors.New("Invalid aggregation type")
	ErrInvalidAggTimeInterval = errors.New("Invalid aggregtaion time interval")
	ErrESNotAcknowledged      = errors.New("Elasticsearch: Not acknowledged")
	ErrESIndexNotExisted      = errors.New("Elasticsearch: Index not exists")
	ErrESTermsNotFound        = errors.New("Elasticsearch: Terms not found")
	ErrESMaxBucketNotFound    = errors.New("Elasticsearch: Max bucket not found")
	ErrESMinBucketNotFound    = errors.New("Elasticsearch: Min bucket not found")
	ErrESCardinalityNotFound  = errors.New("Elasticsearch: Cardinality not found")
	ErrESTopHitNotFound       = errors.New("Elasticsearch: TopHit not found")
	ErrESNameBucketNotFound   = errors.New("Elasticsearch: Name bucket not found")
	ErrNotInit                = errors.New("Elasticsearch client is not inited successfully")
)

const (
	ErrCodeInvalidParameterType = iota
	ErrCodeInvalidParameterT1
	ErrCodeInvalidParameterT2
	ErrCodeInvalidParameterCategory
	ErrCodeInvalidParameterFilter
	ErrCodeInvalidRequestBody
)

type ErrorResponse struct {
	Message string `json:"message"`
}

func NewErrorResponse(message string) ErrorResponse {
	return ErrorResponse{
		Message: message,
	}
}

type ErrorResponseWithCode struct {
	ErrorResponse
	Code int `json:"code"`
}

func NewErrorResponseWithCode(code int, message string) ErrorResponseWithCode {
	return ErrorResponseWithCode{
		ErrorResponse: ErrorResponse{
			Message: message,
		},
		Code: code,
	}
}

func NewBadRequestResponse(code int, column string) ErrorResponseWithCode {
	message := fmt.Sprintf("Invalid parameter: %s", column)
	return NewErrorResponseWithCode(code, message)
}
