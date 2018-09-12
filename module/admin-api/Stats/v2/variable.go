package v2

import "emotibot.com/emotigo/module/admin-api/util"

var (
	// moduleName used to get correct environment name
	moduleName = "statistic"
	// EntryList will be merged in the module controller
	EntryList = []util.EntryPoint{
		util.NewEntryPointWithVer("POST", "audit/robot", []string{"view"}, handleListRobotAudit, 2),
		util.NewEntryPointWithVer("POST", "audit/enterprise", []string{"view"}, handleListEnterpriseAudit, 2),
		util.NewEntryPointWithVer("POST", "audit/system", []string{"view"}, handleListSystemAudit, 2),
	}
)
