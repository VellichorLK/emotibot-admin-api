package util

import (
	"bytes"
	"fmt"
	"math"
	"math/rand"
	"runtime"
	"time"

	"emotibot.com/emotigo/module/admin-api/ApiError"
	"emotibot.com/emotigo/pkg/logger"
)

func init() {
	rand.Seed(time.Now().UnixNano())
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
	logger.Trace.Printf(buf.String())
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

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func GenRandomString(length int) string {
	b := make([]rune, length)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func GenRandomBytes(length int) []byte {
	b := make([]byte, length)
	for i := range b {
		b[i] = byte(rand.Intn(math.MaxInt8))
	}
	return b
}
