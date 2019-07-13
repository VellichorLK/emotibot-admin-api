package BF

import (
	"database/sql"
	"strconv"

	"emotibot.com/emotigo/module/admin-api/ApiError"
	"emotibot.com/emotigo/module/admin-api/util"
	"github.com/tealeg/xlsx"
	"emotibot.com/emotigo/module/admin-api/util/AdminErrors"
	"emotibot.com/emotigo/module/admin-api/util/localemsg"
	"mime/multipart"
	"path"
	"regexp"
	"strings"
	"encoding/json"
	"time"
	"bytes"
	"bufio"
	"fmt"
	"emotibot.com/emotigo/pkg/services/fileservice"
)

func GetCmds(appid string) (*CmdClass, error) {
	return getCmds(appid)
}

func GetCmdsOfLabel(appid string, labelID int) ([]*Cmd, error) {
	return getCmdsOfLabel(appid, labelID)
}

func GetCmd(appid string, id int) (*Cmd, error) {
	return getCmd(appid, id)
}

func DeleteCmd(appid string, id int) error {
	err := deleteCmd(appid, id)
	if err == sql.ErrNoRows {
		return nil
	}

	return err
}

func AddCmd(appid string, cmd *Cmd, cid int) (int, int, error) {
	id, err := addCmd(appid, cmd, cid)
	if err == errDuplicate {
		return 0, ApiError.REQUEST_ERROR, util.GenDuplicatedError(util.Msg["Name"], util.Msg["Cmd"])
	} else if err != nil {
		return 0, ApiError.DB_ERROR, err
	}
	return id, ApiError.SUCCESS, nil
}

func UpdateCmd(appid string, id int, cmd *Cmd) (int, error) {
	err := updateCmd(appid, id, cmd)
	if err == errDuplicate {
		return ApiError.REQUEST_ERROR, util.GenDuplicatedError(util.Msg["Name"], util.Msg["Cmd"])
	} else if err == sql.ErrNoRows {
		return ApiError.NOT_FOUND_ERROR, util.ErrNotFound
	} else if err != nil {
		return ApiError.DB_ERROR, err
	}
	return ApiError.SUCCESS, nil
}

func GetLabelsOfCmd(appid string, cmdID int) ([]*Label, error) {
	labels, err := getLabelsOfCmd(appid, cmdID)
	if err != nil {
		return nil, err
	}
	countMap, err := GetCmdCountOfLabels(appid)
	if err != nil {
		return nil, err
	}
	for _, l := range labels {
		lid, err := strconv.Atoi(l.ID)
		if err != nil {
			return nil, err
		}
		if count, ok := countMap[lid]; ok {
			l.CmdCount = count
		}
	}
	return labels, nil
}

func GetCmdCountOfLabels(appid string) (map[int]int, error) {
	return getCmdCountOfLabels(appid)
}

func GetCmdClass(appid string, classID int) (*CmdClass, int, error) {
	class, err := getCmdClass(appid, classID)
	if err == sql.ErrNoRows {
		return nil, ApiError.NOT_FOUND_ERROR, util.ErrNotFound
	} else if err != nil {
		return nil, ApiError.DB_ERROR, err
	}
	if class == nil {
		return nil, ApiError.NOT_FOUND_ERROR, util.ErrNotFound
	}
	return class, ApiError.SUCCESS, nil
}

func UpdateCmdClass(appid string, classID int, newClassName string) (int, error) {
	err := updateCmdClass(appid, classID, newClassName)
	if err == sql.ErrNoRows {
		return ApiError.NOT_FOUND_ERROR, util.ErrNotFound
	} else if err == errDuplicate {
		return ApiError.REQUEST_ERROR, util.GenDuplicatedError(util.Msg["Name"], util.Msg["CmdClass"])
	} else if err != nil {
		return ApiError.DB_ERROR, err
	}
	return ApiError.SUCCESS, nil
}

func AddCmdClass(appid string, pid *int, className string) (int, int, error) {
	id, err := addCmdClass(appid, pid, className)
	if err == errDuplicate {
		return 0, ApiError.REQUEST_ERROR, util.GenDuplicatedError(util.Msg["Name"], util.Msg["CmdClass"])
	} else if err != nil {
		return 0, ApiError.DB_ERROR, err
	}
	return id, ApiError.SUCCESS, nil
}

