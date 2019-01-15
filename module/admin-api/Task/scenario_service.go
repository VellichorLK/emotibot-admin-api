package Task

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"emotibot.com/emotigo/module/admin-api/ApiError"
	"emotibot.com/emotigo/module/admin-api/autofill"
	autofillData "emotibot.com/emotigo/module/admin-api/autofill/data"
	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/pkg/logger"
	uuid "github.com/satori/go.uuid"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// GetScenarioInfoList get the scenario info list for the specified appid
func GetScenarioInfoList(appid string, userid string) ([]*ScenarioInfo, int, error) {
	scenarioInfoList, err := getScenarioInfoList(appid, userid)
	if err != nil {
		return nil, ApiError.DB_ERROR, err
	}
	return scenarioInfoList, ApiError.SUCCESS, nil
}

// GetTemplateScenarioInfoList get the template scenario info list
func GetTemplateScenarioInfoList() ([]*ScenarioInfo, int, error) {
	templateScenarioInfoList, err := getTemplateScenarioInfoList()
	if err != nil {
		return nil, ApiError.DB_ERROR, err
	}
	return templateScenarioInfoList, ApiError.SUCCESS, nil
}

// CreateScenario create a scenario record in taskenginescenario
func CreateScenario(scenarioid, userid, appid, content, layout string, public,
	editing int, editingContent, editingLayout string, onoff int) (int, error) {
	err := createScenario(scenarioid, userid, appid, content, layout, public, editing, editingContent, editingLayout, onoff)
	if err != nil {
		return ApiError.DB_ERROR, err
	}
	return ApiError.SUCCESS, nil
}

// CreateInitialScenario create an initial scenario record in taskenginescenario
func CreateInitialScenario(appid, userid, scenarioName string) (*ContentMetadata, int, error) {
	scenarioid, err := generateScenarioID()
	if err != nil {
		return nil, ApiError.DB_ERROR, err
	}
	// generate content json string
	now := time.Now()
	updateTime := now.Format("2006-01-02 15:04:05")
	newScenarioName, errno, err := suffixIndexToScenarioNameIfAlreadyExist(appid, userid, scenarioName)
	if err != nil {
		return nil, errno, err
	}
	metadata := &ContentMetadata{
		ScenarioName: newScenarioName,
		UpdateTime:   updateTime,
		UpdateUser:   userid,
		ScenarioID:   scenarioid,
	}
	content := InitialScenarioContent{
		Metadata: metadata,
	}
	contentBytes, err := json.Marshal(content)
	if err != nil {
		return nil, ApiError.JSON_PARSE_ERROR, err
	}
	contentString := string(contentBytes)

	// create scenario
	err = createScenario(scenarioid, userid, appid, contentString, "{}", 0, 0, contentString, "{}", 1)
	if err != nil {
		return nil, ApiError.DB_ERROR, err
	}
	return metadata, ApiError.SUCCESS, nil
}

// CreateScenarioWithTemplate create a new scenario that is a copy of the target template scenario
func CreateScenarioWithTemplate(appid, userid, scenarioName, template string) (*ContentMetadata, int, error) {
	scenarioid, err := generateScenarioID()
	if err != nil {
		return nil, ApiError.DB_ERROR, err
	}
	now := time.Now()
	updateTime := now.Format("2006-01-02 15:04:05")
	newScenarioName, errno, err := suffixIndexToScenarioNameIfAlreadyExist(appid, userid, scenarioName)
	if err != nil {
		return nil, errno, err
	}
	metadata := &ContentMetadata{
		ScenarioName: newScenarioName,
		UpdateTime:   updateTime,
		UpdateUser:   userid,
		ScenarioID:   scenarioid,
	}
	// get template scenario content
	scenario, err := getScenario(template)
	if err != nil {
		return nil, ApiError.DB_ERROR, err
	}
	content := &map[string]interface{}{}
	err = json.Unmarshal([]byte(scenario.Content), content)
	if err != nil {
		return nil, ApiError.JSON_PARSE_ERROR, err
	}
	(*content)["metadata"] = metadata
	contentBytes, err := json.Marshal(content)
	if err != nil {
		return nil, ApiError.JSON_PARSE_ERROR, err
	}
	contentString := string(contentBytes)
	// create scenario
	err = createScenario(scenarioid, userid, appid, contentString, scenario.Layout, 0, 0, contentString, scenario.Layout, 1)
	if err != nil {
		return nil, ApiError.DB_ERROR, err
	}
	return metadata, ApiError.SUCCESS, nil
}

// ImportScenarios import scenarios to the specified appid
func ImportScenarios(appid, userid string, useNewID bool, datas []interface{}) {
	for _, data := range datas {
		ImportScenario(appid, userid, useNewID, data)
	}
}

