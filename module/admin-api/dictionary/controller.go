package dictionary

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"emotibot.com/emotigo/module/admin-api/util/AdminErrors"

	"emotibot.com/emotigo/module/admin-api/ApiError"
	"emotibot.com/emotigo/module/admin-api/auth"
	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/admin-api/util/audit"
	"emotibot.com/emotigo/module/admin-api/util/localemsg"
	"emotibot.com/emotigo/module/admin-api/util/requestheader"
	"emotibot.com/emotigo/pkg/logger"
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
			util.NewEntryPoint("DELETE", "wordbank/{id}", []string{"delete"}, handleDeleteWordbank),
			util.NewEntryPoint("DELETE", "wordbank-dir/dir", []string{"delete"}, handleDeleteWordbankDir),

			util.NewEntryPointWithVer("POST", "upload", []string{"view"}, handleUploadToMySQL, 2),
			util.NewEntryPointWithVer("GET", "download/{file}", []string{}, handleDownloadFromMySQL, 2),
			util.NewEntryPointWithVer("GET", "words/{appid}", []string{}, handleGetWord, 2),
			util.NewEntryPointWithVer("GET", "synonyms/{appid}", []string{}, handleGetSynonyms, 2),

			util.NewEntryPointWithVer("GET", "wordbanks", []string{"view"}, handleGetWordbanksV3, 3),
			util.NewEntryPointWithVer("GET", "wordbank/{id}", []string{"view"}, handleGetWordbankV3, 3),
			util.NewEntryPointWithVer("GET", "class/{id}", []string{"view"}, handleGetWordbankClassV3, 3),
			util.NewEntryPointWithVer("DELETE", "wordbank/{id}", []string{"delete"}, handleDeleteWordbankV3, 3),
			util.NewEntryPointWithVer("DELETE", "class/{id}", []string{"delete"}, handleDeleteWordbankClassV3, 3),
			util.NewEntryPointWithVer("POST", "wordbank", []string{"create"}, handleAddWordbankV3, 3),
			util.NewEntryPointWithVer("POST", "class", []string{"create"}, handleAddWordbankClassV3, 3),
			util.NewEntryPointWithVer("PUT", "wordbank/{id}", []string{"edit"}, handleUpdateWordbankV3, 3),
			util.NewEntryPointWithVer("PUT", "wordbank/{id}/move", []string{"edit"}, handleMoveWordbankV3, 3),
			util.NewEntryPointWithVer("PUT", "class/{id}", []string{"edit"}, handleUpdateWordbankClassV3, 3),

			util.NewEntryPointWithVer("POST", "upload", []string{"view"}, handleUploadToMySQLV3, 3),
			util.NewEntryPointWithVer("GET", "export", []string{}, handleExportFromMySQLV3, 3),
			util.NewEntryPointWithVer("GET", "download/{file}", []string{}, handleDownloadFromMySQLV3, 3),
			util.NewEntryPointWithVer("GET", "words/{appid}", []string{}, handleGetWordV3, 3),
			util.NewEntryPointWithVer("GET", "synonyms/{appid}", []string{}, handleGetSynonymsV3, 3),
			util.NewEntryPointWithCustom("Get", "sync", []string{"edit"}, handleSyncConsul, 3, false, true),
			util.NewEntryPointWithCustom("Get", "sync/{appid}", []string{"edit"}, handleSyncConsul, 3, false, true),
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
	appid := requestheader.GetAppID(r)
	id, err := util.GetMuxIntVar(r, "id")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %s", err.Error()), http.StatusBadRequest)
	}

	wordbank, err := GetWordbank(appid, id)
	if err != nil {
		logger.Info.Printf("Error when get wordbank: %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	if wordbank == nil {
		http.Error(w, "", http.StatusNotFound)
		return
	}
	util.WriteJSON(w, util.GenRetObj(ApiError.SUCCESS, wordbank))
}

func handleUpdateWordbank(w http.ResponseWriter, r *http.Request) {
	appid := requestheader.GetAppID(r)

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
		SyncWordbank(appid, 2)
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
	// util.AddAuditLog(appid, userID, userIP, audit.AuditModuleDictionary, audit.AuditOperationEdit, auditMessage, auditRet)
	audit.AddAuditFromRequestAuto(r, auditMessage, auditRet)
}

