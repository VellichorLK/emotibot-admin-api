package Task

import (
	"database/sql"
	"errors"

	"emotibot.com/emotigo/module/admin-api/util"
)

var errDBNotInit = errors.New("DB not init")

// 4c8cad6375894d487327cd1e7c5d5ef4 is special init userID
// it may change to NULL in future
var specialInitUser = "4c8cad6375894d487327cd1e7c5d5ef4"

func setAllScenarioStatus(appid string, status bool) error {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return errDBNotInit
	}

	queryStr := "UPDATE taskenginescenario as s, taskengineapp as app set s.onoff = ? where s.scenarioID = app.scenarioID and app.appID = ?"
	_, err := mySQL.Exec(queryStr, status, appid)
	return err
}

func setScenarioStatus(appid string, scenarioID string, status bool) error {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return errDBNotInit
	}

	queryStr := "UPDATE taskenginescenario onoff = ? where scenarioID = ?"
	_, err := mySQL.Exec(queryStr, status, scenarioID)
	return err
}

func getMapTableList(appid, userID string) ([]string, error) {
	var err error
	defer func() {
		util.ShowError(err)
	}()
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return nil, errDBNotInit
	}

	queryStr := ""
	var rows *sql.Rows
	if userID == "templateadmin" {
		queryStr = `
			SELECT mapping_table_name
			FROM taskenginemappingtable
			WHERE (update_user IS NULL OR update_user = ?)
				AND appID IS NULL order by update_time`
		rows, err = mySQL.Query(queryStr, specialInitUser)
	} else {
		queryStr = `
			SELECT mapping_table_name
			FROM taskenginemappingtable
			WHERE appID = ? OR (update_user = ? AND appID IS NULL) order by update_time`
		rows, err = mySQL.Query(queryStr, appid, userID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ret := []string{}
	for rows.Next() {
		temp := ""
		err = rows.Scan(&temp)
		if err != nil {
			return nil, err
		}
		ret = append(ret, temp)
	}
	return ret, nil
}

func getMapTableContent(appid, userID, tableName string) (string, error) {
	var err error
	defer func() {
		util.ShowError(err)
	}()
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return "", errDBNotInit
	}

	queryStr := `
		SELECT content
		FROM taskenginemappingtable
		WHERE mapping_table_name = ? AND
			(appid = ? OR update_user = ? OR update_user = ?)`
	row := mySQL.QueryRow(queryStr, tableName, appid, userID, specialInitUser)
	if err != nil {
		return "", err
	}

	content := ""
	err = row.Scan(&content)
	if err != nil {
		return "", err
	}
	return content, nil
}

func saveMappingTable(userID, appid, fileName, content string) error {
	var err error
	defer func() {
		util.ShowError(err)
	}()
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return errDBNotInit
	}

	queryStr := `
		INSERT INTO taskenginemappingtable
		(mapping_table_name, update_time, update_user, appID, content)
		VALUES
		(?, CURRENT_TIME, ?, ?, ?)`
	_, err = mySQL.Exec(queryStr, fileName, userID, appid, content)
	return err
}

func deleteMappingTable(appid, userID, tableName string) error {
	var err error
	defer func() {
		util.ShowError(err)
	}()
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return errDBNotInit
	}

	queryStr := `
		DELETE FROM taskenginemappingtable
		WHERE mapping_table_name = ?
		AND (update_user = ? OR appID = ?)`
	_, err = mySQL.Exec(queryStr, tableName, userID, appid)
	return err
}
