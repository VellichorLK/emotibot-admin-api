package auth

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"time"
)

type ErrStruct struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

func RespJson(w http.ResponseWriter, es interface{}) {
	//log.Println(es)
	js, err := json.Marshal(es)
	log.Println(js)
	if HandleHttpError(http.StatusInternalServerError, err, w) {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	log.Printf("js: %s", js)
	fmt.Fprintf(w, string(js))
}

func HandleHttpMethodError(request_method string, allow_method string) bool {
	if request_method != allow_method {
		return true
	}
	return false
}

func HandleError(err_code int, err error, w http.ResponseWriter) bool {
	if err == nil {
		return false
	}
	_, fn, line, _ := runtime.Caller(1)
	log.Printf("%s:%d, %s", fn, line, err.Error())
	es := ErrStruct{err_code, err.Error()}
	RespJson(w, es)
	return true
}

func HandleHttpError(err_code int, err error, w http.ResponseWriter) bool {
	//return: true if err is not nil
	//return: false if err is nil
	if err == nil {
		return false
	}
	_, fn, line, _ := runtime.Caller(1)
	log.Printf("[%s]%s: %d: %s", time.Now(), fn, line, err.Error())
	http.Error(w, err.Error(), err_code)
	return true
}

func genMD5ID(seed string) string {
	t := fmt.Sprintf("%s-%s", seed, time.Now().Format("20060102150405"))
	s := fmt.Sprintf("%x", md5.Sum([]byte(t)))
	return s
}

func GenEnterpriseId() string {
	return genMD5ID("enterprise")
}

func GenAppId() string {
	return genMD5ID("app")
}

func GenUserId() string {
	return genMD5ID("user")
}