func DeleteCmdClass(appid string, classID int) error {
	return deleteCmdClass(appid, classID)
}

func MoveCmd(appid string, id int, cid int) (int, error) {
	err := moveCmd(appid, id, cid)
	if err == errDuplicate {
		return ApiError.REQUEST_ERROR, util.GenDuplicatedError(util.Msg["Name"], util.Msg["Cmd"])
	} else if err == sql.ErrNoRows {
		return ApiError.NOT_FOUND_ERROR, util.ErrNotFound
	} else if err != nil {
		return ApiError.DB_ERROR, err
	}

	return ApiError.SUCCESS, nil
}


func GetFormatCmdByteForExport(appid string, locale string) (ret []byte, err AdminErrors.AdminError) {
	commandRecordList, errFormat := GetFormatCmdForExport(appid, locale)
	if errFormat != nil {
		return
	}

	file := xlsx.NewFile()
	cmdSheet, xlsxErr := file.AddSheet(localemsg.Get(locale, "CmdSheetName"))
	if xlsxErr != nil {
		err = AdminErrors.New(AdminErrors.ErrnoIOError, xlsxErr.Error())
		return
	}

	sheets := []*xlsx.Sheet{cmdSheet}

	for _, sheet := range sheets {
		if sheet.Name == localemsg.Get(locale, "CmdSheetName") {
			headerRow := sheet.AddRow()
			headerRow.AddCell().SetString(headerTitle[locale][0])
			headerRow.AddCell().SetString(headerTitle[locale][1])
			headerRow.AddCell().SetString(headerTitle[locale][2])
			headerRow.AddCell().SetString(headerTitle[locale][3])
			headerRow.AddCell().SetString(headerTitle[locale][4])
			headerRow.AddCell().SetString(headerTitle[locale][5])
			headerRow.AddCell().SetString(headerTitle[locale][6])
			headerRow.AddCell().SetString(headerTitle[locale][7])
			headerRow.AddCell().SetString(headerTitle[locale][8])
		}
	}

	for _, cmd := range commandRecordList {
			row := cmdSheet.AddRow()
			row.AddCell().SetString(cmd.Class)
			row.AddCell().SetString(cmd.Name)
			row.AddCell().SetString(cmd.Target)
			row.AddCell().SetString(cmd.Tags)
			row.AddCell().SetString(cmd.Keywords)
			row.AddCell().SetString(cmd.Regex)
			row.AddCell().SetString(cmd.Period)
			row.AddCell().SetString(cmd.Answer)
			row.AddCell().SetString(cmd.ResponseType)
	}

	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)
	ioErr := file.Write(writer)
	if ioErr != nil {
		err = AdminErrors.New(AdminErrors.ErrnoIOError, ioErr.Error())
		return
	}
	return buf.Bytes(), nil
}



func GetFormatCmdForExport(appid string, locale string) (commandRecordList []*CommandRecord, err error)  {
	commandList := []*CommandRecord{}

	commands, err := getAllCommand(appid)

	labels, err := getSSMLabels(appid)
	labelsMap := make(map[int]string)
	if err == nil {
		for _, label := range labels {
			labelsMap[label.ID] = label.Name
		}
	}

	cmdClass, err := getCmdClassList(appid)
	cmdClassMap := make(map[int]string)

	if err == nil {
		for _, cls := range cmdClass {
			cmdClassMap[cls.ID] = cls.Name
		}
	}

	for _, command := range commands {
		commandRecord := CommandRecord{}
		if command.ClassId != -1 {
			className := cmdClassMap[command.ClassId]
			commandRecord.Class = className
		}
		commandRecord.Name = command.Name
		commandRecord.Target = convertFlagToTarget(command.Target, locale)
		commandRecord.Tags = convertIdToTagsString(labelsMap, command.Tags)
		keywords, regex := splitRuleToKeywordsAndRegex(command.Rule)
		commandRecord.Keywords = keywords
		commandRecord.Regex = regex
		commandRecord.ResponseType = convertFlagToResponseType(command.ResponseType, locale)
		commandRecord.Answer = command.Answer
		commandRecord.Period = combineBeginAndEndTimeToPeriod(command.BeginTime, command.EndTime)
		commandList = append(commandList, &commandRecord)
	}

	return commandList, nil
}





