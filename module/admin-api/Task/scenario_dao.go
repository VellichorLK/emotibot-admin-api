package Task

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"emotibot.com/emotigo/module/admin-api/util"
)

func getScenarioInfoList(appid, userid string) (scenarioInfoList []*ScenarioInfo, err error) {
	defer func() {
		util.ShowError(err)
	}()

	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = util.ErrDBNotInit
		return nil, err
	}
	queryStr := `
		SELECT a.scenarioID, a.editingContent, b.pk
		FROM taskenginescenario AS a left join taskengineapp AS b on a.scenarioID=b.scenarioID
		WHERE a.appID = ? or (a.appID IS NULL and a.userID = ?)`
	var rows *sql.Rows
	rows, err = mySQL.Query(queryStr, appid, userid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	scenarioInfoList = []*ScenarioInfo{}
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

		scenarioInfo := &ScenarioInfo{
			ScenarioID:   scenarioContent.Metadata.ScenarioID,
			ScenarioName: scenarioContent.Metadata.ScenarioName,
			Enable:       pk.Valid,
			Version:      scenarioContent.Version,
		}
		scenarioInfoList = append(scenarioInfoList, scenarioInfo)
	}
	return scenarioInfoList, nil
}

func getTemplateScenarioInfoList() (templateScenarioInfoList []*ScenarioInfo, err error) {
	defer func() {
		util.ShowError(err)
	}()

	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = util.ErrDBNotInit
		return nil, err
	}
	queryStr := `
		SELECT scenarioID, editingContent
		FROM taskenginescenario
		WHERE public = 1`
	var rows *sql.Rows
	rows, err = mySQL.Query(queryStr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	templateScenarioInfoList = []*ScenarioInfo{}
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

		scenarioInfo := &ScenarioInfo{
			ScenarioID:   scenarioContent.Metadata.ScenarioID,
			ScenarioName: scenarioContent.Metadata.ScenarioName,
			Enable:       true,
			Version:      scenarioContent.Version,
		}
		templateScenarioInfoList = append(templateScenarioInfoList, scenarioInfo)
	}
	return templateScenarioInfoList, nil
}

func createScenario(scenarioid, userid, appid, content, layout string, public,
	editing int, editingContent, editingLayout string, onoff int) (err error) {
	defer func() {
		util.ShowError(err)
	}()

	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = util.ErrDBNotInit
		return err
	}
	tx, err := mySQL.Begin()
	if err != nil {
		return err
	}
	defer util.ClearTransition(tx)

	queryStr := `
		INSERT INTO taskenginescenario
		(scenarioID, userID, appID, content, layout, public, editing, editingContent, editingLayout, updatetime, onoff)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), ?)`
	_, err = tx.Exec(queryStr, scenarioid, userid, appid, content, layout, public, editing, editingContent, editingLayout, onoff)
	if err != nil {
		return err
	}
	return tx.Commit()
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

func updateScenario(scenarioid, appid, userid, editingContent, editingLayout string) (err error) {
	defer func() {
		util.ShowError(err)
	}()

	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = util.ErrDBNotInit
		return err
	}
	tx, err := mySQL.Begin()
	if err != nil {
		return err
	}
	defer util.ClearTransition(tx)

	queryStr := `
		UPDATE taskenginescenario
		SET userID=?, editing=1, editingContent=?, editingLayout=?, updatetime=NOW()
		WHERE scenarioID=? AND appID=?`
	_, err = tx.Exec(queryStr, userid, editingContent, editingLayout, scenarioid, appid)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func getTaskEngineIntentList(appID, scenarioID string) (intents []string, err error) {
	defer func() {
		util.ShowError(err)
	}()

	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = util.ErrDBNotInit
		return nil, err
	}
	queryStr := `
		SELECT name
		FROM taskengine_intents
		WHERE app_id = ? AND scenario_id = ?`
	var rows *sql.Rows
	rows, err = mySQL.Query(queryStr, appID, scenarioID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	intents = []string{}
	for rows.Next() {
		var intent string
		err = rows.Scan(&intent)
		if err != nil {
			return nil, err
		}
		intents = append(intents, intent)
	}
	return intents, nil
}

