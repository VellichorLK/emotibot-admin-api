package QA

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"emotibot.com/emotigo/module/vipshop-admin/util"
	"github.com/kataras/iris/context"
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
			util.NewEntryPoint(http.MethodGet, "questions/operations/{id:int}/download", []string{"view"}, download),
			util.NewEntryPoint(http.MethodGet, "questions/operations/{id:int}/progress", []string{"view"}, progress),
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

func importExcel(ctx context.Context) {
	type returnJSON struct {
		Message string `json:"message,omitempty"`
		StateID int    `json:"state_id,omitempty"`
		UserID  string `json:"user_id,omitempty"`
		Action  string `json:"action,omitempty"`
	}
	var jsonResponse returnJSON
	var userID = util.GetUserID(ctx)
	var userIP = util.GetUserIP(ctx)
	var status = 0 // 0 == failed, 1 == success
	var auditMessage string

	_, fileHeader, err := ctx.FormFile("file")
	if err != nil {
		jsonResponse.Message = "请上传档案"
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonResponse)
		util.AddAuditLog(userID, userIP, util.AuditModuleQA, util.AuditOperationImport, "导入檔案传送失败", status)
		return
	}
	ext := filepath.Ext(fileHeader.Filename)
	if strings.Compare(ext, ".xlsx") != 0 {
		jsonResponse.Message = "[" + fileHeader.Filename + "] 后缀名为" + ext + " 请上传后缀名为.xlsx的文件!"
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonResponse)
		util.AddAuditLog(userID, userIP, util.AuditModuleQA, util.AuditOperationImport, "导入檔案传送失败", status)
		return
	}

	mode := ctx.FormValue("mode")
	switch mode {
	case "incre":
		auditMessage = fmt.Sprintf("增量导入文件 %s", fileHeader.Filename)
	case "full":
		auditMessage = fmt.Sprintf("全量导入文件 %s", fileHeader.Filename)
	default:
		jsonResponse.Message = "上傳模式只支援[全量(incre), 增量(full)]"
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(jsonResponse)
		util.AddAuditLog(userID, userIP, util.AuditModuleQA, util.AuditOperationImport, auditMessage, status)
	}

	response, err := apiClient.McImportExcel(*fileHeader, userID, userIP, mode)

	switch err {
	case util.ErrorMCLock: //503
		jsonResponse.Message = err.Error()
		jsonResponse.UserID = response.SyncInfo.UserID
		jsonResponse.Action = response.SyncInfo.Action
		ctx.StatusCode(http.StatusServiceUnavailable)
		ctx.JSON(jsonResponse)
		util.AddAuditLog(userID, userIP, util.AuditModuleQA, util.AuditOperationImport, auditMessage, status)
	case nil: //200
		status = 1
		jsonResponse.StateID = response.SyncInfo.StatID
		ctx.JSON(jsonResponse)
		util.AddAuditLog(userID, userIP, util.AuditModuleQA, util.AuditOperationImport, auditMessage, status)
	default: //500
		jsonResponse.Message = "服务器不正常, " + err.Error()
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(jsonResponse)
		util.AddAuditLog(userID, userIP, util.AuditModuleQA, util.AuditOperationImport, auditMessage, status)
	}

}

func exportExcel(ctx context.Context) {
	type successJSON struct {
		StateID int `json:"state_id"`
	}
	type errorJSON struct {
		Message string `json:"message"`
		UserID  string `json:"user_id,omitempty"`
	}
	var userID = util.GetUserID(ctx)
	var userIP = util.GetUserIP(ctx)
	var status = 0 // 0 == failed, 1 == success
	var auditMessage = "全量导出"

	mcResponse, err := apiClient.McExportExcel(userID, userIP)
	switch err {
	case nil: // 200
		status = 1
		ctx.StatusCode(http.StatusOK)
		ctx.JSON(successJSON{StateID: mcResponse.SyncInfo.StatID})
		util.AddAuditLog(userID, userIP, util.AuditModuleQA, util.AuditOperationExport, auditMessage, status)
	case util.ErrorMCLock: //503 MCError
		ctx.StatusCode(http.StatusServiceUnavailable)
		ctx.JSON(errorJSON{err.Error(), mcResponse.SyncInfo.UserID})
		util.AddAuditLog(userID, userIP, util.AuditModuleQA, util.AuditOperationExport, auditMessage, status)
	default: //500 error
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(errorJSON{err.Error(), ""})
		util.AddAuditLog(userID, userIP, util.AuditModuleQA, util.AuditOperationExport, auditMessage, status)
	}
}

