package Robot

import (
	"errors"
	"fmt"

	"emotibot.com/emotigo/module/admin-api/util"
)

func getDBFunction(appid string, code string) (ret *Function, err error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = errors.New("DB not init")
		return
	}

	queryStr := fmt.Sprintf(`
		SELECT module_name, module_name_zh, on_off, remark, intent
		FROM %s_function
		WHERE module_name = ?`, appid)
	row := mySQL.QueryRow(queryStr, code)

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

func getDBFunctions(appid string) (ret []*Function, err error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = errors.New("DB not init")
		return
	}

	queryStr := fmt.Sprintf(`
		SELECT module_name, module_name_zh, on_off, remark, intent
		FROM %s_function
		WHERE status != -1`, appid)
	rows, err := mySQL.Query(queryStr)
	if err != nil {
		return
	}

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

func setDBFunctionActiveStatus(appid string, code string, active bool) (ret bool, err error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = errors.New("DB not init")
		return
	}

	queryStr := fmt.Sprintf("UPDATE %s_function SET on_off = ? where module_name = ?", appid)
	val := 0
	if active {
		val = 1
	}
	_, err = mySQL.Exec(queryStr, val, code)
	if err == nil {
		ret = true
	}
	return
}

func setDBMultiFunctionActiveStatus(appid string, active map[string]bool) (ret bool, err error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = errors.New("DB not init")
		return
	}

	t, err := mySQL.Begin()
	if err != nil {
		return
	}

	for code, val := range active {
		queryStr := fmt.Sprintf("UPDATE %s_function SET on_off = ? where module_name = ?", appid)
		sqlVal := 0
		if val {
			sqlVal = 1
		}
		_, err = t.Exec(queryStr, sqlVal, code)
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
