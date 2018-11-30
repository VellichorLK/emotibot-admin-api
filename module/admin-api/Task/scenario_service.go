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