func saveIntents(appID, scenarioID string, intents *[]string) (err error) {
	defer func() {
		util.ShowError(err)
	}()

	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = util.ErrDBNotInit
		return err
	}
	tx, err := mySQL.Begin()
	if err != nil {
		return err
	}
	defer util.ClearTransition(tx)

	// delete old intents
	err = deleteIntentsWithTransaction(appID, scenarioID, tx)
	if err != nil {
		return err
	}

	// save new intents
	for _, intent := range *intents {
		err = insertIntentWithTransaction(appID, scenarioID, intent, tx)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func deleteIntentsWithTransaction(appID, scenarioID string, tx *sql.Tx) (err error) {
	queryStr := `
		DELETE
		FROM taskengine_intents
		WHERE app_id = ? AND scenario_id = ?`
	_, err = tx.Exec(queryStr, appID, scenarioID)
	return err
}

func insertIntentWithTransaction(appID, scenarioID, intentName string, tx *sql.Tx) (err error) {
	queryStr := `
		INSERT INTO taskengine_intents (app_id, scenario_id, name)
		VALUES (?, ?, ?)`
	_, err = tx.Exec(queryStr, appID, scenarioID, intentName)
	return err
}

func deleteScenario(scenarioid, appid string) (err error) {
	defer func() {
		util.ShowError(err)
	}()

	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = util.ErrDBNotInit
		return err
	}
	tx, err := mySQL.Begin()
	if err != nil {
		return err
	}
	defer util.ClearTransition(tx)

	queryStr := `
		DELETE
		FROM taskenginescenario
		WHERE scenarioID=? AND appID=?`
	_, err = tx.Exec(queryStr, scenarioid, appid)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func getAppScenarioList(appid string) (scenarioids []string, err error) {
	defer func() {
		util.ShowError(err)
	}()

	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = util.ErrDBNotInit
		return nil, err
	}
	queryStr := `
		SELECT scenarioID
		FROM taskengineapp
		WHERE appID = ?`
	var rows *sql.Rows
	rows, err = mySQL.Query(queryStr, appid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	scenarioids = []string{}
	for rows.Next() {
		var scenarioid string
		err = rows.Scan(&scenarioid)
		if err != nil {
			return nil, err
		}
		scenarioids = append(scenarioids, scenarioid)
	}
	return scenarioids, nil
}

func createAppScenario(scenarioid string, appid string) (err error) {
	defer func() {
		util.ShowError(err)
	}()

	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = util.ErrDBNotInit
		return err
	}
	tx, err := mySQL.Begin()
	if err != nil {
		return err
	}
	defer util.ClearTransition(tx)

	queryStr := `
		INSERT INTO taskengineapp (pk, appID, scenarioID)
		SELECT ?, ?, ?
		WHERE NOT EXISTS (SELECT * FROM taskengineapp WHERE pk=? AND appId=? AND scenarioID=?)`
	pk := fmt.Sprintf("%s%s", appid, scenarioid)
	_, err = tx.Exec(queryStr, pk, appid, scenarioid, pk, appid, scenarioid)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func deleteAppScenario(scenarioid string, appid string) (err error) {
	defer func() {
		util.ShowError(err)
	}()

	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = util.ErrDBNotInit
		return err
	}
	tx, err := mySQL.Begin()
	if err != nil {
		return err
	}
	defer util.ClearTransition(tx)

	queryStr := `
		DELETE
		FROM taskengineapp
		WHERE scenarioID=? AND appID=?`
	_, err = tx.Exec(queryStr, scenarioid, appid)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func publishScenario(scenarioid, appid, userid string) (err error) {
	defer func() {
		util.ShowError(err)
	}()

	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = util.ErrDBNotInit
		return err
	}
	tx, err := mySQL.Begin()
	if err != nil {
		return err
	}
	defer util.ClearTransition(tx)

	queryStr := `
		UPDATE taskenginescenario
		SET userID=?, editing=0, content=editingContent, layout=editingLayout, updatetime=NOW()
		WHERE scenarioID=? AND appID=?`
	_, err = tx.Exec(queryStr, userid, scenarioid, appid)
	if err != nil {
		return err
	}
	return tx.Commit()
}
