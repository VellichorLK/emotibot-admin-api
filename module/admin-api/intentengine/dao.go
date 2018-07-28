package intentengine

import (
	"database/sql"
	"errors"
	"fmt"

	"emotibot.com/emotigo/module/admin-api/util"
)

// getIntents returns all intent names of given app ID
func getIntents(appID string, version int) ([]string, error) {
	db := util.GetMainDB()
	if db == nil {
		return nil, errors.New("DB not init")
	}

	queryStr := fmt.Sprintf("SELECT name FROM intents WHERE app_id = ? AND intent_version_id = ?")
	rows, err := db.Query(queryStr, appID, version)
	if err != nil {
		util.LogError.Printf("SQL: %s failed. %s\n", queryStr, err.Error())
		return nil, err
	}
	defer rows.Close()

	intents := make([]string, 0)

	for rows.Next() {
		var intent string

		err := rows.Scan(&intent)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}

			return nil, err
		}
		intents = append(intents, intent)
	}

	rows.Close()
	if err = rows.Err(); err != nil {
		util.LogError.Printf("SQL: %s failed. %s\n", queryStr, err.Error())
		return nil, err
	}

	return intents, nil
}

// getIntents returns all intent and their correspond sentences of given app ID
func getIntentDetails(appID string, version int) ([]*Intent, error) {
	db := util.GetMainDB()
	if db == nil {
		return nil, errors.New("DB not init")
	}

	queryStr := `
        SELECT i.intent_id, i.name, s.sentence
        FROM intents AS i
        LEFT JOIN intent_train_sets AS s
        ON i.intent_id = s.intent_id
        WHERE i.app_id = ? AND i.intent_version_id = ?
        ORDER BY i.intent_id ASC`

	rows, err := db.Query(queryStr, appID, version)
	if err != nil {
		util.LogError.Printf("SQL: %s failed. %s\n", queryStr, err.Error())
		return nil, err
	}
	defer rows.Close()

	var intents = make([]*Intent, 0)
	var intent *Intent

	for rows.Next() {
		var intentID int
		var intentName string
		var sentence sql.NullString

		err := rows.Scan(&intentID, &intentName, &sentence)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}

			return nil, err
		}

		// Compare intentID with current intent's ID:
		//   If different:
		//     1. Add current intent to intents list and initiate a
		//        new intent struct
		//	   2. Append the sentence of the row to new initiated intent's
		//		  sentences list
		//   If same:
		//     1. Append the sentence of the row to current intent's
		//		  sentences list
		if intent == nil || intent.ID != intentID {
			intentObj := NewIntent()
			intent = &intentObj
			intent.ID = intentID
			intent.AppID = appID
			intent.Name = intentName

			if sentence.Valid {
				intent.Sentences = append(intent.Sentences, sentence.String)
			}

			intents = append(intents, intent)
		} else {
			if sentence.Valid {
				intent.Sentences = append(intent.Sentences, sentence.String)
			}
		}
	}

	rows.Close()
	if err = rows.Err(); err != nil {
		util.LogError.Printf("SQL: %s failed. %s\n", queryStr, err.Error())
		return nil, err
	}

	return intents, nil
}

// insertIntents insert intents with given app ID
func insertIntents(appID string, intents map[string][]string,
	fileName string, renamedFileName string) (version int, err error) {
	db := util.GetMainDB()
	if db == nil {
		err = errors.New("DB not init")
		return
	}

	tx, _ := db.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	// Create new version, version ID is auto increment
	result, err := tx.Exec("INSERT INTO intent_versions(app_id, orig_file_name, file_name) VALUES (?, ?, ?)", appID, fileName, renamedFileName)
	if err != nil {
		util.LogError.Printf("Insert intents failed. %s\n", err.Error())
		return
	}

	ver, err := result.LastInsertId()
	if err != nil {
		util.LogError.Printf("Insert intents failed. %s\n", err.Error())
		return version, err
	}

	version = int(ver)

	intentStmt, err := tx.Prepare("INSERT INTO intents(app_id, name, intent_version_id) VALUES (?, ?, ?)")
	if err != nil {
		return
	}
	defer intentStmt.Close()

	intentTrainStmt, err := tx.Prepare("INSERT INTO intent_train_sets(sentence, intent_id, intent_version_id) VALUES (?, ?, ?)")
	if err != nil {
		return
	}
	defer intentTrainStmt.Close()

	var intentID int

	for name, sentences := range intents {
		result, err := intentStmt.Exec(appID, name, version)
		if err != nil {
			util.LogError.Printf("Insert intents failed. %s\n", err.Error())
			return version, err
		}

		id, err := result.LastInsertId()
		if err != nil {
			util.LogError.Printf("Insert intents failed. %s\n", err.Error())
			return version, err
		}

		intentID = int(id)

		for _, sentence := range sentences {
			_, err = intentTrainStmt.Exec(sentence, intentID, version)
			if err != nil {
				util.LogError.Printf("Insert intents failed. %s\n", err.Error())
				return version, err
			}
		}
	}

	return
}

