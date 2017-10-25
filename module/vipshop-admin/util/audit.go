package util

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

const (
	mySQLTimeout      string = "10s"
	mySQLWriteTimeout string = "30s"
	mySQLReadTimeout  string = "30s"
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

var (
	auditDB *sql.DB
)

// AuditDBInit should be called before insert all audit log
func AuditDBInit(auditDBURL string, auditDBUser string, auditDBPass string, auditDBDB string) error {
	if len(auditDBURL) == 0 || len(auditDBUser) == 0 || len(auditDBPass) == 0 || len(auditDBDB) == 0 {
		return errors.New("invalid parameters")
	}
	LogInfo.Printf("auditDBURL: %s, auditDBUser: %s, auditDBPass: %s, auditDB_name: %s", auditDBURL, auditDBUser, auditDBPass, auditDBDB)

	url := fmt.Sprintf("%s:%s@tcp(%s)/%s?timeout=%s&readTimeout=%s&writeTimeout=%s&parseTime=true", auditDBUser, auditDBPass, auditDBURL, auditDBDB, mySQLTimeout, mySQLReadTimeout, mySQLWriteTimeout)
	LogInfo.Printf("url: %s", url)

	var err error
	auditDB, err = sql.Open("mysql", url)
	if err != nil {
		LogInfo.Printf("open db(%s) failed: %s", url, err)
		return err
	}
	return nil
}

// AddAuditLog will add audit log to mysql-audit
func AddAuditLog(userID string, userIP string, module string, operation string, content string, result int) error {
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
