package Task

import (
	"errors"

	"emotibot.com/emotigo/module/admin-api/util"
)

func setAllScenarioStatus(appid string, status bool) error {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return errors.New("DB not init")
	}

	queryStr := "UPDATE taskenginescenario as s, taskengineapp as app set s.onoff = ? where s.scenarioID = app.scenarioID and app.appID = ?"
	_, err := mySQL.Exec(queryStr, status, appid)
	return err
}

func setScenarioStatus(appid string, scenarioID string, status bool) error {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return errors.New("DB not init")
	}

	queryStr := "UPDATE taskenginescenario onoff = ? where scenarioID = ?"
	_, err := mySQL.Exec(queryStr, status, scenarioID)
	return err
}
