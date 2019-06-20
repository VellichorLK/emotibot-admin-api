package dictionary

import (
	"database/sql"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"time"

	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/pkg/logger"
)

var errDuplicated = errors.New("duplicate item")

func getWordbank(appid string, id int) (*WordBank, error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return nil, errors.New("DB not init")
	}

	ret := &WordBank{}
	queryStr := fmt.Sprintf("SELECT entity_name, similar_words, answer from %s_entity where id = ?", appid)
	row := mySQL.QueryRow(queryStr, id)
	err := row.Scan(&ret.Name, &ret.SimilarWords, &ret.Answer)
	ret.ID = &id
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, nil
		}
		return nil, err
	}

	return ret, nil
}

func addWordbankDir(appid string, paths []string) error {
	wb := &WordBank{}
	addWordbank(appid, paths, wb)
	return nil
}

func updateWordbank(appid string, wordbank *WordBank) error {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return errors.New("DB not init")
	}

	sqlStr := fmt.Sprintf(`UPDATE %s_entity SET similar_words = ?, answer = ? WHERE id = ?`, appid)

	_, err := mySQL.Exec(sqlStr, wordbank.SimilarWords, wordbank.Answer, wordbank.ID)
	if err != nil {
		return err
	}
	return nil
}

func addWordbank(appid string, paths []string, wordbank *WordBank) error {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return errors.New("DB not init")
	}
	vals := make([]interface{}, len(paths))
	for idx := range paths {
		vals[idx] = paths[idx]
	}

	vals = append(vals, wordbank.Name, wordbank.SimilarWords, wordbank.Answer)

	sqlStr := fmt.Sprintf(`INSERT INTO %s_entity
		(level1, level2, level3, level4, status_flag, entity_name, similar_words, answer)
		VALUES (?, ?, ?, ?, 1, ?, ?, ?)`, appid)

	res, err := mySQL.Exec(sqlStr, vals...)
	if err != nil {
		return err
	}
	if wordbank.Name != "" {
		id, _ := res.LastInsertId()
		intID := int(id)
		wordbank.ID = &intID
	} else {
		wordbank.ID = nil
	}

	return nil
}

func checkDirExist(appid string, paths []string) (bool, error) {
	return checkWordbankExist(appid, paths, "")
}

func checkWordbankExist(appid string, paths []string, name string) (bool, error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return false, errors.New("DB not init")
	}

	conditions := []string{}
	vals := []interface{}{}
	for idx := range paths {
		if paths[idx] == "" {
			break
		}
		conditions = append(conditions, fmt.Sprintf("level%d = ?", idx+1))
		vals = append(vals, paths[idx])
	}

	if name != "" {
		// conditionStr += "entity_name = ?"
		conditions = append(conditions, "entity_name = ?")
		vals = append(vals, name)
	}

	sqlStr := fmt.Sprintf(`SELECT COUNT(*) from %s_entity WHERE %s`, appid, strings.Join(conditions, " and "))
	rows := mySQL.QueryRow(sqlStr, vals...)
	count := 0
	err := rows.Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// getProcessStatus will get status of latest wordbank process
func getProcessStatus(appid string) (string, error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return "", errors.New("DB not init")
	}

	rows, err := mySQL.Query("SELECT status from process_status where app_id = ? and module = 'wordbank' order by id desc limit 1", appid)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var status string
	ret := rows.Next()
	if !ret {
		return "", nil
	}
	if err := rows.Scan(&status); err != nil {
		return "", err
	}

	return status, nil
}

// getFullProcessStatus will get more status info from latest wordbank process
func getFullProcessStatus(appid string) (*StatusInfo, error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return nil, errors.New("DB not init")
	}

	rows, err := mySQL.Query("SELECT status, UNIX_TIMESTAMP(start_at), message from process_status where app_id = ? and module = 'wordbank' order by id desc limit 1", appid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	status := StatusInfo{}
	ret := rows.Next()
	if !ret {
		return nil, nil
	}

	var timestamp int64
	if err := rows.Scan(&status.Status, &timestamp, &status.Message); err != nil {
		return nil, err
	}
	status.StartTime = time.Unix(timestamp, 0)

	emptyMsg := ""
	if status.Message == nil {
		status.Message = &emptyMsg
	}
	return &status, nil
}

