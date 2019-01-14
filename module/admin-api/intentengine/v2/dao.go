package intentenginev2

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/pkg/logger"
)

type intentDaoInterface interface {
	// GetIntents will get intents from db
	// version can be NULL, will return data that version column in NULL
	// each return object only has id, name and count of positive/negative sentences
	GetIntents(appid string, version *int, keyword string) ([]*IntentV2, error)

	// GetIntent will get intent from db with specific appid and id
	// It will return full data of the intent, containing id, name, count and sentences
	GetIntent(appid string, intentID int64, keyword string) (*IntentV2, error)

	// AddIntent will add new intent of appid. The version of new intent will be NULL, which
	// means that intent hasn't been trained
	AddIntent(appid, name string, positive, negative []string) (*IntentV2, error)

	// ModifyIntent will update intent with diff information
	// In updateSentence, if id is 0, means new sentence.
	// If intent version is not NULL, return error because it is readonly
	ModifyIntent(appid string, intentID int64, name string, updateSentence []*SentenceV2WithType, deleteSentences []int64) error

	// DeleteIntent will delete intent with provided intentID only when intent is not read only
	DeleteIntent(appid string, intentID int64) error
	// DeleteIntents will delete intents with provided intentIDs only when intent is not read only
	DeleteIntents(appid string, intentIDs []int64) error

	// GetIntentsDetail will return all intents with full information of specific version
	GetIntentsDetail(appid string, version *int) ([]*IntentV2, error)

	// CommitIntent will save current intent list into a snapshot for training
	// After commit intent, it will get a version and intent engine can get train data with version number
	CommitIntent(appid string) (version int, ret []*IntentV2, err error)
	NeedCommit(appid string) (ret bool, err error)
	GetVersionInfo(appid string, version int) (ret *VersionInfoV2, err error)
	GetLatestVersion(appid string) (version int, err error)
	UpdateLatestIntents(appid string, intents []*IntentV2) (err error)
	UpdateVersionStart(version int, start int64, modelID string) (err error)
	UpdateVersionStatus(version int, end int64, status int) (err error)

	// SearchIntentOfSentence will check if sentence existed in specific version or not.
	// If not existed, return err will be sql.ErrNoRows
	// The extended params will at most only one, sentenceType(int)
	SearchIntentOfSentence(appid string, version *int, content string, params ...interface{}) (intent *IntentV2, sentence *SentenceV2WithType, err error)
}

// intentDaoV2 implement interface of intentDaoInterface, which will store for service to use
type intentDaoV2 struct {
	db *sql.DB
}

