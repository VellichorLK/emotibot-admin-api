package FAQ

import (
	"database/sql"
	"fmt"

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
func selectQuestion(qid int, appid string) (StdQuestion, error) {
	var q StdQuestion
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return q, fmt.Errorf("Main DB has not init")
	}

	err := mySQL.QueryRow("SELECT Content, Category_Id from "+appid+"_question WHERE Question_Id = ?", qid).Scan(&q.Content, &q.CategoryID)
	if err == sql.ErrNoRows {
		return q, err
	} else if err != nil {
		return q, fmt.Errorf("SQL query error: %s", err)
	}
	q.QuestionID = qid

	return q, nil
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

func searchQuestionByContent(content string) (StdQuestion, error) {
	var q StdQuestion
	db := util.GetMainDB()
	if db == nil {
		return q, fmt.Errorf("main db connection pool is nil")
	}
	rawQuery := "SELECT Question_id, Content FROM vipshop_question WHERE Content = ? ORDER BY Question_id DESC"
	results, err := db.Query(rawQuery, content)
	if err != nil {
		return q, fmt.Errorf("sql query %s failed, %v", rawQuery, err)
	}
	defer results.Close()
	if results.Next() {
		results.Scan(&q.QuestionID, &q.Content)
	} else { //404 Not Found
		return q, util.ErrSQLRowNotFound
	}

	if err = results.Err(); err != nil {
		return q, fmt.Errorf("scanning data have failed, %s", err)
	}

	return q, nil

}

// GetCategoryFullPath will return full name of category by ID
func GetCategoryFullPath(categoryID int) (string, error) {
	db := util.GetMainDB()
	if db == nil {
		return "", fmt.Errorf("main db connection pool is nil")
	}

	rows, err := db.Query("SELECT CategoryId, CategoryName, ParentId FROM vipshop_categories")
	if err != nil {
		return "", fmt.Errorf("query category table failed, %v", err)
	}
	defer rows.Close()
	var categories = make(map[int]Category)

	for rows.Next() {
		var c Category
		rows.Scan(&c.ID, &c.Name, &c.ParentID)
		categories[c.ID] = c
	}
	if err = rows.Err(); err != nil {
		return "", fmt.Errorf("Rows scaning failed, %v", err)
	}

	if c, ok := categories[categoryID]; ok {
		switch c.ParentID {
		case 0:
			fallthrough
		case -1:
			return "/" + c.Name, nil
		}
		var fullPath string
		for ; ok; c, ok = categories[c.ParentID] {
			fullPath = "/" + c.Name + fullPath
			if c.ParentID == 0 {
				break
			}
		}
		if !ok {
			return "", fmt.Errorf("category id %d has invalid parentID %d", c.ID, c.ParentID)
		}
		return fullPath, nil
	} else {
		return "", fmt.Errorf("Cant find category id %d in db", categoryID)
	}

}