// getLastTwoSuccess will return last two record which status is success, order by time
func getLastTwoSuccess(appid string) ([]*DownloadMeta, error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return nil, errors.New("DB not init")
	}

	rows, err := mySQL.Query("SELECT UNIX_TIMESTAMP(start_at),entity_file_name from process_status where app_id = ? and module = 'wordbank' and status = 'success' order by start_at desc limit 2", appid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ret := []*DownloadMeta{}
	for rows.Next() {
		var meta DownloadMeta
		var startTime int64
		if err := rows.Scan(&startTime, &meta.UploadFile); err != nil {
			return nil, err
		}

		meta.UploadTime = time.Unix(startTime, 0)

		ret = append(ret, &meta)
	}

	return ret, nil
}

// insertProcess will create a file record into process_status, which status is running
func insertProcess(appid string, status Status, filename string, message string) error {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return errors.New("DB not init")
	}

	_, err := mySQL.Exec("insert process_status(app_id, module, status, entity_file_name, message) values (?, 'wordbank', ?, ?, ?)", appid, status, filename, message)
	if err != nil {
		return err
	}

	return nil
}

func getEntities(appid string) ([]*WordBank, error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return nil, errors.New("DB not init")
	}

	queryStr := fmt.Sprintf("SELECT level1, level2, level3, level4, entity_name, similar_words, answer, id from %s_entity where status_flag = 1", appid)
	rows, err := mySQL.Query(queryStr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cache := make(map[int](map[string]*WordBank))
	for idx := 0; idx < 4; idx++ {
		cache[idx] = make(map[string]*WordBank)
	}
	for rows.Next() {
		categories := make([]sql.NullString, 4)
		var entityName sql.NullString
		var similarWord sql.NullString
		var answer sql.NullString
		var id int

		if err := rows.Scan(&categories[0], &categories[1], &categories[2], &categories[3], &entityName, &similarWord, &answer, &id); err != nil {
			return nil, err
		}

		var lastCategory *WordBank
		for idx, category := range categories {
			if !category.Valid || category.String == "" {
				break
			}

			if _, ok := cache[idx][category.String]; !ok {
				newWordBank := &WordBank{nil, category.String, 0, make([]*WordBank, 0), "", ""}
				cache[idx][category.String] = newWordBank
				if lastCategory != nil {
					lastCategory.Children = append(lastCategory.Children, newWordBank)
				}
			}
			lastCategory = cache[idx][category.String]
		}
		if lastCategory == nil {
			logger.Error.Printf("Level 1 should not be empty in wordbank, skip it")
			continue
		}
		if entityName.Valid && entityName.String != "" {
			newWordBank := &WordBank{&id, entityName.String, 1, make([]*WordBank, 0), "", ""}
			if similarWord.Valid && similarWord.String != "" {
				newWordBank.SimilarWords = similarWord.String
			}
			if answer.Valid && answer.String != "" {
				newWordBank.Answer = answer.String
			}
			lastCategory.Children = append(lastCategory.Children, newWordBank)
		}
	}

	ret := []*WordBank{}
	for _, wordbank := range cache[0] {
		ret = append(ret, wordbank)
	}

	return ret, nil
}

func getWordbankRows(appid string) (ret []*WordBankRow, err error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = errors.New("DB not init")
		return
	}
	queryStr := fmt.Sprintf(`
		SELECT level1, level2, level3, level4, entity_name, similar_words, answer
		FROM %s_entity`, appid)

	rows, err := mySQL.Query(queryStr)
	if err != nil {
		return
	}

	ret = []*WordBankRow{}
	for rows.Next() {
		temp := WordBankRow{}
		err = rows.Scan(
			&temp.Level1, &temp.Level2, &temp.Level3, &temp.Level4,
			&temp.Name, &temp.SimilarWords, &temp.Answer)
		if err != nil {
			return
		}
		ret = append(ret, &temp)
	}

	return
}

