package controller

import (
	"emotibot.com/emotigo/module/token-auth/service"
	"encoding/json"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"strconv"
)

func RolesGetHandlerHX(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	enterpriseID := vars["enterpriseID"]
	var retData, err = service.GetAllRoles(enterpriseID)
	if err != nil {
		returnInternalError(w, err.Error())
		return
	} else if retData == nil {
		returnNotFound(w)
		return
	}
	returnSuccess(w, retData)
}

func ModulesGetHandlerHX(w http.ResponseWriter, r *http.Request) {
	var retData, err = service.GetAllModules();
	if err != nil {
		returnInternalError(w, err.Error())
		return
	} else if retData == nil {
		returnNotFound(w)
		return
	}
	returnSuccess(w, retData)
}

func GetUserPrivilegesHandlerHx(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	var userCode = vars["userCode"]
	var enterpriseID = vars["enterpriseID"]
	var retData, err = service.GetUserPrivileges(enterpriseID,userCode);
	if err != nil {
		returnInternalError(w, err.Error())
		return
	} else if retData == nil {
		returnNotFound(w)
		return
	}
	returnSuccess(w, retData)
}

func GetRolePrivilegesHandlerHx(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var roleId,error= strconv.Atoi(vars["roleId"])
	if error != nil {
		returnInternalError(w, error.Error())
	}
	var enterpriseID = vars["enterpriseID"]
	var retData, err = service.GetRolePrivileges(enterpriseID,roleId);
	if err != nil {
		returnInternalError(w, err.Error())
		return
	} else if retData == nil {
		returnNotFound(w)
		return
	}
	returnSuccess(w, retData)
}

func PrivilegesUpdateHandlerHX(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var roleId,error= strconv.Atoi(vars["roleId"])
	var enterpriseID = vars["enterpriseID"]
	if error != nil {
		returnInternalError(w, error.Error())
		return
	}

	var req map[string]map[string]map[string][]string
	body, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(body, &req)

	var err = service.UpdateRolePrivileges(enterpriseID,roleId,req);
	if err != nil {
		returnInternalError(w, err.Error())
		return
	}
	returnSuccess(w, true)
}


func GetLabelUsersHandlerHX(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var enterpriseID = vars["enterpriseID"]
	var role,error= strconv.Atoi(vars["role"])
	if error != nil {
		returnInternalError(w, error.Error())
		return
	}
	var retData, err = service.GetLabelUsers(enterpriseID,role)
	if err != nil {
		returnInternalError(w, err.Error())
		return
	} else if retData == nil {
		returnNotFound(w)
		return
	}
	returnSuccess(w, retData)
}

func GetUserAccessInfoHandlerHX(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var userCode = vars["userCode"]
  	var retData, err = service.GetUserAccessInfo(userCode);
	if err != nil {
		returnInternalError(w, err.Error())
		return
	} else if retData == nil {
		returnNotFound(w)
		return
	}
	returnSuccess(w, retData)
}