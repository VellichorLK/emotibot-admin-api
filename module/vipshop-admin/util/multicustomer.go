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
	logTraceMC("req", reqURL)

	body, resErr := HTTPGetSimpleWithTimeout(reqURL, 5)
	if resErr != nil {
		logMCError(resErr)
		return ApiError.DICT_SERVICE_ERROR, resErr
	}
	logTraceMC("update wordbank", body)
	return ApiError.SUCCESS, nil
}

func McUpdateFunction(appid string) (int, error) {
	mcURL := getGlobalEnv(MulticustomerURLKey)
	// robot_config?
	// app_id, type=function
	// $curl = "curl '" . SETTING_API . $appid . "&type=function' >> /dev/null &";
	reqURL := fmt.Sprintf("%s/robot_config?app_id=%s&type=function", mcURL, appid)
	logTraceMC("req", reqURL)

	body, resErr := HTTPGetSimpleWithTimeout(reqURL, 5)
	if resErr != nil {
		logMCError(resErr)
		return ApiError.DICT_SERVICE_ERROR, resErr
	}
	logTraceMC("update function", body)
	return ApiError.SUCCESS, nil
}

func McRebuildRobotQA(appid string) (int, error) {
	mcURL := getGlobalEnv(MulticustomerURLKey)
	// manual_edit
	// app_id
	reqURL := fmt.Sprintf("%s/manual_edit?app_id=%s&type=robot", mcURL, appid)
	logTraceMC("req", reqURL)

	body, resErr := HTTPGetSimpleWithTimeout(reqURL, 5)
	if resErr != nil {
		logMCError(resErr)
		return ApiError.DICT_SERVICE_ERROR, resErr
	}
	logTraceMC("rebuild robotQA", body)
	return ApiError.SUCCESS, nil
}

func McManualBusiness(appid string) (int, error) {
	mcURL := getGlobalEnv(MulticustomerURLKey)
	// manual_edit
	// app_id
	reqURL := fmt.Sprintf("%s/manual_business?app_id=%s&type=robot", mcURL, appid)
	logTraceMC("req", reqURL)

	body, resErr := HTTPGetSimpleWithTimeout(reqURL, 5)
	if resErr != nil {
		logMCError(resErr)
		return ApiError.DICT_SERVICE_ERROR, resErr
	}
	logTraceMC("rebuild question", body)
	return ApiError.SUCCESS, nil
}

func logTraceMC(function string, msg string) {
	LogTrace.Printf("[MC][%s]:%s", function, msg)
}

func logMCError(err error) {
	logTraceMC("connect error", err.Error())
}
