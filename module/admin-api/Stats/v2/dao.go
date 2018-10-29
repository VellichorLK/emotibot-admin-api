package v2

import (
	"errors"
	"fmt"
	"strings"

	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/pkg/logger"
)

func getAuditList(enterprise []string, appid []string, userid *string, module []string, operation *string, start int, end int, page int, listPerPage int) ([]*AuditLog, int, error) {
	mySQL := util.GetAuditDB()
	if mySQL == nil {
		return nil, 0, errors.New("DB is not inited")
	}

	columns := []string{"enterprise", "appid", "user_id", "ip_source", "create_time", "module", "operation", "content", "result"}
	conditions := []string{}
	args := []interface{}{}

	if enterprise != nil && len(enterprise) > 0 {
		orCond := fmt.Sprintf("( enterprise = ? %s )", strings.Repeat("OR enterprise = ?", len(enterprise)-1))
		conditions = append(conditions, orCond)
		for idx := range enterprise {
			args = append(args, enterprise[idx])
		}
	}

	if appid != nil && len(appid) > 0 {
		orCond := fmt.Sprintf("( appid = ? %s )", strings.Repeat("OR appid = ?", len(appid)-1))
		conditions = append(conditions, orCond)
		for idx := range appid {
			args = append(args, appid[idx])
		}
	}

	if userid != nil {
		conditions = append(conditions, "userid = ?")
		args = append(args, *userid)
	}

	if module != nil && len(module) > 0 {
		orCond := fmt.Sprintf("( module = ? %s )", strings.Repeat("OR module = ?", len(module)-1))
		conditions = append(conditions, orCond)
		for idx := range module {
			args = append(args, module[idx])
		}
	}

	if operation != nil && *operation != "all" {
		conditions = append(conditions, "operation = ?")
		args = append(args, *operation)
	}

	conditions = append(conditions, "(UNIX_TIMESTAMP(create_time) BETWEEN ? and ?)")
	args = append(args, start)
	args = append(args, end)

	queryStr := ""
	getAll := page <= 0
	total := 0
	if getAll {
		queryStr = fmt.Sprintf("SELECT %s FROM audit_record WHERE %s order by create_time desc", strings.Join(columns, ","), strings.Join(conditions, " and "))
	} else {
		queryStr = fmt.Sprintf("SELECT %s FROM audit_record WHERE %s order by create_time desc limit ? offset ?", strings.Join(columns, ","), strings.Join(conditions, " and "))
		shift := (page - 1) * listPerPage
		args = append(args, listPerPage, shift)
	}
	logger.Trace.Printf("Query for audit: %s", queryStr)
	logger.Trace.Printf("Query param: %#v", args)

	rows, err := mySQL.Query(queryStr, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	ret := []*AuditLog{}
	for rows.Next() {
		temp := AuditLog{}
		rows.Scan(&temp.EnterpriseID, &temp.AppID, &temp.UserID, &temp.UserIP, &temp.CreateTime, &temp.Module, &temp.Operation, &temp.Content, &temp.Result)
		ret = append(ret, &temp)
	}

	if getAll {
		total = len(ret)
	} else {
		queryStr = fmt.Sprintf("SELECT count(*) FROM audit_record WHERE %s", strings.Join(conditions, " and "))
		err = mySQL.QueryRow(queryStr, args[:len(args)-2]...).Scan(&total)
		if err != nil {
			return nil, 0, err
		}
	}

	return ret, total, nil
}