func deleteWordbankDir(appid string, paths []string) (int, error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return 0, errors.New("DB not init")
	}

	if len(paths) <= 1 || len(paths) > 4 {
		return 0, errors.New("Error path")
	}

	queryParam := []interface{}{}
	queryCondition := []string{}

	for idx, path := range paths {
		queryParam = append(queryParam, path)
		queryCondition = append(queryCondition, fmt.Sprintf("level%d = ?", idx+1))
	}

	queryStr := fmt.Sprintf("DELETE FROM %s_entity WHERE %s", appid, strings.Join(queryCondition, " and "))
	result, err := mySQL.Exec(queryStr, queryParam...)
	if err != nil {
		return 0, err
	}

	count, err := result.RowsAffected()

	return int(count), err
}

func deleteWordbank(appid string, id int) (err error) {
	defer showError(err)
	mySQL := util.GetMainDB()
	if mySQL == nil {
		return errors.New("DB not init")
	}

	t, err := mySQL.Begin()
	if err != nil {
		return err
	}
	defer util.ClearTransition(t)

	// TODO: check if wordbank is last item in directory, if yes, add new row for the directory only
	queryStr := fmt.Sprintf("SELECT level1, level2, level3, level4 FROM %s_entity WHERE id = ?", appid)
	row := t.QueryRow(queryStr, id)
	paths := make([]string, 4)
	err = row.Scan(&paths[0], &paths[1], &paths[2], &paths[3])
	if err != nil {
		return err
	}

	queryParam := []interface{}{}
	queryCondition := []string{}

	for idx, path := range paths {
		if path == "" {
			break
		}
		queryParam = append(queryParam, path)
		queryCondition = append(queryCondition, fmt.Sprintf("level%d = ?", idx+1))
	}
	if len(queryParam) == 0 {
		return errors.New("Path empty error")
	}

	count := 0
	queryStr = fmt.Sprintf("SELECT count(*) from %s_entity WHERE %s", appid, strings.Join(queryCondition, " and "))
	row = t.QueryRow(queryStr, queryParam...)
	err = row.Scan(&count)
	if err != nil {
		return err
	}

	if count <= 1 {
		queryStr = fmt.Sprintf(`
			UPDATE %s_entity
			SET entity_name = '', similar_words = '', answer = ''
			WHERE id = ?`, appid)
	} else {
		queryStr = fmt.Sprintf("DELETE FROM %s_entity WHERE id = ?", appid)
	}
	_, err = t.Exec(queryStr, id)
	if err != nil {
		return err
	}

	err = t.Commit()
	return err
}

func getWordbankRow(appid string, id int) (ret *WordBankRow, err error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = errors.New("DB not init")
		return
	}
	queryStr := fmt.Sprintf(`
		SELECT level1, level2, level3, level4, entity_name, similar_words, answer
		FROM %s_entity
		WHERE id = ?`, appid)

	row := mySQL.QueryRow(queryStr, id)
	if err != nil {
		return
	}

	temp := WordBankRow{}
	err = row.Scan(
		&temp.Level1, &temp.Level2, &temp.Level3, &temp.Level4,
		&temp.Name, &temp.SimilarWords, &temp.Answer)
	if err != nil {
		return
	}
	ret = &temp

	return
}

func getWordbanksV3(appid string) (ret *WordBankClassV3, err error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = errors.New("DB not init")
		return
	}

	queryStr := `
		SELECT id, appid, name, pid, editable, intent_engine, rule_engine
		FROM entity_class
		WHERE appid = ? ORDER BY id DESC`
	rows, err := mySQL.Query(queryStr, appid)
	if err != nil {
		return
	}
	defer rows.Close()

	rootMap, classMap, err := parseWordbankClassRowsV3(rows)
	if err != nil {
		return
	}

	// queryParam is used in next mysql params
	queryParam := []interface{}{}
	// quereyQuestion will used to form the question mark in sql query
	queryQuestion := []string{}
	for id := range classMap {
		queryParam = append(queryParam, id)
		queryQuestion = append(queryQuestion, "?")
	}
	if len(queryParam) <= 0 {
		// no WordBankClass found for this appid, return a empty root
		ret = &WordBankClassV3{-1, "", []*WordBankV3{}, []*WordBankClassV3{}, false, true, true}
		return
	}
	queryParam = append(queryParam, appid)

	if len(queryParam) == 1 {
		// only has appid condition
		queryStr = `
			SELECT id, appid, name, editable, cid, similar_words, answer
			FROM entities
			WHERE appid = ? ORDER BY id DESC`
	} else {
		queryStr = fmt.Sprintf(`
			SELECT id, appid, name, editable, cid, similar_words, answer
			FROM entities
			WHERE cid in (%s) OR appid = ? ORDER BY id DESC`, strings.Join(queryQuestion, ","))
	}
	entityRows, err := mySQL.Query(queryStr, queryParam...)
	if err != nil {
		return
	}
	defer entityRows.Close()

	rootMap, err = parseWordbankRowsV3(rootMap, classMap, entityRows)
	if err != nil {
		return
	}
	ret = rootMap[appid]
	return
}

