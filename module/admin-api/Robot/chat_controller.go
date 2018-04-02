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
	id, err := util.GetParamInt(r, "id")
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
