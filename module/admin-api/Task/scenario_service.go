package Task

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"emotibot.com/emotigo/module/admin-api/ApiError"
	"emotibot.com/emotigo/pkg/logger"
	uuid "github.com/satori/go.uuid"
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
	// generate new scenarioid
	scenUUID, err := uuid.NewV4()
	if err != nil {
		logger.Error.Printf("Failed to generate uuid. %s\n", err)
		return nil, ApiError.DB_ERROR, err
	}
	scenarioid := scenUUID.String()
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

// UpdateScenario update the scenario editingContent and editingLayout
func UpdateScenario(scenarioid, appid, userid, editingContent, editingLayout string) (int, error) {
	err := updateScenario(scenarioid, appid, userid, editingContent, editingLayout)
	if err != nil {
		return ApiError.DB_ERROR, err
	}
	return ApiError.SUCCESS, nil
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
