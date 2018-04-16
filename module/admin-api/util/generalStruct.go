package util

import (
	"net/http"

	"emotibot.com/emotigo/module/admin-api/ApiError"
)

// EntryPoint is used in every module define
type EntryPoint struct {
	AllowMethod string
	EntryPath   string
	Callback    func(w http.ResponseWriter, r *http.Request)
	Version     int
	Command     []string
}

// NewEntryPoint create new instance of EntryPoint with version 1
func NewEntryPoint(method string, path string, cmd []string, callback func(w http.ResponseWriter, r *http.Request)) EntryPoint {
	entrypoint := EntryPoint{}
	entrypoint.AllowMethod = method
	entrypoint.EntryPath = path
	entrypoint.Callback = callback
	entrypoint.Version = 1
	entrypoint.Command = cmd
	return entrypoint
}

// NewEntryPointWithVer create new instance of EntryPoint with custom version
func NewEntryPointWithVer(method string, path string, cmd []string, callback func(w http.ResponseWriter, r *http.Request), version int) EntryPoint {
	entrypoint := EntryPoint{}
	entrypoint.AllowMethod = method
	entrypoint.EntryPath = path
	entrypoint.Callback = callback
	entrypoint.Version = version
	entrypoint.Command = cmd
	return entrypoint
}

// ModuleInfo if used to defined
type ModuleInfo struct {
	// ModuleName is needed for every Dictionary for get path
	ModuleName string

	// EntryPoints is needed for every Dictionary for set route
	EntryPoints []EntryPoint

	Environments map[string]string
}

type RetObj struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Result  interface{} `json:"result"`
}

func GenRetObj(status int, result interface{}) RetObj {
	LogTrace.Printf("status: [%d] msg: [%s] obj: [%+v]", status, ApiError.GetErrorMsg(status), result)
	return RetObj{
		Status:  status,
		Message: ApiError.GetErrorMsg(status),
		Result:  result,
	}
}

func GenRetObjWithCustomMsg(status int, message string, result interface{}) RetObj {
	LogTrace.Printf("status: [%d] msg: [%s] obj: [%+v]", status, message, result)
	return RetObj{
		Status:  status,
		Message: message,
		Result:  result,
	}
}

func GenSimpleRetObj(status int) RetObj {
	LogTrace.Printf("status: [%d] msg: [%s]", status, ApiError.GetErrorMsg(status))
	return RetObj{
		Status:  status,
		Message: ApiError.GetErrorMsg(status),
	}
}
