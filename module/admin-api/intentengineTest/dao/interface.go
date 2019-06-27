package dao

import (
	"database/sql"
	"errors"

	"emotibot.com/emotigo/module/admin-api/intentengineTest/data"
)

var (
	// ErrDBNotInit is used to be returned if dao is not initialized properly
	ErrDBNotInit = errors.New("DB is not init")
)

type IntentTestDao struct {
	db *sql.DB
}

func NewIntentTestDao(db *sql.DB) IntentTestDao {
	return IntentTestDao{
		db: db,
	}
}

type Dao interface {
	// GetIntentTests will return the intent tests information
	// including both 'MaxLatestTests' unsaved intent tests and saved intent tests
	GetIntentTests(appID string) (results *data.IntentTestResults, err error)

	// GetIntentTest will return the given version of intent test
	// It will return error if the specified version of intent test does not exist
	GetIntentTest(appID string, version int64, keyword string,
		locale string) (result *data.IntentTest, err error)

	// GetLatestIntents will return the latest editing intents
	// (which version is NULL) of intent tests
	GetLatestIntents(appID string, keyword string,
		locale string) (results []*data.IntentTestIntent, err error)

	// GetIntentTestIntentSentences will return the test sentences of the given
	// test intent
	// It will return error if the specified version of test intent does not exist
	GetIntentTestIntentSentences(testIntentID int64,
		keyword string) (results []*data.IntentTestSentence, err error)

	// GetLatestIntentTestSentences will return the latest editing intent test's
	// sentences (which version is NULL)
	GetLatestIntentTestSentences(appID string) (results []*data.IntentTestSentence,
		err error)

	// ModifyIntentTestSentences will update the test sentences of the given
	// test intent
	// In updateSentences, id = 0 means creating new test sentence
	// It will return error if the specified version of intent test does not exist
	ModifyIntentTestSentences(testIntentID int64, updateList []*data.UpdateCmd,
		deleteList []int64) (err error)

	// TestIntents will create a new intent test task
	// It will also ensure currently there no running test task exist before
	// creating a new version of intent test
	TestIntents(appID string, userID string, ieModelID string) (version int64,
		status int64, err error)

	// UpdateTestIntentsProgress will update the running progress of the
	// given intent test task
	// It will return error if the specified version of intent test does not exist
	UpdateTestIntentsProgress(version int64, newProgress int) (err error)

	// TestIntentsStarted will mark the given intent test task as testing
	// It will return error if the specified version of intent test does not exist
	TestIntentsStarted(version int64) (err error)

	// TestIntentsFinished will mark the given intent test task as finished
	// with test results
	// It will return error if the specified version of intent test does not exist
	TestIntentsFinished(appID string, version int64,
		sentences []*data.IntentTestSentence, testResult *data.TestResult) (err error)

	// TestIntentsFailed will mark the given intent test task as failed
	// with error message
	// It will return error if the specified version of intent test does not exist
	TestIntentsFailed(version int64, testErrMsg string) (err error)

	// GetIntentTestStatus will return latest test task status and progress
	// of specified robot
	// It will return 'PENDING' status with 0 progress if the specified robot
	// has never tested or edited tests before
	GetIntentTestStatus(appID string) (version int64, status int64,
		sentencesCount int64, progress int64, err error)

	// UpdateIntentTest will update the given test intent information, ex: name... etc
	UpdateIntentTest(version int64, name string) error

	// GetUsableModels will return the list of usable Intent Engine models
	// and thier corresponding information
	GetUsableModels(appID string) (models *data.UseableModels, err error)

	// IntentTestSave will mark the given test intent as saved
	IntentTestSave(version int64, name string) error

	// IntentTestUnsave will mark the given test intent as unsaved
	IntentTestUnsave(version int64) error

	// IntentTestImport will replace latest editing test intents with uploaded file
	IntentTestImport(appID string,
		testIntents []*data.IntentTestIntent) (err error)

	// IntentTestExport will export all test intents and their corresponded
	// test sentences
	// If version is not null, the specified test intents are exported
	// If version is null, latest editing test intents are exported
	IntentTestExport(appID string,
		version *int64) (results []*data.IntentTestSentence, err error)

	// IntentTestRestore will:
	// 	1. Ensure the specified intent test version does exist
	// 	2. Ensure the correspond intent version does exist for restore
	// 	3. Delete all NULL version intent test's intents and related test sentences
	//	   of given app ID
	// 	4. Copy the specified intent test version's intents and related sentences
	// 	   with version set to NULL
	// 	5. Restore intents to the specified version
	IntentTestRestore(appID string, version int64) (err error)

	// GetIntentIDByName will return the intent ID of given intent name
	// 'found' will be false if correspond intent cannot be found
	GetIntentIDByName(appID string, intentName string) (found bool, intentID int64,
		err error)

	// GetCurrentIEModelID will return the Intent Engine model ID currently in used
	// of specified robot
	// 'found' will be false if currently in used Intent Engine model ID
	// cannot be found
	GetCurrentIEModelID(appID string) (found bool, ieModelID string, err error)

	// GetRestoreIEModelID will return the Intent Engine model ID to be restored
	// of the specified intent test version
	// 'found' will be false if the specified intent test version cannot be found
	GetRestoreIEModelID(version int64) (found bool, ieModelID string, err error)

	// GetLatestIntentNames return the names of latest intents
	// (which version is NULL)
	GetLatestIntentNames(appID string) (intentNames []string, err error)
}
