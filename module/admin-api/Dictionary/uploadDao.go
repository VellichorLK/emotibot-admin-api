package Dictionary

import (
	"errors"
	"fmt"
	"strings"

	"emotibot.com/emotigo/module/admin-api/util"
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
