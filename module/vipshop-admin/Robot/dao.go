package Robot

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"emotibot.com/emotigo/module/vipshop-admin/util"
)

func getFunctionList(appid string) (map[string]*FunctionInfo, error) {
	filePath := util.GetFunctionSettingPath(appid)
	ret := make(map[string]*FunctionInfo)

	// If file not exist, return empty map
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		util.LogInfo.Printf("File of function setting not existed for appid = [%s]", filePath)
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
	filePath := util.GetFunctionSettingTmpPath(appid)

	lines := []string{}
	for key, info := range infos {
		lines = append(lines, fmt.Sprintf("%s = %t\n", key, info.Status))
	}
	content := strings.Join(lines, "")
	err := ioutil.WriteFile(filePath, []byte(content), 0644)
	util.LogTrace.Printf("Upload function properties to %s", content)

	now := time.Now()
	nowStr := now.Format("2000-01-01 00:00:00")
	ioutil.WriteFile(filePath, []byte(nowStr), 0644)

	return err
}

func getAllRobotQAList(appid string) ([]*QAInfo, error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return nil, errors.New("DB not init")
	}

	cols := []string{"q_id", "created_at", "content", "content2", "content3", "content4", "content5", "content6", "content7", "content8", "content9", "content10"}
	queryStr := fmt.Sprintf("SELECT %s FROM `%s_robotquestion` ORDER BY q_id", strings.Join(cols, ","), appid)
	rows, err := mySQL.Query(queryStr)
	if err != nil {
		return nil, err
	}

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
			util.LogTrace.Printf("Scan row error: %s", err.Error())
			return nil, err
		}

		info := convertQuestionRowToQAInfo(dest)
		ret = append(ret, info)
		questionMap[info.ID] = info
	}

	// Load all answer and put into corresponded question
	answerMap, err := getAllAnswer(appid)
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

func getAllAnswer(appid string) (map[int][]string, error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return nil, errors.New("DB not init")
	}

	ret := make(map[int][]string)
	queryStr := fmt.Sprintf("SELECT parent_q_id, content FROM `%s_robotanswer`", appid)
	rows, err := mySQL.Query(queryStr)
	if err != nil {
		return nil, err
	}

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

	for i := 2; i < len(row); i++ {
		question := string(*row[i].(*[]byte))
		if len(question) > 0 {
			info.Questions = append(info.Questions, question)
		}
	}
	return &info
}

func getRobotQAListPage(appid string, page int, listPerPage int) ([]*QAInfo, error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return nil, errors.New("DB not init")
	}

	if listPerPage < 0 || page <= 0 {
		return nil, fmt.Errorf("Param error page:%d, list per page: %d", page, listPerPage)
	}

	start := listPerPage * (page - 1)
	cols := []string{"q_id", "created_at", "content", "content2", "content3", "content4", "content5", "content6", "content7", "content8", "content9", "content10"}
	queryStr := fmt.Sprintf("SELECT %s FROM `%s_robotquestion` ORDER BY q_id LIMIT %d OFFSET %d", strings.Join(cols, ","), appid, listPerPage, start)
	util.LogTrace.Printf("Query part of question: %s", queryStr)
	rows, err := mySQL.Query(queryStr)
	if err != nil {
		return nil, err
	}

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
			util.LogTrace.Printf("Scan row error: %s", err.Error())
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
	answerMap, err := getAnswerOfQuestion(appid, ids)
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

func getAnswerOfQuestion(appid string, ids []int) (map[int][]string, error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return nil, errors.New("DB not init")
	}

	if len(ids) == 0 {
		return nil, nil
	}

	ret := make(map[int][]string)
	idsStr := strings.Trim(strings.Replace(fmt.Sprint(ids), " ", ",", -1), "[]")
	queryStr := fmt.Sprintf("SELECT parent_q_id, content FROM `%s_robotanswer` WHERE parent_q_id IN (%s)", appid, idsStr)
	util.LogTrace.Printf("Select part of answer: %s", queryStr)

	rows, err := mySQL.Query(queryStr)
	if err != nil {
		return nil, err
	}

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

func getAllRobotQACnt(appid string) (int, error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return 0, errors.New("DB not init")
	}

	queryStr := fmt.Sprintf("SELECT COUNT(*) AS total FROM %s_robotquestion", appid)
	util.LogTrace.Printf("Query total count: %s", queryStr)
	rows, err := mySQL.Query(queryStr)
	if err != nil {
		return 0, err
	}

	count := 0
	if rows.Next() {
		err := rows.Scan(&count)
		if err != nil {
			return 0, err
		}
	}
	return count, nil
}
