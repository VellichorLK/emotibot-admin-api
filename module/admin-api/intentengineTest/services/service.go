package services

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"emotibot.com/emotigo/module/admin-api/util/zhconverter"

	"emotibot.com/emotigo/module/admin-api/intentengineTest/dao"
	"emotibot.com/emotigo/module/admin-api/intentengineTest/data"
	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/admin-api/util/AdminErrors"
	"emotibot.com/emotigo/module/admin-api/util/localemsg"
	"emotibot.com/emotigo/pkg/logger"

	"github.com/tealeg/xlsx"
)

const defaultIntentEngineURL = "http://127.0.0.1:15001"
const maxRetries = 3
const maxPredictThreads = 10

var intentTestDao dao.Dao

func InitDao() {
	db := util.GetMainDB()
	intentTestDao = dao.NewIntentTestDao(db)
}

func GetIntentTests(appID string) (*data.IntentTestResults, AdminErrors.AdminError) {
	results, err := intentTestDao.GetIntentTests(appID)
	if err != nil {
		return results, AdminErrors.New(AdminErrors.ErrnoDBError, err.Error())
	}
	return results, nil
}

func GetIntentTestStatus(appID string) (int64, int64, int64,
	int64, AdminErrors.AdminError) {
	version, status, sentencesCount, progress,
		err := intentTestDao.GetIntentTestStatus(appID)
	if err != nil {
		return 0, 0, 0, 0, AdminErrors.New(AdminErrors.ErrnoDBError, err.Error())
	}
	return version, status, sentencesCount, progress, nil
}

func GetIntentTest(appID string, version int64, keyword string,
	locale string) (*data.IntentTest,
	AdminErrors.AdminError) {
	result, err := intentTestDao.GetIntentTest(appID, version, keyword, locale)
	if err != nil {
		if err == data.ErrTestNotFound {
			return nil, AdminErrors.New(AdminErrors.ErrnoNotFound,
				localemsg.Get(locale, "TestNotFoundError"))
		}
		return nil, AdminErrors.New(AdminErrors.ErrnoDBError, err.Error())
	}
	return result, nil
}

func PatchIntentTest(version int64, name string, locale string) AdminErrors.AdminError {
	err := intentTestDao.UpdateIntentTest(version, name)
	if err != nil {
		if err == data.ErrTestNotFound {
			return AdminErrors.New(AdminErrors.ErrnoNotFound,
				localemsg.Get(locale, "TestNotFoundError"))
		}
		return AdminErrors.New(AdminErrors.ErrnoDBError, err.Error())
	}
	return nil
}

func SaveIntentTest(version int64, name string, locale string) AdminErrors.AdminError {
	err := intentTestDao.IntentTestSave(version, name)
	if err != nil {
		if err == data.ErrTestNotFound {
			return AdminErrors.New(AdminErrors.ErrnoNotFound,
				localemsg.Get(locale, "TestNotFoundError"))
		}
		return AdminErrors.New(AdminErrors.ErrnoDBError, err.Error())
	}
	return nil
}

func UnsaveIntentTest(version int64, locale string) AdminErrors.AdminError {
	err := intentTestDao.IntentTestUnsave(version)
	if err != nil {
		if err == data.ErrTestNotFound {
			return AdminErrors.New(AdminErrors.ErrnoNotFound,
				localemsg.Get(locale, "TestNotFoundError"))
		}
		return AdminErrors.New(AdminErrors.ErrnoDBError, err.Error())
	}
	return nil
}

func ExportIntentTest(appID string, version *int64,
	locale string) ([]byte, AdminErrors.AdminError) {
	sentences, err := intentTestDao.IntentTestExport(appID, version)
	if err != nil {
		if err == data.ErrTestNotFound {
			return nil, AdminErrors.New(AdminErrors.ErrnoNotFound,
				localemsg.Get(locale, "TestNotFoundError"))
		}
		return nil, AdminErrors.New(AdminErrors.ErrnoDBError, err.Error())
	}

	return createExportIntentTestXlsx(sentences, locale)
}

