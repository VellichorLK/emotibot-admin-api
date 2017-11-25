package Stats

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"emotibot.com/emotigo/module/vipshop-admin/util"
)

const (
	RECORD_TABLE_FORMAT = "%s_record"
	RECORD_INFO_TABLE   = "static_record_info"
)

func getAuditList(appid string, input *AuditInput) ([]*AuditLog, error) {
	// Audit log is not splited by appid for now
	mySQL := util.GetAuditDB()
	if mySQL == nil {
		return nil, errors.New("DB is not inited")
	}

	columns := []string{"user_id", "ip_source", "create_time", "module", "operation", "content", "result"}

	conditions := []string{}
	args := []interface{}{}

	if input.Filter != nil && input.Filter.Module != "-1" {
		conditions = append(conditions, "module = ?")
		args = append(args, input.Filter.Module)
	}
	if input.Filter != nil && input.Filter.Operation != "-1" {
		conditions = append(conditions, "operation = ?")
		args = append(args, input.Filter.Operation)
	}
	if input.Filter != nil && input.Filter.UserID != "" {
		conditions = append(conditions, "user_id = ?")
		args = append(args, input.Filter.UserID)
	}

	conditions = append(conditions, "(UNIX_TIMESTAMP(create_time) BETWEEN ? and ?)")
	args = append(args, input.Start)
	args = append(args, input.End)

	queryStr := fmt.Sprintf("SELECT %s FROM audit_record WHERE %s order by create_time desc", strings.Join(columns, ","), strings.Join(conditions, " and "))
	util.LogTrace.Printf("Query for audit: %s", queryStr)
	util.LogTrace.Printf("Query param: %#v", args)

	rows, err := mySQL.Query(queryStr, args...)
	if err != nil {
		return nil, err
	}

	ret := []*AuditLog{}
	for rows.Next() {
		temp := AuditLog{}
		rows.Scan(&temp.UserID, &temp.UserIP, &temp.CreateTime, &temp.Module, &temp.Operation, &temp.Content, &temp.Result)
		ret = append(ret, &temp)
	}

	return ret, nil
}

func getAuditListPage(appid string, input *AuditInput, page int, listPerPage int) ([]*AuditLog, int, error) {
	// Audit log is not splited by appid for now
	mySQL := util.GetAuditDB()
	if mySQL == nil {
		return nil, 0, errors.New("DB is not inited")
	}

	util.LogTrace.Printf("Search for audit: %#v", input.Filter)

	columns := []string{"id", "user_id", "ip_source", "UNIX_TIMESTAMP(create_time)", "module", "operation", "content", "result"}

	conditions := []string{}
	args := []interface{}{}

	if input.Filter != nil && input.Filter.Module != "-1" {
		conditions = append(conditions, "module = ?")
		args = append(args, input.Filter.Module)
	}
	if input.Filter != nil && input.Filter.Operation != "-1" {
		conditions = append(conditions, "operation = ?")
		args = append(args, input.Filter.Operation)
	}
	if input.Filter != nil && input.Filter.UserID != "" {
		conditions = append(conditions, "user_id = ?")
		args = append(args, input.Filter.UserID)
	}

	conditions = append(conditions, "(UNIX_TIMESTAMP(create_time) BETWEEN ? and ?)")
	args = append(args, input.Start)
	args = append(args, input.End)

	shift := (page - 1) * listPerPage
	queryStr := fmt.Sprintf("SELECT %s FROM audit_record WHERE %s order by create_time desc limit ? offset ?", strings.Join(columns, ","), strings.Join(conditions, " and "))
	args = append(args, listPerPage)
	args = append(args, shift)

	util.LogTrace.Printf("Query for audit: %s", queryStr)
	util.LogTrace.Printf("Query param: %#v", args)

	rows, err := mySQL.Query(queryStr, args...)
	if err != nil {
		return nil, 0, err
	}

	ret := []*AuditLog{}
	for rows.Next() {
		temp := AuditLog{}
		var id int
		var timestamp int64
		rows.Scan(&id, &temp.UserID, &temp.UserIP, &timestamp, &temp.Module, &temp.Operation, &temp.Content, &temp.Result)
		temp.CreateTime = time.Unix(timestamp, 0)
		ret = append(ret, &temp)
	}

	cnt, err := getAuditListCnt(appid, input)
	if err != nil {
		return nil, 0, err
	}

	return ret, cnt, nil
}

func getAuditListCnt(appid string, input *AuditInput) (int, error) {
	// Audit log is not splited by appid for now
	mySQL := util.GetAuditDB()
	if mySQL == nil {
		return 0, errors.New("DB is not inited")
	}

	conditions := []string{}
	args := []interface{}{}

	if input.Filter != nil && input.Filter.Module != "-1" {
		conditions = append(conditions, "module = ?")
		args = append(args, input.Filter.Module)
	}
	if input.Filter != nil && input.Filter.Operation != "-1" {
		conditions = append(conditions, "operation = ?")
		args = append(args, input.Filter.Operation)
	}
	if input.Filter != nil && input.Filter.UserID != "" {
		conditions = append(conditions, "user_id = ?")
		args = append(args, input.Filter.UserID)
	}

	conditions = append(conditions, "(UNIX_TIMESTAMP(create_time) BETWEEN ? and ?)")
	args = append(args, input.Start)
	args = append(args, input.End)

	queryStr := fmt.Sprintf("SELECT COUNT(*) FROM audit_record WHERE %s", strings.Join(conditions, " and "))
	util.LogTrace.Printf("Query for audit: %s", queryStr)
	util.LogTrace.Printf("Query param: %#v", args)

	rows, err := mySQL.Query(queryStr, args...)
	if err != nil {
		return 0, err
	}

	var ret int
	if rows.Next() {
		err = rows.Scan(&ret)
		if err != nil {
			return 0, err
		}
	}

	return ret, nil
}

func initStatDB(url string, user string, pass string, db string) (*sql.DB, error) {
	return util.InitDB(url, user, pass, db)
}

func getUnresolveQuestionsStatistic(appid string, start int64, end int64) ([]*StatRow, error) {
	mySQL := getStatsDB()
	if mySQL == nil {
		return nil, errors.New("DB is not inited")
	}

	table := fmt.Sprintf(RECORD_TABLE_FORMAT, appid)
	queryPart := fmt.Sprintf("SELECT r.user_q, COUNT(*) as cnt, MAX(r.answer), MAX(r.score), r.std_q FROM %s AS r LEFT JOIN %s AS info USING(unique_id)", table, RECORD_INFO_TABLE)
	condition := "WHERE info.qa_solved = -1 and r.created_time between FROM_UNIXTIME(?) and FROM_UNIXTIME(?) GROUP BY r.user_q, r.std_q ORDER BY cnt DESC"

	queryStr := queryPart + " " + condition

	util.LogTrace.Printf("Query for stats unresolve question: %s, with [%d, %d]", queryStr, start, end)
	rows, err := mySQL.Query(queryStr, start, end)
	if err != nil {
		return nil, err
	}

	ret := []*StatRow{}
	for rows.Next() {
		temp := StatRow{}
		rows.Scan(&temp.UserQuery, &temp.Count, &temp.Answer, &temp.Score, &temp.StandardQuestion)
		util.LogTrace.Printf("==== %#v", temp)
		ret = append(ret, &temp)
	}
	return ret, nil
}
