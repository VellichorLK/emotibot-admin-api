package v2

import "emotibot.com/emotigo/module/admin-api/util"

var (
	// moduleName used to get correct environment name
	moduleName = "statistic"
	// EntryList will be merged in the module controller
	EntryList = []util.EntryPoint{
		util.NewEntryPointWithCustom("POST", "audit/robot", []string{"view"}, handleListRobotAudit, 2, false),
		util.NewEntryPointWithCustom("POST", "audit/enterprise", []string{"view"}, handleListEnterpriseAudit, 2, false),
		util.NewEntryPointWithCustom("POST", "audit/system", []string{"view"}, handleListSystemAudit, 2, false),
	}
)

var robotAuditHeaders = []*AuditHeader{
	&AuditHeader{
		Text: "用戶 ID",
		ID:   "user",
	},
	&AuditHeader{
		Text: "操作模塊",
		ID:   "module",
	},
	&AuditHeader{
		Text: "操作類型",
		ID:   "operation",
	},
	&AuditHeader{
		Text: "動作描述",
		ID:   "content",
	},
	&AuditHeader{
		Text: "動作結果",
		ID:   "result",
	},
	&AuditHeader{
		Text: "發生時間",
		ID:   "create_time",
	},
	&AuditHeader{
		Text: "IP 地址",
		ID:   "user_ip",
	},
}