func handlePutWordbank(w http.ResponseWriter, r *http.Request) {
	appid := requestheader.GetAppID(r)

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
		SyncWordbank(appid, 2)
	}
	if newWordBank != nil {
		util.WriteJSON(w, util.GenRetObj(ApiError.SUCCESS, newWordBank))
	}
	// util.AddAuditLog(appid, userID, userIP, audit.AuditModuleDictionary, audit.AuditOperationAdd, auditMessage, auditRet)
	audit.AddAuditFromRequestAuto(r, auditMessage, auditRet)
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
	appid := requestheader.GetAppID(r)

	wordbanks, err := GetEntities(appid)
	if err != nil {
		logger.Info.Printf("Error when get entities: %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	util.WriteJSON(w, util.GenRetObj(ApiError.SUCCESS, wordbanks))
}

// handleFileCheck will call api to check if uploaded dictionary is finished
func handleFileCheck(w http.ResponseWriter, r *http.Request) {
	appid := requestheader.GetAppID(r)

	logger.Trace.Printf("Check dictionary info from [%s]", appid)

	ret, err := CheckProcessStatus(appid)
	if err != nil {
		util.WriteJSON(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()))
	} else {
		util.WriteJSON(w, util.GenRetObj(ApiError.SUCCESS, ret))
	}
}

// handleFileCheck will call api to check if uploaded dictionary is finished
func handleFullFileCheck(w http.ResponseWriter, r *http.Request) {
	appid := requestheader.GetAppID(r)

	logger.Trace.Printf("Check dictionary full info from [%s]", appid)

	ret, err := CheckFullProcessStatus(appid)
	if err != nil {
		util.WriteJSON(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()))
	} else {
		util.WriteJSON(w, util.GenRetObj(ApiError.SUCCESS, ret))
	}
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	appid := requestheader.GetAppID(r)
	userID := requestheader.GetUserID(r)
	userIP := requestheader.GetUserIP(r)

	file, info, err := r.FormFile("file")
	defer file.Close()
	logger.Info.Printf("Receive uploaded file: %s", info.Filename)
	logger.Trace.Printf("Uploaded file info %#v", info.Header)

	// 1. receive upload file and check file
	retFile, errCode, err := CheckUploadFile(appid, file, info)
	if err != nil {
		errMsg := ApiError.GetErrorMsg(errCode)
		util.WriteJSON(w, util.GenRetObj(errCode, err.Error()))
		// util.AddAuditLog(appid, userID, userIP, audit.AuditModuleDictionary, audit.AuditOperationImport, fmt.Sprintf("%s: %s", errMsg, err.Error()), 0)
		audit.AddAuditFromRequestAuto(r, fmt.Sprintf("%s: %s", errMsg, err.Error()), 0)
		return
	} else if errCode != ApiError.SUCCESS {
		errMsg := ApiError.GetErrorMsg(errCode)
		util.WriteJSON(w, util.GenSimpleRetObj(errCode))
		// util.AddAuditLog(appid, userID, userIP, audit.AuditModuleDictionary, audit.AuditOperationImport, fmt.Sprintf("%s: %s", errMsg, err.Errorfmt.Sprintf("%s: %s", errMsg, err.Error()), 0)
		audit.AddAuditFromRequestAuto(r, fmt.Sprintf("%s: %s", errMsg, err.Error()), 0)
		return
	}

	// 2. http request to multicustomer
	errCode, err = util.McUpdateWordBank(appid, userID, userIP, retFile)
	if err != nil {
		util.WriteJSON(w, util.GenRetObj(errCode, err.Error()))
		// util.AddAuditLog(appid, userID, userIP, audit.AuditModuleDictionary, audit.AuditOperationImport, fmt.Sprintf("%s %s", util.Msg["Server"], util.Msg["Error"]), 0)
		audit.AddAuditFromRequestAuto(r, fmt.Sprintf("%s %s", util.Msg["Server"], util.Msg["Error"]), 0)
		logger.Error.Printf("update wordbank with multicustomer error: %s", err.Error())
	} else {
		errCode = ApiError.SUCCESS
		util.WriteJSON(w, util.GenSimpleRetObj(errCode))
		// util.AddAuditLog(appid, userID, userIP, audit.AuditModuleDictionary, audit.AuditOperationImport, fmt.Sprintf("%s %s", util.Msg["UploadFile"], info.Filename), 1)
		audit.AddAuditFromRequestAuto(r, fmt.Sprintf("%s %s", util.Msg["UploadFile"], info.Filename), 1)
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
	appid := requestheader.GetAppID(r)
	ret, err := GetDownloadMeta(appid)
	if err != nil {
		util.WriteJSON(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()))
	} else {
		util.WriteJSON(w, util.GenRetObj(ApiError.SUCCESS, ret))
	}
}

func handleUploadToMySQL(w http.ResponseWriter, r *http.Request) {
	errMsg := ""
	appid := requestheader.GetAppID(r)
	now := time.Now()
	var err error
	buf := []byte{}
	defer func() {
		logger.Info.Println("Audit: ", errMsg)
		ret := 0
		if err == nil {
			ret = 1
		}
		// util.AddAuditLog(appid, userID, userIP, audit.AuditModuleDictionary, audit.AuditOperationImport, errMsg, ret)
		audit.AddAuditFromRequestAuto(r, errMsg, ret)

		filename := fmt.Sprintf("wordbank_%s.xlsx", now.Format("20060102150405"))
		RecordDictionaryImportProcess(appid, filename, buf, err)
	}()

	file, info, err := r.FormFile("file")
	defer file.Close()
	logger.Info.Printf("Receive uploaded file: %s", info.Filename)
	logger.Trace.Printf("Uploaded file info %#v", info.Header)
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
	util.WriteJSON(w, util.GenRetObj(ApiError.SUCCESS, wordbanks))
	TriggerUpdateWordbank(appid, wordbanks, 2)
}

