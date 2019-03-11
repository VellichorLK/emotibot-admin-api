package controller

import (
	"fmt"
	"net/http"
	"strings"

	"emotibot.com/emotigo/module/token-auth/internal/enum"

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

func getRequesterV3(r *http.Request) *data.UserDetailV3 {
	authorization := r.Header.Get("Authorization")
	vals := strings.Split(authorization, " ")
	if len(vals) < 2 {
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
				return nil
			}
		}
		return &userInfo
	} else if strings.ToLower(vals[1]) == "api" {
		userInfo := data.UserDetailV3{}
		appid, enterprise, err := service.GetApiKeyOwner(vals[1])
		if err != nil {
			return nil
		}

		if appid != "" {
			userInfo.ID = fmt.Sprintf("%s API", appid)
			userInfo.Type = enum.NormalUser
		} else if enterprise != "" {
			userInfo.ID = fmt.Sprintf("%s API", enterprise)
			userInfo.Type = enum.AdminUser
		} else {
			userInfo.ID = "System API"
			userInfo.Type = enum.SuperAdminUser
		}
		userInfo.UserName = userInfo.ID
		userInfo.Status = 1

		return &userInfo
	}
	return nil
}
