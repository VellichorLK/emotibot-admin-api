package SelfLearning

import (
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"

	"emotibot.com/emotigo/module/vipshop-admin/util"
)

func initSelfLearningDB(url string, user string, pass string, db string) (*sql.DB, error) {
	return util.InitDB(url, user, pass, db)
}

func sqlQuery(sql string, params ...interface{}) (*sql.Rows, error) {
	db := util.GetDB(ModuleInfo.ModuleName)
	if db == nil {
		return nil, errors.New("No module(" + ModuleInfo.ModuleName + ") db connection")
	}
	return db.Query(sql, params...)
}

func getTx() (*sql.Tx, error) {
	db := util.GetDB(ModuleInfo.ModuleName)
	if db == nil {
		return nil, errors.New("No module(" + ModuleInfo.ModuleName + ") db connection")
	}
	return db.Begin()
}

func sqlExec(sql string, params ...interface{}) (sql.Result, error) {
	db := util.GetDB(ModuleInfo.ModuleName)
	if db == nil {
		return nil, errors.New("No module(" + ModuleInfo.ModuleName + ") db connection")
	}

	stmt, err := db.Prepare(sql)
	if err != nil {
		return nil, err
	}

	defer stmt.Close()
	return execStmt(stmt, params...)
}

func execStmt(stmt *sql.Stmt, params ...interface{}) (sql.Result, error) {
	return stmt.Exec(params...)
}

// GetReports search db for report based on id or limit, if id is empty then search all reports until given limit.
// If id is given then only one report will return
func GetReports(id string, limit int) (reports []Report, err error) {
	reports = []Report{}
	db := util.GetDB(ModuleInfo.ModuleName)
	if db == nil {
		err = fmt.Errorf("Could not Get the Self learn DB pool")
		return
	}

	REPORT := TableProps.report
	var wherePart string
	var parameters []interface{}
	if id != "" { // if specify
		wherePart = fmt.Sprintf("WHERE %s = ? ", REPORT.id)
		parameters = append(parameters, id)
	} else { //Only Successful ones
		wherePart = fmt.Sprintf("WHERE %s = 1", REPORT.status)
	}

	rawQuery := fmt.Sprintf("SELECT %s, %s, %s, %s FROM %s %s ORDER BY %s LIMIT %d",
		REPORT.id, REPORT.startTime, REPORT.endTime, REPORT.status,
		REPORT.name, wherePart, REPORT.startTime, limit)

	results, err := db.Query(rawQuery, parameters...)
	if err != nil {
		err = fmt.Errorf("sql query %s failed, %s", rawQuery, err)
		return
	}
	defer results.Close()

	for results.Next() {
		var r = Report{}
		results.Scan(&r.ID, &r.StartTime, &r.EndTime, &r.Status)
		reports = append(reports, r)
	}
	if err = results.Err(); err != nil {
		err = fmt.Errorf("scaning data failed, %s", err)
		return
	}

	for i, r := range reports {
		rawQuery = fmt.Sprintf("SELECT count(DISTINCT %s),count(%s) FROM %s WHERE %s = ?",
			TableProps.clusterResult.clusterID, TableProps.clusterResult.id, TableProps.clusterResult.name, TableProps.clusterResult.reportID)
		result := db.QueryRow(rawQuery, r.ID)
		if err = result.Scan(&r.ClusterSize, &r.UserQuestionSize); err != nil {
			err = fmt.Errorf("sql query %s failed, %s", rawQuery, err)
			return
		}
		reports[i] = r
	}
	return
}

