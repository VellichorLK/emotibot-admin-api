package Task

import (
	"database/sql"
	"encoding/json"

	"emotibot.com/emotigo/module/admin-api/util"
)

func getScenarioInfoList(appid, userid string) (scenarioInfoList []ScenarioInfo, err error) {
	defer func() {
		util.ShowError(err)
	}()

	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = util.ErrDBNotInit
		return nil, err
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

	scenarioInfoList = []ScenarioInfo{}
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
		scenarioInfoList = append(scenarioInfoList, scenarioInfo)
	}
	return scenarioInfoList, nil
}

func getTemplateScenarioInfoList() (templateScenarioInfoList []ScenarioInfo, err error) {
	defer func() {
		util.ShowError(err)
	}()

	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = util.ErrDBNotInit
		return nil, err
	}
	queryStr := `
		SELECT scenarioID, content
		FROM taskenginescenario
		WHERE public = 1`
	var rows *sql.Rows
	rows, err = mySQL.Query(queryStr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	templateScenarioInfoList = []ScenarioInfo{}
	var scenarioid, content string
	for rows.Next() {
		err = rows.Scan(&scenarioid, &content)
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
			Enable:       true,
			Version:      scenarioContent.Version,
		}
		templateScenarioInfoList = append(templateScenarioInfoList, scenarioInfo)
	}
	return templateScenarioInfoList, nil
}

func getScenario(scenarioid string) (scenario *Scenario, err error) {
	defer func() {
		util.ShowError(err)
	}()

	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = util.ErrDBNotInit
		return nil, err
	}
	queryStr := `
		SELECT *
		FROM taskenginescenario
		WHERE scenarioID = ?`
	var rows *sql.Rows
	rows, err = mySQL.Query(queryStr, scenarioid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	scenario = &Scenario{}
	if rows.Next() {
		err = rows.Scan(&scenario.ScenarioID, &scenario.UserID, &scenario.AppID, &scenario.Content, &scenario.Layout,
			&scenario.Public, &scenario.Editing, &scenario.EditingContent, &scenario.EditingLayout, &scenario.Updatetime,
			&scenario.OnOff)
		if err != nil {
			return nil, err
		}
		return scenario, nil
	}
	return nil, nil
}
