package Robot

import (
	"database/sql"
	"emotibot.com/emotigo/module/admin-api/util/localemsg"
	"emotibot.com/emotigo/module/admin-api/util/zhconverter"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/pkg/logger"
)

func getFunctionList(appid string) (map[string]*FunctionInfo, error) {
	filePath := util.GetFunctionSettingPath(appid)
	ret := make(map[string]*FunctionInfo)

	// If file not exist, return empty map
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		logger.Info.Printf("File of function setting not existed for appid = [%s]", filePath)
		return ret, nil
	}

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// each line is function_name = on/off pair
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		pair := strings.Split(line, "=")
		if len(pair) == 2 {
			append := FunctionInfo{}
			function := strings.Trim(pair[0], " \"")
			value := strings.Trim(pair[1], " \"")
			if value == "true" || value == "on" || value == "1" {
				append.Status = true
			} else {
				append.Status = false
			}
			ret[function] = &append
		}
	}

	return ret, nil
}

func updateFunctionList(appid string, infos map[string]*FunctionInfo) error {
	filePath := util.GetFunctionSettingPath(appid)

	lines := []string{}
	for key, info := range infos {
		lines = append(lines, fmt.Sprintf("%s = %t\n", key, info.Status))
	}
	content := strings.Join(lines, "")
	err := ioutil.WriteFile(filePath, []byte(content), 0644)
	logger.Trace.Printf("Upload function properties to %s", content)

	tmpFilePath := util.GetFunctionSettingTmpPath(appid)
	now := time.Now()
	nowStr := now.Format("2000-01-01 00:00:00")
	ioutil.WriteFile(tmpFilePath, []byte(nowStr), 0644)

	return err
}

func getAllRobotQAList(appid string, version int) ([]*QAInfo, error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return nil, util.ErrDBNotInit
	}
	var err error
	var rows *sql.Rows
	var queryStr string

	cols := []string{"q_id", "created_at", "content", "content2", "content3", "content4", "content5", "content6", "content7", "content8", "content9", "content10"}
	switch version {
	case 1:
		queryStr = fmt.Sprintf("SELECT %s FROM `%s_robotquestion` WHERE status >= 0 ORDER BY q_id", strings.Join(cols, ","), appid)
		rows, err = mySQL.Query(queryStr)
	case 2:
		queryStr = fmt.Sprintf("SELECT %s FROM `robot_question` WHERE appid = ? AND status >= 0 ORDER BY q_id", strings.Join(cols, ","))
		rows, err = mySQL.Query(queryStr, appid)
	default:
		err = errInvalidVersion
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	dest := make([]interface{}, len(cols))
	rawResult := make([][]byte, len(cols)-2)
	var id int
	var createdTime time.Time

	dest[0] = &id
	dest[1] = &createdTime
	for i := range rawResult {
		// Put pointers to each string in the interface slice
		dest[i+2] = &rawResult[i]
	}

	ret := []*QAInfo{}
	questionMap := make(map[int]*QAInfo)

	// Set all info about question
	for rows.Next() {
		err = rows.Scan(dest...)
		if err != nil {
			logger.Trace.Printf("Scan row error: %s", err.Error())
			return nil, err
		}

		info := convertQuestionRowToQAInfo(dest)
		ret = append(ret, info)
		questionMap[info.ID] = info
	}

	// Load all answer and put into corresponded question
	answerMap, err := getAllAnswer(appid, version)
	if err != nil {
		return nil, err
	}
	for key, arr := range answerMap {
		if qaInfo, ok := questionMap[key]; ok {
			qaInfo.Answers = arr
		}
	}

	return ret, nil
}

func getAllAnswer(appid string, version int) (map[int][]string, error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return nil, util.ErrDBNotInit
	}

	ret := make(map[int][]string)

	var queryStr string
	var rows *sql.Rows
	var err error
	switch version {
	case 1:
		queryStr = fmt.Sprintf("SELECT parent_q_id, content FROM `%s_robotanswer`", appid)
		rows, err = mySQL.Query(queryStr)
	case 2:
		queryStr = "SELECT parent_q_id, content FROM `robot_answer` WHERE appid = ?"
		rows, err = mySQL.Query(queryStr, appid)
	default:
		err = errInvalidVersion
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var qid int
	var content string
	for rows.Next() {
		err = rows.Scan(&qid, &content)
		if err != nil {
			return nil, err
		}
		ret[qid] = append(ret[qid], content)
	}

	return ret, nil
}

