package v2

import (
	"net/http"

	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/admin-api/util/AdminErrors"
	"emotibot.com/emotigo/module/admin-api/util/requestheader"
	"emotibot.com/emotigo/module/admin-api/util/audit"
)

func handleListRobotAudit(w http.ResponseWriter, r *http.Request) {
	filter := AuditInput{}
	jsonErr := util.ReadJSON(r, &filter)
	if jsonErr != nil {
		util.Return(w, AdminErrors.New(AdminErrors.ErrnoRequestError, jsonErr.Error()), nil)
		return
	}

	locale := requestheader.GetLocale(r)
	records, err := GetRobotAuditRecord(&filter, locale)
	if err != nil {
		util.Return(w, err, nil)
	} else {
		util.Return(w, nil, records)
	}
}
func handleListEnterpriseAudit(w http.ResponseWriter, r *http.Request) {
	filter := AuditInput{}
	jsonErr := util.ReadJSON(r, &filter)
	if jsonErr != nil {
		util.Return(w, AdminErrors.New(AdminErrors.ErrnoRequestError, jsonErr.Error()), nil)
		return
	}

	locale := requestheader.GetLocale(r)
	records, err := GetEnterpriseAuditRecord(&filter, locale)
	if err != nil {
		util.Return(w, err, nil)
	} else {
		util.Return(w, nil, records)
	}
}
func handleListSystemAudit(w http.ResponseWriter, r *http.Request) {
	filter := AuditInput{}
	jsonErr := util.ReadJSON(r, &filter)
	if jsonErr != nil {
		util.Return(w, AdminErrors.New(AdminErrors.ErrnoRequestError, jsonErr.Error()), nil)
		return
	}

	locale := requestheader.GetLocale(r)
	records, err := GetSystemAuditRecord(&filter, locale)
	if err != nil {
		util.Return(w, err, nil)
	} else {
		util.Return(w, nil, records)
	}
}

func handleAddAuditRecord(w http.ResponseWriter, r *http.Request) {
	filter := AuditLog{}
	jsonErr := util.ReadJSON(r, &filter)
	if jsonErr != nil {
		util.Return(w, AdminErrors.New(AdminErrors.ErrnoJSONParse, jsonErr.Error()), nil)
		return
	}

	checkMsg := ""
	enterpriseId := filter.EnterpriseID
	if enterpriseId == "" {
		checkMsg = checkMsg + "操作记录请求参数：enterprise 不能为空字串; "
	}
	appid := filter.AppID
	if enterpriseId == "" {
		checkMsg = checkMsg + "操作记录请求参数：appid 不能为空字串; "
	}
	userId := filter.UserID
	if enterpriseId == "" {
		checkMsg = checkMsg + "操作记录请求参数：user 不能为空字串; "
	}
	userIp := filter.UserIP
	if enterpriseId == "" {
		checkMsg = checkMsg + "操作记录请求参数：user_ip 不能为空字串; "
	}
	module := filter.Module
	if enterpriseId == "" {
		checkMsg = checkMsg + "操作记录请求参数：module 不能为空字串; "
	}
	operation := filter.Operation
	if enterpriseId == "" {
		checkMsg = checkMsg + "操作记录请求参数：operation 不能为空字串; "
	}
	operationDesc := filter.Content
	if enterpriseId == "" {
		checkMsg = checkMsg + "操作记录请求参数：content 不能为空字串; "
	}
	operationResult := filter.ResultStr
	if enterpriseId == "" {
		checkMsg = checkMsg + "操作记录请求参数：result 不能为空字串; "
	}

	if checkMsg == "" {
		auditRet := 0
		if operationResult == "success" {
			auditRet = 1
		}

		err := audit.AddAuditLog(enterpriseId, appid, userId, userIp, module, operation, operationDesc, auditRet)

		if err != nil {
			util.Return(w, AdminErrors.New(AdminErrors.ErrnoDBError, "fail"), "添加操作记录失败")
		} else {
			util.Return(w, AdminErrors.New(AdminErrors.ErrnoSuccess, "success"), "添加操作记录成功")
		}
	} else {
		util.Return(w, AdminErrors.New(AdminErrors.ErrnoRequestError, "fail"), checkMsg)
	}

}
