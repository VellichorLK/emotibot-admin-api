package CustomChat

import (
	"github.com/tealeg/xlsx"
	"emotibot.com/emotigo/module/admin-api/util/localemsg"
	"emotibot.com/emotigo/pkg/logger"
	"fmt"
	"strings"
	"emotibot.com/emotigo/module/admin-api/util/AdminErrors"
	"bytes"
	"bufio"
	"emotibot.com/emotigo/module/admin-api/util"
	"time"
	"database/sql"
	"strconv"
	"emotibot.com/emotigo/module/admin-api/util/zhconverter"
	"encoding/json"
	"emotibot.com/emotigo/module/admin-api/Service"
	"unicode/utf8"
)

func ParseImportQuestionFile(buf []byte, locale string) (customQuestions []*CustomQuestions, err error) {
	file, err := xlsx.OpenBinary(buf)
	if err != nil {
		return nil, err
	}
	hasContent := false

	sheets := file.Sheets
	if len(sheets) <= 0 {
		return nil, AdminErrors.New(AdminErrors.ErrnoRequestError,
			localemsg.Get(locale, "CustomChatUploadSheetErr"))
	}

	for idx := range sheets {
		if sheets[idx].Name == localemsg.Get(locale, "CustomChatQuestionSheetName"){
			hasContent = true
			break
		}
	}

	if hasContent != true {
		return nil, AdminErrors.New(AdminErrors.ErrnoRequestError,
			localemsg.Get(locale, "CustomChatUploadSheetErr"))
	}

	return parseCustomChatQuestionSheets(sheets, locale)
}

func parseCustomChatQuestionSheets(sheets []*xlsx.Sheet, locale string) (customQuestions []*CustomQuestions, err error) {
	customQuestionsMap := map[string]*CustomQuestions{}
	questionsMap := map[string]*Question{}

	for idx := range sheets {

		if sheets[idx].Name == localemsg.Get(locale, "CustomChatQuestionSheetName") {
			rows := sheets[idx].Rows
			if len(rows) == 0 {
				return nil, AdminErrors.New(AdminErrors.ErrnoRequestError,
					fmt.Sprintf(localemsg.Get(locale, "CustomChatUploadNoHeaderTpl"), sheets[idx].Name))
			}
			categoryIdx, questionIdx, answerIdx := getQuestionColumnIdx(rows[0], locale)
			if categoryIdx < 0 || categoryIdx > 2 || questionIdx < 0 || questionIdx > 2 || answerIdx < 0 || answerIdx > 2 {
				return nil,  AdminErrors.New(AdminErrors.ErrnoRequestError,
					fmt.Sprintf(localemsg.Get(locale, "CustomChatUploadNoHeaderTpl"), sheets[idx].Name))
			}

			rows = rows[1:]
			for rowIdx := range rows {
				cells := rows[rowIdx].Cells
				if len(cells) < 2 {
					return nil, AdminErrors.New(AdminErrors.ErrnoRequestError,
						fmt.Sprintf(localemsg.Get(locale, "CustomChatUploadQuestionRowInvalidTpl"), sheets[idx].Name, rowIdx+1))
				}
				category := strings.TrimSpace(cells[categoryIdx].String())
				question := strings.TrimSpace(cells[questionIdx].String())
				answer := strings.TrimSpace(cells[answerIdx].String())

				qLength := utf8.RuneCountInString(question)
				if qLength > 50{
					return nil, AdminErrors.New(AdminErrors.ErrnoRequestError,
						fmt.Sprintf(localemsg.Get(locale, "CustomChatUploadQuestionExceedLimit"), sheets[idx].Name, rowIdx+1, 50))

				}
				aLength := utf8.RuneCountInString(answer)
				if aLength > 1500{
					return nil, AdminErrors.New(AdminErrors.ErrnoRequestError,
						fmt.Sprintf(localemsg.Get(locale, "CustomChatUploadQuestionExceedLimit"), sheets[idx].Name, rowIdx+1, 1500))
				}

				if question == "" && answer == "" {
					continue
				}
				if question == "" {
					return nil, AdminErrors.New(AdminErrors.ErrnoRequestError,
						fmt.Sprintf(localemsg.Get(locale, "CustomChatUploadQuestionRowNoQuestionTpl"), sheets[idx].Name, rowIdx+1))

				}
				if answer == "" {
					return nil, AdminErrors.New(AdminErrors.ErrnoRequestError,
						fmt.Sprintf(localemsg.Get(locale, "CustomChatUploadQuestionRowNoAnswerTpl"), sheets[idx].Name, rowIdx+1))
				}
				if _, ok := questionsMap[question]; !ok {
					questionEnt := Question{}

					questionEnt.Content = question
					questionEnt.Answers = []Answer{}
					questionEnt.AnswerCount = 0
					questionEnt.Category = category
					questionsMap[question] = &questionEnt

				}
				ans := Answer{}
				ans.Content =answer
				answerList := append(questionsMap[question].Answers, ans)
				questionsMap[question].Category = category
				questionsMap[question].Answers = answerList
				questionsMap[question].AnswerCount++

			}

			for k,v := range questionsMap {

				if _, ok := customQuestionsMap[v.Category]; !ok {
					customQuestionsEnt := CustomQuestions{}

					customQuestionsEnt.Category = v.Category
					customQuestionsEnt.Questions = []Question{}
					customQuestionsEnt.QuestionCount = 0

					customQuestionsMap[v.Category] = &customQuestionsEnt
				}

				customQuestionsMap[v.Category].Questions = append(customQuestionsMap[v.Category].Questions, *questionsMap[k])
				customQuestionsMap[v.Category].QuestionCount = len(customQuestionsMap[v.Category].Questions)
			}

		} else {
			continue
		}
	}

	customQuestions = []*CustomQuestions{}
	for category := range customQuestionsMap {
		customQuestions = append(customQuestions, customQuestionsMap[category])
	}
	return
}

