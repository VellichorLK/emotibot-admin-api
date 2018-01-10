package util

import (
	"errors"
	"fmt"
	"time"

	"emotibot.com/emotigo/module/vipshop-admin/ApiError"
)

const (
	ConsulURLKey = "CONSUL_URL"
)

func ConsulUpdateTaskEngine(appid string, val bool) (int, error) {
	// key contains no appid, becaues this can be use in vipshop for now
	key := "te/enabled"
	return ConsulUpdateVal(key, val)
}

func ConsulUpdateRobotChat(appid string) (int, error) {
	key := fmt.Sprintf("%sdata/%s", appid, appid)
	now := time.Now().Unix()
	return ConsulUpdateVal(key, now)
}

func ConsulUpdateVal(key string, val interface{}) (int, error) {
	consulURL := getGlobalEnv(ConsulURLKey)
	if consulURL == "" {
		return ApiError.CONSUL_SERVICE_ERROR, errors.New("Consul URL unavailable")
	}

	reqURL := fmt.Sprintf("%s/%s", consulURL, key)
	logTraceConsul("update", reqURL)
	_, resErr := HTTPPut(reqURL, val, 5)
	if resErr != nil {
		logConsulError(resErr)
		return ApiError.CONSUL_SERVICE_ERROR, resErr
	}
	return ApiError.SUCCESS, nil
}

func logTraceConsul(function string, msg string) {
	LogTrace.Printf("[CONSUL][%s]:%s", function, msg)
}

func logConsulError(err error) {
	logTraceConsul("connect error", err.Error())
}
