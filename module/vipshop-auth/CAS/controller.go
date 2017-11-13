package CAS

import (
	"emotibot.com/emotigo/module/vipshop-auth/util"
	"github.com/kataras/iris/context"
)

var (
	// ModuleInfo is needed for module define
	ModuleInfo util.ModuleInfo
)

func init() {
	ModuleInfo = util.ModuleInfo{
		ModuleName: "cas",
		EntryPoints: []util.EntryPoint{
			util.NewEntryPoint("POST", "login", handleLogin),
		},
	}
}

func handleLogin(ctx context.Context) {

}
