package intentenginev2

import (
	"bufio"
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"emotibot.com/emotigo/module/admin-api/Dictionary"
	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/admin-api/util/AdminErrors"
	"emotibot.com/emotigo/module/admin-api/util/localemsg"
	"emotibot.com/emotigo/pkg/logger"
	"github.com/tealeg/xlsx"
)

// GetIntents will get all intents of appid and version with keyword
func GetIntents(appid string, version *int, keyword string) ([]*IntentV2, AdminErrors.AdminError) {
	intents, err := dao.GetIntents(appid, version, keyword)
	if err == sql.ErrNoRows {
		return []*IntentV2{}, nil
	} else if err != nil {
		return nil, AdminErrors.New(AdminErrors.ErrnoDBError, err.Error())
	}

	return intents, nil
}

// GetIntent will get intent of appid and intentID with keyword
func GetIntent(appid string, intentID int64, keyword string) (*IntentV2, AdminErrors.AdminError) {
	intent, err := dao.GetIntent(appid, intentID, keyword)
	if err == sql.ErrNoRows {
		return nil, AdminErrors.New(AdminErrors.ErrnoNotFound, "")
	} else if err != nil {
		return nil, AdminErrors.New(AdminErrors.ErrnoDBError, err.Error())
	}

	return intent, nil
}

func AddIntent(appid, name string, positive, negative []string) (*IntentV2, AdminErrors.AdminError) {
	intent, err := dao.AddIntent(appid, name, positive, negative)
	if err == sql.ErrNoRows {
		return nil, AdminErrors.New(AdminErrors.ErrnoDBError, "Add fail")
	} else if err != nil {
		return nil, AdminErrors.New(AdminErrors.ErrnoDBError, err.Error())
	}

	return intent, nil
}

func ModifyIntent(appid string, intentID int64, name string,
	updateSentence []*SentenceV2WithType, deleteSentences []int64) (*IntentV2, AdminErrors.AdminError) {
	err := dao.ModifyIntent(appid, intentID, name, updateSentence, deleteSentences)
	if err == sql.ErrNoRows {
		return nil, AdminErrors.New(AdminErrors.ErrnoNotFound, "")
	} else if err != nil {
		return nil, AdminErrors.New(AdminErrors.ErrnoDBError, err.Error())
	}

	intent, err := dao.GetIntent(appid, intentID, "")
	if err != nil {
		return nil, AdminErrors.New(AdminErrors.ErrnoDBError, err.Error())
	}

	return intent, nil
}

func DeleteIntent(appid string, intentID int64) AdminErrors.AdminError {
	err := dao.DeleteIntent(appid, intentID)
	if err == sql.ErrNoRows {
		return nil
	} else if err != nil {
		return AdminErrors.New(AdminErrors.ErrnoDBError, err.Error())
	}

	return nil
}

func GetIntentEngineStatus(appid string) (ret *StatusV2, err AdminErrors.AdminError) {
	ret = &StatusV2{
		Status: statusNeedTrain,
	}
	version, daoErr := dao.GetLatestVersion(appid)
	if daoErr == sql.ErrNoRows {
		logger.Trace.Println("No any version, NEED_TRAIN")
		return
	} else if daoErr != nil {
		return nil, AdminErrors.New(AdminErrors.ErrnoDBError, daoErr.Error())
	}
	latestInfo, daoErr := dao.GetVersionInfo(appid, version)
	if daoErr != nil {
		return nil, AdminErrors.New(AdminErrors.ErrnoDBError, daoErr.Error())
	}
	ret.CurrentStartTime = latestInfo.TrainStartTime
	ret.LastFinishTime = latestInfo.TrainEndTime
	ret.Progress = latestInfo.Progress
	ret.Version = latestInfo.Version
	if ret.CurrentStartTime == nil && ret.LastFinishTime == nil {
		logger.Trace.Println("Version hasn't start, return NEED_TRAIN")
		ret.Status = statusNeedTrain
		return
	} else if ret.LastFinishTime == nil {
		logger.Trace.Println("Version hasn't end, return TRAINING")
		ret.Status = statusTraining
		return
	} else if latestInfo.TrainResult == trainResultFail {
		logger.Trace.Println("Version fail, return NEED_TRAIN")
		ret.Status = statusNeedTrain
		return
	}

	// If any intents has modified, return "NEED_TRAIN", or return "TRAINED"
	needTrain, daoErr := dao.NeedCommit(appid)
	if daoErr != nil {
		return nil, AdminErrors.New(AdminErrors.ErrnoDBError, daoErr.Error())
	}
	if needTrain {
		logger.Trace.Println("Some intents has modified, return NEED_TRAIN")
		ret.Status = statusNeedTrain
	} else {
		logger.Trace.Println("No intent modified, return TRAINED")
		ret.Status = statusFinish
	}
	return
}

