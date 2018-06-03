package Robot

import (
	"database/sql"
	"errors"
	"fmt"

	"emotibot.com/emotigo/module/admin-api/util"
)

var errInvalidVersion = errors.New("invalid version")

func getDBFunction(appid string, code string, version int) (ret *Function, err error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = errors.New("DB not init")
		return
	}

	var queryStr string
	var row *sql.Row
	if version == 1 {
		queryStr = fmt.Sprintf(`
			SELECT module_name, module_name_zh, on_off, remark, intent
			FROM %s_function
			WHERE module_name = ? AND status != -1`, appid)
		row = mySQL.QueryRow(queryStr, code)
	} else if version == 2 {
		queryStr = `
			SELECT module_name, module_name_zh, on_off, remark, intent
			FROM function_switch
			WHERE module_name = ? AND appid = ? AND status != -1`
		row = mySQL.QueryRow(queryStr, code, appid)
	} else {
		err = errInvalidVersion
		return
	}

	temp := Function{}
	var active int
	err = row.Scan(&temp.Code, &temp.Name, &active, &temp.Remark, &temp.Intent)
	if err != nil {
		return
	}
	temp.Active = active == 1
	ret = &temp
	return
}

func getDBFunctions(appid string, version int) (ret []*Function, err error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = errors.New("DB not init")
		return
	}

	var queryStr string
	var rows *sql.Rows
	if version == 1 {
		queryStr = fmt.Sprintf(`
			SELECT module_name, module_name_zh, on_off, remark, intent
			FROM %s_function
			WHERE status != -1`, appid)
		rows, err = mySQL.Query(queryStr)
	} else if version == 2 {
		queryStr = `
			SELECT module_name, module_name_zh, on_off, remark, intent
			FROM function_switch
			WHERE status != -1 AND appid = ?`
		rows, err = mySQL.Query(queryStr, appid)
	} else {
		err = errInvalidVersion
	}
	if err != nil {
		return
	}
	defer rows.Close()

	ret = []*Function{}
	for rows.Next() {
		temp := Function{}
		var active int
		err = rows.Scan(&temp.Code, &temp.Name, &active, &temp.Remark, &temp.Intent)
		if err != nil {
			return
		}
		temp.Active = active == 1
		ret = append(ret, &temp)
	}
	return
}

func setDBFunctionActiveStatus(appid string, code string, active bool, version int) (ret bool, err error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = errors.New("DB not init")
		return
	}
	val := 0
	if active {
		val = 1
	}

	var queryStr string
	if version == 1 {
		queryStr = fmt.Sprintf("UPDATE %s_function SET on_off = ? WHERE module_name = ?", appid)
		_, err = mySQL.Exec(queryStr, val, code)
	} else if version == 2 {
		queryStr = "UPDATE function_switch SET on_off = ? WHERE module_name = ? AND appid = ?"
		_, err = mySQL.Exec(queryStr, val, code, appid)
	} else {
		err = errInvalidVersion
		return
	}
	if err == nil {
		ret = true
	}
	return
}

func setDBMultiFunctionActiveStatus(appid string, active map[string]bool, version int) (ret bool, err error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = errors.New("DB not init")
		return
	}

	t, err := mySQL.Begin()
	if err != nil {
		return
	}

	var queryStr = ""
	if version == 1 {
		queryStr = fmt.Sprintf("UPDATE %s_function SET on_off = ? WHERE module_name = ?", appid)
	} else if version == 2 {
		queryStr = "UPDATE function_switch SET on_off = ? WHERE module_name = ? AND appid = ?"
	} else {
		err = errInvalidVersion
		return
	}

	for code, val := range active {
		sqlVal := 0
		if val {
			sqlVal = 1
		}

		if version == 1 {
			_, err = t.Exec(queryStr, sqlVal, code)
		} else {
			_, err = t.Exec(queryStr, sqlVal, code, appid)
		}

		if err != nil {
			t.Rollback()
			return
		}
	}
	err = t.Commit()
	if err == nil {
		ret = true
	}
	return
}