// getWordbanksAllV3 get wordbanks for all appid
func getWordbanksAllV3() (retRootMap map[string]*WordBankClassV3, err error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = errors.New("DB not init")
		return
	}

	queryStr := `
		SELECT id, appid, name, pid, editable, intent_engine, rule_engine
		FROM entity_class
		ORDER BY id`
	rows, err := mySQL.Query(queryStr)
	if err != nil {
		return
	}
	defer rows.Close()

	rootMap, classMap, err := parseWordbankClassRowsV3(rows)
	if err != nil {
		return
	}

	queryStr = `
		SELECT id, appid, name, editable, cid, similar_words, answer
		FROM entities ORDER BY id DESC`
	entityRows, err := mySQL.Query(queryStr)
	if err != nil {
		return
	}
	defer entityRows.Close()

	rootMap, err = parseWordbankRowsV3(rootMap, classMap, entityRows)
	if err != nil {
		return
	}
	return rootMap, nil
}

// parseWordbankClassRowsV3 parse WordbankClass rows and return rootMap, classMap
func parseWordbankClassRowsV3(rows *sql.Rows) (retRootMap map[string]*WordBankClassV3, retClassMap map[int]*WordBankClassV3, err error) {
	// rootMap maps appid to its root wordbank class
	rootMap := map[string]*WordBankClassV3{}
	// classMap map classID to wordbank class
	classMap := map[int]*WordBankClassV3{}
	// childrenMap map classID to it's children wordbank classes
	childrenMap := map[int][]*WordBankClassV3{}

	for rows.Next() {
		temp := &WordBankClassV3{}
		temp.Children = []*WordBankClassV3{}
		temp.Wordbank = []*WordBankV3{}
		var pidPtr *int
		var appid string
		err = rows.Scan(&temp.ID, &appid, &temp.Name, &pidPtr, &temp.Editable, &temp.IntentEngine, &temp.RuleEngine)
		if err != nil {
			return
		}
		if _, ok := rootMap[appid]; !ok {
			// initialize a root WordbankClass
			rootMap[appid] = &WordBankClassV3{-1, "", []*WordBankV3{}, []*WordBankClassV3{}, false, true, true}
		}
		classMap[temp.ID] = temp

		if pidPtr != nil {
			pid := *pidPtr
			if _, ok := childrenMap[pid]; !ok {
				childrenMap[pid] = []*WordBankClassV3{}
			}
			childrenMap[pid] = append(childrenMap[pid], temp)
		} else {
			rootMap[appid].Children = append(rootMap[appid].Children, temp)
		}
	}

	for id, class := range classMap {
		class.Children = childrenMap[id]
	}
	return rootMap, classMap, nil
}

// parseWordbankRowsV3 parse Wordbank rows and return rootMap
func parseWordbankRowsV3(rootMap map[string]*WordBankClassV3, classMap map[int]*WordBankClassV3, entityRows *sql.Rows) (retRootMap map[string]*WordBankClassV3, err error) {
	for entityRows.Next() {
		temp := &WordBankV3{}
		var appid string
		var editable *int
		var cid *int
		similarWordStr := ""
		err = entityRows.Scan(&temp.ID, &appid, &temp.Name, &editable, &cid, &similarWordStr, &temp.Answer)
		if err != nil {
			return
		}
		if similarWordStr == "" {
			temp.SimilarWords = []string{}
		} else {
			temp.SimilarWords = strings.Split(similarWordStr, ",")
		}
		if editable != nil {
			temp.Editable = *editable != 0
		} else if cid != nil && *cid != -1 && classMap[*cid] != nil {
			temp.Editable = classMap[*cid].Editable
		} else {
			temp.Editable = true
		}

		if cid == nil || *cid == -1 {
			//rootMap[appid].Wordbank = append(rootMap[appid].Wordbank, temp)
			if _, ok := rootMap[appid]; ok {
				rootMap[appid].Wordbank = append(rootMap[appid].Wordbank, temp)
			}
		} else {
			var classObj *WordBankClassV3
			classObj = classMap[*cid]
			if classObj == nil {
				classObj = rootMap[appid]
			}
			classObj.Wordbank = append(classObj.Wordbank, temp)
		}
	}
	return rootMap, nil
}

