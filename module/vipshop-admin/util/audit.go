package util

import (
	"errors"

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

// AddAuditLog will add audit log to mysql-audit
func AddAuditLog(userID string, userIP string, module string, operation string, content string, result int) error {
	auditDB := GetAuditDB()
	if auditDB == nil {
		LogError.Printf("Audit DB connection hasn't init")
		return errors.New("DB not init")
	}

	_, err := auditDB.Exec("insert audit_record(user_id, ip_source, module, operation, content, result) values (?, ?, ?, ?, ?, ?)", userID, userIP, module, operation, content, result)
	if err != nil {
		LogError.Printf("insert audit fail: %s", err.Error())
		return err
	}

	return nil
}