func UpdateLatestCustomChatQuestions(appid string, customChatQuestions []*CustomQuestions) AdminErrors.AdminError {
	dbErr := dao.UpdateLatestCustomChatQuestions(appid, customChatQuestions)
	if dbErr != nil {
		return AdminErrors.New(AdminErrors.ErrnoDBError, dbErr.Error())
	}
	return nil
}

func UpdateLatestCustomChatExtends(appid string, questions []*Question) AdminErrors.AdminError {
	dbErr := dao.UpdateLatestCustomChatExtends(appid, questions)
	if dbErr != nil {
		return AdminErrors.New(AdminErrors.ErrnoDBError, dbErr.Error())
	}
	return nil
}


func ParseImportExtendFile(buf []byte, locale string) (extends []*Question, err error) {
	file, err := xlsx.OpenBinary(buf)
	if err != nil {
		return nil, err
	}
	hasContent := false

	sheets := file.Sheets
	if len(sheets) <= 0 {
		return nil, AdminErrors.New(AdminErrors.ErrnoRequestError,
			fmt.Sprintf(localemsg.Get(locale, "CustomChatUploadSheetErr")))
	}

	for idx := range sheets {
		if sheets[idx].Name == localemsg.Get(locale, "CustomChatExtendSheetName"){
			hasContent = true
			break
		}
	}

	if hasContent != true {
		return nil, AdminErrors.New(AdminErrors.ErrnoRequestError,
			fmt.Sprintf(localemsg.Get(locale, "CustomChatUploadSheetErr")))
	}

	return parseCustomChatExtendSheets(sheets, locale)
}

