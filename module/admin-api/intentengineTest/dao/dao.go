package dao

import (
	"database/sql"
	"errors"
	"fmt"

	//"regexp"
	"strings"
	"time"

	"emotibot.com/emotigo/module/admin-api/auth"
	"emotibot.com/emotigo/module/admin-api/intentengineTest/data"
	"emotibot.com/emotigo/module/admin-api/util/localemsg"
	"emotibot.com/emotigo/pkg/logger"
)

const (
	IntentsTable             = "intents"
	IntentTrainSetsTable     = "intent_train_sets"
	IntentVersionsTable      = "intent_versions"
	IntentTestIntentsTable   = "intent_test_intents"
	IntentTestSentencesTable = "intent_test_sentences"
	IntentTestVersionsTable  = "intent_test_versions"
)

func (dao IntentTestDao) GetIntentTests(appID string) (results *data.IntentTestResults, err error) {
	if dao.db == nil {
		return nil, ErrDBNotInit
	}

	// Retrieve latest test records
	queryStr := fmt.Sprintf(`
		SELECT tv.id, tv.start_time, tv.intents_count, tv.sentences_count,
			tv.tester, tv.true_positives, tv.false_positives,
			tv.true_negatives, tv.false_negatives, iv.version, iv.commit_time,
			i.intents_count, iv.sentence_count, tv.saved
		FROM %s AS tv
		INNER JOIN %s as iv
		ON tv.ie_model_id = iv.ie_model_id
		INNER JOIN (
			SELECT COUNT(1) AS intents_count, version
			FROM %s
			WHERE appid = ?
			GROUP BY version) AS i
		ON i.version = iv.version
		WHERE app_id = ? AND status = ?
		ORDER BY tv.id DESC
		LIMIT ?`, IntentTestVersionsTable, IntentVersionsTable, IntentsTable)
	rows, err := dao.db.Query(queryStr, appID, appID, data.TestStatusTested, data.MaxLatestTests)
	if err != nil {
		return nil, err
	}

	var tester string
	latests := make([]*data.IntentTestResult, 0)

	for rows.Next() {
		latest := data.IntentTestResult{}
		err = rows.Scan(&latest.IntentTest.ID, &latest.IntentTest.UpdatedTime,
			&latest.IntentTest.TestIntentsCount,
			&latest.IntentTest.TestSentencesCount, &tester,
			&latest.IntentTest.TruePositives, &latest.IntentTest.FalsePositives,
			&latest.IntentTest.TrueNegatives, &latest.IntentTest.FalseNegatives,
			&latest.IntentModel.Version, &latest.IntentModel.UpdatedTime,
			&latest.IntentModel.IntentsCount, &latest.IntentModel.SentencesCount,
			&latest.IntentTest.Saved)
		if err != nil {
			return nil, err
		}

		testerName, err := auth.GetUserName(tester)
		if err != nil && err != sql.ErrNoRows {
			return nil, err
		}
		latest.IntentTest.Tester = testerName

		latests = append(latests, &latest)
	}

	// Retrieve saved test records
	saves := make([]*data.IntentTestResult, 0)

	queryStr = fmt.Sprintf(`
		SELECT tv.id, tv.name, tv.start_time, tv.intents_count, tv.sentences_count,
			tv.tester, tv.true_positives, tv.false_positives,
			tv.true_negatives, tv.false_negatives, iv.version, iv.commit_time,
			i.intents_count, iv.sentence_count, tv.saved
		FROM %s AS tv
		INNER JOIN %s as iv
		ON tv.ie_model_id = iv.ie_model_id
		INNER JOIN (
			SELECT COUNT(1) AS intents_count, version
			FROM %s
			WHERE appid = ?
			GROUP BY version) AS i
		ON i.version = iv.version
		WHERE app_id = ? AND status = ? AND saved = 1
		ORDER BY tv.id DESC`, IntentTestVersionsTable, IntentVersionsTable,
		IntentsTable)
	rows, err = dao.db.Query(queryStr, appID, appID, data.TestStatusTested)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		saved := data.IntentTestResult{}
		err = rows.Scan(&saved.IntentTest.ID, &saved.IntentTest.Name,
			&saved.IntentTest.UpdatedTime, &saved.IntentTest.TestIntentsCount,
			&saved.IntentTest.TestSentencesCount, &tester,
			&saved.IntentTest.TruePositives, &saved.IntentTest.FalsePositives,
			&saved.IntentTest.TrueNegatives, &saved.IntentTest.FalseNegatives,
			&saved.IntentModel.Version, &saved.IntentModel.UpdatedTime,
			&saved.IntentModel.IntentsCount, &saved.IntentModel.SentencesCount,
			&saved.IntentTest.Saved)
		if err != nil {
			return nil, err
		}

		testerName, err := auth.GetUserName(tester)
		if err != nil && err != sql.ErrNoRows {
			return nil, err
		}
		saved.IntentTest.Tester = testerName

		saves = append(saves, &saved)
	}

	results = &data.IntentTestResults{
		Latest: latests,
		Saved:  saves,
	}

	return
}

