package audit

import (
	"emotibot.com/emotigo/module/token-auth/dao"
	"emotibot.com/emotigo/module/token-auth/internal/data"
	"emotibot.com/emotigo/module/token-auth/internal/util"
)

const (
	AuditOperationAdd    = "add"    // "新增"
	AuditOperationEdit   = "edit"   // "修改"
	AuditOperationDelete = "delete" // "删除"
	AuditOperationLogin  = "login"  // "登入"

	AuditModuleManageUser       = "manage_user"
	AuditModuleManageRobot      = "manage_robot"
	AuditModuleManageAdmin      = "manage_admin"
	AuditModuleManageEnterprise = "manage_enterprise"
)

var auditDB dao.DB

func SetDB(db dao.DB) {
	auditDB = db
}

var auditChannel chan data.AuditLog

// AddAuditLog will add audit log to mysql-audit
func AddAuditLog(enterpriseID string, appid string, userID string, userIP string, module string, operation string, content string, result int) error {
	if auditChannel == nil {
		auditChannel = make(chan data.AuditLog)
		go logRoutine()
	}
	log := data.AuditLog{
		EnterpriseID: enterpriseID,
		AppID:        appid,
		UserID:       userID,
		UserIP:       userIP,
		Module:       module,
		Operation:    operation,
		Content:      content,
		Result:       result,
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
