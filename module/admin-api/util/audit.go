package util

import (
	"fmt"
	"net/http"

	"emotibot.com/emotigo/module/admin-api/util/requestheader"
	"emotibot.com/emotigo/pkg/logger"
	_ "github.com/go-sql-driver/mysql"
)

const (
	AuditOperationAdd    = "0" // "新增"
	AuditOperationEdit   = "1" // "修改"
	AuditOperationDelete = "2" // "删除"
	AuditOperationImport = "3" // "导入"
	AuditOperationExport = "4" // "导出"
	// AuditOperationRollback = "5" "回复", 目前無相關行為
	AuditOperationLogin    = "6" // "登入"
	AuditOperationPublish  = "7" // 發布
	AuditOperationActive   = "8" // 啟動
	AuditOperationDeactive = "9" // 關閉

	AuditModuleBotMessage     = "0"  // "话术设置"
	AuditModuleFunctionSwitch = "1"  // "技能设置"
	AuditModuleQA             = "2"  // "问答库"
	AuditModuleRobotProfile   = "3"  // "形象设置"
	AuditModuleSwitchList     = "4"  // "开关管理"
	AuditModuleDictionary     = "5"  // "词库管理"
	AuditModuleStatistics     = "6"  // "数据管理"
	AuditModuleMembers        = "7"  // "用户管理"
	AuditModuleRole           = "8"  // "角色管理"
	AuditModuleTaskEngine     = "9"  // "Task-Engine"
	AuditModuleIntentEngine   = "10" // "意圖引擎"
)

var moduleMap = map[string]string{
	"intents": AuditModuleIntentEngine,
}

type auditLog struct {
	AppID     string
	UserID    string
	UserIP    string
	Module    string
	Operation string
	Content   string
	Result    int
}

func (log auditLog) toString() string {
	return fmt.Sprintf("[%s] %s@%s %s,%s: %s [%d]", log.AppID, log.UserID, log.UserIP, log.Module, log.Operation, log.Content, log.Result)
}

var auditChannel chan auditLog
var (
	AuditCustomHeader = "X-AuditModule"
)

// AddAuditFromRequest will get userID, userIP and appid from request, and add audit log to mysql-audit
func AddAuditFromRequest(r *http.Request, module string, operation string, msg string, result int) {
	userID := requestheader.GetUserID(r)
	userIP := requestheader.GetUserIP(r)
	appid := requestheader.GetAppID(r)

	AddAuditLog(appid, userID, userIP, module, operation, msg, result)
}

func AddAuditFromRequestAuto(r *http.Request, msg string, result int) {
	userID := requestheader.GetUserID(r)
	userIP := requestheader.GetUserIP(r)
	appid := requestheader.GetAppID(r)

	module := r.Header.Get(AuditCustomHeader)
	operation := ""
	switch r.Method {
	case http.MethodPost:
		operation = AuditOperationAdd
	case http.MethodDelete:
		operation = AuditOperationDelete
	case http.MethodPatch:
		fallthrough
	case http.MethodPut:
		operation = AuditOperationEdit
	default:
		operation = ""
	}

	moduleCode := moduleMap[module]

	AddAuditLog(appid, userID, userIP, moduleCode, operation, msg, result)
}

// AddAuditLog will add audit log to mysql-audit
func AddAuditLog(appid string, userID string, userIP string, module string, operation string, content string, result int) error {
	if auditChannel == nil {
		auditChannel = make(chan auditLog)
		go logRoutine()
	}
	log := auditLog{appid, userID, userIP, module, operation, content, result}
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
		logger.Error.Printf("Audit DB connection hasn't init")
		return
	}
	_, errWithAppID := auditDB.Exec("insert audit_record(appid, user_id, ip_source, module, operation, content, result) values (?, ?, ?, ?, ?, ?, ?)",
		log.AppID, log.UserID, log.UserIP, log.Module, log.Operation, log.Content, log.Result)
	if errWithAppID == nil {
		return
	}
	logger.Warn.Println("Schema of audit_record should be upgraded")

	_, errWithoutAppID := auditDB.Exec("insert audit_record(user_id, ip_source, module, operation, content, result) values (?, ?, ?, ?, ?, ?)",
		log.UserID, log.UserIP, log.Module, log.Operation, log.Content, log.Result)
	if errWithAppID != nil && errWithoutAppID != nil {
		logger.Error.Printf("insert audit fail: %s", errWithoutAppID.Error())
	}
}