func (dao IntentTestDao) GetIntentTest(appID string, version int64,
	keyword string, locale string) (result *data.IntentTest, err error) {
	if dao.db == nil {
		return nil, ErrDBNotInit
	}

	var testAppID string
	var tester string
	var intentsCount sql.NullInt64
	result = &data.IntentTest{}

	// Note: 'i.intent_count' may be NULL after left outer join
	queryStr := fmt.Sprintf(`
		SELECT tv.id, tv.app_id, tv.name, tv.start_time, tv.intents_count, 
			tv.sentences_count, tv.true_positives, tv.false_positives, 
			tv.true_negatives, tv.false_negatives, tv.tester, tv.saved, 
			i.intent_count, iv.version, iv.sentence_count, iv.commit_time
		FROM %s AS tv
		INNER JOIN %s AS iv
		ON tv.ie_model_id = iv.ie_model_id
		LEFT JOIN (
			SELECT COUNT(1) AS intent_count, version
			FROM %s
			WHERE appid = ?
			GROUP BY version
		) AS i
		ON i.version = iv.version
		WHERE tv.id = ?`, IntentTestVersionsTable, IntentVersionsTable,
		IntentsTable)
	err = dao.db.QueryRow(queryStr, appID, version).Scan(&result.ID,
		&testAppID, &result.Name, &result.UpdatedTime, &result.TestIntentsCount,
		&result.TestSentencesCount, &result.TruePositives, &result.FalsePositives,
		&result.TrueNegatives, &result.FalseNegatives, &tester,
		&result.Saved, &intentsCount, &result.IEModelVersion, &result.SentencesCount,
		&result.IEModelUpdateTime)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, data.ErrTestNotFound
		}
		return nil, err
	}

	// Validate intent test does belong to robot
	if testAppID != appID {
		return nil, data.ErrTestNotFound
	}

	testerName, err := auth.GetUserName(tester)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	result.Tester = testerName

	var _intentCounts int64
	if intentsCount.Valid {
		_intentCounts = intentsCount.Int64
	}
	result.IntentsCount = &_intentCounts

	// Note: 's.sentence' may be NULL after left outer join
	queryStr = fmt.Sprintf(`
		SELECT ti.id, ti.intent_name, ts.sentence, ts.result
		FROM %s AS ti
		LEFT JOIN %s AS ts
		ON ts.test_intent = ti.id
		WHERE ti.version = ?`, IntentTestIntentsTable, IntentTestSentencesTable)
	queryParams := []interface{}{version}

	if keyword != "" {
		queryStr = fmt.Sprintf("%s AND (ti.intent_name LIKE ? OR ts.sentence LIKE ?)", queryStr)
		queryParams = append(queryParams, getLikeValue(keyword), getLikeValue(keyword))
	}

	rows, err := dao.db.Query(queryStr, queryParams...)
	if err != nil {
		return nil, err
	}

	testIntents := make([]*data.IntentTestIntent, 0)
	testIntentsMap := make(map[int64]*data.IntentTestIntent)

	for rows.Next() {
		var id int64
		var intentName string
		var sentenceResult sql.NullInt64
		var resultType bool
		var _intentName, _sentence sql.NullString

		err = rows.Scan(&id, &_intentName, &_sentence, &sentenceResult)
		if err != nil {
			return nil, err
		}

		if !_intentName.Valid {
			// Negative test intent
			intentName = localemsg.Get(locale, "NegativeTestIntentName")
			resultType = false
		} else {
			intentName = _intentName.String
			resultType = true
		}

		testIntent, ok := testIntentsMap[id]
		if !ok {
			var positivesCount int64 = 0
			testIntent = &data.IntentTestIntent{
				ID:             id,
				IntentName:     &intentName,
				PositivesCount: &positivesCount,
				SentencesCount: 0,
				Type:           &resultType,
			}
			testIntentsMap[id] = testIntent
			testIntents = append(testIntents, testIntent)
		}

		if _sentence.Valid {
			if keyword != "" {
				if strings.Contains(_sentence.String, keyword) {
					testIntent.SentencesCount++
					if sentenceResult.Valid && sentenceResult.Int64 == data.TestResultCorrect {
						*testIntent.PositivesCount = *testIntent.PositivesCount + 1
					}
				}
			} else {
				testIntent.SentencesCount++
				if sentenceResult.Valid && sentenceResult.Int64 == data.TestResultCorrect {
					*testIntent.PositivesCount = *testIntent.PositivesCount + 1
				}
			}
		}
	}

	result.TestIntents = testIntents
	return
}

func (dao IntentTestDao) GetLatestIntents(appID string,
	keyword string, locale string) (results []*data.IntentTestIntent,
	err error) {
	if dao.db == nil {
		return nil, ErrDBNotInit
	}

	// Note: 'ts.sentence' may be NULL after left outer join
	queryStr := fmt.Sprintf(`
		SELECT ti.id, ti.intent_name, ts.sentence
		FROM %s AS ti
		LEFT JOIN %s AS ts
		ON ts.test_intent = ti.id
		WHERE ti.app_id = ? AND ti.version IS NULL`, IntentTestIntentsTable,
		IntentTestSentencesTable)
	queryParams := []interface{}{appID}

	if keyword != "" {
		queryStr = fmt.Sprintf("%s AND (ti.intent_name LIKE ? OR ts.sentence LIKE ?)", queryStr)
		queryParams = append(queryParams, getLikeValue(keyword), getLikeValue(keyword))
	}

	rows, err := dao.db.Query(queryStr, queryParams...)
	if err != nil {
		return nil, err
	}

	results = make([]*data.IntentTestIntent, 0)
	testIntentsMap := make(map[int64]*data.IntentTestIntent)

	for rows.Next() {
		var id int64
		var intentName string
		var resultType bool
		var _intentName, _sentence sql.NullString

		err = rows.Scan(&id, &_intentName, &_sentence)
		if err != nil {
			return nil, err
		}

		if !_intentName.Valid {
			// Negative test intent
			intentName = localemsg.Get(locale, "NegativeTestIntentName")
			resultType = false
		} else {
			intentName = _intentName.String
			resultType = true
		}

		result, ok := testIntentsMap[id]
		if !ok {
			result = &data.IntentTestIntent{
				ID:             id,
				IntentName:     &intentName,
				SentencesCount: 0,
				Type:           &resultType,
			}
			testIntentsMap[id] = result
			results = append(results, result)
		}

		if _sentence.Valid {
			if keyword != "" {
				if strings.Contains(_sentence.String, keyword) {
					result.SentencesCount++
				}
			} else {
				result.SentencesCount++
			}
		}
	}

	return
}

func (dao IntentTestDao) GetIntentTestIntentSentences(testIntentID int64,
	keyword string) (results []*data.IntentTestSentence,
	err error) {
	if dao.db == nil {
		return nil, ErrDBNotInit
	}

	exist, err := dao.checkTestIntentExist(testIntentID)
	if err != nil {
		return nil, err
	} else if !exist {
		return nil, data.ErrTestIntentNotFound
	}

	results = make([]*data.IntentTestSentence, 0)

	queryStr := fmt.Sprintf(`
		SELECT id, sentence, result, score, answer
		FROM %s
		WHERE test_intent = ?`, IntentTestSentencesTable)
	queryParams := []interface{}{testIntentID}

	if keyword != "" {
		queryStr = fmt.Sprintf("%s AND sentence LIKE ?", queryStr)
		queryParams = append(queryParams, getLikeValue(keyword))
	}

	rows, err := dao.db.Query(queryStr, queryParams...)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var score sql.NullInt64
		var answer sql.NullString
		result := data.IntentTestSentence{}
		err = rows.Scan(&result.ID, &result.Sentence, &result.Result,
			&score, &answer)
		if err != nil {
			return nil, err
		}

		if score.Valid {
			result.Score = &score.Int64
		}

		if answer.Valid {
			result.Answer = &answer.String
		}

		results = append(results, &result)
	}

	return
}

func (dao IntentTestDao) GetLatestIntentTestSentences(appID string) (results []*data.IntentTestSentence,
	err error) {
	if err != nil {
		return nil, err
	}

	results = make([]*data.IntentTestSentence, 0)

	queryStr := fmt.Sprintf(`
		SELECT ts.id, ti.intent_name, ts.sentence
		FROM %s ts
		INNER JOIN %s AS ti
		ON ti.id = ts.test_intent
		WHERE app_id = ? AND version IS NULL`, IntentTestSentencesTable,
		IntentTestIntentsTable)
	rows, err := dao.db.Query(queryStr, appID)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var intentName sql.NullString
		intentSentence := data.IntentTestSentence{}
		err = rows.Scan(&intentSentence.ID, &intentName, &intentSentence.Sentence)
		if err != nil {
			return nil, err
		}

		if intentName.Valid {
			intentSentence.IntentName = intentName.String
		}

		results = append(results, &intentSentence)
	}

	return results, nil
}

