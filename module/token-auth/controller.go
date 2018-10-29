package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"emotibot.com/emotigo/module/token-auth/internal/data"
	"emotibot.com/emotigo/module/token-auth/internal/enum"
	"emotibot.com/emotigo/module/token-auth/internal/util"
	"emotibot.com/emotigo/module/token-auth/service"
	"github.com/gorilla/mux"
	captcha "github.com/mojocn/base64Captcha"
)

var userTryCount map[string]int
var captchaConfig captcha.ConfigCharacter

const (
	banRetryTimes = 5
)

func init() {
	captchaConfig = initCaptchaConfig()
}

func EnterprisesGetHandler(w http.ResponseWriter, r *http.Request) {
	retData, err := service.GetEnterprises()

	var errMsg string
	if err != nil {
		errMsg = err.Error()
	} else {
		errMsg = ""
	}

	returnOKMsg(w, errMsg, retData)
}

func EnterpriseGetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	enterpriseID := vars["enterpriseID"]
	if !util.IsValidUUID(enterpriseID) {
		returnBadRequest(w, "enterpriseID")
		return
	}

	retData, err := service.GetEnterprise(enterpriseID)
	if err != nil {
		returnMsg(w, err.Error(), retData)
	} else {
		returnMsg(w, "", retData)
	}
}

func UsersGetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	enterpriseID := vars["enterpriseID"]
	if !util.IsValidUUID(enterpriseID) {
		returnBadRequest(w, "enterpriseID")
		return
	}

	retData, err := service.GetUsers(enterpriseID)
	if err != nil {
		returnInternalError(w, err.Error())
	} else {
		returnSuccess(w, retData)
	}
}

func UserGetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	enterpriseID := vars["enterpriseID"]
	if !util.IsValidUUID(enterpriseID) {
		returnBadRequest(w, "enterpriseID")
		return
	}

	userID := vars["userID"]
	if !util.IsValidUUID(userID) {
		returnBadRequest(w, "userID")
		return
	}

	retData, err := service.GetUser(enterpriseID, userID)
	if err != nil {
		returnInternalError(w, err.Error())
	} else {
		returnSuccess(w, retData)
	}
}

func AppsGetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	enterpriseID := vars["enterpriseID"]
	if !util.IsValidUUID(enterpriseID) {
		returnBadRequest(w, "enterpriseID")
		return
	}

	retData, errMsg := service.GetApps(enterpriseID)
	returnMsg(w, errMsg, retData)
}

func AppGetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	enterpriseID := vars["enterpriseID"]
	if !util.IsValidUUID(enterpriseID) {
		returnBadRequest(w, "enterpriseID")
		return
	}

	appID := vars["appID"]
	if !util.IsValidUUID(appID) {
		returnBadRequest(w, "appID")
		return
	}
	retData, errMsg := service.GetApp(enterpriseID, appID)
	returnMsg(w, errMsg, retData)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	account := r.Form.Get("account")
	passwd := r.Form.Get("passwd")
	if !util.IsValidString(&passwd) || !util.IsValidString(&account) {
		returnBadRequest(w, "")
		return
	}
	// if user is banned, return Forbidden
	if util.UserBanInfos.IsUserBanned(account) {
		returnForbidden(w)
		writeErrJSONWithObj(w, "forbidden", util.UserBanInfos[account])
		return
	}

	enterprise, user, errMsg := service.Login(account, passwd)
	if errMsg != "" {
		returnInternalError(w, errMsg)
		return
	} else if enterprise == nil && user == nil {
		// login fail
		addUserTryCount(account)
		fmt.Printf("User %s login fail: %d\n", account, userTryCount[account])
		// ban user if it's retry time more than 5
		if getUserTryCount(account) > banRetryTimes {
			util.UserBanInfos.BanUser(account)
			resetUserTryCount(account)
		}
		returnForbidden(w)
		writeErrJSONWithObj(w, "forbidden", util.UserBanInfos[account])
		return
	}
	// login success, clear ban info
	util.UserBanInfos.ClearBanInfo(account)

	token, err := user.GenerateToken()
	if err != nil {
		returnInternalError(w, err.Error())
		return
	}

	loginRet := data.LoginInfo{
		Token: token,
		Info:  user,
	}
	returnOKMsg(w, errMsg, loginRet)

}

