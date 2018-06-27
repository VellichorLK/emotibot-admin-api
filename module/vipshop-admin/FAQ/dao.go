package FAQ

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"emotibot.com/emotigo/module/vipshop-admin/imagesManager"
	"emotibot.com/emotigo/module/vipshop-admin/util"
)

const SEPARATOR = "#SEPARATE_TOKEN#"
const (
	DynamicMenu     = iota
	RelatedQuestion = iota
)

//errorNotFound represent SQL select query fetch zero item
// var errorNotFound = errors.New("items not found")
var tagMap map[int]string

func addApiCategory(appid string, name string, parentID int, level int) (int, error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return 0, errors.New("DB not init")
	}

	queryStr := fmt.Sprintf("INSERT INTO %s_categories (CategoryName, ParentId, Status, level, ParentPath, SelfPath) VALUES(?, ?, 1, ?, '', '')", appid)
	res, err := mySQL.Exec(queryStr, name, parentID, level)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func getQuestionCountInCategories(appid string, IDs []int) (int, error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return 0, errors.New("DB not init")
	} else if len(IDs) == 0 {
		return 0, nil
	}

	queryStr := fmt.Sprintf("SELECT COUNT(*) FROM %s_question WHERE CategoryId in (?"+strings.Repeat(",?", len(IDs)-1)+")", appid)
	args := make([]interface{}, len(IDs))
	for idx := range IDs {
		args[idx] = IDs[idx]
	}
	rows := mySQL.QueryRow(queryStr, args...)
	count := 0
	err := rows.Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func getCategories(appid string) (map[int]*APICategory, error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return nil, errors.New("DB not init")
	}

	queryStr := fmt.Sprintf("SELECT CategoryId, CategoryName, ParentId FROM `%s_categories` where Status > 0 and CategoryId > 0", appid)
	rows, err := mySQL.Query(queryStr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ret []*APICategory
	for rows.Next() {
		temp := APICategory{}
		err = rows.Scan(&temp.ID, &temp.Name, &temp.ParentID)
		temp.Children = make([]*APICategory, 0)
		if err != nil {
			return nil, err
		}
		ret = append(ret, &temp)
	}

	categoryMap := make(map[int]*APICategory)
	for _, category := range ret {
		categoryMap[category.ID] = category
	}

	return categoryMap, nil
}

func deleteCategories(appid string, IDs []int) error {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return errors.New("DB not init")
	} else if len(IDs) == 0 {
		return nil
	}

	queryStr := fmt.Sprintf("DELETE FROM %s_categories WHERE CategoryId in (?"+strings.Repeat(",?", len(IDs)-1)+")", appid)
	args := make([]interface{}, len(IDs))
	for idx := range IDs {
		args[idx] = IDs[idx]
	}
	_, err := mySQL.Exec(queryStr, args...)
	if err != nil {
		return err
	}

	return nil
}

func disableQuestionInCategories(appid string, IDs []int) error {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return errors.New("DB not init")
	} else if len(IDs) == 0 {
		return nil
	}

	queryStr := fmt.Sprintf("UPDATE %s_question SET Status = -1 WHERE CategoryId in (?"+strings.Repeat(",?", len(IDs)-1)+")", appid)
	args := make([]interface{}, len(IDs))
	for idx := range IDs {
		args[idx] = IDs[idx]
	}
	_, err := mySQL.Exec(queryStr, args...)
	if err != nil {
		return err
	}

	return nil
}

func updateCategoryName(appid string, categoryID int, name string) error {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return errors.New("DB not init")
	}

	queryStr := fmt.Sprintf("UPDATE %s_categories SET CategoryName = ? where CategoryId = ?", appid)
	_, err := mySQL.Exec(queryStr, name, categoryID)
	if err != nil {
		return err
	}

	return nil
}

func selectSimilarQuestions(qID int, appID string) ([]SimilarQuestion, error) {
	query := fmt.Sprintf("SELECT SQ_Id, Content FROM %s_squestion WHERE Question_Id = ?", appID)
	db := util.GetMainDB()
	if db == nil {
		return nil, fmt.Errorf("DB not init")
	}
	rows, err := db.Query(query, qID)
	if err != nil {
		return nil, fmt.Errorf("query execute failed: %s", err)
	}
	defer rows.Close()
	var similarQuestions []SimilarQuestion = make([]SimilarQuestion, 0)

	for rows.Next() {
		var sq SimilarQuestion
		rows.Scan(&sq.Id, &sq.Content)
		similarQuestions = append(similarQuestions, sq)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("Scanning query failed: %s", err)
	}

	return similarQuestions, nil
}

// selectQuestion will return StdQuestion struct of the qid.
// if some input are missed it will return error, if no rows are found it will return sql.ErrNoRows.
func selectQuestions(groupID []int, appid string) ([]StdQuestion, error) {
	var questions = make([]StdQuestion, 0)
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
	defer result.Close()
	for result.Next() {
		var q StdQuestion
		result.Scan(&q.QuestionID, &q.Content, &q.CategoryID)
		questions = append(questions, q)
	}
	if err := result.Err(); err != nil {
		return nil, err
	}

	size := len(questions)
	if size == 0 {
		return nil, sql.ErrNoRows
	}
	if size != len(groupID) {
		return nil, fmt.Errorf("query can not found some of id that passed in")
	}

	return questions, nil
}

func FilterQuestions(appID string, content []string) ([]StdQuestion, error) {
	if len(content) == 0 {
		return nil, fmt.Errorf("content should have at least one element")
	}
	db := util.GetMainDB()
	query := fmt.Sprintf("SELECT Question_id, Content, CategoryId FROM %s_question WHERE Content IN (?%s) AND status = 0", appID, strings.Repeat(", ?", len(content)-1))
	var parameters = make([]interface{}, len(content))
	for i, c := range content {
		parameters[i] = c
	}
	rows, err := db.Query(query, parameters...)
	if err != nil {
		// Do not show detail query in output error message.
		util.LogError.Printf("query %s failed", query)
		return nil, fmt.Errorf("select questions by content query failed, %v", err)
	}
	defer rows.Close()
	var results = make([]StdQuestion, 0)
	for rows.Next() {
		var (
			content   string
			catID, id int
		)
		rows.Scan(&id, &content, &catID)
		stdQ := StdQuestion{QuestionID: id, Content: content, CategoryID: catID}
		results = append(results, stdQ)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("io failed, %v", err)
	}
	return results, nil
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

		// update standard question status
		sqlStr = fmt.Sprintf("UPDATE %s_question SET Status = 1 where Question_Id = %d", appid, qid)
		updapteStmt, err := t.Prepare(sqlStr)
		if err != nil {
			return fmt.Errorf("SQL Prepare err, %s", err)
		}
		defer updapteStmt.Close()

		_, err = updapteStmt.Exec()
		if err != nil {
			return fmt.Errorf("SQL Execution err, %s", err)
		}
	}

	return nil
}

