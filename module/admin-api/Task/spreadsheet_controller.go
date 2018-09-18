package Task

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"emotibot.com/emotigo/module/admin-api/ApiError"
	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/admin-api/util/audit"
	"emotibot.com/emotigo/pkg/logger"
)

func handleUploadSpreadSheet(w http.ResponseWriter, r *http.Request) {
	appID := r.FormValue("appId")
	scenarioID := r.FormValue("scenarioId")
	scenarioString := r.FormValue("scenario")
	retCode := ApiError.SUCCESS
	var auditMsg bytes.Buffer
	var ret string
	defer func() {
		status := ApiError.GetHttpStatus(retCode)
		logger.Trace.Printf("Upload spreadsheet ret: %d, %s\n", retCode, ret)
		util.WriteJSONWithStatus(w, map[string]interface{}{
			"error":  ret,
			"return": retCode,
		}, status)

		if retCode == ApiError.SUCCESS {
			addAuditLog(r, audit.AuditOperationImport, auditMsg.String(), true)
		} else {
			auditMsg.WriteString(fmt.Sprintf(", %s", ret))
			addAuditLog(r, audit.AuditOperationImport, auditMsg.String(), false)
		}
	}()
	auditMsg.WriteString(fmt.Sprintf("%s%s", util.Msg["UploadFile"], util.Msg["Spreadsheet"]))

	file, info, err := r.FormFile("spreadsheet")
	if err != nil {
		retCode = ApiError.IO_ERROR
		ret = fmt.Sprintf("%s: %s", util.Msg["ErrorReadFileError"], err.Error())
		return
	}
	defer file.Close()
	logger.Info.Printf("Receive uploaded file: %s", info.Filename)
	auditMsg.WriteString(info.Filename)

	size := info.Size
	if size == 0 {
		retCode = ApiError.REQUEST_ERROR
		ret = util.Msg["ErrorUploadEmptyFile"]
		return
	}

	buf := make([]byte, size)
	_, err = file.Read(buf)
	if err != nil {
		retCode = ApiError.IO_ERROR
		ret = fmt.Sprintf("%s: %s", util.Msg["ErrorReadFileError"], err.Error())
		return
	}

	scenario, err := ParseUploadSpreadsheet(appID, scenarioString, buf)
	if err != nil {
		retCode = ApiError.REQUEST_ERROR
		ret = fmt.Sprintf("%s: %s", util.Msg["ParseError"], err.Error())
		return
	}

	// save scenario
	content, err := json.Marshal(scenario.EditingContent)
	layout, err := json.Marshal(scenario.EditingLayout)
	logger.Trace.Printf("Save scenario content: %s", string(content))
	retCode, err = UpdateScenario(scenarioID, string(content), string(layout))
	if err != nil {
		retCode = ApiError.DB_ERROR
		ret = fmt.Sprintf("%s: %s", util.Msg["ServerError"], err.Error())
		return
	}
}