func GetCmdImportProcess(recordId int) (process int,  error AdminErrors.AdminError){

	ret, err := getCommandImportProgress(recordId)

	if err != nil {
		return 0, AdminErrors.New(AdminErrors.ErrnoAPIError, err.Error())
	}

	return ret, nil
}

func GetCmdImportReport(recordId int) (ret []byte, name string,  error AdminErrors.AdminError){

	path, err := getCommandImportReportPath(recordId)


	if err != nil {
		return nil, path, AdminErrors.New(AdminErrors.ErrnoAPIError, err.Error())
	}

	buf, err := fileservice.GetFile(CmdFileMinioNamespace, path)
	if err != nil {
		return nil, path, AdminErrors.New(AdminErrors.ErrnoAPIError, err.Error())
	}
	return buf, path, nil
}

func ProcessImportCmdFile(appid string, userid string, buf []byte, info *multipart.FileHeader, locale string) (ret int, recordId int, err error) {

	statusCode := 0
	file, err := xlsx.OpenBinary(buf)
	if err != nil {
		return StatusFileOpenError, -1, err
	}

	statusCode, err = checkTemplateFormat(file, info, locale)

	if err != nil || statusCode != StatusOk {
		return statusCode, -1 , err
	}

	statusCode, err = checkFileContent(file, locale)

	if err != nil || statusCode != StatusOk {
		return statusCode, -1 , err
	}

	commands, valid, parseErr := parseCommandFromFile(file, locale)

	if parseErr != nil {
		err = AdminErrors.New(AdminErrors.ErrnoRequestError, parseErr.Error())
		return
	}

	record, err := addCmdImportHistoryRecord(appid, userid, info.Filename, len(commands), valid)

	go processImportToDb(appid, commands, locale, record)

	return statusCode, record, nil
}

func getXlsxSheetsRows(sheet *xlsx.Sheet) (ret int) {
	if sheet == nil{
		return -1
	}
	rows := sheet.Rows

	return len(rows)
}

func checkTemplateFormat(file *xlsx.File, info *multipart.FileHeader, locale string) (ret int, err error) {
	contentSheetName := localemsg.Get(locale, "CmdSheetName")
	contentSheet := file.Sheet[contentSheetName]

	if path.Ext(info.Filename) != ".xlsx" {

		return StatusFileFormat, AdminErrors.New(AdminErrors.ErrnoRequestError,
			localemsg.Get(locale, "CmdUploadSheetErr"))
	}

	if contentSheet == nil {
		return StatusFileFormat, AdminErrors.New(AdminErrors.ErrnoRequestError,
			localemsg.Get(locale, "CmdUploadSheetErr"))
	}

	headerRow := contentSheet.Row(0)

	if len(headerRow.Cells) != len(headerTitle[locale]){
		return StatusTemplateFormat, nil
	} else {
		for idx, cell := range headerRow.Cells {
			if cell.Value != headerTitle[locale][idx]{
				return StatusTemplateFormat, nil
			}
		}
	}

	if info.Size > MaxFileSize {
		return StatusSizeExceed, nil
	}

	return StatusOk, nil
}


func checkFileContent(file *xlsx.File, locale string) (ret int, err error){

	contentSheetName := localemsg.Get(locale, "CmdSheetName")
	contentSheet := file.Sheet[contentSheetName]

	if contentSheet != nil{
		item := getXlsxSheetsRows(contentSheet)
		if item > MaxFileLength + 1 {
			return StatusRowsExceed, AdminErrors.New(AdminErrors.ErrnoRequestError,
				localemsg.Get(locale, "CmdUploadSheetErr"))
		}
	}

	return StatusOk, nil
}

func parseCommandFromFile(file *xlsx.File, locale string) (commandRecords []*CommandRecord, valid int, err error)  {

	contentSheetName := localemsg.Get(locale, "CmdSheetName")
	contentSheet := file.Sheet[contentSheetName]

	validNum := 0

	for i, row := range contentSheet.Rows {
		if i == 0 {
			continue
		}
		commandRecord := CommandRecord{}
		//commandRecord.Class = row.Cells[]
		//if len(row.Cells) != len(headerTitle[locale]){
		//	return StatusTemplateFormat, nil
		//}

		status := checkStatus(row, locale)

		if status.Status == 0 {
			validNum ++
		}

		commandRecord.Class = row.Cells[0].Value
		commandRecord.Name = row.Cells[1].Value
		commandRecord.Target = row.Cells[2].Value
		commandRecord.Tags = row.Cells[3].Value
		commandRecord.Keywords = row.Cells[4].Value
		commandRecord.Regex = row.Cells[5].Value
		commandRecord.Period = row.Cells[6].Value
		commandRecord.Answer = row.Cells[7].Value
		commandRecord.ResponseType = row.Cells[8].Value
		commandRecord.CheckStatus = status

		commandRecords = append(commandRecords, &commandRecord)
	}
	return commandRecords, validNum, nil
}

