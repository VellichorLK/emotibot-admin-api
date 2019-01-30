package Task

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
	"runtime"
	"strings"

	"emotibot.com/emotigo/module/admin-api/ApiError"
	"emotibot.com/emotigo/module/admin-api/dictionary"
	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/pkg/logger"
)

const (
	teDefaultConfig = `{"system":"bfop","task_engine_v2":{"enable_js_code":false,"enable_node":{"nlu_pc":false,"action":false}}}`
)

func EnableAllScenario(appid string) {
	setAllScenarioStatus(appid, true)
}

func DisableAllScenario(appid string) {
	setAllScenarioStatus(appid, false)
}

func EnableScenario(appid string, scenarioID string) {
	setScenarioStatus(appid, scenarioID, true)
}

func DisableScenario(appid string, scenarioID string) {
	setScenarioStatus(appid, scenarioID, false)
}

func ReadUploadJSON(file multipart.File) (string, error) {
	buf := bytes.NewBuffer(nil)
	size, err := io.Copy(buf, file)
	if err != nil {
		return "", err
	}
	if size == 0 {
		return "", fmt.Errorf("Size of file is zero")
	}
	return buf.String(), nil
}

func BfbImportScenarios(appid string, newID bool, datas []interface{}) {
	for _, data := range datas {
		BfbImportScenario(appid, newID, data)
	}
}

func BfbImportScenario(appid string, newID bool, data interface{}) {
	// logger.Info.Printf("Import scenario, use new ID: %t \ndata: %#v\n", newID, data)
	if !newID {
		bfbUpdateScenario(appid, data)
	} else {
		ret, err := addScenario(appid, data)
		if err != nil {
			fmt.Printf("Add scenario content: %s, error: %s\n", ret, err.Error())
		}
	}
}

func parseJSONData(data interface{}) (string, string, interface{}, interface{}) {
	var content map[string]interface{}
	var layout interface{}
	taskEngineData, ok := data.(*map[string]interface{})
	if ok {
		content = (*taskEngineData)["taskScenario"].(map[string]interface{})
		layout = (*taskEngineData)["taskLayouts"]
	} else {
		taskEngineData := data.(map[string]interface{})
		content = taskEngineData["taskScenario"].(map[string]interface{})
		layout = taskEngineData["taskLayouts"]
	}
	metadata := content["metadata"].(map[string]interface{})
	id := metadata["scenario_id"].(string)
	name := metadata["scenario_name"].(string)
	return id, name, content, layout
}

func bfbUpdateScenario(appid string, data interface{}) (string, error) {
	id, _, content, layout := parseJSONData(data)
	existed, err := checkScenarioExist(appid, id)
	if err != nil {
		return "", err
	}
	if !existed {
		return addScenario(appid, data)
	}

	postData := map[string]interface{}{
		"appid":   appid,
		"userid":  appid,
		"content": content,
		"layout":  layout,
	}
	url := fmt.Sprintf("%s/%s/%s",
		getEnvironment("SERVER_URL"),
		taskScenarioEntry,
		id)
	ret, err := util.HTTPPutForm(url, postData, 0)
	return ret, err
}
func addScenario(appid string, data interface{}) (retStr string, err error) {
	defer func() {
		// recover from panic if one occured. Set err to nil otherwise.
		if r, ok := recover().(error); ok {
			_, file, line, _ := runtime.Caller(1)
			fmt.Printf("Recovered in %s:%d - %s\n", file, line, r)
			err = errors.New("Handle add scenario fail")
		}
	}()
	_, name, content, layout := parseJSONData(data)

	postData := map[string]string{
		"appid":        appid,
		"userid":       appid,
		"scenarioname": name,
	}
	url := fmt.Sprintf("%s/%s", getEnvironment("SERVER_URL"), taskScenarioEntry)
	retData, err := util.HTTPPostForm(url, postData, 0)
	if err != nil {
		logger.Trace.Printf("Err: %s", err.Error())
		return "", err
	}
	retObj := map[string]interface{}{}
	err = json.Unmarshal([]byte(retData), &retObj)
	if err != nil {
		logger.Trace.Printf("Err: %s", err.Error())
		return "", err
	}
	id := retObj["scenarioID"].(string)
	contentObj := content.(map[string]interface{})
	metadata := contentObj["metadata"].(map[string]interface{})
	logger.Info.Printf("metadata: %#v\n", metadata)
	metadata["scenario_id"] = id

	updateData := map[string]interface{}{
		"appid":   appid,
		"userid":  appid,
		"content": content,
		"layout":  layout,
	}
	url = fmt.Sprintf("%s/%s/%s",
		getEnvironment("SERVER_URL"),
		taskScenarioEntry,
		id)
	logger.Trace.Printf("Call %s to addScenario\n", url)
	retStr, err = util.HTTPPutForm(url, updateData, 0)
	return
}

