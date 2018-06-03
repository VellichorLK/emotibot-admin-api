package QA

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"emotibot.com/emotigo/module/admin-api/FAQ"
	"emotibot.com/emotigo/module/admin-api/util"
)

// ModuleInfo is web info of questions entrypoints
var ModuleInfo util.ModuleInfo
var apiClient util.MultiCustomerClient

func init() {
	ModuleInfo = util.ModuleInfo{
		ModuleName: "qa",
		EntryPoints: []util.EntryPoint{
			util.NewEntryPoint(http.MethodGet, "questions/operations", []string{"view"}, viewOperations),
			util.NewEntryPoint(http.MethodPost, "questions/operations/import", []string{"import"}, importExcel),
			util.NewEntryPoint(http.MethodPost, "questions/operations/export", []string{"export"}, exportExcel),
			util.NewEntryPoint(http.MethodGet, "questions/operations/{id}/download", []string{}, download),
			util.NewEntryPoint(http.MethodGet, "questions/operations/{id}/progress", []string{"view"}, progress),
		},
	}
	t, _ := time.ParseDuration("30s")
	apiClient = util.MultiCustomerHttpClient{
		Timeout: t,
	}

}

var successAuditMessage string
var failedAuditMessage string

type errorJSON struct {
	Message string `json:"message"`
}

func importExcel(w http.ResponseWriter, r *http.Request) {
	type returnJSON struct {
		Message string `json:"message,omitempty"`
		StateID int    `json:"state_id,omitempty"`
		UserID  string `json:"user_id,omitempty"`
		Action  string `json:"action,omitempty"`
	}
	var jsonResponse returnJSON
	var userID = util.GetUserID(r)
	var userIP = util.GetUserIP(r)
	var appid = util.GetAppID(r)
	var status = 0 // 0 == failed, 1 == success
	var fileName, reason string

	var method string
	mode := r.FormValue("mode")
	switch mode {
	case "incre":
		method = "增量导入"
	case "full":
		method = "全量导入"
	default:
		jsonResponse.Message = "上传模式只支援[全量(incre), 增量(full)]"
		util.WriteJSONWithStatus(w, jsonResponse, http.StatusBadRequest)
		return
	}

	_, fileHeader, err := r.FormFile("file")
	if err != nil {
		jsonResponse.Message = "无上传档案"
		util.WriteJSONWithStatus(w, jsonResponse, http.StatusBadRequest)
		fileName = "无"
		reason = jsonResponse.Message
		auditMessage := fmt.Sprintf("[%s]:%s:%s", method, fileName, reason)
		util.AddAuditLog(appid, userID, userIP, util.AuditModuleQA, util.AuditOperationImport, auditMessage, status)
		return
	}
	ext := filepath.Ext(fileHeader.Filename)
	if strings.Compare(ext, ".xlsx") != 0 {
		jsonResponse.Message = "[" + fileHeader.Filename + "] 后缀名为" + ext + " 请上传后缀名为.xlsx的文件!"
		util.WriteJSONWithStatus(w, jsonResponse, http.StatusBadRequest)
		fileName = fileHeader.Filename
		reason = jsonResponse.Message
		auditMessage := fmt.Sprintf("[%s]:%s:%s", method, fileName, reason)
		util.AddAuditLog(appid, userID, userIP, util.AuditModuleQA, util.AuditOperationImport, auditMessage, status)
		return
	}
	fileName = fileHeader.Filename
	response, err := apiClient.McImportExcel(*fileHeader, userID, userIP, mode, appid)

	switch err {
	case util.ErrorMCLock: //503
		jsonResponse.Message = err.Error()
		jsonResponse.UserID = response.SyncInfo.UserID
		jsonResponse.Action = response.SyncInfo.Action
		util.WriteJSONWithStatus(w, jsonResponse, http.StatusServiceUnavailable)
		reason = "导入正在执行中"
		auditMessage := fmt.Sprintf("[%s]:%s:%s", method, fileName, reason)
		util.AddAuditLog(appid, userID, userIP, util.AuditModuleQA, util.AuditOperationImport, auditMessage, status)
	case nil: //200
		status = 1
		jsonResponse.StateID = response.SyncInfo.StatID
		util.WriteJSON(w, jsonResponse)
	default: //500
		jsonResponse.Message = "服务器不正常, " + err.Error()
		util.WriteJSONWithStatus(w, jsonResponse, http.StatusInternalServerError)
		reason = "服务器不正常"
		auditMessage := fmt.Sprintf("[%s]:%s:%s", method, fileName, reason)
		util.AddAuditLog(appid, userID, userIP, util.AuditModuleQA, util.AuditOperationImport, auditMessage, status)
	}

}