func handleDownloadFromMySQL(w http.ResponseWriter, r *http.Request) {
	ret := 0
	appid := requestheader.GetAppID(r)
	filename := util.GetMuxVar(r, "file")
	errMsg := fmt.Sprintf("%s%s: %s", util.Msg["DownloadFile"], util.Msg["Wordbank"], filename)
	defer func() {
		// util.AddAuditLog(appid, userID, userIP, audit.AuditModuleDictionary, audit.AuditOperationExport, errMsg, ret)
		audit.AddAuditFromRequestAuto(r, errMsg, ret)
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

func handleGetWord(w http.ResponseWriter, r *http.Request) {
	appid := util.GetMuxVar(r, "appid")
	if appid == "" {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, "invalid appid"), http.StatusBadRequest)
		return
	}

	err, wordLines, _ := GetWordData(appid)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(strings.Join(wordLines, "\n") + "\n"))
}

func handleGetSynonyms(w http.ResponseWriter, r *http.Request) {
	appid := util.GetMuxVar(r, "appid")
	if appid == "" {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, "invalid appid"), http.StatusBadRequest)
		return
	}

	err, _, synonymLines := GetWordData(appid)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(strings.Join(synonymLines, "\n") + "\n"))
}

func handleDeleteWordbank(w http.ResponseWriter, r *http.Request) {
	appid := requestheader.GetAppID(r)
	errMsg := fmt.Sprintf("%s%s", util.Msg["Delete"], util.Msg["Wordbank"])
	ret := 0
	defer func() {
		// util.AddAuditLog(appid, userID, userIP, audit.AuditModuleDictionary, audit.AuditOperationDelete, errMsg, ret)
		audit.AddAuditFromRequestAuto(r, errMsg, ret)
	}()

	id, err := util.GetMuxIntVar(r, "id")
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, "invalid id"), http.StatusBadRequest)
		return
	}

	wordbankRow, err := GetWordbankRow(appid, id)
	if err != nil {
		if err != sql.ErrNoRows {
			util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
			return
		}
		util.WriteJSON(w, util.GenRetObj(ApiError.SUCCESS, errMsg))
		return
	}

	if wordbankRow != nil {
		errMsg += fmt.Sprintf(": %s/%s", wordbankRow.GetPath(), wordbankRow.Name)
	}

	err = DeleteWordbank(appid, id)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}
	ret = 1
	util.WriteJSON(w, util.GenRetObj(ApiError.SUCCESS, errMsg))
	SyncWordbank(appid, 2)
}

func handleDeleteWordbankDir(w http.ResponseWriter, r *http.Request) {
	appid := requestheader.GetAppID(r)
	path := r.URL.Query().Get("path")
	paths := strings.Split(path, "/")
	errMsg := fmt.Sprintf("%s%s %s", util.Msg["Delete"], util.Msg["WordbankDir"], path)
	ret := 0
	defer func() {
		// util.AddAuditLog(appid, userID, userIP, audit.AuditModuleDictionary, audit.AuditOperationDelete, errMsg, ret)
		audit.AddAuditFromRequestAuto(r, errMsg, ret)
	}()

	if appid == "" {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, "invalid appid"), http.StatusBadRequest)
		return
	}

	if len(paths) <= 1 {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, "invalid path"), http.StatusBadRequest)
		return
	}

	rowCount, err := DeleteWordbankDir(appid, paths)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}
	errMsg += fmt.Sprintf(": %s %d %s", util.Msg["Delete"], rowCount, util.Msg["Row"])
	ret = 1
	util.WriteJSON(w, util.GenRetObj(ApiError.SUCCESS, errMsg))
	SyncWordbank(appid, 2)
}

func handleGetWordbanksV3(w http.ResponseWriter, r *http.Request) {
	appid := requestheader.GetAppID(r)

	wordbanks, errno, err := GetWordbanksV3(appid)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(errno, err.Error()), ApiError.GetHttpStatus(errno))
	} else {
		util.WriteJSON(w, util.GenRetObj(ApiError.SUCCESS, wordbanks))
	}
}

