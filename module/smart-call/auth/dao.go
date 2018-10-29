package auth

import (
	"fmt"
	"strings"

	"emotibot.com/emotigo/module/admin-api/util"
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