func UserAddHandler(w http.ResponseWriter, r *http.Request) {
	requester := getRequester(r)
	vars := mux.Vars(r)
	enterpriseID := vars["enterpriseID"]
	if !util.IsValidUUID(enterpriseID) {
		returnBadRequest(w, "enterpriseID")
		return
	}
	user, err := parseAddUserFromRequest(r)
	if err != nil {
		returnBadRequest(w, err.Error())
		return
	}
	roleID := r.FormValue("role")
	if roleID == "" && user.Type == 2 {
		returnBadRequest(w, "Role")
		return
	}

	if requester.Type > user.Type {
		returnForbidden(w)
		return
	}

	if roleID != "" {
		role, err := service.GetRole(enterpriseID, roleID)
		if err != nil && err != sql.ErrNoRows {
			util.LogError.Printf("Error when get role %s: %s\n", roleID, err.Error())
			returnInternalError(w, err.Error())
			return
		} else if role == nil {
			returnBadRequest(w, "Role")
		}
	}

	id, err := service.AddUser(enterpriseID, user, roleID)
	if err != nil {
		returnInternalError(w, err.Error())
		return
	}
	newUser, err := service.GetUser(enterpriseID, id)
	if err != nil {
		returnInternalError(w, err.Error())
		return
	}
	returnSuccess(w, newUser)
}

func UserDeleteHandler(w http.ResponseWriter, r *http.Request) {
	requester := getRequester(r)
	vars := mux.Vars(r)
	enterpriseID := vars["enterpriseID"]
	if !util.IsValidUUID(enterpriseID) {
		returnBadRequest(w, "enterpriseID")
		return
	}
	userID := vars["userID"]
	if !util.IsValidUUID(userID) {
		returnBadRequest(w, "userID")
		return
	}

	user, err := service.GetUser(enterpriseID, userID)
	if err != nil && err != sql.ErrNoRows {
		returnInternalError(w, err.Error())
		return
	} else if user == nil {
		returnSuccess(w, "")
		return
	}

	if requester.Type > user.Type {
		returnForbidden(w)
		return
	}

	err = service.DeleteUser(enterpriseID, userID)
	if err != nil {
		returnInternalError(w, err.Error())
		return
	}

	returnSuccess(w, "")
}

func UserUpdateHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	enterpriseID := vars["enterpriseID"]
	if !util.IsValidUUID(enterpriseID) {
		returnBadRequest(w, "enterpriseID")
		return
	}

	userID := vars["userID"]
	if !util.IsValidUUID(userID) {
		returnBadRequest(w, "userID")
		return
	}

	origUser, err := service.GetUser(enterpriseID, userID)
	if err != nil {
		returnInternalError(w, err.Error())
		return
	} else if origUser == nil {
		returnNotFound(w)
		return
	}

	newUser, err := parseUpdateUserFromRequest(r)
	if err != nil {
		returnBadRequest(w, err.Error())
		return
	}
	newUser.Type = origUser.Type

	newUser.ID = userID
	newUser.Enterprise = &enterpriseID
	err = service.UpdateUser(enterpriseID, newUser)
	if err != nil {
		returnInternalError(w, err.Error())
		return
	}

	updatedUser, err := service.GetUser(enterpriseID, userID)
	if err != nil {
		returnInternalError(w, err.Error())
		return
	}
	returnSuccess(w, updatedUser)
}

func RolesGetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	enterpriseID := vars["enterpriseID"]
	if !util.IsValidUUID(enterpriseID) {
		returnBadRequest(w, "enterpriseID")
		return
	}

	retData, err := service.GetRoles(enterpriseID)
	if err != nil {
		returnInternalError(w, err.Error())
	} else {
		returnSuccess(w, retData)
	}
}
func RoleGetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	enterpriseID := vars["enterpriseID"]
	if !util.IsValidUUID(enterpriseID) {
		returnBadRequest(w, "enterpriseID")
		return
	}
	roleID := vars["roleID"]
	if !util.IsValidUUID(roleID) {
		returnBadRequest(w, "roleID")
		return
	}

	retData, err := service.GetRole(enterpriseID, roleID)
	if err != nil {
		returnInternalError(w, err.Error())
	} else {
		if retData == nil {
			returnNotFound(w)
		} else {
			returnSuccess(w, retData)
		}
	}
}
func RoleDeleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	enterpriseID := vars["enterpriseID"]
	if !util.IsValidUUID(enterpriseID) {
		returnBadRequest(w, "enterpriseID")
		return
	}
	roleID := vars["roleID"]
	if !util.IsValidUUID(roleID) {
		returnBadRequest(w, "roleID")
		return
	}

	retData, err := service.DeleteRole(enterpriseID, roleID)
	if err != nil {
		returnInternalError(w, err.Error())
	} else {
		returnSuccess(w, retData)
	}
}
func RoleAddHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	enterpriseID := vars["enterpriseID"]
	if !util.IsValidUUID(enterpriseID) {
		returnBadRequest(w, "enterpriseID")
		return
	}
	role, err := parseRoleFromRequest(r)
	if err != nil {
		returnBadRequest(w, err.Error())
		return
	}
	id, err := service.AddRole(enterpriseID, role)
	if err != nil {
		returnInternalError(w, err.Error())
		return
	}
	newRole, err := service.GetRole(enterpriseID, id)
	if err != nil {
		returnInternalError(w, err.Error())
		return
	}
	returnSuccess(w, newRole)
}
func RoleUpdateHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	enterpriseID := vars["enterpriseID"]
	if !util.IsValidUUID(enterpriseID) {
		returnBadRequest(w, "enterpriseID")
		return
	}
	roleID := vars["roleID"]
	if !util.IsValidUUID(roleID) {
		returnBadRequest(w, "roleID")
		return
	}
	role, err := parseRoleFromRequest(r)
	if err != nil {
		returnBadRequest(w, err.Error())
		return
	}
	ret, err := service.UpdateRole(enterpriseID, roleID, role)
	if err != nil {
		if err == sql.ErrNoRows {
			returnNotFound(w)
			return
		}
		returnInternalError(w, err.Error())
		return
	}
	returnSuccess(w, ret)
}

func ModulesGetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	enterpriseID := vars["enterpriseID"]
	if !util.IsValidUUID(enterpriseID) {
		returnBadRequest(w, "enterpriseID")
		return
	}

	retData, err := service.GetModules(enterpriseID)
	if err != nil {
		returnInternalError(w, err.Error())
	} else {
		returnSuccess(w, retData)
	}
}

func ValidateTokenHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	token := vars["token"]

	if token == "" {
		params := strings.Split(r.Header.Get("Authorization"), " ")
		if len(params) >= 2 {
			token = params[1]
		}
	}
	if token == "" {
		returnBadRequest(w, "token")
		return
	}

	userInfo := data.User{}
	err := userInfo.SetValueWithToken(token)
	if err != nil {
		util.LogInfo.Println("Check token fail: ", err.Error())
		returnBadRequest(w, "token")
		return
	}
	returnSuccess(w, nil)
}

func parseRoleFromRequest(r *http.Request) (*data.Role, error) {
	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		return nil, errors.New("Invalid name")
	}
	description := r.FormValue("description")
	privilegeStr := r.FormValue("privilege")

	privileges := map[string][]string{}
	err := json.Unmarshal([]byte(privilegeStr), &privileges)
	if err != nil {
		util.LogError.Printf("Cannot decode privilegeStr: %s\n", err.Error())
		return nil, err
	}
	ret := data.Role{
		Name:        name,
		Description: description,
		Privileges:  privileges,
	}
	return &ret, nil
}

func loadUserFromRequest(r *http.Request) *data.User {
	user := data.User{}
	username := r.FormValue("username")
	name := r.FormValue("name")
	email := r.FormValue("email")
	password := r.FormValue("password")
	role := r.FormValue("role")
	user.Email = &email
	user.DisplayName = &name
	user.Password = &password
	user.Role = &role
	user.UserName = &username

	userType, err := strconv.Atoi(r.FormValue("type"))
	if err != nil {
		userType = enum.NormalUser
	} else if userType > enum.NormalUser || userType < enum.AdminUser {
		userType = enum.NormalUser
	}
	user.Type = userType

	customStr := r.FormValue("custom")
	if customStr != "" {
		customInfo := map[string]string{}
		err = json.Unmarshal([]byte(customStr), &customInfo)
		if err == nil {
			user.CustomInfo = &customInfo
		} else {
			util.LogTrace.Println("Parse json error: ", err.Error())
		}
	}
	return &user
}
func parseAddUserFromRequest(r *http.Request) (*data.User, error) {
	user := loadUserFromRequest(r)

	// if user.Email == nil || *user.Email == "" {
	// 	return nil, errors.New("invalid email")
	// }
	if user.Password == nil || *user.Password == "" {
		return nil, errors.New("invalid password")
	}
	if user.UserName == nil || *user.UserName == "" {
		return nil, errors.New("invalid username")
	}

	return user, nil
}
func parseUpdateUserFromRequest(r *http.Request) (*data.User, error) {
	user := loadUserFromRequest(r)

	// if user.Email == nil || *user.Email == "" {
	// 	return nil, errors.New("invalid email")
	// }

	return user, nil
}

