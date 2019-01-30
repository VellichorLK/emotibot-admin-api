package intentengine

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

	"emotibot.com/emotigo/module/admin-api/ApiError"
	"emotibot.com/emotigo/module/admin-api/dictionary"
	"emotibot.com/emotigo/module/admin-api/util/localemsg"
	"emotibot.com/emotigo/pkg/logger"
	"github.com/tealeg/xlsx"
)

const intentFilesPath = "./statics"

const (
	TrainBothEngines = iota
	TrainIntentEngine
	TrainRuleEngine
)

func GetIntents(appID string, version int) (intents []string, retCode int, err error) {
	retCode = ApiError.SUCCESS
	intents = []string{}
	if version == 0 {
		// Use latest version of intents dataset
		version, err = getLatestIntentsVersion(appID)
		if err != nil {
			retCode = ApiError.DB_ERROR
			return
		} else if version == -1 {
			// no existed intents
			return
		}
	}

	intents, err = getIntents(appID, version)
	if err != nil {
		retCode = ApiError.DB_ERROR
		return
	}

	retCode = ApiError.SUCCESS
	return
}

func UploadIntents(appID string, file multipart.File, info *multipart.FileHeader) (version int, retCode int, err error) {
	logger.Info.Printf("Receive uploaded file: %s", info.Filename)
	logger.Trace.Printf("Uploaded file info %#v", info.Header)

	buf := make([]byte, info.Size)
	_, err = file.Read(buf)
	if err != nil {
		retCode = ApiError.IO_ERROR
		return
	}

	// TODO: Check upload file

	intents, err := ParseIntentsFromXLSX(buf)
	if err != nil {
		retCode = ApiError.INTENT_FORMAT_ERROR
		return
	}

	now := time.Now()
	renamedFileName := fmt.Sprintf("intents_%s.xlsx", now.Format("20060102150405"))

	version, err = insertIntents(appID, intents, info.Filename, renamedFileName)
	if err != nil {
		retCode = ApiError.DB_ERROR
		return
	}

	err = saveIntentsFile(buf, renamedFileName)
	if err != nil {
		retCode = ApiError.IO_ERROR
		return
	}

	retCode = ApiError.SUCCESS
	return
}

func GetDownloadIntents(appID string, version int, format string) ([]byte, string, int, error) {
	if version == 0 {
		// Use latest version of intents dataset
		ver, err := getLatestIntentsVersion(appID)
		if err != nil {
			return nil, "", ApiError.DB_ERROR, err
		} else if ver == -1 {
			// No any version of intents existed
			return nil, "", ApiError.NOT_FOUND_ERROR, errors.New("No any valid intent")
		}

		version = ver
	}

	_, origFileName, err := getIntentsXSLXFileName(appID, version)
	if err != nil {
		errMsg := fmt.Sprintf("Cannot find intents file with version %v", version)
		logger.Error.Printf(errMsg)
		return nil, "", ApiError.NOT_FOUND_ERROR, errors.New(errMsg)
	}

	intents, err := getIntentDetails(appID, version)
	if err != nil {
		return nil, "", ApiError.DB_ERROR, err
	}

	xlsxFile := xlsx.NewFile()
	if strings.ToLower(format) == "bfop" {
		err = fillBFOPExport(xlsxFile, intents)
	} else {
		err = fillBF2Export(xlsxFile, intents)
	}
	if err != nil {
		return nil, "", ApiError.IO_ERROR, err
	}

	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)
	err = xlsxFile.Write(writer)
	if err != nil {
		return nil, "", ApiError.IO_ERROR, err
	}
	return buf.Bytes(), string(origFileName), ApiError.SUCCESS, nil
}

func fillBFOPExport(xlsxFile *xlsx.File, intents []*Intent) error {
	for idx := range intents {
		intent := intents[idx]
		sheet, err := xlsxFile.AddSheet(intent.Name)
		if err != nil {
			return err
		}
		for _, sentence := range intent.Sentences {
			sheet.AddRow().AddCell().SetString(sentence)
		}
	}
	return nil
}

func fillBF2Export(xlsxFile *xlsx.File, intents []*Intent) error {
	sheet, err := xlsxFile.AddSheet(localemsg.Get("zh-cn", "IntentBF2Sheet1Name"))
	row := sheet.AddRow()
	row.AddCell().SetString(localemsg.Get("zh-cn", "IntentName"))
	row.AddCell().SetString(localemsg.Get("zh-cn", "IntentSentence"))

	if err != nil {
		return err
	}
	for idx := range intents {
		intent := intents[idx]
		for _, sentence := range intent.Sentences {
			row := sheet.AddRow()
			row.AddCell().SetString(intent.Name)
			row.AddCell().SetString(sentence)
		}
	}
	return nil
}

