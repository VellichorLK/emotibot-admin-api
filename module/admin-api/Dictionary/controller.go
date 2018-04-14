package Dictionary

import (
	"bytes"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"emotibot.com/emotigo/module/admin-api/ApiError"
	"emotibot.com/emotigo/module/admin-api/util"
)

const (
	defaultInternalURL = "http://127.17.0.1"
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
			util.NewEntryPoint("GET", "wordbank/{id}", []string{"view"}, handleGetWordbank),

			util.NewEntryPointWithVer("POST", "upload", []string{"view"}, handleUploadToMySQL, 2),
			util.NewEntryPointWithVer("GET", "download/{file}", []string{"view"}, handleDownloadFromMySQL, 2),
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

func handleGetWordbank(w http.ResponseWriter, r *http.Request) {
	appid := util.GetAppID(r)
	id, err := util.GetMuxIntVar(r, "id")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %s", err.Error()), http.StatusBadRequest)
	}

	wordbank, err := GetWordbank(appid, id)
	if err != nil {
		util.LogInfo.Printf("Error when get wordbank: %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	if wordbank == nil {
		http.Error(w, "", http.StatusNotFound)
		return
	}
	util.WriteJSON(w, util.GenRetObj(ApiError.SUCCESS, wordbank))
}

func handleUpdateWordbank(w http.ResponseWriter, r *http.Request) {
	appid := util.GetAppID(r)
	userID := util.GetUserID(r)
	userIP := util.GetUserIP(r)

	updatedWordbank := &WordBank{}
	err := util.ReadJSON(r, updatedWordbank)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	origWordbank, err := GetWordbank(appid, *updatedWordbank.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	retCode, err := UpdateWordbank(appid, updatedWordbank)
	auditRet := 1
	if err != nil {
		if retCode == ApiError.REQUEST_ERROR {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		auditRet = 0
	} else {
		http.Error(w, "", http.StatusOK)
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

func handlePutWordbank(w http.ResponseWriter, r *http.Request) {
	appid := util.GetAppID(r)
	userID := util.GetUserID(r)
	userIP := util.GetUserIP(r)

	paths, newWordBank, err := getWordbankFromReq(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
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
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		auditRet = 0
	} else {
		http.Error(w, "", http.StatusOK)
	}
	if newWordBank != nil {
		util.WriteJSON(w, util.GenRetObj(ApiError.SUCCESS, newWordBank))
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

func getWordbankFromReq(r *http.Request) ([]string, *WordBank, error) {
	paths := make([]string, maxDirDepth)
	for idx := 0; idx < maxDirDepth; idx++ {
		paths[idx] = r.FormValue(fmt.Sprintf("level%d", idx+1))
		if paths[idx] == "" {
			break
		}
	}
	if !checkLevel1Valid(paths[0]) {
		return paths, nil, fmt.Errorf("path error")
	}

	ret := &WordBank{}
	nodeType, err := strconv.Atoi(r.FormValue("type"))
	if err != nil || nodeType > 1 {
		ret.Type = 0
	}
	ret.Type = nodeType

	if ret.Type == 0 {
		return paths, nil, nil
	}

	ret.Name = r.FormValue("name")
	if ret.Name == "" {
		return paths, nil, fmt.Errorf("name cannot be empty")
	}

	ret.Answer = r.FormValue("answer")
	ret.SimilarWords = r.FormValue("similar_words")
	return paths, ret, nil
}

func handleGetWordbanks(w http.ResponseWriter, r *http.Request) {
	appid := util.GetAppID(r)

	wordbanks, err := GetEntities(appid)
	if err != nil {
		util.LogInfo.Printf("Error when get entities: %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	util.WriteJSON(w, util.GenRetObj(ApiError.SUCCESS, wordbanks))
}

// handleFileCheck will call api to check if uploaded dictionary is finished
func handleFileCheck(w http.ResponseWriter, r *http.Request) {
	appid := util.GetAppID(r)

	util.LogTrace.Printf("Check dictionary info from [%s]", appid)

	ret, err := CheckProcessStatus(appid)
	if err != nil {
		util.WriteJSON(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()))
	} else {
		util.WriteJSON(w, util.GenRetObj(ApiError.SUCCESS, ret))
	}
}

// handleFileCheck will call api to check if uploaded dictionary is finished
func handleFullFileCheck(w http.ResponseWriter, r *http.Request) {
	appid := util.GetAppID(r)

	util.LogTrace.Printf("Check dictionary full info from [%s]", appid)

	ret, err := CheckFullProcessStatus(appid)
	if err != nil {
		util.WriteJSON(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()))
	} else {
		util.WriteJSON(w, util.GenRetObj(ApiError.SUCCESS, ret))
	}
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	appid := util.GetAppID(r)
	userID := util.GetUserID(r)
	userIP := util.GetUserIP(r)

	file, info, err := r.FormFile("file")
	defer file.Close()
	util.LogInfo.Printf("Receive uploaded file: %s", info.Filename)
	util.LogTrace.Printf("Uploaded file info %#v", info.Header)

	// 1. receive upload file and check file
	retFile, errCode, err := CheckUploadFile(appid, file, info)
	if err != nil {
		errMsg := ApiError.GetErrorMsg(errCode)
		util.WriteJSON(w, util.GenRetObj(errCode, err.Error()))
		util.AddAuditLog(userID, userIP, util.AuditModuleDictionary, util.AuditOperationImport, fmt.Sprintf("%s: %s", errMsg, err.Error()), 0)
		return
	} else if errCode != ApiError.SUCCESS {
		errMsg := ApiError.GetErrorMsg(errCode)
		util.WriteJSON(w, util.GenSimpleRetObj(errCode))
		util.AddAuditLog(userID, userIP, util.AuditModuleDictionary, util.AuditOperationImport, fmt.Sprintf("%s: %s", errMsg, err.Error()), 0)
		return
	}

	// 2. http request to multicustomer
	errCode, err = util.McUpdateWordBank(appid, userID, userIP, retFile)
	if err != nil {
		util.WriteJSON(w, util.GenRetObj(errCode, err.Error()))
		util.AddAuditLog(userID, userIP, util.AuditModuleDictionary, util.AuditOperationImport, fmt.Sprintf("%s %s", util.Msg["Server"], util.Msg["Error"]), 0)
		util.LogError.Printf("update wordbank with multicustomer error: %s", err.Error())
	} else {
		errCode = ApiError.SUCCESS
		util.WriteJSON(w, util.GenSimpleRetObj(errCode))
		util.AddAuditLog(userID, userIP, util.AuditModuleDictionary, util.AuditOperationImport, fmt.Sprintf("%s %s", util.Msg["UploadFile"], info.Filename), 1)
	}
}

func handleDownload(w http.ResponseWriter, r *http.Request) {
	d := map[string]interface{}{
		"result": true,
		"entry":  "download",
	}

	// TODO: WIP
	// 1. get file from input version
	// 2. output raw

	util.WriteJSON(w, d)
}

func handleDownloadMeta(w http.ResponseWriter, r *http.Request) {
	// 1. select from db last two row
	// 2. return response
	appid := util.GetAppID(r)
	ret, err := GetDownloadMeta(appid)
	if err != nil {
		util.WriteJSON(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()))
	} else {
		util.WriteJSON(w, util.GenRetObj(ApiError.SUCCESS, ret))
	}
}

func handleUploadToMySQL(w http.ResponseWriter, r *http.Request) {
	errMsg := ""
	appid := util.GetAppID(r)
	now := time.Now()
	var err error
	buf := []byte{}
	userID := util.GetUserID(r)
	userIP := util.GetUserIP(r)
	defer func() {
		util.LogInfo.Println("Audit: ", errMsg)
		ret := 0
		if err == nil {
			ret = 1
		}
		util.AddAuditLog(userID, userIP, util.AuditModuleDictionary, util.AuditOperationImport, errMsg, ret)

		filename := fmt.Sprintf("wordbank_%s.xlsx", now.Format("20060102150405"))
		RecordDictionaryImportProcess(appid, filename, buf, err)
	}()

	file, info, err := r.FormFile("file")
	defer file.Close()
	util.LogInfo.Printf("Receive uploaded file: %s", info.Filename)
	util.LogTrace.Printf("Uploaded file info %#v", info.Header)
	errMsg = fmt.Sprintf("%s%s: %s", util.Msg["UploadFile"], util.Msg["Wordbank"], info.Filename)

	// 1. parseFile
	size := info.Size
	buf = make([]byte, size)
	_, err = file.Read(buf)
	if err != nil {
		errMsg = err.Error()
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
		return
	}
	wordbanks, err := parseDictionaryFromXLSX(buf)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
		return
	}

	// 2. save to mysql
	err = SaveWordbankRows(appid, wordbanks)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}

	// 3. save to local file which can be get from url
	err, md5Words, md5Synonyms := SaveWordbankToFile(appid, wordbanks)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}
	util.WriteJSON(w, util.GenRetObj(ApiError.SUCCESS, wordbanks))

	// 4. Update consul key
	// TODO: use relative to compose the url
	url := getEnvironment("INTERNAL_URL")
	if url == "" {
		url = defaultInternalURL
	}
	consulJSON := map[string]interface{}{
		"url":         fmt.Sprintf("%s/Files/settings/%s/%s.txt", url, appid, appid),
		"md5":         md5Words,
		"synonym-url": fmt.Sprintf("%s/Files/settings/%s/%s_synonyms.txt", url, appid, appid),
		"synonym-md5": md5Synonyms,
		"timestamp":   now.UnixNano() / 1000000,
	}
	util.ConsulUpdateEntity(appid, consulJSON)
	util.LogInfo.Printf("Update to consul:\n%+v\n", consulJSON)
}

func handleDownloadFromMySQL(w http.ResponseWriter, r *http.Request) {
	ret := 0
	appid := util.GetAppID(r)
	userID := util.GetUserID(r)
	userIP := util.GetUserIP(r)
	filename := util.GetMuxVar(r, "file")
	errMsg := fmt.Sprintf("%s%s: %s", util.Msg["DownloadFile"], util.Msg["Wordbank"], filename)
	defer func() {
		util.AddAuditLog(userID, userIP, util.AuditModuleDictionary, util.AuditOperationImport, errMsg, ret)
	}()

	if filename == "" {
		util.WriteJSONWithStatus(w,
			util.GenRetObj(ApiError.REQUEST_ERROR, "invalid filename"),
			http.StatusBadRequest)
		return
	}

	buf, err := GetWordbankFile(appid, filename)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "", http.StatusNotFound)
			return
		}
		util.WriteJSONWithStatus(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ret = 1
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Header().Set("Content-Type", "application/vnd.ms-excel")
	w.Write(buf)
}
