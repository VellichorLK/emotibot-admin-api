package FAQ

import (
	"database/sql"
	"fmt"
	"strings"

	"emotibot.com/emotigo/module/vipshop-admin/util"
)

//errorNotFound represent SQL select query fetch zero item
// var errorNotFound = errors.New("items not found")

func selectSimilarQuestions(qID int, appID string) ([]string, error) {
	query := fmt.Sprintf("SELECT Content FROM %s_squestion WHERE Question_Id = ?", appID)
	db := util.GetMainDB()
	if db == nil {
		return nil, fmt.Errorf("DB not init")
	}
	rows, err := db.Query(query, qID)
	if err != nil {
		return nil, fmt.Errorf("query execute failed: %s", err)
	}
	defer rows.Close()
	var contents []string

	for rows.Next() {
		var content string
		rows.Scan(&content)
		contents = append(contents, content)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("Scanning query failed: %s", err)
	}

	return contents, nil
}

//selectQuestion will return StdQuestion struct of the qid, if not found will return sql.ErrNoRows
func selectQuestions(groupID []int, appid string) ([]StdQuestion, error) {
	var questions []StdQuestion
	db := util.GetMainDB()
	if db == nil {
		return nil, fmt.Errorf("Main DB has not init")
	}
	var parameters = make([]interface{}, len(groupID))
	for i, id := range groupID {
		parameters[i] = id
	}
	rawQuery := fmt.Sprintf("SELECT Question_id, Content, CategoryId from %s_question WHERE Question_Id IN (? %s)",
		appid, strings.Repeat(",?", len(groupID)-1))
	result, err := db.Query(rawQuery, parameters...)
	if err != nil {
		return nil, fmt.Errorf("SQL query %s error: %s", rawQuery, err)
	}
	for result.Next() {
		var q StdQuestion
		result.Scan(&q.QuestionID, &q.Content, &q.CategoryID)
		questions = append(questions, q)
	}
	if err := result.Err(); err != nil {
		return nil, err
	}

	return questions, nil
}

func deleteSimilarQuestionsByQuestionID(t *sql.Tx, qid int, appid string) error {
	queryStr := fmt.Sprintf("DELETE FROM %s_squestion WHERE Question_Id = ?", appid)
	_, err := t.Exec(queryStr, qid)
	if err != nil {
		return fmt.Errorf("DELETE SQL execution failed, %s", err)
	}
	return nil
}

func insertSimilarQuestions(t *sql.Tx, qid int, appid string, user string, sqs []SimilarQuestion) error {

	if len(sqs) > 0 {
		// prepare insert sql
		sqlStr := fmt.Sprintf("INSERT INTO %s_squestion(Question_Id, Content, CreatedUser, CreatedTime) VALUES ", appid)
		vals := []interface{}{}

		for _, sq := range sqs {
			sqlStr += "(?, ?, ?, now()),"
			vals = append(vals, qid, sq.Content, user)
		}

		//trim the last ,
		sqlStr = sqlStr[0 : len(sqlStr)-1]

		//prepare the statement
		stmt, err := t.Prepare(sqlStr)
		if err != nil {
			return fmt.Errorf("SQL Prepare err, %s", err)
		}
		defer stmt.Close()

		//format all vals at once
		_, err = stmt.Exec(vals...)
		if err != nil {
			return fmt.Errorf("SQL Execution err, %s", err)
		}
	}

	// hack here, because houta use SQuestion_count to store sq count instead of join similar question table
	// so we have to update SQuestion_count in question table, WTF .....
	// TODO: rewrite query function and left join squestion table
	sqlStr := fmt.Sprintf("UPDATE %s_question SET SQuestion_count = %d, Status = 1 WHERE Question_Id = ?", appid, len(sqs))
	_, err := t.Exec(sqlStr, qid)
	if err != nil {
		return fmt.Errorf("SQL Execution err, %s", err)
	}

	return nil
}

