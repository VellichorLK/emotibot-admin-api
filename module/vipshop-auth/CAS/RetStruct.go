package CAS

import (
	"encoding/json"
	"fmt"

	"emotibot.com/emotigo/module/vipshop-auth/CAuth"
)

type LoginRes struct {
	Appid        string `json:"appid,omitempty"`
	UsrID        string `json:"user_id,omitempty"`
	UsrName      string `json:"user_name,omitempty"`
	UsrType      int    `json:"user_type,omitempty"`
	EnterpriseID string `json:"enterprise_id,omitempty"`
	Privilege    string `json:"privilege,omitempty"`
	RoleName     string `json:"role_name,omitempty"`
}

type CASRetStruct struct {
	Code int `json:"code"`
}

type CaptchaRet struct {
	Data string `json:"data"`
	ID   string `json:"id"`
}

func getUserPrivs(userID string, pwd string) (*LoginRes, error) {
	sres, err := CAuth.GetUserRoles(userID)
	if err != nil {
		return nil, err
	}

	lr := &LoginRes{}
	lr.Appid = validAppID
	lr.UsrName = userID
	if len(sres) == 0 {
		return nil, fmt.Errorf("User %s has no role", userID)
	}

	for _, sre := range sres {
		rolesPriv, err := CAuth.GetRolePrivs(sre.RoleName)
		if err != nil {
			return nil, err
		}

		lr.RoleName = sre.RoleName

		rolesPrivString, err := json.Marshal(rolesPriv)
		if err != nil {
			return nil, err
		}

		lr.Privilege = string(rolesPrivString)

		//assume only one role
		break
	}

	return lr, nil
}
