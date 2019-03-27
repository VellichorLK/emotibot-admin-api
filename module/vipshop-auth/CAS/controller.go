package CAS

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"emotibot.com/emotigo/module/vipshop-admin/ApiError"
	"emotibot.com/emotigo/module/vipshop-admin/util"
	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
	captcha "github.com/mojocn/base64Captcha"
)

var (
	// ModuleInfo is needed for module define
	ModuleInfo util.ModuleInfo
)

const validAppID = "vipshop"

const (
	StoreCollectNum = 1024
	StoreExpireDur  = 60 * time.Minute
)

func init() {
	ModuleInfo = util.ModuleInfo{
		ModuleName: "cas",
		EntryPoints: []util.EntryPoint{
			util.NewEntryPoint("POST", "login", []string{}, handleLogin),
			util.NewEntryPoint("GET", "captcha", []string{}, handleGetCatpcha),
		},
	}

	// init catpch store
	store := captcha.NewMemoryStore(StoreCollectNum, StoreExpireDur)
	captcha.SetCustomStore(store)
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

	// verify captcha
	captchaCode := ctx.FormValue("captcha")
	captchaID := ctx.FormValue("captcha_id")
	if captchaCode == "" || captchaID == "" {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.JSON(util.GenRetObj(ApiError.REQUEST_ERROR, "captcha is invalid"))
		return
	}

	verifyResult := captcha.VerifyCaptcha(captchaID, captchaCode)
	if !verifyResult {
		// verify failed
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.JSON(util.GenRetObj(ApiError.REQUEST_ERROR, "captcha is incorrect"))
		return
	}

	// getURL := fmt.Sprintf("%s?type=json&appid=%s&ac=%s&pw=%s", getCASServer(), getCASAppid(), userID, pwd)
	getURL := fmt.Sprintf("%s?type=json&appid=%s&ac=%s&pw=%s", getCASServer(), getCASAppid(), url.QueryEscape(userID), url.QueryEscape(pwd))

	_, resp, err := HTTPSGetRequest(getURL, 5, true)

	result := 0
	logMsg := userID + " 登录"
	userIP := util.GetUserIP(ctx)

	util.LogInfo.Printf("cas userId:[%s] getURL[%s] resp: [%s]", url.QueryEscape(userID), getURL, resp)

	defer func() {
		util.AddAuditLog(userID, userIP, util.AuditModuleMembers, util.AuditOperationLogin, logMsg, result)
	}()

	if err != nil {
		util.LogInfo.Printf("msg: [%s]", err.Error())
		casServerError(ctx, err)
		return
	}

	if resp == nil {
		err = errors.New("empty response")
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

func handleGetCatpcha(ctx context.Context) {
	config := captchaConfig()
	captchaId, captcaInterfaceInstance := captcha.GenerateCaptcha("", *config)

	base64blob := captcha.CaptchaWriteToBase64Encoding(captcaInterfaceInstance)

	var response CaptchaRet = CaptchaRet{
		Data: base64blob,
		ID:   captchaId,
	}

	ctx.JSON(response)
}

func captchaConfig() *captcha.ConfigCharacter {
	return &captcha.ConfigCharacter{
		Height: 60,
		Width:  265,
		//const CaptchaModeNumber:数字,CaptchaModeAlphabet:字母,CaptchaModeArithmetic:算术,CaptchaModeNumberAlphabet:数字字母混合.
		Mode:               captcha.CaptchaModeNumber,
		ComplexOfNoiseText: captcha.CaptchaComplexLower,
		ComplexOfNoiseDot:  captcha.CaptchaComplexLower,
		IsUseSimpleFont:    true,
		IsShowHollowLine:   false,
		IsShowNoiseDot:     false,
		IsShowNoiseText:    false,
		IsShowSlimeLine:    false,
		IsShowSineLine:     false,
		CaptchaLen:         5,
	}
}