func handleGetWordbankV3(w http.ResponseWriter, r *http.Request) {
	var ret interface{}
	var err error
	errno := ApiError.SUCCESS
	defer func() {
		if err == nil {
			util.WriteJSON(w, util.GenRetObj(errno, ret))
		} else {
			util.WriteJSONWithStatus(w, util.GenRetObj(errno, err.Error()), ApiError.GetHttpStatus(errno))
		}
	}()

	appid := requestheader.GetAppID(r)
	id, err := util.GetMuxIntVar(r, "id")
	if err != nil {
		errno, err = ApiError.REQUEST_ERROR, util.GenBadRequestError("ID")
		return
	}

	wordbank, errno, err := GetWordbankV3(appid, id)
	if err != nil {
		return
	}
	ret = wordbank
	return
}
func handleGetWordbankClassV3(w http.ResponseWriter, r *http.Request) {
	var ret interface{}
	var err error
	errno := ApiError.SUCCESS
	defer func() {
		if err == nil {
			util.WriteJSON(w, util.GenRetObj(errno, ret))
		} else {
			util.WriteJSONWithStatus(w, util.GenRetObj(errno, err.Error()), ApiError.GetHttpStatus(errno))
		}
	}()

	appid := requestheader.GetAppID(r)
	id, err := util.GetMuxIntVar(r, "id")
	if err != nil {
		errno, err = ApiError.REQUEST_ERROR, util.GenBadRequestError("ID")
		return
	}

	wordbankClass, errno, err := GetWordbankClassV3(appid, id)
	if err != nil {
		return
	}

	ret = wordbankClass
	return
}

func handleDeleteWordbankV3(w http.ResponseWriter, r *http.Request) {
	var ret interface{}
	var err error
	errno := ApiError.SUCCESS
	appid := requestheader.GetAppID(r)
	errMsg := fmt.Sprintf("%s%s", util.Msg["Delete"], util.Msg["Wordbank"])
	result := 0
	defer func() {
		if err == nil {
			util.WriteJSON(w, util.GenRetObj(errno, ret))
		} else {
			util.WriteJSONWithStatus(w, util.GenRetObj(errno, err.Error()), ApiError.GetHttpStatus(errno))
		}
		// util.AddAuditLog(appid, userID, userIP, audit.AuditModuleDictionary, audit.AuditOperationDelete, errMsg, result)
		audit.AddAuditFromRequestAuto(r, errMsg, result)
	}()

	id, err := util.GetMuxIntVar(r, "id")
	if err != nil {
		errMsg += ": " + util.Msg["IDError"]
		errno, err = ApiError.REQUEST_ERROR, util.GenBadRequestError("ID")
		return
	}

	wordbank, errno, err := GetWordbankV3(appid, id)
	if err == util.ErrNotFound {
		errno, err = ApiError.SUCCESS, nil
		return
	}
	if err != nil {
		return
	}
	if !wordbank.Editable {
		err = errors.New(util.Msg["ErrorNotEditable"])
		errno = ApiError.REQUEST_ERROR
		return
	}

	errno, err = DeleteWordbankV3(appid, id)
	if err != nil {
		return
	}
	result = 1
	go TriggerUpdateWordbankV3(appid)
	return
}
func handleDeleteWordbankClassV3(w http.ResponseWriter, r *http.Request) {
	appid := requestheader.GetAppID(r)
	errMsg := fmt.Sprintf("%s%s", util.Msg["Delete"], util.Msg["WordbankDir"])
	ret := 0
	defer func() {
		// util.AddAuditLog(appid, userID, userIP, audit.AuditModuleDictionary, audit.AuditOperationDelete, errMsg, ret)
		audit.AddAuditFromRequestAuto(r, errMsg, ret)
	}()
	id, err := util.GetMuxIntVar(r, "id")
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
		return
	}

	class, err := GetWordbanksWithChildrenV3(appid, id)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.IO_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}
	if class == nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.NOT_FOUND_ERROR, "Wordbank class not found"), http.StatusNotFound)
		return
	}
	if !class.Editable {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, "Wordbank class not editable"), http.StatusBadRequest)
		return
	}

	// Check all child wordbank classes and wordbanks of the deleted wordbank are editable before deletion
	if !wordbankClassDeletable(class) {
		err = errors.New("Wordbank class contains more than one child which is not editable")
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
		return
	}

	err = DeleteWordbankClassV3(appid, id)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.IO_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}
	util.WriteJSON(w, util.GenSimpleRetObj(ApiError.SUCCESS))
	go TriggerUpdateWordbankV3(appid)
	return
}