func StartTrain(appid string) (version int, err AdminErrors.AdminError) {
	status, err := GetIntentEngineStatus(appid)
	if err != nil {
		return 0, err
	}
	if status.Status == statusTraining {
		return 0, AdminErrors.New(AdminErrors.ErrnoRequestError, util.Msg["PreviousStillRunning"])
	}
	if status.Status != statusNeedTrain {
		return status.Version, nil
	}
	version, _, daoErr := dao.CommitIntent(appid)
	if daoErr != nil {
		return 0, AdminErrors.New(AdminErrors.ErrnoDBError, daoErr.Error())
	}

	start := time.Now().Unix()
	modelID, apiErr := trainIntent(appid)
	if apiErr != nil {
		return 0, AdminErrors.New(AdminErrors.ErrnoAPIError, apiErr.Error())
	}
	daoErr = dao.UpdateVersionStart(version, start, modelID)
	if daoErr != nil {
		return 0, AdminErrors.New(AdminErrors.ErrnoDBError, daoErr.Error())
	}

	go checkIntentModelStatus(appid, modelID, version)

	return version, nil
}

func trainIntent(appid string) (modelID string, err error) {
	payload := map[string]interface{}{
		"app_id":      appid,
		"auto_reload": true,
	}

	intentEngineURL := getEnvironment("INTENT_ENGINE_URL")
	if intentEngineURL == "" {
		intentEngineURL = defaultIntentEngineURL
	}

	trainURL := fmt.Sprintf("%s/train", intentEngineURL)

	body, err := util.HTTPPostJSON(trainURL, payload, 30)
	if err != nil {
		return "", err
	}
	logger.Trace.Println("Get response when training intent-engine:", body)

	ret := IETrainStatus{}
	err = json.Unmarshal([]byte(body), &ret)
	if err != nil {
		return
	}
	return ret.ModelID, nil
}

func checkIntentModelStatus(appid, modelID string, version int) {
	time.Sleep(time.Second * 5)
	payload := map[string]string{
		"app_id":   appid,
		"model_id": modelID,
	}

	intentEngineURL := getEnvironment("INTENT_ENGINE_URL")
	if intentEngineURL == "" {
		intentEngineURL = defaultIntentEngineURL
	}

	trainURL := fmt.Sprintf("%s/status", intentEngineURL)

	body, err := util.HTTPPostJSON(trainURL, payload, 30)
	if err != nil {
		return
	}
	logger.Trace.Println("Get response when training intent-engine:", body)
	ret := IETrainStatus{}
	err = json.Unmarshal([]byte(body), &ret)
	if err != nil {
		return
	}

	now := time.Now().Unix()
	switch ret.Status {
	case statusIETrainError:
		dao.UpdateVersionStatus(version, now, trainResultFail)
	case statusIETrainReady:
		dao.UpdateVersionStatus(version, now, trainResultSuccess)
		util.ConsulUpdateIntent(appid)
	default:
		go checkIntentModelStatus(appid, modelID, version)
	}
}

func GetTrainData(appid string) (*TrainDataResponse, AdminErrors.AdminError) {
	wordbanks, errno, err := Dictionary.GetWordbanksV3(appid)
	if err != nil {
		return nil, AdminErrors.New(errno, err.Error())
	}
	version, err := dao.GetLatestVersion(appid)
	if err == sql.ErrNoRows {
		return nil, AdminErrors.New(AdminErrors.ErrnoRequestError, "")
	} else if err != nil {
		return nil, AdminErrors.New(AdminErrors.ErrnoDBError, err.Error())
	}
	intents, err := dao.GetIntentsDetail(appid, &version)
	if err == sql.ErrNoRows {
		return nil, AdminErrors.New(AdminErrors.ErrnoRequestError, "")
	} else if err != nil {
		return nil, AdminErrors.New(AdminErrors.ErrnoDBError, err.Error())
	}

	ret := NewTrainDataResponse(appid)
	for idx := range intents {
		temp := TrainIntent{}
		temp.Load(intents[idx])
		ret.Intent = append(ret.Intent, &temp)
	}
	intentDict, err := getIntentDictResp(wordbanks, []string{})
	if err != nil {
		return nil, AdminErrors.New(AdminErrors.ErrnoIOError, err.Error())
	}

	if intentDict != nil {
		ret.IntentDict = intentDict
	}

	return ret, nil
}

