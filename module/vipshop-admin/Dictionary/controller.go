package Dictionary

import (
	"fmt"

	"emotibot.com/emotigo/module/vipshop-admin/ApiError"
	"emotibot.com/emotigo/module/vipshop-admin/util"
	"github.com/kataras/iris/context"
)

var (
	// ModuleInfo is needed for module define
	ModuleInfo util.ModuleInfo
)

const (
	MulticustomerURLKey = "MC_URL"
)

func init() {
	ModuleInfo = util.ModuleInfo{
		ModuleName: "dictionary",
		EntryPoints: []util.EntryPoint{
			util.NewEntryPoint("POST", "upload", handleUpload),
			util.NewEntryPoint("GET", "download", handleDownload),
			util.NewEntryPoint("GET", "download-meta", handleDownloadMeta),
			util.NewEntryPoint("GET", "check", handleFileCheck),
		},
	}
}

// InitDatabase will init dao in module, which should called after read env
func InitDatabase() {
	url := getGlobalEnv("MYSQL_URL")
	user := getGlobalEnv("MYSQL_USER")
	pass := getGlobalEnv("MYSQL_PASS")
	db := getGlobalEnv("MYSQL_DB")
	DaoInit(url, user, pass, db)
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

// handleFileCheck will call api to check if uploaded dictionary is finished
func handleFileCheck(ctx context.Context) {
	appid := util.GetAppID(ctx)

	util.LogTrace.Printf("Check dictionary info from [%s]", appid)

	ret, err := CheckProcessStatus(appid)
	if err != nil {
		errMsg := ApiError.GetErrorMsg(ApiError.DB_ERROR)
		ctx.JSON(util.GenRetObj(ApiError.DB_ERROR, errMsg, err.Error()))
	} else {
		errMsg := ApiError.GetErrorMsg(ApiError.SUCCESS)
		ctx.JSON(util.GenRetObj(ApiError.SUCCESS, errMsg, ret))
	}
}

func handleUpload(ctx context.Context) {
	appid := util.GetAppID(ctx)
	userID := util.GetUserID(ctx)
	userIP := util.GetUserIP(ctx)
	mcURL := getGlobalEnv(MulticustomerURLKey)

	file, info, err := ctx.FormFile("file")
	defer file.Close()
	util.LogInfo.Printf("Receive uploaded file: %s", info.Filename)
	util.LogTrace.Printf("Uploaded file info %#v", info.Header)

	// 1. receive upload file and check file
	retFile, errCode, err := CheckUploadFile(appid, file, info)
	if err != nil {
		errMsg := ApiError.GetErrorMsg(errCode)
		ctx.JSON(util.GenRetObj(errCode, errMsg, err.Error()))
		util.AddAuditLog(userID, userIP, util.AuditModuleDictionary, util.AuditOperationImport, fmt.Sprintf("%s: %s", errMsg, err.Error()), 0)
		return
	} else if errCode != ApiError.SUCCESS {
		errMsg := ApiError.GetErrorMsg(errCode)
		ctx.JSON(util.GenSimpleRetObj(errCode, errMsg))
		util.AddAuditLog(userID, userIP, util.AuditModuleDictionary, util.AuditOperationImport, fmt.Sprintf("%s: %s", errMsg, err.Error()), 0)
		return
	}

	// 2. http request to multicustomer
	// http://172.16.101.47:14501/entity?
	// app_id, userip, userid, file_name
	reqURL := fmt.Sprintf("%s/entity?app_id=%s&userid=%s&userip=%s&file_name=%s", mcURL, appid, userID, userIP, retFile)
	util.LogTrace.Printf("mc req: %s", reqURL)

	_, resErr := util.HTTPGetSimpleWithTimeout(reqURL, 1)
	if resErr != nil {
		errCode = ApiError.DICT_SERVICE_ERROR
		errMsg := ApiError.GetErrorMsg(errCode)
		ctx.JSON(util.GenRetObj(errCode, errMsg, resErr.Error()))
		util.AddAuditLog(userID, userIP, util.AuditModuleDictionary, util.AuditOperationImport, fmt.Sprintf("service error: %s", resErr.Error()), 0)
	} else {
		errCode = ApiError.SUCCESS
		errMsg := ApiError.GetErrorMsg(errCode)
		ctx.JSON(util.GenSimpleRetObj(errCode, errMsg))
		util.AddAuditLog(userID, userIP, util.AuditModuleDictionary, util.AuditOperationImport, fmt.Sprintf("upload %s", retFile), 1)
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
		errMsg := ApiError.GetErrorMsg(ApiError.DB_ERROR)
		ctx.JSON(util.GenRetObj(ApiError.DB_ERROR, errMsg, err.Error()))
	} else {
		errMsg := ApiError.GetErrorMsg(ApiError.SUCCESS)
		ctx.JSON(util.GenRetObj(ApiError.SUCCESS, errMsg, ret))
	}
}
