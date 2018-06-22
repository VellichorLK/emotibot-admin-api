package BF

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"emotibot.com/emotigo/module/admin-api/util"
)

func addUser(userid, account, password, enterprise string) error {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return errors.New("DB not init")
	}

	queryStr := `
		INSERT INTO api_user
		(UserId, Email, CreatedTime, Password, NickName, Type, Status, UpdatedTime, enterprise_id)
		VALUES
		(?, ?, CURRENT_TIMESTAMP(), ?, ?, 0, 1, CURRENT_TIMESTAMP(), ?)`
	_, err := mySQL.Exec(queryStr, userid, account, password, account, enterprise)
	if err != nil {
		return err
	}
	return nil
}

func deleteUser(userid string) error {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return errors.New("DB not init")
	}

	queryStr := `
		DELETE FROM api_user
		WHERE UserId = ?`
	_, err := mySQL.Exec(queryStr, userid)
	if err != nil {
		return err
	}
	return nil
}

func addEnterprise(id, name string) error {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return errors.New("DB not init")
	}

	queryStr := `
		INSERT INTO api_enterprise
		(id, enterprise_name, account_type, account_status, create_time, modify_time, enterprise_type)
		VALUES
		(?, ?, 2, 0, CURRENT_TIMESTAMP(), CURRENT_TIMESTAMP(), 1)`
	_, err := mySQL.Exec(queryStr, id, name)
	if err != nil {
		return err
	}
	return nil
}

func deleteEnterprise(id string) error {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return errors.New("DB not init")
	}

	t, err := mySQL.Begin()
	if err != nil {
		return err
	}
	defer util.ClearTransition(t)

	queryStr := `
		SELECT UserId FROM api_user
		WHERE enterprise_id = ?`
	rows, err := t.Query(queryStr, id)
	if err != nil {
		return err
	}
	defer rows.Close()

	users := []interface{}{}
	qMarks := []string{}
	for rows.Next() {
		userid := ""
		err = rows.Scan(&userid)
		if err != nil {
			return err
		}
		users = append(users, userid)
		qMarks = append(qMarks, "?")
	}

	if len(users) > 0 {
		queryStr = fmt.Sprintf("DELETE FROM api_userkey WHERE UserId in (%s)",
			strings.Join(qMarks, ","))
		_, err = t.Exec(queryStr, users...)
		if err != nil {
			return err
		}
	}

	queryStr = "DELETE FROM api_user WHERE enterprise_id = ?"
	_, err = t.Exec(queryStr, id)
	if err != nil {
		return err
	}

	queryStr = "DELETE FROM api_enterprise WHERE id = ?"
	_, err = t.Exec(queryStr, id)
	if err != nil {
		return err
	}

	err = t.Commit()
	return err
}

func addApp(appid, userid, name string) error {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return errors.New("DB not init")
	}

	queryStr := `
		INSERT INTO api_userkey
		(UserId, Count, Version, CreatedTime, PreductName, ApiKey, Status)
		VALUES
		(?, 0, 0, CURRENT_TIMESTAMP(), ?, ?, 1)`
	_, err := mySQL.Exec(queryStr, userid, name, appid)
	if err != nil {
		return err
	}
	return nil
}

func deleteApp(appid string) error {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return errors.New("DB not init")
	}

	queryStr := "DELETE FROM api_userkey WHERE ApiKey = ?"
	_, err := mySQL.Exec(queryStr, appid)
	if err != nil {
		return err
	}
	return nil
}

var cmdMap = map[string][]int{
	"edit":   []int{6, 8, 9, 10, 11},
	"export": []int{13, 29, 30},
	"import": []int{12, 27, 28},
}

