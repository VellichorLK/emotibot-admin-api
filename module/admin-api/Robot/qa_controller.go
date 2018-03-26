package Robot

import (
	"bytes"
	"fmt"
	"strings"

	"emotibot.com/emotigo/module/admin-api/ApiError"
	"emotibot.com/emotigo/module/admin-api/util"
	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
)

const (
	defaultListPerPage = 30
)

func handleRobotQA(ctx context.Context) {
	appid := util.GetAppID(ctx)

	id, err := ctx.Params().GetInt("id")
	if err != nil || id <= 0 {
		ctx.StatusCode(iris.StatusBadRequest)
		return
	}

	info, errCode, err := GetRobotQA(appid, id)
	if errCode != ApiError.SUCCESS {
		ctx.JSON(util.GenRetObj(errCode, err.Error()))
	} else {
		ctx.JSON(util.GenRetObj(errCode, info))
	}
}

func handleRobotQAList(ctx context.Context) {
	appid := util.GetAppID(ctx)

	page, err := ctx.URLParamInt("page")
	if err != nil {
		util.LogInfo.Printf("Param error: %s", err.Error())
		ctx.StatusCode(iris.StatusBadRequest)
		return
	}
	listPerPage, err := ctx.URLParamInt("per_page")
	if err != nil {
		util.LogInfo.Printf("Param error: %s", err.Error())
		ctx.StatusCode(iris.StatusBadRequest)
		return
	}

	if page < 0 || listPerPage < 0 {
		ctx.StatusCode(iris.StatusBadRequest)
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
		ctx.JSON(util.GenRetObj(errCode, err.Error()))
		return
	}

	ctx.JSON(util.GenRetObj(errCode, list))
}

func handleRobotQAModelRebuild(ctx context.Context) {
	appid := util.GetAppID(ctx)
	auditLog := ""
	result := 0

	errCode, err := util.McRebuildRobotQA(appid)
	if errCode != ApiError.SUCCESS {
		ctx.JSON(util.GenRetObj(errCode, err.Error()))
		auditLog = fmt.Sprintf("%s%s%s",
			util.Msg["RobotProfile"], util.Msg["Rebuild"], util.Msg["Error"])
	} else {
		ctx.JSON(util.GenSimpleRetObj(errCode))
		auditLog = fmt.Sprintf("%s%s%s",
			util.Msg["RobotProfile"], util.Msg["Rebuild"], util.Msg["Success"])
		result = 1
	}
	addAudit(ctx, util.AuditModuleRobotProfile, util.AuditOperationEdit, auditLog, result)
}

func handleUpdateRobotQA(ctx context.Context) {
	appid := util.GetAppID(ctx)
	auditLog := ""
	result := 0
	errCode := ApiError.SUCCESS
	errMsg := ""
	var retObj interface{}

	failMsg := fmt.Sprintf("%s%s%s",
		util.Msg["Modify"], util.Msg["RobotProfile"], util.Msg["Error"])
	successMsg := fmt.Sprintf("%s%s%s",
		util.Msg["Modify"], util.Msg["RobotProfile"], util.Msg["Success"])

	id, err := ctx.Params().GetInt("id")
	info := loadQAInfoFromContext(ctx)
	if err != nil || id <= 0 || info == nil {
		ctx.StatusCode(iris.StatusBadRequest)
		auditLog = fmt.Sprintf("%s: %s%s", failMsg, util.Msg["Request"], util.Msg["Error"])
		errCode = ApiError.REQUEST_ERROR
		addAudit(ctx, util.AuditModuleRobotProfile, util.AuditOperationEdit, auditLog, result)
		ctx.JSON(util.GenRetObj(errCode, retObj))
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
	addAudit(ctx, util.AuditModuleRobotProfile, util.AuditOperationEdit, auditLog, result)
	ctx.JSON(util.GenRetObj(errCode, retObj))
}

func loadQAInfoFromContext(ctx context.Context) *QAInfo {
	input := &QAInfo{}
	err := ctx.ReadJSON(input)
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