func createExportIntentTestXlsx(sentences []*data.IntentTestSentence, locale string) ([]byte,
	AdminErrors.AdminError) {
	file := xlsx.NewFile()
	sheet, err := file.AddSheet(localemsg.Get(locale, "IntentTestSentences"))
	if err != nil {
		return nil, AdminErrors.New(AdminErrors.ErrnoIOError, err.Error())
	}

	headerRow := sheet.AddRow()
	headerRow.AddCell().SetString(localemsg.Get(locale, "IntentName"))
	headerRow.AddCell().SetString(localemsg.Get(locale, "TestSentence"))

	for _, sentence := range sentences {
		row := sheet.AddRow()
		row.AddCell().SetString(sentence.IntentName)
		row.AddCell().SetString(sentence.Sentence)
	}

	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)
	ioErr := file.Write(writer)
	if ioErr != nil {
		return nil, AdminErrors.New(AdminErrors.ErrnoIOError, ioErr.Error())
	}

	return buf.Bytes(), nil
}

func RestoreIntentTest(appID string, version int64,
	locale string) AdminErrors.AdminError {
	currentFound, currentIEModelID, err := intentTestDao.GetCurrentIEModelID(appID)
	if err != nil {
		return AdminErrors.New(AdminErrors.ErrnoDBError, err.Error())
	}

	restoreFound, restoreIEModelID, err := intentTestDao.GetRestoreIEModelID(version)
	if err != nil {
		return AdminErrors.New(AdminErrors.ErrnoDBError, err.Error())
	} else if !restoreFound {
		if err == data.ErrTestNotFound {
			return AdminErrors.New(AdminErrors.ErrnoNotFound,
				localemsg.Get(locale, "TestNotFoundError"))
		}
	}

	// Restore Intent Engine model
	apiErr := loadIEModel(appID, restoreIEModelID)
	if apiErr != nil {
		return AdminErrors.New(AdminErrors.ErrnoAPIError, apiErr.Error())
	}

	err = intentTestDao.IntentTestRestore(appID, version)
	if err != nil {
		// Restore failed, rollback to previous Intent Engine model
		if currentFound {
			apiErr = loadIEModel(appID, currentIEModelID)
			if apiErr != nil {
				return AdminErrors.New(AdminErrors.ErrnoAPIError, apiErr.Error())
			}
		}

		if err == data.ErrTestNotFound {
			return AdminErrors.New(AdminErrors.ErrnoNotFound,
				localemsg.Get(locale, "TestNotFoundError"))
		}
		return AdminErrors.New(AdminErrors.ErrnoDBError, err.Error())
	}

	return nil
}

func GetLatestIntents(appID string, keyword string,
	locale string) ([]*data.IntentTestIntent, AdminErrors.AdminError) {
	results, err := intentTestDao.GetLatestIntents(appID, keyword, locale)
	if err != nil {
		return nil, AdminErrors.New(AdminErrors.ErrnoDBError, err.Error())
	}
	return results, nil
}

func ImportLatestIntentTest(appID string, buf []byte,
	locale string) AdminErrors.AdminError {
	testIntents, negativeTestIntent, err := parseImportTestFile(appID, buf, locale)
	if err != nil {
		switch err {
		case data.ErrTestImportSheetFormat:
			return AdminErrors.New(AdminErrors.ErrnoRequestError,
				localemsg.Get(locale, "IntentTestUploadSheetFormatErr"))
		case data.ErrTestImportSheetNoHeader:
			return AdminErrors.New(AdminErrors.ErrnoRequestError,
				localemsg.Get(locale, "IntentTestUploadNoHeaderTpl"))
		case data.ErrTestImportSheetEmptySentence:
			return AdminErrors.New(AdminErrors.ErrnoRequestError,
				localemsg.Get(locale, "IntestTestSentenceEmptyErr"))
		default:
			return AdminErrors.New(AdminErrors.ErrnoDBError, err.Error())
		}
	}

	latestIntents, err := intentTestDao.GetLatestIntentNames(appID)
	if err != nil {
		return AdminErrors.New(AdminErrors.ErrnoDBError, err.Error())
	}

	intents := map[string]bool{}

	for _, intentName := range latestIntents {
		intents[intentName] = false
	}

	_testIntents := []*data.IntentTestIntent{}

	// Ensure the imported test intents exists in current intents.
	for _, testIntent := range testIntents {
		if testIntent.IntentName == nil {
			// Safety check, should not happen.
			continue
		}

		if _, ok := intents[*testIntent.IntentName]; !ok {
			// Imported test intent does not exist in current intents,
			// move test sentences under negative test intent.
			negativeTestIntent.Sentences = append(negativeTestIntent.Sentences,
				testIntent.Sentences...)
		} else {
			_testIntents = append(_testIntents, testIntent)
			intents[*testIntent.IntentName] = true
		}
	}

	// Ensure all test intents equals to current intents, even imported test intents may not include them.
	for intentName, ok := range intents {
		if !ok {
			// Make a copy of 'intentName'.
			// (We cannot use the address of 'intentName' directly,
			// 	for range reuses the same variable for different values during iteration.)
			_intentName := intentName
			_testIntents = append(_testIntents, &data.IntentTestIntent{
				IntentName: &_intentName,
			})
		}
	}

	err = intentTestDao.IntentTestImport(appID, append(_testIntents, negativeTestIntent))
	if err != nil {
		return AdminErrors.New(AdminErrors.ErrnoDBError, err.Error())
	}

	return nil
}