func getWordbankV3(appid string, id int) (ret *WordBankV3, cid int, err error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = errors.New("DB not init")
		return
	}

	queryStr := `
		SELECT e.name, e.editable, e.similar_words, e.answer, e.cid, c.editable FROM (
			SELECT name, editable, similar_words, answer, cid
			FROM entities 
			WHERE id = ? AND appid = ?
		) as e
		LEFT JOIN entity_class as c
		ON e.cid = c.id`
	row := mySQL.QueryRow(queryStr, id, appid)
	logger.Trace.Printf("Get wordbank of %d@%s", id, appid)

	ret = &WordBankV3{}
	similarWordsStr := ""
	var classID *int
	var editable *int
	var parentEditable *int
	err = row.Scan(&ret.Name, &editable, &similarWordsStr, &ret.Answer, &classID, &parentEditable)
	if err != nil {
		if err == sql.ErrNoRows {
			ret = nil
			err = nil
		}
		return
	}

	if editable != nil {
		ret.Editable = *editable != 0
	} else if parentEditable != nil {
		ret.Editable = *parentEditable != 0
	} else {
		ret.Editable = true
	}
	similarWordsStr = strings.TrimSpace(similarWordsStr)
	if similarWordsStr == "" {
		ret.SimilarWords = []string{}
	} else {
		ret.SimilarWords = strings.Split(similarWordsStr, ",")
	}
	if classID != nil {
		cid = *classID
	} else {
		cid = -1
	}
	ret.ID = id
	return
}

func getWordbankClassV3(appid string, id int) (ret *WordBankClassV3, pid int, err error) {
	defer func() {
		util.ShowError(err)
	}()
	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = errors.New("DB not init")
		return
	}

	queryStr := `
		SELECT name, editable, intent_engine, rule_engine, pid
		FROM entity_class
		WHERE id = ? AND appid = ?`
	row := mySQL.QueryRow(queryStr, id, appid)

	ret = &WordBankClassV3{}
	var pidScan *int
	var editable *int
	ie := 0
	re := 0
	err = row.Scan(&ret.Name, &editable, &ie, &re, &pidScan)
	if err != nil {
		if err == sql.ErrNoRows {
			ret = nil
			err = nil
		}
		return
	}

	if pidScan == nil {
		pid = -1
	} else {
		pid = *pidScan
	}

	ret.Editable = false
	if editable != nil && *editable != 0 {
		ret.Editable = true
	}
	ret.IntentEngine = ie != 0
	ret.RuleEngine = re != 0
	ret.ID = id
	return
}

func getWordbankClassParents(appid string, id int) (ret []*WordBankClassV3, err error) {
	defer showError(err)
	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = errors.New("DB not init")
		return
	}

	queryStr := `
		SELECT id, name, editable, intent_engine, rule_engine
		FROM (
			SELECT * FROM entity_class ORDER BY id DESC) AS sorted,
			(SELECT @pv := ?) AS tmp
		WHERE
			find_in_set(id, @pv)
			AND (pid IS NULL OR length(@pv := concat(@pv, ',', pid)))
			AND appid = ?
		ORDER BY id
	`
	rows, err := mySQL.Query(queryStr, id, appid)
	if err != nil {
		return
	}

	classes := []*WordBankClassV3{}
	for rows.Next() {
		temp := &WordBankClassV3{}
		editable := 0
		cid := 0
		ie := 0
		re := 0
		err = rows.Scan(&cid, &temp.Name, &editable, &ie, &re)
		if err != nil {
			return
		}

		temp.Editable = editable != 0
		temp.IntentEngine = ie != 0
		temp.RuleEngine = re != 0
		temp.ID = cid
		classes = append(classes, temp)
	}

	ret = classes
	return
}

