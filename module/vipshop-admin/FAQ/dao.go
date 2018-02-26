package FAQ

import (
	"strings"
	"database/sql"
	"fmt"
	"regexp"
	"strconv"

	"emotibot.com/emotigo/module/vipshop-admin/util"
)

var SEPARATOR = "#SEPARATE_TOKEN#"

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
	rows, err := db.Query(rawQuery, content)
	if err != nil {
		return q, fmt.Errorf("sql query %s failed, %v", rawQuery, err)
	}
	defer rows.Close()
	if rows.Next() {
		rows.Scan(&q.QuestionID, &q.Content)
	} else { //404 Not Found
		return q, util.ErrSQLRowNotFound
	}

	if err = rows.Err(); err != nil {
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

func FilterQuestionIDs(condition QueryCondition, appid string) ([]int, error) {
	db := util.GetMainDB()
	var qids []int
	searchKeyword :=!condition.SearchAnswer && !condition.SearchDynamicMenu && !condition.SearchRelativeQuestion && !condition.SearchQuestion
	if !condition.NotShow && condition.Keyword == "" && searchKeyword && !condition.TimeSet && condition.CategoryId == 0 && len(condition.Dimension) == 0 {
		// basically no filter
		sql := fmt.Sprintf("select q.Question_Id from %s_question as q where q.status > -1 order by q.Question_Id desc", appid)
		rows, err := db.Query(sql)
		if err != nil {
			return qids, err
		}
		defer rows.Close()

		for rows.Next() {
			var qid int
			rows.Scan(&qid)
			qids = append(qids, qid)
		}

		return qids, nil
	}
	return qids, nil
}

func FetchQuestions(condition QueryCondition, qids []int, appid string) ([]Question, error) {
	db := util.GetMainDB()

	sql := "select q.Question_Id, q.CategoryId, q.Content, q.SQuestion_count, q.CategoryName, a.Answer_Id, a.Content as acontent, a.Content_String as aContentString, a.Answer_CMD, a.Answer_CMD_Msg, a.Not_Show_In_Relative_Q, a.Begin_Time, a.End_Time, group_concat(DISTINCT rq.RelatedQuestion SEPARATOR '%s') as RelatedQuestion, group_concat(DISTINCT dm.DynamicMenu SEPARATOR '%s') as DynamicMenu, %s"
	if len(condition.Dimension) == 0 {
		sql = fmt.Sprintf(sql, SEPARATOR, SEPARATOR, "GROUP_CONCAT(DISTINCT tag.Tag_Id) as tag_ids")
	} else {
		sql = fmt.Sprintf(sql, SEPARATOR, SEPARATOR, "tag_ids")
	}
	sql += fmt.Sprintf(" from (%s) as q", questionSQL(condition, qids, appid))

	sql += fmt.Sprintf(" inner join (%s) as a on q.Question_Id = a.Question_Id", answerSQL(condition, appid))

	// dynamic menu
	if condition.SearchDynamicMenu {
		// TODO: handle dynamic menu condition here
	} else {
		sql += fmt.Sprintf(" left join %s_dynamic_menu as dm on dm.Answer_id = a.Answer_Id", appid)
	}

	// relateive questions
	if condition.SearchRelativeQuestion {
		// TODO: relative question condition here
	} else {
		sql += fmt.Sprintf(" left join %s_related_question as rq on rq.Answer_id = a.Answer_Id", appid)
	}

	// dimension
	if len(condition.Dimension) == 0 {
		sql += fmt.Sprintf(" left join (%s) as tag on tag.a_id = a.Answer_Id", dimensionSQL(condition, appid))
	} else {
		sql += fmt.Sprintf(" inner join (%s) as tag on tag.a_id = a.Answer_Id", dimensionSQL(condition, appid))
	}

	sql += " group by a.Answer_Id order by q.Question_Id desc, a.Answer_Id"

	// fetch
	rows, err := db.Query(sql)
	var questions []Question
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
		var rq string
		var dm string
		var tagIDs string
		var answerString string
		rows.Scan(&question.QuestionId, &question.CategoryId, &question.Content, &question.SQuestionConunt, &question.CategoryName, &answer.AnswerId, &answer.Content, &answerString, &answer.AnswerCmd, &answer.AnswerCmdMsg, &answer.NotShow, &answer.BeginTime, &answer.EndTime, &rq, &dm, &tagIDs)

		// encode answer content
		answer.Content = Escape(answer.Content)

		// transform tag id format
		if tagIDs == "" {
			answer.Dimension = []string {"", "", "", "", ""}
		} else {
			answer.Dimension = FormDimension(strings.Split(tagIDs, ","), tagMap)
		}

		if currentQuestion == nil || currentQuestion.QuestionId != question.QuestionId {
			question.Answers = append(question.Answers, answer)
			questions = append(questions, question)
			currentQuestion = &question
		} else {
			currentQuestion.Answers = append(currentQuestion.Answers, answer)
		}
		util.LogError.Printf("questions %+v", question)
	}

	return questions, nil
}