func checkStatus(row *xlsx.Row, locale string) (RecordStatus) {

	status := RecordStatus{}
	cls := row.Cells[0].Value
	if len(cls) > 20 {
		status.Status = RecordStatusClassExceedMax
		status.Content = localemsg.Get(locale, "CmdRecordStatusClassExceedMax")
		return  status
	}
	name := row.Cells[1].Value

	if len(name) == 0 || len(name) > 100{
		status.Status = RecordStatusNameError
		status.Content = localemsg.Get(locale, "CmdRecordStatusNameUnnormal")
		return  status
	}
	target := row.Cells[2].Value

	if !(target == localemsg.Get(locale, "CmdTargetQuestion") ||
		            target == localemsg.Get(locale, "CmdTargetAnswer")){
		status.Status = RecordStatusTargetError
		status.Content = localemsg.Get(locale, "CmdRecordStatusTargetError")
		return  status
	}
	period := row.Cells[6].Value

	regx := regexp.MustCompile(CmdPeriodFormatRegex)

	if len(period) != 0 && !regx.MatchString(period){
		status.Status = RecordStatusPriodError
		status.Content = localemsg.Get(locale, "CmdRecordStatusPriodError")
		return  status
	}

	responseType := row.Cells[8].Value
	if !(responseType == localemsg.Get(locale, "CmdResponseReplace") ||
					responseType == localemsg.Get(locale, "CmdResponseBefore") ||
					responseType == localemsg.Get(locale, "CmdResponseAfter")){

		status.Status = RecordStatusResponseError
		status.Content = localemsg.Get(locale, "CmdRecordStatusResponseTypeError")
		return  status
	}
	status.Status = RecordStatusOk
	status.Content = localemsg.Get(locale, "CmdRecordStatusOK")

	return status
}

func processImportToDb(appid string, commandRecords []*CommandRecord, locale string, recordId int) (err error)  {

	labels, err := getSSMLabels(appid)
	labelsMap := make(map[string]int)
	if err == nil {
		for _, label := range labels {
			labelsMap[label.Name] = label.ID
		}
	}

	cmdClass, err := getCmdClassList(appid)
	cmdClassMap := make(map[string]int)

	if err == nil {
		for _, cls := range cmdClass {
			cmdClassMap[cls.Name] = cls.ID
		}
	}

	err = deleteAllCmd(appid)

	if err != nil {
		return  err
	}

	for idx, command := range commandRecords {
		if command.CheckStatus.Status == 0 {
			commandObj := CommandObj{}

			if _, ok := cmdClassMap[command.Class];!ok{
				classMap, err := saveAndUpdateCmdClass(appid, command.Class, cmdClassMap)
				if err != nil {
					cmdClassMap = classMap
				}
			}
			commandObj.ClassId = cmdClassMap[command.Class]
			commandObj.Target = convertTargetToFlag(command.Target, locale)
			commandObj.Rule = combineKeywordsAndRegexToRule(command.Keywords, command.Regex)
			commandObj.ResponseType = convertResponseTypeToFlag(command.ResponseType, locale)
			commandObj.Name = command.Name
			commandObj.Answer = command.Answer
			commandObj.Tags = convertTagsToId(labelsMap, command.Tags)
			begin, end := convertPeriodToBeginAndEndTime(command.Period)
			commandObj.BeginTime = begin
			commandObj.EndTime = end
			commandObj.Status = 0

			cmd := Cmd{}
			cmd.Name = commandObj.Name
			cmd.Target = CmdTarget(commandObj.Target)
			cmd.Status = commandObj.Status > 0
			cmd.Answer = commandObj.Answer
			cmd.Rule = combineKeywordsAndRegexToRuleObj(command.Keywords, command.Regex)
			cmd.Type = ResponseType(commandObj.ResponseType)
			cmd.Begin = commandObj.BeginTime
			cmd.End = commandObj.EndTime
			cmd.LinkLabel = commandObj.Tags

			_, retCode, err := AddCmd(appid, &cmd, commandObj.ClassId)
			if err != nil  || retCode != 0{
				recordStatus := RecordStatus{}
				recordStatus.Status = retCode
				recordStatus.Content = localemsg.Get(locale, "CmdImportStatusError")
				command.CheckStatus = recordStatus
				return err
			}
		}
		err = updateCmdImportProgress(recordId, idx + 1)

	}

	err = recordImportStatusToFile(appid, commandRecords, locale, recordId)
	return err
}

