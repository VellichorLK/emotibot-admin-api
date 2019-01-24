package util

import (
	"bytes"
	"fmt"
	"math/rand"
	"net/http"
	"regexp"
	"runtime"
	"time"

	"emotibot.com/emotigo/pkg/logger"
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
	logger.Trace.Printf(buf.String())
}

// GenRandomString will generate a string with length, which will only contains 0-9, a-z, A-Z
func GenRandomString(length int) string {
	rand.Seed(time.Now().Unix())
	buf := bytes.Buffer{}
	for buf.Len() < length {
		var c uint8
		for c < uint8('0') ||
			(c > uint8('9') && c < uint8('A')) ||
			(c > uint8('Z') && c < uint8('a')) ||
			(c > uint8('z')) {
			c = uint8(rand.Uint32() % 128)
		}
		buf.WriteRune(rune(c))
	}
	return buf.String()
}
