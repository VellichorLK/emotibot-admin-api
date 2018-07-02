package ApiError

import (
	"net/http"
)

var (
	errorMsgMap = map[int]string{
		0:   "success",
		-1:  "db error",
		-2:  "io error",
		-4:  "consul server unavailable",
		-7:  "Error when send request to other API server",
		-8:  "Error when resource not found",
		101: "Uploaded file still processing",
		102: "File extension should be xlsx",
		103: "File size should smaller than 2MB",
		104: "Multicustomer is not available",
		201: "Get no info of given id",
		301: "Return from openapi has error",
		401: "Intent file format error",
	}

	// SUCCESS
	SUCCESS = 0

	// DB error, like db disconnect, schema error, etc
	DB_ERROR = -1

	// IO error, like create dir fail, create file err, etc
	IO_ERROR = -2

	// Request error, like error input
	REQUEST_ERROR = -3

	// Consul service is not available
	CONSUL_SERVICE_ERROR = -4

	// Openapi service is not available
	OPENAPI_URL_ERROR = -5

	// Parse from json error
	JSON_PARSE_ERROR = -6

	// API Request Error
	WEB_REQUEST_ERROR = -7

	// Resource not found
	NOT_FOUND_ERROR = -8

	// Base64 decode error
	BASE64_PARSE_ERROR = -9

	// Dictionary error: last file still running
	DICT_STILL_RUNNING = 101
	// Dictionary error: extension of uploaded file error
	DICT_FORMAT_ERROR = 102
	// Dictionary error: uploaded file too large
	DICT_SIZE_ERROR = 103
	// Dictionary error: multicustomer service is not available
	DICT_SERVICE_ERROR = 104

	// Switch-manage error: return from database is empty
	SWITCH_NO_ROWS = 201

	// QA test error: return format from openapi has error
	QA_TEST_FORMAT_ERROR = 301

	// Intent file format error
	INTENT_FORMAT_ERROR = 401
)

func GetErrorMsg(errno int) string {
	if errMsg, ok := errorMsgMap[errno]; ok {
		return errMsg
	}
	return ""
}

func GetHttpStatus(errno int) int {
	switch errno {
	case SUCCESS:
		return http.StatusOK
	case NOT_FOUND_ERROR:
		return http.StatusNotFound
	case REQUEST_ERROR:
		return http.StatusBadRequest
	case DB_ERROR:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