func recordImportStatusToFile(appid string, commandRecords []*CommandRecord, locale string, recordId int) (err error){

	file := xlsx.NewFile()

	sheetReport, xlsxErr := file.AddSheet(localemsg.Get(locale, "CmdSheetName"))
	if xlsxErr != nil {
		err = AdminErrors.New(AdminErrors.ErrnoIOError, xlsxErr.Error())
		return
	}
	sheets := []*xlsx.Sheet{sheetReport}

	for _, sheet := range sheets {
		if sheet.Name == localemsg.Get(locale, "CmdSheetName") {
			headerRow := sheet.AddRow()
			headerRow.AddCell().SetString(headerTitle[locale][0])
			headerRow.AddCell().SetString(headerTitle[locale][1])
			headerRow.AddCell().SetString(headerTitle[locale][2])
			headerRow.AddCell().SetString(headerTitle[locale][3])
			headerRow.AddCell().SetString(headerTitle[locale][4])
			headerRow.AddCell().SetString(headerTitle[locale][5])
			headerRow.AddCell().SetString(headerTitle[locale][6])
			headerRow.AddCell().SetString(headerTitle[locale][7])
			headerRow.AddCell().SetString(headerTitle[locale][8])
			headerRow.AddCell().SetString(localemsg.Get(locale, "CmdImportErrorReason"))
		}
	}

	for _, cmd := range commandRecords {
		row := sheetReport.AddRow()
		row.AddCell().SetString(cmd.Class)
		row.AddCell().SetString(cmd.Name)
		row.AddCell().SetString(cmd.Target)
		row.AddCell().SetString(cmd.Tags)
		row.AddCell().SetString(cmd.Keywords)
		row.AddCell().SetString(cmd.Regex)
		row.AddCell().SetString(cmd.Period)
		row.AddCell().SetString(cmd.Answer)
		row.AddCell().SetString(cmd.ResponseType)
		row.AddCell().SetString(cmd.CheckStatus.Content)

	}

	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)
	ioErr := file.Write(writer)
	if ioErr != nil {
		err = AdminErrors.New(AdminErrors.ErrnoIOError, ioErr.Error())
		return err
	}
	now := time.Now()
	id := fmt.Sprintf("%s-%s", now.Format("20060102150405"), util.GenRandomString(IDLength))
	path := fmt.Sprintf("%s/%s_report.xlsx", appid, id)
	err = fileservice.AddFile(CmdFileMinioNamespace, path, bytes.NewReader(buf.Bytes()))

	if err != nil {
		return err
	}

	err = updateCmdImportReportFile(recordId, path)

	return err
}

func convertPeriodToBeginAndEndTime(period string)(*time.Time, *time.Time){

	if len(period) == 0 {
		return nil, nil
	}
	nodeArray := strings.Split(period, "-")
	begin, _ := time.Parse("2006/01/02", nodeArray[0])
	end, _ := time.Parse("2006/01/02", nodeArray[1])

	return &begin, &end
}

func combineBeginAndEndTimeToPeriod(begin_time *time.Time, end_time *time.Time  )( string){

	if begin_time == nil || end_time == nil {
		return ""
	}
	begin := begin_time.Format("2006/01/02")
	end := end_time.Format("2006/01/02")
	period  := strings.Join([]string{begin, end}, "-")
	return period
}

func convertTagsToId(tagMap map[string]int, tags string)([]int){
	var ret []int
	tagArray := strings.Split(tags, "/")

	for _, item := range tagArray {
		if _, ok := tagMap[item]; ok {
			ret = append(ret, tagMap[item])
		}
	}
	return ret
}