func Escape(target string) string {
	re := regexp.MustCompile("<img(.*)src=\"([^\"]+)\"[^>]*>")
	return re.ReplaceAllString(target, "[图片]")
}

func FormDimension(tagIDs []string, tagMap map[string]Tag) []string {
	// get tag string to type
	tags := []string {"", "", "", "", ""}
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

func questionSQL(condition QueryCondition, qids []int, appid string) string {
	sql := `select tmp_q.Question_Id, tmp_q.CategoryId, tmp_q.Content, tmp_q.SQuestion_count, fullc.CategoryName from (
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

	sql = fmt.Sprintf(sql, appid, appid, appid, appid, appid, appid, appid, appid, appid, appid)

	if len(qids) == 0 {
		sql = strings.Replace(sql, "#QUESITION_CONDTION#", "", -1)
	} else {
		// TODO: replace quetion condition here
		idStr := ""
		for i, qid := range qids {
			if i == 0 {
				idStr += strconv.Itoa(qid)
			} else {
				idStr += fmt.Sprintf(",%s", strconv.Itoa(qid))
			}
		}
		questionCondition := fmt.Sprintf(" and %s_question.Question_Id in(%s)", appid, idStr)
		sql = strings.Replace(sql, "#QUESITION_CONDTION#", questionCondition, -1)
	}

	if condition.CategoryId == 0 {
		sql = strings.Replace(sql, "#CATEGORY_CONDITION#", "", -1)
	} else {
		// TODO: fectch parent categorires & replace category condition here
	}

	if condition.SearchQuestion && condition.Keyword != "" {
		// TODO: replace keyword condition here
	} else {
		sql = strings.Replace(sql, "#KEYWORD_CONDITION#", "", -1)
	}

	return sql
}

func answerSQL(condition QueryCondition, appid string) string {
	sql := `select tmp_a.Answer_Id, tmp_a.Content, tmp_a.Content_String, tmp_a.Answer_CMD, tmp_a.Answer_CMD_Msg, tmp_a.Not_Show_In_Relative_Q, tmp_a.Begin_Time, tmp_a.End_Time, tmp_a.Question_Id from %s_answer as tmp_a
			#TIME_CONDITION#
			#KEYWORD_CONDTION#
			#NOT_SHOW_CONDITION#`

	sql = fmt.Sprintf(sql, appid)

	if condition.TimeSet {
		// TODO: replace time condition
	} else {
		sql = strings.Replace(sql, "#TIME_CONDITION#", "", -1)
	}

	if condition.SearchAnswer && condition.Keyword != "" {
		// TODO: replace keword condition
	} else {
		sql = strings.Replace(sql, "#KEYWORD_CONDTION#", "", -1)
	}

	if condition.NotShow {
		// TODO: replace not show condition
	} else {
		sql = strings.Replace(sql, "#NOT_SHOW_CONDITION#", "", -1)
	}
	return sql
}

func dimensionSQL(condition QueryCondition, appid string) string {
	// sql without condition
	sql := `select %s_answertag.Answer_Id as a_id,%s_answertag.Tag_Id, %s_tag.Tag_Type, %s_tag.Tag_Name from %s_answertag
	left join %s_tag on %s_tag.Tag_Id = %s_answertag.Tag_Id
	left join %s_tag_type on %s_tag_type.Type_id = %s_tag.Tag_Type`

	if len(condition.Dimension) == 0 {
		appids := []interface{}{appid, appid, appid, appid, appid, appid, appid, appid, appid, appid, appid}
		return fmt.Sprintf(sql, appids...)
	}

	sql = `select Answer_Id as a_id, tag_ids from (
		SELECT answer_id, GROUP_CONCAT(DISTINCT ans_tag.Tag_Id) as tag_ids from (SELECT answer_id, Tag_Type, anst.Tag_Id
		FROM   %s_answertag as anst, %s_tag as tag
		WHERE  anst.tag_id IN ( %s ) and anst.tag_id = tag.Tag_Id
		GROUP  BY answer_id, Tag_Type) as ans_tag group by answer_id having count(*) = %d
	) as tmp_tags`

	// TODO create tag id string
	sql = fmt.Sprintf(sql, appid, appid, "tag_id_string", len(condition.Dimension))

	return sql
}