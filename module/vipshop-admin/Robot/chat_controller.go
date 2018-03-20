package Robot

import (
	"encoding/json"
	"fmt"
	"strings"

	"emotibot.com/emotigo/module/vipshop-admin/ApiError"
	"emotibot.com/emotigo/module/vipshop-admin/util"
	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
)

func handleChatInfoList(ctx context.Context) {
	appid := util.GetAppID(ctx)
	errCode := ApiError.SUCCESS

	chatList, errCode, err := GetRobotChatInfoList(appid)
	if errCode != ApiError.SUCCESS {
		ctx.JSON(util.GenRetObj(errCode, err.Error()))
		return
	}
	ctx.JSON(util.GenRetObj(errCode, chatList))
}

func handleChatList(ctx context.Context) {
	appid := util.GetAppID(ctx)
	errCode := ApiError.SUCCESS

	chatList, errCode, err := GetRobotChatList(appid)
	if errCode != ApiError.SUCCESS {
		ctx.JSON(util.GenRetObj(errCode, err.Error()))
		return
	}
	ctx.JSON(util.GenRetObj(errCode, chatList))
}

func handleMultiChatModify(ctx context.Context) {
	appid := util.GetAppID(ctx)
	errCode := ApiError.SUCCESS

	inputs := []*ChatInfoInput{}
	err := ctx.ReadJSON(&inputs)
	if err != nil {
		util.LogError.Println("Input invalid")
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.JSON(util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()))
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
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.JSON(util.GenRetObj(ApiError.REQUEST_ERROR, "Empty input"))
		return
	}

	var types []int
	for _, info := range validInput {
		types = append(types, info.Type)
	}

	origInfos, errCode, err := GetMultiRobotChat(appid, types)
	if err != nil {
		ctx.JSON(util.GenRetObj(errCode, err.Error()))
		return
	}

	errCode, err = UpdateMultiChat(appid, validInput)
	if err != nil {
		ctx.JSON(util.GenRetObj(errCode, err.Error()))
		auditMultiChatModify(ctx, origInfos, validInput, 0)
	} else {
		ctx.JSON(util.GenSimpleRetObj(errCode))
		auditMultiChatModify(ctx, origInfos, validInput, 1)
	}
	ret, err := util.ConsulUpdateRobotChat(appid)
	if err != nil {
		util.LogInfo.Printf("Update consul result: %d, %s", ret, err.Error())
	} else {
		util.LogInfo.Printf("Update consul result: %d", ret)
	}
}

func auditMultiChatModify(ctx context.Context, origInfos []*ChatInfo, newInfos []*ChatInfoInput, result int) {
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

	addAudit(ctx, util.AuditModuleBotMessage, util.AuditOperationEdit, strings.Join(msgs, "\n"), result)
}