func exportExcel(w http.ResponseWriter, r *http.Request) {
	type successJSON struct {
		StateID int `json:"state_id"`
	}
	type errorJSON struct {
		Message string `json:"message"`
		UserID  string `json:"user_id,omitempty"`
	}
	var err error
	var userID = util.GetUserID(r)
	var userIP = util.GetUserIP(r)
	var appid = util.GetAppID(r)

	// check if we need should do any db query
	condition, err := FAQ.ParseCondition(r)
	if err != nil {
		util.LogError.Printf("Error happened while parsing conditions: %s", err.Error())
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	var answerIDs []string
	var mcResponse util.MCResponse

	originCategoryId := condition.CategoryId
	condition.CategoryId = 0
	if FAQ.HasCondition(condition) {
		answerIDs = []string{}
		condition.CategoryId = originCategoryId
		_, aids, err := FAQ.DoFilter(condition, appid)
		if err != nil {
			util.LogError.Printf("Error happened while fetch question ids & answer ids: %s", err.Error())
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		for _, aidSlice := range aids {
			for _, aid := range aidSlice {
				answerIDs = append(answerIDs, aid)
			}
		}
		mcResponse, err = apiClient.McExportExcel(userID, userIP, answerIDs, appid, -2)
	} else {
		if originCategoryId != 0 {
			mcResponse, err = apiClient.McExportExcel(userID, userIP, answerIDs, appid, originCategoryId)
		} else {
			mcResponse, err = apiClient.McExportExcel(userID, userIP, answerIDs, appid, -2)
		}
	}

	switch err {
	case nil: // 200
		util.WriteJSON(w, successJSON{StateID: mcResponse.SyncInfo.StatID})
		auditLogContent, err := genQAExportAuditLog(appid, &condition, mcResponse.SyncInfo.StatID)
		if err != nil {
			util.LogError.Printf("Error happened while parsing conditions: %s", err.Error())
		}
		util.AddAuditLog(appid, userID, userIP, util.AuditModuleQA, util.AuditOperationExport, auditLogContent, 1)
	case util.ErrorMCLock: //503 MCError
		util.WriteJSONWithStatus(w, errorJSON{err.Error(), mcResponse.SyncInfo.UserID}, http.StatusServiceUnavailable)
		util.AddAuditLog(appid, userID, userIP, util.AuditModuleQA, util.AuditOperationExport, "[全量导出]: 其他操作正在进行中", 0)
	default: //500 error
		util.WriteJSONWithStatus(w, errorJSON{err.Error(), ""}, http.StatusInternalServerError)
		util.AddAuditLog(appid, userID, userIP, util.AuditModuleQA, util.AuditOperationExport, "[全量导出]: 服务器不正常", 0)
	}
}

func download(w http.ResponseWriter, r *http.Request) {
	id := util.GetMuxVar(r, "id")
	db := util.GetMainDB()
	if db == nil {
		util.LogError.Println("Main DB Connection failed")
		util.WriteJSONWithStatus(w, errorJSON{"Main DB Connection failed"}, http.StatusInternalServerError)
	}
	rows, err := db.Query("SELECT content, status FROM state_machine WHERE state_id = ?", id)
	if err != nil {
		util.WriteJSONWithStatus(w, errorJSON{err.Error()}, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// var userID = util.GetUserID(r)
	// var userIP = util.GetUserIP(r)
	var content []byte
	var status string

	if !rows.Next() {
		http.Error(w, "", http.StatusNoContent)
		return
	}
	rows.Scan(&content, &status)
	if strings.Compare(status, "success") == 0 || strings.Compare(status, "fail") == 0 {
		w.Header().Set("Content-Disposition", "attachment; filename=other_"+id+".xlsx")
		w.Header().Set("Cache-Control", "public")
		w.Write([]byte(content))
	} else if strings.Compare(status, "running") == 0 {
		util.WriteJSONWithStatus(w, errorJSON{"still running, download later."}, http.StatusServiceUnavailable)
	} else {
		util.WriteJSONWithStatus(w, errorJSON{"unknown status '" + status + "'."}, http.StatusInternalServerError)
	}

}

func progress(w http.ResponseWriter, r *http.Request) {
	type successJSON struct {
		ID          int          `json:"state_id"`
		Status      string       `json:"status"`
		CreatedTime JSONUnixTime `json:"created_time"`
		ExtraInfo   string       `json:"extra_info"`
	}

	db := util.GetMainDB()
	if db == nil {
		util.LogError.Println("Main DB Connection failed")
		util.WriteJSONWithStatus(w, errorJSON{"Main DB Connection failed"}, http.StatusInternalServerError)
		return
	}
	statID, err := util.GetMuxIntVar(r, "id")
	if err != nil {
		util.WriteJSONWithStatus(w, errorJSON{Message: err.Error()}, http.StatusBadRequest)
		return
	}
	rows, err := db.Query("SELECT status, created_time, extra_info FROM state_machine WHERE state_id = ?", statID)
	var returnJSON = successJSON{ID: statID}
	if err != nil {
		util.WriteJSONWithStatus(w, errorJSON{Message: err.Error()}, http.StatusInternalServerError)
		return
	}
	if !rows.Next() {
		http.Error(w, "", http.StatusNotFound)
		return
	}
	rows.Scan(&returnJSON.Status, &returnJSON.CreatedTime, &returnJSON.ExtraInfo)
	defer rows.Close()

	if err = rows.Err(); err != nil {
		util.WriteJSONWithStatus(w, errorJSON{Message: err.Error()}, http.StatusInternalServerError)
	} else {
		util.WriteJSON(w, returnJSON)
	}
}

func viewOperations(w http.ResponseWriter, r *http.Request) {
	type operation struct {
		StateID     int          `json:"state_id"`
		Action      string       `json:"action"`
		Status      string       `json:"status"`
		UserID      string       `json:"user_id"`
		CreatedTime JSONUnixTime `json:"created_time"`
		UpdatedTime JSONUnixTime `json:"updated_time"`
		ExtraInfo   string       `json:"extra_info"`
	}

	db := util.GetMainDB()
	if db == nil {
		util.LogError.Println("Main DB Connection failed")
		util.WriteJSONWithStatus(w, errorJSON{"Main DB Connection failed"}, http.StatusInternalServerError)
		return
	}
	values := r.Form
	var limitedQuery string
	var whereQuery []string
	var parameters []interface{}
	if numString, ok := values["num"]; ok {
		num, err := strconv.Atoi(numString[0])
		if err != nil {
			util.WriteJSONWithStatus(w, errorJSON{err.Error()}, http.StatusBadRequest)
			return
		}
		limitedQuery = fmt.Sprintf("LIMIT %d", num)
	}
	if userID, ok := values["userID"]; ok {
		parameters = append(parameters, userID[0])
		whereQuery = append(whereQuery, "user_id = ?")
	}
	if action, ok := values["action"]; ok {
		parameters = append(parameters, action[0])
		whereQuery = append(whereQuery, "action = ?")
	}
	if status, ok := values["status"]; ok {
		parameters = append(parameters, status[0])
		whereQuery = append(whereQuery, "status = ?")
	}
	var SQLQuery = "SELECT state_id, user_id, action, status, created_time, updated_time, extra_info FROM state_machine "
	if len(parameters) > 0 {
		SQLQuery += "WHERE " + strings.Join(whereQuery, " AND ")
	}
	rows, err := db.Query(SQLQuery+" ORDER BY updated_time DESC "+limitedQuery, parameters...)
	if err != nil {
		util.WriteJSONWithStatus(w, errorJSON{err.Error()}, http.StatusInternalServerError)
		return
	}
	type returnJSON struct {
		Records []operation `json:"records"`
	}
	var operations []operation

	for rows.Next() {
		var op = operation{}
		rows.Scan(&op.StateID, &op.UserID, &op.Action, &op.Status, &op.CreatedTime, &op.UpdatedTime, &op.ExtraInfo)
		operations = append(operations, op)
	}
	defer rows.Close()

	util.WriteJSON(w, returnJSON{Records: operations})
}

//JSONUnixTime are use for formatting to Unix Time Mill Second
type JSONUnixTime time.Time

// MarshalJSON is for JSON Marshal usage
func (t JSONUnixTime) MarshalJSON() ([]byte, error) {
	//do your serializing here
	stamp := fmt.Sprintf("%d", time.Time(t).UnixNano()/1000000)
	return []byte(stamp), nil
}