func checkScenarioExist(appid string, id string) (bool, error) {
	values := url.Values{
		"userid": []string{appid},
	}
	url := fmt.Sprintf("%s/%s/%s?%s",
		getEnvironment("SERVER_URL"),
		taskScenarioEntry,
		id,
		values.Encode())
	content, err := util.HTTPGetSimple(url)
	if err != nil {
		logger.Info.Printf("Check scenario %s existed fail: %s", id, err.Error())
		return false, err
	}

	ret := map[string]*interface{}{}
	err = json.Unmarshal([]byte(content), &ret)
	if err != nil {
		logger.Info.Printf("Check scenario %s existed fail: %s", id, err.Error())
		return false, err
	}

	if ret["result"] == nil {
		logger.Trace.Printf("Check scenario: %s not exist\n", id)
		return false, nil
	}
	logger.Trace.Printf("Check scenario: %s exist\n", id)
	return true, nil
}

func GetMapTableList(appid, userID string) ([]string, int, error) {
	list, err := getMapTableList(appid, userID)
	if err == sql.ErrNoRows {
		return []string{}, ApiError.SUCCESS, nil
	} else if err != nil {
		return []string{}, ApiError.DB_ERROR, err
	}
	return list, ApiError.SUCCESS, nil
}

// GetMapTableAllV2 parse and generate mapping table map from a WordBankClassV3 tree
func GetMapTableAllV2(rootMap map[string]*dictionary.WordBankClassV3) map[string]map[string][]*MapTuple {
	// appidToMapTupleMap map appid to a MapTupleMap which map mapTableName to a list of MapTuple
	appidToMapTupleMap := map[string]map[string][]*MapTuple{}
	for appid, root := range rootMap {
		// mtMap map mapTableName to a list of WordBankV3
		mtMap := parseWordBankClassV3Tree(root)
		// mapTupleMap map mapTableName to a list of MapTuple
		mapTupleMap := map[string][]*MapTuple{}
		for mapTableName, wbList := range mtMap {
			mapTupleList := []*MapTuple{}
			for _, wb := range wbList {
				mtList := wordbankV3ToMapTupleList(wb)
				mapTupleList = append(mapTupleList, mtList...)
			}
			mapTupleMap[mapTableName] = mapTupleList
		}
		appidToMapTupleMap[appid] = mapTupleMap
	}
	return appidToMapTupleMap
}

func wordbankV3ToMapTupleList(wb *dictionary.WordBankV3) []*MapTuple {
	mapTupleList := []*MapTuple{}
	for _, word := range wb.SimilarWords {
		mapTuple := MapTuple{Key: word, Value: wb.Name}
		mapTupleList = append(mapTupleList, &mapTuple)
	}
	return mapTupleList
}

// GetMapTableListV2 parse and generate mapping table name list from a WordBankClassV3 tree
func GetMapTableListV2(root *dictionary.WordBankClassV3) []string {
	// mtMap map mapTableName to a list of WordBankV3
	mtMap := parseWordBankClassV3Tree(root)

	mtList := make([]string, len(mtMap))
	i := 0
	for key := range mtMap {
		mtList[i] = key
		i++
	}
	logger.Trace.Printf("GetMapTableListV2: %+v", mtList)
	return mtList
}

func parseWordBankClassV3Tree(root *dictionary.WordBankClassV3) (retMtMap map[string][]*dictionary.WordBankV3) {
	teRoot := root.GetChildByName(util.Msg["TaskEngineWordbank"])
	if teRoot == nil {
		return
	}
	// mtMap map mapTableName to a list of WordBankV3
	mtMap := map[string][]*dictionary.WordBankV3{}
	for _, child := range teRoot.Children {
		tmpPath := make([]string, 0)
		dfsMapTableScan(child, tmpPath, mtMap)
	}
	return mtMap
}

