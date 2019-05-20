package controller

import (
	"fmt"
	"net/http"
	"strings"

	"emotibot.com/emotigo/module/token-auth/internal/enum"
	"emotibot.com/emotigo/module/token-auth/internal/util"

	"emotibot.com/emotigo/module/token-auth/service"

	"emotibot.com/emotigo/module/token-auth/internal/data"
)

func getRequester(r *http.Request) *data.User {
	authorization := r.Header.Get("Authorization")
	vals := strings.Split(authorization, " ")
	if len(vals) < 2 {
		return nil
	}

	userInfo := data.User{}
	err := userInfo.SetValueWithToken(vals[1])
	if err != nil {
		return nil
	}

	return &userInfo
}

func GetRequesterV3(r *http.Request) *data.UserDetailV3 {
	authorization := r.Header.Get("Authorization")
	vals := strings.Split(authorization, " ")
	if len(vals) < 2 {
		util.LogError.Println("Invalid token:", authorization)
		return nil
	}

	if vals[0] == "Bearer" {
		userInfo := data.UserDetailV3{}
		if vals[1] == "EMOTIBOTDEBUGGER" {
			userInfo.ID = "System Trace"
			userInfo.Type = enum.SuperAdminUser
			userInfo.UserName = userInfo.ID
			userInfo.Status = 1
		} else {
			err := userInfo.SetValueWithToken(vals[1])
			if err != nil {
				util.LogError.Println("Cannot set user value with token:", err.Error())
				return nil
			}
		}
		return &userInfo
	} else if strings.ToLower(vals[0]) == "api" {
		userInfo := data.UserDetailV3{}
		appid, enterprise, err := service.GetApiKeyOwner(vals[1])
		if err != nil {
			util.LogError.Println("Get api owner fail:", err.Error())
			return nil
		}

		if appid != "" {
			userInfo.ID = fmt.Sprintf("%s API", appid)
			userInfo.Type = enum.NormalUser
			userInfo.Enterprise = &enterprise
		} else if enterprise != "" {
			userInfo.ID = fmt.Sprintf("%s API", enterprise)
			userInfo.Type = enum.AdminUser
			userInfo.Enterprise = &enterprise
		} else {
			userInfo.ID = "System API"
			userInfo.Type = enum.SuperAdminUser
		}
		userInfo.UserName = userInfo.ID
		userInfo.Status = 1

		return &userInfo
	}
	util.LogError.Println("Invalid token:", authorization)
	return nil
}