// ImportScenario import the scenario to the specified appid
func ImportScenario(appid, userid string, useNewID bool, jsonData interface{}) (int, error) {
	scenarioid, _, content, layout := parseJSONData(jsonData)
	if useNewID {
		id, err := generateScenarioID()
		if err != nil {
			return ApiError.DB_ERROR, err
		}
		scenarioid = id
		contentObj := content.(map[string]interface{})
		metadata := contentObj["metadata"].(map[string]interface{})
		metadata["scenario_id"] = id
	}

	// prepare contentString
	contentBytes, err := json.Marshal(content)
	if err != nil {
		return ApiError.JSON_PARSE_ERROR, err
	}
	contentString := string(contentBytes)
	// encrypt js_code if exist
	contentString, err = encryptJSCode(contentString)
	if err != nil {
		return ApiError.BASE64_PARSE_ERROR, err
	}

	// prepare layoutString
	layoutBytes, err := json.Marshal(layout)
	if err != nil {
		return ApiError.JSON_PARSE_ERROR, err
	}
	layoutString := string(layoutBytes)
	// create scenario
	err = createScenario(scenarioid, userid, appid, contentString, layoutString, 0, 0, contentString, layoutString, 1)
	if err != nil {
		return ApiError.DB_ERROR, err
	}
	return ApiError.SUCCESS, nil
}

func generateScenarioID() (string, error) {
	scenUUID, err := uuid.NewV4()
	if err != nil {
		logger.Error.Printf("Failed to generate uuid. %s\n", err)
		return "", err
	}
	return scenUUID.String(), nil
}

func suffixIndexToScenarioNameIfAlreadyExist(appid, userid, scenarioName string) (string, int, error) {
	scenarioInfoList, errno, err := GetScenarioInfoList(appid, appid)
	if err != nil {
		return "", errno, err
	}
	index := 0
	for _, scenarioInfo := range scenarioInfoList {
		if scenarioInfo.ScenarioName == scenarioName && index == 0 {
			index = 1
		} else {
			pattern := fmt.Sprintf(`^%s_(\d+)$`, scenarioName)
			var re = regexp.MustCompile(pattern)
			match := re.FindStringSubmatch(scenarioInfo.ScenarioName)
			if match != nil {
				i, err := strconv.Atoi(match[1])
				if err == nil && (i+1) > index {
					index = i + 1
				}
			}
		}
	}
	newScenarioName := scenarioName
	if index != 0 {
		newScenarioName = fmt.Sprintf("%s_%d", scenarioName, index)
	}
	return newScenarioName, ApiError.SUCCESS, nil
}

// GetScenario get the scenario content and layout
func GetScenario(scenarioid string) (*Scenario, int, error) {
	scenario, err := getScenario(scenarioid)
	if err != nil {
		return nil, ApiError.DB_ERROR, err
	}
	return scenario, ApiError.SUCCESS, nil
}

// GetDecryptedScenario get the scenario content and layout then decrypt the js_code
func GetDecryptedScenario(scenarioid string) (*Scenario, int, error) {
	scenario, err := getScenario(scenarioid)
	if err != nil {
		return nil, ApiError.DB_ERROR, err
	} else if scenario == nil {
		return scenario, ApiError.SUCCESS, nil
	}
	// check if text_type is cipher
	value := gjson.Get(scenario.EditingContent, "js_code.text_type")
	if !value.Exists() {
		logger.Trace.Printf("no js_code.text_type found, skip decryption process.")
		return scenario, ApiError.SUCCESS, nil
	}
	textType := value.String()
	if textType == "plain" {
		logger.Trace.Printf("text_type: plain, skip decryption process.")
		return scenario, ApiError.SUCCESS, nil
	}

	// decrypt js_code.main
	value = gjson.Get(scenario.EditingContent, "js_code.main")
	if !value.Exists() {
		logger.Trace.Printf("no js_code.main found, skip decryption process.")
		return scenario, ApiError.SUCCESS, nil
	}
	mainCipher := value.String()
	logger.Trace.Printf("mainCipher: %s", mainCipher)
	mainPlain, err := util.DesDecrypt(mainCipher, []byte(util.TEJSCodeEncryptKey))
	if err != nil {
		return nil, ApiError.BASE64_PARSE_ERROR, err
	}
	logger.Trace.Printf("mainPlain: %s", mainPlain)
	newContent, err := sjson.Set(scenario.EditingContent, "js_code.main", mainPlain)
	if err != nil {
		return nil, ApiError.JSON_PARSE_ERROR, err
	}
	newContent, err = sjson.Set(newContent, "js_code.text_type", "plain")
	if err != nil {
		return nil, ApiError.JSON_PARSE_ERROR, err
	}
	scenario.EditingContent = newContent
	logger.Trace.Printf("js_code.main decryption success")
	return scenario, ApiError.SUCCESS, nil
}

// UpdateScenario update the scenario editingContent and editingLayout
func UpdateScenario(scenarioid, appid, userid, editingContent, editingLayout string) (int, error) {
	// encrypt js_code if exist
	editingContent, err := encryptJSCode(editingContent)
	if err != nil {
		return ApiError.BASE64_PARSE_ERROR, err
	}
	err = updateScenario(scenarioid, appid, userid, editingContent, editingLayout)
	if err != nil {
		return ApiError.DB_ERROR, err
	}
	return ApiError.SUCCESS, nil
}

