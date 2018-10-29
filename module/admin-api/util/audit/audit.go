package audit

import (
	"fmt"
	"net/http"

	"emotibot.com/emotigo/module/admin-api/util/localemsg"

	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/admin-api/util/requestheader"
	"emotibot.com/emotigo/pkg/logger"
	_ "github.com/go-sql-driver/mysql"
)

const (
	AuditOperationAdd      = "add"     // "新增"
	AuditOperationEdit     = "edit"    // "修改"
	AuditOperationDelete   = "delete"  // "删除"
	AuditOperationImport   = "import"  // "导入"
	AuditOperationExport   = "export"  // "导出"
	AuditOperationLogin    = "login"   // "登入"
	AuditOperationPublish  = "publish" // 發布
	AuditOperationActive   = "on"      // 啟動
	AuditOperationDeactive = "off"     // 關閉

	AuditModuleSSM               = "ssm"
	AuditModuleFAQ               = "faq"
	AuditModuleQALabel           = "qa_label"           //  "標籤管理"
	AuditModuleTaskEngine        = "task_engine"        //  "任務引擎"
	AuditModuleIntentManage      = "intent_manage"      //  "意圖引擎"
	AuditModuleWordbank          = "wordbank"           //  "詞庫"
	AuditModuleStatisticDaily    = "statistic_daily"    //  "日誌管理"
	AuditModuleStatisticAnalysis = "statistic_analysis" //  "統計分析"
	AuditModuleAudit             = "statistic_audit"
	AuditModuleRobotProfile      = "robot_profile"    //  "機器人形象"
	AuditModuleRobotChatSkill    = "robot_chat_skill" //  "話術設置"
	AuditModuleRobotFunction     = "robot_function"   //  "技能設置"
	AuditModuleRobotCommand      = "robot_command"    //  "指令設置"
	AuditModuleIntegration       = "integration"      //  "接入部署"
	AuditModuleManageUser        = "manage_user"
	AuditModuleManageRobot       = "manage_robot"
	AuditModuleManageAdmin       = "manage_admin"
	AuditModuleManageEnterprise  = "manage_enterprise"
)

var moduleMap = map[string]string{
	"intents":    AuditModuleIntentManage,
	"dictionary": AuditModuleWordbank,
}

var operationLocaleKeyMap = map[string]string{
	"0": "AuditOperationAdd",
	"1": "AuditOperationEdit",
	"2": "AuditOperationDelete",
	"3": "AuditOperationImport",
	"4": "AuditOperationExport",
	"6": "AuditOperationLogin",
	"7": "AuditOperationPublish",
	"8": "AuditOperationActive",
	"9": "AuditOperationDeactive",

	AuditOperationAdd:      "AuditOperationAdd",
	AuditOperationEdit:     "AuditOperationEdit",
	AuditOperationDelete:   "AuditOperationDelete",
	AuditOperationImport:   "AuditOperationImport",
	AuditOperationExport:   "AuditOperationExport",
	AuditOperationLogin:    "AuditOperationLogin",
	AuditOperationPublish:  "AuditOperationPublish",
	AuditOperationActive:   "AuditOperationActive",
	AuditOperationDeactive: "AuditOperationDeactive",
}

var moduleLocalKeyMap = map[string]string{
	"2":  "AuditModuleSSM",
	"9":  "AuditModuleTaskEngine",
	"10": "AuditModuleIntentManage",
	"5":  "AuditModuleWordbank",
	"6":  "AuditModuleStatisticAnalysis",
	"3":  "AuditModuleRobotProfile",
	"0":  "AuditModuleRobotChatSkill",
	"1":  "AuditModuleRobotFunction",

	AuditModuleSSM:               "AuditModuleSSM",
	AuditModuleFAQ:               "AuditModuleFAQ",
	AuditModuleQALabel:           "AuditModuleQALabel",
	AuditModuleTaskEngine:        "AuditModuleTaskEngine",
	AuditModuleIntentManage:      "AuditModuleIntentManage",
	AuditModuleWordbank:          "AuditModuleWordbank",
	AuditModuleStatisticDaily:    "AuditModuleStatisticDaily",
	AuditModuleStatisticAnalysis: "AuditModuleStatisticAnalysis",
	AuditModuleAudit:             "AuditModuleAudit",
	AuditModuleRobotProfile:      "AuditModuleRobotProfile",
	AuditModuleRobotChatSkill:    "AuditModuleRobotChatSkill",
	AuditModuleRobotFunction:     "AuditModuleRobotFunction",
	AuditModuleRobotCommand:      "AuditModuleRobotCommand",
	AuditModuleIntegration:       "AuditModuleIntegration",
	AuditModuleManageUser:        "AuditModuleManageUser",
	AuditModuleManageRobot:       "AuditModuleManageRobot",
	AuditModuleManageAdmin:       "AuditModuleManageAdmin",
	AuditModuleManageEnterprise:  "AuditModuleManageEnterprise",
}

