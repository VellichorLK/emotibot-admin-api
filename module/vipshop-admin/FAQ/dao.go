package FAQ

import (
	"database/sql"
	"errors"
	"fmt"

	"emotibot.com/emotigo/module/vipshop-admin/util"
)

//errorNotFound represent SQL select query fetch zero item
// var errorNotFound = errors.New("items not found")

func selectSimilarQuestions(qID string, appID string) ([]string, error) {
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

	if rows.Next() {
		var content string
		rows.Scan(&content)
		contents = append(contents, content)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("Scanning query failed: %s", err)
	}

	return contents, nil
}

func selectQuestion(qid string, appid string) (string, error) {
	var content string
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return content, errors.New("DB not init")
	}

	queryStr := fmt.Sprintf("SELECT Content from %s_question WHERE Question_Id = ?", appid)
	rows, err := mySQL.Query(queryStr, qid)
	if err != nil {
		util.LogInfo.Printf("error: %s", err.Error())
		return content, fmt.Errorf("SQL Query Failed, err: %s", err)
	}
	defer rows.Close()
	if !rows.Next() {
		return content, nil
	}

	rows.Scan(&content)
	if err = rows.Err(); err != nil {
		return "", fmt.Errorf("SQL scan error: %s", err)
	}

	return content, nil
}

func deleteSimilarQuestionsByQuestionID(t *sql.Tx, qid string, appid string) error {
	queryStr := fmt.Sprintf("DELETE FROM %s_squestion WHERE Question_Id = ?", appid)
	_, err := t.Exec(queryStr, qid)
	if err != nil {
		return fmt.Errorf("DELETE SQL execution failed, %s", err)
	}
	return nil
}

func insertSimilarQuestions(t *sql.Tx, qid string, appid string, user string, sqs []SimilarQuestion) error {

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
