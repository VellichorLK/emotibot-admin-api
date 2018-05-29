package FAQ

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"emotibot.com/emotigo/module/vipshop-admin/util"
)

//RFQuestion is removed Feedback question(移除解決未解決的問題)
type RFQuestion struct {
	Content string `json:"content"`
}

//UpdateRFQUestionsArgs are Post API JSON arguments
type UpdateRFQuestionsArgs struct {
	Contents []string `json:"contents"`
}

// GetRFQuestions return RemoveFeedbackQuestions.
// It need to joined with StdQuestions table, because it need to validate the data.
func GetRFQuestions(appid string) ([]RFQuestion, error) {
	var questions = make([]RFQuestion, 0)
	db := util.GetMainDB()
	if db == nil {
		return nil, fmt.Errorf("main db connection pool is nil")
	}
	rawQuery := fmt.Sprintf("SELECT Question_Content FROM %s_removeFeedbackQuestion ORDER BY id", appid)
	rows, err := db.Query(rawQuery)
	if err != nil {
		return nil, fmt.Errorf("query %s failed, %v", rawQuery, err)
	}
	defer rows.Close()
	for rows.Next() {
		var q RFQuestion
		rows.Scan(&q.Content)
		questions = append(questions, q)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("scan rows failed, %v", err)
	}
	return questions, nil
}

func GetRFQuestionsByCategoryId(appid string, categories []int64) (map[int64][]string, error) {
	fmt.Println(categories)
	db := util.GetMainDB()
	if db == nil {
		return nil, fmt.Errorf("main db connection pool is nil")
	}
	if len(categories) == 0 {
		return nil, fmt.Errorf("categories should have at least one element")
	}
	var rawQuery = "SELECT rf.Question_Content, stdQ.CategoryId FROM " + appid +
		"_removeFeedbackQuestion AS rf JOIN " + appid +
		"_question AS stdQ ON stdQ.CategoryId IN (?" + strings.Repeat(", ?", len(categories)-1) + ") AND stdQ.Content = rf.Question_Content"
	var parameters = make([]interface{}, len(categories))
	for i, c := range categories {
		parameters[i] = c
	}
	rows, err := db.Query(rawQuery, parameters...)
	if err != nil {
		util.LogError.Println("error query:" + rawQuery)
		return nil, fmt.Errorf("query failed, %v", err)
	}
	var results = make(map[int64][]string, 0)
	for rows.Next() {
		var content string
		var categoryID int64
		rows.Scan(&content, &categoryID)
		contents := results[categoryID]
		contents = append(contents, content)
		results[categoryID] = contents
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("sql scan error, %v", err)
	}
	return results, nil
}

func FilterRFQuestions(appid string, q []string) ([]string, error) {
	if len(q) == 0 {
		return []string{}, nil
	}

	query := fmt.Sprintf("SELECT Question_Content FROM %s_removeFeedbackQuestion WHERE Question_Content In (?%s)", appid, strings.Repeat(", ?", len(q)-1))
	db := util.GetMainDB()
	if db == nil {
		return nil, fmt.Errorf("main db not init")
	}
	var parameters = make([]interface{}, len(q))
	for i, content := range q {
		parameters[i] = content
	}
	rows, err := db.Query(query, parameters...)
	if err != nil {
		util.LogError.Printf("query %s failed, %s\n", query, err)
		return nil, fmt.Errorf("select query failed, %v", err)
	}
	var results []string
	for rows.Next() {
		var content string
		rows.Scan(&content)
		results = append(results, content)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("scan io failed, %v", err)
	}

	return results, nil
}

// SetRFQuestions will reset RFQuestion table and save given content as RFQuestion.
// It will try to Update consul as well, if failed, table will be rolled back.
func SetRFQuestions(contents []string, appid string) error {

	db := util.GetMainDB()
	if db == nil {
		return fmt.Errorf("main db connection pool is nil")
	}
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("transaction start failed, %v", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	_, err = tx.Exec(fmt.Sprintf("DELETE FROM %s_removeFeedbackQuestion", appid))
	if err != nil {
		return fmt.Errorf("DELETE RFQuestions Table failed, %v", err)
	}
	var insertStmt *sql.Stmt
	insertStmt, err = tx.Prepare(fmt.Sprintf("INSERT INTO %s_removeFeedbackQuestion(Question_Content) VALUES(?)", appid))
	if err != nil {
		return fmt.Errorf("preparing insert remove feedback question query failed, %v", err)
	}
	defer insertStmt.Close()
	for _, c := range contents {
		_, err = insertStmt.Exec(c)
		if err != nil {
			return fmt.Errorf("insert %s failed, %v", c, err)
		}
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("db commit failed, %v", err)
	}
	unixTime := time.Now().UnixNano() / 1000000
	_, err = util.ConsulUpdateVal("vipshopdata/RFQuestion", unixTime)
	if err != nil {
		return fmt.Errorf("consul update failed, %v", err)
	}

	return nil
}
