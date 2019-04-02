package Robot

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"

	"emotibot.com/emotigo/module/admin-api/ApiError"
	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/admin-api/util/AdminErrors"
	"emotibot.com/emotigo/module/admin-api/util/audit"
	"emotibot.com/emotigo/module/admin-api/util/requestheader"
	"emotibot.com/emotigo/pkg/logger"
)

func handleRobotQAListV3(w http.ResponseWriter, r *http.Request) {
	appid := requestheader.GetAppID(r)
	locale := requestheader.GetLocale(r)

	logger.Trace.Println("Get robot qa list of", appid)
	qainfos, errno, err := GetRobotQAListV3(appid, locale)
	if err != nil {
		status := ApiError.GetHttpStatus(errno)
		util.WriteJSONWithStatus(w, util.GenRetObj(errno, err.Error()), status)
		logger.Error.Println("Get robot qa list err: ", err.Error())
		return
	}

	util.WriteJSON(w, util.GenRetObj(errno, qainfos))
}

func handleRobotQAV3(w http.ResponseWriter, r *http.Request) {
	appid := requestheader.GetAppID(r)
	qid, err := util.GetMuxIntVar(r, "id")
	if err != nil {
		err := util.GenBadRequestError("ID")
		util.WriteJSONWithStatus(w,
			util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()),
			http.StatusBadRequest)
		return
	}

	logger.Trace.Println("Get robot qa of", appid, qid)
	qainfo, errno, err := GetRobotQAV3(appid, qid)
	if err != nil {
		status := ApiError.GetHttpStatus(errno)
		util.WriteJSONWithStatus(w, util.GenRetObj(errno, err.Error()), status)
		logger.Error.Println("Get robot qa err: ", err.Error())
		return
	}

	util.WriteJSON(w, util.GenRetObj(errno, qainfo))
}

func handleCreateRobotQAV3(w http.ResponseWriter, r *http.Request) {
	panic("TODO")
}
func handleUpdateRobotQAV3(w http.ResponseWriter, r *http.Request) {
	panic("TODO")
}
func handleAddRobotQAAnswerV3(w http.ResponseWriter, r *http.Request) {
	errno := ApiError.SUCCESS
	var auditBuffer bytes.Buffer
	var err error
	var ret interface{}

	defer func() {
		result := 0
		if err != nil {
			util.WriteJSONWithStatus(w,
				util.GenRetObj(errno, err.Error()),
				ApiError.GetHttpStatus(errno))
		} else {
			util.WriteJSON(w,
				util.GenRetObj(errno, ret))
			result = 1
		}

		if auditBuffer.Len() > 0 && err != nil {
			auditBuffer.WriteString(fmt.Sprintf(": %s", err.Error()))
		}

		auditMsg := auditBuffer.String()
		if auditMsg != "" {
			addAudit(r, audit.AuditModuleRobotProfile, audit.AuditOperationAdd, auditMsg, result)
		}
	}()

	appid := requestheader.GetAppID(r)
	qid, err := util.GetMuxIntVar(r, "id")
	if err != nil {
		err = util.GenBadRequestError("ID")
		return
	}

	basicQuestion, err := GetBasicQusetionV3(qid)
	if err != nil {
		errno = ApiError.DB_ERROR
		return
	}
	if basicQuestion == nil {
		errno = ApiError.NOT_FOUND_ERROR
		err = util.GenNotFoundError(util.Msg["RobotProfileQuestion"])
		return
	}
	auditBuffer.WriteString(fmt.Sprintf(util.Msg["AddRobotProfileAnswerTemplate"], basicQuestion.Content))

	answer := strings.TrimSpace(r.FormValue("content"))
	if answer == "" {
		err = util.GenBadRequestError(util.Msg["Content"])
		return
	}
	auditBuffer.WriteString(fmt.Sprintf(" [%s]", answer))

	id, errno, err := AddRobotQAAnswerV3(appid, qid, answer)
	if err != nil {
		return
	}

	ret = InfoV3{
		ID:      id,
		Content: answer,
	}
	go SyncRobotProfileToSolr()
	return
}