func parseCustomChatExtendSheets(sheets []*xlsx.Sheet, locale string) (questions []*Question, err error) {
	//questionsArray := map[string]*Question{}
	questionsMap := map[string]*Question{}

	for idx := range sheets {

		if sheets[idx].Name == localemsg.Get(locale, "CustomChatExtendSheetName") {
			rows := sheets[idx].Rows
			if len(rows) == 0 {
				return nil, AdminErrors.New(AdminErrors.ErrnoRequestError,
					fmt.Sprintf(localemsg.Get(locale, "CustomChatUploadNoHeaderTpl"), sheets[idx].Name))
			}
			questionIdx, extendIdx := getExtendColumnIdx(rows[0], locale)
			if questionIdx < 0 || questionIdx > 1 || extendIdx < 0 || extendIdx > 1 {
				return nil, AdminErrors.New(AdminErrors.ErrnoRequestError,
					fmt.Sprintf(localemsg.Get(locale, "CustomChatUploadNoHeaderTpl"), sheets[idx].Name))
			}

			rows = rows[1:]
			for rowIdx := range rows {
				cells := rows[rowIdx].Cells
				if len(cells) < 2 {
					return nil, AdminErrors.New(AdminErrors.ErrnoRequestError,
						fmt.Sprintf(localemsg.Get(locale, "CustomChatUploadQuestionRowInvalidTpl"), sheets[idx].Name, rowIdx+1))
				}

				question := strings.TrimSpace(cells[questionIdx].String())
				extend := strings.TrimSpace(cells[extendIdx].String())
				extendLength := utf8.RuneCountInString(extend)
				if extendLength > 50{
					return nil, AdminErrors.New(AdminErrors.ErrnoRequestError,
						fmt.Sprintf(localemsg.Get(locale, "CustomChatUploadQuestionExceedLimit"), sheets[idx].Name, rowIdx+1, 50))
				}

				if question == "" && extend == "" {
					continue
				}
				if question == "" {
					return nil, AdminErrors.New(AdminErrors.ErrnoRequestError,
						fmt.Sprintf(localemsg.Get(locale, "CustomChatUploadQuestionRowNoQuestionTpl"), sheets[idx].Name, rowIdx+1))
				}
				if extend == "" {
					return nil, AdminErrors.New(AdminErrors.ErrnoRequestError,
						fmt.Sprintf(localemsg.Get(locale, "CustomChatUploadQuestionRowNoExtendTpl"), sheets[idx].Name, rowIdx+1))
				}


				if _, ok := questionsMap[question]; !ok {
					questionEnt := Question{}

					questionEnt.Content = question
					questionEnt.Extends = []Extend{}
					questionEnt.ExtendCount = 0
					questionsMap[question] = &questionEnt

				}
				ent := Extend{}
				ent.Content = extend
				extendList := append(questionsMap[question].Extends, ent)
				questionsMap[question].Extends = extendList
				questionsMap[question].ExtendCount++
			}

		} else {
			continue
		}

	}

	questionsArray := []*Question{}
	for _, q := range questionsMap {
		questionsArray = append(questionsArray, q)
	}
	return questionsArray, nil
}

