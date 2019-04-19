package CustomChat

import (
	"emotibot.com/emotigo/module/admin-api/util"
	"time"
	"database/sql"
	"fmt"
	"strings"
)

type customChatDaoInterface interface {

	UpdateLatestCustomChatQuestions(appid string, customChatQuestions []*CustomQuestions) (err error)
	GetCustomChatDetail(appid string) (ret []*CustomQuestions, err error)
	UpdateLatestCustomChatExtends(appid string, questions []*Question) (err error)
	GetCustomChatQuestionsWithStatus(appid string) (ret []*CustomQuestions, err error)

}

type customChatDao struct {
	db *sql.DB
}

func (dao *customChatDao) checkDB() {
	if dao.db == nil {
		dao.db = util.GetMainDB()
	}
}

func (dao customChatDao) UpdateLatestCustomChatQuestions(appid string, customQuestions []*CustomQuestions) (err error) {
	defer func() {
		util.ShowError(err)
	}()
	dao.checkDB()
	if dao.db == nil {
		return util.ErrDBNotInit
	}

	tx, err := dao.db.Begin()
	if err != nil {
		return
	}
	defer util.ClearTransition(tx)

	deleteAnswerStr := "DELETE FROM custom_chat_answer WHERE qid in (select id from custom_chat_question where appid = ?)"
	deleteExpandStr := "DELETE FROM custom_chat_extend WHERE qid in (select id from custom_chat_question where appid = ?)"
	deleteQuestionStr := "delete from custom_chat_question where appid = ?"
	_, err = tx.Exec(deleteAnswerStr, appid)
	if err != nil {
		return
	}
	_, err = tx.Exec(deleteExpandStr, appid)
	if err != nil {
		return
	}
	_, err = tx.Exec(deleteQuestionStr, appid)
	if err != nil {
		return
	}

	now := time.Now().Unix()

	err = insertCustomChatQuestion(tx, appid, customQuestions, now)
	if err != nil {
		return
	}
	err = tx.Commit()
	return
}


func insertCustomChatQuestion(tx db, appid string, customQuestions []*CustomQuestions, now int64) (err error) {
	if tx == nil {
		return util.ErrDBNotInit
	}

	insertQuestionStr := `
		INSERT INTO custom_chat_question (appid, category, content, status)
		VALUES (?, ?, ?, ?)`

	insertAnswerStr := `
		INSERT INTO custom_chat_answer (qid, content, status)
		VALUES (?, ?, ?)`

	var result sql.Result
	for _, cq := range customQuestions {
		for _, question := range cq.Questions{
			result, err = tx.Exec(insertQuestionStr, appid, question.Category, question.Content, 1)
			if err != nil {
				return
			}
			question.ID, err = result.LastInsertId()
			if err != nil {
				return
			}

			for _, answer := range question.Answers{
				result, err = tx.Exec(insertAnswerStr, question.ID, answer.Content, 1)
				if err != nil {
					return
				}
			}
		}
	}

	return
}


func (dao customChatDao) UpdateLatestCustomChatExtends(appid string, questions []*Question) (err error) {
	defer func() {
		util.ShowError(err)
	}()
	dao.checkDB()
	if dao.db == nil {
		return util.ErrDBNotInit
	}

	tx, err := dao.db.Begin()
	if err != nil {
		return
	}
	defer util.ClearTransition(tx)

	deleteExpandStr := "DELETE FROM custom_chat_extend WHERE qid in (select id from custom_chat_question where appid = ?)"

	_, err = tx.Exec(deleteExpandStr, appid)
	if err != nil {
		return
	}

	now := time.Now().Unix()

	err = insertCustomChatExtend(tx, appid, questions, now)
	if err != nil {
		return
	}
	err = tx.Commit()
	return
}


