package Robot

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"emotibot.com/emotigo/module/admin-api/ApiError"
	"emotibot.com/emotigo/module/admin-api/util"
)

func handleChatInfoList(w http.ResponseWriter, r *http.Request) {
	appid := util.GetAppID(r)
	errCode := ApiError.SUCCESS

	chatList, errCode, err := GetRobotChatInfoList(appid)
	if errCode != ApiError.SUCCESS {
		util.WriteJSON(w, util.GenRetObj(errCode, err.Error()))
		return
	}
	util.WriteJSON(w, util.GenRetObj(errCode, chatList))
}

func handleGetChat(w http.ResponseWriter, r *http.Request) {
	appid := util.GetAppID(r)
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
	appid := util.GetAppID(r)
	errCode := ApiError.SUCCESS

	chatList, errCode, err := GetRobotChatList(appid)
	if errCode != ApiError.SUCCESS {
		util.WriteJSON(w, util.GenRetObj(errCode, err.Error()))
		return
	}
	util.WriteJSON(w, util.GenRetObj(errCode, chatList))
}

func handleMultiChatModify(w http.ResponseWriter, r *http.Request) {
	appid := util.GetAppID(r)
	errCode := ApiError.SUCCESS

	inputs := []*ChatInfoInput{}
	err := util.ReadJSON(r, &inputs)
	if err != nil {
		util.LogError.Println("Input invalid")
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
		util.LogError.Println("No valid input")
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
		util.LogInfo.Printf("Update consul result: %d, %s", ret, err.Error())
	} else {
		util.LogInfo.Printf("Update consul result: %d", ret)
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

	addAudit(r, util.AuditModuleBotMessage, util.AuditOperationEdit, strings.Join(msgs, "\n"), result)
}

func handleGetRobotWords(w http.ResponseWriter, r *http.Request) {
	appid := util.GetAppID(r)
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
	appid := util.GetAppID(r)
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
	appid := util.GetAppID(r)
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
	defer func() {
		util.WriteJSONWithStatus(w, util.GenRetObj(errno, ret), httpStatus)
	}()
	appid := util.GetAppID(r)
	id, err := util.GetMuxIntVar(r, "id")
	if err != nil {
		ret, errno, httpStatus = "Invalid ID", ApiError.REQUEST_ERROR, http.StatusBadRequest
		return
	}

	content := r.FormValue("content")
	if content == "" {
		ret, errno, httpStatus = "Invalid content", ApiError.REQUEST_ERROR, http.StatusBadRequest
		return
	}

	ret, errno, err = AddRobotWordContent(appid, id, content)
	if err != nil {
		httpStatus, ret = ApiError.GetHttpStatus(errno), err.Error()
	}
}
func handleUpdateRobotWordContent(w http.ResponseWriter, r *http.Request) {
	httpStatus := http.StatusOK
	var ret interface{}
	var errno int
	defer func() {
		util.WriteJSONWithStatus(w, util.GenRetObj(errno, ret), httpStatus)
	}()
	appid := util.GetAppID(r)
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

	content := r.FormValue("content")
	if content == "" {
		ret, errno, httpStatus = "Invalid content", ApiError.REQUEST_ERROR, http.StatusBadRequest
		return
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
}
func handleDeleteRobotWordContent(w http.ResponseWriter, r *http.Request) {
	httpStatus := http.StatusOK
	var ret interface{}
	var errno int
	defer func() {
		util.WriteJSONWithStatus(w, util.GenRetObj(errno, ret), httpStatus)
	}()
	appid := util.GetAppID(r)
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

	errno, err = DeleteRobotWordContent(appid, id, cid)
	if err != nil {
		httpStatus, ret = ApiError.GetHttpStatus(errno), err.Error()
	}
}