func (dao IntentTestDao) ModifyIntentTestSentences(testIntentID int64,
	updateList []*data.UpdateCmd, deleteList []int64) (err error) {
	if dao.db == nil {
		return ErrDBNotInit
	}

	exist, err := dao.checkTestIntentExist(testIntentID)
	if err != nil {
		return err
	} else if !exist {
		return data.ErrTestIntentNotFound
	}

	queryStr := fmt.Sprintf(`
		UPDATE %s
		SET updated_time = ?
		WHERE id = ?`, IntentTestIntentsTable)

	insertQueryStr := fmt.Sprintf(`
		INSERT INTO %s (sentence, test_intent)
		VALUES (?, ?)`, IntentTestSentencesTable)

	updateQueryStr := fmt.Sprintf(`
		UPDATE %s
		SET sentence = ?, result = 0, score = NULL, answer = NULL
		WHERE id = ?`, IntentTestSentencesTable)

	tx, _ := dao.db.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	// Update test intent updated time
	now := time.Now().Unix()
	_, err = tx.Exec(queryStr, now, testIntentID)
	if err != nil {
		return
	}

	for _, cmd := range updateList {
		if cmd.ID == 0 {
			// Insert new test sentence
			_, err = tx.Exec(insertQueryStr, cmd.Content, testIntentID)
		} else {
			// Update test sentence
			_, err = tx.Exec(updateQueryStr, cmd.Content, cmd.ID)
		}

		if err != nil {
			return err
		}
	}

	if len(deleteList) > 0 {
		queryParams := make([]interface{}, len(deleteList))
		for i := range deleteList {
			queryParams[i] = deleteList[i]
		}

		deleteQueryStr := fmt.Sprintf(`
			DELETE
			FROM %s
			WHERE id IN (?%s)`, IntentTestSentencesTable,
			strings.Repeat(", ?", len(deleteList)-1))
		_, err = tx.Exec(deleteQueryStr, queryParams...)
		if err != nil {
			return err
		}
	}

	return nil
}

func (dao IntentTestDao) TestIntents(appID string, userID string,
	ieModelID string) (version int64, status int64, err error) {
	if dao.db == nil {
		return 0, 0, ErrDBNotInit
	}

	now := time.Now()
	var tx *sql.Tx
	var id, startTime int64

	createNewTest := func() (version int64, status int64, err error) {
		if tx == nil {
			tx, _ = dao.db.Begin()
			defer func() {
				if err != nil {
					tx.Rollback()
				} else {
					tx.Commit()
				}
			}()
		}

		var intentsCount, sentencesCount int64
		var _sentencesCount sql.NullInt64

		// Note: 'ts.sentences_count' may be NULL after left outer join
		queryStr := fmt.Sprintf(`
			SELECT COUNT(1), SUM(ts.sentences_count)
			FROM %s AS ti
			LEFT JOIN (
				SELECT COUNT(1) AS sentences_count, test_intent
				FROM %s
				GROUP BY test_intent) AS ts
			ON ts.test_intent = ti.id
			WHERE ti.app_id = ? AND ti.version IS NULL`,
			IntentTestIntentsTable, IntentTestSentencesTable)
		err = tx.QueryRow(queryStr, appID).Scan(&intentsCount, &_sentencesCount)
		if err != nil {
			return 0, 0, err
		}

		if _sentencesCount.Valid {
			sentencesCount = _sentencesCount.Int64
		}

		queryStr = fmt.Sprintf(`
			INSERT INTO %s (app_id, tester, ie_model_id,
				intents_count, sentences_count, start_time, status)
			VALUES (?, ?, ?, ?, ?, ?, ?)`, IntentTestVersionsTable)
		result, err := tx.Exec(queryStr, appID, userID, ieModelID, intentsCount,
			sentencesCount, now.Unix(), data.TestStatusTesting)
		if err != nil {
			return 0, 0, err
		}

		version, err = result.LastInsertId()
		if err != nil {
			return 0, 0, err
		}

		// Test task created, return pending status for service
		// to run actual intent test
		return version, data.TestStatusPending, nil
	}

	// Check whether currently there's already a running test task exists
	queryStr := fmt.Sprintf(`
		SELECT id, start_time
		FROM %s
		WHERE app_id = ? AND status = ?
		ORDER BY id DESC
		LIMIT 1`, IntentTestVersionsTable)
	err = dao.db.QueryRow(queryStr, appID, data.TestStatusTesting).Scan(&id,
		&startTime)
	if err != nil {
		if err == sql.ErrNoRows {
			// No running test task exists
			return createNewTest()
		}
		return 0, 0, err
	}

	// Currently there's already a running test task exists,
	// check whether it has expired or not
	testStartTime := time.Unix(startTime, 0)

	if now.Sub(testStartTime) > data.TestTaskExpiredDuration {
		// Test task has expired, update its end time, status, error message
		// and create a new test task
		tx, _ = dao.db.Begin()
		defer func() {
			if err != nil {
				tx.Rollback()
			} else {
				tx.Commit()
			}
		}()

		queryStr = fmt.Sprintf(`
			UPDATE %s
			SET end_time = ?, status = ?, message = ?
			WHERE id = ?`, IntentTestVersionsTable)
		_, err = tx.Exec(queryStr, now.Unix(), data.TestStatusFailed,
			data.ErrTestTaskExpired.Error(), id)
		if err != nil {
			return 0, 0, err
		}

		return createNewTest()
	}

	// Test task not expired yet, YOU SHALL NOT PASS!!
	return version, status, data.ErrTestTaskInProcess
}