func handleUploadToMySQLV3(w http.ResponseWriter, r *http.Request) {
	var ret interface{}
	var err error
	errno := ApiError.SUCCESS
	errMsg := ""
	appid := requestheader.GetAppID(r)
	locale := requestheader.GetLocale(r)
	now := time.Now()
	buf := []byte{}
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Recovered in %v", r)
		}
		if err == nil {
			util.WriteJSON(w, util.GenRetObj(errno, ret))
		} else {
			util.WriteJSONWithStatus(w, util.GenRetObj(errno, err.Error()), ApiError.GetHttpStatus(errno))
		}
		filename := fmt.Sprintf("wordbank_%s.xlsx", now.Format("20060102150405"))
		importResult := 0
		if err == nil {
			errMsg += " => " + filename
			importResult = 1
		} else {
			errMsg += ": " + err.Error()
		}
		logger.Trace.Println("Audit: ", errMsg)
		// util.AddAuditLog(appid, userID, userIP, audit.AuditModuleDictionary, audit.AuditOperationImport, errMsg, importResult)
		audit.AddAuditFromRequestAuto(r, errMsg, importResult)

		RecordDictionaryImportProcess(appid, filename, buf, err)
	}()

	file, info, err := r.FormFile("file")
	if err != nil {
		errno = ApiError.IO_ERROR
		err = fmt.Errorf("%s: %s", util.Msg["ErrorReadFileError"], err.Error())
		return
	}
	defer file.Close()
	logger.Info.Printf("Receive uploaded file: %s", info.Filename)
	errMsg = fmt.Sprintf("%s%s: %s", util.Msg["UploadFile"], util.Msg["Wordbank"], info.Filename)

	// 1. parseFile
	size := info.Size
	if size == 0 {
		errno = ApiError.REQUEST_ERROR
		err = errors.New(util.Msg["ErrorUploadEmptyFile"])
		return
	}

	buf = make([]byte, size)
	_, err = file.Read(buf)
	if err != nil {
		errno = ApiError.IO_ERROR
		err = fmt.Errorf("%s: %s", util.Msg["ErrorReadFileError"], err.Error())
		return
	}
	root, err := parseDictionaryFromXLSXV3(buf, locale)
	if err != nil {
		errno = ApiError.REQUEST_ERROR
		return
	}

	// 2. save to mysql
	err = SaveWordbankV3Rows(appid, root)
	if err != nil {
		errno = ApiError.DB_ERROR
		return
	}

	ret = map[string]interface{}{
		"root": root,
	}
	go TriggerUpdateWordbankV3(appid)
}

func handleDownloadFromMySQLV3(w http.ResponseWriter, r *http.Request) {
	handleDownloadFromMySQL(w, r)
}

func handleGetWordV3(w http.ResponseWriter, r *http.Request) {
	appid := util.GetMuxVar(r, "appid")
	if appid == "" {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, "invalid appid"), http.StatusBadRequest)
		return
	}
	locale := requestheader.GetLocale(r)

	err, wordLines, _ := GetWordDataV3(appid, locale)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(strings.Join(wordLines, "\n") + "\n"))
}

func handleGetSynonymsV3(w http.ResponseWriter, r *http.Request) {
	appid := util.GetMuxVar(r, "appid")
	if appid == "" {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, "invalid appid"), http.StatusBadRequest)
		return
	}

	locale := requestheader.GetLocale(r)

	err, _, synonymLines := GetWordDataV3(appid, locale)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(strings.Join(synonymLines, "\n") + "\n"))
}

func handleExportFromMySQLV3(w http.ResponseWriter, r *http.Request) {
	appid := requestheader.GetAppID(r)
	locale := requestheader.GetLocale(r)
	if appid == "" {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, "invalid appid"), http.StatusBadRequest)
		return
	}

	buf, err := ExportWordbankV3(appid, locale)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
		return
	}

	now := time.Now()
	filename := fmt.Sprintf("wordbank_%s.xlsx", now.Format("20060102150405"))
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Header().Set("Content-Type", "application/vnd.ms-excel")
	w.Write(buf.Bytes())
}

