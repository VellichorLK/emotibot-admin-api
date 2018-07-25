package util

import (
	"bytes"
	"fmt"
	"math/rand"
	"net/http"
	"regexp"
	"runtime"
	"strings"
	"time"

	"emotibot.com/emotigo/module/admin-api/ApiError"
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

func GenRandomUUIDSameAsOpenAPI() string {
	now := time.Now()
	randomNum := rand.Intn(900) + 100
	ret := fmt.Sprintf("%d%02d%02d%02d%02d%02d%06d%03d",
		now.Year(), now.Month(), now.Day(),
		now.Hour(), now.Minute(), now.Second(), now.Nanosecond()/1000,
		randomNum)
	return ret
}

// PrintRuntimeStack will print run stack with most layer of maxStack
func PrintRuntimeStack(maxStack int) {
	pc := make([]uintptr, maxStack)
	n := runtime.Callers(0, pc)
	if n == 0 {
		return
	}
	pc = pc[:n]
	frames := runtime.CallersFrames(pc)
	var buf bytes.Buffer

	buf.WriteString("Stack: \n")
	for {
		frame, more := frames.Next()
		buf.WriteString(fmt.Sprintf("\t[%s:%d]\n", frame.File, frame.Line))
		if !more {
			break
		}
	}
	LogTrace.Printf(buf.String())
}

type RetObj struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Result  interface{} `json:"result"`
}

func GenRetObj(status int, result interface{}) RetObj {
	return RetObj{
		Status:  status,
		Message: ApiError.GetErrorMsg(status),
		Result:  result,
	}
}

func GenRetObjWithCustomMsg(status int, message string, result interface{}) RetObj {
	return RetObj{
		Status:  status,
		Message: message,
		Result:  result,
	}
}

func GenSimpleRetObj(status int) RetObj {
	return RetObj{
		Status:  status,
		Message: ApiError.GetErrorMsg(status),
	}
}
