package auth

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"time"
)

const (
	const_appid_length        int = 32 // md5sum length
	const_enterpriseid_length int = 32 // md5sum length
	const_userid_length       int = 32 // md5sum length
)

type ErrStruct struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

// validation
func IsValidAppId(aid string) bool {
	if len(aid) != const_appid_length {
		return false
	}
	return true
}

func IsValidEnterpriseId(eid string) bool {
	if len(eid) != const_enterpriseid_length {
		return false
	}
	return true
}

func IsValidUserId(uid string) bool {
	if len(uid) != const_userid_length {
		return false
	}
	return true
}

func RespJson(w http.ResponseWriter, es interface{}) {
	js, err := json.Marshal(es)
	if HandleHttpError(http.StatusInternalServerError, err, w) {
		LogError.Printf("jsonize %s failed. %s", es, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	LogInfo.Printf("js: %s", js)
	fmt.Fprintf(w, string(js))
}

// return true: invalid
func HandleHttpMethodError(request_method string, allowed_method []string, w http.ResponseWriter) bool {
	for _, m := range allowed_method {
		if request_method == m {
			return false
		}
	}
	HandleHttpError(http.StatusMethodNotAllowed, errors.New("Method Not Allowed"), w)
	return true
}

func HandleError(err_code int, err error, w http.ResponseWriter) bool {
	if err == nil {
		return false
	}
	_, fn, _, _ := runtime.Caller(1)
	LogError.Printf("%s: %s", fn, err.Error())
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

	_, fn, _, _ := runtime.Caller(1)
	LogError.Printf("%s: %s", fn, err.Error())
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
