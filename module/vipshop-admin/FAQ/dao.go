package FAQ

import (
	"database/sql"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"emotibot.com/emotigo/module/vipshop-admin/util"
)

const SEPARATOR = "#SEPARATE_TOKEN#"

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
// It need to joined with StdQuestions table, because it need to validate the data.
func GetRFQuestions() ([]RFQuestion, error) {
	var questions = make([]RFQuestion, 0)
	db := util.GetMainDB()
	if db == nil {
		return nil, fmt.Errorf("main db connection pool is nil")
	}
	rawQuery := "SELECT stdQ.Question_Id, rf.Question_Content, stdQ.CategoryId  FROM vipshop_removeFeedbackQuestion AS rf LEFT JOIN vipshop_question AS stdQ ON stdQ.Content = rf.Question_Content"
	rows, err := db.Query(rawQuery)
	if err != nil {
		return nil, fmt.Errorf("query %s failed, %v", rawQuery, err)
	}
	defer rows.Close()
	for rows.Next() {
		var q RFQuestion
		var id sql.NullInt64
		rows.Scan(&id, &q.Content, &q.CategoryID)
		if id.Valid {
			q.IsValid = true
			q.ID = int(id.Int64)
		} else {
			q.ID = 0
			q.CategoryID = 0
		}
		questions = append(questions, q)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("scan rows failed, %v", err)
	}
	return questions, nil
}

// SetRFQuestions will reset RFQuestion table and save given content as RFQuestion.
// It will try to Update consul as well, if failed, table will be rolled back.
func SetRFQuestions(contents []string) error {

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
	_, err = tx.Exec("TRUNCATE vipshop_removeFeedbackQuestion")
	if err != nil {
		return fmt.Errorf("truncate RFQuestions Table failed, %v", err)
	}
	if len(contents) > 0 {
		rawQuery := "INSERT INTO vipshop_removeFeedbackQuestion(Question_Content) VALUES(?)" + strings.Repeat(",(?)", len(contents)-1)
		var parameters = make([]interface{}, len(contents))
		for i, c := range contents {
			parameters[i] = c
		}
		_, err = tx.Exec(rawQuery, parameters...)
		if err != nil {
			return fmt.Errorf("insert failed, %v", err)
		}
	}

	unixTime := time.Now().UnixNano() / 1000000
	_, err = util.ConsulUpdateVal("vipshopdata/RFQuestion", unixTime)
	if err != nil {
		return fmt.Errorf("consul update failed, %v", err)
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("db commit failed, %v", err)
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
				sqlParams = append(sqlParams, "%"+condition.Keyword+"%")
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
				sqlParams = append(sqlParams, "%"+condition.Keyword+"%")
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
			param := "%" + condition.Keyword + "%"
			sqlParams = append(sqlParams, param, param, param, param)
		}

		query += " group by q.Question_Id order by q.Question_Id desc"

	} else {
		// no filter
		query = `SELECT q.Question_Id, GROUP_CONCAT(DISTINCT a.Answer_Id) as aids from vipshop_question as q
				inner join vipshop_answer as a on q.Question_Id = a.Question_Id
				group by q.Question_Id
				order by q.Question_Id desc`
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
	db := util.GetMainDB()

	query := "select q.Question_Id, q.CategoryId, q.Content, q.SQuestion_count, q.CategoryName, a.Answer_Id, a.Content as acontent, a.Content_String as aContentString, a.Answer_CMD, a.Answer_CMD_Msg, a.Not_Show_In_Relative_Q, a.Begin_Time, a.End_Time, group_concat(DISTINCT rq.RelatedQuestion SEPARATOR '%s') as RelatedQuestion, group_concat(DISTINCT dm.DynamicMenu SEPARATOR '%s') as DynamicMenu, %s"
	query = fmt.Sprintf(query, SEPARATOR, SEPARATOR, "GROUP_CONCAT(DISTINCT tag.Tag_Id) as tag_ids")

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
			question.Answers = append(question.Answers, answer)
			questions = append(questions, question)
			currentQuestion = &question
		} else {
			currentQuestion.Answers = append(currentQuestion.Answers, answer)
		}
	}

	return questions, nil
}

func Escape(target string) string {
	re := regexp.MustCompile("<img(.*)src=\"([^\"]+)\"[^>]*>")
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
	query := `select tmp_q.Question_Id, tmp_q.CategoryId, tmp_q.Content, tmp_q.SQuestion_count, fullc.CategoryName from (
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
	) as fullc on fullc.id = tmp_q.CategoryId`

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
		if len(category.Children) > 0 {
			idStr += fmt.Sprintf(",%s", GenIdStr(category.Children))
		}

		categoryCondition := fmt.Sprintf(" and vipshop_question.CategoryId in(%s)", idStr)
		query = strings.Replace(query, "#CATEGORY_CONDITION#", categoryCondition, -1)
	}

	if condition.SearchQuestion && condition.Keyword != "" {
		// replace keyword condition
		keywordCondition := " and vipshop_question.content like ?"
		newParam := append(*sqlParam, "%"+condition.Keyword+"%")
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
		newParam := append(*sqlParam, "%"+condition.Keyword+"%")
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
			tagIDs = append(tagIDs, dimensionToIdMAP[dimension])
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