func getIntentDictResp(wordBankClass *Dictionary.WordBankClassV3,
	classNames []string) (dicts []*TrainDict, err error) {
	if !wordBankClass.IntentEngine {
		return nil, nil
	}

	if wordBankClass.Children != nil && len(wordBankClass.Children) > 0 {
		for _, child := range wordBankClass.Children {
			_classNames := make([]string, len(classNames))

			// Ignore workbank class name of root node (mock node, ID = -1)
			if wordBankClass.ID != -1 {
				_classNames = append(classNames, wordBankClass.Name)
			}

			// Recursively call to iterate all children nodes
			subDicts, err := getIntentDictResp(child, _classNames)
			if err != nil {
				return nil, err
			}
			dicts = append(dicts, subDicts...)
		}
	}

	if wordBankClass.Wordbank != nil && len(wordBankClass.Wordbank) > 0 {
		for _, w := range wordBankClass.Wordbank {
			// Ignore wordbanks under root node (mock node, ID = -1)
			if wordBankClass.ID != -1 {
				_classNames := make([]string, len(classNames))
				_classNames = append(classNames, wordBankClass.Name)

				d := TrainDict{}
				d.ClassName = _classNames
				d.DictName = w.Name
				d.Words = w.SimilarWords
				dicts = append(dicts, &d)
			}
		}
	}

	return nil, nil
}

func UpdateLatestIntents(appid string, intents []*IntentV2) AdminErrors.AdminError {
	dbErr := dao.UpdateLatestIntents(appid, intents)
	if dbErr != nil {
		return AdminErrors.New(AdminErrors.ErrnoDBError, dbErr.Error())
	}
	return nil
}

func ParseImportIntentFile(buf []byte, locale string) (intents []*IntentV2, err error) {
	file, err := xlsx.OpenBinary(buf)
	if err != nil {
		return nil, err
	}
	format := typeBFOP

	sheets := file.Sheets
	if len(sheets) <= 0 {
		return nil, errors.New(localemsg.Get(locale, "IntentUploadSheetErr"))
	}

	for idx := range sheets {
		if sheets[idx].Name == localemsg.Get(locale, "IntentBF2Sheet1Name") ||
			sheets[idx].Name == localemsg.Get(locale, "IntentBF2Sheet2Name") {
			format = typeBF2
			break
		}
	}

	// if len(sheets) == 2 {
	// 	if sheets[0].Name == localemsg.Get(locale, "IntentBF2Sheet1Name") &&
	// 		sheets[1].Name == localemsg.Get(locale, "IntentBF2Sheet2Name") {
	// 		format = typeBF2
	// 	}
	// }

	if format == typeBFOP {
		logger.Trace.Println("Parse file with BFOP type")
		return parseBFOPSheets(sheets, locale)
	} else if format == typeBF2 {
		logger.Trace.Println("Parse file with BF2 type")
		return parseBF2Sheets(sheets, locale)
	}
	// this line must not happen
	return nil, errors.New("Invalid type")
}

func getBFOPColumnIdx(row *xlsx.Row, locale string) (posIdx, negIdx int) {
	posIdx, negIdx = -1, -1
	for idx := range row.Cells {
		cellStr := row.Cells[idx].String()
		if cellStr == localemsg.Get(locale, "IntentPositive") {
			posIdx = idx
		} else if cellStr == localemsg.Get(locale, "IntentNegative") {
			negIdx = idx
		}
	}
	return
}