func (dao IntentTestDao) GetUsableModels(appID string) (models *data.UseableModels,
	err error) {
	if dao.db == nil {
		return nil, ErrDBNotInit
	}

	tx, _ := dao.db.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	testIntents, err := GetLatestIntentTestIntentsWithTx(tx, appID)
	if err != nil {
		return nil, err
	}

	intentNamesMap := make(map[int64][]string)
	models = data.NewUseableModels()

	intentNamesQueryStr := fmt.Sprintf(`
		SELECT name
		FROM %s
		WHERE version = ?`, IntentsTable)

	getIntentNames := func(intentVersion int64) ([]string, error) {
		if intentNames, ok := intentNamesMap[intentVersion]; ok {
			return intentNames, nil
		}

		rows, err := tx.Query(intentNamesQueryStr, intentVersion)
		if err != nil {
			return nil, err
		}

		intentNames := []string{}

		for rows.Next() {
			var intentName string
			err = rows.Scan(&intentName)
			if err != nil {
				return nil, err
			}

			intentNames = append(intentNames, intentName)
		}

		intentNamesMap[intentVersion] = intentNames
		return intentNames, nil
	}

	var ieModel *data.IEModel
	var modelID sql.NullString

	// In used Intent Engine model
	queryStr := fmt.Sprintf(`
		SELECT iv.version, iv.ie_model_id, iv.start_train, i.intent_count, iv.sentence_count
		FROM %s AS iv
		INNER JOIN (
			SELECT COUNT(1) AS intent_count, version
			FROM %s
			WHERE appid = ?
			GROUP BY version
		) AS i
		ON i.version = iv.version
		WHERE appid = ? AND in_used = 1 AND result = 1
		ORDER BY iv.version DESC
		LIMIT 1`, IntentVersionsTable, IntentsTable)

	ieModel = data.NewIEModel()
	err = tx.QueryRow(queryStr, appID, appID).Scan(&ieModel.IntentVersion, &modelID,
		&ieModel.TrainTime, &ieModel.IntentsCount, &ieModel.SentencesCount)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	if err != sql.ErrNoRows && modelID.Valid {
		ieModel.ModelID = modelID.String
		models.InUsed = ieModel

		// Get intent names of in used Intent Enginge model
		// and diff them with test intent's intent names
		intentNames, err := getIntentNames(ieModel.IntentVersion)
		if err != nil {
			return nil, err
		}

		intentDiffs, testIntentDiffs := diffIntentTestIntents(testIntents, intentNames)
		models.InUsed.Diffs.Intents = append(models.InUsed.Diffs.Intents, intentDiffs...)
		models.InUsed.Diffs.TestIntents = append(models.InUsed.Diffs.TestIntents, testIntentDiffs...)
		models.InUsed.DiffsCount = int64(len(intentDiffs) + len(testIntentDiffs))
	}

	// Recent trained Intent Engine models
	queryStr = fmt.Sprintf(`
		SELECT iv.version, iv.ie_model_id, iv.start_train, i.intent_count, iv.sentence_count
		FROM %s AS iv
		INNER JOIN (
			SELECT COUNT(1) AS intent_count, version
			FROM %s
			WHERE appid = ?
			GROUP BY version
		) AS i
		ON i.version = iv.version
		WHERE appid = ? AND result = 1
		ORDER BY iv.version DESC
		LIMIT ?`, IntentVersionsTable, IntentsTable)
	rows, err := tx.Query(queryStr, appID, appID, data.MaxRecentTrained)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	for rows.Next() {
		ieModel := data.NewIEModel()
		err = rows.Scan(&ieModel.IntentVersion, &modelID, &ieModel.TrainTime,
			&ieModel.IntentsCount, &ieModel.SentencesCount)
		if err != nil {
			return nil, err
		}

		if modelID.Valid {
			ieModel.ModelID = modelID.String
			models.RecentTrained = append(models.RecentTrained, ieModel)
		}
	}

	// Get intent names of in recent trained Intent Enginge models
	// and diff them with test intent's intent names
	for _, trained := range models.RecentTrained {
		intentNames, err := getIntentNames(trained.IntentVersion)
		if err != nil {
			return nil, err
		}

		intentDiffs, testIntentDiffs := diffIntentTestIntents(testIntents, intentNames)
		trained.Diffs.Intents = append(trained.Diffs.Intents, intentDiffs...)
		trained.Diffs.TestIntents = append(trained.Diffs.TestIntents, testIntentDiffs...)
		trained.DiffsCount = int64(len(intentDiffs) + len(testIntentDiffs))
	}

	// Recent tested intent tests' Intent Engine models
	queryStr = fmt.Sprintf(`
		SELECT iv.version, tv.ie_model_id, iv.start_train, i.intent_count, iv.sentence_count
		FROM %s AS tv
		INNER JOIN %s AS iv
		ON tv.ie_model_id = iv.ie_model_id
		INNER JOIN (
			SELECT COUNT(1) AS intent_count, version
			FROM %s
			WHERE appid = ?
			GROUP BY version
		) AS i
		ON i.version = iv.version
		WHERE tv.app_id = ? AND iv.result = 1
		ORDER BY tv.id DESC
		LIMIT ?`, IntentTestVersionsTable, IntentVersionsTable, IntentsTable)
	rows, err = tx.Query(queryStr, appID, appID, data.MaxRecentTested)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	for rows.Next() {
		ieModel := data.NewIEModel()
		err = rows.Scan(&ieModel.IntentVersion, &ieModel.ModelID, &ieModel.TrainTime,
			&ieModel.IntentsCount, &ieModel.SentencesCount)
		if err != nil {
			return nil, err
		}
		models.RecentTested = append(models.RecentTested, ieModel)
	}

	// Get intent names of in recent tested Intent Enginge models
	// and diff them with test intent's intent names
	for _, tested := range models.RecentTested {
		intentNames, err := getIntentNames(tested.IntentVersion)
		if err != nil {
			return nil, err
		}

		intentDiffs, testIntentDiffs := diffIntentTestIntents(testIntents, intentNames)
		tested.Diffs.Intents = append(tested.Diffs.Intents, intentDiffs...)
		tested.Diffs.TestIntents = append(tested.Diffs.TestIntents, testIntentDiffs...)
		tested.DiffsCount = int64(len(intentDiffs) + len(testIntentDiffs))
	}

	// Recent saved intent tests' Intent Engine models
	queryStr = fmt.Sprintf(`
		SELECT iv.version, tv.ie_model_id, iv.start_train, i.intent_count, iv.sentence_count
		FROM %s AS tv
		INNER JOIN %s AS iv
		ON tv.ie_model_id = iv.ie_model_id
		INNER JOIN (
			SELECT COUNT(1) AS intent_count, version
			FROM %s
			WHERE appid = ?
			GROUP BY version
		) AS i
		ON i.version = iv.version
		WHERE tv.app_id = ? AND tv.saved = 1 AND iv.result = 1
		ORDER BY tv.id DESC
		LIMIT ?`, IntentTestVersionsTable, IntentVersionsTable, IntentsTable)
	rows, err = tx.Query(queryStr, appID, appID, data.MaxRecentSaved)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	for rows.Next() {
		ieModel := data.NewIEModel()
		err = rows.Scan(&ieModel.IntentVersion, &ieModel.ModelID, &ieModel.TrainTime,
			&ieModel.IntentsCount, &ieModel.SentencesCount)
		if err != nil {
			return nil, err
		}
		models.RecentSaved = append(models.RecentSaved, ieModel)
	}

	// Get intent names of in recent tested and saved Intent Enginge models
	// and diff them with test intent's intent names
	for _, saved := range models.RecentSaved {
		intentNames, err := getIntentNames(saved.IntentVersion)
		if err != nil {
			return nil, err
		}

		intentDiffs, testIntentDiffs := diffIntentTestIntents(testIntents, intentNames)
		saved.Diffs.Intents = append(saved.Diffs.Intents, intentDiffs...)
		saved.Diffs.TestIntents = append(saved.Diffs.TestIntents, testIntentDiffs...)
		saved.DiffsCount = int64(len(intentDiffs) + len(testIntentDiffs))
	}

	return models, nil
}

func (dao IntentTestDao) UpdateTestIntentsProgress(version int64,
	newProgress int) (err error) {
	if dao.db == nil {
	}

	exist, err := dao.checkTestExist(version)
	if err != nil {
		return err
	} else if !exist {
		return data.ErrTestNotFound
	}

	queryStr := fmt.Sprintf(`
		UPDATE %s
		SET progress = progress + ?
		WHERE id = ?`, IntentTestVersionsTable)
	_, err = dao.db.Exec(queryStr, newProgress, version)
	if err != nil {
		return err
	}

	return nil
}

func (dao IntentTestDao) TestIntentsStarted(version int64) (err error) {
	if dao.db == nil {
		return ErrDBNotInit
	}

	exist, err := dao.checkTestExist(version)
	if err != nil {
		return err
	} else if !exist {
		return data.ErrTestNotFound
	}

	queryStr := fmt.Sprintf(`
		UPDATE %s
		SET status = ?
		WHERE id = ?`, IntentTestVersionsTable)
	_, err = dao.db.Exec(queryStr, data.TestStatusTesting, version)
	if err != nil {
		return err
	}

	return nil
}

