package Robot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"emotibot.com/emotigo/module/admin-api/ApiError"
	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/admin-api/util/audit"
	"emotibot.com/emotigo/module/admin-api/util/requestheader"
	"emotibot.com/emotigo/pkg/logger"
)

func handleChatInfoList(w http.ResponseWriter, r *http.Request) {
	appid := requestheader.GetAppID(r)
	errCode := ApiError.SUCCESS

	chatList, errCode, err := GetRobotChatInfoList(appid)
	if errCode != ApiError.SUCCESS {
		util.WriteJSON(w, util.GenRetObj(errCode, err.Error()))
		return
	}
	util.WriteJSON(w, util.GenRetObj(errCode, chatList))
}

func handleGetChat(w http.ResponseWriter, r *http.Request) {
	appid := requestheader.GetAppID(r)
	errCode := ApiError.SUCCESS
	id, err := util.GetMuxIntVar(r, "id")
	if err != nil || id <= 0 {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	chat, errCode, err := GetRobotChat(appid, id)
	if errCode != ApiError.SUCCESS {
		util.WriteJSON(w, util.GenRetObj(errCode, err.Error()))
		return
	}
	util.WriteJSON(w, util.GenRetObj(errCode, chat))
}

func handleChatList(w http.ResponseWriter, r *http.Request) {
	appid := requestheader.GetAppID(r)
	errCode := ApiError.SUCCESS

	chatList, errCode, err := GetRobotChatList(appid)
	if errCode != ApiError.SUCCESS {
		util.WriteJSON(w, util.GenRetObj(errCode, err.Error()))
		return
	}
	util.WriteJSON(w, util.GenRetObj(errCode, chatList))
}

func handleMultiChatModify(w http.ResponseWriter, r *http.Request) {
	appid := requestheader.GetAppID(r)
	errCode := ApiError.SUCCESS

	inputs := []*ChatInfoInput{}
	err := util.ReadJSON(r, &inputs)
	if err != nil {
		logger.Error.Println("Input invalid")
		http.Error(w, "", http.StatusBadRequest)
		util.WriteJSON(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()))
		return
	}

	validInput := []*ChatInfoInput{}
	for _, input := range inputs {
		if input.Type > 0 {
			validInput = append(validInput, input)
		}
	}

	if len(validInput) <= 0 {
		logger.Error.Println("No valid input")
		http.Error(w, "", http.StatusBadRequest)
		util.WriteJSON(w, util.GenRetObj(ApiError.REQUEST_ERROR, "Empty input"))
		return
	}

	var types []int
	for _, info := range validInput {
		types = append(types, info.Type)
	}

	origInfos, errCode, err := GetMultiRobotChat(appid, types)
	if err != nil {
		util.WriteJSON(w, util.GenRetObj(errCode, err.Error()))
		return
	}

	errCode, err = UpdateMultiChat(appid, validInput)
	if err != nil {
		util.WriteJSON(w, util.GenRetObj(errCode, err.Error()))
		auditMultiChatModify(r, origInfos, validInput, 0)
	} else {
		util.WriteJSON(w, util.GenSimpleRetObj(errCode))
		auditMultiChatModify(r, origInfos, validInput, 1)
	}
	ret, err := util.ConsulUpdateRobotChat(appid)
	if err != nil {
		logger.Info.Printf("Update consul result: %d, %s", ret, err.Error())
	} else {
		logger.Info.Printf("Update consul result: %d", ret)
	}
}

func auditMultiChatModify(r *http.Request, origInfos []*ChatInfo, newInfos []*ChatInfoInput, result int) {
	origInfoMap := make(map[int]*ChatInfo)
	for _, v := range origInfos {
		origInfoMap[v.Type] = v
	}

	msgs := []string{}
	for _, newInfo := range newInfos {
		newStr, _ := json.Marshal(newInfo)
		if origInfo, ok := origInfoMap[newInfo.Type]; ok {
			origStr, _ := json.Marshal(origInfo)
			msgs = append(msgs, fmt.Sprintf("%s: [%s] => [%s]", newInfo.Name, origStr, newStr))
		} else {
			msgs = append(msgs, fmt.Sprintf("%s: [] => [%s]", newInfo.Name, newStr))
		}
	}

	addAudit(r, audit.AuditModuleRobotChatSkill, audit.AuditOperationEdit, strings.Join(msgs, "\n"), result)
}

