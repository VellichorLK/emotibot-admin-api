package data

import (
	"net/http"
)

type ResponseLogger struct {
	http.ResponseWriter
	StatusCode int
}

func NewResponseLogger(w http.ResponseWriter) *ResponseLogger {
	return &ResponseLogger{
		ResponseWriter: w,
		StatusCode:     http.StatusOK,
	}
}

func (l *ResponseLogger) WriteHeader(statusCode int) {
	l.StatusCode = statusCode
	l.ResponseWriter.WriteHeader(statusCode)
}