func (dao IntentTestDao) TestIntentsFinished(appID string, version int64,
	sentences []*data.IntentTestSentence, testResult *data.TestResult) (err error) {
	if dao.db == nil {
		return ErrDBNotInit
	}

	exist, err := dao.checkTestExist(version)
	if err != nil {
		return err
	} else if !exist {
		return data.ErrTestNotFound
	}

	tx, _ := dao.db.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	err = updateIntentSentencesWithTx(tx, sentences)
	if err != nil {
		return err
	}

	now := time.Now().Unix()
	queryStr := fmt.Sprintf(`
		UPDATE %s
		SET status = ?, end_time = ?,
			true_positives = ?, false_positives = ?,
			true_negatives = ?, false_negatives = ?,
			progress = sentences_count
		WHERE id = ?`, IntentTestVersionsTable)
	_, err = tx.Exec(queryStr, data.TestStatusTested, now,
		testResult.TruePositives, testResult.FalsePositives,
		testResult.TrueNegatives, testResult.FalseNegatives, version)
	if err != nil {
		return err
	}

	// Copy current editing test intents and sentences
	queryStr = fmt.Sprintf(`
		SELECT intent_name
		FROM %s
		WHERE app_id = ? AND version IS NULL`, IntentTestIntentsTable)
	rows, err := tx.Query(queryStr, appID)
	if err != nil {
		return err
	}

	copiedTestIntents := make([]*data.IntentTestIntent, 0)
	copiedTestIntentsMap := make(map[string]*data.IntentTestIntent)

	for rows.Next() {
		var intentName sql.NullString
		intentTestIntent := data.IntentTestIntent{
			Version:   &version,
			Sentences: make([]*data.IntentTestSentence, 0),
		}

		err = rows.Scan(&intentName)
		if err != nil {
			return err
		}

		if intentName.Valid {
			intentTestIntent.IntentName = &intentName.String
			copiedTestIntentsMap[intentName.String] = &intentTestIntent
		} else {
			// Negative test intent
			copiedTestIntentsMap[""] = &intentTestIntent
		}

		copiedTestIntents = append(copiedTestIntents, &intentTestIntent)
	}

	for _, sentence := range sentences {
		testIntent, ok := copiedTestIntentsMap[sentence.IntentName]
		if !ok {
			logger.Error.Printf("Cannot find correspond test sentence :%s's test intent %s",
				sentence.Sentence, sentence.IntentName)
			return errors.New("Cannot find correspond test sentence's test intent")
		}

		testIntent.Sentences = append(testIntent.Sentences, sentence)
	}

	err = insertIntentTestsWithTx(tx, appID, copiedTestIntents, &now)
	if err != nil {
		return err
	}

	// Toggle the 'in_used' flag
	return toggleInUsedIntentTestWithTx(tx, appID, version)
}

func (dao IntentTestDao) TestIntentsFailed(version int64,
	testErrMsg string) (err error) {
	if dao.db == nil {
		return ErrDBNotInit
	}

	exist, err := dao.checkTestExist(version)
	if err != nil {
		return err
	} else if !exist {
		return data.ErrTestNotFound
	}

	now := time.Now().Unix()

	queryStr := fmt.Sprintf(`
		UPDATE %s
		SET status = ?, end_time = ?, progress = 0, message = ?
		WHERE id = ?`, IntentTestVersionsTable)
	_, err = dao.db.Exec(queryStr, data.TestStatusFailed, now,
		testErrMsg, version)
	if err != nil {
		return err
	}

	return nil
}

func (dao IntentTestDao) GetIntentTestStatus(appID string) (version int64,
	status int64, sentencesCount int64, progress int64, err error) {
	if dao.db == nil {
		return 0, 0, 0, 0, ErrDBNotInit
	}

	// Get latest test task status and progress, make sure there's no test task in progress
	queryStr := fmt.Sprintf(`
		SELECT id, status, sentences_count, progress
		FROM %s
		WHERE app_id = ?
		ORDER BY id DESC
		LIMIT 1`, IntentTestVersionsTable)
	err = dao.db.QueryRow(queryStr, appID).Scan(&version, &status, &sentencesCount,
		&progress)
	if err != nil && err == sql.ErrNoRows {
		// Never tested before, check if there are editing test intents
		var exist bool
		queryStr = fmt.Sprintf(`
			SELECT 1
			FROM %s
			WHERE app_id = ?`, IntentTestIntentsTable)
		err = dao.db.QueryRow(queryStr, appID).Scan(&exist)
		if err != nil {
			if err == sql.ErrNoRows {
				// No editing test intents, return 'pending' status
				return 0, data.TestStatusPending, 0, 0, nil
			}
			return
		}

		// There are editing test intents, return 'need test' status
		return 0, data.TestStatusNeedTest, 0, 0, nil
	}

	if status == data.TestStatusTesting {
		return
	}

	// Latest test has finished or failed
	// Check if there are editing test intents newer than current in used intent test
	var latestTestIntentUpdatedTime int64
	queryStr = fmt.Sprintf(`
		SELECT MAX(updated_time)
		FROM %s
		WHERE app_id = ? AND version IS NULL`, IntentTestIntentsTable)
	err = dao.db.QueryRow(queryStr, appID).Scan(&latestTestIntentUpdatedTime)
	if err != nil {
		return
	}

	var latestTestTaskEndTime sql.NullInt64
	queryStr = fmt.Sprintf(`
		SELECT id, end_time, status, sentences_count, progress
		FROM %s
		WHERE app_id = ? AND in_used = 1`, IntentTestVersionsTable)
	err = dao.db.QueryRow(queryStr, appID).Scan(&version, &latestTestTaskEndTime,
		&status, &sentencesCount, &progress)
	if err != nil {
		if err == sql.ErrNoRows {
			err = fmt.Errorf("Cannot find in used intent test, but intent tests does exist")
			logger.Error.Println(err.Error())
			return
		}
		return
	}

	if latestTestTaskEndTime.Valid {
		if latestTestIntentUpdatedTime > latestTestTaskEndTime.Int64 {
			// Newer editing test intents exist, return 'need test' status
			return 0, data.TestStatusNeedTest, 0, 0, nil
		}
	} else {
		err = fmt.Errorf("end_time should not be null for in used intent test version: %d, status: %d",
			version, status)
		logger.Error.Println(err.Error())
		return
	}

	// No newer editing test intents exist, return in used test task's
	// status and progress
	return
}

func (dao IntentTestDao) UpdateIntentTest(version int64, name string) error {
	if dao.db == nil {
		return ErrDBNotInit
	}

	exist, err := dao.checkTestExist(version)
	if err != nil {
		return err
	} else if !exist {
		return data.ErrTestNotFound
	}

	queryStr := fmt.Sprintf(`
		UPDATE %s
		SET name = ?
		WHERE id = ?`, IntentTestVersionsTable)
	_, err = dao.db.Exec(queryStr, name, version)
	if err != nil {
		return err
	}

	return nil
}

func (dao IntentTestDao) IntentTestSave(version int64, name string) error {
	if dao.db == nil {
		return ErrDBNotInit
	}

	exist, err := dao.checkTestExist(version)
	if err != nil {
		return err
	} else if !exist {
		return data.ErrTestNotFound
	}

	queryStr := fmt.Sprintf(`
		UPDATE %s
		SET name = ?, saved = 1
		WHERE id = ?`, IntentTestVersionsTable)
	_, err = dao.db.Exec(queryStr, name, version)
	if err != nil {
		return err
	}

	return nil
}

