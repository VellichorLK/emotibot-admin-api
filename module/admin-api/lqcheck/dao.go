package lqcheck

import (
	"emotibot.com/emotigo/module/admin-api/util"
	"errors"
)

var errDBNotInit = errors.New("DB not init")

func queryDB(sql string, params ...interface{}) (map[int]map[string]string, error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return nil, errors.New("DB not init")
	}

	rows, err := mySQL.Query(sql, params...)
	if err != nil {
		return nil, err
	}
	cols, _ := rows.Columns()
	// TODO
	//colTypes, _ := rows.ColumnTypes()
	//for _, v := range colTypes {
	//	fmt.Println(v.Name(), v.DatabaseTypeName(), v.ScanType())
	//}

	values := make([][]byte, len(cols))
	scans := make([]interface{}, len(cols))
	for i := range values {
		scans[i] = &values[i]
	}

	ret := map[int]map[string]string{}

	i := 0
	for rows.Next() {
		if err := rows.Scan(scans...); err != nil {
			return nil, err
		}

		row := map[string]string{}

		for k, v := range values {
			key := cols[k]
			row[key] = string(v)
		}
		ret[i] = row
		i++
	}
	defer rows.Close()

	return ret, nil
}

func getHealthCheckStatus(appid string) (map[int]map[string]string, error) {
	sql := `
		select * 
		from health_check_status
		where appid = ? 
		order by update_time desc 
		limit 1 
	`
	params := make([]interface{}, 1)
	params[0] = appid

	return queryDB(sql, params...)
}

func getLatestHealthCheckReport(appid string) (map[int]map[string]string, error) {
	sql := `
		select * 
		from health_check_report 
		where appid = ? 
		order by update_time desc 
		limit 1
	`
	params := make([]interface{}, 1)
	params[0] = appid

	return queryDB(sql, params...)
}

func saveReportRecord(params []interface{}) (interface{}, error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return nil, errors.New("DB not init")
	}

	sql := `
		insert into health_check_report (task_id, appid, report) values (?, ?, ?)
	`
	res, _ := mySQL.Exec(sql, params...)

	return res, nil
}
