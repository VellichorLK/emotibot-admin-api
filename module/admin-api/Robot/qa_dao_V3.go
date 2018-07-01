package Robot

import (
	"database/sql"

	"emotibot.com/emotigo/module/admin-api/util"
)

type scanners interface {
	Scan(dest ...interface{}) error
	Next() bool
}
type scanner interface {
	Scan(dest ...interface{}) error
}

func scanInfoMap(rows scanners) (map[int][]*InfoV3, error) {
	ret := map[int][]*InfoV3{}
	for rows.Next() {
		qid, temp := 0, InfoV3{}
		err := rows.Scan(&qid, &temp.ID, &temp.Content)
		if err != nil {
			return nil, err
		}
		if _, ok := ret[qid]; !ok {
			ret[qid] = []*InfoV3{&temp}
		} else {
			ret[qid] = append(ret[qid], &temp)
		}
	}
	return ret, nil
}

func scanInfos(rows scanners) ([]*InfoV3, error) {
	ret := []*InfoV3{}
	for rows.Next() {
		temp, err := scanInfo(rows)
		if err != nil {
			return nil, err
		}
		ret = append(ret, temp)
	}
	return ret, nil
}

func scanInfo(row scanner) (*InfoV3, error) {
	ret := InfoV3{}
	err := row.Scan(&ret.ID, &ret.Content)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

func getRobotQAListV3(appid string) (ret []*QAInfoV3, err error) {
	defer func() {
		util.ShowError(err)
	}()
	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = util.ErrDBNotInit
		return
	}

	t, err := mySQL.Begin()
	if err != nil {
		return
	}
	defer util.ClearTransition(t)

	qaList := []*QAInfoV3{}
	// Get orig questions
	queryStr := `SELECT id, content FROM robot_profile_question ORDER BY id`
	origQRows, err := t.Query(queryStr)
	if err != nil {
		return
	}
	defer origQRows.Close()

	qaMap := map[int]*QAInfoV3{}
	for origQRows.Next() {
		id, content := 0, ""
		err = origQRows.Scan(&id, &content)
		if err != nil {
			return
		}
		qaMap[id] = &QAInfoV3{
			ID:               id,
			Question:         content,
			RelatedQuestions: []*InfoV3{},
			Answers:          []*InfoV3{},
		}
		qaList = append(qaList, qaMap[id])
	}

	// Get extend questions
	queryStr = `SELECT qid, id, content FROM robot_profile_extend
		WHERE appid = ?`
	extendQRows, err := t.Query(queryStr, appid)
	if err != nil {
		return
	}
	defer extendQRows.Close()
	extendQMap, err := scanInfoMap(extendQRows)
	if err != nil {
		return
	}
	for id, val := range extendQMap {
		if _, ok := qaMap[id]; ok {
			qaMap[id].RelatedQuestions = append(qaMap[id].RelatedQuestions, val...)
		}
	}

	// Get answers
	queryStr = `SELECT qid, id, content FROM robot_profile_answer
		WHERE appid = ?`
	ansRows, err := t.Query(queryStr, appid)
	if err != nil {
		return
	}
	defer ansRows.Close()
	ansMap, err := scanInfoMap(ansRows)
	if err != nil {
		return
	}
	for id, val := range ansMap {
		if _, ok := qaMap[id]; ok {
			qaMap[id].Answers = append(qaMap[id].Answers, val...)
		}
	}

	ret = qaList

	err = t.Commit()
	return
}