func (dao IntentTestDao) IntentTestUnsave(version int64) error {
	if dao.db == nil {
		return ErrDBNotInit
	}

	exist, err := dao.checkTestExist(version)
	if err != nil {
		return err
	} else if !exist {
		return data.ErrTestNotFound
	}

	queryStr := fmt.Sprintf(`
		UPDATE %s
		SET saved = 0
		WHERE id = ?`, IntentTestVersionsTable)
	_, err = dao.db.Exec(queryStr, version)
	if err != nil {
		return err
	}

	return nil
}

func (dao IntentTestDao) IntentTestImport(appID string,
	testIntents []*data.IntentTestIntent) (err error) {
	if dao.db == nil {
		return ErrDBNotInit
	}

	tx, _ := dao.db.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	err = deleteLatestIntentTestsWithTx(tx, appID)
	if err != nil {
		return err
	}

	now := time.Now().Unix()
	err = insertIntentTestsWithTx(tx, appID, testIntents, &now)
	return
}

func (dao IntentTestDao) IntentTestExport(appID string,
	version *int64) (results []*data.IntentTestSentence, err error) {
	if dao.db == nil {
		return nil, ErrDBNotInit
	}

	if version != nil {
		exist, err := dao.checkTestExist(*version)
		if err != nil {
			return nil, err
		} else if !exist {
			return nil, data.ErrTestNotFound
		}
	}

	results = make([]*data.IntentTestSentence, 0)

	whereClause := "WHERE ti.app_id = ?"
	orderClause := "ORDER BY ti.intent_name DESC"
	queryParams := []interface{}{appID}

	queryStr := fmt.Sprintf(`
		SELECT ts.id, ti.intent_name, ts.sentence
		FROM %s AS ts
		INNER JOIN %s AS ti
		ON ti.id = ts.test_intent`, IntentTestSentencesTable, IntentTestIntentsTable)

	if version != nil {
		whereClause = fmt.Sprintf("%s AND ti.version = ? %s",
			whereClause, orderClause)
		queryParams = append(queryParams, *version)
	} else {
		whereClause = fmt.Sprintf("%s AND ti.version IS NULL %s",
			whereClause, orderClause)
	}

	queryStr = fmt.Sprintf("%s %s", queryStr, whereClause)
	rows, err := dao.db.Query(queryStr, queryParams...)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var intentName sql.NullString
		result := data.IntentTestSentence{}
		err = rows.Scan(&result.ID, &intentName, &result.Sentence)
		if err != nil {
			return nil, err
		}

		if intentName.Valid {
			result.IntentName = intentName.String
		}

		results = append(results, &result)
	}

	return
}

func (dao IntentTestDao) IntentTestRestore(appID string,
	version int64) (err error) {
	if dao.db == nil {
		return ErrDBNotInit
	}

	// 1. Ensure the specified intent test version does exist
	exist, err := dao.checkTestExist(version)
	if err != nil {
		return err
	} else if !exist {
		return data.ErrTestNotFound
	}

	// 2. Ensure the correspond intent version does exist for restore
	exist, intentVersion, err := dao.checkIntentExist(version)
	if err != nil {
		return err
	} else if !exist {
		return data.ErrRestoredIntentNotFound
	}

	tx, _ := dao.db.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	// 3. Delete all NULL version intent test's intents and related test sentences
	//    of given app ID
	err = deleteLatestIntentTestsWithTx(tx, appID)
	if err != nil {
		return err
	}

	// 4. Copy the specified intent test version's intents and related sentences,
	//    then restore them with version set to NULL
	copiedTestIntents, err := copyIntentTestsWithTx(tx, appID, &version)
	if err != nil {
		return err
	}

	err = insertIntentTestsWithTx(tx, appID, copiedTestIntents, nil)
	if err != nil {
		return err
	}

	// 5. Toggle the 'in_used' flag
	err = toggleInUsedIntentTestWithTx(tx, appID, version)
	if err != nil {
		return err
	}

	// 6. Restore intents and their related sentences to the specified version
	return restoreIntentsWithTx(tx, appID, intentVersion)
}

func (dao IntentTestDao) GetIntentIDByName(appID string,
	intentName string) (found bool, intentID int64, err error) {
	if dao.db == nil {
		return false, 0, ErrDBNotInit
	}

	queryStr := fmt.Sprintf(`
		SELECT id
		FROM %s
		WHERE appid = ? AND name = ? AND version IS NULL
		ORDER BY id DESC
		LIMIT 1`, IntentsTable)
	err = dao.db.QueryRow(queryStr, appID, intentName).Scan(&intentID)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, 0, nil
		}
		return false, 0, err
	}

	return true, intentID, nil
}

func (dao IntentTestDao) GetCurrentIEModelID(appID string) (found bool,
	ieModelID string, err error) {
	if dao.db == nil {
		return false, "", ErrDBNotInit
	}

	queryStr := fmt.Sprintf(`
		SELECT ie_model_id
		FROM %s
		WHERE appid = ? AND in_used = 1
		ORDER By version DESC
		LIMIT 1`, IntentVersionsTable)
	err = dao.db.QueryRow(queryStr, appID).Scan(&ieModelID)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, "", nil
		}
		return false, "", err
	}

	return true, ieModelID, nil
}

func (dao IntentTestDao) GetRestoreIEModelID(version int64) (found bool,
	ieModelID string, err error) {
	if dao.db == nil {
		return false, "", ErrDBNotInit
	}

	queryStr := fmt.Sprintf(`
		SELECT ie_model_id
		FROM %s
		WHERE id = ?`, IntentTestVersionsTable)
	err = dao.db.QueryRow(queryStr, version).Scan(&ieModelID)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, "", nil
		}
		return false, "", err
	}

	return true, ieModelID, nil
}

func GetLatestIntentTestIntentsWithTx(tx *sql.Tx,
	appID string) (results []*data.IntentTestIntent, err error) {
	if tx == nil {
		return nil, ErrDBNotInit
	}

	queryStr := fmt.Sprintf(`
		SELECT id, intent_name
		FROM %s
		WHERE app_id = ? AND VERSION IS NULL`, IntentTestIntentsTable)
	rows, err := tx.Query(queryStr, appID)
	if err != nil {
		return nil, err
	}

	results = []*data.IntentTestIntent{}

	for rows.Next() {
		var intentName sql.NullString
		intentTestIntent := data.IntentTestIntent{}
		err = rows.Scan(&intentTestIntent.ID, &intentName)
		if err != nil {
			return nil, err
		}

		if intentName.Valid {
			intentTestIntent.IntentName = &intentName.String
		}

		results = append(results, &intentTestIntent)
	}

	return
}

func deleteLatestIntentTestsWithTx(tx *sql.Tx, appID string) (err error) {
	if tx == nil {
		return ErrDBNotInit
	}

	queryStr := fmt.Sprintf(`
		DELETE
		FROM %s
		WHERE test_intent IN (
			SELECT id
			FROM %s
			WHERE app_id = ? AND version IS NULL
		)`, IntentTestSentencesTable, IntentTestIntentsTable)
	_, err = tx.Exec(queryStr, appID)
	if err != nil {
		return err
	}

	queryStr = fmt.Sprintf(`
		DELETE
		FROM %s
		WHERE app_id = ? AND version IS NULL`, IntentTestIntentsTable)
	_, err = tx.Exec(queryStr, appID)
	if err != nil {
		return err
	}

	return nil
}

