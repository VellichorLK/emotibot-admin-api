package util

import (
	"errors"

	_ "github.com/go-sql-driver/mysql"
)

const (
	AuditOperationAdd    = "新增"
	AuditOperationEdit   = "修改"
	AuditOperationDelete = "删除"
	AuditOperationImport = "导入"
	AuditOperationExport = "导出"
	AuditOperationLogin  = "登入"

	AuditModuleBotMessage     = "话术设置"
	AuditModuleFunctionSwitch = "技能设置"
	AuditModuleQASetting      = "问答库"
	AuditModuleRobotProfile   = "形象设置"
	AuditModuleSwitchList     = "开关管理"
	AuditModuleDictionary     = "词库管理"
	AuditModuleStatistics     = "数据管理"
	AuditModuleMembers        = "用户管理"
	AuditModuleRole           = "角色管理"
)

// AddAuditLog will add audit log to mysql-audit
func AddAuditLog(userID string, userIP string, module string, operation string, content string, result int) error {
	auditDB := GetAuditDB()
	if auditDB == nil {
		LogInfo.Printf("Audit DB connection hasn't init")
		return errors.New("DB not init")
	}

	_, err := auditDB.Query("insert audit_record(user_id, ip_source, module, operation, content, result) values (?, ?, ?, ?, ?, ?)", userID, userIP, module, operation, content, result)
	if err != nil {
		LogInfo.Printf("insert audit fail: %s", err.Error())
		return err
	}

	return nil
}
