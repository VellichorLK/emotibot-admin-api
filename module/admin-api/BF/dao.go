package BF

import (
	"crypto/md5"
	"crypto/sha1"
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/pkg/logger"
)

var (
	useDB *sql.DB
)

func addUser(userid, account, password, enterprise string) error {
	checkDB()
	mySQL := useDB
	if mySQL == nil {
		return errors.New("DB not init")
	}

	queryStr := ""
	var err error
	if enterprise == "" {
		queryStr = `
			INSERT INTO api_user
			(UserId, Email, CreatedTime, Password, NickName, Type, Status, UpdatedTime, enterprise_id)
			VALUES
			(?, ?, CURRENT_TIMESTAMP(), ?, ?, 0, 1, CURRENT_TIMESTAMP(), NULL)`
		_, err = mySQL.Exec(queryStr, userid, account, password, account)
	} else {
		queryStr = `
			INSERT INTO api_user
			(UserId, Email, CreatedTime, Password, NickName, Type, Status, UpdatedTime, enterprise_id)
			VALUES
			(?, ?, CURRENT_TIMESTAMP(), ?, ?, 0, 1, CURRENT_TIMESTAMP(), ?)`
		_, err = mySQL.Exec(queryStr, userid, account, password, account, enterprise)
	}
	if err != nil {
		return err
	}
	return nil
}

func deleteUser(userid string) error {
	checkDB()
	mySQL := useDB
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
	checkDB()
	mySQL := useDB
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

func updateEnterprise(id, name string) (err error) {
	defer util.ShowError(err)
	checkDB()
	mySQL := useDB
	if mySQL == nil {
		return errors.New("DB not init")
	}

	queryStr := "UPDATE api_enterprise SET enterprise_name = ? WHERE id = ?"
	_, err = mySQL.Exec(queryStr, name, id)
	return err
}

func deleteEnterprise(id string) (err error) {
	defer util.ShowError(err)
	checkDB()
	mySQL := useDB
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
	logger.Trace.Printf("Get users of enterprise [%s]: %+v\n", id, users)

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
	checkDB()
	mySQL := useDB
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

func updateApp(appid, name string) error {
	checkDB()
	mySQL := useDB
	if mySQL == nil {
		return errors.New("DB not init")
	}

	queryStr := "UPDATE api_userkey SET PreductName = ? WHERE ApiKey = ?"
	_, err := mySQL.Exec(queryStr, name, appid)
	return err
}

func deleteApp(appid string) error {
	checkDB()
	mySQL := useDB
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
	checkDB()
	mySQL := useDB
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
	checkDB()
	mySQL := useDB
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

	logger.Trace.Printf("role: %s (%d), %+v\n", uuid, id, rightIDs)
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
	checkDB()
	mySQL := useDB
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

	logger.Trace.Printf("Delete bf role with id: %d, uuid: %s\n", id, uuid)
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
	checkDB()
	mySQL := useDB
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
	checkDB()
	mySQL := useDB
	if mySQL == nil {
		return errors.New("DB not init")
	}

	var err error
	if enterprise != "" {
		queryStr := "UPDATE api_user SET Password = ? WHERE UserId = ? AND enterprise_id = ?"
		_, err = mySQL.Exec(queryStr, password, userid, enterprise)
	} else {
		queryStr := "UPDATE api_user SET Password = ? WHERE UserId = ? AND enterprise_id IS NULL"
		_, err = mySQL.Exec(queryStr, password, userid)
	}
	if err != nil {
		return err
	}
	return nil
}

func getSSMCategories(appid string, containSoftDelete bool) (*Category, error) {
	var err error
	checkDB()
	mySQL := useDB
	if mySQL == nil {
		return nil, errors.New("DB not init")
	}

	queryStr := "SELECT id, pid, name, label FROM tbl_sq_category WHERE app_id = ? AND is_del = 0"
	if containSoftDelete {
		queryStr = "SELECT id, pid, name, label FROM tbl_sq_category WHERE app_id = ?"
	}
	rows, err := mySQL.Query(queryStr, appid)
	if err != nil {
		return nil, err
	}

	var root *Category
	categoryMap := map[int]*Category{}
	for rows.Next() {
		tmp := &Category{}
		tmp.Children = []*Category{}
		err = rows.Scan(&tmp.ID, &tmp.Parent, &tmp.Name, &tmp.CatID)
		if err != nil {
			return nil, err
		}

		categoryMap[tmp.ID] = tmp
		if tmp.Parent == 0 {
			root = tmp
		}
	}

	for _, category := range categoryMap {
		if parent, ok := categoryMap[category.Parent]; ok {
			parent.Children = append(parent.Children, category)
		}
	}

	return root, nil
}

func getSSMLabels(appid string) ([]*SSMLabel, error) {
	var err error
	checkDB()
	mySQL := useDB
	if mySQL == nil {
		return nil, errors.New("DB not init")
	}

	queryStr := "SELECT id, name, description FROM tbl_robot_tag WHERE app_id = ?"
	rows, err := mySQL.Query(queryStr, appid)
	if err != nil {
		return nil, err
	}

	ret := []*SSMLabel{}

	for rows.Next() {
		tmp := &SSMLabel{}
		err = rows.Scan(&tmp.ID, &tmp.Name, &tmp.Description)
		if err != nil {
			return nil, err
		}
		ret = append(ret, tmp)
	}
	return ret, nil
}

func checkDB() {
	if useDB == nil {
		useDB = util.GetMainDB()
	}
}

func getBFAccessToken(userid string) (string, error) {
	checkDB()
	mySQL := useDB
	if mySQL == nil {
		return "", errors.New("DB not init")
	}

	now := time.Now()
	data := []byte(fmt.Sprintf("%s%s", now.Format("2006-01-02 15:04:05"), userid))
	md5Part := fmt.Sprintf("%x", md5.Sum(data))
	sha1Part := fmt.Sprintf("%x", sha1.Sum([]byte(strconv.FormatFloat(rand.Float64(), 'f', 17, 64))))
	accessToken := fmt.Sprintf("%s-%s", md5Part, sha1Part)
	logger.Trace.Printf("Gen new access token: %s\n", accessToken)

	queryStr := "INSERT INTO tbl_user_access_token (USER_ID, access_token, expiration, create_datetime) VALUES (?, ?, ?, ?)"
	_, err := mySQL.Exec(queryStr, userid, accessToken, accessTokenExpire, now)
	if err != nil {
		return "", err
	}

	return accessToken, nil
}
