package util

import (
	"github.com/kataras/iris/context"
)

// EntryPoint is used in every module define
type EntryPoint struct {
	AllowMethod string
	EntryPath   string
	Callback    func(ctx context.Context)
	Version     int
	Command     []string
}

// NewEntryPoint create new instance of EntryPoint with version 1
func NewEntryPoint(method string, path string, cmd []string, callback func(ctx context.Context)) EntryPoint {
	entrypoint := EntryPoint{}
	entrypoint.AllowMethod = method
	entrypoint.EntryPath = path
	entrypoint.Callback = callback
	entrypoint.Version = 1
	entrypoint.Command = cmd
	return entrypoint
}

// NewEntryPointWithVer create new instance of EntryPoint with version 1
func NewEntryPointWithVer(method string, path string, cmd []string, callback func(ctx context.Context), version int) EntryPoint {
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

func GenRetObj(status int, message string, result interface{}) RetObj {
	LogInfo.Printf("status: [%d] msg: [%s] obj: [%T]", status, message, result)
	return RetObj{
		Status:  status,
		Message: message,
		Result:  result,
	}
}

func GenSimpleRetObj(status int, message string) RetObj {
	LogInfo.Printf("status: [%d] msg: [%s]", status, message)
	return RetObj{
		Status:  status,
		Message: message,
	}
}