// GetClusters search db for clusters based on given Report.
func GetClusters(r Report) (clusters []Cluster, err error) {
	db := util.GetDB(ModuleInfo.ModuleName)
	if db == nil {
		err = fmt.Errorf("Could not Get the Self learn DB pool")
		return
	}
	rawQuery := fmt.Sprintf("SELECT %s, count(*) FROM %s WHERE %s = ? GROUP BY %s",
		TableProps.clusterResult.clusterID, TableProps.clusterResult.name, TableProps.clusterResult.reportID, TableProps.clusterResult.clusterID)
	results, err := db.Query(rawQuery, r.ID)
	if err != nil {
		err = fmt.Errorf("sql query %s failed, %s", rawQuery, err)
		return
	}
	defer results.Close()
	dict := make(map[int]*Cluster)
	for results.Next() {
		c := Cluster{Tags: []string{}}
		results.Scan(&c.ID, &c.UserQuestionSize)
		dict[c.ID] = &c
	}
	if err = results.Err(); err != nil {
		err = fmt.Errorf("scaning data failed, %s", err)
		return
	}
	results.Close()

	rawQuery = fmt.Sprintf("SELECT %s AS tag, %s As ClusterID FROM %s WHERE %s = ?",
		TableProps.clusterTag.tag, TableProps.clusterTag.clusteringID, TableProps.clusterTag.name, TableProps.clusterTag.reportID)
	tags, err := db.Query(rawQuery, r.ID)
	if err != nil {
		err = fmt.Errorf("sql query %s failed, %s", rawQuery, err)
		return
	}
	defer tags.Close()
	for tags.Next() {
		var (
			clusterID int
			tag       string
		)
		tags.Scan(&tag, &clusterID)
		if c, ok := dict[clusterID]; ok {
			c.Tags = append(c.Tags, tag)
		} else {
			err = fmt.Errorf("report %d have a tag %s without cluster No.%d", r.ID, tag, clusterID)
			return
		}
	}
	if err = tags.Err(); err != nil {
		err = fmt.Errorf("scaning data failed, %s", err)
		return
	}

	clusters = make([]Cluster, len(dict))
	var keys []int
	for k := range dict {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	for i, k := range keys {
		clusters[i] = *dict[k]
	}
	return
}

// GetUserQuestions search DB for UserQuestions base on ReportID & ClusterID. support pagination(page start from 0).
func GetUserQuestions(reportID string, clusterID string, page int, limit int) (uQuestions []UserQuestion, err error) {
	uQuestions = []UserQuestion{}
	db := util.GetDB(ModuleInfo.ModuleName)
	if db == nil {
		err = fmt.Errorf("Could not Get the Self learn DB pool")
		return
	}

	if reportID == "" {
		err = fmt.Errorf("reportID should not be empty")
		return
	}
	f := TableProps.feedback
	r := TableProps.clusterResult
	var parameters []interface{}
	wherePart := fmt.Sprintf("WHERE r.%s = ?", r.reportID)
	parameters = append(parameters, reportID)
	if clusterID != "" {
		wherePart += fmt.Sprintf(" AND r.%s = ?", r.clusterID)
		parameters = append(parameters, clusterID)
	}

	var limitPart string
	if page == 0 && limit == 0 {
		limitPart = ""
	} else {
		limitPart = fmt.Sprintf(" LIMIT %d, %d ", page*limit, limit)
	}

	rawQuery := fmt.Sprintf("SELECT f.%s, f.%s, f.%s, f.%s, f.%s FROM %s as r INNER JOIN %s as f ON r.%s = f.%s %s ORDER BY f.%s %s",
		f.id, f.question, f.stdQuestion, f.createdTime, f.updatedTime, r.name, f.name, r.feedbackID, f.id, wherePart, TableProps.feedback.id, limitPart)
	results, err := db.Query(rawQuery, parameters...)
	if err != nil {
		err = fmt.Errorf("sql query %s failed. %s", rawQuery, err)
		return
	}
	defer results.Close()
	for results.Next() {
		var q UserQuestion
		var stdQuestion sql.NullString
		results.Scan(&q.ID, &q.Question, &stdQuestion, &q.CreatedTime, &q.UpdatedTime)
		q.StdQuestion = stdQuestion.String
		uQuestions = append(uQuestions, q)
	}
	if err = results.Err(); err != nil {
		err = fmt.Errorf("scaning data failed, %s", err)
		return
	}

	return
}

//GetUserQuestion get an UserQuestion by id.
func GetUserQuestion(id int) (UserQuestion, error) {
	var uQuestion UserQuestion
	db := util.GetDB(ModuleInfo.ModuleName)
	if db == nil {
		return uQuestion, fmt.Errorf("Could not Get the Self learn DB pool")
	}
	f := TableProps.feedback

	rawQuery := fmt.Sprintf("SELECT %s, %s, %s, %s, %s FROM %s WHERE %s = ?",
		f.id, f.question, f.stdQuestion, f.createdTime, f.updatedTime, f.name, f.id)
	result := db.QueryRow(rawQuery, id)

	err := result.Scan(&uQuestion.ID, &uQuestion.Question, &uQuestion.StdQuestion, &uQuestion.CreatedTime, &uQuestion.UpdatedTime)
	if err != nil {
		return uQuestion, fmt.Errorf("sql query %s failed. %s", rawQuery, err)
	}
	return uQuestion, nil
}

// ErrRowNotFound represent SQL query not found error
var ErrRowNotFound = errors.New("Not Found")

// ErrAlreadyOccupied represent rows already have value, should not updated it.
var ErrAlreadyOccupied = errors.New("db row already updated")

//UpdateStdQuestion update an array of feedback's ID of the stdQuestion parameter.
func UpdateStdQuestions(feedbacks []int, stdQuestion string) error {
	db := util.GetDB(ModuleInfo.ModuleName)
	if db == nil {
		return fmt.Errorf("could not get the Self learn DB pool")
	}
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("transaction start failed, %s", err)
	}

	rawQuery := fmt.Sprintf("SELECT %s FROM %s WHERE %s = ?", TableProps.feedback.stdQuestion, TableProps.feedback.name, TableProps.feedback.id)
	selectStmt, err := tx.Prepare(rawQuery)
	if err != nil {
		return fmt.Errorf("sql query %s prepare failed, %s", rawQuery, err)
	}
	defer selectStmt.Close()

	rawQuery = fmt.Sprintf("UPDATE %s SET %s = ? WHERE %s = ?", TableProps.feedback.name, TableProps.feedback.stdQuestion, TableProps.feedback.id)
	stmt, err := tx.Prepare(rawQuery)
	if err != nil {
		return fmt.Errorf("sql query %s prepare failed, %s", rawQuery, err)
	}
	defer stmt.Close()

	for _, id := range feedbacks {
		var currentStdQuestion sql.NullString
		result, err := selectStmt.Query(id)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("select query failed, %s", err)
		}
		defer result.Close()

		if !result.Next() {
			tx.Rollback()
			return ErrRowNotFound
		}
		if err = result.Scan(&currentStdQuestion); err != nil {
			tx.Rollback()
			return fmt.Errorf("scanning data failed, %s", err)
		}
		result.Close()
		if currentStdQuestion.String != "" {
			tx.Rollback()
			return ErrAlreadyOccupied
		}

		execResult, err := stmt.Exec(stdQuestion, id)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("db updated feedback failed, %s", err)
		}
		if rowEffected, err := execResult.RowsAffected(); err != nil {
			tx.Rollback()
			return fmt.Errorf("db updated feedback failed, %s", err)
		} else if rowEffected != 1 {
			tx.Rollback()
			return fmt.Errorf("didnt updated feedback id [%d]", id)
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("Transaction failed, %s", err)
	}
	return nil

}