func TestIntents(appID string, userID string, ieModelID string,
	locale string) AdminErrors.AdminError {
	version, status, err := intentTestDao.TestIntents(appID, userID, ieModelID)
	if err != nil {
		if err == data.ErrTestTaskInProcess {
			return AdminErrors.New(AdminErrors.ErrnoRequestError,
				localemsg.Get(locale, "PreviousTestStillRunning"))
		}
		return AdminErrors.New(AdminErrors.ErrnoDBError, err.Error())
	}

	if status == data.TestStatusPending {
		// Test task created (pending), run intents test
		go runIntentsTest(appID, version, ieModelID)
	}

	return nil
}

func GetUsableModels(appID string) (*data.UseableModels, AdminErrors.AdminError) {
	models, err := intentTestDao.GetUsableModels(appID)
	if err != nil {
		return nil, AdminErrors.New(AdminErrors.ErrnoDBError, err.Error())
	}
	return models, nil
}

func GetIntent(testIntentID int64, keyword string,
	locale string) ([]*data.IntentTestSentence, AdminErrors.AdminError) {
	results, err := intentTestDao.GetIntentTestIntentSentences(testIntentID, keyword)
	if err != nil {
		if err == data.ErrTestIntentNotFound {
			return nil, AdminErrors.New(AdminErrors.ErrnoRequestError,
				localemsg.Get(locale, "TestIntentNotFoundError"))
		}
		return nil, AdminErrors.New(AdminErrors.ErrnoDBError, err.Error())
	}
	return results, nil
}

func UpdateIntent(testIntentID int64, updateList []*data.UpdateCmd,
	deleteList []int64, locale string) AdminErrors.AdminError {
	err := intentTestDao.ModifyIntentTestSentences(testIntentID, updateList,
		deleteList)
	if err != nil {
		if err == data.ErrTestIntentNotFound {
			return AdminErrors.New(AdminErrors.ErrnoRequestError,
				localemsg.Get(locale, "TestIntentNotFoundError"))
		}
		return AdminErrors.New(AdminErrors.ErrnoDBError, err.Error())
	}
	return nil
}

