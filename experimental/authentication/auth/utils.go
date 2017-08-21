package auth

import (
	"crypto/md5"
	"database/sql"
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
	const_roleid_length       int = 32 // md5sum length
)

type ErrStruct struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

// NullableString is used for compatiable using with mysql and json
type NullableString struct {
	sql.NullString
}

// MarshalJSON is used for json.stringify of NullableString
func (v NullableString) MarshalJSON() ([]byte, error) {
	if v.Valid {
		return json.Marshal(v.String)
	}
	return json.Marshal(nil)
}

// UnmarshalJSON is used for parse NullableString from json string
func (v *NullableString) UnmarshalJSON(data []byte) error {
	// Unmarshalling into a pointer will let us detect null
	var s *string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if s != nil {
		v.Valid = true
		v.String = *s
	} else {
		v.Valid = false
	}
	return nil
}

// validation
func IsValidAppId(aid string) bool {
	if len(aid) != const_appid_length {
		return false
	}
	if !HasOnlyNumEng(aid) {
		return false
	}
	return true
}

func IsValidEnterpriseId(eid string) bool {
	if len(eid) != const_enterpriseid_length {
		return false
	}
	if !HasOnlyNumEng(eid) {
		return false
	}
	return true
}

func IsValidUserId(uid string) bool {
	if len(uid) != const_userid_length {
		return false
	}
	if !HasOnlyNumEng(uid) {
		return false
	}
	return true
}

func IsValidRoleId(rid string) bool {
	if len(rid) != const_roleid_length {
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

func HandleSuccess(w http.ResponseWriter, result interface{}) bool {
	ret := make(map[string]interface{})
	ret["status"] = 0
	ret["message"] = "success"
	ret["result"] = result
	RespJson(w, ret)
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

func GenRoleId() string {
	return genMD5ID("role")
}

func HasOnlyNumEng(input string) bool {
	for _, c := range input {
		if (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') && (c < '0' || c > '9') {
			return false
		}
	}
	return true
}