func dfsMapTableScan(wbc *dictionary.WordBankClassV3, path []string, mtMap map[string][]*dictionary.WordBankV3) {
	newPath := make([]string, len(path))
	copy(newPath, path)
	newPath = append(newPath, wbc.Name)
	if len(wbc.Children) == 0 {
		if len(wbc.Wordbank) != 0 {
			mtName := strings.Join(newPath, "/")
			mtMap[mtName] = wbc.Wordbank
		}
	} else {
		for _, child := range wbc.Children {
			dfsMapTableScan(child, newPath, mtMap)
		}
	}
}

func GetMapTableContent(appid, userID, tableName string) (string, int, error) {
	content, err := getMapTableContent(appid, userID, tableName)
	if err == sql.ErrNoRows {
		return "", ApiError.NOT_FOUND_ERROR, err
	} else if err != nil {
		return "", ApiError.DB_ERROR, err
	}
	return content, ApiError.SUCCESS, nil
}

func SaveMappingTable(userID, appid, fileName, content string) (int, error) {
	err := saveMappingTable(userID, appid, fileName, content)
	if err != nil {
		return ApiError.DB_ERROR, err
	}
	return ApiError.SUCCESS, nil
}

func ParseUploadMappingTable(buf []byte) ([]*MapTuple, error) {
	// REMOVE utf-8 BOM if existed
	buf = bytes.Trim(buf, "\xef\xbb\xbf")

	buf = bytes.Replace(buf, []byte("\r\n"), []byte("\n"), -1)
	buf = bytes.Replace(buf, []byte("\r"), []byte("\n"), -1)

	lines := strings.Split(string(buf), "\n")

	ret := []*MapTuple{}
	for idx, line := range lines {
		params := strings.Split(line, ",")
		if len(params) != 2 {
			logger.Trace.Printf("Parse csv in line %d format error", idx+1)
			continue
		}

		temp := MapTuple{
			Key:   params[0],
			Value: params[1],
		}
		ret = append(ret, &temp)
	}
	return ret, nil
}

func DeleteMappingTable(appid, userID, tableName string) error {
	err := deleteMappingTable(appid, userID, tableName)
	if err == sql.ErrNoRows {
		return nil
	}
	return err
}

// UpdateSpreadsheetScenario updates task engine scenario
func UpdateSpreadsheetScenario(scenarioID, content, layout string) (int, error) {
	// TODO check duplicate scenario name
	err := updateSpreadsheetScenario(scenarioID, content, layout)
	if err != nil {
		return ApiError.DB_ERROR, nil
	}
	return ApiError.SUCCESS, nil
}

// UpdateIntentV1 register or update intent to intent engine 1.0
func UpdateIntentV1(appID string, intentName string, triggerPhrases []string) error {
	sentences := []*IntentSentenceV1{}
	for _, phrase := range triggerPhrases {
		sentence := IntentSentenceV1{
			Keywords: make([]string, 0),
			Sentence: phrase,
		}
		sentences = append(sentences, &sentence)
	}
	intentV1 := IntentV1{
		AppID:      appID,
		IntentID:   fmt.Sprintf("%s_%s", appID, intentName),
		IntentName: intentName,
		Sentences:  sentences,
	}
	dataString, err := json.Marshal([]IntentV1{intentV1})
	url := fmt.Sprintf("%s/intent/update", getEnvironment("INTENT_ENGINE_V1_URL"))
	params := map[string]string{
		"app_id": appID,
		"data":   string(dataString),
	}
	_, err = util.HTTPPostForm(url, params, 0)
	if err != nil {
		return err
	}
	return nil
}

// GetTaskEngineConfig return the task-engine config
func GetTaskEngineConfig() (*TEConfig, int, error) {
	configString, errno, err := util.ConsulGetTaskEngineConfig()
	if err != nil {
		return nil, errno, err
	}
	if configString == "" {
		configString = teDefaultConfig
	}
	teConfig := &TEConfig{}
	err = json.Unmarshal([]byte(configString), teConfig)
	if err != nil {
		return nil, ApiError.JSON_PARSE_ERROR, err
	}
	teConfigString, _ := json.Marshal(teConfig)
	logger.Trace.Printf("config of taskengine: %s", teConfigString)
	return teConfig, ApiError.SUCCESS, nil
}