func handleAddWordbankClassV3(w http.ResponseWriter, r *http.Request) {
	var err error
	var result interface{}
	retCode := ApiError.SUCCESS

	appid := requestheader.GetAppID(r)
	var className string
	defer func() {
		ret := 0

		auditMsg := fmt.Sprintf("%s%s %s", util.Msg["Add"], util.Msg["WordbankDir"], className)
		if err == nil {
			ret = 1
			util.WriteJSON(w, util.GenRetObj(retCode, result))
		} else {
			switch retCode {
			case ApiError.REQUEST_ERROR:
				util.WriteJSONWithStatus(w, util.GenRetObj(retCode, result), http.StatusBadRequest)
			case ApiError.DB_ERROR:
				util.WriteJSONWithStatus(w, util.GenSimpleRetObj(retCode), http.StatusInternalServerError)
			}
			if result != nil {
				auditMsg += fmt.Sprintf(":%v (%s)", result, ApiError.GetErrorMsg(retCode))
			} else {
				auditMsg += ": " + ApiError.GetErrorMsg(retCode)
			}
		}
		// util.AddAuditLog(appid, userID, userIP, audit.AuditModuleDictionary, audit.AuditOperationAdd, auditMsg, ret)
		audit.AddAuditFromRequestAuto(r, auditMsg, ret)
	}()
	pid, err := strconv.Atoi(r.FormValue("pid"))
	if err != nil {
		retCode = ApiError.REQUEST_ERROR
		result = fmt.Sprintf("get pid fail: %s", err.Error())
		return
	}
	if pid != -1 {
		parentClass, _, err := GetWordbankClassV3(appid, pid)
		if err != nil {
			retCode = ApiError.DB_ERROR
			return
		}
		if parentClass == nil {
			retCode = ApiError.NOT_FOUND_ERROR
			result = "Parent not existed"
			return
		}
	}

	className = strings.TrimSpace(r.FormValue("name"))
	if className == "" {
		retCode = ApiError.REQUEST_ERROR
		result = "Class name should not be empty"
		return
	}
	if len(className) > maxDirNameLen {
		retCode, result = ApiError.REQUEST_ERROR, util.Msg["ErrorAPIPathTooLong"]
		return
	}

	class, err := AddWordbankClassV3(appid, className, pid)
	if err != nil {
		retCode = ApiError.DB_ERROR
		result = err.Error()
	} else {
		result = class
	}
	go TriggerUpdateWordbankV3(appid)
}
func handleUpdateWordbankClassV3(w http.ResponseWriter, r *http.Request) {
	var err error
	var result interface{}
	retCode := ApiError.SUCCESS
	var origName string
	var newName string

	appid := requestheader.GetAppID(r)
	defer func() {
		ret := 0

		auditMsg := fmt.Sprintf("%s%s %s -> %s", util.Msg["Modify"], util.Msg["WordbankDir"], origName, newName)
		if err == nil {
			ret = 1
			util.WriteJSON(w, util.GenRetObj(retCode, result))
		} else {
			switch retCode {
			case ApiError.REQUEST_ERROR:
				util.WriteJSONWithStatus(w, util.GenRetObj(retCode, result), http.StatusBadRequest)
			case ApiError.DB_ERROR:
				util.WriteJSONWithStatus(w, util.GenSimpleRetObj(retCode), http.StatusInternalServerError)
			case ApiError.NOT_FOUND_ERROR:
				util.WriteJSONWithStatus(w, util.GenSimpleRetObj(retCode), http.StatusNotFound)
			}
			auditMsg += ": " + ApiError.GetErrorMsg(retCode)
		}
		// util.AddAuditLog(appid, userID, userIP, audit.AuditModuleDictionary, audit.AuditOperationAdd, auditMsg, ret)
		audit.AddAuditFromRequestAuto(r, auditMsg, ret)
	}()

	newName = r.FormValue("name")
	if strings.TrimSpace(newName) == "" {
		retCode, result = ApiError.REQUEST_ERROR, errors.New("Class name should not be empty")
		return
	}

	id, err := util.GetMuxIntVar(r, "id")
	if err != nil {
		retCode, result = ApiError.REQUEST_ERROR, err.Error()
		return
	}
	origClass, _, err := GetWordbankClassV3(appid, id)
	if err != nil {
		retCode, result = ApiError.DB_ERROR, err.Error()
		return
	}
	if origClass == nil {
		err = errors.New("Wordbank class not existed")
		retCode = ApiError.NOT_FOUND_ERROR
		return
	}
	if !origClass.Editable {
		err = errors.New("Wordbank class not editable")
		retCode = ApiError.REQUEST_ERROR
		return
	}

	err = UpdateWordbankClassV3(appid, id, newName)
	if err != nil {
		retCode, result = ApiError.DB_ERROR, err.Error()
	}
	go TriggerUpdateWordbankV3(appid)
}