func getAnswerOfQuestion(appid string, qid int, version int) ([]string, error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return nil, util.ErrDBNotInit
	}

	ret := []string{}

	var err error
	var rows *sql.Rows
	var queryStr string
	switch version {
	case 1:
		queryStr = fmt.Sprintf("SELECT content FROM `%s_robotanswer` where parent_q_id = ?", appid)
		rows, err = mySQL.Query(queryStr, qid)
	case 2:
		queryStr = "SELECT content FROM `robot_answer` where parent_q_id = ? AND appid = ?"
		rows, err = mySQL.Query(queryStr, qid, appid)
	default:
		err = errInvalidVersion
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var content string
	for rows.Next() {
		err = rows.Scan(&content)
		if err != nil {
			return nil, err
		}
		ret = append(ret, content)
	}

	return ret, nil
}

func getAnswerOfQuestions(appid string, ids []int, version int) (map[int][]string, error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return nil, util.ErrDBNotInit
	}

	if len(ids) == 0 {
		return nil, nil
	}

	ret := make(map[int][]string)

	var queryStr string
	var rows *sql.Rows
	var err error
	idsStr := strings.Trim(strings.Replace(fmt.Sprint(ids), " ", ",", -1), "[]")
	switch version {
	case 1:
		queryStr = fmt.Sprintf("SELECT parent_q_id, content FROM `%s_robotanswer` WHERE parent_q_id IN (%s)", appid, idsStr)
		rows, err = mySQL.Query(queryStr)
	case 2:
		queryStr = fmt.Sprintf("SELECT parent_q_id, content FROM `robot_answer` WHERE parent_q_id IN (%s) AND appid = ?", idsStr)
		rows, err = mySQL.Query(queryStr, appid)
	default:
		err = errInvalidVersion
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var qid int
	var content string
	for rows.Next() {
		err = rows.Scan(&qid, &content)
		if err != nil {
			return nil, err
		}
		ret[qid] = append(ret[qid], content)
	}

	return ret, nil
}

func convertQuestionRowToQAInfo(row []interface{}) *QAInfo {
	info := QAInfo{}

	info.ID = *row[0].(*int)
	info.CreatedTime = *row[1].(*time.Time)
	info.Question = string(*row[2].(*[]byte))
	info.Answers = []string{}

	for i := 3; i < len(row); i++ {
		question := string(*row[i].(*[]byte))
		if len(question) > 0 {
			info.RelatedQuestions = append(info.RelatedQuestions, question)
		}
	}
	return &info
}

func getRobotQAListPage(appid string, page int, listPerPage int, version int) ([]*QAInfo, error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return nil, util.ErrDBNotInit
	}

	if listPerPage < 0 || page <= 0 {
		return nil, fmt.Errorf("Param error page:%d, list per page: %d", page, listPerPage)
	}

	start := listPerPage * (page - 1)
	cols := []string{"q_id", "created_at", "content", "content2", "content3", "content4", "content5", "content6", "content7", "content8", "content9", "content10"}
	var err error
	var rows *sql.Rows
	var queryStr string

	switch version {
	case 1:
		queryStr = fmt.Sprintf(`
			SELECT %s FROM %s_robotquestion
			WHERE status >= 0
			ORDER BY q_id LIMIT %d
			OFFSET %d`, strings.Join(cols, ","), appid, listPerPage, start)
		rows, err = mySQL.Query(queryStr)
	case 2:
		queryStr = fmt.Sprintf(`
			SELECT %s FROM robot_question
			WHERE appid = ? AND status >= 0
			ORDER BY q_id LIMIT %d
			OFFSET %d`, strings.Join(cols, ","), listPerPage, start)
		rows, err = mySQL.Query(queryStr, appid)
	default:
		err = errInvalidVersion
		return nil, err
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// ret := []*QAInfo{}
	dest := make([]interface{}, len(cols))
	rawResult := make([][]byte, len(cols)-2)
	var id int
	var createdTime time.Time

	dest[0] = &id
	dest[1] = &createdTime
	for i := range rawResult {
		// Put pointers to each string in the interface slice
		dest[i+2] = &rawResult[i]
	}

	ret := []*QAInfo{}
	questionMap := make(map[int]*QAInfo)

	// Set all info about question
	for rows.Next() {
		err = rows.Scan(dest...)
		if err != nil {
			logger.Trace.Printf("Scan row error: %s", err.Error())
			return nil, err
		}

		info := convertQuestionRowToQAInfo(dest)
		ret = append(ret, info)
		questionMap[info.ID] = info
	}

	var ids []int
	for id := range questionMap {
		ids = append(ids, id)
	}

	// Load all answer and put into corresponded question
	answerMap, err := getAnswerOfQuestions(appid, ids, version)
	if err != nil {
		return nil, err
	}
	if answerMap != nil {
		for key, arr := range answerMap {
			if qaInfo, ok := questionMap[key]; ok {
				qaInfo.Answers = arr
			}
		}
	}

	return ret, nil
}