// UpdateTaskEngineIntents update the trigger intent list of a scenario
func UpdateTaskEngineIntents(appID, scenarioID string, intents *[]string) (int, error) {
	intentsInDB, err := getTaskEngineIntentList(appID, scenarioID)
	if err != nil {
		return ApiError.DB_ERROR, err
	}

	identical := compareStringSlice(*intents, intentsInDB)
	if identical {
		logger.Trace.Printf("Skip intent list update process for scenario: %s. Intent list is identical with the one in DB", scenarioID)
		return ApiError.SUCCESS, nil
	}
	logger.Trace.Printf("Found new intent list: %s, update intents for scenario: %s", *intents, scenarioID)

	// save new intent list
	err = saveIntents(appID, scenarioID, intents)
	if err != nil {
		return ApiError.DB_ERROR, err
	}
	// update autofills
	option := autofillData.AutofillOption{
		Module:   autofillData.AutofillModuleIntent,
		TaskMode: autofillData.SyncTaskModeUpdate,
	}
	autofill.UpdateAutofills(appID, &option)
	return ApiError.SUCCESS, nil
}

func compareStringSlice(a, b []string) bool {
	if (a == nil) != (b == nil) {
		return false
	}
	if len(a) != len(b) {
		return false
	}
	sort.Strings(a)
	sort.Strings(b)
	for i, s := range a {
		if s != b[i] {
			return false
		}
	}
	return true
}

// encryptJSCode encrypt the js_code.main in editingContent
func encryptJSCode(editingContent string) (string, error) {
	value := gjson.Get(editingContent, "js_code.text_type")
	if !value.Exists() {
		logger.Trace.Printf("no js_code.text_type found, skip encryption process.")
		return editingContent, nil
	}
	textType := value.String()
	logger.Trace.Printf("js_code.text_type: %s", textType)
	if textType == "cipher" {
		return editingContent, nil
	} else if textType == "plain" {
		value = gjson.Get(editingContent, "js_code.main")
		if !value.Exists() {
			logger.Trace.Printf("no js_code.main found, skip encryption process.")
			return editingContent, nil
		}
		logger.Trace.Printf("encrypt js_code.main")
		mainPlain := value.String()
		mainCipher, err := util.DesEncrypt([]byte(mainPlain), []byte(util.TEJSCodeEncryptKey))
		if err != nil {
			return "", err
		}
		editingContent, err = sjson.Set(editingContent, "js_code.main", mainCipher)
		if err != nil {
			return "", err
		}
		editingContent, err := sjson.Set(editingContent, "js_code.text_type", "cipher")
		if err != nil {
			return "", err
		}
		return editingContent, nil
	}
	return "", fmt.Errorf("unknown text_type in js_code: %s", textType)
}

// DeleteScenario delete the scenario with the specified scenarioid
func DeleteScenario(scenarioid, appid string) (int, error) {
	err := deleteScenario(scenarioid, appid)
	if err != nil {
		return ApiError.DB_ERROR, err
	}
	return ApiError.SUCCESS, nil
}

// GetAppScenarioList get the scenarioid list for the specified appid
func GetAppScenarioList(appid string) ([]string, int, error) {
	scenarioList, err := getAppScenarioList(appid)
	if err != nil {
		return nil, ApiError.DB_ERROR, err
	}
	return scenarioList, ApiError.SUCCESS, nil
}

// CreateAppScenario create an app-scenario pair to taskengineapp
func CreateAppScenario(scenarioid, appid string) (int, error) {
	err := createAppScenario(scenarioid, appid)
	if err != nil {
		return ApiError.DB_ERROR, err
	}
	return ApiError.SUCCESS, nil
}

// DeleteAppScenario delete the app-scenario pair with the specified scenarioid in taskengineapp
func DeleteAppScenario(scenarioid, appid string) (int, error) {
	err := deleteAppScenario(scenarioid, appid)
	if err != nil {
		return ApiError.DB_ERROR, err
	}
	return ApiError.SUCCESS, nil
}

// PublishScenario copy the scenario editingContent to content
func PublishScenario(scenarioid, appid, userid string) (int, error) {
	err := publishScenario(scenarioid, appid, userid)
	if err != nil {
		return ApiError.DB_ERROR, err
	}
	return ApiError.SUCCESS, nil
}

// UpdateAppScenarioPairToConsul update the app-scenario pair to consul key te/app
func UpdateAppScenarioPairToConsul(appid string) (int, error) {
	scenarioList, errno, err := GetAppScenarioList(appid)
	if err != nil {
		return errno, err
	}
	value := strings.Join(scenarioList, ",")
	errno, err = util.ConsulUpdateTaskEngineApp(appid, value)
	if err != nil {
		logger.Error.Printf("Failed to update consul key:te/app/%s errno: %d, %s", appid, errno, err.Error())
		return errno, err
	}
	return ApiError.SUCCESS, nil
}