func parseBFOPSheets(sheets []*xlsx.Sheet, locale string) (intents []*IntentV2, err error) {
	intents = []*IntentV2{}
	// parse each sheet, each sheet is an intnet
	for idx := range sheets {
		rows := sheets[idx].Rows
		if len(rows) == 0 {
			return nil, fmt.Errorf(localemsg.Get(locale, "IntentUploadNoHeaderTpl"), sheets[idx].Name)
		}
		intent := IntentV2{}
		intent.Name = sheets[idx].Name
		posIdx, negIdx := getBFOPColumnIdx(rows[0], locale)
		if posIdx < 0 || posIdx > 1 || negIdx < 0 || negIdx > 1 {
			return nil, fmt.Errorf(localemsg.Get(locale, "IntentUploadHeaderErrTpl"), sheets[idx].Name)
		}
		posList := []*SentenceV2{}
		negList := []*SentenceV2{}
		// parse each row
		rows = rows[1:]
		for rowIdx := range rows {
			cells := rows[rowIdx].Cells
			pos, neg := "", ""
			if posIdx < len(cells) {
				pos = strings.TrimSpace(cells[posIdx].String())
			}
			if negIdx < len(cells) {
				neg = strings.TrimSpace(cells[negIdx].String())
			}
			if pos != "" {
				posList = append(posList, &SentenceV2{Content: pos})
			}
			if neg != "" {
				negList = append(negList, &SentenceV2{Content: neg})
			}
		}
		intent.Positive = &posList
		intent.PositiveCount = len(posList)
		intent.Negative = &negList
		intent.NegativeCount = len(negList)

		intents = append(intents, &intent)
	}
	return
}

func getBF2ColumnIdx(row *xlsx.Row, locale string) (nameIdx, sentenceIdx int) {
	nameIdx, sentenceIdx = -1, -1
	for idx := range row.Cells {
		cellStr := row.Cells[idx].String()
		if cellStr == localemsg.Get(locale, "IntentName") {
			nameIdx = idx
		} else if cellStr == localemsg.Get(locale, "IntentSentence") {
			sentenceIdx = idx
		}
	}
	logger.Trace.Printf("Uplaod column idx: %d %d\n", nameIdx, sentenceIdx)
	return
}

func parseBF2Sheets(sheets []*xlsx.Sheet, locale string) (intents []*IntentV2, err error) {
	// if len(sheets) != 2 {
	// 	return nil, errors.New(localemsg.Get(locale, "IntentUploadSheetErr"))
	// }

	intentMap := map[string]*IntentV2{}
	for idx := range sheets {
		var sentenceType int
		if sheets[idx].Name == localemsg.Get(locale, "IntentBF2Sheet1Name") {
			sentenceType = typePositive
		} else if sheets[idx].Name == localemsg.Get(locale, "IntentBF2Sheet2Name") {
			sentenceType = typeNegative
		} else {
			continue
		}
		rows := sheets[idx].Rows
		if len(rows) == 0 {
			return nil, fmt.Errorf(localemsg.Get(locale, "IntentUploadNoHeaderTpl"), sheets[idx].Name)
		}

		nameIdx, sentenceIdx := getBF2ColumnIdx(rows[0], locale)
		if nameIdx < 0 || nameIdx > 1 || sentenceIdx < 0 || sentenceIdx > 1 {
			return nil, fmt.Errorf(localemsg.Get(locale, "IntentUploadHeaderErrTpl"), sheets[idx].Name)
		}

		rows = rows[1:]
		for rowIdx := range rows {
			cells := rows[rowIdx].Cells
			if len(cells) != 2 {
				return nil, fmt.Errorf(localemsg.Get(locale, "IntentUploadBF2RowInvalidTpl"),
					sheets[idx].Name, rowIdx+1)
			}
			name := strings.TrimSpace(cells[nameIdx].String())
			sentence := strings.TrimSpace(cells[sentenceIdx].String())
			if name == "" {
				return nil, fmt.Errorf(localemsg.Get(locale, "IntentUploadBF2RowNoNameTpl"),
					sheets[idx].Name, rowIdx+1)
			}
			if sentence == "" {
				return nil, fmt.Errorf(localemsg.Get(locale, "IntentUploadBF2RowNoSentenceTpl"),
					sheets[idx].Name, rowIdx+1)
			}

			if _, ok := intentMap[name]; !ok {
				intentMap[name] = &IntentV2{Name: name}
				intentMap[name].Positive = &([]*SentenceV2{})
				intentMap[name].Negative = &([]*SentenceV2{})
			}

			if sentenceType == typePositive {
				newList := append(*intentMap[name].Positive, &SentenceV2{Content: sentence})
				intentMap[name].PositiveCount++
				intentMap[name].Positive = &newList
			} else {
				newList := append(*intentMap[name].Negative, &SentenceV2{Content: sentence})
				intentMap[name].NegativeCount++
				intentMap[name].Negative = &newList
			}
		}
	}

	intents = []*IntentV2{}
	for name := range intentMap {
		intents = append(intents, intentMap[name])
	}
	return
}

