package UI

import (
	"emotibot.com/emotigo/module/admin-api/util/AdminErrors"
)

func GetUIModules(appid string) ([]*Module, AdminErrors.AdminError) {
	ret, err := getUIModules(appid)
	if err != nil {
		return nil, AdminErrors.New(AdminErrors.ErrnoDBError, err.Error())
	}
	return ret, nil
}
