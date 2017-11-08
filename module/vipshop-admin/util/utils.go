package util

import (
	"github.com/kataras/iris/context"
)

const (
	// ConstAuthorizationHeaderKey is header used for auth, content will be appid only
	ConstAuthorizationHeaderKey = "Authorization"

	// ConstUserIDHeaderKey is header record the userid
	ConstUserIDHeaderKey = "X-UserID"

	// ConstUserIPHeaderKey is header record the userip
	ConstUserIPHeaderKey = "X-Real-IP"
)

// GetAppID will get AppID from http header
func GetAppID(ctx context.Context) string {
	return ctx.GetHeader(ConstAuthorizationHeaderKey)
}

// GetUserID will get UserID from http header
func GetUserID(ctx context.Context) string {
	return ctx.GetHeader(ConstUserIDHeaderKey)
}

// GetUserIP will get User addr from http header
func GetUserIP(ctx context.Context) string {
	return ctx.GetHeader(ConstUserIPHeaderKey)
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
