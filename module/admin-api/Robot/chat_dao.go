package Robot

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"emotibot.com/emotigo/module/admin-api/util"
)

var errDuplicate = errors.New("duplicate")

func getRobotWords(appid string) (ret []*ChatInfoV2, err error) {
	defer func() {
		util.ShowError(err)
	}()

	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = errors.New("DB not init")
		return
	}
	queryStr := "SELECT type,name,comment FROM robot_words_type"
	typeRows, err := mySQL.Query(queryStr)
	if err != nil {
		return
	}
	defer typeRows.Close()

	infos := []*ChatInfoV2{}
	infoMap := map[int]*ChatInfoV2{}
	for typeRows.Next() {
		tmp := &ChatInfoV2{}
		tmp.Contents = []*ChatContentInfoV2{}
		err = typeRows.Scan(&tmp.Type, &tmp.Name, &tmp.Comment)
		if err != nil {
			return
		}
		infos = append(infos, tmp)
		infoMap[tmp.Type] = tmp
	}

	queryStr = "SELECT id,content,type FROM robot_words WHERE appid = ? ORDER BY id"
	contentRows, err := mySQL.Query(queryStr, appid)
	if err != nil {
		return
	}
	defer contentRows.Close()

	for contentRows.Next() {
		tmp := &ChatContentInfoV2{}
		var wordType int
		err = contentRows.Scan(&tmp.ID, &tmp.Content, &wordType)
		if err != nil {
			return
		}
		if info, ok := infoMap[wordType]; ok {
			info.Contents = append(info.Contents, tmp)
		}
	}
	ret = infos

	return
}

func getRobotWord(appid string, id int) (ret *ChatInfoV2, err error) {
	defer func() {
		util.ShowError(err)
	}()

	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = errors.New("DB not init")
		return
	}
	tx, err := mySQL.Begin()
	if err != nil {
		return
	}
	defer util.ClearTransition(tx)

	ret, err = getRobotWordWithTx(appid, id, tx)
	if err != nil {
		return
	}

	err = tx.Commit()
	return
}

func getRobotWordWithTx(appid string, id int, tx *sql.Tx) (ret *ChatInfoV2, err error) {
	info := &ChatInfoV2{}
	info.Contents = []*ChatContentInfoV2{}

	queryStr := "SELECT name,comment FROM robot_words_type WHERE type = ?"
	typeRow := tx.QueryRow(queryStr, id)

	err = typeRow.Scan(&info.Name, &info.Comment)
	if err != nil {
		return
	}

	queryStr = "SELECT id,content,type FROM robot_words WHERE appid = ? AND type = ?"
	contentRows, err := tx.Query(queryStr, appid, id)
	if err != nil {
		return
	}
	defer contentRows.Close()

	for contentRows.Next() {
		tmp := &ChatContentInfoV2{}
		var wordType int
		err = contentRows.Scan(&tmp.ID, &tmp.Content, &wordType)
		if err != nil {
			return
		}
		info.Contents = append(info.Contents, tmp)
	}
	ret = info
	return
}
func getRobotWordContentWithTx(appid string, typeID int, cid int, tx *sql.Tx) (ret *ChatContentInfoV2, err error) {
	queryStr := "SELECT id,content FROM robot_words WHERE appid = ? AND type = ? AND id = ?"
	contentRow := tx.QueryRow(queryStr, appid, typeID, cid)

	tmp := &ChatContentInfoV2{}
	err = contentRow.Scan(&tmp.ID, &tmp.Content)
	if err != nil {
		return
	}
	ret = tmp
	return
}

func updateRobotWord(appid string, typeID int, contents []string) (ret []*ChatContentInfoV2, err error) {
	defer func() {
		util.ShowError(err)
	}()

	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = errors.New("DB not init")
		return
	}
	tx, err := mySQL.Begin()
	if err != nil {
		return
	}
	defer util.ClearTransition(tx)

	queryStr := "DELETE FROM robot_words WHERE appid = ? AND type = ?"
	_, err = tx.Exec(queryStr, appid, typeID)
	if err != nil {
		return
	}

	if len(contents) <= 0 {
		return
	}

	questionMarkArr := make([]string, len(contents))
	for idx := range questionMarkArr {
		questionMarkArr[idx] = "(?, ?, ?)"
	}
	queryStr = fmt.Sprintf(`
		INSERT INTO robot_words
		(content, type, appid)
		VALUES %s`, strings.Join(questionMarkArr, ","))
	params := []interface{}{}
	for idx := range contents {
		params = append(params, contents[idx], typeID, appid)
	}
	_, err = tx.Exec(queryStr, params...)
	if err != nil {
		return
	}

	info, err := getRobotWordWithTx(appid, typeID, tx)
	if err != nil {
		return
	}
	ret = info.Contents
	err = tx.Commit()
	return
}

func addRobotWordContent(appid string, typeID int, content string) (ret *ChatContentInfoV2, err error) {
	defer func() {
		util.ShowError(err)
	}()

	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = errors.New("DB not init")
		return
	}
	tx, err := mySQL.Begin()
	if err != nil {
		return
	}
	defer util.ClearTransition(tx)

	queryStr := "SELECT count(*) FROM robot_words WHERE appid = ? AND type = ? AND content = ?"
	row := tx.QueryRow(queryStr, appid, typeID, content)
	count := 0
	err = row.Scan(&count)
	if err != nil {
		if err != sql.ErrNoRows {
			return
		}
		err = nil
	}
	if count > 0 {
		err = errDuplicate
		return
	}

	queryStr = "INSERT INTO robot_words (content, type, appid) VALUES (?, ?, ?)"
	result, err := tx.Exec(queryStr, content, typeID, appid)
	if err != nil {
		return
	}
	lastID, err := result.LastInsertId()
	if err != nil {
		return
	}

	ret, err = getRobotWordContentWithTx(appid, typeID, int(lastID), tx)
	if err != nil {
		return
	}

	err = tx.Commit()
	return
}

func updateRobotWordContent(appid string, typeID int, contentID int, content string) (err error) {
	defer func() {
		util.ShowError(err)
	}()

	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = errors.New("DB not init")
		return
	}
	tx, err := mySQL.Begin()
	if err != nil {
		return
	}
	defer util.ClearTransition(tx)

	queryStr := "SELECT count(*) FROM robot_words WHERE appid = ? AND type = ? AND content = ?"
	row := tx.QueryRow(queryStr, appid, typeID, content)
	count := 0
	err = row.Scan(&count)
	if err != nil {
		if err != sql.ErrNoRows {
			return
		}
		err = nil
	}
	if count > 0 {
		err = errDuplicate
		return
	}

	queryStr = "UPDATE robot_words SET content = ? WHERE appid = ? AND type = ? AND id = ?"
	_, err = tx.Exec(queryStr, content, appid, typeID, contentID)
	if err != nil {
		return
	}

	err = tx.Commit()
	return
}

func deleteRobotWordContent(appid string, typeID int, contentID int) (err error) {
	defer func() {
		util.ShowError(err)
	}()

	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = errors.New("DB not init")
		return
	}

	queryStr := "DELETE FROM robot_words WHERE appid = ? AND type = ? AND id = ?"
	_, err = mySQL.Exec(queryStr, appid, typeID, contentID)
	if err != nil {
		return
	}

	return
}
