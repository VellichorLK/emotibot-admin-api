package Task

import (
	"database/sql"
	"encoding/json"

	"emotibot.com/emotigo/module/admin-api/util"
)

func getScenarioInfoList(appid, userid string) (scenarilInfoList []ScenarioInfo, err error) {
	defer func() {
		util.ShowError(err)
	}()

	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = util.ErrDBNotInit
		return
	}
	queryStr := `
		SELECT a.scenarioID, a.content, b.pk
		FROM taskenginescenario AS a left join taskengineapp AS b on a.scenarioID=b.scenarioID
		WHERE a.appID = ? or (a.appID IS NULL and a.userID = ?)`
	var rows *sql.Rows
	rows, err = mySQL.Query(queryStr, appid, userid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	scenarilInfoList = []ScenarioInfo{}
	var scenarioid, content string
	var pk sql.NullString
	for rows.Next() {
		err = rows.Scan(&scenarioid, &content, &pk)
		if err != nil {
			return nil, err
		}
		scenarioContent := ScenarioContent{}
		err = json.Unmarshal([]byte(content), &scenarioContent)
		if err != nil {
			return nil, err
		}

		scenarioInfo := ScenarioInfo{
			ScenarioID:   scenarioContent.Metadata.ScenarioID,
			ScenarioName: scenarioContent.Metadata.ScenarioName,
			Enable:       pk.Valid,
			Version:      scenarioContent.Version,
		}
		scenarilInfoList = append(scenarilInfoList, scenarioInfo)
	}
	return scenarilInfoList, nil
}