//searchQuestionByContent return standard question based on content given.
//return util.ErrSQLRowNotFound if query is empty
func searchQuestionByContent(content string) (StdQuestion, error) {
	var q StdQuestion
	db := util.GetMainDB()
	if db == nil {
		return q, fmt.Errorf("main db connection pool is nil")
	}
	rawQuery := "SELECT Question_id, Content, CategoryId FROM vipshop_question WHERE Content = ? ORDER BY Question_id DESC"
	results, err := db.Query(rawQuery, content)
	if err != nil {
		return q, fmt.Errorf("sql query %s failed, %v", rawQuery, err)
	}
	defer results.Close()
	if results.Next() {
		results.Scan(&q.QuestionID, &q.Content, &q.CategoryID)
	} else { //404 Not Found
		return q, util.ErrSQLRowNotFound
	}

	if err = results.Err(); err != nil {
		return q, fmt.Errorf("scanning data have failed, %s", err)
	}

	return q, nil

}

// GetCategory will return find Category By ID.
// return error sql.ErrNoRows if category can not be found with given ID
func GetCategory(ID int) (Category, error) {
	db := util.GetMainDB()
	var c Category
	if db == nil {
		return c, fmt.Errorf("main db connection pool is nil")
	}
	rawQuery := "SELECT CategoryId, CategoryName, ParentId FROM vipshop_categories WHERE CategoryId = ?"
	err := db.QueryRow(rawQuery, ID).Scan(&c.ID, &c.Name, &c.ParentID)
	if err == sql.ErrNoRows {
		return c, err
	} else if err != nil {
		return c, fmt.Errorf("query row failed, %v", err)
	}
	return c, nil
}

// GetRFQuestions return RemoveFeedbackQuestions.
// It need to joined with StdQuestions table, because there is no way to make sure data consistency.
func GetRFQuestions() ([]RFQuestion, error) {
	var questions = make([]RFQuestion, 0)
	db := util.GetMainDB()
	if db == nil {
		return nil, fmt.Errorf("main db connection pool is nil")
	}
	rawQuery := "SELECT rf.id, rf.Question_Content FROM vipshop_question AS stdQ INNER JOIN vipshop_removeFeedbackQuestion AS rf ON stdQ.Content = rf.Content"
	rows, err := db.Query(rawQuery)
	if err != nil {
		return nil, fmt.Errorf("query %s failed, %v", rawQuery, err)
	}
	defer rows.Close()
	for rows.Next() {
		var q RFQuestion
		rows.Scan(&q.ID, &q.Content)
		questions = append(questions, q)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("scan rows failed, %v", err)
	}
	return questions, nil
}

//InsertRFQuestions will save content as RFQuestion into DB.
func InsertRFQuestions(contents []string) error {
	db := util.GetMainDB()
	if db == nil {
		return fmt.Errorf("main db connection pool is nil")
	}
	rawQuery := "INSERT INTO vipshop_removeFeedbackQuestion(Question_Content) " + strings.Repeat("VALUES(?) ", len(contents))
	var parameters = make([]interface{}, len(contents))
	for i, c := range contents {
		parameters[i] = c
	}
	_, err := db.Exec(rawQuery, parameters...)
	if err != nil {
		return fmt.Errorf("insert failed, %v", err)
	}

	return nil
}

//GetQuestionsByCategories search all the questions contained in given categories.
func GetQuestionsByCategories(categories []Category) ([]StdQuestion, error) {
	db := util.GetMainDB()
	if db == nil {
		return nil, fmt.Errorf("main db connection pool is nil")
	}
	rawQuery := "SELECT Question_id, Content, CategoryId FROM vipshop_question WHERE CategoryId IN (? " + strings.Repeat(",? ", len(categories)-1) + ")"
	var args = make([]interface{}, len(categories))
	for i, c := range categories {
		args[i] = c.ID
	}
	rows, err := db.Query(rawQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("query question failed, %v", err)
	}
	defer rows.Close()
	var questions []StdQuestion
	for rows.Next() {
		var q StdQuestion
		rows.Scan(&q.QuestionID, &q.Content, &q.CategoryID)
		questions = append(questions, q)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("scan failed, %v", err)
	}

	return questions, nil
}