func parseImportTestFile(appID string,
	buf []byte, locale string) (testIntents []*data.IntentTestIntent,
	negativeTestIntent *data.IntentTestIntent, err error) {
	file, err := xlsx.OpenBinary(buf)
	if err != nil {
		return nil, nil, err
	}

	testIntents = []*data.IntentTestIntent{}
	negativeTestIntent = &data.IntentTestIntent{
		Sentences: make([]*data.IntentTestSentence, 0),
	}
	testIntentsMap := make(map[string]*data.IntentTestIntent)
	sheets := file.Sheets

	if len(sheets) != 1 {
		return nil, nil, data.ErrTestImportSheetFormat
	}

	if len(sheets[0].Rows) == 0 {
		return nil, nil, data.ErrTestImportSheetNoHeader
	}

	headerRow := sheets[0].Rows[0]

	if len(headerRow.Cells) < 2 {
		return nil, nil, data.ErrTestImportSheetNoHeader
	}

	intentNameHeader := headerRow.Cells[0].String()
	sentenceHeader := headerRow.Cells[1].String()

	if intentNameHeader != localemsg.Get(locale, "IntentName") &&
		intentNameHeader != localemsg.Get("", "IntentName") {
		return nil, nil, data.ErrTestImportSheetNoHeader
	}

	if sentenceHeader != localemsg.Get(locale, "TestSentence") &&
		sentenceHeader != localemsg.Get("", "TestSentence") {
		return nil, nil, data.ErrTestImportSheetNoHeader
	}

	// Skip header rows
	rows := sheets[0].Rows[1:]

	for idx := range rows {
		cells := rows[idx].Cells
		intentName := strings.TrimSpace(cells[0].String())
		sentence := strings.TrimSpace(cells[1].String())

		if sentence == "" {
			return nil, nil, data.ErrTestImportSheetEmptySentence
		}

		testSentence := data.IntentTestSentence{
			Sentence: sentence,
		}

		if intentName != "" {
			testIntent, ok := testIntentsMap[intentName]
			if !ok {
				testIntent = &data.IntentTestIntent{
					IntentName: &intentName,
					Sentences:  make([]*data.IntentTestSentence, 0),
				}
				testIntents = append(testIntents, testIntent)
				testIntentsMap[intentName] = testIntent
			}

			testIntent.Sentences = append(testIntent.Sentences, &testSentence)
		} else {
			// Negative test intent
			negativeTestIntent.Sentences = append(negativeTestIntent.Sentences, &testSentence)
		}
	}

	return testIntents, negativeTestIntent, nil
}

func runIntentsTest(appID string, version int64, ieModelID string) {
	var err, loadModelErr, predictErr error
	var sentences []*data.IntentTestSentence
	overallTestResult := &data.TestResult{}

	defer func() {
		if loadModelErr != nil {
			intentTestDao.TestIntentsFailed(version, loadModelErr.Error())
			logger.Error.Printf("Intent test task failed: %s.", loadModelErr.Error())
		} else if predictErr != nil {
			intentTestDao.TestIntentsFailed(version, predictErr.Error())
			logger.Error.Printf("Intent test task failed: %s.", predictErr.Error())
		} else if err != nil {
			intentTestDao.TestIntentsFailed(version, err.Error())
			logger.Error.Printf("Intent test task failed: %s.", err.Error())
		} else {
			err = intentTestDao.TestIntentsFinished(appID, version, sentences,
				overallTestResult)
			logger.Info.Println("Intent test task finished.")
			if err != nil {
				intentTestDao.TestIntentsFailed(version, err.Error())
				logger.Error.Printf("Intent test task failed: %s.\n", err.Error())
			}
		}
	}()

	// Change test task status to 'Testing'
	err = intentTestDao.TestIntentsStarted(version)
	if err != nil {
		return
	}
	logger.Info.Println("Intent test task started.")

	// Load test intents
	sentences, err = intentTestDao.GetLatestIntentTestSentences(appID)
	if err != nil {
		return
	}

	if len(sentences) == 0 {
		logger.Info.Println("No intent sentence to tested.")
		return
	}

	logger.Info.Printf("Total %d intent sentences to be tested.", len(sentences))

	// Create a fake app ID for training
	now := time.Now()
	fakeAppID := fmt.Sprintf("%s_%s", appID, now.Format("20060102_150405"))

	// Load model with fake app ID
	logger.Info.Printf("Load Intent Engine model: %s with fake app ID: %s.\n",
		ieModelID, fakeAppID)
	err = loadIEModel(fakeAppID, ieModelID)
	if err != nil {
		return
	}

	defer func() {
		// Unload the model of fake app ID
		logger.Info.Printf("Unload Intent Engine model with fake app ID: %s.\n", fakeAppID)
		err = unloadIEModel(fakeAppID)
	}()

	// Ensure intent model is loaded
	loadRetry := 0
	payload := map[string]interface{}{
		"app_id":   fakeAppID,
		"sentence": "测试语句",
	}

	logger.Info.Println("Waiting for Intent Engine model loaded...")

	for {
		time.Sleep(10 * time.Second)
		loadRetry++
		if loadRetry > maxRetries {
			loadModelErr = fmt.Errorf("Intent Engine model not loaded, exceeded max retries: %d", maxRetries)
			logger.Error.Printf(loadModelErr.Error())
			return
		}

		logger.Info.Println("Checking Intent Engine model load status...")
		loaded, loadModelErr := ieModelLoaded(payload)
		if loadModelErr != nil {
			logger.Error.Printf(loadModelErr.Error())
			return
		}

		if loaded {
			logger.Info.Println("Intent Engine model is loaded")
			break
		}
		logger.Info.Println("Intent Engine model not loaded")
	}

	// Start predicting
	var numOfPredictThreads int

	chunkSize := len(sentences) / maxPredictThreads
	if chunkSize == 0 {
		numOfPredictThreads = 1
	} else {
		numOfPredictThreads = maxPredictThreads
	}

	resultsChan := make(chan *data.TestResult, numOfPredictThreads)

	for i := 0; i < numOfPredictThreads; i++ {
		start := i * chunkSize
		var end int

		if i < numOfPredictThreads-1 {
			end = start + chunkSize
		} else {
			end = start + len(sentences[start:])
		}

		// Predict sentence
		go predictSentences(version, fakeAppID, sentences[start:end], resultsChan)
	}

	// Wait for all predictions complete
	for i := 0; i < numOfPredictThreads; i++ {
		testResult := <-resultsChan

		overallTestResult.TruePositives += testResult.TruePositives
		overallTestResult.FalsePositives += testResult.FalsePositives
		overallTestResult.TrueNegatives += testResult.TrueNegatives
		overallTestResult.FalseNegatives += testResult.FalseNegatives

		if testResult.Error != nil {
			predictErr = testResult.Error
		}
	}

	return
}