func Train(appID string, version int, autoReload bool, trainEngine int) (retCode int, err error) {
	if version == 0 {
		// Use latest version of intents dataset
		version, err = getLatestIntentsVersion(appID)
		if err != nil {
			retCode = ApiError.DB_ERROR
			return
		} else if version == -1 {
			// No any version of intents existed
			retCode = ApiError.NOT_FOUND_ERROR
			return
		}
	}

	// Train Intent Engine if required
	if trainEngine == TrainBothEngines || trainEngine == TrainIntentEngine {
		ieModelID, err := getIntentEngineModelID(appID, version)
		if err != nil {
			if string(ieModelID) == "" {
				// Cannot find corresponded version (sql.ErrNoRows)
				retCode = ApiError.REQUEST_ERROR
				return retCode, err
			}
		}

		if ieModelID == nil {
			retCode, err = trainIntentEngine(appID, version, autoReload)
			if err != nil {
				return retCode, err
			}
		}
	}

	// Train Rule Engine if required
	if trainEngine == TrainBothEngines || trainEngine == TrainRuleEngine {
		reModelID, err := getRuleEngineModelID(appID, version)
		if err != nil {
			if string(reModelID) == "" {
				// Cannot find corresponded version (sql.ErrNoRows)
				retCode = ApiError.REQUEST_ERROR
				return retCode, err
			}
		}

		if reModelID == nil {
			retCode, err = trainRuleEngine(appID, version, autoReload)
			if err != nil {
				return retCode, err
			}
		}
	}

	retCode = ApiError.SUCCESS
	return
}

func trainIntentEngine(appID string, version int, autoReload bool) (retCode int, err error) {
	// TODO: Supports different versions of intents dataset

	payload := make(map[string]interface{})
	payload["app_id"] = appID
	payload["auto_reload"] = autoReload
	payloadStr, _ := json.Marshal(payload)

	intentEngineURL := getEnvironment("INTENT_ENGINE_URL")
	if intentEngineURL == "" {
		intentEngineURL = defaultIntentEngineURL
	}

	ieTrainURL := fmt.Sprintf("%s/%s", intentEngineURL, "train")
	resp, err := http.Post(ieTrainURL, "application/json; charset=utf-8", bytes.NewBuffer(payloadStr))
	if err != nil {
		retCode = ApiError.WEB_REQUEST_ERROR
		return
	}
	defer resp.Body.Close()

	response := TrainResponse{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		retCode = ApiError.WEB_REQUEST_ERROR
		return
	}

	err = updateIntentEngineModelID(appID, []byte(response.ModelID), version)
	if err != nil {
		retCode = ApiError.DB_ERROR
		return
	}

	retCode = ApiError.SUCCESS
	return
}

func trainRuleEngine(appID string, version int, autoReload bool) (retCode int, err error) {
	// TODO: Supports different versions of dictionary dataset

	payload := make(map[string]interface{})
	payload["app_id"] = appID
	payload["auto_reload"] = autoReload
	payloadStr, _ := json.Marshal(payload)

	ruleEngineURL := getEnvironment("RULE_ENGINE_URL")
	if ruleEngineURL == "" {
		ruleEngineURL = defaultRuleEngineURL
	}

	reTrainURL := fmt.Sprintf("%s/%s", ruleEngineURL, "train")
	resp, err := http.Post(reTrainURL, "application/json; charset=utf-8", bytes.NewBuffer(payloadStr))
	if err != nil {
		retCode = ApiError.WEB_REQUEST_ERROR
		return
	}
	defer resp.Body.Close()

	response := TrainResponse{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		retCode = ApiError.WEB_REQUEST_ERROR
		return
	}

	err = updateRuleEngineModelID(appID, []byte(response.ModelID), version)
	if err != nil {
		retCode = ApiError.DB_ERROR
		return
	}

	retCode = ApiError.SUCCESS
	return
}