func getAllRobotQACnt(appid string, version int) (int, error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return 0, util.ErrDBNotInit
	}
	var err error
	var rows *sql.Rows
	var queryStr string
	switch version {
	case 1:
		queryStr = fmt.Sprintf("SELECT COUNT(*) AS total FROM %s_robotquestion WHERE status >= 0", appid)
		rows, err = mySQL.Query(queryStr)
	case 2:
		queryStr = "SELECT COUNT(*) AS total FROM robot_question WHERE appid = ? AND status >= 0"
		rows, err = mySQL.Query(queryStr, appid)
	default:
		err = errInvalidVersion
		return 0, err
	}

	if err != nil {
		return 0, err
	}
	defer rows.Close()

	count := 0
	if rows.Next() {
		err := rows.Scan(&count)
		if err != nil {
			return 0, err
		}
	}
	return count, nil
}

func getRobotQA(appid string, id int, version int) (*QAInfo, error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return nil, util.ErrDBNotInit
	}

	cols := []string{"q_id", "created_at", "content", "content2", "content3", "content4", "content5", "content6", "content7", "content8", "content9", "content10"}
	var err error
	var rows *sql.Rows
	var queryStr string
	if version == 1 {
		queryStr = fmt.Sprintf("SELECT %s FROM `%s_robotquestion` WHERE q_id = ?", strings.Join(cols, ","), appid)
		rows, err = mySQL.Query(queryStr, id)
	} else if version == 2 {
		queryStr = fmt.Sprintf("SELECT %s FROM `robot_question` WHERE q_id = ? AND appid = ?", strings.Join(cols, ","))
		rows, err = mySQL.Query(queryStr, id, appid)
	} else {
		err = errInvalidVersion
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	dest := make([]interface{}, len(cols))
	rawResult := make([][]byte, len(cols)-2)
	var qid int
	var createdTime time.Time

	dest[0] = &qid
	dest[1] = &createdTime
	for i := range rawResult {
		// Put pointers to each string in the interface slice
		dest[i+2] = &rawResult[i]
	}

	ret := &QAInfo{}

	// Set all info about question
	if rows.Next() {
		err = rows.Scan(dest...)
		if err != nil {
			logger.Trace.Printf("Scan row error: %s", err.Error())
			return nil, err
		}

		ret = convertQuestionRowToQAInfo(dest)
	} else {
		return nil, nil
	}

	// Load all answer and put into corresponded question
	answerList, err := getAnswerOfQuestion(appid, id, version)
	if err != nil {
		return nil, err
	}
	ret.Answers = answerList
	return ret, nil
}

func updateRobotQA(appid string, id int, info *QAInfo) error {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return util.ErrDBNotInit
	}
	emptyStr := ""

	link, err := mySQL.Begin()
	if err != nil {
		return err
	}
	defer link.Commit()

	questionCols := 10
	// Update question table
	queryStr := fmt.Sprintf("UPDATE %s_robotquestion SET content = ?, content2 = ?, content3 = ?, content4 = ?, content5 = ?, content6 = ?, content7 = ?, content8 = ?, content9 = ?, content10 = ?, status = 1, answer_count = ? WHERE q_id = ?", appid)
	args := make([]interface{}, questionCols+2)
	args[0] = info.Question
	for i := 1; i < questionCols; i++ {
		if i-1 < len(info.RelatedQuestions) {
			args[i] = info.RelatedQuestions[i-1]
		} else {
			args[i] = emptyStr
		}
	}
	args[questionCols] = len(info.Answers)
	args[questionCols+1] = id

	_, err = link.Exec(queryStr, args...)
	if err != nil {
		link.Rollback()
		return err
	}

	// Delete orig answer
	queryStr = fmt.Sprintf("DELETE FROM %s_robotanswer WHERE parent_q_id = ?", appid)
	_, err = link.Exec(queryStr, id)
	if err != nil {
		link.Rollback()
		return err
	}

	// Insert new answer
	queryStr = fmt.Sprintf("INSERT INTO %s_robotanswer(parent_q_id, content, user) VALUES (?,?,?)", appid)
	stmt, err := link.Prepare(queryStr)
	if err != nil {
		stmt.Close()
		link.Rollback()
		return err
	}

	for _, answer := range info.Answers {
		_, err = stmt.Exec(id, answer, appid)
		if err != nil {
			stmt.Close()
			link.Rollback()
			return err
		}
	}
	stmt.Close()
	return nil
}

