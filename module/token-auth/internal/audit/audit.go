package audit

import (
	"emotibot.com/emotigo/module/token-auth/dao"
	"emotibot.com/emotigo/module/token-auth/internal/data"
	"emotibot.com/emotigo/module/token-auth/internal/util"
)

const (
	AuditOperationAdd    = "0" // "新增"
	AuditOperationEdit   = "1" // "修改"
	AuditOperationDelete = "2" // "删除"
	AuditOperationImport = "3" // "导入"
	AuditOperationExport = "4" // "导出"
	// AuditOperationRollback = "5" "回复", 目前無相關行為
	AuditOperationLogin = "6" // "登入"

	AuditModuleBotMessage     = "0"  // "话术设置"
	AuditModuleFunctionSwitch = "1"  // "技能设置"
	AuditModuleQA             = "2"  // "问答库"
	AuditModuleRobotProfile   = "3"  // "形象设置"
	AuditModuleSwitchList     = "4"  // "开关管理"
	AuditModuleDictionary     = "5"  // "词库管理"
	AuditModuleStatistics     = "6"  // "数据管理"
	AuditModuleMembers        = "7"  // "用户管理"
	AuditModuleRole           = "8"  // "角色管理"
	AuditModuleRobotGroup     = "9"  // "机器人群组"
	AuditModuleRobot          = "10" // "机器人"
)

var auditDB dao.DB

func SetDB(db dao.DB) {
	auditDB = db
}

var auditChannel chan data.AuditLog

// AddAuditLog will add audit log to mysql-audit
func AddAuditLog(appid string, userID string, userIP string, module string, operation string, content string, result int) error {
	if auditChannel == nil {
		auditChannel = make(chan data.AuditLog)
		go logRoutine()
	}

	log := data.AuditLog{
		AppID:     appid,
		UserID:    userID,
		UserIP:    userIP,
		Module:    module,
		Operation: operation,
		Content:   content,
		Result:    result,
	}
	auditChannel <- log

	return nil
}

func logRoutine() {
	for {
		log := <-auditChannel
		addAuditLog(log)
	}
}

func addAuditLog(log data.AuditLog) {
	if auditDB == nil {
		util.LogError.Printf("Audit DB connection hasn't init")
		return
	}

	err := auditDB.AddAuditLog(log)
	if err != nil {
		util.LogError.Printf("Cannot add audit log")
		return
	}
}
