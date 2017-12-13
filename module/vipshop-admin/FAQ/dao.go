package FAQ

import (
	"fmt"
	"time"
	"errors"
	"emotibot.com/emotigo/module/vipshop-admin/util"
)

func findQuestion(qid string, appid string) (bool, error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return false, errors.New("DB not init")
	}

	queryStr := fmt.Sprintf("SELECT Question_Id from %s_question WHERE Question_Id = %s", appid, qid)
	rows, err := mySQL.Query(queryStr)
	if err != nil {
		util.LogInfo.Printf("error: ", err.Error())
		return false, err
	}

	ret := rows.Next()
	if !ret {
		return false, nil
	}

	return true, nil
}

func deleteSimilarQuestionsByQuestionId(qid string, appid string) error {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return errors.New("DB not init")
	}

	queryStr := fmt.Sprintf("DELETE FROM %s_squestion WHERE Question_Id = ?", appid)
	_, err := mySQL.Query(queryStr, qid)
	if err != nil {
		return err
	}

	return nil
}

func insertSimilarQuestions(qid string, appid string, user string, sqs []SimilarQuestion) error {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return errors.New("DB not init")
	}

	// prepare insert sql
	sqlStr := fmt.Sprintf("INSERT INTO %s_squestion(Question_Id, Content, CreatedUser, CreatedTime) VALUES ", appid)
	vals := []interface{}{}

	for _, sq := range sqs {
		sqlStr += "(?, ?, ?, now()),"
		vals = append(vals, qid, sq.Content, user)
	}
	
	//trim the last ,
	sqlStr = sqlStr[0:len(sqlStr)-1]

	//prepare the statement
	stmt, err := mySQL.Prepare(sqlStr)
	if err != nil {
		util.LogInfo.Printf("error: ", err.Error())
	}

	//format all vals at once
	_, err = stmt.Exec(vals...)
	if err != nil {
		util.LogInfo.Printf("error: ", err.Error())
	}

	util.LogInfo.Printf("6")
	return nil
}

