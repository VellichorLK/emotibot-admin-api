package requestheader

import (
	"net/http"
	"regexp"
	"strings"

	"emotibot.com/emotigo/module/admin-api/util/localemsg"
)

const (
	// ConstAuthorizationHeaderKey is header used for auth, content will be appid only
	ConstAuthorizationHeaderKey = "Authorization"

	// ConstUserIDHeaderKey is header record the userid
	ConstUserIDHeaderKey = "X-UserID"

	// ConstUserIPHeaderKey is header record the userip
	ConstUserIPHeaderKey = "X-Real-IP"

	// ConstLocaleHeaderKey is header record the request locale, which may be zh-cn or zh-tw
	ConstLocaleHeaderKey = "X-Locale"
	defaultLocale        = localemsg.ZhCn

	ConstEnterpriseIDHeaderKey = "X-EnterpriseID"
	ConstAppIDHeaderKey        = "X-AppID"
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
	return header
}

// GetEnterpriseID will get enterprise ID from http header
func GetEnterpriseID(r *http.Request) string {
	enterprise := r.Header.Get(ConstEnterpriseIDHeaderKey)
	match, _ := regexp.MatchString("[a-zA-Z0-9]+", enterprise)
	if match {
		return enterprise
	}
	return ""
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

func GetLocale(r *http.Request) string {
	locale := r.Header.Get(ConstLocaleHeaderKey)
	if locale != localemsg.ZhCn && locale != localemsg.ZhTw {
		locale = ""
	}
	if locale != "" {
		return locale
	}

	locale = r.URL.Query().Get("locale")
	if locale == "" {
		locale = defaultLocale
	}
	return locale
}

// SetEnterprise will set EnterpriseID to http header
func SetEnterprise(r *http.Request, enterprise string) {
	r.Header.Set(ConstEnterpriseIDHeaderKey, enterprise)
}

// SetAppID will set AppID to http header
func SetAppID(r *http.Request, appid string) {
	r.Header.Set(ConstAppIDHeaderKey, appid)
}

// SetUserID will get UserID from http header
func SetUserID(r *http.Request, user string) {
	r.Header.Set(ConstUserIDHeaderKey, user)
}