func insertIntentTestsWithTx(tx *sql.Tx, appID string,
	testIntents []*data.IntentTestIntent, updatedTime *int64) (err error) {
	if tx == nil {
		return ErrDBNotInit
	}

	sentenceValues := make([]interface{}, 0)

	for _, testIntent := range testIntents {
		var queryStr string
		var queryParams []interface{}
		var _updatedTime int64

		if updatedTime != nil {
			_updatedTime = *updatedTime
		} else {
			_updatedTime = testIntent.UpdatedTime
		}

		if testIntent.Version == nil {
			// Insert into 'intent_test_intents' table
			queryStr = fmt.Sprintf(`
				INSERT INTO %s (app_id, intent_name, version, updated_time)
				VALUES (?, ?, NULL, ?)`, IntentTestIntentsTable)
			queryParams = []interface{}{appID, testIntent.IntentName, _updatedTime}
		} else {
			// Insert into 'intent_test_intents' table
			queryStr = fmt.Sprintf(`
				INSERT INTO %s (app_id, intent_name, version, updated_time)
				VALUES (?, ?, ?, ?)`, IntentTestIntentsTable)
			queryParams = []interface{}{appID, testIntent.IntentName,
				*testIntent.Version, _updatedTime}
		}

		result, err := tx.Exec(queryStr, queryParams...)
		if err != nil {
			return nil
		}

		lastInsertID, err := result.LastInsertId()
		if err != nil {
			return err
		}

		for _, sentence := range testIntent.Sentences {
			sentenceValues = append(sentenceValues, sentence.Sentence,
				lastInsertID, sentence.Result, sentence.Score, sentence.Answer)
		}
	}

	// Insert into 'intent_test_sentences' table
	sentencesPerOp := 2000
	start := 0

	if len(sentenceValues) > 0 {
		for {
			end := start + sentencesPerOp*5
			if end > len(sentenceValues) {
				end = len(sentenceValues)
			}

			params := sentenceValues[start:end]

			queryStr := fmt.Sprintf(`
				INSERT INTO %s (sentence, test_intent, result, score, answer)
				VALUES (?, ?, ?, ?, ?)%s`, IntentTestSentencesTable,
				strings.Repeat(", (?, ?, ?, ?, ?)", len(params)/5-1))
			_, err := tx.Exec(queryStr, params...)
			if err != nil {
				return err
			}

			if end == len(sentenceValues) {
				return nil
			}
			start = end
		}
	}

	return nil
}

func copyIntentTestsWithTx(tx *sql.Tx, appID string,
	version *int64) (copiedTestIntents []*data.IntentTestIntent, err error) {
	copiedTestIntents = make([]*data.IntentTestIntent, 0)
	if tx == nil {
		return nil, ErrDBNotInit
	}

	whereClause := "WHERE ti.app_id = ?"
	queryParams := []interface{}{appID}
	if version == nil {
		whereClause = fmt.Sprintf("%s AND ti.version IS NULL", whereClause)
	} else {
		whereClause = fmt.Sprintf("%s AND ti.version = ?", whereClause)
		queryParams = append(queryParams, *version)
	}

	// Note:
	// 	'ti.intent', 'ts.sentence' and 'ts.result' may be NULL after left outer join
	queryStr := fmt.Sprintf(`
		SELECT ti.intent_name, ti.updated_time, ts.sentence, ts.result, ts.score, ts.answer
		FROM %s AS ti
		LEFT JOIN %s AS ts
		ON ts.test_intent = ti.id
		%s`, IntentTestIntentsTable, IntentTestSentencesTable, whereClause)
	rows, err := tx.Query(queryStr, queryParams...)
	if err != nil {
		return nil, err
	}

	intentsMap := make(map[string]*data.IntentTestIntent)

	for rows.Next() {
		var result, score sql.NullInt64
		var intentName sql.NullString
		var sentence, answer sql.NullString
		var updatedTime int64
		err = rows.Scan(&intentName, &updatedTime, &sentence, &result, &score, &answer)
		if err != nil {
			return nil, err
		}

		// Note: Negative test intent's name will be empty string key
		var name string
		if intentName.Valid {
			name = intentName.String
		}

		intent, ok := intentsMap[name]
		if !ok {
			if name != "" {
				intent = &data.IntentTestIntent{
					IntentName:  &name,
					UpdatedTime: updatedTime,
					Sentences:   make([]*data.IntentTestSentence, 0),
				}
			} else {
				// Negative test intent
				intent = &data.IntentTestIntent{
					UpdatedTime: updatedTime,
					Sentences:   make([]*data.IntentTestSentence, 0),
				}
			}

			copiedTestIntents = append(copiedTestIntents, intent)
			intentsMap[name] = intent
		}

		if sentence.Valid {
			s := data.IntentTestSentence{
				Sentence: sentence.String,
			}

			if result.Valid {
				s.Result = result.Int64
			}

			if score.Valid {
				s.Score = &score.Int64
			}

			if answer.Valid {
				s.Answer = &answer.String
			}

			intent.Sentences = append(intent.Sentences, &s)
		}
	}

	return
}

func updateIntentSentencesWithTx(tx *sql.Tx,
	sentences []*data.IntentTestSentence) (err error) {
	if tx == nil {
		return ErrDBNotInit
	}

	sentenceValues := make([]interface{}, len(sentences)*6)

	for i := 0; i < len(sentences); i++ {
		offset := i * 2
		scoreOffset := len(sentences)*2 + offset
		answerOffset := len(sentences)*4 + offset
		sentenceValues[offset] = sentences[i].ID
		sentenceValues[offset+1] = sentences[i].Result
		sentenceValues[scoreOffset] = sentences[i].ID
		sentenceValues[scoreOffset+1] = sentences[i].Score
		sentenceValues[answerOffset] = sentences[i].ID
		sentenceValues[answerOffset+1] = sentences[i].Answer
	}

	// Update 'intent_test_sentences' table
	sentencesPerOp := 2000
	start := 0

	if len(sentenceValues) > 0 {
		whenCases := make([]string, len(sentenceValues)/6)
		for i := 0; i < len(sentenceValues)/6; i++ {
			whenCases[i] = "WHEN ? THEN ?"
		}
		whenCasesStr := strings.Join(whenCases, " ")

		for {
			end := start + sentencesPerOp*6
			if end > len(sentenceValues) {
				end = len(sentenceValues)
			}

			params := sentenceValues[start:end]
			queryStr := fmt.Sprintf(`
				UPDATE %s
				SET result =
						CASE id
							%s
							ELSE result
						END,
					score =
						CASE id
							%s
							ELSE score
						END,
					answer =
						CASE id
							%s
							ELSE answer
						END`, IntentTestSentencesTable,
				whenCasesStr, whenCasesStr, whenCasesStr)
			_, err = tx.Exec(queryStr, params...)
			if err != nil {
				return err
			}

			if end == len(sentenceValues) {
				return nil
			}
			start = end
		}
	}

	return nil
}

