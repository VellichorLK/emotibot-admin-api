package Switch

import (
	"emotibot.com/emotigo/module/admin-api/ApiError"
)

func GetSwitches(appid string) ([]*SwitchInfo, int, error) {
	list, err := getSwitchList(appid)
	if err != nil {
		return nil, ApiError.DB_ERROR, err
	}

	return list, ApiError.SUCCESS, nil
}

func GetSwitch(appid string, id int) (*SwitchInfo, int, error) {
	ret, err := getSwitch(appid, id)
	if err != nil {
		return nil, ApiError.DB_ERROR, err
	}

	if ret == nil {
		return nil, ApiError.SWITCH_NO_ROWS, nil
	}

	return ret, ApiError.SUCCESS, nil
}

func UpdateSwitch(appid string, id int, input *SwitchInfo) (int, error) {
	err := updateSwitch(appid, id, input)
	if err != nil {
		return ApiError.DB_ERROR, err
	}

	return ApiError.SUCCESS, nil
}

func DeleteSwitch(appid string, id int) (int, error) {
	err := deleteSwitch(appid, id)
	if err != nil {
		return ApiError.DB_ERROR, err
	}

	return ApiError.SUCCESS, nil
}

func InsertSwitch(appid string, input *SwitchInfo) (int, error) {
	err := insertSwitch(appid, input)
	if err != nil {
		return ApiError.DB_ERROR, err
	}

	return ApiError.SUCCESS, nil
}