func (dao intentDaoV2) GetIntents(appid string, version *int, keyword string) (ret []*IntentV2, err error) {
	defer func() {
		util.ShowError(err)
	}()
	dao.checkDB()
	if dao.db == nil {
		return nil, util.ErrDBNotInit
	}

	tx, err := dao.db.Begin()
	if err != nil {
		return
	}
	defer util.ClearTransition(tx)

	conditions := []string{"appid = ?", "status = 0"}
	params := []interface{}{appid}
	if version != nil {
		conditions = append(conditions, "version = ?")
		params = append(params, version)
	} else {
		conditions = append(conditions, "version is NULL")
	}

	// Get intents id and name
	queryStr := fmt.Sprintf("SELECT id, name FROM intents WHERE %s ORDER BY id desc", strings.Join(conditions, " AND "))
	intentRows, err := tx.Query(queryStr, params...)
	if err != nil {
		return
	}
	defer intentRows.Close()

	intentMap := map[int64]*IntentV2{}
	intents := []*IntentV2{}
	for intentRows.Next() {
		intent := &IntentV2{}
		err = intentRows.Scan(&intent.ID, &intent.Name)
		if err != nil {
			return
		}
		intents = append(intents, intent)
		intentMap[intent.ID] = intent
	}

	// If try to get specific version and get empty,
	if version != nil && len(intents) == 0 {
		return nil, sql.ErrNoRows
	}

	// Get train data count, type 0 is Positive, type 1 is Negative
	conditions = []string{"intents.appid = ?", "intents.id = s.intent"}
	params = []interface{}{appid}

	if version != nil {
		conditions = append(conditions, "intents.version = ?")
		params = append(params, version)
	} else {
		conditions = append(conditions, "intents.version is NULL")
	}

	if keyword != "" {
		conditions = append(conditions, "s.sentence like ?")
		params = append(params, getLikeValue(keyword))
	}

	queryStr = fmt.Sprintf(`
		SELECT s.intent, s.type, count(*)
		FROM intent_train_sets as s,intents
		WHERE %s
		GROUP BY s.intent, s.type`, strings.Join(conditions, " AND "))
	countRows, err := tx.Query(queryStr, params...)
	if err != nil {
		return
	}
	defer countRows.Close()

	for countRows.Next() {
		var id int64
		contentType, count := 0, 0
		err = countRows.Scan(&id, &contentType, &count)
		if err != nil {
			return
		}
		if _, ok := intentMap[id]; !ok {
			continue
		}
		switch contentType {
		case 0:
			intentMap[id].PositiveCount = count
		case 1:
			intentMap[id].NegativeCount = count
		}
	}

	err = tx.Commit()
	if err != nil {
		return
	}

	ret = []*IntentV2{}
	for idx := range intents {
		// intent has sentence with keyword
		if intents[idx].PositiveCount > 0 || intents[idx].NegativeCount > 0 {
			ret = append(ret, intents[idx])
			continue
		}

		// ignore case in english when checking name of intent
		intentName := strings.ToLower(intents[idx].Name)
		key := strings.ToLower(keyword)
		if strings.Index(intentName, key) >= 0 {
			ret = append(ret, intents[idx])
		}
	}
	return
}
func (dao intentDaoV2) GetIntent(appid string, intentID int64, keyword string) (ret *IntentV2, err error) {
	defer func() {
		util.ShowError(err)
	}()
	dao.checkDB()
	if dao.db == nil {
		return nil, util.ErrDBNotInit
	}

	queryStr := "SELECT id, name FROM intents WHERE appid = ? AND id = ? AND status = 0"
	intentRow := dao.db.QueryRow(queryStr, appid, intentID)

	intent := IntentV2{}
	err = intentRow.Scan(&intent.ID, &intent.Name)
	if err != nil {
		return
	}

	negativeList := []*SentenceV2{}
	positiveList := []*SentenceV2{}
	// Get train data count, type 0 is Positive, type 1 is Negative
	var contentRows *sql.Rows
	if keyword != "" {
		queryStr = "SELECT id, sentence, type FROM intent_train_sets WHERE intent = ? AND sentence like ?"
		contentRows, err = dao.db.Query(queryStr, intentID, getLikeValue(keyword))
	} else {
		queryStr = "SELECT id, sentence, type FROM intent_train_sets WHERE intent = ?"
		contentRows, err = dao.db.Query(queryStr, intentID)
	}
	if err != nil {
		if err == sql.ErrNoRows {
			err = nil
			intent.Negative = &[]*SentenceV2{}
			intent.Positive = &[]*SentenceV2{}
			ret = &intent
		}
		return
	}
	defer contentRows.Close()

	for contentRows.Next() {
		sentence := SentenceV2{}
		sentenceType := 0
		err = contentRows.Scan(&sentence.ID, &sentence.Content, &sentenceType)
		if sentenceType == typePositive {
			positiveList = append(positiveList, &sentence)
		} else if sentenceType == typeNegative {
			negativeList = append(negativeList, &sentence)
		}
	}
	intent.NegativeCount = len(negativeList)
	intent.PositiveCount = len(positiveList)
	intent.Negative = &negativeList
	intent.Positive = &positiveList
	ret = &intent
	return
}
func (dao intentDaoV2) AddIntent(appid, name string, positive, negative []string) (ret *IntentV2, err error) {
	defer func() {
		util.ShowError(err)
	}()
	dao.checkDB()
	if dao.db == nil {
		return nil, util.ErrDBNotInit
	}

	tx, err := dao.db.Begin()
	if err != nil {
		return
	}
	defer util.ClearTransition(tx)

	intent := IntentV2{}
	timestamp := time.Now().Unix()
	queryStr := "INSERT INTO intents (appid, name, version, updatetime) VALUES (?, ?, NULL, ?)"
	result, err := tx.Exec(queryStr, appid, name, timestamp)
	if err != nil {
		return nil, err
	}

	newID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	intent.ID = newID
	intent.Name = name

	positiveList := []*SentenceV2{}
	negativeList := []*SentenceV2{}

	queryStr = fmt.Sprintf("INSERT INTO intent_train_sets (sentence, intent, type) VALUES (?, %d, ?)", newID)
	sentenceMap := map[int][]string{
		typePositive: positive,
		typeNegative: negative,
	}

	for contentType := range sentenceMap {
		list := sentenceMap[contentType]
		if list != nil {
			for idx := range list {
				content := list[idx]
				result, err = tx.Exec(queryStr, content, contentType)
				if err != nil {
					return
				}
				var sentenceID int64
				sentenceID, err = result.LastInsertId()
				if err != nil {
					return
				}

				sentence := SentenceV2{}
				sentence.ID = sentenceID
				sentence.Content = content
				switch contentType {
				case typePositive:
					positiveList = append(positiveList, &sentence)
				case typeNegative:
					negativeList = append(negativeList, &sentence)
				}
			}
		}
	}
	intent.Negative = &negativeList
	intent.NegativeCount = len(negativeList)
	intent.Positive = &positiveList
	intent.PositiveCount = len(positiveList)

	err = tx.Commit()
	if err != nil {
		return
	}
	ret = &intent
	return
}
func (dao intentDaoV2) ModifyIntent(appid string, intentID int64, name string,
	updateSentence []*SentenceV2WithType, deleteSentences []int64) (err error) {
	defer func() {
		util.ShowError(err)
	}()
	dao.checkDB()
	if dao.db == nil {
		return util.ErrDBNotInit
	}

	if name == "" {
		return util.ErrParameter
	}

	tx, err := dao.db.Begin()
	if err != nil {
		return
	}
	defer util.ClearTransition(tx)

	readOnly, err := checkIntent(tx, appid, intentID)
	if err != nil {
		return
	}
	if readOnly {
		return ErrReadOnlyIntent
	}

	timestamp := time.Now().Unix()

	queryStr := "UPDATE intents SET name = ?, updatetime = ? WHERE id = ?"
	_, err = tx.Exec(queryStr, name, timestamp, intentID)
	if err != nil {
		return
	}

	insertQuery := "INSERT INTO intent_train_sets (sentence, intent, type) VALUES (?, ?, ?)"
	updateQuery := "UPDATE intent_train_sets SET sentence = ? WHERE id = ?"
	for _, sentence := range updateSentence {
		if sentence.ID == 0 {
			_, err = tx.Exec(insertQuery, sentence.Content, intentID, sentence.Type)
		} else {
			_, err = tx.Exec(updateQuery, sentence.Content, sentence.ID)
		}
		if err != nil {
			return
		}
	}

	if len(deleteSentences) > 0 {
		params := make([]interface{}, len(deleteSentences))
		for idx := range deleteSentences {
			params[idx] = deleteSentences[idx]
		}
		queryStr := fmt.Sprintf("DELETE FROM intent_train_sets WHERE id in (?%s)",
			strings.Repeat(",?", len(deleteSentences)-1))
		_, err = tx.Exec(queryStr, params...)
		if err != nil {
			return
		}
	}

	err = tx.Commit()
	return
}
func (dao intentDaoV2) DeleteIntent(appid string, intentID int64) (err error) {
	defer func() {
		util.ShowError(err)
	}()
	dao.checkDB()
	if dao.db == nil {
		return util.ErrDBNotInit
	}

	tx, err := dao.db.Begin()
	if err != nil {
		return
	}
	defer util.ClearTransition(tx)

	readOnly, err := checkIntent(tx, appid, intentID)
	if err == sql.ErrNoRows {
		return nil
	} else if err != nil {
		return err
	}
	if readOnly {
		return ErrReadOnlyIntent
	}

	queryStr := "DELETE FROM intent_train_sets WHERE intent = ?"
	_, err = tx.Exec(queryStr, intentID)
	if err != nil {
		return err
	}

	now := time.Now().Unix()
	fmt.Printf("\n\n%d, %s, %d\n\n", intentID, appid, now)
	queryStr = "UPDATE intents SET status = -1, updatetime = ? WHERE id = ? AND appid = ?"
	_, err = tx.Exec(queryStr, now, intentID, appid)
	if err != nil {
		return err
	}

	err = tx.Commit()
	return
}
func (dao intentDaoV2) DeleteIntents(appid string, intentIDs []int64) (err error) {
	defer func() {
		util.ShowError(err)
	}()
	dao.checkDB()
	if dao.db == nil {
		return util.ErrDBNotInit
	}

	tx, err := dao.db.Begin()
	if err != nil {
		return
	}
	defer util.ClearTransition(tx)

	deletedIDs := []interface{}{}
	for _, id := range intentIDs {
		readOnly, err := checkIntent(tx, appid, id)
		if err != nil {
			continue
		}
		if !readOnly {
			deletedIDs = append(deletedIDs, id)
		}
	}

	if len(deletedIDs) == 0 {
		return ErrReadOnlyIntent
	}

	queryStr := fmt.Sprintf("DELETE FROM intent_train_sets WHERE intent IN (?%s)", strings.Repeat(",?", len(deletedIDs)-1))
	_, err = tx.Exec(queryStr, deletedIDs...)
	if err != nil {
		return err
	}

	now := time.Now().Unix()
	updateParams := []interface{}{now, appid}
	updateParams = append(updateParams, deletedIDs...)
	queryStr = fmt.Sprintf("UPDATE intents SET status = -1, updatetime = ? WHERE appid = ? AND id IN (?%s)", strings.Repeat(",?", len(deletedIDs)-1))
	_, err = tx.Exec(queryStr, updateParams...)
	if err != nil {
		return err
	}

	err = tx.Commit()
	return
}
func (dao intentDaoV2) GetIntentsDetail(appid string, version *int) (ret []*IntentV2, err error) {
	defer func() {
		util.ShowError(err)
	}()
	dao.checkDB()
	if dao.db == nil {
		return nil, util.ErrDBNotInit
	}

	tx, err := dao.db.Begin()
	if err != nil {
		return
	}
	defer util.ClearTransition(tx)

	intents, err := getIntents(tx, appid, version, true)
	if err != nil {
		return
	}

	err = tx.Commit()
	if err != nil {
		return
	}
	ret = intents
	return
}
func (dao intentDaoV2) CommitIntent(appid string) (version int, ret []*IntentV2, err error) {
	defer func() {
		util.ShowError(err)
	}()
	dao.checkDB()
	if dao.db == nil {
		return 0, nil, util.ErrDBNotInit
	}

	tx, err := dao.db.Begin()
	if err != nil {
		return
	}
	defer util.ClearTransition(tx)

	need, err := needCommit(tx, appid)
	if err != nil {
		return
	}
	logger.Trace.Println("Check if need commit: ", need)
	if !need {
		version, err = getLatestVersion(tx, appid)
		logger.Trace.Println("No need commit, version:", version)
		return
	}
	intents, err := getIntents(tx, appid, nil, true)
	now := time.Now().Unix()

	logger.Trace.Printf("Commit %d intents\n", len(intents))
	version, err = commitNewVersion(tx, appid, intents, now)
	if err != nil {
		return
	}
	logger.Trace.Println("Commit new intent, version:", version)

	queryStr := "DELETE FROM intents WHERE status = -1 AND appid = ?"
	_, err = tx.Exec(queryStr, appid)
	if err != nil {
		return
	}

	err = tx.Commit()
	if err != nil {
		return
	}
	ret = intents
	return
}
func (dao intentDaoV2) NeedCommit(appid string) (ret bool, err error) {
	defer func() {
		util.ShowError(err)
	}()
	dao.checkDB()
	if dao.db == nil {
		return false, util.ErrDBNotInit
	}
	return needCommit(dao.db, appid)
}
func (dao intentDaoV2) GetVersionInfo(appid string, version int) (ret *VersionInfoV2, err error) {
	defer func() {
		util.ShowError(err)
	}()
	dao.checkDB()
	if dao.db == nil {
		return nil, util.ErrDBNotInit
	}
	queryStr := `
		SELECT ie_model_id, re_model_id, in_used, start_train, end_train, sentence_count, result
		FROM intent_versions
		WHERE appid = ? AND version = ?`
	info := VersionInfoV2{}
	inUse, count := 0, 0
	err = dao.db.QueryRow(queryStr, appid, version).Scan(
		&info.IntentEngineModel, &info.RuleEngineModel, &inUse,
		&info.TrainStartTime, &info.TrainEndTime, &count, &info.TrainResult)
	info.InUse = inUse != 0
	if info.TrainStartTime == nil {
		info.Progress = 0
	} else if info.TrainEndTime != nil {
		info.Progress = 100
	} else {
		const sentencePerSecond = 10 // TODO: check this average value
		predictSeconds := count / sentencePerSecond
		if predictSeconds == 0 {
			info.Progress = 99
		} else {
			now := time.Now().Unix()
			info.Progress = int(now-*info.TrainStartTime) * 100 / predictSeconds
		}
	}
	return &info, nil
}
func (dao intentDaoV2) GetLatestVersion(appid string) (version int, err error) {
	defer func() {
		util.ShowError(err)
	}()
	dao.checkDB()
	if dao.db == nil {
		return 0, util.ErrDBNotInit
	}
	return getLatestVersion(dao.db, appid)
}
func (dao intentDaoV2) UpdateLatestIntents(appid string, intents []*IntentV2) (err error) {
	defer func() {
		util.ShowError(err)
	}()
	dao.checkDB()
	if dao.db == nil {
		return util.ErrDBNotInit
	}

	tx, err := dao.db.Begin()
	if err != nil {
		return
	}
	defer util.ClearTransition(tx)

	queryStr := "DELETE FROM intents WHERE version is NULL AND appid = ?"
	_, err = tx.Exec(queryStr, appid)
	if err != nil {
		return
	}

	// delete all sentence without parent intent
	queryStr = `
	DELETE s FROM intent_train_sets as s RIGHT JOIN
	(
		SELECT t1.id as id, intents.id as iid
		FROM intent_train_sets as t1
		LEFT JOIN intents ON t1.intent = intents.id
	) as t2
	ON t2.iid is NULL AND s.id = t2.id`
	_, err = tx.Exec(queryStr)
	if err != nil {
		return
	}

	now := time.Now().Unix()
	err = insertIntents(tx, appid, nil, intents, now)
	if err != nil {
		return
	}

	err = tx.Commit()
	return
}
func (dao intentDaoV2) UpdateVersionStart(version int, start int64, modelID string) (err error) {
	defer func() {
		util.ShowError(err)
	}()
	dao.checkDB()
	if dao.db == nil {
		return util.ErrDBNotInit
	}

	queryStr := "UPDATE intent_versions SET ie_model_id = ?, start_train = ?, end_train = NULL, result = 0 WHERE version = ?"
	_, err = dao.db.Exec(queryStr, modelID, start, version)
	return
}
func (dao intentDaoV2) UpdateVersionStatus(version int, end int64, status int) (err error) {
	defer func() {
		util.ShowError(err)
	}()
	dao.checkDB()
	if dao.db == nil {
		return util.ErrDBNotInit
	}

	queryStr := "UPDATE intent_versions SET end_train = ?, result = ? WHERE version = ?"
	_, err = dao.db.Exec(queryStr, end, status, version)
	return
}