func returnMsg(w http.ResponseWriter, errMsg string, retData interface{}) {
	if reflect.ValueOf(retData).IsNil() && errMsg == "" {
		returnNotFound(w)
	} else {
		returnOKMsg(w, errMsg, retData)
	}
}

func returnOKMsg(w http.ResponseWriter, errMsg string, retData interface{}) {
	if errMsg != "" {
		writeErrJSON(w, errMsg)
	} else {
		returnSuccess(w, retData)
	}
}

func returnNotFound(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
	writeErrJSON(w, "Resource not found")
}

func returnBadRequest(w http.ResponseWriter, column string) {
	errMsg := ""
	w.WriteHeader(http.StatusBadRequest)
	if column != "" {
		errMsg = fmt.Sprintf("Column input error: %s", column)
	} else {
		errMsg = "Bad request"
	}
	writeErrJSON(w, errMsg)
}

func returnUnauthorized(w http.ResponseWriter) {
	http.Error(w, "Unauthorized", http.StatusUnauthorized)
}

func returnForbidden(w http.ResponseWriter) {
	http.Error(w, "", http.StatusForbidden)
}

func returnUnprocessableEntity(w http.ResponseWriter, errMsg string) {
	w.WriteHeader(http.StatusUnprocessableEntity)
	writeErrJSON(w, errMsg)
}

func returnInternalError(w http.ResponseWriter, errMsg string) {
	w.WriteHeader(http.StatusInternalServerError)
	writeErrJSON(w, errMsg)
}

func returnBFSuccess(w http.ResponseWriter, retData interface{}) {
	ret := data.BFReturn{
		ErrorCode: 0,
		ErrorMsg:  "success",
		Data:      &retData,
	}

	writeResponseJSON(w, &ret)
}

func returnSuccess(w http.ResponseWriter, retData interface{}) {
	ret := data.Return{
		ReturnMessage: "success",
		ReturnObj:     &retData,
	}

	writeResponseJSON(w, &ret)
}

func writeErrJSON(w http.ResponseWriter, errMsg string) {
	ret := data.Return{
		ReturnMessage: errMsg,
		ReturnObj:     nil,
	}
	writeResponseJSON(w, &ret)
}

func writeErrJSONWithObj(w http.ResponseWriter, errMsg string, obj interface{}) {
	ret := data.Return{
		ReturnMessage: errMsg,
		ReturnObj:     obj,
	}
	writeResponseJSON(w, &ret)
}

func writeResponseJSON(w http.ResponseWriter, ret interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ret)
}

func addUserTryCount(userID string) {
	if userTryCount == nil {
		userTryCount = make(map[string]int)
	}
	if _, ok := userTryCount[userID]; ok {
		userTryCount[userID]++
	} else {
		userTryCount[userID] = 1
	}
}

func getUserTryCount(userID string) int {
	if cnt, ok := userTryCount[userID]; ok {
		return cnt
	}
	return 0
}

func resetUserTryCount(userID string) {
	userTryCount[userID] = 0
}

func initCaptchaConfig() captcha.ConfigCharacter {
	return captcha.ConfigCharacter{
		Height:             60,
		Width:              265,
		Mode:               captcha.CaptchaModeNumber,
		ComplexOfNoiseText: captcha.CaptchaComplexMedium,
		ComplexOfNoiseDot:  captcha.CaptchaComplexMedium,
		IsUseSimpleFont:    true,
		IsShowHollowLine:   false,
		IsShowNoiseDot:     true,
		IsShowNoiseText:    false,
		IsShowSlimeLine:    true,
		IsShowSineLine:     false,
		CaptchaLen:         6,
	}
}

func CaptchaGetHandler(w http.ResponseWriter, r *http.Request) {
	captchaID, captcaInterfaceInstance := captcha.GenerateCaptcha("", captchaConfig)
	base64blob := captcha.CaptchaWriteToBase64Encoding(captcaInterfaceInstance)

	response := map[string]string{
		"data": base64blob,
		"id":   captchaID,
	}

	writeResponseJSON(w, response)
}
