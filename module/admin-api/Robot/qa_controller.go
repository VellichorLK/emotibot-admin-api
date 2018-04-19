package Robot

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"

	"emotibot.com/emotigo/module/admin-api/ApiError"
	"emotibot.com/emotigo/module/admin-api/util"
)

const (
	defaultListPerPage = 30
)

func handleRobotQA(w http.ResponseWriter, r *http.Request) {
	appid := util.GetAppID(r)

	id, err := util.GetMuxIntVar(r, "id")
	if err != nil || id <= 0 {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	info, errCode, err := GetRobotQA(appid, id)
	if errCode != ApiError.SUCCESS {
		util.WriteJSON(w, util.GenRetObj(errCode, err.Error()))
	} else {
		util.WriteJSON(w, util.GenRetObj(errCode, info))
	}
}

func handleRobotQAList(w http.ResponseWriter, r *http.Request) {
	appid := util.GetAppID(r)

	page, err := util.GetParamInt(r, "page")
	if err != nil {
		util.LogInfo.Printf("Param error: %s", err.Error())
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	listPerPage, err := util.GetParamInt(r, "per_page")
	if err != nil {
		util.LogInfo.Printf("Param error: %s", err.Error())
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
		list, errCode, err = GetRobotQAList(appid)
	} else {
		list, errCode, err = GetRobotQAPage(appid, page, listPerPage)
	}
	if err != nil {
		util.WriteJSON(w, util.GenRetObj(errCode, err.Error()))
		return
	}

	util.WriteJSON(w, util.GenRetObj(errCode, list))
}

func handleRobotQAModelRebuild(w http.ResponseWriter, r *http.Request) {
	appid := util.GetAppID(r)
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
	addAudit(r, util.AuditModuleRobotProfile, util.AuditOperationEdit, auditLog, result)
}

func handleUpdateRobotQA(w http.ResponseWriter, r *http.Request) {
	appid := util.GetAppID(r)
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
		addAudit(r, util.AuditModuleRobotProfile, util.AuditOperationEdit, auditLog, result)
		util.WriteJSON(w, util.GenRetObj(errCode, retObj))
		return
	}

	origInfo, errCode, err := GetRobotQA(appid, id)
	if err != nil {
		errMsg = ApiError.GetErrorMsg(errCode)
		retObj = err.Error()
		auditLog = fmt.Sprintf("%s: %s", failMsg, errMsg)
	} else {
		errCode, err = UpdateRobotQA(appid, id, info)
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
	addAudit(r, util.AuditModuleRobotProfile, util.AuditOperationEdit, auditLog, result)
	util.WriteJSON(w, util.GenRetObj(errCode, retObj))
}

func loadQAInfoFromContext(r *http.Request) *QAInfo {
	input := &QAInfo{}
	err := util.ReadJSON(r, input)
	if err != nil {
		util.LogInfo.Printf("Bad request when loading from input: %s", err.Error())
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
