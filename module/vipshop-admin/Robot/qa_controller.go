package Robot

import (
	"emotibot.com/emotigo/module/vipshop-admin/ApiError"
	"emotibot.com/emotigo/module/vipshop-admin/util"
	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
)

const (
	defaultListPerPage = 30
)

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
	errMsg := ApiError.GetErrorMsg(errCode)
	if err != nil {
		ctx.JSON(util.GenRetObj(errCode, errMsg, err.Error()))
		return
	}

	ctx.JSON(util.GenRetObj(errCode, errMsg, list))
}
