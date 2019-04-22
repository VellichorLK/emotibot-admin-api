package integration

import (
	"emotibot.com/emotigo/module/admin-api/util"
	"errors"
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

	//queryStr = "SELECT pkey FROM integration WHERE appid = '' AND platform = ?"
	//rows, err = mySQL.Query(queryStr, platform)
	//if err != nil {
	//	return nil, err
	//}
	//for rows.Next() {
	//	key := ""
	//	err = rows.Scan(&key)
	//	if err != nil {
	//		return nil, err
	//	}
	//	ret[key] = ""
	//}

	return ret, nil
}

func setPlatformConfig(params map[string]interface{}) ([]map[string]interface{}, error) {
	sql := `
		SELECT appid, platform, pkey, pvalue 
		FROM integration 
		WHERE appid = ? 
		AND platform = ? 
	`
	result, err := util.DbQuery(sql, params["appid"], params["platform"])
	if err != nil {
		return nil, err
	}

	ret := make([]map[string]interface{}, 0)
	rowMap := make(map[string]map[string]interface{})
	for _, v := range result {
		rowKey := v["appid"].(string) + "_" + v["platform"].(string)

		if _, ok := rowMap[rowKey]; !ok {
			rowMap[rowKey] = make(map[string]interface{})

			rowMap[rowKey]["appid"] = v["appid"].(string)
			rowMap[rowKey]["platform"] = v["platform"].(string)
			rowMap[rowKey][v["pkey"].(string)] = v["pvalue"].(string)
		} else {
			rowMap[rowKey][v["pkey"].(string)] = v["pvalue"].(string)
		}
	}

	if len(rowMap) > 0 {
		for _, v := range rowMap {
			ret = append(ret, v)
		}
		return ret, errors.New("already bound")
	}

	sql = `
		INSERT INTO integration(appid, platform, pkey, pvalue)
		values(?, ?, ?, ?)
		,(?, ?, ?, ?)
		,(?, ?, ?, ?)
		,(?, ?, ?, ?)
		,(?, ?, ?, ?)
		,(?, ?, ?, ?)
	`
	_, err = util.DbExec(sql,
		params["appid"], params["platform"], "corp_id", params["corp_id"],
		params["appid"], params["platform"], "agent_id", params["agent_id"],
		params["appid"], params["platform"], "secret", params["secret"],
		params["appid"], params["platform"], "url", params["url"],
		params["appid"], params["platform"], "token", params["token"],
		params["appid"], params["platform"], "encoded-aes", params["encoded-aes"],
	)
	//tmp := make([]interface{}, 0)
	//util.DbExec(sql, tmp...)
	if err != nil {
		return nil, err
	}

	//ret = append(ret, res.RowsAffected())

	return ret, nil
}

func deletePlatformConfig(params map[string]interface{}) ([]map[string]interface{}, error) {
	ret := make([]map[string]interface{}, 0)

	sql := `
		DELETE FROM integration 
		WHERE appid = ? 
		AND platform = ? 
	`
	_, err := util.DbExec(sql, params["appid"], params["platform"])
	if err != nil {
		return nil, err
	}

	return ret, nil
}
