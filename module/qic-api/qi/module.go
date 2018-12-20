package qi

import (
	"emotibot.com/emotigo/module/admin-api/util"
)

var (
	// ModuleInfo is needed for module define
	ModuleInfo util.ModuleInfo
)

func init() {
	ModuleInfo = util.ModuleInfo{
		ModuleName: "qi",
		EntryPoints: []util.EntryPoint{
			util.NewEntryPoint("POST", "groups", []string{}, handleCreateGroup),
			util.NewEntryPoint("GET", "groups", []string{}, handleGetGroups),
			util.NewEntryPoint("GET", "groups/{id:[0-9]+}", []string{}, handleGetGroup),
			util.NewEntryPoint("PUT", "groups/{id:[0-9]+}", []string{}, handleUpdateGroup),
			util.NewEntryPoint("DELETE", "groups/{id:[0-9]+}", []string{}, handleDeleteGroup),
		},
	}
}