//searchQuestionByContent return standard question based on content given.
//return util.ErrSQLRowNotFound if query is empty
func searchQuestionByContent(content string, appid string) (StdQuestion, error) {
	var q StdQuestion
	db := util.GetMainDB()
	if db == nil {
		return q, fmt.Errorf("main db connection pool is nil")
	}
	rawQuery := fmt.Sprintf("SELECT Question_id, Content, CategoryId FROM %s_question WHERE Content = ? ORDER BY Question_id DESC", appid)
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
func GetCategory(ID int, appid string) (Category, error) {
	db := util.GetMainDB()
	var c Category
	if db == nil {
		return c, fmt.Errorf("main db connection pool is nil")
	}
	rawQuery := fmt.Sprintf("SELECT CategoryId, CategoryName, ParentId FROM %s_categories WHERE CategoryId = ?", appid)
	err := db.QueryRow(rawQuery, ID).Scan(&c.ID, &c.Name, &c.ParentID)
	if err == sql.ErrNoRows {
		return c, err
	} else if err != nil {
		return c, fmt.Errorf("query row failed, %v", err)
	}
	return c, nil
}

//GetQuestionsByCategories search all the questions contained in given categories.
func GetQuestionsByCategories(categories []Category, appid string) ([]StdQuestion, error) {
	db := util.GetMainDB()
	if db == nil {
		return nil, fmt.Errorf("main db connection pool is nil")
	}
	rawQuery := fmt.Sprintf("SELECT Question_id, Content, CategoryId FROM %s_question WHERE CategoryId IN (? %s)", appid, strings.Repeat(",? ", len(categories)-1))
	var args = make([]interface{}, len(categories))
	for i, c := range categories {
		args[i] = c.ID
	}
	rows, err := db.Query(rawQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("query question failed, %v", err)
	}
	defer rows.Close()
	var questions = make([]StdQuestion, 0)
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

func FilterQuestion(condition QueryCondition, appid string) ([]int, map[int]string, error) {
	var qids = make([]int, 0)
	var aidMap = make(map[int]string)
	var query string
	var sqlParams []interface{}

	if HasCondition(condition) {
		// TODO: filter by condition
		query = `SELECT q.Question_Id, GROUP_CONCAT(DISTINCT a.Answer_Id) as aids from`

		var qids []int
		qSQL, err := questionSQL(condition, qids, &sqlParams, appid)
		if err != nil {
			return qids, aidMap, err
		}
		query += fmt.Sprintf(" (%s) as q", qSQL)

		var aids [][]string
		query += fmt.Sprintf(" inner join (%s) as a on q.Question_Id = a.Question_Id", answerSQL(condition, aids, &sqlParams, appid))

		dmCondition := "select Answer_Id, DynamicMenu from %s_dynamic_menu"
		dmCondition = fmt.Sprintf(dmCondition, appid)
		if condition.SearchDynamicMenu {
			if condition.Keyword != "" {
				dmCondition += " where DynamicMenu like ?"
				sqlParams = append(sqlParams, "%"+util.EscapeQuery(condition.Keyword)+"%")
			}
			query += fmt.Sprintf(" inner join (%s) as dm on dm.Answer_Id = a.Answer_Id", dmCondition)
		} else if condition.SearchAll {
			query += fmt.Sprintf(" left join (%s) as dm on dm.Answer_Id = a.Answer_Id", dmCondition)
		}

		rqCondition := "select Answer_Id, RelatedQuestion from %s_related_question"
		rqCondition = fmt.Sprintf(rqCondition, appid)
		if condition.SearchRelativeQuestion {
			if condition.Keyword != "" {
				rqCondition += " where RelatedQuestion like ?"
				sqlParams = append(sqlParams, "%"+util.EscapeQuery(condition.Keyword)+"%")
			}
			query += fmt.Sprintf(" inner join (%s) as rq on rq.Answer_Id = a.Answer_Id", rqCondition)
		} else if condition.SearchAll {
			query += fmt.Sprintf(" left join (%s) as rq on rq.Answer_Id = a.Answer_Id", rqCondition)
		}

		if len(condition.Dimension) != 0 {
			sqlStr, err := dimensionSQL(condition, appid)
			if err != nil {
				return qids, aidMap, err
			}
			query += fmt.Sprintf(" inner join (%s) as tag on tag.a_id = a.Answer_Id", sqlStr)
		}

		if condition.SearchAll && condition.Keyword != "" {
			query += " where q.Content like ? or a.Content like ? or rq.RelatedQuestion like ? or dm.DynamicMenu like ?"
			param := "%" + util.EscapeQuery(condition.Keyword) + "%"
			sqlParams = append(sqlParams, param, param, param, param)
		}

		query += " group by q.Question_Id order by q.Question_Id desc"

	} else {
		// no filter
		query = fmt.Sprintf(`SELECT q.Question_Id, GROUP_CONCAT(DISTINCT a.Answer_Id) as aids from %s_question as q
				inner join %s_answer as a on q.Question_Id = a.Question_Id
				group by q.Question_Id
				order by q.Question_Id desc`, appid, appid)
	}

	db := util.GetMainDB()
	rows, err := db.Query(query, sqlParams...)
	if err != nil {
		return qids, aidMap, err
	}

	for rows.Next() {
		var qid int
		var aidStr string

		rows.Scan(&qid, &aidStr)
		qids = append(qids, qid)
		aidMap[qid] = aidStr
	}
	return qids, aidMap, nil
}

func HasCondition(condition QueryCondition) bool {
	searchKeyword := condition.SearchAnswer || condition.SearchDynamicMenu || condition.SearchRelativeQuestion || condition.SearchQuestion || condition.SearchAll

	if !condition.NotShow && !searchKeyword && !condition.TimeSet && condition.CategoryId == 0 && len(condition.Dimension) == 0 {
		return false
	} else {
		return true
	}
}

func FetchQuestions(condition QueryCondition, qids []int, aids [][]string, appid string) ([]Question, error) {
	var questions []Question
	var sqlParams []interface{}
	var timeFormat string = "%Y-%m-%d %H:%i:%s"
	db := util.GetMainDB()

	query := "select q.Question_Id, q.CategoryId, q.Content, q.sQCount, q.CategoryName, a.Answer_Id, a.Content as acontent, a.Content_String as aContentString, a.Answer_CMD, a.Answer_CMD_Msg, a.Not_Show_In_Relative_Q, DATE_FORMAT(a.Begin_Time, '%s') as Begin_Time, DATE_FORMAT(a.End_Time, '%s') as End_Time, group_concat(DISTINCT rq.RelatedQuestion SEPARATOR '%s') as RelatedQuestion, group_concat(DISTINCT dm.DynamicMenu SEPARATOR '%s') as DynamicMenu, %s"
	query = fmt.Sprintf(query, timeFormat, timeFormat, SEPARATOR, SEPARATOR, "GROUP_CONCAT(DISTINCT tag.Tag_Id) as tag_ids")

	qSQL, err := questionSQL(condition, qids, &sqlParams, appid)
	if err != nil {
		return questions, err
	}
	query += fmt.Sprintf(" from (%s) as q", qSQL)
	query += fmt.Sprintf(" inner join (%s) as a on q.Question_Id = a.Question_Id", answerSQL(condition, aids, &sqlParams, appid))
	query += fmt.Sprintf(" left join %s_dynamic_menu as dm on dm.Answer_id = a.Answer_Id", appid)
	query += fmt.Sprintf(" left join %s_related_question as rq on rq.Answer_id = a.Answer_Id", appid)

	dimensionSQL, err := dimensionSQL(condition, appid)
	if err != nil {
		return questions, err
	}
	query += fmt.Sprintf(" left join (%s) as tag on tag.a_id = a.Answer_Id", dimensionSQL)
	query += " group by a.Answer_Id order by q.Question_Id desc, a.Answer_Id"

	// fetch
	rows, err := db.Query(query)
	if err != nil {
		return questions, err
	}
	defer rows.Close()

	// construct Questions
	var currentQuestion *Question
	tagMap, err := TagMapFactory(appid)
	if err != nil {
		return questions, err
	}

	for rows.Next() {
		var answer Answer
		var question Question
		var rq sql.NullString
		var dm sql.NullString
		var tagIDs sql.NullString
		var answerString string

		rows.Scan(&question.QuestionId, &question.CategoryId, &question.Content, &question.SQuestionConunt, &question.CategoryName, &answer.AnswerId, &answer.Content, &answerString, &answer.AnswerCmd, &answer.AnswerCmdMsg, &answer.NotShow, &answer.BeginTime, &answer.EndTime, &rq, &dm, &tagIDs)

		// encode answer content
		answer.Content = Escape(answer.Content)

		// transform tag id format
		if tagIDs.String == "" {
			answer.Dimension = []string{"", "", "", "", ""}
		} else {
			answer.Dimension = FormDimension(strings.Split(tagIDs.String, ","), tagMap)
		}

		answer.RelatedQuestion = rq.String
		answer.DynamicMenu = dm.String

		if currentQuestion == nil || currentQuestion.QuestionId != question.QuestionId {
			if currentQuestion != nil {
				questions = append(questions, *currentQuestion)
			}
			currentQuestion = &question
		}
		currentQuestion.Answers = append(currentQuestion.Answers, answer)
	}
	if currentQuestion != nil {
		questions = append(questions, *currentQuestion)
	}

	return questions, nil
}

func Escape(target string) string {
	re := regexp.MustCompile("<img.*?>")
	return re.ReplaceAllString(target, "[图片]")
}

func FormDimension(tagIDs []string, tagMap map[string]Tag) []string {
	// get tag string to type
	tags := []string{"", "", "", "", ""}
	if len(tagIDs) == 0 {
		return tags
	}

	for _, tagID := range tagIDs {
		tag := tagMap[tagID]
		index := tag.Type - 1

		if tags[index] == "" {
			tags[index] += tag.Content
		} else {
			tags[index] += fmt.Sprintf(",%s", tag.Content)
		}
	}

	return tags
}

func TagMapFactory(appid string) (map[string]Tag, error) {
	var tagMap = make(map[string]Tag)

	mySQL := util.GetMainDB()
	sql := `select tag.Tag_Id, tag_type.Type_id, tag.Tag_Name from %s_tag as tag
	left join %s_tag_type as tag_type on tag.Tag_Type = tag_type.Type_id`

	sql = fmt.Sprintf(sql, appid, appid)
	rows, err := mySQL.Query(sql)
	if err != nil {
		return tagMap, err
	}
	defer rows.Close()

	for rows.Next() {
		tag := Tag{}
		var tagID int
		var content string

		rows.Scan(&tagID, &tag.Type, &content)

		tagIDstr := strconv.Itoa(tagID)
		tag.Content = strings.Replace(content, "#", "", -1)

		tagMap[tagIDstr] = tag
	}

	return tagMap, nil
}

func questionSQL(condition QueryCondition, qids []int, sqlParam *[]interface{}, appid string) (string, error) {
	query := `select tmp_q.Question_Id, tmp_q.CategoryId, tmp_q.Content, tmp_q.SQuestion_count, fullc.CategoryName, count(vipshop_squestion.Content) as sQCount from (
		select * from %s_question
		where %s_question.status >= 0
		#CATEGORY_CONDITION#
		#KEYWORD_CONDITION#
		#QUESITION_CONDTION#
		order by %s_question.Question_Id desc
	) as tmp_q
	left join %s_categories on %s_categories.CategoryId = tmp_q.CategoryId
	left join (
		select level5.categoryid as id,concat_ws('/',level1.categoryname, level2.categoryname, level3.categoryname, level4.categoryname, level5.categoryname) AS CategoryName from (
			select categoryid, categoryname, parentid from %s_categories) as level5
			left join (select * from %s_categories) as level4 on level4.categoryid = level5.parentid
			left join (select * from %s_categories) as level3 on level3.categoryid = level4.parentid
			left join (select * from %s_categories) as level2 on level2.categoryid = level3.parentid
			left join (select * from %s_categories) as level1 on level1.categoryid = level2.parentid
	) as fullc on fullc.id = tmp_q.CategoryId
	left join vipshop_squestion on vipshop_squestion.Question_Id = tmp_q.Question_Id
    group by tmp_q.Question_Id`

	query = fmt.Sprintf(query, appid, appid, appid, appid, appid, appid, appid, appid, appid, appid)

	if len(qids) == 0 {
		query = strings.Replace(query, "#QUESITION_CONDTION#", "", -1)
	} else {
		idStr := GenIdStr(qids)
		questionCondition := fmt.Sprintf(" and %s_question.Question_Id in(%s)", appid, idStr)
		query = strings.Replace(query, "#QUESITION_CONDTION#", questionCondition, -1)
	}

	if condition.CategoryId == 0 {
		query = strings.Replace(query, "#CATEGORY_CONDITION#", "", -1)
	} else {
		// fectch parent categorires & replace category condition here
		categoryMap, err := GenCagtegoryMap(appid)

		if err != nil {
			return query, err
		}
		category := categoryMap[condition.CategoryId]
		idStr := strconv.Itoa(condition.CategoryId)
		if category != nil && len(category.Children) > 0 {
			idStr += fmt.Sprintf(",%s", GenIdStr(category.Children))
		}

		categoryCondition := fmt.Sprintf(" and %s_question.CategoryId in(%s)", appid, idStr)
		query = strings.Replace(query, "#CATEGORY_CONDITION#", categoryCondition, -1)
	}

	if condition.SearchQuestion && condition.Keyword != "" {
		// replace keyword condition
		keywordCondition := fmt.Sprintf(" and %s_question.content like ?", appid)
		newParam := append(*sqlParam, "%"+util.EscapeQuery(condition.Keyword)+"%")
		*sqlParam = newParam
		query = strings.Replace(query, "#KEYWORD_CONDITION#", keywordCondition, -1)
	} else {
		query = strings.Replace(query, "#KEYWORD_CONDITION#", "", -1)
	}

	return query, nil
}

func GenIdStr(ids []int) string {
	idStr := ""
	for i, id := range ids {
		if i == 0 {
			idStr += strconv.Itoa(id)
		} else {
			idStr += fmt.Sprintf(",%s", strconv.Itoa(id))
		}
	}
	return idStr
}

func GenCagtegoryMap(appid string) (map[int]*Category, error) {
	query := `select CategoryId, ParentId from %s_categories order by ParentId`
	query = fmt.Sprintf(query, appid)

	db := util.GetMainDB()
	categoryMap := make(map[int]*Category)

	rows, err := db.Query(query)
	if err != nil {
		return categoryMap, err
	}
	defer rows.Close()

	for rows.Next() {
		var category Category
		var parent int
		rows.Scan(&category.ID, &category.ParentID)
		categoryMap[category.ID] = &category

		parent = category.ParentID
		for parent != 0 {
			parentCategory := categoryMap[parent]
			parentCategory.Children = append(parentCategory.Children, category.ID)
			parent = parentCategory.ParentID
		}
	}

	return categoryMap, nil
}

func answerSQL(condition QueryCondition, aids [][]string, sqlParam *[]interface{}, appid string) string {
	query := `select tmp_a.Answer_Id, tmp_a.Content, tmp_a.Content_String, tmp_a.Answer_CMD, tmp_a.Answer_CMD_Msg, tmp_a.Not_Show_In_Relative_Q, tmp_a.Begin_Time, tmp_a.End_Time, tmp_a.Question_Id from %s_answer as tmp_a
			#ANSWER_CONDITION#
			#TIME_CONDITION#
			#KEYWORD_CONDTION#
			#NOT_SHOW_CONDITION#`

	query = fmt.Sprintf(query, appid)
	hasWhere := false

	if len(aids) == 0 {
		query = strings.Replace(query, "#ANSWER_CONDITION#", "", -1)
	} else {
		idStr := ""
		for i, ids := range aids {
			for j, id := range ids {
				if i == 0 && j == 0 {
					idStr += id
				} else {
					idStr += fmt.Sprintf(",%s", id)
				}
			}
		}

		var answerCondition string
		if hasWhere {
			answerCondition = fmt.Sprintf(" and tmp_a.Answer_Id in(%s)", idStr)
		} else {
			answerCondition = fmt.Sprintf(" where tmp_a.Answer_Id in(%s)", idStr)
			hasWhere = true
		}
		query = strings.Replace(query, "#ANSWER_CONDITION#", answerCondition, -1)
	}

	if condition.TimeSet && condition.BeginTime != "" && condition.EndTime != "" {
		// replace time condition
		var timeCondition string

		if hasWhere {
			timeCondition = fmt.Sprintf(" and tmp_a.Begin_Time >= '%s' and tmp_a.End_Time <= '%s'", condition.BeginTime, condition.EndTime)
		} else {
			hasWhere = true
			timeCondition = fmt.Sprintf(" where tmp_a.Begin_Time >= '%s' and tmp_a.End_Time <= '%s'", condition.BeginTime, condition.EndTime)
		}
		query = strings.Replace(query, "#TIME_CONDITION#", timeCondition, -1)
	} else {
		query = strings.Replace(query, "#TIME_CONDITION#", "", -1)
	}

	if condition.SearchAnswer && condition.Keyword != "" {
		// replace keword condition
		var keywordCondition string
		if hasWhere {
			keywordCondition = " and tmp_a.Content_String like ?"
		} else {
			hasWhere = true
			keywordCondition = " where tmp_a.Content_String like ?"
		}
		query = strings.Replace(query, "#KEYWORD_CONDTION#", keywordCondition, -1)
		newParam := append(*sqlParam, "%"+util.EscapeQuery(condition.Keyword)+"%")
		*sqlParam = newParam
	} else {
		query = strings.Replace(query, "#KEYWORD_CONDTION#", "", -1)
	}

	if condition.NotShow {
		// replace not show condition
		var notShowCondition string
		if hasWhere {
			notShowCondition = " and tmp_a.Not_Show_In_Relative_Q = 1"
		} else {
			notShowCondition = " where tmp_a.Not_Show_In_Relative_Q = 1"
			hasWhere = true
		}
		query = strings.Replace(query, "#NOT_SHOW_CONDITION#", notShowCondition, -1)
	} else {
		query = strings.Replace(query, "#NOT_SHOW_CONDITION#", "", -1)
	}
	return query
}

func dimensionSQL(condition QueryCondition, appid string) (string, error) {
	// sql without condition
	query := `select %s_answertag.Answer_Id as a_id,%s_answertag.Tag_Id, %s_tag.Tag_Type, %s_tag.Tag_Name from %s_answertag
	left join %s_tag on %s_tag.Tag_Id = %s_answertag.Tag_Id
	left join %s_tag_type on %s_tag_type.Type_id = %s_tag.Tag_Type`

	if len(condition.Dimension) == 0 {
		appids := []interface{}{appid, appid, appid, appid, appid, appid, appid, appid, appid, appid, appid}
		return fmt.Sprintf(query, appids...), nil
	}

	query = `select Answer_Id as a_id, tag_ids from (
		SELECT answer_id, GROUP_CONCAT(DISTINCT ans_tag.Tag_Id) as tag_ids from (SELECT answer_id, Tag_Type, anst.Tag_Id
		FROM   %s_answertag as anst, %s_tag as tag
		WHERE  anst.tag_id IN ( %s ) and anst.tag_id = tag.Tag_Id
		GROUP  BY answer_id, Tag_Type) as ans_tag group by answer_id having count(*) = %d
	) as tmp_tags`

	// create tag id string
	// get dimension to tag id map
	var tagIDs []int
	dimensionToIdMAP, err := DimensionToIdMapFactory(appid)
	if err != nil {
		return query, err
	}
	for _, dimensionGroup := range condition.Dimension {
		dimensions := strings.Split(dimensionGroup.Content, ",")
		for _, dimension := range dimensions {
			if id, ok := dimensionToIdMAP[dimension]; ok {
				tagIDs = append(tagIDs, id)
			}
		}
	}
	tagIDstr := GenIdStr(tagIDs)

	query = fmt.Sprintf(query, appid, appid, tagIDstr, len(condition.Dimension))
	return query, nil
}

func DimensionToIdMapFactory(appid string) (map[string]int, error) {
	var dimensionIdMap = make(map[string]int)
	query := "select Tag_Id, Tag_Name from %s_tag"
	query = fmt.Sprintf(query, appid)

	db := util.GetMainDB()
	rows, err := db.Query(query)
	if err != nil {
		return dimensionIdMap, err
	}
	defer rows.Close()

	for rows.Next() {
		var dimension string
		var tagID int
		rows.Scan(&tagID, &dimension)

		dimensionIdMap[dimension] = tagID
	}

	return dimensionIdMap, nil
}

func UpdateQuestion(appid string, question *Question) (err error) {
	db := util.GetMainDB()
	if db == nil {
		err = fmt.Errorf("main db connection pool is nil")
		return
	}

	tx, err := db.Begin()
	if err != nil {
		return
	}
	defer util.ClearTransition(tx)

	err = updateQuestion(appid, question, tx)
	if err != nil {
		return
	}

	err = tx.Commit()
	return
}

func updateQuestion(appid string, question *Question, tx *sql.Tx) (err error) {
	parameters := []interface{}{
		question.Content,
		question.CategoryId,
	}
	sqlStr := fmt.Sprintf("UPDATE `%s_question` SET `Content`=?, `CategoryId`=?, `Status`=1 WHERE Question_Id=%d", appid, question.QuestionId)

	_, err = tx.Exec(sqlStr, parameters...)
	if err != nil {
		return
	}

	var answerIDs []int
	for _, answer := range question.Answers {
		answer.QuestionId = question.QuestionId
		var answerID int
		if answer.AnswerId != 0 {
			err = updateAnswer(appid, &answer, tx)
			answerID = answer.AnswerId
		} else {
			aid, tmpErr := insertAnswer(appid, int64(question.QuestionId), &answer, tx)
			if tmpErr != nil {
				err = tmpErr
				return
			}
			answerID = int(aid)
		}

		if err != nil {
			return
		}
		answerIDs = append(answerIDs, answerID)
	}

	var targetIDSQL string
	for index, aid := range answerIDs {
		if index == 0 {
			targetIDSQL = fmt.Sprintf("Answer_Id != %d", aid)
		} else {
			targetIDSQL += fmt.Sprintf(" and Answer_Id != %d", aid)
		}
	}
	sqlStr = fmt.Sprintf("DELETE FROM %s_answer WHERE (%s) and Question_Id=%d", appid, targetIDSQL, question.QuestionId)
	_, err = tx.Exec(sqlStr)
	return
}

func updateAnswer(appid string, answer *Answer, tx *sql.Tx) (err error) {
	sqlStr := fmt.Sprintf("UPDATE %s_answer SET Question_Id =?, Content=?, Answer_CMD=?, Begin_Time=?, End_Time=?, Status=1, Not_Show_In_Relative_Q=?, Content_String=?, Answer_CMD_Msg=? WHERE Answer_Id=?", appid)
	answerCmdMsg := answerCmd(answer.AnswerCmd, answer.AnswerCmdMsg)
	parameters := []interface{}{
		answer.QuestionId,
		answer.Content,
		answer.AnswerCmd,
		answer.BeginTime,
		answer.EndTime,
		answer.NotShow,
		answer.Content,
		answerCmdMsg,
		answer.AnswerId,
	}

	_, err = tx.Exec(sqlStr, parameters...)
	if err != nil {
		return
	}

	// delete old dynamic menu, related questions, dimensions
	sqlStr = fmt.Sprintf("DELETE FROM %s_dynamic_menu WHERE Answer_id = %d", appid, answer.AnswerId)
	_, err = tx.Exec(sqlStr)
	if err != nil {
		return
	}

	sqlStr = fmt.Sprintf("DELETE FROM %s_related_question WHERE Answer_id = %d", appid, answer.AnswerId)
	_, err = tx.Exec(sqlStr)
	if err != nil {
		return
	}

	sqlStr = fmt.Sprintf("DELETE FROM %s_answertag WHERE Answer_Id = %d", appid, answer.AnswerId)
	_, err = tx.Exec(sqlStr)
	if err != nil {
		return
	}

	// add back
	if len(answer.DynamicMenus) > 0 {
		err = insertAnswerLabels(appid, int64(answer.AnswerId), DynamicMenu, answer.DynamicMenus, tx)
		if err != nil {
			return
		}
	}

	if len(answer.RelatedQuestions) > 0 {
		err = insertAnswerLabels(appid, int64(answer.AnswerId), RelatedQuestion, answer.RelatedQuestions, tx)
		if err != nil {
			return
		}
	}
	if len(answer.DimensionIDs) > 0 {
		err = insertAnswerDimensions(appid, int64(answer.AnswerId), answer.DimensionIDs, tx)
		if err != nil {
			return
		}
	}

	err = imagesManager.DeleteMediaRef(answer.AnswerId)
	if err != nil {
		return
	}
	if len(answer.Images) > 0 {
		err = imagesManager.CreateMediaRef(answer.AnswerId, answer.Images)
	}

	return
}

func answerCmd(cmd, cmdMsg string) string {
	if cmd == "" {
		return ""
	} else if cmd == "shopping" {
		return cmdMsg
	}

	return answerCmdLog(cmd)
}

func answerCmdLog(cmd string) (chineseCmd string) {
	switch cmd {
	case "":
		chineseCmd = "无指令"
	case "order_track":
		chineseCmd = "物流信息查询"
	case "order_info":
		chineseCmd = "订单信息查询"
	case "scene_id":
		chineseCmd = "场景标识"
	case "cash":
		chineseCmd = "提现"
	case "order_cancel":
		chineseCmd = "取消订单"
	case "apply_for_return":
		chineseCmd = "退货申请"
	case "exchange_goods":
		chineseCmd = "换货申请"
	case "vip_finance":
		chineseCmd = "唯品金融"
	case "query_refund":
		chineseCmd = "查询退款"
	case "shopping":
		chineseCmd = "购物"
	}
	return
}

func InsertQuestion(appid string, question *Question, answers []Answer) (qid int64, err error) {
	db := util.GetMainDB()
	if db == nil {
		err = fmt.Errorf("main db connection pool is nil")
		return
	}

	tx, err := db.Begin()
	if err != nil {
		err = fmt.Errorf("transaction start failed, %v", err)
		return
	}
	defer util.ClearTransition(tx)

	qid, err = insertQuestion(appid, question, answers, tx)
	if err != nil {
		return
	}

	err = tx.Commit()
	return

}

func insertQuestion(appid string, question *Question, answers []Answer, tx *sql.Tx) (qid int64, err error) {
	columValues := []interface{}{question.Content, question.CategoryId, question.User, question.SQuestionConunt}
	sql := fmt.Sprintf("INSERT INTO %s_question (Content, CategoryId, CreatedUser, SQuestion_count, CreatedTime) VALUES (?, ?, ?, ?, now())", appid)

	result, err := tx.Exec(sql, columValues...)
	if err != nil {
		return
	}
	qid, err = result.LastInsertId()

	err = insertAnswers(appid, qid, answers, tx)
	if err != nil {
		return
	}

	return
}

func insertAnswers(appid string, qid int64, answers []Answer, tx *sql.Tx) (err error) {
	for _, answer := range answers {
		_, err = insertAnswer(appid, qid, &answer, tx)
		if err != nil {
			return
		}
	}

	return
}

func insertAnswer(appid string, qid int64, answer *Answer, tx *sql.Tx) (answerID int64, err error) {
	sqlStr := fmt.Sprintf("INSERT INTO %s_answer (Question_Id, Content, Answer_CMD, Begin_Time, End_Time, Not_Show_In_Relative_Q, Answer_CMD_Msg, Content_String) VALUES (?, ?, ?, ?, ?, ?, ?, ?);", appid)
	answerCMDMsg := answerCmd(answer.AnswerCmd, answer.AnswerCmdMsg)
	columnValues := []interface{}{qid, answer.Content, answer.AnswerCmd, answer.BeginTime, answer.EndTime, answer.NotShow, answerCMDMsg, answer.Content}

	result, err := tx.Exec(sqlStr, columnValues...)
	if err != nil {
		util.LogError.Printf("error: %s", err.Error())
		return
	}

	answerID, err = result.LastInsertId()
	if err != nil {
		return
	}

	if len(answer.RelatedQuestions) != 0 {
		err = insertAnswerLabels(appid, answerID, RelatedQuestion, answer.RelatedQuestions, tx)
		if err != nil {
			return
		}
	}

	if len(answer.DynamicMenus) != 0 {
		err = insertAnswerLabels(appid, answerID, DynamicMenu, answer.DynamicMenus, tx)
		if err != nil {
			return
		}
	}

	if len(answer.DimensionIDs) != 0 {
		err = insertAnswerDimensions(appid, answerID, answer.DimensionIDs, tx)
		if err != nil {
			return
		}
	}

	if len(answer.Images) != 0 {
		err = imagesManager.CreateMediaRef(int(answerID), answer.Images)
	}

	return
}

func insertAnswerLabels(appid string, answerID int64, labelType int, labels []string, tx *sql.Tx) (err error) {
	var table string
	var column string
	if labelType == RelatedQuestion {
		table = fmt.Sprintf("%s_related_question", appid)
		column = "RelatedQuestion"
	} else if labelType == DynamicMenu {
		table = fmt.Sprintf("%s_dynamic_menu", appid)
		column = "DynamicMenu"
	} else {
		return fmt.Errorf("Error Label Type")
	}

	var values []interface{}
	sqlStr := fmt.Sprintf("INSERT INTO %s (Answer_id, %s) VALUES", table, column)
	for index, label := range labels {
		if index == 0 {
			sqlStr += " (?, ?)"
		} else {
			sqlStr += ", (?, ?)"
		}
		values = append(values, answerID)
		values = append(values, label)
	}
	sqlStr += ";"

	_, err = tx.Exec(sqlStr, values...)
	return
}

func insertAnswerDimensions(appid string, answerID int64, dimensions []int, tx *sql.Tx) (err error) {
	sqlStr := fmt.Sprintf("INSERT IGNORE INTO %s_answertag (Answer_Id, Tag_Id, CreatedTime) VALUES", appid)

	var values []interface{}
	for index, dimension := range dimensions {
		if index == 0 {
			sqlStr += " (?, ?, now())"
		} else {
			sqlStr += ", (?, ?, now())"
		}
		values = append(values, answerID)
		values = append(values, dimension)
	}

	sqlStr += ";"

	_, err = tx.Exec(sqlStr, values...)
	return
}

func FindQuestions(appid string, targets []Question) (questions []Question, err error) {
	db := util.GetMainDB()

	if len(targets) == 0 {
		return
	}

	var conditions []interface{}
	appids := []interface{}{appid, appid, appid, appid, appid, appid}
	sql := fmt.Sprintf(`SELECT Question_Id, Content, CategoryId, categoryname FROM %s_question as q
		left join (
			select level5.categoryid as id,concat_ws('/',level1.categoryname, level2.categoryname, level3.categoryname, level4.categoryname, level5.categoryname) AS CategoryName from (
				select categoryid, categoryname, parentid from %s_categories) as level5
				left join (select * from %s_categories) as level4 on level4.categoryid = level5.parentid
				left join (select * from %s_categories) as level3 on level3.categoryid = level4.parentid
				left join (select * from %s_categories) as level2 on level2.categoryid = level3.parentid
				left join (select * from %s_categories) as level1 on level1.categoryid = level2.parentid
		) as fullc on fullc.id = q.CategoryId where`, appids...)

	var shouldOr bool = false
	for _, dao := range targets {
		clause, condition := genQuestionWhereClause(&dao)
		if len(condition) == 0 {
			continue
		}
		conditions = append(conditions, condition...)

		if shouldOr {
			sql += fmt.Sprintf(" or %s", clause)
		} else {
			sql += fmt.Sprintf(" %s", clause)
			shouldOr = true
		}
	}

	rows, err := db.Query(sql, conditions...)
	if err != nil {
		return
	}

	for rows.Next() {
		question := Question{}
		rows.Scan(&question.QuestionId, &question.Content, &question.CategoryId, &question.CategoryName)

		questions = append(questions, question)
	}

	return
}

func DeleteQuestions(appid string, targets []Question) error {
	db := util.GetMainDB()
	if db == nil {
		return fmt.Errorf("main db connection pool is nil")
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("transaction start failed, %v", err)
	}
	defer util.ClearTransition(tx)

	err = deleteQuestions(appid, targets, tx)
	tx.Commit()
	return err
}

func deleteQuestions(appid string, targets []Question, tx *sql.Tx) error {
	// what we should de when delete questions
	// 1. find answers of questions
	// 2. delete answers
	// 3. delete quesions

	var targetAnswers []Answer
	for _, question := range targets {
		targetAnswer := Answer{
			QuestionId: question.QuestionId,
		}

		targetAnswers = append(targetAnswers, targetAnswer)
	}
	targetAnswers, err := findAnswers(appid, targetAnswers, tx)

	var answerIDs []Answer = make([]Answer, len(targetAnswers))
	for index, answer := range targetAnswers {
		targetAnswer := Answer{
			AnswerId: answer.AnswerId,
		}
		answerIDs[index] = targetAnswer
	}

	// delete answers
	err = deleteAnswers(appid, answerIDs, tx)
	if err != nil {
		return err
	}

	// delete questions
	sql, conditions := genDeleteQuestionSQL(appid, targets)
	_, err = tx.Exec(sql, conditions...)
	return err
}

func genDeleteQuestionSQL(appid string, targets []Question) (string, []interface{}) {
	sql := fmt.Sprintf("UPDATE %s_question SET Status = -1 where", appid)

	var shouldOr bool = false
	var conditions []interface{}
	for _, dao := range targets {
		clause, condition := genQuestionWhereClause(&dao)
		if len(condition) == 0 {
			continue
		}
		conditions = append(conditions, condition...)

		if shouldOr {
			sql += fmt.Sprintf(" or %s", clause)
		} else {
			sql += fmt.Sprintf(" %s", clause)
			shouldOr = true
		}
	}

	return sql, conditions
}

func genQuestionWhereClause(dao *Question) (string, []interface{}) {
	var conditions []interface{}
	var shouldAnd bool = false
	whereClause := "("

	if dao.QuestionId != 0 {
		appendClauseAndConditions(&whereClause, &conditions, dao.QuestionId, "Question_Id", &shouldAnd)
	}

	if dao.Content != "" {
		appendClauseAndConditions(&whereClause, &conditions, dao.Content, "Content", &shouldAnd)
	}

	if dao.CategoryId != 0 {
		appendClauseAndConditions(&whereClause, &conditions, dao.CategoryId, "CategoryId", &shouldAnd)
	}
	whereClause += ")"

	if len(conditions) == 0 {
		return "", conditions
	}

	return whereClause, conditions
}

func FindAnswers(appid string, targets []Answer) (answers []Answer, err error) {
	db := util.GetMainDB()
	if db == nil {
		err = fmt.Errorf("main db connection pool is nil")
		return
	}

	tx, err := db.Begin()
	if err != nil {
		err = fmt.Errorf("transaction start failed, %v", err)
		return
	}
	defer util.ClearTransition(tx)

	answers, err = findAnswers(appid, targets, tx)
	tx.Commit()
	return
}

func findAnswers(appid string, targets []Answer, tx *sql.Tx) ([]Answer, error) {
	sql, conditions := genFindAnswersSQL(appid, targets)

	var answers []Answer
	rows, err := tx.Query(sql, conditions...)
	if err != nil {
		return answers, err
	}

	for rows.Next() {
		answer := Answer{}
		err = rows.Scan(&answer.AnswerId, &answer.QuestionId, &answer.Content, &answer.AnswerCmd, &answer.BeginTime, &answer.EndTime, &answer.AnswerCmdMsg, &answer.NotShow)
		if err != nil {
			return answers, err
		}
		answers = append(answers, answer)
	}

	return answers, nil
}

func genFindAnswersSQL(appid string, targets []Answer) (string, []interface{}) {
	var conditions []interface{}
	var timeFormat string = "%Y-%m-%d %H:%i:%s"
	sql := fmt.Sprintf("SELECT Answer_Id, Question_Id, Content, Answer_CMD, DATE_FORMAT(Begin_Time, '%s'), DATE_FORMAT(End_Time, '%s'), Answer_CMD_Msg, Not_Show_In_Relative_Q FROM %s_answer", timeFormat, timeFormat, appid)

	if len(targets) == 0 {
		return sql, conditions
	}

	var shouldOr bool = false
	for _, dao := range targets {
		clause, condition := genAnswerWhereClause(&dao)
		if len(condition) == 0 {
			continue
		}

		conditions = append(conditions, condition...)
		if shouldOr {
			sql += fmt.Sprintf(" or %s", clause)
		} else {
			sql += fmt.Sprintf(" where %s", clause)
			shouldOr = true
		}
	}

	return sql, conditions
}

func whereDateStringTemplate(start string, end string, shouldAnd *bool) {
	clause := ""
	if *shouldAnd {
		clause += "and"
	}

	*shouldAnd = true

	if start != "" {
		clause += "Begin_Time > ?"
	}
}

func DeleteAnswers(appid string, targets []Answer) error {
	db := util.GetMainDB()
	if db == nil {
		return fmt.Errorf("main db connection pool is nil")
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("transaction start failed, %v", err)
	}
	defer util.ClearTransition(tx)

	err = deleteAnswers(appid, targets, tx)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func deleteAnswers(appid string, answers []Answer, tx *sql.Tx) error {
	// what we should do when delete an answer
	// 1. delete relative questions
	// 2. delete dynamic menu
	// 3. delete answer

	if len(answers) == 0 {
		return nil
	}

	var targetLabels []AnswerLabelDAO
	var conditions []interface{}
	var shouldOr bool = false
	var err error
	sqlStr := fmt.Sprintf("DELETE FROM %s_answer where", appid)
	for _, answer := range answers {
		var targetLabel AnswerLabelDAO = AnswerLabelDAO{
			AnswerId: answer.AnswerId,
		}
		targetLabels = append(targetLabels, targetLabel)

		clause, condition := genAnswerWhereClause(&answer)
		if len(condition) == 0 {
			continue
		}
		conditions = append(conditions, condition...)

		if shouldOr {
			sqlStr = fmt.Sprintf("%s or %s", sqlStr, clause)
		} else {
			sqlStr = fmt.Sprintf("%s %s", sqlStr, clause)
			shouldOr = true
		}

		// delete image ref
		if len(answer.Images) > 0 {
			err = imagesManager.DeleteMediaRef(answer.AnswerId)
			if err != nil {
				return err
			}
		}
	}

	// delete relative questions
	if len(targetLabels) != 0 {
		err = deleteAnswerLabels(appid, RelatedQuestion, tx, targetLabels)
		if err != nil {
			return err
		}

		// delete dynamic menu
		err = deleteAnswerLabels(appid, DynamicMenu, tx, targetLabels)
		if err != nil {
			return err
		}
	}

	// delete answer
	_, err = tx.Exec(sqlStr, conditions...)
	return err
}

func genAnswerWhereClause(dao *Answer) (string, []interface{}) {
	var conditions []interface{}
	var shouldAnd bool = false
	whereClause := "("

	if dao.QuestionId != 0 {
		appendClauseAndConditions(&whereClause, &conditions, dao.QuestionId, "Question_Id", &shouldAnd)
	}

	if dao.AnswerId != 0 {
		appendClauseAndConditions(&whereClause, &conditions, dao.AnswerId, "Answer_Id", &shouldAnd)
	}

	if dao.Content != "" {
		appendClauseAndConditions(&whereClause, &conditions, dao.Content, "Content", &shouldAnd)
	}

	if dao.BeginTime != "" {
		if shouldAnd {
			whereClause += " AND"
		}
		shouldAnd = true

		whereClause += " Begin_Time > ?"
		conditions = append(conditions, dao.BeginTime)
	}

	if dao.EndTime != "" {
		if shouldAnd {
			whereClause += " AND"
		}
		shouldAnd = true

		whereClause += " End_Time < ?"
		conditions = append(conditions, dao.EndTime)
	}
	whereClause += ")"

	if len(conditions) == 0 {
		return "", conditions
	}

	return whereClause, conditions
}

func FindAnswerLabels(appid string, labelType int, targets []AnswerLabelDAO) (labels []AnswerLabelDAO, err error) {
	db := util.GetMainDB()
	if db == nil {
		err = fmt.Errorf("main db connection pool is nil")
		return
	}

	tx, err := db.Begin()
	if err != nil {
		err = fmt.Errorf("transaction start failed, %v", err)
		return
	}
	defer util.ClearTransition(tx)

	labels, err = findAnswerLabels(appid, labelType, tx, targets)
	if err != nil {
		return
	}

	tx.Commit()
	return
}

func findAnswerLabels(appid string, labelType int, tx *sql.Tx, targetLabels []AnswerLabelDAO) (labels []AnswerLabelDAO, err error) {
	table, column, err := getAnswerLabelTableAndColumn(appid, labelType)
	if err != nil {
		return
	}

	sql, conditions := genFindAnswerLabelSQL(appid, table, column, targetLabels)
	if len(conditions) == 0 {
		return
	}

	rows, err := tx.Query(sql, conditions...)
	if err != nil {
		return
	}

	for rows.Next() {
		label := AnswerLabelDAO{}

		err = rows.Scan(&label.Id, &label.AnswerId, &label.Content)
		if err != nil {
			return
		}
		labels = append(labels, label)
	}
	return
}

func genFindAnswerLabelSQL(appid, table, column string, targets []AnswerLabelDAO) (sql string, conditions []interface{}) {
	sql = fmt.Sprintf("select id, Answer_id, %s FROM %s WHERE", column, table)

	shouldOr := false
	for _, dao := range targets {
		clause, condition := genAnswerLableWhereClause(dao, column)
		if len(condition) == 0 {
			continue
		}
		conditions = append(conditions, condition...)

		if shouldOr {
			sql += fmt.Sprintf(" or %s", clause)
		} else {
			sql += fmt.Sprintf(" %s", clause)
		}
		shouldOr = true
	}
	return
}

func deleteAnswerLabels(appid string, labelType int, tx *sql.Tx, targetLabels []AnswerLabelDAO) (err error) {
	var table string
	var column string

	if len(targetLabels) == 0 {
		return nil
	}

	table, column, err = getAnswerLabelTableAndColumn(appid, labelType)
	if err != nil {
		return
	}

	sql, conditions := genDeleteAnswerLabelSQL(table, column, targetLabels)

	_, err = tx.Exec(sql, conditions...)
	return err
}

func getAnswerLabelTableAndColumn(appid string, labelType int) (table string, column string, err error) {
	if labelType == RelatedQuestion {
		table = fmt.Sprintf("%s_related_question", appid)
		column = "RelatedQuestion"
	} else if labelType == DynamicMenu {
		table = fmt.Sprintf("%s_dynamic_menu", appid)
		column = "DynamicMenu"
	} else {
		err = fmt.Errorf("Error Label Type")
	}
	return
}

func genDeleteAnswerLabelSQL(table string, column string, targets []AnswerLabelDAO) (string, []interface{}) {
	var conditions []interface{}

	var shouldOr bool = false
	sql := fmt.Sprintf("DELETE FROM %s where", table)
	for _, dao := range targets {
		clause, condition := genAnswerLableWhereClause(dao, column)
		if len(condition) == 0 {
			continue
		}
		conditions = append(conditions, condition...)

		if shouldOr {
			sql += fmt.Sprintf(" or %s", clause)
		} else {
			sql += clause
		}

		shouldOr = true
	}

	return sql, conditions
}

func genAnswerLableWhereClause(dao AnswerLabelDAO, column string) (string, []interface{}) {
	var conditions []interface{}
	whereClause := "("
	shouldAnd := false

	if dao.Id != 0 {
		appendClauseAndConditions(&whereClause, &conditions, dao.Id, "id", &shouldAnd)
	}

	if dao.AnswerId != 0 {
		appendClauseAndConditions(&whereClause, &conditions, dao.AnswerId, "Answer_id", &shouldAnd)
	}

	if dao.Content != "" {
		appendClauseAndConditions(&whereClause, &conditions, dao.AnswerId, column, &shouldAnd)
	}
	whereClause += ")"

	if len(conditions) == 0 {
		return "", conditions
	}

	return whereClause, conditions
}

func genSimilarQuestionWhereClause(dao SimilarQuestionDAO) (string, []interface{}) {
	var conditions []interface{}
	whereClause := "("
	shouldAnd := false
	if dao.Qid != 0 {
		conditions = append(conditions, dao.Qid)
		clause := whereClauseTemplate("Question_Id", &shouldAnd)
		whereClause += clause
	}

	if dao.Content != "" {
		conditions = append(conditions, dao.Content)
		clause := whereClauseTemplate("Content", &shouldAnd)
		whereClause += clause
	}

	if dao.Sid != 0 {
		conditions = append(conditions, dao.Sid)
		clause := whereClauseTemplate("SQ_Id", &shouldAnd)
		whereClause += clause
	}
	whereClause += ")"

	if len(conditions) == 0 {
		return "", conditions
	}

	return whereClause, conditions
}

func whereClauseTemplate(column string, shouldAnd *bool) string {
	clause := ""
	if *shouldAnd {
		clause += fmt.Sprintf(" and %s = ?", column)
	} else {
		clause += fmt.Sprintf("%s = ?", column)
		*shouldAnd = true
	}

	return clause
}

func appendClauseAndConditions(whereClause *string, conditions *[]interface{}, argument interface{}, column string, shouldAnd *bool) {
	*conditions = append(*conditions, argument)

	clause := whereClauseTemplate(column, shouldAnd)
	*whereClause += clause
}

// CategoryTree is the implement of tree strucure of all Category data.
// Root category id should always be 0.
// Index should keep the references of all node to do quick search.
type CategoryTree struct {
	Index map[int64]*Category
	Root  *Category
}

//NewCategoryTree create a CategoryTree
func NewCategoryTree(appid string) (*CategoryTree, error) {
	rawQuery := "SELECT CategoryId, CategoryName, ParentId FROM " + appid + "_categories"
	db := util.GetMainDB()
	if db == nil {
		return nil, fmt.Errorf("main db is not init")
	}
	rows, err := db.Query(rawQuery)
	if err != nil {
		return nil, fmt.Errorf("query failed, %v", err)
	}
	defer rows.Close()
	var rootCategory = &Category{ID: 0}
	var tree = &CategoryTree{
		Index: map[int64]*Category{0: rootCategory},
		Root:  rootCategory,
	}
	for rows.Next() {
		var (
			catID    int64
			name     string
			parentId int64
		)
		err := rows.Scan(&catID, &name, &parentId)
		if err != nil {
			return nil, fmt.Errorf("scan io error, %v", err)
		}
		c, exists := tree.Index[catID]
		if !exists {
			c = &Category{ID: int(catID)}
		}
		c.Name = name
		c.ParentID = int(parentId)
		parent, haveParent := tree.Index[parentId]
		if haveParent {
			parent.Children = append(parent.Children, int(catID))
		} else {
			parent = &Category{
				ID:       int(parentId),
				Children: []int{int(catID)},
			}
		}

		tree.Index[catID] = c
	}
	return tree, nil
}

func (t *CategoryTree) SubCategories(catID int64) []*Category {
	r := t.Index[catID]
	var categories []*Category
	for _, child := range r.Children {
		subcats := t.SubCategories(int64(child))
		categories = append(categories, subcats...)
	}
	categories = append(categories, r)
	return categories
}

func LabelMap(appid string) (map[int]string, error) {
	if tagMap != nil {
		return tagMap, nil
	}

	tm := make(map[int]string)
	tagMap = tm

	sqlStr := fmt.Sprintf("SELECT Tag_Id, Tag_Name FROM %s_tag", appid)

	db := util.GetMainDB()
	if db == nil {
		err := fmt.Errorf("Unable to get mysql connection")
		return tagMap, err
	}

	rows, err := db.Query(sqlStr)
	if err != nil {
		return tagMap, err
	}
	for rows.Next() {
		var tagID int
		var tagName string

		rows.Scan(&tagID, &tagName)
		tagMap[tagID] = strings.Replace(tagName, "#", "", -1)
	}
	return tagMap, nil
}