func GetExportIntents(appid string, locale string) (ret []byte, err AdminErrors.AdminError) {
	intents, daoErr := dao.GetIntentsDetail(appid, nil)
	if daoErr != nil {
		return
	}

	file := xlsx.NewFile()
	for idx := range intents {
		intent := intents[idx]
		sheet, xlsxErr := file.AddSheet(intent.Name)
		if xlsxErr != nil {
			err = AdminErrors.New(AdminErrors.ErrnoIOError, xlsxErr.Error())
			return
		}
		// add header
		headerRow := sheet.AddRow()
		headerRow.AddCell().SetString(localemsg.Get(locale, "IntentPositive"))
		headerRow.AddCell().SetString(localemsg.Get(locale, "IntentNegative"))

		// each row is a positive sentence or negative sentence
		for idx := 0; idx < intent.PositiveCount || idx < intent.NegativeCount; idx++ {
			row := sheet.AddRow()
			posCell := row.AddCell()
			negCell := row.AddCell()
			if idx < intent.PositiveCount {
				posCell.SetString((*intent.Positive)[idx].Content)
			}
			if idx < intent.NegativeCount {
				negCell.SetString((*intent.Negative)[idx].Content)
			}
		}
	}

	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)
	ioErr := file.Write(writer)
	if ioErr != nil {
		err = AdminErrors.New(AdminErrors.ErrnoIOError, ioErr.Error())
		return
	}
	return buf.Bytes(), nil
}

func GetExportIntentsBFFormat(appid string, locale string) (ret []byte, err AdminErrors.AdminError) {
	intents, daoErr := dao.GetIntentsDetail(appid, nil)
	if daoErr != nil {
		return
	}

	file := xlsx.NewFile()
	sheetPositive, xlsxErr := file.AddSheet(localemsg.Get(locale, "IntentBF2Sheet1Name"))
	if xlsxErr != nil {
		err = AdminErrors.New(AdminErrors.ErrnoIOError, xlsxErr.Error())
		return
	}
	sheetNegative, xlsxErr := file.AddSheet(localemsg.Get(locale, "IntentBF2Sheet2Name"))
	if xlsxErr != nil {
		err = AdminErrors.New(AdminErrors.ErrnoIOError, xlsxErr.Error())
		return
	}
	sheets := []*xlsx.Sheet{sheetPositive, sheetNegative}
	for _, sheet := range sheets {
		headerRow := sheet.AddRow()
		headerRow.AddCell().SetString(localemsg.Get(locale, "IntentName"))
		headerRow.AddCell().SetString(localemsg.Get(locale, "IntentSentence"))
	}

	for idx := range intents {
		intent := intents[idx]
		if intent.Positive != nil {
			for _, sentence := range *intent.Positive {
				row := sheetPositive.AddRow()
				row.AddCell().SetString(intent.Name)
				row.AddCell().SetString(sentence.Content)
			}
		}
		if intent.Negative != nil {
			for _, sentence := range *intent.Negative {
				row := sheetNegative.AddRow()
				row.AddCell().SetString(intent.Name)
				row.AddCell().SetString(sentence.Content)
			}
		}
	}

	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)
	ioErr := file.Write(writer)
	if ioErr != nil {
		err = AdminErrors.New(AdminErrors.ErrnoIOError, ioErr.Error())
		return
	}
	return buf.Bytes(), nil
}

// SearchSentence will do search of intent contain the sentence equals to content, and return intent name, sentence type
// If err happens, intentName will be empty string and sentenceType will be 0, err will be an not null Error
func SearchSentence(appid string, version *int, content string) (intentName string, sentenceType int, err AdminErrors.AdminError) {
	intent, sentence, dbErr := dao.SearchIntentOfSentence(appid, version, content)
	if dbErr == sql.ErrNoRows {
		err = AdminErrors.New(AdminErrors.ErrnoNotFound, "")
		return
	}
	return intent.Name, sentence.Type, nil
}

// SearchSentenceWithType will do same with SearchSentence, only add sentenceType as filter
func SearchSentenceWithType(appid string, version *int, content string, sentenceType int) (intentName string, err AdminErrors.AdminError) {
	intent, _, dbErr := dao.SearchIntentOfSentence(appid, version, content, sentenceType)
	if dbErr == sql.ErrNoRows {
		err = AdminErrors.New(AdminErrors.ErrnoNotFound, "")
		return
	} else if dbErr != nil {
		err = AdminErrors.New(AdminErrors.ErrnoDBError, dbErr.Error())
		return
	}
	return intent.Name, nil
}