func GetExportCustomChat(appid string, locale string) (ret []byte, err AdminErrors.AdminError) {
	customQuestions, daoErr := dao.GetCustomChatDetail(appid)
	if daoErr != nil {
		return
	}

	file := xlsx.NewFile()
	sheetQuestion, xlsxErr := file.AddSheet(localemsg.Get(locale, "CustomChatQuestionSheetName"))
	if xlsxErr != nil {
		err = AdminErrors.New(AdminErrors.ErrnoIOError, xlsxErr.Error())
		return
	}
	sheetExtend, xlsxErr := file.AddSheet(localemsg.Get(locale, "CustomChatExtendSheetName"))
	if xlsxErr != nil {
		err = AdminErrors.New(AdminErrors.ErrnoIOError, xlsxErr.Error())
		return
	}
	sheets := []*xlsx.Sheet{sheetQuestion, sheetExtend}

	for _, sheet := range sheets {
		if sheet.Name == localemsg.Get(locale, "CustomChatQuestionSheetName"){
			headerRow := sheet.AddRow()
			headerRow.AddCell().SetString(localemsg.Get(locale, "CustomChatCategory"))
			headerRow.AddCell().SetString(localemsg.Get(locale, "CustomChatQuestion"))
			headerRow.AddCell().SetString(localemsg.Get(locale, "CustomChatAnswer"))
		}else if sheet.Name == localemsg.Get(locale, "CustomChatExtendSheetName"){
			headerRow := sheet.AddRow()
			headerRow.AddCell().SetString(localemsg.Get(locale, "CustomChatQuestion"))
			headerRow.AddCell().SetString(localemsg.Get(locale, "CustomChatExtend"))
		}
	}

	for _, cq := range customQuestions{
		for _, question := range cq.Questions{
			for _, answer := range question.Answers{
				row := sheetQuestion.AddRow()
				row.AddCell().SetString(cq.Category)
				row.AddCell().SetString(question.Content)
				row.AddCell().SetString(answer.Content)
			}
			for _, extend := range question.Extends{
				row := sheetExtend.AddRow()
				row.AddCell().SetString(question.Content)
				row.AddCell().SetString(extend.Content)
			}
		}
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


//func parseCustomChatQuestionSheets(sheets []*xlsx.Sheet, locale string) (customQuestions []*CustomQuestions, err error) {
//	customQuestionsMap := map[string]*CustomQuestions{}
//	questionsMap := map[string]*Question{}
//
//	for idx := range sheets {
//		//var sentenceType int
//		if sheets[idx].Name == localemsg.Get(locale, "CustomChatQuestionSheetName") {
//			rows := sheets[idx].Rows
//			if len(rows) == 0 {
//				return nil, fmt.Errorf(localemsg.Get(locale, "CustomChatUploadNoHeaderTpl"), sheets[idx].Name)
//			}
//			categoryIdx, questionIdx, answerIdx := getQuestionColumnIdx(rows[0], locale)
//			if categoryIdx < 0 || categoryIdx > 2 || questionIdx < 0 || questionIdx > 2 || answerIdx < 0 || answerIdx > 2 {
//				return nil, fmt.Errorf(localemsg.Get(locale, "CustomChatUploadNoHeaderTpl"), sheets[idx].Name)
//			}
//
//
//
//
//
//
//			rows = rows[1:]
//			for rowIdx := range rows {
//				cells := rows[rowIdx].Cells
//				if len(cells) < 2 {
//					return nil, fmt.Errorf(localemsg.Get(locale, "CustomChatUploadQuestionRowInvalidTpl"),
//						sheets[idx].Name, rowIdx+1)
//				}
//				category := strings.TrimSpace(cells[categoryIdx].String())
//				question := strings.TrimSpace(cells[questionIdx].String())
//				answer := strings.TrimSpace(cells[answerIdx].String())
//
//				if question == "" && answer == "" {
//					continue
//				}
//				if question == "" {
//					return nil, fmt.Errorf(localemsg.Get(locale, "CustomChatUploadQuestionRowNoQuestionTpl"),
//						sheets[idx].Name, rowIdx+1)
//				}
//				if answer == "" {
//					return nil, fmt.Errorf(localemsg.Get(locale, "CustomChatUploadQuestionRowNoAnswerTpl"),
//						sheets[idx].Name, rowIdx+1)
//				}
//				if _, ok := questionsMap[question]; !ok {
//					questionEnt := Question{}
//
//					questionEnt.Content = question
//					questionEnt.Answers = []string{}
//					questionEnt.AnswerCount = 0
//					questionsMap[question] = &questionEnt
//
//				}
//				answerList := append(questionsMap[question].Answers, answer)
//
//				questionsMap[question].Answers = answerList
//				questionsMap[question].AnswerCount++
//
//
//
//				if _, ok := customQuestionsMap[category]; !ok {
//					customQuestionsEnt := CustomQuestions{}
//
//					customQuestionsEnt.Category = category
//					customQuestionsEnt.Questions = &([]*Question{})
//					customQuestionsEnt.QuestionCount = 0
//
//					customQuestionsMap[category] = &customQuestionsEnt
//				}
//
//				questionList := append(*customQuestionsMap[category].Questions, questionsMap[question])
//
//				customQuestionsMap[category].Questions = &questionList
//				customQuestionsMap[category].QuestionCount = len(questionList)
//
//			}
//
//
//		} else {
//			continue
//		}
//
//	}
//
//	customQuestions = []*CustomQuestions{}
//	for category := range customQuestionsMap {
//		customQuestions = append(customQuestions, customQuestionsMap[category])
//	}
//	return
//}

func getExtendColumnIdx(row *xlsx.Row, locale string) (questionIdx, extendIdx int) {
	questionIdx, extendIdx = -1, -1
	for idx := range row.Cells {
		cellStr := row.Cells[idx].String()
		if cellStr == localemsg.Get(locale, "CustomChatQuestion"){
			questionIdx = idx
		} else if cellStr == localemsg.Get(locale, "CustomChatExtend"){
			extendIdx = idx
		}
	}
	logger.Trace.Printf("Upload column idx: %d %d\n", questionIdx, extendIdx)
	return
}

func getQuestionColumnIdx(row *xlsx.Row, locale string) (categoryIdx, questionIdx, answerIdx int) {
	categoryIdx, questionIdx, answerIdx = -1, -1, -1
	for idx := range row.Cells {
		cellStr := row.Cells[idx].String()
		if cellStr == localemsg.Get(locale, "CustomChatCategory")  {
			categoryIdx = idx
		} else if cellStr == localemsg.Get(locale, "CustomChatQuestion"){
			questionIdx = idx
		} else if cellStr == localemsg.Get(locale, "CustomChatAnswer"){
			answerIdx = idx
		}
	}
	logger.Trace.Printf("Upload column idx: %d %d %d\n", categoryIdx, questionIdx, answerIdx)
	return
}


func SyncCustomChatToSolr(appid string ,force bool) (err error) {

	return ForceSyncCustomChatToSolr(appid, true)
}


func ForceSyncCustomChatToSolr(appid string, force bool) (err error) {
	restart := false
	body := ""
	defer func() {
		if err != nil {
			logger.Error.Println("Error when sync to solr:", err.Error())
			return
		}
	}()

	if !force {
		var start bool
		var pid int
		start, pid, err = tryStartSyncProcess(syncSolrTimeout)
		if err != nil {
			return
		}
		if !start {
			logger.Info.Println("Pass sync, there is still process running")
			return
		}
		defer func() {
			if r := recover(); r != nil {
				msg := ""
				switch r.(type) {
				case error:
					msg = (r.(error)).Error()
				case string:
					msg = r.(string)
				default:
					msg = fmt.Sprintf("%v", r)
				}
				finishSyncProcess(pid, false, msg)
			} else if err != nil {
				finishSyncProcess(pid, false, err.Error())
			} else {
				finishSyncProcess(pid, true, "")
			}

			restart, err = needProcessCustomChatData(appid)
			if err != nil {
				logger.Error.Println("Check status fail: ", err.Error())
				return
			}
			if restart {
				logger.Trace.Println("Restart sync process")
				time.Sleep(time.Second)
				go ForceSyncCustomChatToSolr(appid, false)
			}
		}()
	}

	var pid int
	_, pid, err = tryStartSyncProcess(syncSolrTimeout)

	customQuestions, err := dao.GetCustomChatQuestionsWithStatus(appid)
	if err != nil {
		return
	}

	questions := []*ChatQuestionTagging{}
	for _, cq := range customQuestions {

		for _, q := range cq.Questions {
			question := ChatQuestionTagging{}
			qID:=strconv.FormatInt(q.ID,10)
			for _, a := range q.Answers {

				aID:=strconv.FormatInt(a.ID,10)
				answer := ChatAnswerTagging{}
				answer.AnswerID = qID + "_" + "0" + "_" + aID
				answer.Answer = a.Content
				answer.Keyword = ""
				answer.Segment = ""
				answer.WordPos = ""
				answer.SentenceType = ""
				question.Answers = append(question.Answers, &answer)
			}
			question.QuestionID = qID + "_" + "0"
			question.Question = q.Content
			question.Keyword = ""
			question.SentenceType = ""
			question.Segment = ""
			question.WordPos = ""
			question.AppID = appid

			questions = append(questions, &question)

			for _, e := range q.Extends{
				extend := ChatQuestionTagging{}
				eID:=strconv.FormatInt(e.ID,10)
				extend.QuestionID = qID + "_" + eID
				extend.Question = e.Content
				extend.Keyword = ""
				extend.SentenceType = ""
				extend.Segment = ""
				extend.WordPos = ""
				extend.AppID = appid
				extend.Answers = question.Answers
				questions = append(questions, &extend)
			}

		}
	}

	questions = convertQuestionContentWithZhCn(questions)

	if len(questions) > 0 {
		err = fillNLUChatQuestionTaggingInfos(questions)
		if err != nil {
			logger.Error.Printf("Get NLUInfo fail: %s\n", err.Error())
			return
		}

		jsonStr, _ := json.Marshal(questions)

		body, err = Service.DeleteInSolrByFiled("id", appid + "_other*")

		if err != nil {
			logger.Error.Printf("Solr-etl fail, err: %s, response: %s, \n", err.Error(), body)
			return
		}

		logger.Trace.Printf("JSON send to solr: %s\n", jsonStr)
		body, err = Service.IncrementAddSolrByType(jsonStr, "other")
		if err != nil {
			logger.Error.Printf("Solr-etl fail, err: %s, response: %s, \n", err.Error(), body)
			return
		}
	}

	UpdateCustomChatStatus(customQuestions)
	finishSyncProcess(pid, true, "")
	return
}


func convertQuestionContentWithZhCn(questions []*ChatQuestionTagging) []*ChatQuestionTagging {
	if questions == nil {
		return questions
	}
	for _, q := range questions {
		if q == nil {
			continue
		}
		q.Question = zhconverter.T2S(q.Question)
		if q.Answers == nil {
			continue
		}
		for _, answer := range q.Answers {
			answer.Answer = zhconverter.T2S(answer.Answer)
		}
	}
	return questions
}


func tryStartSyncProcess(syncSolrTimeout int) (ret bool, processID int, err error) {
	// this sync status no need to check appid
	defer func() {
		util.ShowError(err)
	}()
	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = util.ErrDBNotInit
		return
	}

	t, err := mySQL.Begin()
	if err != nil {
		return
	}
	defer util.ClearTransition(t)

	var start int
	var running bool
	var status = ""

	queryStr := `
		SELECT UNIX_TIMESTAMP(start_at), status FROM process_status
		WHERE module = 'custom-chat'
		ORDER BY id desc limit 1
	`
	row := t.QueryRow(queryStr)
	err = row.Scan(&start, &status)
	if err == sql.ErrNoRows {
		running, err = false, nil
	} else if err != nil {
		return
	} else {
		running, err = status == "running", nil
	}

	now := time.Now().Unix()
	if running {
		logger.Trace.Printf("Previous still running from %d", start)
		if int(now)-start <= syncSolrTimeout {
			return
		}
	}

	queryStr = `
		INSERT INTO process_status
		(app_id, module, status) VALUES ('', 'custom-chat', 'running')
	`
	result, err := t.Exec(queryStr)
	if err != nil {
		return
	}

	id64, err := result.LastInsertId()
	if err != nil {
		return
	}

	processID = int(id64)
	err = t.Commit()
	if err != nil {
		return
	}

	ret = true
	return
}


func finishSyncProcess(pid int, result bool, msg string) (err error) {
	// this sync status no need to check appid
	defer func() {
		util.ShowError(err)
	}()
	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = util.ErrDBNotInit
		return
	}

	status := "success"
	if !result {
		status = "fail"
	}

	queryStr := `
		UPDATE process_status
		SET status = ?, message = ?
		WHERE id = ?`
	_, err = mySQL.Exec(queryStr, status, msg, pid)
	return
}


func needProcessCustomChatData(appid string) (ret bool, err error) {
	ret = false
	//defer func() {
	//	util.ShowError(err)
	//}()
	//mySQL := util.GetMainDB()
	//if mySQL == nil {
	//	err = util.ErrDBNotInit
	//	return
	//}
	//
	//t, err := mySQL.Begin()
	//if err != nil {
	//	return
	//}
	//defer util.ClearTransition(t)
	//
	//
	//
	//err = t.Commit()
	return
}

func fillNLUChatQuestionTaggingInfos(questions []*ChatQuestionTagging) error {
	sentences := make([]string, 0, len(questions))
	for _, q := range questions {
		sentences = append(sentences, q.Question)
	}

	sentenceResult, err := Service.BatchGetNLUResults("", sentences)
	if err != nil {
		return err
	}

	for _, q := range questions {
		if _, ok := sentenceResult[q.Question]; !ok {
			continue
		}
		result := sentenceResult[q.Question]
		q.Segment = result.Segment.ToString()
		q.WordPos = result.Segment.ToFullString()
		q.Keyword = result.Keyword.ToString()
		q.SentenceType = result.SentenceType
	}

	return nil
}