func convertIdToTagsString(tagMap map[int]string, tags []int)(string){
	var ret string
	var tagArray []string

	for _, item := range tags {
		if _, ok := tagMap[item]; ok {
			tagArray = append(tagArray, tagMap[item])
		}
	}

	ret = strings.Join(tagArray, "/")

	return ret
}

func convertResponseTypeToFlag(responseType string, local string)(int){
	if responseType == localemsg.Get(local, "CmdResponseReplace"){
		return 0
	} else if responseType == localemsg.Get(local, "CmdResponseBefore"){
		return 1
	} else {
		return 2
	}
}

func convertFlagToResponseType(flag int, local string)(string){
	if flag ==  0 {
		return localemsg.Get(local, "CmdResponseReplace")
	} else if flag ==  1 {
		return localemsg.Get(local, "CmdResponseBefore")
	} else {
		return localemsg.Get(local, "CmdResponseAfter")
	}
}

func combineKeywordsAndRegexToRuleObj(keywords string, regex string)([]*CmdContent){
	cmdRule := []*CmdContent{}

	keywordArray := strings.Split(keywords, "/")

	for _, item := range keywordArray {
		rule := CmdContent{}
		rule.Type = "keyword"
		rule.Value = append(rule.Value, item)
		cmdRule = append(cmdRule, &rule)
	}

	if len(regex) > 0 {
		rule := CmdContent{}
		rule.Type = "regex"
		rule.Value = append(rule.Value, regex)
		cmdRule = append(cmdRule, &rule)
	}

	return cmdRule
}


func splitRuleToKeywordsAndRegex(rule string)(string, string){
	cmdRules := []*CmdRule{}

	keywordArray := []string{}
	regexArray := []string{}

	err := json.Unmarshal([]byte(rule), &cmdRules)

	if err != nil {
		err = fmt.Errorf("Invalid rule content: %s", err.Error())
		return "", ""
	}

	for _, rule := range cmdRules {
		if rule.Type == "keyword" {
			for _, item := range rule.Value{
				keywordArray = append(keywordArray, item)
			}
		}

		if rule.Type == "regex" {
			for _, item := range rule.Value{
				regexArray = append(regexArray, item)
			}
		}
	}

	keywordString := strings.Join(keywordArray, "/")
	regexString := strings.Join(regexArray, "/")
	return keywordString, regexString
}


func combineKeywordsAndRegexToRule(keywords string, regex string)(string){
	cmdRule := []*CmdRule{}

	var ret string

	keywordArray := strings.Split(keywords, "/")

	for _, item := range keywordArray {
		rule := CmdRule{}
		rule.Type = "keyword"
		rule.Value = append(rule.Value, item)
		cmdRule = append(cmdRule, &rule)
	}

	if len(regex) > 0 {
		rule := CmdRule{}
		rule.Type = "regex"
		rule.Value = append(rule.Value, regex)
		cmdRule = append(cmdRule, &rule)
	}

	data, _ := json.Marshal(cmdRule)

	ret = string(data)
	return ret
}

func convertTargetToFlag(target string, local string)(int){
	if target == localemsg.Get(local, "CmdTargetQuestion"){
		return  0
	}else {
		return  1
	}
}
func convertFlagToTarget(targetId int, local string)(string){
	if targetId == 0{
		return  localemsg.Get(local, "CmdTargetQuestion")
	}else {
		return  localemsg.Get(local, "CmdTargetAnswer")
	}
}

func saveAndUpdateCmdClass(appid string, className string, classMap map[string]int)( map[string]int, error){
	var pid *int
	classID, _, err := AddCmdClass(appid, pid, className)
	if err != nil {
		return classMap, err
	}
	classMap[className] = classID

	return classMap, nil
}

func addCmdImportHistoryRecord(appid string, userid string, filename string, rows int, valid_rows int) (id int, err error) {
      return addCmdImportRecord(appid, userid, filename, rows, valid_rows)
}

func updateCmdImportProgress(recordId int, finish_rows int)  (err error){
	return updateCommandImportProgress(recordId, finish_rows)
}
func updateCmdImportReportFile(recordId int, path string)  (err error){
    return updateCommandReportFile(recordId, path)
}

func deleteAllCmd(appid string)  (err error){
     return deleteAllCommand(appid)
}
