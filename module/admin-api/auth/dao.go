package auth

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"emotibot.com/emotigo/module/admin-api/util"
)

var (
	apiKeyCache = map[string]int64{}
	apiKeyApp   = map[string]string{}
)

func GetUserNames(ids []string) (ret map[string]string, err error) {
	defer func() {
		util.ShowError(err)
	}()

	db := util.GetDB(ModuleInfo.ModuleName)
	if db == nil {
		err = util.ErrDBNotInit
		return
	}

	if len(ids) == 0 {
		return
	}

	queryStr := fmt.Sprintf("SELECT uuid, user_name FROM users WHERE uuid in (?%s)", strings.Repeat(",?", len(ids)-1))
	params := make([]interface{}, len(ids))
	for idx := range ids {
		params[idx] = ids[idx]
	}
	rows, err := db.Query(queryStr, params...)
	if err != nil {
		return
	}
	defer rows.Close()

	ret = map[string]string{}
	for rows.Next() {
		key, val := "", ""
		err = rows.Scan(&key, &val)
		if err != nil {
			return
		}
		ret[key] = val
	}

	return
}

func GetAllUserNames(appid string) (ret map[string]string, err error) {
	defer func() {
		util.ShowError(err)
	}()

	db := util.GetDB(ModuleInfo.ModuleName)
	if db == nil {
		err = util.ErrDBNotInit
		return
	}

	queryStr := "SELECT enterprise FROM apps WHERE uuid = ?"
	enterprise := ""
	err = db.QueryRow(queryStr, appid).Scan(&enterprise)
	if err != nil {
		return
	}

	queryStr = "SELECT uuid, user_name FROM users WHERE enterprise = ?"
	rows, err := db.Query(queryStr, enterprise)
	if err != nil {
		return
	}
	defer rows.Close()

	ret = map[string]string{}
	for rows.Next() {
		key, val := "", ""
		err = rows.Scan(&key, &val)
		if err != nil {
			return
		}
		ret[key] = val
	}

	return
}

func GetUserID(username string) (id string, err error) {
	defer func() {
		util.ShowError(err)
	}()

	db := util.GetDB(ModuleInfo.ModuleName)
	if db == nil {
		err = util.ErrDBNotInit
		return
	}

	queryStr := fmt.Sprintf("SELECT uuid FROM users WHERE user_name = ?")
	err = db.QueryRow(queryStr, username).Scan(&id)
	return
}

func GetUserName(id string) (name string, err error) {
	defer func() {
		util.ShowError(err)
	}()

	db := util.GetDB(ModuleInfo.ModuleName)
	if db == nil {
		err = util.ErrDBNotInit
		return
	}

	queryStr := fmt.Sprintf("SELECT user_name FROM users WHERE uuid = ?")
	err = db.QueryRow(queryStr, id).Scan(&name)
	return
}

var currentTimeGetter = getCurrentTimestamp

func GetAppOwner(apiKey string) (string, string, error) {
	var err error
	defer func() {
		util.ShowError(err)
	}()

	currentTime := currentTimeGetter()
	if expire, ok := apiKeyCache[apiKey]; ok {
		if currentTime <= expire {
			if appid, ok := apiKeyApp[apiKey]; ok {
				return "", appid, nil
			}
		} else {
			return "", "", sql.ErrNoRows
		}
	}

	db := util.GetDB(ModuleInfo.ModuleName)
	if db == nil {
		err = util.ErrDBNotInit
		return "", "", err
	}

	queryStr := "SELECT enterprise, appid, expire_time FROM api_key WHERE api_key = ? AND expire_time > ?"
	row := db.QueryRow(queryStr, apiKey, currentTime)
	appid := ""
	enterprise := ""
	expire := int64(0)
	err = row.Scan(&enterprise, &appid, &expire)
	if err != nil {
		return "", "", err
	}

	apiKeyCache[apiKey] = expire
	apiKeyApp[apiKey] = appid

	return enterprise, appid, nil
}

func getCurrentTimestamp() int64 {
	return time.Now().Unix()
}

func GetSystemAdminID() ([]string, error) {
	var err error
	defer func() {
		util.ShowError(err)
	}()

	db := util.GetDB(ModuleInfo.ModuleName)
	if db == nil {
		err = util.ErrDBNotInit
		return nil, err
	}

	queryStr := "SELECT uuid FROM users WHERE enterprise = NULL order by id"
	rows, err := db.Query(queryStr)
	if err != nil {
		return nil, err
	}

	ret := []string{}
	for rows.Next() {
		uuid := ""
		err = rows.Scan(&uuid)
		if err != nil {
			return nil, err
		}
		ret = append(ret, uuid)
	}

	return ret, nil
}

func GetEnterpriseAdminOfRobot(appid string) ([]string, error) {
	var err error
	defer func() {
		util.ShowError(err)
	}()

	db := util.GetDB(ModuleInfo.ModuleName)
	if db == nil {
		err = util.ErrDBNotInit
		return nil, err
	}

	queryStr := "SELECT users.uuid FROM users, apps WHERE apps.uuid = ? AND apps.enterprise = users.enterprise"
	rows, err := db.Query(queryStr, appid)
	if err != nil {
		return nil, err
	}

	ret := []string{}
	for rows.Next() {
		uuid := ""
		err = rows.Scan(&uuid)
		if err != nil {
			return nil, err
		}
		ret = append(ret, uuid)
	}

	return ret, nil
}

func GetAllApps() ([]string, error) {
	var err error
	defer func() {
		util.ShowError(err)
	}()

	db := util.GetDB(ModuleInfo.ModuleName)
	if db == nil {
		err = util.ErrDBNotInit
		return nil, err
	}

	queryStr := "SELECT uuid FROM apps"
	rows, err := db.Query(queryStr)
	if err != nil {
		return nil, err
	}

	ret := []string{}
	for rows.Next() {
		uuid := ""
		err = rows.Scan(&uuid)
		if err != nil {
			return nil, err
		}
		ret = append(ret, uuid)
	}

	return ret, nil
}