func handleGetRobotWords(w http.ResponseWriter, r *http.Request) {
	appid := requestheader.GetAppID(r)
	httpStatus := http.StatusOK
	var ret interface{}
	var errno int
	defer func() {
		util.WriteJSONWithStatus(w, util.GenRetObj(errno, ret), httpStatus)
	}()

	ret, errno, err := GetRobotWords(appid)
	if err != nil {
		httpStatus, ret = ApiError.GetHttpStatus(errno), err.Error()
	}
}
func handleGetRobotWord(w http.ResponseWriter, r *http.Request) {
	httpStatus := http.StatusOK
	var ret interface{}
	var errno int
	defer func() {
		util.WriteJSONWithStatus(w, util.GenRetObj(errno, ret), httpStatus)
	}()
	appid := requestheader.GetAppID(r)
	id, err := util.GetMuxIntVar(r, "id")
	if err != nil {
		ret, errno, httpStatus = "Invalid ID", ApiError.REQUEST_ERROR, http.StatusBadRequest
		return
	}

	ret, errno, err = GetRobotWord(appid, id)
	if err != nil {
		httpStatus, ret = ApiError.GetHttpStatus(errno), err.Error()
	}
}
func handleUpdateRobotWord(w http.ResponseWriter, r *http.Request) {
	httpStatus := http.StatusOK
	var ret interface{}
	var errno int
	defer func() {
		util.WriteJSONWithStatus(w, util.GenRetObj(errno, ret), httpStatus)
	}()
	appid := requestheader.GetAppID(r)
	id, err := util.GetMuxIntVar(r, "id")
	if err != nil {
		ret, errno, httpStatus = "Invalid ID", ApiError.REQUEST_ERROR, http.StatusBadRequest
		return
	}

	contentStr := r.FormValue("content")
	contents := []string{}
	err = json.Unmarshal([]byte(contentStr), &contents)
	if err != nil {
		ret, errno, httpStatus = "Invalid contents", ApiError.REQUEST_ERROR, http.StatusBadRequest
		return
	}

	ret, errno, err = UpdateRobotWord(appid, id, contents)
	if err != nil {
		httpStatus, ret = ApiError.GetHttpStatus(errno), err.Error()
	}
}
func handleAddRobotWordContent(w http.ResponseWriter, r *http.Request) {
	httpStatus := http.StatusOK
	var ret interface{}
	var errno int
	var content string
	var wordsType *ChatInfoV2
	defer func() {
		util.WriteJSONWithStatus(w, util.GenRetObj(errno, ret), httpStatus)
		auditBuffer := bytes.Buffer{}
		auditBuffer.WriteString(util.Msg["Add"])
		if wordsType != nil {
			auditBuffer.WriteString(wordsType.Name)
		} else {
			auditBuffer.WriteString(util.Msg["RobotWords"])
		}
		if content != "" {
			auditBuffer.WriteString(fmt.Sprintf(" %s", content))
		}
		retVal := 0
		if errno == ApiError.SUCCESS {
			retVal = 1
		}
		addAudit(r, audit.AuditModuleRobotChatSkill, audit.AuditOperationAdd, auditBuffer.String(), retVal)
	}()
	appid := requestheader.GetAppID(r)
	id, err := util.GetMuxIntVar(r, "id")
	if err != nil {
		ret, errno, httpStatus = "Invalid type ID", ApiError.REQUEST_ERROR, http.StatusBadRequest
		return
	}

	wordsType, errno, err = GetRobotWord(appid, id)
	if err != nil {
		ret, httpStatus = "Invalid type ID", ApiError.GetHttpStatus(errno)
	}

	content = r.FormValue("content")
	if content == "" {
		ret, errno, httpStatus = "Invalid content", ApiError.REQUEST_ERROR, http.StatusBadRequest
		return
	}

	ret, errno, err = AddRobotWordContent(appid, id, content)
	if err != nil {
		httpStatus, ret = ApiError.GetHttpStatus(errno), err.Error()
	}
	updateWordsConsul(appid)
}
func handleUpdateRobotWordContent(w http.ResponseWriter, r *http.Request) {
	httpStatus := http.StatusOK
	var ret interface{}
	var errno int

	var origContent string
	var content string
	var wordsType *ChatInfoV2
	defer func() {
		util.WriteJSONWithStatus(w, util.GenRetObj(errno, ret), httpStatus)
		auditBuffer := bytes.Buffer{}
		auditBuffer.WriteString(util.Msg["Modify"])
		if wordsType != nil {
			auditBuffer.WriteString(wordsType.Name)
		} else {
			auditBuffer.WriteString(util.Msg["RobotWords"])
		}
		if origContent != "" && content != "" {
			auditBuffer.WriteString(fmt.Sprintf(": %s => %s", origContent, content))
		}
		retVal := 0
		if errno == ApiError.SUCCESS {
			retVal = 1
		}
		addAudit(r, audit.AuditModuleRobotChatSkill, audit.AuditOperationEdit, auditBuffer.String(), retVal)
	}()
	appid := requestheader.GetAppID(r)
	id, err := util.GetMuxIntVar(r, "id")
	if err != nil {
		ret, errno, httpStatus = "Invalid ID", ApiError.REQUEST_ERROR, http.StatusBadRequest
		return
	}
	cid, err := util.GetMuxIntVar(r, "cid")
	if err != nil {
		ret, errno, httpStatus = "Invalid Content ID", ApiError.REQUEST_ERROR, http.StatusBadRequest
		return
	}

	content = r.FormValue("content")
	if content == "" {
		ret, errno, httpStatus = "Invalid content", ApiError.REQUEST_ERROR, http.StatusBadRequest
		return
	}

	wordsType, errno, err = GetRobotWord(appid, id)
	if err != nil {
		ret, httpStatus = "Invalid type ID", ApiError.GetHttpStatus(errno)
	}
	for _, c := range wordsType.Contents {
		if c.ID == cid {
			origContent = c.Content
			break
		}
	}

	errno, err = UpdateRobotWordContent(appid, id, cid, content)
	if err != nil {
		httpStatus, ret = ApiError.GetHttpStatus(errno), err.Error()
	} else {
		ret = ChatContentInfoV2{
			ID:      cid,
			Content: content,
		}
	}
	updateWordsConsul(appid)
}
func handleDeleteRobotWordContent(w http.ResponseWriter, r *http.Request) {
	httpStatus := http.StatusOK
	var ret interface{}
	var errno int

	var origContent string
	var wordsType *ChatInfoV2
	defer func() {
		util.WriteJSONWithStatus(w, util.GenRetObj(errno, ret), httpStatus)
		auditBuffer := bytes.Buffer{}
		auditBuffer.WriteString(util.Msg["Delete"])
		if wordsType != nil {
			auditBuffer.WriteString(wordsType.Name)
		} else {
			auditBuffer.WriteString(util.Msg["RobotWords"])
		}
		if origContent != "" {
			auditBuffer.WriteString(fmt.Sprintf(" %s", origContent))
		}
		retVal := 0
		if errno == ApiError.SUCCESS {
			retVal = 1
		}
		addAudit(r, audit.AuditModuleRobotChatSkill, audit.AuditOperationDelete, auditBuffer.String(), retVal)
	}()
	appid := requestheader.GetAppID(r)
	id, err := util.GetMuxIntVar(r, "id")
	if err != nil {
		ret, errno, httpStatus = "Invalid ID", ApiError.REQUEST_ERROR, http.StatusBadRequest
		return
	}
	cid, err := util.GetMuxIntVar(r, "cid")
	if err != nil {
		ret, errno, httpStatus = "Invalid Content ID", ApiError.REQUEST_ERROR, http.StatusBadRequest
		return
	}

	wordsType, errno, err = GetRobotWord(appid, id)
	if err != nil {
		ret, httpStatus = "Invalid type ID", ApiError.GetHttpStatus(errno)
	}
	for _, c := range wordsType.Contents {
		if c.ID == cid {
			origContent = c.Content
			break
		}
	}

	errno, err = DeleteRobotWordContent(appid, id, cid)
	if err != nil {
		httpStatus, ret = ApiError.GetHttpStatus(errno), err.Error()
	}
	updateWordsConsul(appid)
}
func updateWordsConsul(appid string) {
	consulRet, consulErr := util.ConsulUpdateRobotChat(appid)
	if consulErr != nil {
		logger.Info.Printf("Update consul result: %d, %s", consulRet, consulErr.Error())
	} else {
		logger.Info.Printf("Update consul result: %d", consulRet)
	}
}