func handleAddWordbankV3(w http.ResponseWriter, r *http.Request) {
	var err error
	var result interface{}
	var wb *WordBankV3
	retCode := ApiError.SUCCESS

	appid := requestheader.GetAppID(r)
	defer func() {
		ret := 0

		name := ""
		if wb != nil {
			name = wb.Name
		}
		auditMsg := fmt.Sprintf("%s%s %s", util.Msg["Add"], util.Msg["Wordbank"], name)
		if err == nil {
			ret = 1
			util.WriteJSON(w, util.GenRetObj(retCode, result))
		} else {
			switch retCode {
			case ApiError.REQUEST_ERROR:
				util.WriteJSONWithStatus(w, util.GenRetObj(retCode, result), http.StatusBadRequest)
			case ApiError.DB_ERROR:
				util.WriteJSONWithStatus(w, util.GenSimpleRetObj(retCode), http.StatusInternalServerError)
			}
			if result != nil {
				auditMsg += fmt.Sprintf(":%v (%s)", result, ApiError.GetErrorMsg(retCode))
			} else {
				auditMsg += ": " + ApiError.GetErrorMsg(retCode)
			}
		}
		// util.AddAuditLog(appid, userID, userIP, audit.AuditModuleDictionary, audit.AuditOperationAdd, auditMsg, ret)
		audit.AddAuditFromRequestAuto(r, auditMsg, ret)
	}()

	cid, err := strconv.Atoi(r.FormValue("cid"))
	if err != nil {
		retCode, result = ApiError.REQUEST_ERROR, fmt.Sprintf("get cid fail: %s", err.Error())
		return
	}
	if cid != -1 {
		parentClass, _, err := GetWordbankClassV3(appid, cid)
		if err != nil {
			retCode, result = ApiError.DB_ERROR, fmt.Sprintf("Get parent class error: %s", err.Error())
			return
		}
		if parentClass == nil {
			retCode, result = ApiError.NOT_FOUND_ERROR, "Parent not existed"
			return
		}
	}

	wb, err = parseWordbankV3FromRequest(r)
	if err != nil {
		retCode, result = ApiError.REQUEST_ERROR, fmt.Sprintf(util.Msg["ErrorRequestErrorTpl"], err.Error())
		return
	}

	id, err := AddWordbankV3(appid, cid, wb)
	if err != nil {
		retCode, result = ApiError.DB_ERROR, err.Error()
	} else {
		wb.ID = id
		result = wb
	}
	go TriggerUpdateWordbankV3(appid)
}
func handleUpdateWordbankV3(w http.ResponseWriter, r *http.Request) {
	var err error
	var result interface{}
	var wb *WordBankV3
	retCode := ApiError.SUCCESS

	appid := requestheader.GetAppID(r)
	defer func() {
		ret := 0

		name := ""
		if wb != nil {
			name = wb.Name
		}
		auditMsg := fmt.Sprintf("%s%s %s", util.Msg["Modify"], util.Msg["Wordbank"], name)
		if err == nil {
			ret = 1
			util.WriteJSON(w, util.GenRetObj(retCode, result))
		} else {
			switch retCode {
			case ApiError.REQUEST_ERROR:
				util.WriteJSONWithStatus(w, util.GenRetObj(retCode, result), http.StatusBadRequest)
			case ApiError.DB_ERROR:
				util.WriteJSONWithStatus(w, util.GenSimpleRetObj(retCode), http.StatusInternalServerError)
			}
			if result != nil {
				auditMsg += fmt.Sprintf(":%v (%s)", result, ApiError.GetErrorMsg(retCode))
			} else {
				auditMsg += ": " + ApiError.GetErrorMsg(retCode)
			}
		}
		// util.AddAuditLog(appid, userID, userIP, audit.AuditModuleDictionary, audit.AuditOperationAdd, auditMsg, ret)
		audit.AddAuditFromRequestAuto(r, auditMsg, ret)
	}()

	id, err := util.GetMuxIntVar(r, "id")
	if err != nil {
		retCode = ApiError.REQUEST_ERROR
		result = fmt.Sprintf("id fail: %s", err.Error())
		return
	}
	origWordbank, _, err := GetWordbankV3(appid, id)
	if err != nil {
		retCode = ApiError.DB_ERROR
		return
	}
	if origWordbank == nil {
		err = errors.New("Wordbank not existed")
		retCode = ApiError.NOT_FOUND_ERROR
		return
	}

	wb, err = parseWordbankV3FromRequest(r)
	if err != nil {
		retCode = ApiError.REQUEST_ERROR
		result = fmt.Sprintf(util.Msg["ErrorRequestErrorTpl"], err.Error())
		return
	}

	if !origWordbank.Editable {
		if wb.Name != origWordbank.Name {
			err = errors.New("Wordbank not editable")
			result = err.Error()
			retCode = ApiError.REQUEST_ERROR
			return
		}
	}

	err = UpdateWordbankV3(appid, id, wb)
	if err != nil {
		retCode = ApiError.DB_ERROR
		result = err.Error()
	} else {
		wb.ID = id
		result = wb
	}
	go TriggerUpdateWordbankV3(appid)
}
func handleMoveWordbankV3(w http.ResponseWriter, r *http.Request) {
	var err error
	var result interface{}
	retCode := ApiError.SUCCESS

	var wbName string
	var parentName string
	appid := requestheader.GetAppID(r)
	locale := requestheader.GetLocale(r)
	defer func() {
		ret := 0
		auditBuf := bytes.Buffer{}
		auditBuf.WriteString(fmt.Sprintf("%s%s", util.Msg["Move"], util.Msg["Wordbank"]))

		if wbName != "" {
			auditBuf.WriteRune(' ')
			auditBuf.WriteString(wbName)
		}
		if parentName != "" {
			auditBuf.WriteRune(' ')
			auditBuf.WriteString(util.Msg["To"])
			auditBuf.WriteRune(' ')
			auditBuf.WriteString(parentName)
		}
		if err == nil {
			ret = 1
			util.WriteJSON(w, util.GenRetObj(retCode, result))
		} else {
			switch retCode {
			case ApiError.REQUEST_ERROR:
				util.WriteJSONWithStatus(w, util.GenRetObj(retCode, result), http.StatusBadRequest)
			case ApiError.DB_ERROR:
				util.WriteJSONWithStatus(w, util.GenRetObj(retCode, err.Error()), http.StatusInternalServerError)
			}
			auditBuf.WriteString(": " + ApiError.GetErrorMsg(retCode))
			if s, ok := result.(string); ok {
				auditBuf.WriteString(", " + s)
			}
		}
		// util.AddAuditLog(appid, userID, userIP, audit.AuditModuleDictionary, audit.AuditOperationAdd, auditMsg, ret)
		audit.AddAuditFromRequestAuto(r, auditBuf.String(), ret)
	}()

	id, err := util.GetMuxIntVar(r, "id")
	if err != nil {
		retCode, result = ApiError.REQUEST_ERROR, fmt.Sprintf("id fail: %s", err.Error())
		return
	}
	cid, err := strconv.Atoi(r.FormValue("cid"))
	if err != nil {
		retCode, result = ApiError.REQUEST_ERROR, "invalid target"
		return
	}
	if cid > 0 {
		var parentClass *WordBankClassV3
		parentClass, _, err = GetWordbankClassV3(appid, cid)
		if err != nil {
			retCode, result = ApiError.DB_ERROR, fmt.Sprintf("Get parent class error: %s", err.Error())
			return
		}
		if parentClass == nil {
			err = errors.New("Parent not existed")
			retCode, result = ApiError.NOT_FOUND_ERROR, "Parent not existed"
			return
		}
		parentName = parentClass.Name
	} else {
		parentName = localemsg.Get(locale, "DictionaryNoClass")
	}

	origWordbank, _, err := GetWordbankV3(appid, id)
	if err != nil {
		retCode = ApiError.DB_ERROR
		return
	}
	if origWordbank == nil {
		err = errors.New("Wordbank not existed")
		retCode = ApiError.NOT_FOUND_ERROR
		return
	}
	wbName = origWordbank.Name

	err = MoveWordbankV3(appid, id, cid)
	if err == errDuplicated {
		retCode = ApiError.REQUEST_ERROR
		result = util.Msg["ErrorMoveTarget"]
	} else if err != nil {
		retCode = ApiError.DB_ERROR
		result = err.Error()
	} else {
		go TriggerUpdateWordbankV3(appid)
	}
}