func getWordbanksWithChildren(appid string, id int) (root *WordBankClassV3, err error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = errors.New("DB not init")
		return
	}

	queryStr := `
		SELECT name, editable, intent_engine, rule_engine
		FROM entity_class
		WHERE id = ? AND appid = ?`
	row := mySQL.QueryRow(queryStr, id, appid)

	root = &WordBankClassV3{}
	var editable *int
	ie := 0
	re := 0
	err = row.Scan(&root.Name, &editable, &ie, &re)
	if err != nil {
		if err == sql.ErrNoRows {
			root = nil
			err = nil
		}
		return root, err
	}

	root.Editable = false
	if editable != nil && *editable != 0 {
		root.Editable = true
	}
	root.IntentEngine = ie != 0
	root.RuleEngine = re != 0
	root.ID = id

	queryStr = `
		SELECT id, name, editable, intent_engine, rule_engine
		FROM entity_class
		WHERE appid = ? AND pid = ?`
	rows, err := mySQL.Query(queryStr, appid, id)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		temp := &WordBankClassV3{}
		temp.Children = []*WordBankClassV3{}
		temp.Wordbank = []*WordBankV3{}
		err = rows.Scan(&temp.ID, &temp.Name, &temp.Editable, &temp.IntentEngine, &temp.RuleEngine)
		if err != nil {
			return
		}
		root.Children = append(root.Children, temp)
	}

	for i, child := range root.Children {
		root.Children[i], err = getWordbanksWithChildren(appid, child.ID)
		if err != nil {
			return
		}
	}

	queryStr = `SELECT id, editable, name, cid, similar_words, answer
		FROM entities
		WHERE cid = ?`
	entityRows, err := mySQL.Query(queryStr, id)
	if err != nil {
		return
	}
	defer entityRows.Close()

	for entityRows.Next() {
		temp := &WordBankV3{}
		cid := 0
		var editable *int
		similarWordStr := ""
		err = entityRows.Scan(&temp.ID, &editable, &temp.Name, &cid, &similarWordStr, &temp.Answer)
		if err != nil {
			return
		}
		if similarWordStr == "" {
			temp.SimilarWords = []string{}
		} else {
			temp.SimilarWords = strings.Split(similarWordStr, ",")
		}
		if editable != nil {
			temp.Editable = *editable != 0
		} else {
			temp.Editable = true
		}

		root.Wordbank = append(root.Wordbank, temp)
	}

	return
}

func deleteWordbankV3(appid string, id int) (err error) {
	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = errors.New("DB not init")
		return
	}
	queryStr := `
	DELETE e.* FROM 
	entities as e,
		(SELECT * FROM (
			SELECT e.id as id, c.id as cid, e.editable as editable  FROM
				(SELECT * FROM entities WHERE id = ? AND appid = ?) AS e
			LEFT JOIN entity_class as c
			ON
				e.cid = c.id
			) as temp
			WHERE (temp.cid IS NULL) OR (temp.cid IS NOT NULL AND editable > 0)
		) AS temp2
	WHERE e.id = temp2.id
	`
	_, err = mySQL.Exec(queryStr, id, appid)
	fmt.Print(id)
	fmt.Print(appid)
	return err
}
func deleteWordbankClassV3(appid string, id int) (err error) {
	defer showError(err)

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

	queryStr := `
		SELECT id
		FROM (
			SELECT * FROM entity_class order by id) AS sorted,
			(select @pv := ?) AS tmp
		WHERE 
			find_in_set(pid, @pv)
			AND appid = ?
			AND length(@pv := concat(@pv, ',', id))
	`
	rows, err := t.Query(queryStr, id, appid)
	if err != nil {
		return err
	}
	defer rows.Close()
	ids := []interface{}{}

	for rows.Next() {
		tmp := 0
		err = rows.Scan(&tmp)
		if err != nil {
			return err
		}
		ids = append(ids, tmp)
	}
	ids = append(ids, id)

	qmark := make([]string, len(ids))
	for i := range qmark {
		qmark[i] = "?"
	}

	if len(ids) == 0 {
		return nil
	}

	logger.Trace.Printf("Delete id: %v\n", ids)

	queryStr = fmt.Sprintf("DELETE FROM entity_class WHERE id in (%s)", strings.Join(qmark, ","))
	_, err = t.Exec(queryStr, ids...)
	if err != nil {
		return
	}

	queryStr = fmt.Sprintf("UPDATE entities SET cid = NULL WHERE cid in (%s)", strings.Join(qmark, ","))
	_, err = t.Exec(queryStr, ids...)
	if err != nil {
		return
	}

	err = t.Commit()
	return err
}