func insertCustomChatExtend(tx db, appid string, questions []*Question, now int64) (err error) {
	if tx == nil {
		return util.ErrDBNotInit
	}
	questionsMap := map[string]*Question{}

	for _, question := range questions{

		questionsConditions := []string{"appid = ?", "content = ?"}
		questionsParams := []interface{}{appid, question.Content}

		queryQuestionsStr := fmt.Sprintf("SELECT id, category, content FROM custom_chat_question WHERE %s ", strings.Join(questionsConditions, " AND "))
		questionRows, err := tx.Query(queryQuestionsStr, questionsParams...)
		if err != nil {
			break
		}
		defer questionRows.Close()

		for questionRows.Next() {

			err = questionRows.Scan(&question.ID, &question.Category, &question.Content)
			if err != nil {
				break
			}
			questionsMap[question.Content] = question
		}
	}

	insertExtendStr := `
		INSERT INTO custom_chat_extend (qid, content, status)
		VALUES (?, ?, ?)`

	for _, question := range questionsMap{
		for _, extend := range question.Extends{
			_, err = tx.Exec(insertExtendStr, question.ID, extend.Content, 1)
			if err != nil {
				return
			}
		}
	}
	return
}


func (dao customChatDao) GetCustomChatDetail(appid string) (ret []*CustomQuestions, err error) {
	defer func() {
		util.ShowError(err)
	}()
	dao.checkDB()
	if dao.db == nil {
		return nil, util.ErrDBNotInit
	}

	tx, err := dao.db.Begin()
	if err != nil {
		return
	}
	defer util.ClearTransition(tx)

	customQuestions, err := GetCustomChatQuestions(tx, appid)
	if err != nil {
		return
	}

	err = tx.Commit()
	if err != nil {
		return
	}
	ret = customQuestions
	return
}


func GetCustomChatQuestions(tx db, appid string) (ret []*CustomQuestions, err error) {
	if tx == nil {
		return nil, util.ErrDBNotInit
	}
	customQuestions, err := GetCustomQuestions(tx, appid)
	if err != nil {
		return
	}
	return customQuestions, nil
}


