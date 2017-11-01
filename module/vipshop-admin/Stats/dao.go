package Stats

import (
	"errors"
	"fmt"
	"strings"

	"emotibot.com/emotigo/module/vipshop-admin/util"
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
		rows.Scan(&temp.UserID, &temp.UserIP, &temp.CreateTime, &temp.Module, &temp.Operation, &temp.Content, &temp.Result)
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