func (dao intentDaoV2) SearchIntentOfSentence(appid string, version *int, content string, params ...interface{}) (intent *IntentV2, sentence *SentenceV2WithType, err error) {
	defer func() {
		util.ShowError(err)
	}()
	dao.checkDB()
	if dao.db == nil {
		err = util.ErrDBNotInit
		return
	}

	queryParams := []interface{}{
		appid,
		content,
	}
	conditions := []string{
		"i.appid = ?",
		"s.sentence = ?",
		"i.status = 0",
	}

	if version == nil {
		conditions = append(conditions, "i.version IS NULL")
	} else {
		conditions = append(conditions, "i.version = ?")
		queryParams = append(queryParams, version)
	}

	if len(params) > 0 {
		conditions = append(conditions, "s.type = ?")
		queryParams = append(queryParams, params[0].(int))
	}
	queryStr := fmt.Sprintf(`
		SELECT i.id, i.name, s.id, s.type
		FROM intents AS i, intent_train_sets AS s
		WHERE %s`, strings.Join(conditions, " AND "))

	retIntent := IntentV2{}
	retSentence := SentenceV2WithType{}
	retSentence.Content = content

	err = dao.db.QueryRow(queryStr, queryParams...).Scan(
		&retIntent.ID, &retIntent.Name, &retSentence.ID, &retSentence.Type)
	if err != nil {
		return
	}
	return &retIntent, &retSentence, nil
}

