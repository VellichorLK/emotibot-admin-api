package v2

import (
	"errors"
	"fmt"
	"strings"

	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/pkg/logger"
)

func getAuditList(enterprise *string, appid *string, userid *string, module *string, operation *string, start int, end int, page int, listPerPage int) ([]*AuditLog, error) {
	mySQL := util.GetAuditDB()
	if mySQL == nil {
		return nil, errors.New("DB is not inited")
	}

	columns := []string{"enterprise", "appid", "user_id", "ip_source", "create_time", "module", "operation", "content", "result"}

	conditions := []string{}
	args := []interface{}{}

	if enterprise != nil {
		conditions = append(conditions, "enterprise = ?")
		args = append(args, *enterprise)
	}

	if appid != nil {
		conditions = append(conditions, "appid = ?")
		args = append(args, *appid)
	}

	if userid != nil {
		conditions = append(conditions, "userid = ?")
		args = append(args, *userid)
	}

	if module != nil && *module != "all" {
		conditions = append(conditions, "module = ?")
		args = append(args, *userid)
	}

	if operation != nil && *operation != "all" {
		conditions = append(conditions, "operation = ?")
		args = append(args, *userid)
	}

	conditions = append(conditions, "(UNIX_TIMESTAMP(create_time) BETWEEN ? and ?)")
	args = append(args, start)
	args = append(args, end)

	queryStr := ""
	if page <= 0 {
		queryStr = fmt.Sprintf("SELECT %s FROM audit_record WHERE %s order by create_time desc", strings.Join(columns, ","), strings.Join(conditions, " and "))
	} else {
		queryStr = fmt.Sprintf("SELECT %s FROM audit_record WHERE %s order by create_time desc limit ? offset ?", strings.Join(columns, ","), strings.Join(conditions, " and "))
		shift := (page - 1) * listPerPage
		args = append(args, listPerPage)
		args = append(args, shift)
	}
	logger.Trace.Printf("Query for audit: %s", queryStr)
	logger.Trace.Printf("Query param: %#v", args)

	rows, err := mySQL.Query(queryStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ret := []*AuditLog{}
	for rows.Next() {
		temp := AuditLog{}
		rows.Scan(&temp.EnterpriseID, &temp.AppID, &temp.UserID, &temp.UserIP, &temp.CreateTime, &temp.Module, &temp.Operation, &temp.Content, &temp.Result)
		ret = append(ret, &temp)
	}

	return ret, nil
}