func GetTrainStatus(appID string, version int) (statusResp StatusResponse, retCode int, err error) {
	if version == 0 {
		// Use latest version of intents dataset
		version, err = getLatestIntentsVersion(appID)
		if err != nil {
			retCode = ApiError.DB_ERROR
			return
		} else if version == -1 {
			// No any version of intents existed
			retCode = ApiError.NOT_FOUND_ERROR
			return
		}
	}

	ieModelID, err := getIntentEngineModelID(appID, version)
	if err != nil {
		if string(ieModelID) == "" {
			// Cannot find corresponded version (sql.ErrNoRows)
			retCode = ApiError.NOT_FOUND_ERROR
			return
		}
	}

	if ieModelID == nil {
		statusResp.IntentEngineStatus = NotTrained
	} else {
		// Query Intent Engine status
		payload := map[string]string{
			"app_id":   appID,
			"model_id": string(ieModelID),
		}
		payloadStr, _ := json.Marshal(payload)

		intentEngineURL := getEnvironment("INTENT_ENGINE_URL")
		if intentEngineURL == "" {
			intentEngineURL = defaultIntentEngineURL
		}

		ieStatusURL := fmt.Sprintf("%s/%s", intentEngineURL, "status")
		resp, _err := http.Post(ieStatusURL, "application/json; charset=utf-8", bytes.NewBuffer(payloadStr))
		if err != nil {
			retCode = ApiError.WEB_REQUEST_ERROR
			err = _err
			return
		}
		defer resp.Body.Close()

		ieStatus := IntentEngineStatusResponse{}
		err = json.NewDecoder(resp.Body).Decode(&ieStatus)
		if err != nil {
			retCode = ApiError.WEB_REQUEST_ERROR
			return
		}
		logger.Trace.Printf("Query status with version %d: %+v\n", version, payload)
		logger.Trace.Printf("Get response from intent status: %+v\n", ieStatus)

		switch ieStatus.Status {
		case "training":
			statusResp.IntentEngineStatus = Training
		case "ready":
			statusResp.IntentEngineStatus = Trained
		case "error":
			statusResp.IntentEngineStatus = TrainFailed
		default:
			statusResp.IntentEngineStatus = NotTrained
		}
	}

	reModelID, err := getRuleEngineModelID(appID, version)
	if err != nil {
		if string(reModelID) == "" {
			// Cannot find corresponded version (sql.ErrNoRows)
			retCode = ApiError.NOT_FOUND_ERROR
			return
		}
	}

	if reModelID == nil {
		statusResp.RuleEngineStatus = NotTrained
	} else {
		// Query Rule Engine status
		payload := map[string]string{
			"app_id":   appID,
			"model_id": string(reModelID),
		}
		payloadStr, _ := json.Marshal(payload)

		ruleEngineURL := getEnvironment("RULE_ENGINE_URL")
		if ruleEngineURL == "" {
			ruleEngineURL = defaultRuleEngineURL
		}

		reStatusURL := fmt.Sprintf("%s/%s", ruleEngineURL, "status")
		resp, _err := http.Post(reStatusURL, "application/json; charset=utf-8", bytes.NewBuffer(payloadStr))
		if err != nil {
			retCode = ApiError.WEB_REQUEST_ERROR
			err = _err
			return
		}
		defer resp.Body.Close()

		reStatus := RuleEngineStatusResponse{}
		err = json.NewDecoder(resp.Body).Decode(&reStatus)
		if err != nil {
			retCode = ApiError.WEB_REQUEST_ERROR
			return
		}

		switch reStatus.Status {
		case "training":
			statusResp.RuleEngineStatus = Training
		case "ready":
			statusResp.RuleEngineStatus = Trained
		case "error":
			statusResp.RuleEngineStatus = TrainFailed
		default:
			statusResp.RuleEngineStatus = NotTrained
		}
	}

	retCode = ApiError.SUCCESS
	return
}

func GetTrainingData(appID string, flag string, version ...int) (resp interface{}, retCode int, err error) {
	var ver int

	if len(version) == 0 {
		ver, err = getLatestIntentsVersion(appID)
		if err != nil {
			retCode = ApiError.DB_ERROR
			return
		} else if ver == -1 {
			// No any version of intents existed
			retCode = ApiError.NOT_FOUND_ERROR
			return
		}
	} else {
		ver = version[0]
	}

	wordbanks, _, err := dictionary.GetWordbanksV3(appID)
	if err != nil {
		retCode = ApiError.DB_ERROR
		return
	}

	switch flag {
	case "intent_engine":
		resp, err = getIntentEngineTrainingData(appID, ver, wordbanks)
		if err != nil {
			retCode = ApiError.DB_ERROR
			return
		}
	case "rule_engine":
		resp, err = getRuleEngineTrainingData(appID, ver, wordbanks)
		if err != nil {
			retCode = ApiError.DB_ERROR
			return
		}
	default:
		retCode = ApiError.REQUEST_ERROR
		err = errors.New("flag must be 'intent_engine' or 'rule_engine'")
		return
	}

	retCode = ApiError.SUCCESS
	return
}

