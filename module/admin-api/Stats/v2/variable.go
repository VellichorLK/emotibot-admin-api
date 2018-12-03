package v2

import "emotibot.com/emotigo/module/admin-api/util"

var (
	// moduleName used to get correct environment name
	moduleName = "statistic"
	// EntryList will be merged in the module controller
	EntryList = []util.EntryPoint{
		util.NewEntryPointWithConfig("POST", "audit/robot", []string{"view"}, handleListRobotAudit, util.EntryConfig{
			Version:     2,
			IgnoreAppID: true,
		}),
		util.NewEntryPointWithConfig("POST", "audit/enterprise", []string{"view"}, handleListEnterpriseAudit, util.EntryConfig{
			Version:     2,
			IgnoreAppID: true,
		}),
		util.NewEntryPointWithConfig("POST", "audit/system", []string{"view"}, handleListSystemAudit, util.EntryConfig{
			Version:     2,
			IgnoreAppID: true,
		}),
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