type auditLog struct {
	EnterpriseID string
	AppID        string
	UserID       string
	UserIP       string
	Module       string
	Operation    string
	Content      string
	Result       int
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
	enterpriseID := requestheader.GetEnterpriseID(r)

	AddAuditLog(enterpriseID, appid, userID, userIP, module, operation, msg, result)
}

func AddAuditFromRequestAuto(r *http.Request, msg string, result int) {
	userID := requestheader.GetUserID(r)
	userIP := requestheader.GetUserIP(r)
	appid := requestheader.GetAppID(r)
	enterpriseID := requestheader.GetEnterpriseID(r)
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
	AddAuditLog(enterpriseID, appid, userID, userIP, moduleCode, operation, msg, result)
}

// AddAuditLog will add audit log to mysql-audit
func AddAuditLog(enterpriseID string, appid string, userID string, userIP string, module string, operation string, content string, result int) error {
	if auditChannel == nil {
		auditChannel = make(chan auditLog)
		go logRoutine()
	}
	log := auditLog{enterpriseID, appid, userID, userIP, module, operation, content, result}
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
	auditDB := util.GetAuditDB()
	if auditDB == nil {
		logger.Error.Printf("Audit DB connection hasn't init")
		return
	}

	logger.Trace.Printf("Get audit: %+v", log)

	_, errWithEnterprise := auditDB.Exec("insert audit_record(enterprise, appid, user_id, ip_source, module, operation, content, result) values (?, ?, ?, ?, ?, ?, ?, ?)",
		log.EnterpriseID, log.AppID, log.UserID, log.UserIP, log.Module, log.Operation, log.Content, log.Result)
	if errWithEnterprise == nil {
		return
	}

	logger.Warn.Println("Schema of audit_record should be upgraded or err happen,", errWithEnterprise.Error())
	_, errWithAppID := auditDB.Exec("insert audit_record(appid, user_id, ip_source, module, operation, content, result) values (?, ?, ?, ?, ?, ?, ?)",
		log.AppID, log.UserID, log.UserIP, log.Module, log.Operation, log.Content, log.Result)
	if errWithAppID == nil {
		return
	}
	logger.Warn.Println("Schema of audit_record should be upgraded,", errWithAppID.Error())

	_, errWithoutAppID := auditDB.Exec("insert audit_record(enterprise, user_id, ip_source, module, operation, content, result) values (?, ?, ?, ?, ?, ?)",
		log.UserID, log.UserIP, log.Module, log.Operation, log.Content, log.Result)
	if errWithAppID != nil && errWithoutAppID != nil {
		logger.Error.Printf("insert audit fail: %s", errWithoutAppID.Error())
	}
}

// GetAuditModuleName will get module name by locale
func GetAuditModuleName(locale, module string) string {
	if localeKey, ok := moduleLocalKeyMap[module]; ok {
		msg := localemsg.Get(locale, localeKey)
		if msg != "" {
			return msg
		}
	}
	return module
}

// GetAuditOperationName will get operation name by locale
func GetAuditOperationName(locale, operation string) string {
	if localeKey, ok := operationLocaleKeyMap[operation]; ok {
		msg := localemsg.Get(locale, localeKey)
		if msg != "" {
			return msg
		}
	}
	return operation
}