func updateRobotQAV2(appid string, id int, info *QAInfo) error {
	var err error
	defer func() {
		util.ShowError(err)
	}()

	mySQL := util.GetMainDB()
	if mySQL == nil {
		return util.ErrDBNotInit
	}

	link, err := mySQL.Begin()
	if err != nil {
		return err
	}
	defer util.ClearTransition(link)

	// TODO: update question if different from orig, use custom rows of appid itself
	queryStr := "UPDATE robot_question SET status = 1, answer_count = ? WHERE q_id = ?"
	_, err = link.Exec(queryStr, len(info.Answers), id)
	if err != nil {
		return err
	}

	// Delete orig answer
	queryStr = "DELETE FROM robot_answer WHERE parent_q_id = ? AND appid = ?"
	_, err = link.Exec(queryStr, id, appid)
	if err != nil {
		return err
	}

	// Insert new answer
	queryStr = "INSERT INTO robot_answer (appid, parent_q_id, content) VALUES (?, ?,?)"
	stmt, err := link.Prepare(queryStr)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, answer := range info.Answers {
		_, err = stmt.Exec(appid, id, answer)
		if err != nil {
			return err
		}
	}
	err = link.Commit()
	return err
}

func getRobotChat(appid string, id int) (*ChatInfo, error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return nil, util.ErrDBNotInit
	}

	queryStr := fmt.Sprintf("SELECT content FROM %s_robot_setting where type = ?", appid)

	rows, err := mySQL.Query(queryStr, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	contents := []string{}
	for rows.Next() {
		content := ""
		err = rows.Scan(&content)
		if err != nil {
			return nil, err
		}
		contents = append(contents, content)
	}

	ret := &ChatInfo{
		Type:     id,
		Contents: contents,
	}

	return ret, nil
}

func getRobotChatList(appid string) ([]*ChatInfo, error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return nil, util.ErrDBNotInit
	}

	queryStr := fmt.Sprintf("SELECT type, content FROM %s_robot_setting", appid)

	contentMap := make(map[int][]string)
	rows, err := mySQL.Query(queryStr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		chatType := 0
		content := ""
		err = rows.Scan(&chatType, &content)
		if err != nil {
			return nil, err
		}
		contentMap[chatType] = append(contentMap[chatType], content)
	}

	ret := []*ChatInfo{}
	for key, contents := range contentMap {
		ret = append(ret, &ChatInfo{
			Type:     key,
			Contents: contents,
		})
	}

	return ret, nil
}

func getRobotChatInfoList(appid string) ([]*ChatDescription, error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return nil, util.ErrDBNotInit
	}

	nameMap := make(map[int]string)
	commentMap := make(map[int]string)
	queryStr := fmt.Sprintf("SELECT type, name, comment FROM %s_robotwords_type", appid)
	rows, err := mySQL.Query(queryStr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		chatType := 0
		name := ""
		comment := ""
		err = rows.Scan(&chatType, &name, &comment)
		if err != nil {
			return nil, err
		}
		nameMap[chatType] = name
		commentMap[chatType] = comment
	}

	ret := []*ChatDescription{}
	for key, name := range nameMap {
		comment, _ := commentMap[key]

		ret = append(ret, &ChatDescription{
			Type:    key,
			Name:    name,
			Comment: comment,
		})
	}

	return ret, nil
}

func getMultiRobotChat(appid string, input []int) ([]*ChatInfo, error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return nil, util.ErrDBNotInit
	}

	if len(input) == 0 {
		return []*ChatInfo{}, nil
	}

	idsStr := strings.Trim(strings.Replace(fmt.Sprint(input), " ", ",", -1), "[]")
	queryStr := fmt.Sprintf("SELECT type, content FROM %s_robot_setting where type in (%s)", appid, idsStr)
	rows, err := mySQL.Query(queryStr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	contentMap := make(map[int][]string)
	for rows.Next() {
		chatType := 0
		content := ""
		err = rows.Scan(&chatType, &content)
		if err != nil {
			return nil, err
		}
		contentMap[chatType] = append(contentMap[chatType], content)
	}

	ret := []*ChatInfo{}
	for key, contents := range contentMap {
		ret = append(ret, &ChatInfo{
			Type:     key,
			Contents: contents,
		})
	}
	return ret, nil
}

