package CAS

import (
	"encoding/json"
	"fmt"
	"strings"

	"emotibot.com/emotigo/module/vipshop-admin/ApiError"
	"emotibot.com/emotigo/module/vipshop-admin/util"
	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
)

var (
	// ModuleInfo is needed for module define
	ModuleInfo util.ModuleInfo
)

func init() {
	ModuleInfo = util.ModuleInfo{
		ModuleName: "cas",
		EntryPoints: []util.EntryPoint{
			util.NewEntryPoint("POST", "login", []string{}, handleLogin),
		},
	}
}

func handleLogin(ctx context.Context) {

	userID := ctx.FormValue("user_name")
	if strings.Trim(userID, " ") == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.JSON(util.GenRetObj(ApiError.REQUEST_ERROR, "loginname is invalid"))
		return
	}

	pwd := ctx.FormValue("raw_password")
	if strings.Trim(pwd, " ") == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.JSON(util.GenRetObj(ApiError.REQUEST_ERROR, "password is invalid"))
		return
	}

	getURL := fmt.Sprintf("%s?type=json&appid=%s&ac=%s&pw=%s", getCASServer(), getCASAppid(), userID, pwd)

	_, resp, err := HTTPSGetRequest(getURL, 5, true)

	result := 0
	logMsg := userID + " 登录"
	userIP := util.GetUserIP(ctx)

	defer func() {
		util.AddAuditLog(userID, userIP, util.AuditModuleMembers, util.AuditOperationLogin, logMsg, result)
	}()

	if err != nil {
		util.LogInfo.Printf("msg: [%s]", err.Error())
		casServerError(ctx, err)
		return
	}

	if resp == nil {
		util.LogInfo.Printf("msg: [%s]", err.Error())
		casServerError(ctx, err)
		return
	}

	casResp := &CASRetStruct{}
	err = json.Unmarshal(resp, &casResp)
	if err != nil {
		util.LogInfo.Printf("msg: [%s]", err.Error())
		casServerError(ctx, err)
		return
	}

	if casResp.Code != iris.StatusOK {
		ctx.StatusCode(casResp.Code)
		ctx.JSON(util.GenRetObj(ApiError.WEB_REQUEST_ERROR, "status is not 200"))
		return
	}

	lr, err := getUserPrivs(userID, pwd)
	if err != nil {
		ctx.StatusCode(iris.StatusForbidden)
		ctx.JSON(util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()))
		return
	}

	ctx.StatusCode(iris.StatusOK)
	ctx.JSON(util.GenRetObj(ApiError.SUCCESS, lr))
	result = 1

}

func casServerError(ctx context.Context, err error) {
	ctx.StatusCode(iris.StatusBadGateway)
	ctx.JSON(util.GenRetObj(ApiError.WEB_REQUEST_ERROR, err.Error()))
}
