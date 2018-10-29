package Service

import (
	"emotibot.com/emotigo/module/admin-api/util"
)

var (
	// ModuleInfo is needed for module define
	ModuleInfo util.ModuleInfo
)

func init() {
	ModuleInfo = util.ModuleInfo{
		ModuleName:  "service",
		EntryPoints: []util.EntryPoint{},
	}
}