func updateMultiRobotChat(appid string, input []*ChatInfoInput) error {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return util.ErrDBNotInit
	}

	link, err := mySQL.Begin()
	if err != nil {
		return err
	}
	defer link.Commit()

	// Delete old contents
	ids := []interface{}{}
	for _, info := range input {
		ids = append(ids, info.Type)
	}
	deleteStr := fmt.Sprintf("DELETE FROM %s_robot_setting where type in (?%s)", appid, strings.Repeat(",?", len(ids)-1))
	logger.Trace.Printf("SQL: %s\n", deleteStr)
	logger.Trace.Printf("param: %#v\n", ids)
	_, err = link.Exec(deleteStr, ids...)
	if err != nil {
		link.Rollback()
		return err
	}

	// Insert new contents
	insertStr := fmt.Sprintf("INSERT INTO %s_robot_setting(type, content) VALUES(?,?)", appid)
	for _, info := range input {
		for _, content := range info.Contents {
			_, err = link.Exec(insertStr, info.Type, content)
			if err != nil {
				link.Rollback()
				return err
			}
		}
	}

	return nil
}

// initRobotQA only support in version 2
func initRobotQAData(appid string) (err error) {
	defer func() {
		util.ShowError(err)
	}()
	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = util.ErrDBNotInit
		return
	}

	tx, err := mySQL.Begin()
	if err != nil {
		return
	}
	defer util.ClearTransition(tx)

	// 1. check is there has function set
	queryStr := `
		SELECT count(*)
		FROM robot_question
		WHERE appid = ?`
	count := 0
	row := tx.QueryRow(queryStr, appid)
	err = row.Scan(&count)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	// if existed, return
	if count > 0 {
		return nil
	}

	// copy default function to appid
	queryStr = `
		INSERT INTO robot_question
		(content,appid,created_at,answer_count,
			content2,content3,content4,content5,content6,content7,content8,content9,content10,status)
			SELECT content,?,created_at,answer_count,
			content2,content3,content4,content5,content6,content7,content8,content9,content10,1
			FROM robot_question
			WHERE appid = ''`
	_, err = tx.Exec(queryStr, appid)
	if err != nil {
		return
	}
	return tx.Commit()
}

// initWordbankData only support in version 3
func initWordbankData(appid string) (err error) {
	defer func() {
		util.ShowError(err)
	}()
	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = util.ErrDBNotInit
		return
	}

	tx, err := mySQL.Begin()
	if err != nil {
		return
	}
	defer util.ClearTransition(tx)

	// 1. check is there has function set
	queryStr := `
		SELECT count(*)
		FROM entity_class
		WHERE appid = ?`
	count := 0
	row := tx.QueryRow(queryStr, appid)
	err = row.Scan(&count)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	// if existed, return
	if count > 0 {
		return nil
	}

	// copy default function to appid
	queryStr = `
		INSERT INTO entity_class
		(appid, name, editable, intent_engine, rule_engine)
			SELECT ?, name, editable, intent_engine, rule_engine
			FROM entity_class
			WHERE appid = ''`
	_, err = tx.Exec(queryStr, appid)
	if err != nil {
		return
	}
	return tx.Commit()
}

func initPreinstallWords(appid string, locale string) (err error) {
	defer func() {
		util.ShowError(err)
	}()
	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = util.ErrDBNotInit
		return
	}

	tx, err := mySQL.Begin()
	if err != nil {
		return
	}
	defer util.ClearTransition(tx)

	varmap := make(map[int]string)

	var queryStr string
	queryStr = `SELECT type, content FROM robot_words
		WHERE appid=''`

	rows, err := tx.Query(queryStr)

	var typeid int
	var content string

	for rows.Next() {
		err = rows.Scan(&typeid, &content)
		if err != nil {
			return
		}
		if locale == localemsg.ZhTw {
			content = zhconverter.S2T(content)
		}
		varmap[typeid] = content
	}

	for key, value := range varmap {
		queryStr = `
		INSERT INTO robot_words ( appid, content, type)
	SELECT  ?, ?, ?`
		_, err = tx.Exec(queryStr, appid, value, key)
		if err != nil {
			return
		}
	}


	return tx.Commit()
}