// deleteIntents remove all intents of given app ID
func deleteIntents(appID string) (err error) {
	db := util.GetMainDB()
	if db == nil {
		return errors.New("DB not init")
	}

	_, err = db.Exec("DELETE FROM intents WHERE app_id = ?", appID)
	if err != nil {
		util.LogError.Printf("Delete intents failed. %s\n", err.Error())
	}
	return
}

// updateIntentEngineModelID updates intent engine's model_id of the givein app ID
func updateIntentEngineModelID(appID string, modelID []byte, version int) (err error) {
	db := util.GetMainDB()
	if db == nil {
		return errors.New("DB not init")
	}

	_, err = db.Exec("UPDATE intent_versions SET ie_model_id = ? WHERE intent_version_id = ?",
		modelID, version)
	return
}

// updateRuleEngineModelID updates rule engine's model_id of the givein app ID
func updateRuleEngineModelID(appID string, modelID []byte, version int) (err error) {
	db := util.GetMainDB()
	if db == nil {
		return errors.New("DB not init")
	}

	_, err = db.Exec("UPDATE intent_versions SET re_model_id = ? WHERE intent_version_id = ?",
		modelID, version)
	return
}

func getLatestIntentsVersion(appID string) (version int, err error) {
	db := util.GetMainDB()
	if db == nil {
		err = errors.New("DB not init")
		return
	}

	var v sql.NullInt64
	err = db.QueryRow("SELECT MAX(intent_version_id) FROM intent_versions WHERE app_id = ?", appID).Scan(&v)
	if err != nil || !v.Valid {
		version = -1
		return
	}

	version = int(v.Int64)
	return
}

func getIntentEngineModelID(appID string, version int) (modelID []byte, err error) {
	db := util.GetMainDB()
	if db == nil {
		err = errors.New("DB not init")
		return
	}

	var m sql.NullString
	err = db.QueryRow(`
		SELECT ie_model_id
		FROM intent_versions
		WHERE intent_version_id = (SELECT MAX(intent_version_id) FROM intent_versions)`).Scan(&m)
	if err != nil {
		if err == sql.ErrNoRows {
			modelID = []byte("")
		}
		return
	}

	if m.Valid {
		modelID = []byte(m.String)
	}

	return
}

func getRuleEngineModelID(appID string, version int) (modelID []byte, err error) {
	db := util.GetMainDB()
	if db == nil {
		err = errors.New("DB not init")
		return
	}

	var m sql.NullString
	err = db.QueryRow(`
		SELECT re_model_id
		FROM intent_versions
		WHERE intent_version_id = (SELECT MAX(intent_version_id) FROM intent_versions)`).Scan(&m)
	if err != nil {
		if err == sql.ErrNoRows {
			modelID = []byte("")
		}
		return
	}

	if m.Valid {
		modelID = []byte(m.String)
	}

	return
}

func getIntentsXSLXFileName(appID string, version int) (fileName []byte, origFileName []byte, err error) {
	db := util.GetMainDB()
	if db == nil {
		err = errors.New("DB not init")
		return
	}

	err = db.QueryRow("SELECT file_name, orig_file_name FROM intent_versions WHERE intent_version_id = ?", version).Scan(&fileName, &origFileName)
	return
}
