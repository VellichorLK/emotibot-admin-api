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
	"emotibot.com/emotigo/module/admin-api/util"
	"github.com/tealeg/xlsx"
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

func ImportScenarios(appid string, newID bool, datas []interface{}) {
	for _, data := range datas {
		ImportScenario(appid, newID, data)
	}
}

func ImportScenario(appid string, newID bool, data interface{}) {
	// util.LogInfo.Printf("Import scenario, use new ID: %t \ndata: %#v\n", newID, data)
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
		"userid":       appid,
		"scenarioname": name,
	}
	url := fmt.Sprintf("%s/%s", getEnvironment("SERVER_URL"), taskScenarioEntry)
	retData, err := util.HTTPPostForm(url, postData, 0)
	if err != nil {
		util.LogTrace.Printf("Err: %s", err.Error())
		return "", err
	}
	retObj := map[string]interface{}{}
	err = json.Unmarshal([]byte(retData), &retObj)
	if err != nil {
		util.LogTrace.Printf("Err: %s", err.Error())
		return "", err
	}
	id := retObj["scenarioID"].(string)
	contentObj := content.(map[string]interface{})
	metadata := contentObj["metadata"].(map[string]interface{})
	util.LogInfo.Printf("metadata: %#v\n", metadata)
	metadata["scenario_id"] = id

	updateData := map[string]interface{}{
		"userid":  appid,
		"content": content,
		"layout":  layout,
	}
	url = fmt.Sprintf("%s/%s/%s",
		getEnvironment("SERVER_URL"),
		taskScenarioEntry,
		id)
	util.LogTrace.Printf("Call %s to addScenario\n", url)
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
		util.LogInfo.Printf("Check scenario %s existed fail: %s", id, err.Error())
		return false, err
	}

	ret := map[string]*interface{}{}
	err = json.Unmarshal([]byte(content), &ret)
	if err != nil {
		util.LogInfo.Printf("Check scenario %s existed fail: %s", id, err.Error())
		return false, err
	}

	if ret["result"] == nil {
		util.LogTrace.Printf("Check scenario: %s not exist\n", id)
		return false, nil
	}
	util.LogTrace.Printf("Check scenario: %s exist\n", id)
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
			util.LogTrace.Printf("Parse csv in line %d format error", idx+1)
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

// ParseUploadSpreadsheet will parse and verify uploaded spreadsheet
func ParseUploadSpreadsheet(scenarioString string, fileBuf []byte) ([]string, *Scenario, error) {
	var scenario Scenario
	var triggerPhrases []string
	json.Unmarshal([]byte(scenarioString), &scenario)
	xlFile, err := xlsx.OpenBinary(fileBuf)
	if err != nil {
		return nil, nil, err
	}

	// crreate an default intent named by scenario_name
	trigger := createDefaultTrigger(&scenario)
	triggerList := &scenario.EditingContent.Skills["mainSkill"].TriggerList
	*triggerList = append(*triggerList, trigger)

	var sheet *xlsx.Sheet
	// parse triggier phrases to return
	sheet = xlFile.Sheet[SheetName["triggerPhrase"]]
	if sheet != nil {
		triggerPhrases, err = parseTriggerPhrase(sheet)
		if err != nil {
			return nil, nil, err
		}
	}

	// parser entity and assign to scenario
	sheet = xlFile.Sheet[SheetName["entityCollecting"]]
	if sheet != nil {
		entityList, err := parseEntity(sheet)
		if err != nil {
			return nil, nil, err
		}
		scenario.EditingContent.Skills["mainSkill"].EntityCollectorList = entityList
	}

	sheet = xlFile.Sheet[SheetName["responseMessage"]]
	if sheet != nil {
		actionGroupList, err := parseMsgAction(sheet)
		if err != nil {
			return nil, nil, err
		}
		scenario.EditingContent.Skills["mainSkill"].ActionGroupList = actionGroupList
	}

	return triggerPhrases, &scenario, err
}

func parseTriggerPhrase(sheet *xlsx.Sheet) ([]string, error) {
	phrases := make([]string, 0)
	sheetPhrase := new(SpreadsheetTrigger)
	if sheet.MaxRow == 0 {
		return phrases, errors.New("Missing trigger phrases")
	}
	for i := 0; i < sheet.MaxRow; i++ {
		err := sheet.Rows[i].ReadStruct(sheetPhrase)
		if err != nil {
			return nil, err
		}
		phrases = append(phrases, sheetPhrase.Phrase)
	}
	return phrases, nil
}

func createDefaultTrigger(scenario *Scenario) *Trigger {
	intentName := scenario.EditingContent.Metadata["scenario_name"]
	trigger := Trigger{
		Type:       "intent_engine",
		IntentName: intentName,
		Editable:   true,
	}
	return &trigger
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

func parseEntity(sheet *xlsx.Sheet) ([]*Entity, error) {
	entities := make([]*Entity, 0)
	sheetEntity := new(SpreadsheetEntity)
	if sheet.MaxRow <= 1 {
		// skip entity collecting
		return nil, nil
	}
	for i := 1; i < sheet.MaxRow; i++ {
		err := sheet.Rows[i].ReadStruct(sheetEntity)
		if err != nil {
			return nil, err
		}
		entity := sheetEntity.ToEntity()
		entities = append(entities, &entity)
	}
	return entities, nil
}

// UpdateScenario updates task engine scenario
func UpdateScenario(scenarioID, content, layout string) (int, error) {
	// TODO check duplicate scenario name
	err := updateScenario(scenarioID, content, layout)
	if err != nil {
		return ApiError.DB_ERROR, nil
	}
	return ApiError.SUCCESS, nil
}

func parseMsgAction(sheet *xlsx.Sheet) ([]*ActionGroup, error) {
	actionGroupList := make([]*ActionGroup, 0)
	sheetMsgAction := new(SpreadsheetMsgAction)
	if sheet.MaxRow == 0 {
		return nil, errors.New("Missing response message")
	}
	for i := 0; i < sheet.MaxRow; i++ {
		err := sheet.Rows[i].ReadStruct(sheetMsgAction)
		if err != nil {
			return nil, err
		}
		action := Action{
			Type: "msg",
			Msg:  sheetMsgAction.Msg,
		}
		actionGroup := ActionGroup{
			ActionGroupID: util.GenRandomUUIDSameAsOpenAPI(),
			ActionList:    []*Action{&action},
			ConditionList: make([]*interface{}, 0),
		}

		actionGroupList = append(actionGroupList, &actionGroup)
	}
	return actionGroupList, nil
}