func addRole(uuid string, commands []string) error {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return errors.New("DB not init")
	}
	var err error
	defer func() {
		util.ShowError(err)
	}()

	rightIDs := []int{}
	for _, cmd := range commands {
		if idList, ok := cmdMap[cmd]; ok {
			rightIDs = append(rightIDs, idList...)
		}
	}

	t, err := mySQL.Begin()
	if err != nil {
		return err
	}
	defer util.ClearTransition(t)

	queryStr := "INSERT INTO ent_role (NAME) VALUES (?)"
	ret, err := mySQL.Exec(queryStr, uuid)
	if err != nil {
		return err
	}

	id64, err := ret.LastInsertId()
	if err != nil {
		return err
	}

	queryStr = "INSERT tbl_role_right (ROLE_ID ,RIGHT_ID) VALUES (?, ?)"
	for _, rightID := range rightIDs {
		_, err = t.Exec(queryStr, id64, rightID)
		if err != nil {
			return err
		}
	}

	return t.Commit()
}

func updateRole(uuid string, commands []string) error {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return errors.New("DB not init")
	}

	rightIDs := []int{}
	for _, cmd := range commands {
		if idList, ok := cmdMap[cmd]; ok {
			rightIDs = append(rightIDs, idList...)
		}
	}

	t, err := mySQL.Begin()
	if err != nil {
		return err
	}
	defer util.ClearTransition(t)

	queryStr := "SELECT ID FROM ent_role WHERE NAME = ?"
	row := t.QueryRow(queryStr, uuid)
	id := -1

	err = row.Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}

	queryStr = "DELETE FROM tbl_role_right WHERE ROLE_ID = ?"
	_, err = t.Exec(queryStr, id)
	if err != nil {
		return err
	}

	util.LogTrace.Printf("role: %s (%d), %+v\n", uuid, id, rightIDs)
	queryStr = "INSERT tbl_role_right (ROLE_ID ,RIGHT_ID) VALUES (?, ?)"
	for _, rightID := range rightIDs {
		_, err = t.Exec(queryStr, id, rightID)
		if err != nil {
			return err
		}
	}

	return t.Commit()
}

func deleteRole(uuid string) error {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return errors.New("DB not init")
	}

	t, err := mySQL.Begin()
	if err != nil {
		return err
	}
	defer util.ClearTransition(t)

	queryStr := "SELECT ID FROM ent_role WHERE NAME = ?"
	row := t.QueryRow(queryStr, uuid)
	id := -1

	err = row.Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}

	util.LogTrace.Printf("Delete bf role with id: %d, uuid: %s\n", id, uuid)
	queryStr = "DELETE FROM tbl_role_right WHERE ROLE_ID = ?"
	_, err = t.Exec(queryStr, id)
	if err != nil {
		return err
	}

	queryStr = "DELETE FROM ent_role WHERE NAME = ?"
	_, err = t.Exec(queryStr, uuid)
	if err != nil {
		return err
	}

	return t.Commit()
}

func updateUserRole(enterprise, userid, roleid string) error {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return errors.New("DB not init")
	}
	var err error
	defer func() {
		util.ShowError(err)
	}()

	t, err := mySQL.Begin()
	if err != nil {
		return err
	}

	queryStr := "SELECT ID FROM ent_role WHERE NAME = ?"
	row := t.QueryRow(queryStr, roleid)
	id := 0
	err = row.Scan(&id)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	// If role doesn't exist in BF, set role to default
	if err == sql.ErrNoRows {
		err = nil
		id = 0
	}

	queryStr = "UPDATE api_user SET RoleId = ? WHERE UserId = ? AND enterprise_id = ?"
	_, err = t.Exec(queryStr, id, userid, enterprise)
	if err != nil {
		return err
	}

	return t.Commit()
}

func updateUserPassword(enterprise, userid, password string) error {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return errors.New("DB not init")
	}

	queryStr := "UPDATE api_user SET Password = ? WHERE UserId = ? AND enterprise_id = ?"
	_, err := mySQL.Exec(queryStr, password, userid, enterprise)
	if err != nil {
		return err
	}
	return nil
}
