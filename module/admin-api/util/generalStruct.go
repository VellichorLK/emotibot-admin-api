package util

import (
	"net/http"

	"emotibot.com/emotigo/module/admin-api/ApiError"
)

// EntryPoint is used in every module define
type EntryPoint struct {
	AllowMethod    string
	EntryPath      string
	Callback       func(w http.ResponseWriter, r *http.Request)
	Version        int
	Command        []string
	CheckAppID     bool
	CheckAuthToken bool
}

// NewEntryPoint create new instance of EntryPoint with version 1
func NewEntryPoint(method string, path string, cmd []string, callback func(w http.ResponseWriter, r *http.Request)) EntryPoint {
	entrypoint := EntryPoint{}
	entrypoint.AllowMethod = method
	entrypoint.EntryPath = path
	entrypoint.Callback = callback
	entrypoint.Version = 1
	entrypoint.Command = cmd
	entrypoint.CheckAppID = true
	entrypoint.CheckAuthToken = true
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
	entrypoint.CheckAppID = true
	entrypoint.CheckAuthToken = true
	return entrypoint
}

// NewEntryPointWithCustom create new instance of EntryPoint with custom param
// which is (version int, checkAppID bool, checkAuthToken bool)
func NewEntryPointWithCustom(method string, path string, cmd []string, callback func(w http.ResponseWriter, r *http.Request), param ...interface{}) EntryPoint {
	entrypoint := EntryPoint{}
	entrypoint.AllowMethod = method
	entrypoint.EntryPath = path
	entrypoint.Callback = callback
	entrypoint.Command = cmd

	for idx := range param {
		origVal := param[idx]
		switch idx {
		case 0:
			if val, ok := origVal.(int); ok {
				entrypoint.Version = val
			} else {
				entrypoint.Version = 1
			}
		case 1:
			if val, ok := origVal.(bool); ok {
				entrypoint.CheckAppID = val
			} else {
				entrypoint.CheckAppID = true
			}
		case 2:
			if val, ok := origVal.(bool); ok {
				entrypoint.CheckAuthToken = val
			} else {
				entrypoint.CheckAuthToken = true
			}
		}
	}

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