func getIntentEngineTrainingData(appID string, version int,
	wordbanks *dictionary.WordBankClassV3) (IntentEngineGetDataResponse, error) {
	ret := NewIntentEngineGetDataResponse()
	ret.AppID = appID

	intents, err := getIntentDetails(appID, version)
	if err != nil {
		ret.Status = "ERROR"
		return ret, err
	}

	for _, intent := range intents {
		sentencesResp := NewIntentSentencesResponse()
		sentencesResp.Positive = append(sentencesResp.Positive, intent.Sentences...)

		intentResp := IntentResponse{
			Name:      intent.Name,
			Sentences: &sentencesResp,
			Features:  NewIntentFeaturesResponse(),
		}
		ret.Intent = append(ret.Intent, &intentResp)
	}

	dicts := make([]interface{}, 0)
	classNames := make([]string, 0)

	err = getDictResp("intent_engine", &dicts, wordbanks, classNames)
	if err != nil {
		ret.Status = "ERROR"
		return ret, err
	}

	// Copy elements of []interface and convert element type to *IntentDictResponse
	_dicts := make([]*IntentDictResponse, len(dicts))
	for i, dict := range dicts {
		_dicts[i] = dict.(*IntentDictResponse)
	}

	ret.Status = "OK"
	ret.IntentDict = _dicts

	return ret, nil
}

func getRuleEngineTrainingData(appID string, version int,
	wordbanks *dictionary.WordBankClassV3) (RuleEngineGetDataResponse, error) {
	ret := NewRuleEngineGetDataResponse()
	ret.AppID = appID

	dicts := make([]interface{}, 0)
	classNames := make([]string, 0)

	err := getDictResp("rule_engine", &dicts, wordbanks, classNames)
	if err != nil {
		ret.Status = "ERROR"
		return ret, err
	}

	// Copy elements of []interface and convert element type to *RuleEngineDictResponse
	_dicts := make([]*RuleEngineDictResponse, len(dicts))
	for i, dict := range dicts {
		_dicts[i] = dict.(*RuleEngineDictResponse)
	}

	ret.Status = "OK"
	ret.Dict = _dicts

	return ret, nil
}

func getDictResp(flag string, dicts *[]interface{},
	wordBankClass *dictionary.WordBankClassV3,
	classNames []string) error {
	switch flag {
	case "intent_engine":
		if !wordBankClass.IntentEngine {
			return nil
		}
	case "rule_engine":
		if !wordBankClass.RuleEngine {
			return nil
		}
	default:
		errMsg := fmt.Sprintf("Unknown flag: %s", flag)
		return errors.New(errMsg)
	}

	if wordBankClass.Children != nil && len(wordBankClass.Children) > 0 {
		for _, child := range wordBankClass.Children {
			_classNames := make([]string, len(classNames))

			// Ignore workbank class name of root node (mock node, ID = -1)
			if wordBankClass.ID != -1 {
				_classNames = append(classNames, wordBankClass.Name)
			}

			// Recursively call to iterate all children nodes
			getDictResp(flag, dicts, child, _classNames)
		}
	}

	if wordBankClass.Wordbank != nil && len(wordBankClass.Wordbank) > 0 {
		for _, w := range wordBankClass.Wordbank {
			// Ignore wordbanks under root node (mock node, ID = -1)
			if wordBankClass.ID != -1 {
				_classNames := make([]string, len(classNames))
				_classNames = append(classNames, wordBankClass.Name)

				switch flag {
				case "intent_engine":
					d := NewIntentDictResponse()
					d.ClassName = _classNames
					d.DictName = w.Name
					d.Words = w.SimilarWords
					*dicts = append(*dicts, &d)
				case "rule_engine":
					d := NewRuleEngineDictResponse()
					d.ClassName = _classNames
					d.DictName = w.Name
					d.Words = w.SimilarWords
					*dicts = append(*dicts, &d)
				default:
					errMsg := fmt.Sprintf("Unknown flag: %s", flag)
					return errors.New(errMsg)
				}
			}
		}
	}

	return nil
}

func saveIntentsFile(file []byte, fileName string) (err error) {
	if _, err = os.Stat(intentFilesPath); os.IsNotExist(err) {
		err = os.Mkdir(intentFilesPath, os.ModePerm)
		if err != nil {
			logger.Error.Printf("Fail to ./statics directory")
			return
		}
	}

	f, err := os.OpenFile(intentFilesPath+"/"+fileName, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		logger.Error.Printf("Fail to save uploaded intents file %s", fileName)
		return
	}
	defer f.Close()

	_, err = f.Write(file)
	if err != nil {
		logger.Error.Printf("Fail to save uploaded intents file %s", fileName)
		return
	}

	return
}