func handleUpdateRobotQAAnswerV3(w http.ResponseWriter, r *http.Request) {
	errno := ApiError.SUCCESS
	var auditBuffer bytes.Buffer
	var err error
	var ret interface{}

	defer func() {
		result := 0
		if err != nil {
			util.WriteJSONWithStatus(w,
				util.GenRetObj(errno, err.Error()),
				ApiError.GetHttpStatus(errno))
		} else {
			util.WriteJSON(w,
				util.GenRetObj(errno, ret))
			result = 1
		}

		if auditBuffer.Len() > 0 && err != nil {
			auditBuffer.WriteString(fmt.Sprintf(": %s", err.Error()))
		}

		auditMsg := auditBuffer.String()
		if auditMsg != "" {
			addAudit(r, audit.AuditModuleRobotProfile, audit.AuditOperationEdit, auditMsg, result)
		}
	}()

	appid := requestheader.GetAppID(r)
	qid, err := util.GetMuxIntVar(r, "id")
	if err != nil {
		err = util.GenBadRequestError("ID")
		return
	}
	aid, err := util.GetMuxIntVar(r, "aid")
	if err != nil {
		err = util.GenBadRequestError("AID")
		return
	}

	basicQuestion, err := GetBasicQusetionV3(qid)
	if err != nil {
		errno = ApiError.DB_ERROR
		return
	}
	if basicQuestion == nil {
		errno = ApiError.NOT_FOUND_ERROR
		err = util.GenNotFoundError(util.Msg["RobotProfileQuestion"])
		return
	}
	auditBuffer.WriteString(fmt.Sprintf(util.Msg["EditRobotProfileAnswerTemplate"], basicQuestion.Content))

	origAnswerInfo, err := GetRobotQAAnswerV3(appid, qid, aid)
	if err != nil {
		errno = ApiError.DB_ERROR
		return
	}
	if origAnswerInfo == nil {
		errno = ApiError.NOT_FOUND_ERROR
		err = util.GenNotFoundError(util.Msg["RobotProfileAnswer"])
		return
	}

	answer := strings.TrimSpace(r.FormValue("content"))
	if answer == "" {
		err = util.GenBadRequestError(util.Msg["Content"])
		return
	}
	auditBuffer.WriteString(fmt.Sprintf(" [%s] -> [%s]", origAnswerInfo.Content, answer))

	if answer != origAnswerInfo.Content {
		errno, err = UpdateRobotQAAnswerV3(appid, qid, aid, answer)
		if err != nil {
			return
		}
		origAnswerInfo.Content = answer
	}

	ret = origAnswerInfo
	go SyncRobotProfileToSolr()
	return
}