func commitNewVersion(tx db, appid string, intents []*IntentV2, now int64) (version int, err error) {
	defer func() {
		util.ShowError(err)
	}()
	if tx == nil {
		return 0, util.ErrDBNotInit
	}

	sentenceCount := 0
	for _, intent := range intents {
		sentenceCount += intent.NegativeCount
		sentenceCount += intent.PositiveCount
	}

	queryStr := `
		INSERT INTO intent_versions (appid, commit_time, sentence_count)
		VALUES (?, ?, ?)`
	result, err := tx.Exec(queryStr, appid, now, sentenceCount)
	if err != nil {
		return
	}
	latestVersion, err := result.LastInsertId()
	if err != nil {
		return
	}

	version = int(latestVersion)
	err = insertIntents(tx, appid, &version, intents, now)
	return
}

func insertIntents(tx db, appid string, version *int, intents []*IntentV2, now int64) (err error) {
	if tx == nil {
		return util.ErrDBNotInit
	}
	queryStr := `
		INSERT INTO intents (appid, name, version, updatetime)
		VALUES (?, ?, ?, ?)`
	sentenceValues := []interface{}{}
	var result sql.Result
	for _, intent := range intents {
		result, err = tx.Exec(queryStr, appid, intent.Name, version, now)
		if err != nil {
			return
		}
		intent.ID, err = result.LastInsertId()
		if err != nil {
			return
		}
		if intent.Positive != nil {
			for _, info := range *intent.Positive {
				sentenceValues = append(sentenceValues, info.Content, intent.ID, typePositive)
			}
		}
		if intent.Negative != nil {
			for _, info := range *intent.Negative {
				sentenceValues = append(sentenceValues, info.Content, intent.ID, typeNegative)
			}
		}
	}

	if len(sentenceValues) > 0 {
		start := 0
		recordPerOp := 30000
		for {
			end := start + recordPerOp
			if end > len(sentenceValues) {
				end = len(sentenceValues)
			}

			useParam := sentenceValues[start:end]
			queryStr = fmt.Sprintf(`
				INSERT INTO intent_train_sets (sentence, intent, type)
				VALUES (?, ?, ?)%s`, strings.Repeat(",(?, ?, ?)", len(useParam)/3-1))
			_, err = tx.Exec(queryStr, useParam...)
			if err != nil {
				return
			}

			if end == len(sentenceValues) {
				break
			}
			start = end
		}
	}
	return
}

