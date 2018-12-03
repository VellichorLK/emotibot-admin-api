package Task

import (
	"emotibot.com/emotigo/module/admin-api/ApiError"
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