func handleDeleteRobotQAAnswerV3(w http.ResponseWriter, r *http.Request) {
	errno := ApiError.SUCCESS
	var auditBuffer bytes.Buffer
	var err error
	var ret interface{}

	defer func() {
		result := 0
		if err != nil {
			util.WriteJSONWithStatus(w,
				util.GenRetObj(errno, err.Error()),
				ApiError.GetHttpStatus(errno))
		} else {
			util.WriteJSON(w,
				util.GenRetObj(errno, ret))
			result = 1
		}

		if auditBuffer.Len() > 0 && err != nil {
			auditBuffer.WriteString(fmt.Sprintf(": %s", err.Error()))
		}

		auditMsg := auditBuffer.String()
		if auditMsg != "" {
			addAudit(r, audit.AuditModuleRobotProfile, audit.AuditOperationDelete, auditMsg, result)
		}
	}()

	appid := requestheader.GetAppID(r)
	qid, err := util.GetMuxIntVar(r, "id")
	if err != nil {
		err = util.GenBadRequestError("ID")
		return
	}
	aid, err := util.GetMuxIntVar(r, "aid")
	if err != nil {
		err = util.GenBadRequestError("AID")
		return
	}

	basicQuestion, err := GetBasicQusetionV3(qid)
	if err != nil {
		errno = ApiError.DB_ERROR
		return
	}
	if basicQuestion == nil {
		errno = ApiError.NOT_FOUND_ERROR
		err = util.GenNotFoundError(util.Msg["RobotProfileQuestion"])
		return
	}
	auditBuffer.WriteString(fmt.Sprintf(util.Msg["DelRobotProfileAnswerTemplate"], basicQuestion.Content))

	origAnswerInfo, err := GetRobotQAAnswerV3(appid, qid, aid)
	if err != nil {
		errno = ApiError.DB_ERROR
		return
	}
	if origAnswerInfo == nil {
		auditBuffer.WriteString(fmt.Sprintf(": ID %d %s", aid, util.Msg["Deleted"]))
		return
	}
	auditBuffer.WriteString(fmt.Sprintf(" [%s]", origAnswerInfo.Content))

	errno, err = DeleteRobotQAAnswerV3(appid, qid, aid)
	if err != nil {
		return
	}
	go SyncRobotProfileToSolr()
	return
}
func handleAddRobotQARQuestionV3(w http.ResponseWriter, r *http.Request) {
	errno := ApiError.SUCCESS
	var auditBuffer bytes.Buffer
	var err error
	var ret interface{}

	defer func() {
		result := 0
		if err != nil {
			util.WriteJSONWithStatus(w,
				util.GenRetObj(errno, err.Error()),
				ApiError.GetHttpStatus(errno))
		} else {
			util.WriteJSON(w,
				util.GenRetObj(errno, ret))
			result = 1
		}

		if auditBuffer.Len() > 0 && err != nil {
			auditBuffer.WriteString(fmt.Sprintf(": %s", err.Error()))
		}

		auditMsg := auditBuffer.String()
		if auditMsg != "" {
			addAudit(r, audit.AuditModuleRobotProfile, audit.AuditOperationAdd, auditMsg, result)
		}
	}()

	appid := requestheader.GetAppID(r)
	qid, err := util.GetMuxIntVar(r, "id")
	if err != nil {
		err = util.GenBadRequestError("ID")
		return
	}

	basicQuestion, err := GetBasicQusetionV3(qid)
	if err != nil {
		errno = ApiError.DB_ERROR
		return
	}
	if basicQuestion == nil {
		errno = ApiError.NOT_FOUND_ERROR
		err = util.GenNotFoundError(util.Msg["RobotProfileQuestion"])
		return
	}
	auditBuffer.WriteString(fmt.Sprintf(util.Msg["AddRobotProfileRQuestionTemplate"], basicQuestion.Content))

	relateQuestion := strings.TrimSpace(r.FormValue("content"))
	if relateQuestion == "" {
		errno = ApiError.REQUEST_ERROR
		err = util.GenBadRequestError(util.Msg["Content"])
		return
	}
	auditBuffer.WriteString(fmt.Sprintf(" [%s]", relateQuestion))

	id, errno, err := AddRobotQARQuestionV3(appid, qid, relateQuestion)
	if err != nil {
		return
	}

	ret = InfoV3{
		ID:      id,
		Content: relateQuestion,
	}
	go SyncRobotProfileToSolr()
	return
}
func handleUpdateRobotQARQuestionV3(w http.ResponseWriter, r *http.Request) {
	errno := ApiError.SUCCESS
	var auditBuffer bytes.Buffer
	var err error
	var ret interface{}

	defer func() {
		result := 0
		if err != nil {
			util.WriteJSONWithStatus(w,
				util.GenRetObj(errno, err.Error()),
				ApiError.GetHttpStatus(errno))
		} else {
			util.WriteJSON(w,
				util.GenRetObj(errno, ret))
			result = 1
		}

		if auditBuffer.Len() > 0 && err != nil {
			auditBuffer.WriteString(fmt.Sprintf(": %s", err.Error()))
		}

		auditMsg := auditBuffer.String()
		if auditMsg != "" {
			addAudit(r, audit.AuditModuleRobotProfile, audit.AuditOperationEdit, auditMsg, result)
		}
	}()

	appid := requestheader.GetAppID(r)
	qid, err := util.GetMuxIntVar(r, "id")
	if err != nil {
		err = util.GenBadRequestError("ID")
		return
	}
	rQid, err := util.GetMuxIntVar(r, "qid")
	if err != nil {
		err = util.GenBadRequestError("QID")
		return
	}

	basicQuestion, err := GetBasicQusetionV3(qid)
	if err != nil {
		errno = ApiError.DB_ERROR
		return
	}
	if basicQuestion == nil {
		errno = ApiError.NOT_FOUND_ERROR
		err = util.GenNotFoundError(util.Msg["RobotProfileQuestion"])
		return
	}
	auditBuffer.WriteString(fmt.Sprintf(util.Msg["EditRobotProfileRQuestionTemplate"], basicQuestion.Content))

	origRQuestionInfo, err := GetRobotQARQuestionV3(appid, qid, rQid)
	if err != nil {
		errno = ApiError.DB_ERROR
		return
	}
	if origRQuestionInfo == nil {
		errno = ApiError.NOT_FOUND_ERROR
		err = util.GenNotFoundError(util.Msg["RobotProfileRelateQuestion"])
		return
	}

	relateQuestion := strings.TrimSpace(r.FormValue("content"))
	if relateQuestion == "" {
		err = util.GenBadRequestError(util.Msg["Content"])
		return
	}
	auditBuffer.WriteString(fmt.Sprintf(" [%s] -> [%s]", origRQuestionInfo.Content, relateQuestion))

	if relateQuestion != origRQuestionInfo.Content {
		errno, err = UpdateRobotQARQuestionV3(appid, qid, rQid, relateQuestion)
		if err != nil {
			return
		}
		origRQuestionInfo.Content = relateQuestion
	}

	ret = origRQuestionInfo
	go SyncRobotProfileToSolr()
	return
}
func handleDeleteRobotQARQuestionV3(w http.ResponseWriter, r *http.Request) {
	errno := ApiError.SUCCESS
	var auditBuffer bytes.Buffer
	var err error
	var ret interface{}

	defer func() {
		result := 0
		if err != nil {
			util.WriteJSONWithStatus(w,
				util.GenRetObj(errno, err.Error()),
				ApiError.GetHttpStatus(errno))
		} else {
			util.WriteJSON(w,
				util.GenRetObj(errno, ret))
			result = 1
		}

		if auditBuffer.Len() > 0 && err != nil {
			auditBuffer.WriteString(fmt.Sprintf(": %s", err.Error()))
		}

		auditMsg := auditBuffer.String()
		if auditMsg != "" {
			addAudit(r, audit.AuditModuleRobotProfile, audit.AuditOperationDelete, auditMsg, result)
		}
	}()

	appid := requestheader.GetAppID(r)
	qid, err := util.GetMuxIntVar(r, "id")
	if err != nil {
		err = util.GenBadRequestError("ID")
		return
	}
	rQid, err := util.GetMuxIntVar(r, "qid")
	if err != nil {
		err = util.GenBadRequestError("QID")
		return
	}

	basicQuestion, err := GetBasicQusetionV3(qid)
	if err != nil {
		errno = ApiError.DB_ERROR
		return
	}
	if basicQuestion == nil {
		errno = ApiError.NOT_FOUND_ERROR
		err = util.GenNotFoundError(util.Msg["RobotProfileQuestion"])
		return
	}
	auditBuffer.WriteString(fmt.Sprintf(util.Msg["DelRobotProfileRQuestionTemplate"], basicQuestion.Content))

	origRQuestionInfo, err := GetRobotQARQuestionV3(appid, qid, rQid)
	if err != nil {
		errno = ApiError.DB_ERROR
		return
	}
	if origRQuestionInfo == nil {
		auditBuffer.WriteString(fmt.Sprintf(": ID %d %s", rQid, util.Msg["Deleted"]))
		return
	}
	auditBuffer.WriteString(fmt.Sprintf(" [%s]", origRQuestionInfo.Content))

	errno, err = DeleteRobotQARQuestionV3(appid, qid, rQid)
	if err != nil {
		return
	}
	go SyncRobotProfileToSolr()
	return
}

func handleRebuildRobotQAV3(w http.ResponseWriter, r *http.Request) {
	// TODO: only force update data of robot itself
	err := ForceSyncRobotProfileToSolr(true)
	if err != nil {
		util.ReturnError(w, AdminErrors.ErrnoAPIError, err.Error())
	}
	util.Return(w, nil, "success")
	return
}