func getRobotQAV3(appid string, id int) (ret *QAInfoV3, err error) {
	defer func() {
		util.ShowError(err)
	}()
	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = util.ErrDBNotInit
		return
	}

	t, err := mySQL.Begin()
	if err != nil {
		return
	}
	defer util.ClearTransition(t)

	// Get orig questions
	queryStr := `SELECT content FROM robot_profile_question 
		WHERE id = ?`
	origQRow := t.QueryRow(queryStr, id)
	if err != nil {
		return
	}
	content := ""
	err = origQRow.Scan(&content)
	if err != nil {
		return
	}
	qaInfo := &QAInfoV3{
		ID:               id,
		Question:         content,
		RelatedQuestions: []*InfoV3{},
		Answers:          []*InfoV3{},
	}

	// Get extend questions
	queryStr = `SELECT id, content FROM robot_profile_extend
		WHERE appid = ? AND qid = ?`
	extendQRows, err := t.Query(queryStr, appid, id)
	if err != nil {
		return
	}
	defer extendQRows.Close()
	extendQuestions, err := scanInfos(extendQRows)
	if err != nil {
		return
	}
	qaInfo.RelatedQuestions = extendQuestions

	// Get answers
	queryStr = `SELECT id, content FROM robot_profile_answer
		WHERE appid = ? AND qid = ?`
	ansRows, err := t.Query(queryStr, appid, id)
	if err != nil {
		return
	}
	defer ansRows.Close()
	answers, err := scanInfos(ansRows)
	if err != nil {
		return
	}
	qaInfo.Answers = answers

	ret = qaInfo
	err = t.Commit()
	return
}

func addRobotQAAnswerV3(appid string, qid int, answer string) (id int, err error) {
	defer func() {
		util.ShowError(err)
	}()
	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = util.ErrDBNotInit
		return
	}

	t, err := mySQL.Begin()
	if err != nil {
		return
	}
	defer util.ClearTransition(t)

	queryStr := `SELECT count(*) FROM robot_profile_question WHERE id = ?`
	qRow := t.QueryRow(queryStr, qid)
	count := 0
	err = qRow.Scan(&count)
	if err != nil {
		return
	}
	if count == 0 {
		err = sql.ErrNoRows
		return
	}

	queryStr = `SELECT count(*) FROM robot_profile_answer
		WHERE appid = ? AND qid = ? AND content = ?`
	row := t.QueryRow(queryStr, appid, qid, answer)
	count = 0
	err = row.Scan(&count)
	if err == sql.ErrNoRows {
		err = nil
	} else if err != nil {
		return
	}

	if count > 0 {
		err = util.ErrDuplicated
		return
	}

	queryStr = `
		INSERT INTO robot_profile_answer
		(appid, qid, content) VALUES (?, ?, ?)`
	result, err := t.Exec(queryStr, appid, qid, answer)
	if err != nil {
		return
	}

	id64, err := result.LastInsertId()
	if err != nil {
		return
	}

	id = int(id64)
	err = t.Commit()
	return
}

func getBasicQuestionV3(qid int) (ret *InfoV3, err error) {
	defer func() {
		util.ShowError(err)
	}()
	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = util.ErrDBNotInit
		return
	}

	queryStr := `SELECT id, content FROM robot_profile_question WHERE id = ?`
	row := mySQL.QueryRow(queryStr, qid)
	ret, err = scanInfo(row)
	return
}

func getRobotQAAnswerV3(appid string, qid, aid int) (ret *InfoV3, err error) {
	defer func() {
		util.ShowError(err)
	}()
	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = util.ErrDBNotInit
		return
	}

	queryStr := `SELECT id, content FROM robot_profile_answer
		WHERE appid = ? AND qid = ? AND id = ?`
	row := mySQL.QueryRow(queryStr, appid, qid, aid)
	ret, err = scanInfo(row)
	return
}

func updateRobotQAAnswerV3(appid string, qid, aid int, answer string) (err error) {
	defer func() {
		util.ShowError(err)
	}()
	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = util.ErrDBNotInit
		return
	}
	t, err := mySQL.Begin()
	if err != nil {
		return
	}
	defer util.ClearTransition(t)

	queryStr := `SELECT count(*) FROM robot_profile_answer
		WHERE appid = ? AND qid = ? AND content = ? AND id != ?`
	row := t.QueryRow(queryStr, appid, qid, answer, aid)
	count := 0
	err = row.Scan(&count)
	if err == sql.ErrNoRows {
		err = nil
	} else if err != nil {
		return
	}
	if count > 0 {
		err = util.ErrDuplicated
		return
	}

	queryStr = `UPDATE robot_profile_answer SET content = ?
		WHERE appid = ? AND qid = ? AND id = ?`
	_, err = t.Exec(queryStr, answer, appid, qid, aid)
	if err != nil {
		return
	}

	err = t.Commit()
	return
}

