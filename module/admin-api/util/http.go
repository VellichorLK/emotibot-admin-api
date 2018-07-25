package util

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"emotibot.com/emotigo/module/admin-api/util/AdminErrors"
	"github.com/gorilla/mux"
)

func WriteJSON(w http.ResponseWriter, obj interface{}) error {
	js, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(js)
	return nil
}

func ReadJSON(r *http.Request, target interface{}) error {
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	err := decoder.Decode(target)
	return err
}

func GetMuxVar(r *http.Request, key string) string {
	vars := mux.Vars(r)
	return vars[key]
}

func GetMuxIntVar(r *http.Request, key string) (int, error) {
	vars := mux.Vars(r)
	strVal := vars[key]
	if strVal == "" {
		return 0, fmt.Errorf("Invalid key %s", key)
	}
	return strconv.Atoi(vars[key])
}

func GetParamInt(r *http.Request, key string) (int, error) {
	return strconv.Atoi(r.URL.Query().Get(key))
}

func WriteJSONWithStatus(w http.ResponseWriter, obj interface{}, status int) error {
	js, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Status", fmt.Sprintf("%d", status))
	w.WriteHeader(status)
	w.Write(js)
	return nil
}

func WriteWithStatus(w http.ResponseWriter, content string, status int) {
	w.Header().Set("X-Status", fmt.Sprintf("%d", status))
	w.WriteHeader(status)
	w.Write([]byte(content))
}

func Return(w http.ResponseWriter, adminErr AdminErrors.AdminError, ret interface{}) error {
	var obj RetObj
	status := http.StatusOK
	if adminErr != nil {
		obj = RetObj{
			Status:  adminErr.Errno(),
			Message: adminErr.String(),
			Result:  ret,
		}
		status = AdminErrors.GetReturnStatus(adminErr.Errno())
	} else {
		obj = RetObj{
			Status:  AdminErrors.ErrnoSuccess,
			Message: "",
			Result:  ret,
		}
	}
	js, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	w.Header().Set("X-Status", fmt.Sprintf("%d", status))
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	w.Write(js)
	return nil
}

func ReturnError(w http.ResponseWriter, errno int, msg string) error {
	return Return(w, AdminErrors.New(errno, msg), nil)
}