func loadIEModel(appID string, ieModelID string) (err error) {
	intentEngineURL := getEnvironment("INTENT_ENGINE_URL")
	if intentEngineURL == "" {
		intentEngineURL = defaultIntentEngineURL
	}

	ieLoadModelURL := fmt.Sprintf("%s/%s", intentEngineURL, "load_model")
	payload := map[string]interface{}{
		"app_id":   appID,
		"model_id": ieModelID,
	}

	body, _err := util.HTTPPostJSON(ieLoadModelURL, payload, 30)
	if _err != nil {
		err = _err
		return
	}

	resp := data.IELoadModelResp{}
	err = json.Unmarshal([]byte(body), &resp)
	if err != nil {
		return
	} else if !strings.EqualFold(resp.Status, "OK") {
		if resp.Error != "" {
			err = fmt.Errorf("status: %s; error: %s", resp.Status, resp.Error)
		} else {
			err = fmt.Errorf("status: %s", resp.Status)
		}
	}
	return
}

func ieModelLoaded(payload map[string]interface{}) (bool, error) {
	intentEngineURL := getEnvironment("INTENT_ENGINE_URL")
	if intentEngineURL == "" {
		intentEngineURL = defaultIntentEngineURL
	}
	iePredictURL := fmt.Sprintf("%s/%s", intentEngineURL, "predict")

	body, err := util.HTTPPostJSON(iePredictURL, payload, 30)
	if err != nil {
		return false, err
	}

	resp := data.IEPredictResp{}
	err = json.Unmarshal([]byte(body), &resp)
	if err != nil {
		return false, err
	}

	if !strings.EqualFold(resp.Status, "OK") {
		if strings.HasPrefix(resp.Error, "Model is loading") {
			return false, nil
		}
		return false, errors.New(resp.Error)
	}

	return true, nil
}

