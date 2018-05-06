package util

import (
	_ "github.com/go-sql-driver/mysql"
)

const (
	AuditOperationAdd    = "0" // "新增"
	AuditOperationEdit   = "1" // "修改"
	AuditOperationDelete = "2" // "删除"
	AuditOperationImport = "3" // "导入"
	AuditOperationExport = "4" // "导出"
	// AuditOperationRollback = "5" "回复", 目前無相關行為
	AuditOperationLogin = "6" // "登入"

	AuditModuleBotMessage     = "0" // "话术设置"
	AuditModuleFunctionSwitch = "1" // "技能设置"
	AuditModuleQA             = "2" // "问答库"
	AuditModuleRobotProfile   = "3" // "形象设置"
	AuditModuleSwitchList     = "4" // "开关管理"
	AuditModuleDictionary     = "5" // "词库管理"
	AuditModuleStatistics     = "6" // "数据管理"
	AuditModuleMembers        = "7" // "用户管理"
	AuditModuleRole           = "8" // "角色管理"
)

type auditLog struct {
	UserID    string
	UserIP    string
	Module    string
	Operation string
	Content   string
	Result    int
}

var auditChannel chan auditLog

// AddAuditLog will add audit log to mysql-audit
func AddAuditLog(userID string, userIP string, module string, operation string, content string, result int) error {
	if auditChannel == nil {
		auditChannel = make(chan auditLog)
		go logRoutine()
	}
	log := auditLog{userID, userIP, module, operation, content, result}
	auditChannel <- log
	return nil
}

func logRoutine() {
	for {
		log := <-auditChannel
		addAuditLog(log)
	}
}

func addAuditLog(log auditLog) {
	auditDB := GetAuditDB()
	if auditDB == nil {
		LogError.Printf("Audit DB connection hasn't init")
		return
	}

	_, err := auditDB.Exec("insert audit_record(user_id, ip_source, module, operation, content, result) values (?, ?, ?, ?, ?, ?)",
		log.UserID, log.UserIP, log.Module, log.Operation, log.Content, log.Result)
	if err != nil {
		LogError.Printf("insert audit fail: %s", err.Error())
	}
}
