package util

import (
	"net/http"
	"strings"
)

const (
	// ConstAuthorizationHeaderKey is header used for auth, content will be appid only
	ConstAuthorizationHeaderKey = "Authorization"

	// ConstUserIDHeaderKey is header record the userid
	ConstUserIDHeaderKey = "X-UserID"

	// ConstUserIPHeaderKey is header record the userip
	ConstUserIPHeaderKey = "X-Real-IP"

	ConstAppIDHeaderKey = "X-AppID"
)

func GetAuthToken(r *http.Request) string {
	header := r.Header.Get(ConstAuthorizationHeaderKey)
	params := strings.Split(header, " ")
	if len(params) < 2 {
		return ""
	}
	return params[1]
}

// GetAppID will get AppID from http header
func GetAppID(r *http.Request) string {
	return r.Header.Get(ConstAppIDHeaderKey)
}

// GetUserID will get UserID from http header
func GetUserID(r *http.Request) string {
	return r.Header.Get(ConstUserIDHeaderKey)
}

// GetUserIP will get User addr from http header
func GetUserIP(r *http.Request) string {
	return r.Header.Get(ConstUserIPHeaderKey)
}

// Contains will check if str is in arr or not
func Contains(arr []string, str string) bool {
	for _, s := range arr {
		if s == str {
			return true
		}
	}
	return false
}

func IsValidAppID(id string) bool {
	return len(id) > 0 && HasOnlyNumEng(id)
}

func HasOnlyNumEng(input string) bool {
	for _, c := range input {
		if (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') && (c < '0' || c > '9') {
			return false
		}
	}
	return true
}
