package Task

import "emotibot.com/emotigo/module/admin-api/ApiError"

// GetScenarioInfoList get the scenario info list for the specified appid
func GetScenarioInfoList(appid string, userid string) ([]ScenarioInfo, int, error) {
	scenarioInfoList, err := getScenarioInfoList(appid, userid)
	if err != nil {
		return nil, ApiError.DB_ERROR, err
	}
	return scenarioInfoList, ApiError.SUCCESS, nil
}

// GetTemplateScenarioInfoList get the template scenario info list
func GetTemplateScenarioInfoList() ([]ScenarioInfo, int, error) {
	templateScenarioInfoList, err := getTemplateScenarioInfoList()
	if err != nil {
		return nil, ApiError.DB_ERROR, err
	}
	return templateScenarioInfoList, ApiError.SUCCESS, nil
}

// GetScenario get the scenario content for the specified scenarioid
func GetScenario(scenarioid string) (*Scenario, int, error) {
	scenario, err := getScenario(scenarioid)
	if err != nil {
		return nil, ApiError.DB_ERROR, err
	}
	return scenario, ApiError.SUCCESS, nil
}

// UpdateScenario update the scenario content for the specified scenarioid
func UpdateScenario(scenarioid, appid, userid, editingContent, editingLayout string) (int, error) {
	err := updateScenario(scenarioid, appid, userid, editingContent, editingLayout)
	if err != nil {
		return ApiError.DB_ERROR, err
	}
	return ApiError.SUCCESS, nil
}