// needCommit will check there is existed intent edited after latest commit
func needCommit(tx db, appid string) (ret bool, err error) {
	if tx == nil {
		return false, util.ErrDBNotInit
	}

	queryStr := `
		SELECT count(*) as cnt, max(commit_time) as commit_time
		FROM intent_versions
		WHERE appid = ?
	`
	var lastUpdatePtr *int
	versionCnt := 0
	err = tx.QueryRow(queryStr, appid).Scan(&versionCnt, &lastUpdatePtr)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}
	if err == sql.ErrNoRows || versionCnt == 0 {
		logger.Trace.Println("No any version, need commit")
		return true, nil
	}

	lastUpdate := 0
	if lastUpdatePtr != nil {
		lastUpdate = *lastUpdatePtr
	}
	queryStr = `
		SELECT count(*) FROM intents
		WHERE appid = ? AND updatetime > ? AND version is NULL
	`
	count := 0
	err = tx.QueryRow(queryStr, appid, lastUpdate).Scan(&count)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}
	if err == sql.ErrNoRows || count == 0 {
		logger.Trace.Printf("No any intents is modified after %d, no need commit\n", lastUpdate)
		return false, nil
	}
	logger.Trace.Printf("Get %d intents update time bigger then %d\n", count, lastUpdate)
	return true, nil
}
func getLatestVersion(tx db, appid string) (version int, err error) {
	if tx == nil {
		return 0, util.ErrDBNotInit
	}

	var value *int
	queryStr := `
		SELECT max(version)
		FROM intent_versions
		WHERE appid = ?`
	err = tx.QueryRow(queryStr, appid).Scan(&value)
	if err != nil {
		return
	}
	if value == nil {
		err = sql.ErrNoRows
	} else {
		version = *value
	}
	return
}
func getIntents(tx db, appid string, version *int, detail bool) (ret []*IntentV2, err error) {
	if tx == nil {
		return nil, util.ErrDBNotInit
	}
	intents, err := getIntentsNameOnly(tx, appid, version)
	if err != nil {
		return
	}
	if detail {
		err = fillIntentDetail(tx, intents)
	} else {
		err = fillIntentCount(tx, intents)
	}
	return intents, nil
}
func getIntentsNameOnly(tx db, appid string, version *int) (ret []*IntentV2, err error) {
	if tx == nil {
		return nil, util.ErrDBNotInit
	}
	conditions := []string{"appid = ?", "status = 0"}
	params := []interface{}{appid}
	if version != nil {
		conditions = append(conditions, "version = ?")
		params = append(params, version)
	} else {
		conditions = append(conditions, "version is NULL")
	}

	// Get intents id and name
	queryStr := fmt.Sprintf("SELECT id, name FROM intents WHERE %s ORDER BY id desc", strings.Join(conditions, " AND "))
	intentRows, err := tx.Query(queryStr, params...)
	if err != nil {
		return
	}
	defer intentRows.Close()

	intents := []*IntentV2{}
	for intentRows.Next() {
		intent := &IntentV2{}
		err = intentRows.Scan(&intent.ID, &intent.Name)
		if err != nil {
			return
		}
		intents = append(intents, intent)
	}

	// If try to get specific version and get empty,
	if version != nil && len(intents) == 0 {
		return nil, sql.ErrNoRows
	}
	return intents, nil
}
func fillIntentDetail(tx db, intents []*IntentV2) (err error) {
	if tx == nil {
		return util.ErrDBNotInit
	}
	if len(intents) == 0 {
		return
	}

	intentSentence := map[int64]map[int][]*SentenceV2{}
	params := []interface{}{}
	for idx := range intents {
		intentSentence[intents[idx].ID] = map[int][]*SentenceV2{}
		intentSentence[intents[idx].ID][typePositive] = []*SentenceV2{}
		intentSentence[intents[idx].ID][typeNegative] = []*SentenceV2{}

		params = append(params, intents[idx].ID)
	}

	queryStr := fmt.Sprintf(`
		SELECT intent, id, sentence, type
		FROM intent_train_sets
		WHERE intent in (?%s)`, strings.Repeat(",?", len(intents)-1))
	rows, err := tx.Query(queryStr, params...)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var intentID, id int64
		var sentence string
		var sentenceType int
		err = rows.Scan(&intentID, &id, &sentence, &sentenceType)
		if err != nil {
			return
		}

		if _, ok := intentSentence[intentID]; !ok {
			continue
		}
		sentenceMap := intentSentence[intentID]

		if _, ok := sentenceMap[sentenceType]; !ok {
			continue
		}
		sentenceMap[sentenceType] = append(sentenceMap[sentenceType], &SentenceV2{
			ID:      id,
			Content: sentence,
		})
	}

	// fill data back
	for idx := range intents {
		intent := intents[idx]
		if _, ok := intentSentence[intent.ID]; ok {
			p := intentSentence[intent.ID][typePositive]
			n := intentSentence[intent.ID][typeNegative]
			intent.Positive = &p
			intent.Negative = &n
			intent.PositiveCount = len(p)
			intent.NegativeCount = len(n)
		}
	}

	return
}
func fillIntentCount(tx db, intents []*IntentV2) (err error) {
	if tx == nil {
		return util.ErrDBNotInit
	}
	if len(intents) == 0 {
		return
	}

	intentMap := map[int64]*IntentV2{}
	params := []interface{}{}
	for idx := range intents {
		intentMap[intents[idx].ID] = intents[idx]
		params = append(params, intents[idx].ID)
	}

	// Get train data count, type 0 is Positive, type 1 is Negative
	queryStr := fmt.Sprintf(`
		SELECT s.intent, s.type, count(*)
		FROM intent_train_sets as s,intents
		WHERE intents.id = s.intent AND intents.id in (?%s)
		GROUP BY s.intent, s.type`, strings.Repeat(",?", len(intents)-1))
	countRows, err := tx.Query(queryStr, params...)
	if err != nil {
		return
	}
	defer countRows.Close()

	for countRows.Next() {
		var id int64
		contentType, count := 0, 0
		err = countRows.Scan(&id, &contentType, &count)
		if err != nil {
			return
		}
		if _, ok := intentMap[id]; !ok {
			continue
		}
		switch contentType {
		case 0:
			intentMap[id].PositiveCount = count
		case 1:
			intentMap[id].NegativeCount = count
		}
	}
	return
}

func (dao *intentDaoV2) checkDB() {
	if dao.db == nil {
		dao.db = util.GetMainDB()
	}
}

func checkIntent(tx db, appid string, intentID int64) (readOnly bool, err error) {
	// when error happen, return false
	if tx == nil {
		return false, errors.New("error parameter")
	}
	queryStr := "SELECT version FROM intents WHERE appid = ? AND id = ? AND status = 0"
	row := tx.QueryRow(queryStr, appid, intentID)

	var version *int
	err = row.Scan(&version)
	if err != nil {
		return false, err
	}

	return version != nil, nil
}

func getLikeValue(val string) string {
	return fmt.Sprintf("%%%s%%", strings.Replace(val, "%", "\\%", -1))
}

type db interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Exec(query string, args ...interface{}) (sql.Result, error)
}
