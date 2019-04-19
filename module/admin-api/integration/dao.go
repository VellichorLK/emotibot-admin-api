package integration

import (
	"errors"

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
		ret[key] = ""
	}

	return ret, nil
}