//RevokeUserQuestion remove userQuestion assigned standard question. error contains:
//	`ErrRowNotFound`: can't found row.
func RevokeUserQuestion(id int) error {
	db := util.GetDB(ModuleInfo.ModuleName)
	if db == nil {
		return fmt.Errorf("could not get the Self learn DB pool")
	}

	rawQuery := fmt.Sprintf("SELECT %s FROM %s WHERE %s = ?",
		TableProps.feedback.stdQuestion, TableProps.feedback.name, TableProps.feedback.id)
	result, err := db.Query(rawQuery, id)
	if err != nil {
		return fmt.Errorf("sql query %s failed, %s", rawQuery, err)
	}
	defer result.Close()
	if !result.Next() {
		return ErrRowNotFound
	}
	result.Close()

	rawQuery = fmt.Sprintf("UPDATE %s SET %s = NULL WHERE %s = ?", TableProps.feedback.name, TableProps.feedback.stdQuestion, TableProps.feedback.id)
	_, err = db.Exec(rawQuery, id)
	if err != nil {
		return fmt.Errorf("sql query %s failed, %s", rawQuery, err)
	}

	return nil

}

func DeleteReport(id int) error {
	db := util.GetDB(ModuleInfo.ModuleName)
	if db == nil {
		return fmt.Errorf("could not get the Self learn DB pool")
	}

	rawQuery := fmt.Sprintf("SELECT count(%s) FROM %s WHERE %s = ?", TableProps.report.id, TableProps.report.name, TableProps.report.id)
	var num int
	err := db.QueryRow(rawQuery, id).Scan(&num)
	if err != nil {
		return fmt.Errorf("query %s failed, %v", rawQuery, err)
	}
	if num == 0 {
		return ErrRowNotFound
	}

	rawQuery = fmt.Sprintf("DELETE FROM %s WHERE %s = ?", TableProps.report.name, TableProps.report.id)
	_, err = db.Exec(rawQuery, id)
	if err != nil {
		return fmt.Errorf("delete query %s failed, %v", rawQuery, err)
	}
	return nil
}

//GetQuestionIDByContent get question id from db by question content
func GetQuestionIDByContent(content []interface{}) (map[string]int, error) {
	db := util.GetMainDB()
	if db == nil {
		return nil, errors.New("could not get the Self learn DB pool")
	}
	if len(content) == 0 {
		return make(map[string]int), nil
	}
	querySQL := "select " + NQuestionID + "," + NContent + " from " + QuestionTable +
		" where " + NContent + " in(?" + strings.Repeat(",?", len(content)) + ")"

	rows, err := db.Query(querySQL, content...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	questionIDMap := make(map[string]int)

	var question string
	var id int
	for rows.Next() {
		err := rows.Scan(&id, &question)
		if err != nil {
			return nil, err
		}
		questionIDMap[question] = id
	}
	return questionIDMap, nil
}