func deleteRobotQAAnswerV3(appid string, qid, aid int) (err error) {
	defer func() {
		util.ShowError(err)
	}()
	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = util.ErrDBNotInit
		return
	}

	queryStr := `DELETE FROM robot_profile_answer
		WHERE appid = ? AND qid = ? AND id = ?`
	_, err = mySQL.Exec(queryStr, appid, qid, aid)
	return err
}

func addRobotQARQuestionV3(appid string, qid int, question string) (id int, err error) {
	defer func() {
		util.ShowError(err)
	}()
	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = util.ErrDBNotInit
		return
	}

	t, err := mySQL.Begin()
	if err != nil {
		return
	}
	defer util.ClearTransition(t)

	queryStr := `SELECT count(*) FROM robot_profile_question WHERE id = ?`
	qRow := t.QueryRow(queryStr, qid)
	count := 0
	err = qRow.Scan(&count)
	if err != nil {
		return
	}
	if count == 0 {
		err = sql.ErrNoRows
		return
	}

	queryStr = `SELECT count(*) FROM robot_profile_extend
		WHERE appid = ? AND qid = ? AND content = ?`
	row := t.QueryRow(queryStr, appid, qid, question)
	count = 0
	err = row.Scan(&count)
	if err == sql.ErrNoRows {
		err = nil
	} else if err != nil {
		return
	}

	if count > 0 {
		err = util.ErrDuplicated
		return
	}

	queryStr = `
		INSERT INTO robot_profile_extend
		(appid, qid, content) VALUES (?, ?, ?)`
	result, err := t.Exec(queryStr, appid, qid, question)
	if err != nil {
		return
	}

	id64, err := result.LastInsertId()
	if err != nil {
		return
	}

	id = int(id64)
	err = t.Commit()
	return
}

func getRobotQARQuestionV3(appid string, qid, rQid int) (ret *InfoV3, err error) {
	defer func() {
		util.ShowError(err)
	}()
	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = util.ErrDBNotInit
		return
	}

	queryStr := `SELECT id, content FROM robot_profile_extend
		WHERE appid = ? AND qid = ? AND id = ?`
	row := mySQL.QueryRow(queryStr, appid, qid, rQid)
	ret, err = scanInfo(row)
	return
}

func updateRobotQARQuestionV3(appid string, qid, rQid int, relateQuestion string) (err error) {
	defer func() {
		util.ShowError(err)
	}()
	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = util.ErrDBNotInit
		return
	}
	t, err := mySQL.Begin()
	if err != nil {
		return
	}
	defer util.ClearTransition(t)

	queryStr := `SELECT count(*) FROM robot_profile_extend
		WHERE appid = ? AND qid = ? AND content = ? AND id != ?`
	row := t.QueryRow(queryStr, appid, qid, relateQuestion, rQid)
	count := 0
	err = row.Scan(&count)
	if err == sql.ErrNoRows {
		err = nil
	} else if err != nil {
		return
	}
	if count > 0 {
		err = util.ErrDuplicated
		return
	}

	queryStr = `UPDATE robot_profile_extend SET content = ?
		WHERE appid = ? AND qid = ? AND id = ?`
	_, err = t.Exec(queryStr, relateQuestion, appid, qid, rQid)
	if err != nil {
		return
	}

	err = t.Commit()
	return
}

func deleteRobotQARQuestionV3(appid string, qid, rQid int) (err error) {
	defer func() {
		util.ShowError(err)
	}()
	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = util.ErrDBNotInit
		return
	}

	queryStr := `DELETE FROM robot_profile_extend
		WHERE appid = ? AND qid = ? AND id = ?`
	_, err = mySQL.Exec(queryStr, appid, qid, rQid)
	return err
}
