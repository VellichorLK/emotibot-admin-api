package util

import (
	"net/http"
	"regexp"
)

const (
	// ConstUserIDHeaderKey is header record the userid
	ConstUserIDHeaderKey = "X-UserID"

	// ConstUserIPHeaderKey is header record the userip
	ConstUserIPHeaderKey = "X-Real-IP"

	// ConstAppIDHeaderKey is header record the appid
	ConstAppIDHeaderKey = "X-AppID"

	// ConstEnterpriseIDHeaderKey is header record the appid
	ConstEnterpriseIDHeaderKey = "X-EnterpriseID"
)

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

// GetEnterpriseID will get User addr from http header
func GetEnterpriseID(r *http.Request) string {
	return r.Header.Get(ConstEnterpriseIDHeaderKey)
}
