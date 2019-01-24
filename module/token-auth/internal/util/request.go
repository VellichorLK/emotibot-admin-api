package util

import (
	"encoding/json"
	"net/http"

	"emotibot.com/emotigo/module/admin-api/util/AdminErrors"
)

type RetObj struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Result  interface{} `json:"result"`
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

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	w.Write(js)
	return nil
}