func GetCustomQuestions(tx db, appid string )(ret []*CustomQuestions, err error) {
	if tx == nil {
		return nil, util.ErrDBNotInit
	}

	questionsConditions := []string{"appid = ?", "status = 0"}
	questionsParams := []interface{}{appid}

	queryQuestionsStr := fmt.Sprintf("SELECT id, category, content FROM custom_chat_question WHERE %s ", strings.Join(questionsConditions, " AND "))
	questionRows, err := tx.Query(queryQuestionsStr, questionsParams...)
	if err != nil {
		return
	}
	defer questionRows.Close()
	customQuestionsMap := map[string]*CustomQuestions{}
	questionsMap := map[string]*Question{}

	for questionRows.Next() {
		question := &Question{}
		err = questionRows.Scan(&question.ID, &question.Category, &question.Content)
		if err != nil {
			return
		}
		questionsMap[question.Content] = question
	}

	for _, question := range questionsMap{
		answerConditions := []string{"qid = ?", "status = 0"}
		answerParams := []interface{}{question.ID}

		queryAnswerStr := fmt.Sprintf("SELECT qid, content FROM custom_chat_answer WHERE %s ", strings.Join(answerConditions, " AND "))
		answerRows, err := tx.Query(queryAnswerStr, answerParams...)
		if err != nil {
			break
		}
		defer answerRows.Close()

		for answerRows.Next() {
			answer := &Answer{}
			err = answerRows.Scan(&answer.QID, &answer.Content)
			if err != nil {
				break
			}
			question.Answers = append(question.Answers, *answer)
		}

		extendConditions := []string{"qid = ?", "status = 0"}
		extendParams := []interface{}{question.ID}

		queryExtendStr := fmt.Sprintf("SELECT qid, content FROM custom_chat_extend WHERE %s ", strings.Join(extendConditions, " AND "))
		extendRows, err := tx.Query(queryExtendStr, extendParams...)
		if err != nil {
			break
		}
		defer extendRows.Close()

		for extendRows.Next() {
			extend := &Extend{}
			err = extendRows.Scan(&extend.QID, &extend.Content)
			if err != nil {
				break
			}
			question.Extends = append(question.Extends, *extend)
		}
	}

	for k, v := range questionsMap {

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

	for category := range customQuestionsMap {
		ret = append(ret, customQuestionsMap[category])
	}
	return ret, nil
}


func (dao customChatDao) GetCustomChatQuestionsWithStatus(appid string) (ret []*CustomQuestions, err error) {
	defer func() {
		util.ShowError(err)
	}()
	dao.checkDB()
	if dao.db == nil {
		return nil, util.ErrDBNotInit
	}

	tx, err := dao.db.Begin()
	if err != nil {
		return
	}
	defer util.ClearTransition(tx)

	customQuestions, err := GetQuestionsWithStatus(tx, appid)
	if err != nil {
		return
	}

	err = tx.Commit()
	if err != nil {
		return
	}
	ret = customQuestions
	return
}


func GetQuestionsWithStatus(tx db, appid string) (ret []*CustomQuestions, err error) {
	defer func() {
		util.ShowError(err)
	}()

	if tx == nil {
		return nil, util.ErrDBNotInit
	}

	questionsConditions := []string{"appid = ?"}
	questionsParams := []interface{}{appid}

	queryQuestionsStr := fmt.Sprintf("SELECT id, category, content, status FROM custom_chat_question WHERE %s ", strings.Join(questionsConditions, " AND "))
	questionRows, err := tx.Query(queryQuestionsStr, questionsParams...)
	if err != nil {
		return
	}
	defer questionRows.Close()
	customQuestionsMap := map[string]*CustomQuestions{}
	questionsMap := map[string]*Question{}

	for questionRows.Next() {
		question := &Question{}
		err = questionRows.Scan(&question.ID, &question.Category, &question.Content, &question.Status)
		if err != nil {
			return
		}
		questionsMap[question.Content] = question
	}

	for _, question := range questionsMap{
		answerConditions := []string{"qid = ?"}
		answerParams := []interface{}{question.ID}

		queryAnswerStr := fmt.Sprintf("SELECT id, qid, content, status FROM custom_chat_answer WHERE %s ", strings.Join(answerConditions, " AND "))
		answerRows, err := tx.Query(queryAnswerStr, answerParams...)
		if err != nil {
			break
		}
		defer answerRows.Close()

		for answerRows.Next() {
			answer := &Answer{}
			err = answerRows.Scan(&answer.ID, &answer.QID, &answer.Content, &answer.Status)
			if err != nil {
				break
			}
			question.Answers = append(question.Answers, *answer)
		}

		extendConditions := []string{"qid = ?"}
		extendParams := []interface{}{question.ID}

		queryExtendStr := fmt.Sprintf("SELECT id, qid, content, status FROM custom_chat_extend WHERE %s ", strings.Join(extendConditions, " AND "))
		extendRows, err := tx.Query(queryExtendStr, extendParams...)
		if err != nil {
			break
		}
		defer extendRows.Close()

		for extendRows.Next() {
			extend := &Extend{}
			err = extendRows.Scan(&extend.ID, &extend.QID, &extend.Content, &extend.Status)
			if err != nil {
				break
			}

			question.Extends = append(question.Extends, *extend)
		}
	}

	for k, v := range questionsMap {

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

	for category := range customQuestionsMap {
		ret = append(ret, customQuestionsMap[category])
	}
	return ret, nil
}

func UpdateCustomChatStatus(customQuestions []*CustomQuestions) (err error) {
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
	var qIDList  []interface{}

	for _, cq := range customQuestions {
		for _, q := range cq.Questions {
			qIDList = append(qIDList, q.ID)
		}
	}

	if len(qIDList) > 0 {
		updateQuestionStatusStr := fmt.Sprintf(`
			UPDATE custom_chat_question SET status = 0
			WHERE id in (?%s)`, strings.Repeat(",?", len(qIDList)-1))
		_, err = t.Exec(updateQuestionStatusStr, qIDList...)
		if err != nil {
			return
		}

		updateAnswerStatusStr := fmt.Sprintf(`
			UPDATE custom_chat_answer SET status = 0
			WHERE qid in (?%s)`, strings.Repeat(",?", len(qIDList)-1))
		_, err = t.Exec(updateAnswerStatusStr, qIDList...)
		if err != nil {
			return
		}

		updateExtendStatusStr := fmt.Sprintf(`
			UPDATE custom_chat_extend SET status = 0
			WHERE qid in (?%s)`, strings.Repeat(",?", len(qIDList)-1))
		_, err = t.Exec(updateExtendStatusStr, qIDList...)
		if err != nil {
			return
		}
	}

	err = t.Commit()
	return
}

type db interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Exec(query string, args ...interface{}) (sql.Result, error)
}