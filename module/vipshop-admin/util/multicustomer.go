package util

import (
	"fmt"

	"emotibot.com/emotigo/module/vipshop-admin/ApiError"
)

const (
	MulticustomerURLKey = "MC_URL"
)

func UpdateWordBank(appid string, userID string, userIP string, retFile string) (int, error) {
	mcURL := getGlobalEnv(MulticustomerURLKey)
	// http://172.16.101.47:14501/entity?
	// app_id, userip, userid, file_name
	reqURL := fmt.Sprintf("%s/entity?app_id=%s&userid=%s&userip=%s&file_name=%s", mcURL, appid, userID, userIP, retFile)
	LogTrace.Printf("mc req: %s", reqURL)

	_, resErr := HTTPGetSimpleWithTimeout(reqURL, 5)
	if resErr != nil {
		logMCError(resErr)
		return ApiError.DICT_SERVICE_ERROR, resErr
	}
	return ApiError.SUCCESS, nil
}

func McUpdateFunction(appid string) (int, error) {
	mcURL := getGlobalEnv(MulticustomerURLKey)
	// robot_config?
	// app_id, type=function
	// $curl = "curl '" . SETTING_API . $appid . "&type=function' >> /dev/null &";
	reqURL := fmt.Sprintf("%s/robot_config?app_id=%s&type=function", mcURL, appid)
	LogTrace.Printf("mc req: %s", reqURL)

	body, resErr := HTTPGetSimpleWithTimeout(reqURL, 5)
	if resErr != nil {
		logMCError(resErr)
		return ApiError.DICT_SERVICE_ERROR, resErr
	}
	LogTrace.Printf("%s", body)
	return ApiError.SUCCESS, nil
}

func logMCError(err error) {
	LogTrace.Printf("Link to multicustomer error: %s", err)
}
