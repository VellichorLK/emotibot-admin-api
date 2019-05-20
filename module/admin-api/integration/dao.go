package integration

import (
	"errors"
	"fmt"
	"strings"

	"emotibot.com/emotigo/module/admin-api/util"
)

func getPlatformConfig(appid, platform string) (map[string]string, error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return nil, errors.New("DB not init")
	}

	queryStr := "SELECT pkey, pvalue FROM integration WHERE appid = ? AND platform = ?"
	rows, err := mySQL.Query(queryStr, appid, platform)
	if err != nil {
		return nil, err
	}

	ret := map[string]string{}
	for rows.Next() {
		key, value := "", ""
		err = rows.Scan(&key, &value)
		if err != nil {
			return nil, err
		}
		ret[key] = value
	}
	rows.Close()

	queryStr = "SELECT pkey FROM integration WHERE appid = '' AND platform = ?"
	rows, err = mySQL.Query(queryStr, platform)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		key := ""
		err = rows.Scan(&key)
		if err != nil {
			return nil, err
		}
		if _, ok := ret[key]; !ok {
			ret[key] = ""
		}
	}
	rows.Close()

	return ret, nil
}

func setPlatformConfig(appid, platform string, values map[string]string) (map[string]string, error) {
	db := util.GetMainDB()
	if db == nil {
		return nil, util.ErrDBNotInit
	}

	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer util.ClearTransition(tx)

	// Get default keys from db
	sql := `
		SELECT pkey 
		FROM integration 
		WHERE appid = ''
		AND platform = ? 
	`
	existedKeys := map[string]string{}
	rows, err := tx.Query(sql, platform)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		key := ""
		err = rows.Scan(&key)
		if err != nil {
			return nil, err
		}
		existedKeys[key] = ""
	}

	// platform default keys must set in db
	if len(existedKeys) <= 0 {
		return nil, errors.New("Not valid platform")
	}

	// Remove old setting
	sql = `DELETE FROM integration WHERE appid = ? AND platform = ?`
	_, err = tx.Exec(sql, appid, platform)
	if err != nil {
		return nil, err
	}

	// insert new values from input, but only accept key in platform default key
	sql = fmt.Sprintf(`
		INSERT INTO integration(appid, platform, pkey, pvalue)
		values (?, ?, ?, ?)%s`, strings.Repeat(",(?,?,?,?)", len(existedKeys)-1))
	params := []interface{}{}
	for key := range existedKeys {
		existedKeys[key] = values[key]
		params = append(params, appid, platform, key, values[key])
	}

	_, err = tx.Exec(sql, params...)
	if err != nil {
		return nil, err
	}

	// Finish transition
	err = tx.Commit()
	return existedKeys, err
}

func deletePlatformConfig(appid, platform string) error {
	sql := `
		DELETE FROM integration 
		WHERE appid = ? 
		AND platform = ? 
	`

	db := util.GetMainDB()
	if db == nil {
		return errors.New("DB not init")
	}
	_, err := db.Exec(sql, appid, platform)
	return err
}
