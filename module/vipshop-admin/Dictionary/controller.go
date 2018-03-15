package Dictionary

import (
	"fmt"
	"net/http"

	"emotibot.com/emotigo/module/vipshop-admin/ApiError"
	"emotibot.com/emotigo/module/vipshop-admin/util"
	"github.com/kataras/iris/context"
)

var (
	// ModuleInfo is needed for module define
	ModuleInfo util.ModuleInfo
)

func init() {
	ModuleInfo = util.ModuleInfo{
		ModuleName: "dictionary",
		EntryPoints: []util.EntryPoint{
			util.NewEntryPoint("POST", "upload", []string{"view"}, handleUpload),
			util.NewEntryPoint("GET", "download", []string{"export"}, handleDownload),
			util.NewEntryPoint("GET", "download-meta", []string{"view"}, handleDownloadMeta),
			util.NewEntryPoint("GET", "check", []string{"view"}, handleFileCheck),
			util.NewEntryPoint("GET", "full-check", []string{"view"}, handleFullFileCheck),
			util.NewEntryPoint("GET", "wordbanks", []string{"view"}, handleGetWordbanks),
		},
	}
}

func getEnvironments() map[string]string {
	return util.GetEnvOf(ModuleInfo.ModuleName)
}

func getEnvironment(key string) string {
	envs := util.GetEnvOf(ModuleInfo.ModuleName)
	if envs != nil {
		if val, ok := envs[key]; ok {
			return val
		}
	}
	return ""
}

func getGlobalEnv(key string) string {
	envs := util.GetEnvOf("server")
	if envs != nil {
		if val, ok := envs[key]; ok {
			return val
		}
	}
	return ""
}

func handleGetWordbanks(ctx context.Context) {
	appid := util.GetAppID(ctx)

	wordbanks, err := GetEntities(appid)
	if err != nil {
		util.LogInfo.Printf("Error when get entities: %s\n", err.Error())
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.Writef(err.Error())
	}
	ctx.JSON(wordbanks)
}

// handleFileCheck will call api to check if uploaded dictionary is finished
func handleFileCheck(ctx context.Context) {
	appid := util.GetAppID(ctx)

	util.LogTrace.Printf("Check dictionary info from [%s]", appid)

	ret, err := CheckProcessStatus(appid)
	if err != nil {
		ctx.JSON(util.GenRetObj(ApiError.DB_ERROR, err.Error()))
	} else {
		ctx.JSON(util.GenRetObj(ApiError.SUCCESS, ret))
	}
}

// handleFileCheck will call api to check if uploaded dictionary is finished
func handleFullFileCheck(ctx context.Context) {
	appid := util.GetAppID(ctx)

	util.LogTrace.Printf("Check dictionary full info from [%s]", appid)

	ret, err := CheckFullProcessStatus(appid)
	if err != nil {
		ctx.JSON(util.GenRetObj(ApiError.DB_ERROR, err.Error()))
	} else {
		ctx.JSON(util.GenRetObj(ApiError.SUCCESS, ret))
	}
}

func handleUpload(ctx context.Context) {
	appid := util.GetAppID(ctx)
	userID := util.GetUserID(ctx)
	userIP := util.GetUserIP(ctx)

	file, info, err := ctx.FormFile("file")
	defer file.Close()
	util.LogInfo.Printf("Receive uploaded file: %s", info.Filename)
	util.LogTrace.Printf("Uploaded file info %#v", info.Header)

	// 1. receive upload file and check file
	retFile, errCode, err := CheckUploadFile(appid, file, info)
	if err != nil {
		errMsg := ApiError.GetErrorMsg(errCode)
		ctx.JSON(util.GenRetObj(errCode, err.Error()))
		util.AddAuditLog(userID, userIP, util.AuditModuleDictionary, util.AuditOperationImport, fmt.Sprintf("%s: %s", errMsg, err.Error()), 0)
		return
	} else if errCode != ApiError.SUCCESS {
		errMsg := ApiError.GetErrorMsg(errCode)
		ctx.JSON(util.GenSimpleRetObj(errCode))
		util.AddAuditLog(userID, userIP, util.AuditModuleDictionary, util.AuditOperationImport, fmt.Sprintf("%s: %s", errMsg, err.Error()), 0)
		return
	}

	// 2. http request to multicustomer
	errCode, err = util.UpdateWordBank(appid, userID, userIP, retFile)
	if err != nil {
		ctx.JSON(util.GenRetObj(errCode, err.Error()))
		util.AddAuditLog(userID, userIP, util.AuditModuleDictionary, util.AuditOperationImport, fmt.Sprintf("%s %s", util.Msg["Server"], util.Msg["Error"]), 0)
		util.LogError.Printf("update wordbank with multicustomer error: %s", err.Error())
	} else {
		errCode = ApiError.SUCCESS
		ctx.JSON(util.GenSimpleRetObj(errCode))
		util.AddAuditLog(userID, userIP, util.AuditModuleDictionary, util.AuditOperationImport, fmt.Sprintf("%s %s", util.Msg["UploadFile"], info.Filename), 1)
	}
}

func handleDownload(ctx context.Context) {
	d := map[string]interface{}{
		"result": true,
		"entry":  "download",
	}

	// 1. get file from input version
	// 2. output raw

	ctx.JSON(d)
}

func handleDownloadMeta(ctx context.Context) {
	// 1. select from db last two row
	// 2. return response
	appid := util.GetAppID(ctx)
	ret, err := GetDownloadMeta(appid)
	if err != nil {
		ctx.JSON(util.GenRetObj(ApiError.DB_ERROR, err.Error()))
	} else {
		ctx.JSON(util.GenRetObj(ApiError.SUCCESS, ret))
	}
}