func showError(err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		logger.Error.Printf("DB error [%s:%d]: %s\n", file, line, err.Error())
	}
}

func addWordbankClassV3(appid string, className string, pid int) (id int, err error) {
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

	pidPtr := &pid
	if pid == -1 {
		pidPtr = nil
	}

	queryStr := "SELECT count(*) FROM entity_class WHERE appid = ? AND BINARY name = ? AND pid = ?"
	row := t.QueryRow(queryStr, appid, className, pidPtr)
	count := 0
	err = row.Scan(&count)
	if err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			return
		}
	}
	if count > 0 {
		err = errDuplicated
		return
	}

	columns := []string{"appid", "name", "editable"}
	params := []interface{}{appid, className, 1}
	q := []string{"?", "?", "?"}
	if pid != -1 {
		columns = append(columns, "pid")
		params = append(params, pid)
		q = append(q, "?")
	}

	queryStr = fmt.Sprintf(
		"INSERT INTO entity_class (%s) VALUES (%s)",
		strings.Join(columns, ","),
		strings.Join(q, ","))
	result, err := t.Exec(queryStr, params...)
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

func updateWordbankClassV3(appid string, id int, className string) (err error) {
	defer func() {
		showError(err)
	}()

	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = errors.New("DB not init")
		return
	}

	queryStr := `
		UPDATE entity_class
		SET appid = ?, name = ?
		WHERE id = ?`
	_, err = mySQL.Exec(queryStr, appid, className, id)
	if err != nil {
		return
	}
	return
}

func addWordbankV3(appid string, cid int, wb *WordBankV3) (id int, err error) {
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

	queryStr := "SELECT count(*) FROM entities WHERE cid = ? AND BINARY name = ? AND appid = ?"
	row := t.QueryRow(queryStr, cid, wb.Name, appid)
	count := 0
	err = row.Scan(&count)
	if err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			return
		}
	}
	if count > 0 {
		err = errDuplicated
		return
	}

	classID := &cid
	if cid == -1 {
		classID = nil
	}
	queryStr = `INSERT INTO entities
		(name, appid, cid, similar_words, answer)
		VALUES (?, ?, ?, ?, ?)`
	result, err := t.Exec(queryStr,
		wb.Name, appid, classID, strings.Join(wb.SimilarWords, ","), wb.Answer)
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

func updateWordbankV3(appid string, id int, wb *WordBankV3) (err error) {
	defer func() {
		util.ShowError(err)
	}()

	mySQL := util.GetMainDB()
	if mySQL == nil {
		err = errors.New("DB not init")
		return
	}

	queryStr := `UPDATE entities SET
		name = ?, similar_words = ?, answer = ?
		WHERE id = ?`
	_, err = mySQL.Exec(queryStr,
		wb.Name, strings.Join(wb.SimilarWords, ","), wb.Answer, id)
	return
}

func moveWordbankV3(appid string, id, cid int) (err error) {
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

	classID := &cid
	if cid == -1 {
		classID = nil
	}
	// check if target class has a same name of entity or not
	queryStr := `
		SELECT count(*)
		FROM entities AS e1, entities AS e2
		WHERE e1.id = ? AND e2.cid = ? AND BINARY e2.name = e1.name AND e2.id != e1.id`
	row := t.QueryRow(queryStr, id, classID)
	count := 0
	err = row.Scan(&count)
	if err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			return
		}
	}
	if count > 0 {
		err = errDuplicated
		return
	}

	queryStr = `UPDATE entities SET cid = ? WHERE id = ?`
	_, err = t.Exec(queryStr, classID, id)
	if err != nil {
		return
	}

	err = t.Commit()
	return
}
