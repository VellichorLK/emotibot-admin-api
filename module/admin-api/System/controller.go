package System

import (
	"errors"
	"net/http"
	"strings"

	"emotibot.com/emotigo/module/admin-api/ApiError"
	"emotibot.com/emotigo/module/admin-api/util"
)

var (
	// ModuleInfo is needed for module define
	ModuleInfo util.ModuleInfo
)

func init() {
	ModuleInfo = util.ModuleInfo{
		ModuleName: "system",
		EntryPoints: []util.EntryPoint{
			util.NewEntryPoint("GET", "setting", []string{"view"}, handleGetSetting),
			util.NewEntryPoint("PUT", "setting", []string{"edit"}, handleUpdateSetting),
		},
	}
}

func handleGetSetting(w http.ResponseWriter, r *http.Request) {
	str, status, err := util.ConsulGetControllerSetting()
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(status, nil), http.StatusInternalServerError)
		return
	}
	util.LogTrace.Println("Get setting from consul: ", str)

	controllerSetting := ControllerSetting{}
	if str == "" {
		controllerSetting.SetDefault()
		util.WriteJSON(w, util.GenRetObj(ApiError.SUCCESS, controllerSetting))
		return
	}

	err = controllerSetting.UnMarshalJSON([]byte(str))
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.JSON_PARSE_ERROR, nil), http.StatusInternalServerError)
		return
	}
	util.WriteJSON(w, util.GenRetObj(ApiError.SUCCESS, controllerSetting))
}

func handleUpdateSetting(w http.ResponseWriter, r *http.Request) {
	status := http.StatusOK
	errno := ApiError.SUCCESS
	var err error
	var ret interface{}

	defer func() {
		if err != nil {
			util.WriteJSONWithStatus(w, util.GenRetObj(errno, err.Error()), status)
		} else {
			util.WriteJSON(w, util.GenRetObj(errno, ret))
		}
	}()

	r.ParseForm()
	settingStr := strings.TrimSpace(r.FormValue("setting"))
	if settingStr == "" {
		errno, err, status = ApiError.REQUEST_ERROR, errors.New("Invalid setting"), http.StatusBadRequest
		return
	}

	controllerSetting := ControllerSetting{}
	err = controllerSetting.UnMarshalJSON([]byte(settingStr))
	if err != nil {
		errno, status = ApiError.JSON_PARSE_ERROR, http.StatusBadRequest
		return
	}

	updatedStr, err := controllerSetting.MarshalJSON()
	if err != nil {
		errno, status = ApiError.JSON_PARSE_ERROR, http.StatusInternalServerError
		return
	}
	util.LogTrace.Println("Update controller setting with string: ", string(updatedStr))
	errno, err = util.ConsulSetControllerSetting(string(updatedStr))
	if err != nil {
		status = http.StatusInternalServerError
	}
}
