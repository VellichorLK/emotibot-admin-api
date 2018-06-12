package util

import (
	"net/http"
	"regexp"
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
	if r.Method == "GET" {
		token := r.URL.Query().Get("token")
		if strings.TrimSpace(token) != "" {
			return token
		}
	}

	header := r.Header.Get(ConstAuthorizationHeaderKey)
	params := strings.Split(header, " ")
	if len(params) < 2 {
		return ""
	}
	return params[1]
}

// GetAppID will get AppID from http header
func GetAppID(r *http.Request) string {
	appid := r.Header.Get(ConstAppIDHeaderKey)
	match, _ := regexp.MatchString("[a-zA-Z0-9]+", appid)
	if match {
		return appid
	}
	return ""
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
	return len(id) > 0 && HasOnlyNumEngDash(id)
}

func HasOnlyNumEngDash(input string) bool {
	for _, c := range input {
		if (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') && (c < '0' || c > '9') && c != '-' {
			return false
		}
	}
	return true
}
