package Robot

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/pkg/logger"
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
		WHERE appid = ? AND status >= 0`
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
		WHERE appid = ? AND status >= 0`
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
		WHERE appid = ? AND qid = ? AND status >= 0`
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
		WHERE appid = ? AND qid = ? AND status >= 0`
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
		WHERE appid = ? AND qid = ? AND content = ? AND status >= 0`
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
		(appid, qid, content, status) VALUES (?, ?, ?, 1)`
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
		WHERE appid = ? AND qid = ? AND id = ? AND status >= 0`
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
		WHERE appid = ? AND qid = ? AND content = ? AND id != ? AND status >= 0`
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

	queryStr = `UPDATE robot_profile_answer SET content = ?, status = 1
		WHERE appid = ? AND qid = ? AND id = ? AND status >= 0`
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

	queryStr := `UPDATE robot_profile_answer SET status = -1
		WHERE appid = ? AND qid = ? AND id = ? AND status >= 0`
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
		WHERE appid = ? AND qid = ? AND content = ? AND status >= 0`
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
		(appid, qid, content, status) VALUES (?, ?, ?, 1)`
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
		WHERE appid = ? AND qid = ? AND id = ? AND status >= 0`
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
		WHERE appid = ? AND qid = ? AND content = ? AND id != ? AND status >= 0`
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
		WHERE appid = ? AND qid = ? AND id = ? AND status >= 0`
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

	queryStr := `UPDATE robot_profile_extend SET status = -1
		WHERE appid = ? AND qid = ? AND id = ? AND status >= 0`
	_, err = mySQL.Exec(queryStr, appid, qid, rQid)
	return err
}

func tryStartSyncProcess(syncSolrTimeout int) (ret bool, processID int, err error) {
	// this sync status no need to check appid
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

	var start int
	var running bool
	var status = ""

	queryStr := `
		SELECT UNIX_TIMESTAMP(start_at), status FROM process_status
		WHERE module = 'robot-profile'
		ORDER BY id desc limit 1
	`
	row := t.QueryRow(queryStr)
	err = row.Scan(&start, &status)
	if err == sql.ErrNoRows {
		running, err = false, nil
	} else if err != nil {
		return
	} else {
		running, err = status == "running", nil
	}

	now := time.Now().Unix()
	if running {
		logger.Trace.Printf("Previous still running from %d", start)
		if int(now)-start <= syncSolrTimeout {
			return
		}
	}

	queryStr = `
		INSERT INTO process_status
		(app_id, module, status) VALUES ('', 'robot-profile', 'running')
	`
	result, err := t.Exec(queryStr)
	if err != nil {
		return
	}

	id64, err := result.LastInsertId()
	if err != nil {
		return
	}

	processID = int(id64)
	err = t.Commit()
	if err != nil {
		return
	}

	ret = true
	return
}

func finishSyncProcess(pid int, result bool, msg string) (err error) {
	// this sync status no need to check appid
	defer func() {
		util.ShowError(err)
	}()
	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = util.ErrDBNotInit
		return
	}

	status := "success"
	if !result {
		status = "fail"
	}

	queryStr := `
		UPDATE process_status
		SET status = ?, message = ?
		WHERE id = ?`
	_, err = mySQL.Exec(queryStr, status, msg, pid)
	return
}

func getProcessModifyRobotQA() (rqIDs []interface{}, ansIDs []interface{}, deleteAnsIDs []interface{}, ret []*ManualTagging, appids []string, err error) {
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

	appidMap := map[string]bool{}
	queryStr := `
	SELECT id, qid, content, appid
	FROM
	(
			SELECT id, r.qid as qid, content, appid, status, aid
			FROM robot_profile_extend AS r
			-- select answer changed from answer
			LEFT JOIN
				(
					SELECT qid, GROUP_CONCAT(id) as aid
					FROM robot_profile_answer
					WHERE status = 1 OR status = -1
					GROUP BY qid
				) as changeQ
			ON r.qid = changeQ.qid
	) AS final
	-- only get changed extend q, and extend q whose answer changed
	WHERE final.status = 1 OR final.aid IS NOT NULL`
	rqRows, err := t.Query(queryStr)
	if err != nil {
		return
	}
	defer rqRows.Close()

	rqInfos := []*ManualTagging{}
	// key is standard q id, values are rq of standard q
	rqMap := map[int][]*ManualTagging{}
	qids := []interface{}{}

	for rqRows.Next() {
		temp := ManualTagging{}
		id, qid := 0, 0
		err = rqRows.Scan(&id, &qid, &temp.Question, &temp.AppID)
		temp.SolrID = fmt.Sprintf("%d_%d", qid, id)
		if _, ok := rqMap[qid]; !ok {
			rqMap[qid] = []*ManualTagging{}
			qids = append(qids, qid)
		}
		rqMap[qid] = append(rqMap[qid], &temp)
		rqInfos = append(rqInfos, &temp)
		rqIDs = append(rqIDs, id)
		appidMap[temp.AppID] = true
	}

	queryStr = `
	SELECT DISTINCT a.qid, q.content, a.appid
	FROM robot_profile_answer as a, robot_profile_question as q
	WHERE
		(a.status = 1 OR a.status = -1)
		AND a.qid = q.id;`
	qRows, err := t.Query(queryStr)
	if err != nil {
		return
	}
	for qRows.Next() {
		temp := ManualTagging{}
		id := 0
		err = qRows.Scan(&id, &temp.Question, &temp.AppID)
		temp.SolrID = fmt.Sprintf("%d_0", id)
		if _, ok := rqMap[id]; !ok {
			rqMap[id] = []*ManualTagging{}
			qids = append(qids, id)
		}
		rqMap[id] = append(rqMap[id], &temp)
		rqInfos = append(rqInfos, &temp)
		appidMap[temp.AppID] = true
	}

	if len(qids) == 0 {
		return
	}

	queryStr = fmt.Sprintf(`
		SELECT id, qid, content, appid FROM robot_profile_answer
		WHERE qid in (?%s) AND status >= 0`, strings.Repeat(",?", len(qids)-1))
	ansRows, err := t.Query(queryStr, qids...)
	if err != nil {
		return
	}
	defer ansRows.Close()

	ansIDs = []interface{}{}
	for ansRows.Next() {
		id, qid, content, appid := 0, 0, "", ""
		err = ansRows.Scan(&id, &qid, &content, &appid)
		if err != nil {
			return
		}

		if _, ok := rqMap[qid]; ok {
			for idx := range rqMap[qid] {
				info := rqMap[qid][idx]
				if info.AppID == appid {
					info.Answers = append(info.Answers, &ManualAnswerTagging{
						SolrID: fmt.Sprintf("%s_%d", info.SolrID, id),
						Answer: content,
					})
				}
			}
		}
		ansIDs = append(ansIDs, id)
		appidMap[appid] = true
	}

	queryStr = fmt.Sprintf(`
		SELECT id, appid FROM robot_profile_answer
		WHERE qid in (?%s) AND status = -1`, strings.Repeat(",?", len(qids)-1))
	delAnsRows, err := t.Query(queryStr, qids...)
	if err != nil {
		return
	}
	defer delAnsRows.Close()

	for delAnsRows.Next() {
		id := 0
		appid := ""
		err = delAnsRows.Scan(&id, &appid)
		if err != nil {
			return
		}
		deleteAnsIDs = append(deleteAnsIDs, id)
		appidMap[appid] = true
	}

	ret = rqInfos
	err = t.Commit()

	for key := range appidMap {
		appids = append(appids, key)
	}

	return
}

func getDeleteModifyRobotQA() (mapSolrIDs map[string][]string, rqIDs []interface{}, err error) {
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

	queryStr := `
		SELECT id, appid, r.qid
		FROM robot_profile_extend AS r
		WHERE r.status = -1;`
	rqRows, err := t.Query(queryStr)
	if err != nil {
		return
	}
	defer rqRows.Close()

	mapSolrIDs = map[string][]string{}
	for rqRows.Next() {
		id, qid := 0, 0
		appid := ""
		err = rqRows.Scan(&id, &appid, &qid)
		if err != nil {
			return
		}
		if _, ok := mapSolrIDs[appid]; !ok {
			mapSolrIDs[appid] = []string{}
		}
		mapSolrIDs[appid] = append(mapSolrIDs[appid], fmt.Sprintf("%d_%d", qid, id))
		rqIDs = append(rqIDs, id)
	}
	err = t.Commit()
	return
}

func resetRobotQAData(rqIDs []interface{}, deleteRQIDs []interface{}, ansIDs []interface{}, delAnsIDs []interface{}) (err error) {
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
	queryStr := ""

	if len(rqIDs) > 0 {
		queryStr = fmt.Sprintf(`
			UPDATE robot_profile_extend SET status = 0
			WHERE id in (?%s)`, strings.Repeat(",?", len(rqIDs)-1))
		_, err = t.Exec(queryStr, rqIDs...)
		if err != nil {
			return
		}
	}
	if len(deleteRQIDs) > 0 {
		queryStr = fmt.Sprintf(`
			DELETE FROM robot_profile_extend
			WHERE status = -1 AND id in (?%s)`, strings.Repeat(",?", len(deleteRQIDs)-1))
		_, err = t.Exec(queryStr, deleteRQIDs...)
		if err != nil {
			return
		}
	}
	if len(ansIDs) > 0 {
		queryStr = fmt.Sprintf(`
			UPDATE robot_profile_answer SET status = 0
			WHERE status = 1 AND id in (?%s)`, strings.Repeat(",?", len(ansIDs)-1))
		_, err = t.Exec(queryStr, ansIDs...)
		if err != nil {
			return
		}
	}
	if len(delAnsIDs) > 0 {
		queryStr = fmt.Sprintf(`
			DELETE FROM robot_profile_answer
			WHERE status = -1 AND id in (?%s)`, strings.Repeat(",?", len(delAnsIDs)-1))
		_, err = t.Exec(queryStr, delAnsIDs...)
		if err != nil {
			return
		}
	}

	err = t.Commit()
	return
}

func needProcessRobotData() (ret bool, err error) {
	ret = false
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

	queryStr := `
		SELECT count(*) FROM robot_profile_extend
		WHERE status != 0`
	extendRow := t.QueryRow(queryStr)
	count := 0
	err = extendRow.Scan(&count)
	if err != nil {
		return
	}

	if count > 0 {
		logger.Trace.Println("New modify in robot profile extend")
		ret = true
		return
	}

	queryStr = `
		SELECT count(*) FROM robot_profile_answer
		WHERE status != 0`
	answerRow := t.QueryRow(queryStr)
	count = 0
	err = answerRow.Scan(&count)
	if err != nil {
		return
	}

	if count > 0 {
		logger.Trace.Println("New modify in robot profile answer")
		ret = true
		return
	}

	err = t.Commit()
	return
}