func restoreIntentsWithTx(tx *sql.Tx, appID string, intentVersion int64) (err error) {
	if tx == nil {
		return ErrDBNotInit
	}

	queryStr := fmt.Sprintf(`
		DELETE
		FROM %s
		WHERE intent IN (
			SELECT id
			FROM %s
			WHERE appid = ? AND version IS NULL
		)`, IntentTrainSetsTable, IntentsTable)
	_, err = tx.Exec(queryStr, appID)
	if err != nil {
		return err
	}

	queryStr = fmt.Sprintf(`
		DELETE
		FROM %s
		WHERE appid = ? AND version IS NULL`, IntentsTable)
	_, err = tx.Exec(queryStr, appID)
	if err != nil {
		return err
	}

	queryStr = fmt.Sprintf(`
		SELECT i.name, i.updatetime, it.sentence
		FROM %s AS i
		INNER JOIN %s AS it
		ON it.intent = i.id
		WHERE i.version = ?`, IntentsTable, IntentTrainSetsTable)
	rows, err := tx.Query(queryStr, intentVersion)
	if err != nil {
		return err
	}

	type Intent struct {
		id          int64
		updatedTime int64
	}

	// map[intentName]newIntentID
	intents := make(map[string]*Intent)
	// map[intentTrainSetSentence]intentName
	trainSets := make(map[string]string)

	for rows.Next() {
		var intentName, sentence string
		var updatedTime int64
		err = rows.Scan(&intentName, &updatedTime, &sentence)
		if err != nil {
			return err
		}

		if _, ok := intents[intentName]; !ok {
			intents[intentName] = &Intent{
				updatedTime: updatedTime,
			}
		}

		trainSets[sentence] = intentName
	}

	queryStr = fmt.Sprintf(`
		INSERT INTO %s (appid, name, version, updatetime)
		VALUES (?, ?, NULL, ?)`, IntentsTable)

	for intentName, intent := range intents {
		result, err := tx.Exec(queryStr, appID, intentName, intent.updatedTime)
		if err != nil {
			return err
		}

		id, err := result.LastInsertId()
		if err != nil {
			return err
		}

		intents[intentName].id = id
	}

	sentencesPerOp := 2000

	if len(trainSets) > 0 {
		params := make([]interface{}, 0)
		ops := 0

		for sentence, intentName := range trainSets {
			queryStr = fmt.Sprintf(`
				INSERT INTO %s (sentence, intent)
				VALUES (?, ?)`, IntentTrainSetsTable)
			intentID := intents[intentName].id
			params = append(params, sentence, intentID)
			ops++

			if ops == sentencesPerOp || ops == len(trainSets) {
				queryStr = fmt.Sprintf("%s%s", queryStr,
					strings.Repeat(", (?, ?)", len(params)/2-1))
				_, err := tx.Exec(queryStr, params...)
				if err != nil {
					return err
				}

				params = make([]interface{}, 0)
				ops = 0
			}
		}
	}

	// Toggle the 'in_used' flag
	return toggleInUsedIEModelWithTx(tx, appID, intentVersion)
}

func toggleInUsedIntentTestWithTx(tx *sql.Tx, appID string,
	version int64) (err error) {
	if tx == nil {
		return ErrDBNotInit
	}

	queryStr := fmt.Sprintf(`
		UPDATE %s
		SET in_used = 0
		WHERE app_id = ? AND in_used = 1`, IntentTestVersionsTable)
	_, err = tx.Exec(queryStr, appID)
	if err != nil {
		return err
	}

	queryStr = fmt.Sprintf(`
		UPDATE %s
		SET in_used = 1
		WHERE app_id = ? AND id = ?`, IntentTestVersionsTable)
	_, err = tx.Exec(queryStr, appID, version)
	return err
}

func toggleInUsedIEModelWithTx(tx *sql.Tx, appID string,
	intentVersion int64) (err error) {
	if tx == nil {
		return ErrDBNotInit
	}

	queryStr := fmt.Sprintf(`
		UPDATE %s
		SET in_used = 0
		WHERE appid = ? AND in_used = 1`, IntentVersionsTable)
	_, err = tx.Exec(queryStr, appID)
	if err != nil {
		return err
	}

	queryStr = fmt.Sprintf(`
		UPDATE %s
		SET in_used = 1
		WHERE appid = ? AND version = ?`, IntentVersionsTable)
	_, err = tx.Exec(queryStr, appID, intentVersion)
	return
}

func getLatestIntentNamesWithTx(tx *sql.Tx,
	appID string) (intentNames []string, err error) {
	if tx == nil {
		return nil, ErrDBNotInit
	}

	intentNames = []string{}

	queryStr := fmt.Sprintf(`
		SELECT name
		FROM %s
		WHERE appid = ? AND version IS NULL`, IntentsTable)
	rows, err := tx.Query(queryStr, appID)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	for rows.Next() {
		var intentName string
		err = rows.Scan(&intentName)
		if err != nil {
			return nil, err
		}

		intentNames = append(intentNames, intentName)
	}

	return intentNames, nil
}

func diffIntentTestIntents(testIntents []*data.IntentTestIntent,
	intents []string) (intentDiffs, testIntentDiffs []string) {
	testIntentsMap := make(map[string]bool)
	intentsMap := make(map[string]bool)

	intentDiffs = []string{}
	testIntentDiffs = []string{}

	for _, testIntent := range testIntents {
		// Ignore negative test intent
		if testIntent.IntentName != nil {
			testIntentsMap[*testIntent.IntentName] = true
		}
	}

	for _, intent := range intents {
		intentsMap[intent] = true

		if _, ok := testIntentsMap[intent]; !ok {
			intentDiffs = append(intentDiffs, intent)
		}
	}

	for _, testIntent := range testIntents {
		// Ignore negative test intent
		if testIntent.IntentName != nil {
			testIntentName := *testIntent.IntentName
			if _, ok := intentsMap[testIntentName]; !ok {
				testIntentDiffs = append(testIntentDiffs, testIntentName)
			}
		}
	}

	return
}

func (dao IntentTestDao) checkTestExist(version int64) (exist bool, err error) {
	if dao.db == nil {
		return false, ErrDBNotInit
	}

	queryStr := fmt.Sprintf(`
		SELECT 1
		FROM %s
		WHERE id = ?`, IntentTestVersionsTable)
	err = dao.db.QueryRow(queryStr, version).Scan(&exist)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (dao IntentTestDao) checkTestIntentExist(testIntentID int64) (exist bool,
	err error) {
	if dao.db == nil {
		return false, ErrDBNotInit
	}

	queryStr := fmt.Sprintf(`
		SELECT 1
		FROM %s
		WHERE id = ?`, IntentTestIntentsTable)
	err = dao.db.QueryRow(queryStr, testIntentID).Scan(&exist)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (dao IntentTestDao) checkIntentExist(version int64) (exist bool,
	intentVersion int64, err error) {
	if dao.db == nil {
		return false, 0, ErrDBNotInit
	}

	queryStr := fmt.Sprintf(`
		SELECT version
		FROM %s AS iv
		INNER JOIN %s AS tv
		ON tv.ie_model_id = iv.ie_model_id
		WHERE tv.id = ?`, IntentVersionsTable, IntentTestVersionsTable)
	err = dao.db.QueryRow(queryStr, version).Scan(&intentVersion)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, 0, nil
		}
		return false, 0, err
	}

	return true, intentVersion, nil
}

func getLikeValue(val string) string {
	return fmt.Sprintf("%%%s%%", strings.Replace(val, "%", "\\%", -1))
}