func download(ctx context.Context) {
	id := ctx.Params().Get("id")
	db := util.GetMainDB()
	if db == nil {
		util.LogError.Println("Main DB Connection failed")
		ctx.StatusCode(500)
		ctx.JSON(errorJSON{"Main DB Connection failed"})
	}
	rows, err := db.Query("SELECT content, status FROM state_machine WHERE state_id = ?", id)
	if err != nil {
		ctx.StatusCode(500)
		ctx.JSON(errorJSON{err.Error()})
		return
	}
	defer rows.Close()

	var content []byte
	var status string
	if !rows.Next() {
		ctx.StatusCode(http.StatusNoContent)
		return
	}
	rows.Scan(&content, &status)
	if strings.Compare(status, "success") == 0 || strings.Compare(status, "fail") == 0 {
		ctx.Header("Content-Disposition", "attachment; filename=other_"+id+".xlsx")
		ctx.Header("Cache-Control", "public")
		ctx.Binary(content)
	} else if strings.Compare(status, "running") == 0 {
		ctx.StatusCode(http.StatusServiceUnavailable)
		ctx.JSON(errorJSON{"still running, download later."})
	} else {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(errorJSON{"unknown status '" + status + "'."})
	}

}

func progress(ctx context.Context) {
	type successJSON struct {
		ID          int          `json:"state_id"`
		Status      string       `json:"status"`
		CreatedTime JSONUnixTime `json:"created_time"`
		ExtraInfo   string       `json:"extra_info"`
	}

	db := util.GetMainDB()
	if db == nil {
		util.LogError.Println("Main DB Connection failed")
		ctx.StatusCode(500)
		ctx.JSON("Main DB Connection failed")
	}
	statID, err := strconv.Atoi(ctx.Params().Get("id"))
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(errorJSON{Message: err.Error()})
	}
	rows, err := db.Query("SELECT status, created_time, extra_info FROM state_machine WHERE state_id = ?", statID)
	var returnJSON = successJSON{ID: statID}
	if !rows.Next() {
		ctx.StatusCode(http.StatusNotFound)
		return
	}
	rows.Scan(&returnJSON.Status, &returnJSON.CreatedTime, &returnJSON.ExtraInfo)
	defer rows.Close()

	if err = rows.Err(); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
	}

	ctx.JSON(returnJSON)
}

func viewOperations(ctx context.Context) {
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
		ctx.StatusCode(500)
		ctx.JSON("Main DB Connection failed")
	}
	values := ctx.FormValues()
	var limitedQuery string
	var whereQuery []string
	var parameters []interface{}
	if numString, ok := values["num"]; ok {
		num, err := strconv.Atoi(numString[0])
		if err != nil {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(errorJSON{err.Error()})
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
	rows, err := db.Query(SQLQuery+" ORDER BY updated_time DESC"+limitedQuery, parameters...)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(errorJSON{err.Error()})
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

	ctx.JSON(returnJSON{Records: operations})

}

//JSONUnixTime are use for formatting to Unix Time Mill Second
type JSONUnixTime time.Time

// MarshalJSON is for JSON Marshal usage
func (t JSONUnixTime) MarshalJSON() ([]byte, error) {
	//do your serializing here
	stamp := fmt.Sprintf("%d", time.Time(t).UnixNano()/1000000)
	return []byte(stamp), nil
}
