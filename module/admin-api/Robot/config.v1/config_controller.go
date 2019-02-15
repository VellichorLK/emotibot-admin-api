package config

import (
	"net/http"

	"emotibot.com/emotigo/module/admin-api/util/AdminErrors"

	"emotibot.com/emotigo/module/admin-api/util"

	"emotibot.com/emotigo/module/admin-api/util/requestheader"
)

func HandleGetRobotConfigs(w http.ResponseWriter, r *http.Request) {
	appid := requestheader.GetAppID(r)
	configs, err := GetConfigs(appid)
	util.Return(w, err, configs)
}

func HandleSetRobotConfig(w http.ResponseWriter, r *http.Request) {
	appid := requestheader.GetAppID(r)
	configName := r.FormValue("configName")
	module := r.FormValue("module")
	value := r.FormValue("value")

	if configName == "" {
		util.ReturnError(w, AdminErrors.ErrnoRequestError, "empty config name")
		return
	}

	if module == "" {
		util.ReturnError(w, AdminErrors.ErrnoRequestError, "empty module")
		return
	}

	var err AdminErrors.AdminError
	if value == "" {
		err = SetConfigToDefault(appid, configName)
	} else {
		err = SetConfig(appid, module, configName, value)
	}

	util.Return(w, err, err == nil)
	if module == moduleBFSource {
		go util.ConsulUpdateBFSetting(appid)
	} else {
		go util.ConsulUpdateBFOPSetting()
	}
}
