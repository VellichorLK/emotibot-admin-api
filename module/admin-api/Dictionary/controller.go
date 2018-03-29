package Dictionary

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"emotibot.com/emotigo/module/admin-api/ApiError"
	"emotibot.com/emotigo/module/admin-api/util"
	"github.com/kataras/iris/context"
)

var (
	// ModuleInfo is needed for module define
	ModuleInfo  util.ModuleInfo
	maxDirDepth int
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

			util.NewEntryPoint("PUT", "wordbank", []string{"edit"}, handlePutWordbank),
			util.NewEntryPoint("POST", "wordbank", []string{"edit"}, handleUpdateWordbank),
			util.NewEntryPoint("GET", "wordbank/{id:int}", []string{"view"}, handleGetWordbank),
		},
	}
	maxDirDepth = 4
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

func handleGetWordbank(ctx context.Context) {
	appid := util.GetAppID(ctx)
	id, err := ctx.Params().GetInt("id")
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.Writef("Error: %s", err.Error())
	}

	wordbank, err := GetWordbank(appid, id)
	if err != nil {
		util.LogInfo.Printf("Error when get wordbank: %s\n", err.Error())
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.Writef(err.Error())
	}

	if wordbank == nil {
		ctx.StatusCode(http.StatusNotFound)
		return
	}
	ctx.JSON(wordbank)
}

func handleUpdateWordbank(ctx context.Context) {
	appid := util.GetAppID(ctx)
	userID := util.GetUserID(ctx)
	userIP := util.GetUserIP(ctx)

	updatedWordbank := &WordBank{}
	err := ctx.ReadJSON(updatedWordbank)
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.Writef(err.Error())
		return
	}

	origWordbank, err := GetWordbank(appid, *updatedWordbank.ID)
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.Writef(err.Error())
		return
	}
	retCode, err := UpdateWordbank(appid, updatedWordbank)
	auditRet := 1
	if err != nil {
		if retCode == ApiError.REQUEST_ERROR {
			ctx.StatusCode(http.StatusBadRequest)
		} else {
			ctx.StatusCode(http.StatusInternalServerError)
		}
		ctx.Writef(err.Error())
		auditRet = 0
	} else {
		ctx.StatusCode(http.StatusOK)
	}

	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("%s%s: %s", util.Msg["Modify"], util.Msg["Wordbank"], origWordbank.Name))
	if origWordbank.SimilarWords != updatedWordbank.SimilarWords {
		buffer.WriteString(fmt.Sprintf("\n%s: '%s' => '%s'", util.Msg["SimilarWord"], origWordbank.SimilarWords, updatedWordbank.SimilarWords))
	}
	if origWordbank.Answer != updatedWordbank.Answer {
		buffer.WriteString(fmt.Sprintf("\n%s: '%s' => '%s'", util.Msg["Answer"], origWordbank.Answer, updatedWordbank.Answer))
	}

	auditMessage := buffer.String()
	util.AddAuditLog(userID, userIP, util.AuditModuleDictionary, util.AuditOperationEdit, auditMessage, auditRet)
}

func handlePutWordbank(ctx context.Context) {
	appid := util.GetAppID(ctx)
	userID := util.GetUserID(ctx)
	userIP := util.GetUserIP(ctx)

	paths, newWordBank, err := getWordbankFromReq(ctx)
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.Writef(err.Error())
		return
	}

	retCode, err := AddWordbank(appid, paths, newWordBank)
	auditMessage := ""
	logPath := []string{}
	for idx := range paths {
		if paths[idx] == "" {
			break
		}
		logPath = append(logPath, paths[idx])
	}
	if newWordBank == nil {
		auditMessage = fmt.Sprintf("%s: %s/",
			util.Msg["Add"],
			strings.Join(logPath, "/"))
	} else {
		auditMessage = fmt.Sprintf("%s: %s/%s",
			util.Msg["Add"],
			strings.Join(logPath, "/"), newWordBank.Name)
	}
	auditRet := 1
	if err != nil {
		if retCode == ApiError.REQUEST_ERROR {
			ctx.StatusCode(http.StatusBadRequest)
		} else {
			ctx.StatusCode(http.StatusInternalServerError)
		}
		ctx.Writef(err.Error())
		auditRet = 0
	} else {
		ctx.StatusCode(http.StatusOK)
	}
	if newWordBank != nil {
		ctx.JSON(newWordBank)
	}
	util.AddAuditLog(userID, userIP, util.AuditModuleDictionary, util.AuditOperationAdd, auditMessage, auditRet)
}

func checkLevel1Valid(dir string) bool {
	if dir == "" {
		return false
	}
	if dir == "敏感词库" || dir == "专有词库" {
		return true
	}
	return false
}

func getWordbankFromReq(ctx context.Context) ([]string, *WordBank, error) {
	paths := make([]string, maxDirDepth)
	for idx := 0; idx < maxDirDepth; idx++ {
		paths[idx] = ctx.FormValue(fmt.Sprintf("level%d", idx+1))
		if paths[idx] == "" {
			break
		}
	}
	if !checkLevel1Valid(paths[0]) {
		return paths, nil, fmt.Errorf("path error")
	}

	ret := &WordBank{}
	nodeType, err := strconv.Atoi(ctx.FormValue("type"))
	if err != nil || nodeType > 1 {
		ret.Type = 0
	}
	ret.Type = nodeType

	if ret.Type == 0 {
		return paths, nil, nil
	}

	ret.Name = ctx.FormValue("name")
	if ret.Name == "" {
		return paths, nil, fmt.Errorf("name cannot be empty")
	}

	ret.Answer = ctx.FormValue("answer")
	ret.SimilarWords = ctx.FormValue("similar_words")
	return paths, ret, nil
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
	errCode, err = util.McUpdateWordBank(appid, userID, userIP, retFile)
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