func parseWordbankV3FromRequest(r *http.Request) (*WordBankV3, error) {
	if r == nil {
		return nil, errors.New("Param error")
	}

	ret := &WordBankV3{}
	ret.Answer = r.FormValue("answer")

	similarWordStr := strings.TrimSpace(r.FormValue("similar_words"))
	if similarWordStr == "" {
		ret.SimilarWords = []string{}
	} else {
		ret.SimilarWords = strings.Split(similarWordStr, ",")
	}

	ret.Name = strings.TrimSpace(r.FormValue("name"))

	if ret.Name == "" {
		return nil, errors.New("empty name")
	}

	if utf8.RuneCountInString(ret.Name) > maxNameLen {
		return nil, errors.New(util.Msg["ErrorAPINameTooLong"])
	}
	for idx := range ret.SimilarWords {
		if utf8.RuneCountInString(ret.SimilarWords[idx]) > maxSimilarWordsPerRow {
			return nil, errors.New(util.Msg["ErrorAPISimilarTooLong"])
		}
	}

	ret.Editable = true
	return ret, nil
}

func wordbankClassDeletable(wordbanks *WordBankClassV3) bool {
	if !wordbanks.Editable {
		return false
	}

	// Check child wordbank classes
	if children := wordbanks.Children; children != nil {
		for _, child := range children {
			if !wordbankClassDeletable(child) {
				return false
			}
		}
	}

	// Check child wordbanks
	if wordbanks := wordbanks.Wordbank; wordbanks != nil {
		for _, wordbank := range wordbanks {
			if !wordbank.Editable {
				return false
			}
		}
	}

	return true
}

func handleSyncConsul(w http.ResponseWriter, r *http.Request) {
	appid := util.GetMuxVar(r, "appid")
	var appids []string
	var err error
	if appid == "" {
		appids, err = auth.GetAllApps()
	} else {
		appids = []string{appid}
	}
	if err != nil {
		util.ReturnError(w, AdminErrors.ErrnoDBError, err.Error())
		return
	}

	go syncConsul(appids)
	util.Return(w, nil, appids)
}

func syncConsul(appids []string) {
	for idx, appid := range appids {
		err := TriggerUpdateWordbankV3(appid)
		if err != nil {
			logger.Error.Println("Update wordbank consul fail, ", err.Error())
			return
		} else {
			logger.Info.Printf("Finish wordbank sync %d/%d\n", idx+1, len(appids))
		}
	}
}
