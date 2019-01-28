package dictionary

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/pkg/logger"
)

func saveWordbankRows(appid string, wordbanks []*WordBankRow) (err error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = errors.New("DB not init")
		return
	}

	t, err := mySQL.Begin()
	if err != nil {
		return
	}
	defer util.ClearTransition(t)

	queryStr := fmt.Sprintf("DELETE FROM %s_entity", appid)
	_, err = t.Exec(queryStr)
	if err != nil {
		return
	}

	queryArgs := []interface{}{}
	queryQMark := []string{}

	if len(wordbanks) > 0 {
		for _, wordbank := range wordbanks {
			queryArgs = append(queryArgs,
				wordbank.Level1, wordbank.Level2, wordbank.Level3, wordbank.Level4,
				wordbank.Name, wordbank.SimilarWords, wordbank.Answer)
			queryQMark = append(queryQMark, "(?, ?, ?, ?, ?, ?, ?)")
		}
		queryStr = fmt.Sprintf(`
			INSERT INTO %s_entity
			(level1, level2, level3, level4, entity_name, similar_words, answer)
			VALUES %s`, appid, strings.Join(queryQMark, ","))
		_, err = t.Exec(queryStr, queryArgs...)
		if err != nil {
			return
		}
	}
	err = t.Commit()
	return
}

func insertImportProcess(appid string, filename string, status bool, msg string) (err error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = errors.New("DB not init")
		return
	}
	statusStr := "success"
	if !status {
		statusStr = "fail"
	}
	queryStr := `INSERT INTO process_status
		(app_id, module, status, message, entity_file_name)
		VALUES (?, "wordbank", ?, ?, ?)`
	_, err = mySQL.Exec(queryStr, appid, statusStr, msg, filename)
	return err
}

func insertEntityFile(appid string, filename string, buf []byte) (err error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = errors.New("DB not init")
		return
	}

	queryStr := `INSERT INTO entity_files
		(appid, filename, content)
		VALUES (?, ?, ?)`
	_, err = mySQL.Exec(queryStr, appid, filename, buf)
	return
}

func getWordbankFile(appid string, filename string) (buf []byte, err error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = errors.New("DB not init")
		return
	}

	queryStr := `SELECT content FROM entity_files
		WHERE appid = ? AND filename = ?`
	row := mySQL.QueryRow(queryStr, appid, filename)

	err = row.Scan(&buf)
	return
}

func saveWordbankV3Rows(appid string, root *WordBankClassV3) (err error) {
	defer func() {
		util.ShowError(err)
	}()
	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = errors.New("DB not init")
		return
	}

	t, err := mySQL.Begin()
	if err != nil {
		return
	}
	defer util.ClearTransition(t)

	queryStr := fmt.Sprintf("DELETE FROM entity_class WHERE appid = ?")
	result, err := t.Exec(queryStr, appid)
	if err != nil {
		return
	}
	c, err := result.RowsAffected()
	if err == nil {
		logger.Trace.Printf("Delete %d rows\n", c)
	}

	queryStr = fmt.Sprintf(`
		DELETE FROM entities WHERE id IN (
			SELECT id FROM (
				SELECT e.id AS id, c.id AS cid
				FROM entities as e
				LEFT JOIN entity_class AS c
				ON e.cid = c.id) AS tmp
			WHERE cid IS NULL)`)
	result, err = t.Exec(queryStr)
	if err != nil {
		return
	}
	c, err = result.RowsAffected()
	if err == nil {
		logger.Trace.Printf("Delete %d rows\n", c)
	}

	err = saveWordbankClassV3WithTransaction(appid, nil, root, t)
	if err != nil {
		return
	}
	err = t.Commit()
	return
}

// save wordbank class with DFS
func saveWordbankClassV3WithTransaction(appid string, pid *int, class *WordBankClassV3, tx *sql.Tx) error {
	if class == nil || tx == nil {
		return errors.New("Param of function has error")
	}
	var selfID *int
	var err error
	defer func() {
		util.ShowError(err)
	}()

	// Insert itself to get ID
	if class.ID != -1 {
		// ID -1 is root class, no need to insert
		selfID, err = insertClassV3WithTransaction(appid, pid, class, tx)
		if err != nil {
			return err
		}
	}

	// Use selfID as cid to insert wordbanks
	for _, wordbank := range class.Wordbank {
		err = insertWordbankV3WithTransaction(appid, selfID, wordbank, tx)
		if err != nil {
			return err
		}
	}

	// Use selfID as pid to insert children classes
	for _, child := range class.Children {
		err = saveWordbankClassV3WithTransaction(appid, selfID, child, tx)
		if err != nil {
			return err
		}
	}
	return nil
}

func insertClassV3WithTransaction(appid string, pid *int, class *WordBankClassV3, tx *sql.Tx) (*int, error) {
	queryStr := `
		INSERT INTO entity_class (appid, name, pid, editable, intent_engine, rule_engine)
		VALUES (?, ?, ?, ?, ?, ?)`
	editable := 0
	if class.Editable {
		editable = 1
	}
	ie := 0
	re := 0
	if class.IntentEngine {
		ie = 1
	}
	if class.RuleEngine {
		re = 1
	}
	result, err := tx.Exec(queryStr, appid, class.Name, pid, editable, ie, re)
	if err != nil {
		return nil, err
	}
	id64, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	id32 := int(id64)
	return &id32, nil
}

func insertWordbankV3WithTransaction(appid string, pid *int, wordbank *WordBankV3, tx *sql.Tx) error {
	queryStr := `
		INSERT INTO entities (appid, name, editable, cid, similar_words, answer)
		VALUES (?, ?, ?, ?, ?, ?)`
	editable := 0
	if wordbank.Editable {
		editable = 1
	}
	_, err := tx.Exec(queryStr,
		appid, wordbank.Name, editable, pid, strings.Join(wordbank.SimilarWords, ","), wordbank.Answer)
	return err
}
