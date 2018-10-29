package Robot

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"

	"emotibot.com/emotigo/module/admin-api/ApiError"
	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/admin-api/util/audit"
	"emotibot.com/emotigo/module/admin-api/util/requestheader"
	"emotibot.com/emotigo/pkg/logger"
)

const (
	defaultListPerPage = 30
)

func handleRobotQA(w http.ResponseWriter, r *http.Request) {
	appid := requestheader.GetAppID(r)

	id, err := util.GetMuxIntVar(r, "id")
	if err != nil || id <= 0 {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	info, errCode, err := GetRobotQA(appid, id, 1)
	if errCode != ApiError.SUCCESS {
		util.WriteJSON(w, util.GenRetObj(errCode, err.Error()))
	} else {
		util.WriteJSON(w, util.GenRetObj(errCode, info))
	}
}

func handleRobotQAList(w http.ResponseWriter, r *http.Request) {
	appid := requestheader.GetAppID(r)

	page, err := util.GetParamInt(r, "page")
	if err != nil {
		logger.Info.Printf("Param error: %s", err.Error())
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	listPerPage, err := util.GetParamInt(r, "per_page")
	if err != nil {
		logger.Info.Printf("Param error: %s", err.Error())
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	if page < 0 || listPerPage < 0 {
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	if listPerPage == 0 {
		listPerPage = defaultListPerPage
	}

	var errCode int
	var list *RetQAInfo
	if page == 0 {
		list, errCode, err = GetRobotQAList(appid, 1)
	} else {
		list, errCode, err = GetRobotQAPage(appid, page, listPerPage, 1)
	}
	if err != nil {
		util.WriteJSON(w, util.GenRetObj(errCode, err.Error()))
		return
	}

	util.WriteJSON(w, util.GenRetObj(errCode, list))
}

func handleRobotQAModelRebuild(w http.ResponseWriter, r *http.Request) {
	appid := requestheader.GetAppID(r)
	auditLog := ""
	result := 0

	errCode, err := util.McRebuildRobotQA(appid)
	if errCode != ApiError.SUCCESS {
		util.WriteJSON(w, util.GenRetObj(errCode, err.Error()))
		auditLog = fmt.Sprintf("%s%s%s",
			util.Msg["RobotProfile"], util.Msg["Rebuild"], util.Msg["Error"])
	} else {
		util.WriteJSON(w, util.GenSimpleRetObj(errCode))
		auditLog = fmt.Sprintf("%s%s%s",
			util.Msg["RobotProfile"], util.Msg["Rebuild"], util.Msg["Success"])
		result = 1
	}
	addAudit(r, audit.AuditModuleRobotProfile, audit.AuditOperationEdit, auditLog, result)
}

func handleUpdateRobotQA(w http.ResponseWriter, r *http.Request) {
	appid := requestheader.GetAppID(r)
	auditLog := ""
	result := 0
	errCode := ApiError.SUCCESS
	errMsg := ""
	var retObj interface{}

	failMsg := fmt.Sprintf("%s%s%s",
		util.Msg["Modify"], util.Msg["RobotProfile"], util.Msg["Error"])
	successMsg := fmt.Sprintf("%s%s%s",
		util.Msg["Modify"], util.Msg["RobotProfile"], util.Msg["Success"])

	id, err := util.GetMuxIntVar(r, "id")
	info := loadQAInfoFromContext(r)
	if err != nil || id <= 0 || info == nil {
		http.Error(w, "", http.StatusBadRequest)
		auditLog = fmt.Sprintf("%s: %s%s", failMsg, util.Msg["Request"], util.Msg["Error"])
		errCode = ApiError.REQUEST_ERROR
		addAudit(r, audit.AuditModuleRobotProfile, audit.AuditOperationEdit, auditLog, result)
		util.WriteJSON(w, util.GenRetObj(errCode, retObj))
		return
	}

	origInfo, errCode, err := GetRobotQA(appid, id, 1)
	if err != nil {
		errMsg = ApiError.GetErrorMsg(errCode)
		retObj = err.Error()
		auditLog = fmt.Sprintf("%s: %s", failMsg, errMsg)
	} else {
		errCode, err = UpdateRobotQA(appid, id, info, 1)
		if errCode != ApiError.SUCCESS {
			retObj = err.Error()
			errMsg = ApiError.GetErrorMsg(errCode)
			auditLog = fmt.Sprintf("%s: %s", failMsg, errMsg)
		} else {
			auditLog = fmt.Sprintf("%s: %s",
				successMsg, diffQAInfo(origInfo, info))
			result = 1
		}
	}
	addAudit(r, audit.AuditModuleRobotProfile, audit.AuditOperationEdit, auditLog, result)
	util.WriteJSON(w, util.GenRetObj(errCode, retObj))
}

func loadQAInfoFromContext(r *http.Request) *QAInfo {
	input := &QAInfo{}
	err := util.ReadJSON(r, input)
	if err != nil {
		logger.Info.Printf("Bad request when loading from input: %s", err.Error())
		return nil
	}

	return input
}

func diffQAInfo(origInfo *QAInfo, newInfo *QAInfo) string {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("%s%s \"%s\"\n", util.Msg["Modify"], util.Msg["Question"], origInfo.Question))

	origAnswers := strings.Join(origInfo.Answers, ", ")
	newAnswers := strings.Join(newInfo.Answers, ", ")
	buffer.WriteString(fmt.Sprintf("%s: [%s]\n%s: [%s]",
		util.Msg["Origin"], origAnswers,
		util.Msg["Updated"], newAnswers))

	return buffer.String()
}

func handleRobotQAV2(w http.ResponseWriter, r *http.Request) {
	appid := requestheader.GetAppID(r)

	id, err := util.GetMuxIntVar(r, "id")
	if err != nil || id <= 0 {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	info, errCode, err := GetRobotQA(appid, id, 2)
	if errCode != ApiError.SUCCESS {
		util.WriteJSON(w, util.GenRetObj(errCode, err.Error()))
	} else {
		util.WriteJSON(w, util.GenRetObj(errCode, info))
	}
}

func handleRobotQAListV2(w http.ResponseWriter, r *http.Request) {
	appid := requestheader.GetAppID(r)

	page, err := util.GetParamInt(r, "page")
	if err != nil {
		logger.Info.Printf("Param error: %s", err.Error())
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	listPerPage, err := util.GetParamInt(r, "per_page")
	if err != nil {
		logger.Info.Printf("Param error: %s", err.Error())
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	if page < 0 || listPerPage < 0 {
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	if listPerPage == 0 {
		listPerPage = defaultListPerPage
	}

	var errCode int
	var list *RetQAInfo
	if page == 0 {
		list, errCode, err = GetRobotQAList(appid, 2)
	} else {
		list, errCode, err = GetRobotQAPage(appid, page, listPerPage, 2)
	}
	if err != nil {
		util.WriteJSON(w, util.GenRetObj(errCode, err.Error()))
		return
	}

	util.WriteJSON(w, util.GenRetObj(errCode, list))
}

func handleUpdateRobotQAV2(w http.ResponseWriter, r *http.Request) {
	appid := requestheader.GetAppID(r)
	auditLog := ""
	result := 0
	errCode := ApiError.SUCCESS
	errMsg := ""
	var retObj interface{}

	failMsg := fmt.Sprintf("%s%s%s",
		util.Msg["Modify"], util.Msg["RobotProfile"], util.Msg["Error"])
	successMsg := fmt.Sprintf("%s%s%s",
		util.Msg["Modify"], util.Msg["RobotProfile"], util.Msg["Success"])

	id, err := util.GetMuxIntVar(r, "id")
	info := loadQAInfoFromContext(r)
	if err != nil || id <= 0 || info == nil {
		http.Error(w, "", http.StatusBadRequest)
		auditLog = fmt.Sprintf("%s: %s%s", failMsg, util.Msg["Request"], util.Msg["Error"])
		errCode = ApiError.REQUEST_ERROR
		addAudit(r, audit.AuditModuleRobotProfile, audit.AuditOperationEdit, auditLog, result)
		util.WriteJSON(w, util.GenRetObj(errCode, retObj))
		return
	}

	origInfo, errCode, err := GetRobotQA(appid, id, 2)
	if err != nil {
		errMsg = ApiError.GetErrorMsg(errCode)
		retObj = err.Error()
		auditLog = fmt.Sprintf("%s: %s", failMsg, errMsg)
	} else {
		errCode, err = UpdateRobotQA(appid, id, info, 2)
		if errCode != ApiError.SUCCESS {
			retObj = err.Error()
			errMsg = ApiError.GetErrorMsg(errCode)
			auditLog = fmt.Sprintf("%s: %s", failMsg, errMsg)
		} else {
			auditLog = fmt.Sprintf("%s: %s",
				successMsg, diffQAInfo(origInfo, info))
			result = 1
		}
	}
	addAudit(r, audit.AuditModuleRobotProfile, audit.AuditOperationEdit, auditLog, result)
	util.WriteJSON(w, util.GenRetObj(errCode, retObj))

	mcCode, mcErr := util.McRebuildRobotQA(appid)
	if mcErr != nil {
		errMsg = "SUCCESS"
		if mcErr != nil {
			errMsg = mcErr.Error()
		}
		logger.Info.Printf("Call multicustomer result: %d, %s", mcCode, errMsg)
	}
}