func predictSentences(version int64, appID string,
	sentences []*data.IntentTestSentence, resultsChan chan *data.TestResult) {
	intentEngineURL := getEnvironment("INTENT_ENGINE_URL")
	if intentEngineURL == "" {
		intentEngineURL = defaultIntentEngineURL
	}
	iePredictURL := fmt.Sprintf("%s/%s", intentEngineURL, "predict")

	testResult := &data.TestResult{}
	payload := map[string]interface{}{
		"app_id": appID,
	}

	logger.Info.Printf("Start predicting %d intent sentences...", len(sentences))

	var sentenceErr error
	testCount := 0

	for _, sentence := range sentences {
		retryLeft := maxRetries
		sentenceErr = nil

		for {
			if sentenceErr != nil && retryLeft == 0 {
				logger.Error.Printf("Max predicting retries has exceeded, error: %s.\n",
					sentenceErr.Error())
				break
			}

			if sentenceErr != nil {
				logger.Error.Printf("Predict failed, retry predicting intent sentence, %d retries left...",
					retryLeft)
			}

			retryLeft--
			sentenceErr = nil

			payload["sentence"] = zhconverter.T2S(sentence.Sentence)
			body, err := util.HTTPPostJSON(iePredictURL, payload, 30)
			if err != nil {
				sentenceErr = err
				continue
			}

			resp := data.IEPredictResp{}
			err = json.Unmarshal([]byte(body), &resp)
			if err != nil {
				sentenceErr = err
				continue
			}

			if !strings.EqualFold(resp.Status, "OK") {
				var err error
				if resp.Error != "" {
					err = fmt.Errorf("status: %s; error: %s", resp.Status,
						resp.Error)
				} else {
					err = fmt.Errorf("status: %s", resp.Status)
				}
				sentenceErr = err
				continue
			}

			if sentence.IntentName != "" {
				if len(resp.Predictions) == 0 {
					sentence.Result = data.TestResultWrong
					testResult.FalseNegatives++
				} else {
					prediction := resp.Predictions[0]

					if sentence.IntentName == prediction.Label {
						sentence.Result = data.TestResultCorrect
						sentence.Score = &prediction.Score
						sentence.Answer = &prediction.Label
						testResult.TruePositives++
					} else {
						sentence.Result = data.TestResultWrong
						sentence.Score = &prediction.Score
						sentence.Answer = &prediction.Label
						testResult.FalseNegatives++
					}
				}
			} else {
				// Negatives tests
				if len(resp.Predictions) == 0 {
					sentence.Result = data.TestResultCorrect
					testResult.TrueNegatives++
				} else {
					prediction := resp.Predictions[0]
					sentence.Result = data.TestResultWrong
					sentence.Score = &prediction.Score
					sentence.Answer = &prediction.Label
					testResult.FalsePositives++
				}
			}

			if sentenceErr == nil {
				break
			}
		}

		if sentenceErr != nil {
			logger.Error.Printf("Error while predicting intent sentence: %s.\n", sentenceErr.Error())
			testResult.Error = sentenceErr
			break
		}

		testCount++
		if testCount%500 == 0 {
			logger.Trace.Printf("500 sentences tested")
			// Update test progress
			intentTestDao.UpdateTestIntentsProgress(version, testCount)
			testCount = 0
		}
	}

	// Update test progress
	if testResult.Error == nil {
		logger.Info.Printf("Total %d intent sentences predicted.\n", len(sentences))
		if testCount%500 != 0 {
			// Update test progress
			intentTestDao.UpdateTestIntentsProgress(version, testCount)
		}
	}

	resultsChan <- testResult
}

func unloadIEModel(appID string) (err error) {
	intentEngineURL := getEnvironment("INTENT_ENGINE_URL")
	if intentEngineURL == "" {
		intentEngineURL = defaultIntentEngineURL
	}

	ieUnloadModelURL := fmt.Sprintf("%s/%s", intentEngineURL, "unload_model")
	payload := map[string]interface{}{
		"app_id": appID,
	}

	body, _err := util.HTTPPostJSON(ieUnloadModelURL, payload, 30)
	if _err != nil {
		err = _err
		return
	}

	resp := data.IEUnloadModelResp{}
	err = json.Unmarshal([]byte(body), &resp)
	if err != nil {
		return
	} else if !strings.EqualFold(resp.Status, "OK") {
		err = fmt.Errorf("status: %s", resp.Status)
		return
	}
	return
}

func getEnvironment(key string) string {
	envs := util.GetEnvOf("intents")
	if envs != nil {
		if val, ok := envs[key]; ok {
			return val
		}
	}
	return ""
}
