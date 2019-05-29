package controller

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"emotibot.com/emotigo/module/token-auth/cache"
	"emotibot.com/emotigo/pkg/misc/adminerrors"
	"github.com/gorilla/mux"

	"emotibot.com/emotigo/module/token-auth/internal/audit"
	"emotibot.com/emotigo/module/token-auth/internal/data"
	"emotibot.com/emotigo/module/token-auth/internal/enum"
	"emotibot.com/emotigo/module/token-auth/internal/util"
	"emotibot.com/emotigo/module/token-auth/service"
)

var loginTemplate *template.Template
var authCache cache.Cache

const codeLength = 15
const codeExpire = 600

func init() {
	loginTemplate = template.Must(template.ParseFiles("template/login.html"))
	authCache = cache.NewLocalCache()
}

func GetOAuthLoginPage(w http.ResponseWriter, r *http.Request) {
	responseType := r.URL.Query().Get("response_type")
	clientID := r.URL.Query().Get("client_id")
	redirectURI := r.URL.Query().Get("redirect_uri")
	state := r.URL.Query().Get("state")

	if responseType != "code" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid response type"))
		return
	}

	if clientID == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid client ID"))
		return
	}

	if redirectURI == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid redirect URI"))
		return
	}

	if state == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid state"))
		return
	}

	cookie, _ := r.Cookie("oauth_token")
	if cookie != nil {
		util.LogTrace.Println("Use cookie,", cookie.Value)

		userInfo := data.UserDetailV3{}
		err := userInfo.SetValueWithToken(cookie.Value)
		if err == nil {
			code := util.GenRandomString(codeLength)
			authCache.Set("auth", code, &userInfo, 600)
			w.Header().Set("Location", fmt.Sprintf("%s?code=%s&state=%s", redirectURI, url.PathEscape(code), url.PathEscape(state)))
			w.WriteHeader(http.StatusMovedPermanently)
			return
		}
		util.LogTrace.Println("Invalid user token,", err.Error())
	}

	ret, err := service.CheckOauthLoginRequest(clientID, redirectURI)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	if !ret {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid client id"))
		return
	}

	loginTemplate.Execute(w, struct {
		Redirect string
		State    string
	}{
		Redirect: redirectURI,
		State:    state,
	})
}

func HandleOAuthLoginPage(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	account := r.Form.Get("account")
	passwd := r.Form.Get("passwd")

	var err error
	var user *data.UserDetailV3
	var code string

	redirectURI := r.Form.Get("redirect")
	state := r.Form.Get("state")

	defer func() {
		// Add audit log
		auditMessage := fmt.Sprintf("[%s]: %s", data.AuditLogin, account)

		enterpriseID := ""
		appID := ""
		userID := ""
		userIP := util.GetUserIP(r)
		module := audit.AuditModuleManageUser
		ret := 1
		if err != nil {
			ret = 0
		}

		if user != nil {
			if user.Enterprise != nil {
				// user is enterprise admin or normal user
				enterpriseID = *user.Enterprise
			} else {
				// user is system admin
				module = audit.AuditModuleManageAdmin
			}
		}

		audit.AddAuditLog(enterpriseID, appID, userID, userIP, module, audit.AuditOperationLogin,
			auditMessage, ret)

		if err != nil {
			w.Header().Set("Location", fmt.Sprintf("%s?err=%s&state=%s", redirectURI, err.Error(), url.PathEscape(state)))
		} else {
			w.Header().Set("Location", fmt.Sprintf("%s?code=%s&state=%s", redirectURI, url.PathEscape(code), url.PathEscape(state)))
		}
		w.WriteHeader(http.StatusMovedPermanently)
	}()

	if !util.IsValidString(&passwd) || !util.IsValidString(&account) {
		err = util.ErrInvalidParameter
		return
	}

	// If user is banned, return Forbidden
	if util.UserBanInfos.IsUserBanned(account) {
		err = util.ErrOperationForbidden
		return
	}

	passwd = fmt.Sprintf("%x", md5.Sum([]byte(passwd)))
	user, err = service.LoginV3(account, passwd)
	if err != nil {
		return
	} else if user == nil {
		// Login fail
		addUserTryCount(account)
		fmt.Printf("User %s login fail: %d\n", account, userTryCount[account])
		// Ban user if it's retry time more than 5
		if getUserTryCount(account) > banRetryTimes {
			util.UserBanInfos.BanUser(account)
			resetUserTryCount(account)
		}
		err = util.ErrOperationForbidden
		return
	}

	// Login success, clear ban info
	util.UserBanInfos.ClearBanInfo(account)

	code = util.GenRandomString(codeLength)
	authCache.Set("auth", code, user, 600)

	token, err := user.GenerateToken()
	expiration := time.Now().Add(time.Duration(util.GetJWTExpireTime()) * time.Second)
	cookie := http.Cookie{Name: "oauth_token", Value: token, Expires: expiration}
	http.SetCookie(w, &cookie)
}

func GetOAuthTokenViaCode(w http.ResponseWriter, r *http.Request) {
	grantType := r.URL.Query().Get("grant_type")
	code := r.URL.Query().Get("code")
	clientID := r.URL.Query().Get("client_id")
	clientSecret := r.URL.Query().Get("client_secret")
	redirectURI := r.URL.Query().Get("redirect_uri")

	if grantType != "authorization_code" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid grant type"))
		return
	}

	if clientID == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid client ID"))
		return
	}

	if redirectURI == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid redirect URI"))
		return
	}

	valid, err := service.CheckOauthRequest(clientID, clientSecret, redirectURI)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		// TODO
		return
	}
	if !valid {
		util.Return(w, adminerrors.New(adminerrors.ErrnoRequestError, "invalid oauth client info"), nil)
		return
	}

	userObj := authCache.Get("auth", code)
	if userObj == nil {
		util.Return(w, adminerrors.New(adminerrors.ErrnoRequestError, "invalid code"), nil)
		return
	}
	user, ok := userObj.(*data.UserDetailV3)
	if !ok {
		util.Return(w, adminerrors.New(adminerrors.ErrnoTypeConvert, "invalid type conversion"), nil)
		return
	}

	token, err := user.GenerateToken()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		util.Return(w, adminerrors.New(adminerrors.ErrnoRequestError, "invalid oauth client info"), nil)
		return
	}

	// code can only use for once
	authCache.Set("auth", code, nil, 0)
	ret := data.OauthTokenInto{
		AccessToken:  token,
		TokenType:    "Bearer",
		RefreshToken: "",
		ExpiresIn:    util.GetJWTExpireTime(),
		Scope:        "",
	}
	util.Return(w, nil, &ret)
}

func EnterpriseAddHandlerV4(w http.ResponseWriter, r *http.Request) {
	var enterprise *data.EnterpriseV3
	var err error

	defer func() {
		// Add audit log
		var auditMessage string
		if enterprise != nil {
			auditMessage = fmt.Sprintf("[%s]: %s", data.AuditContentEnterpriseAdd, enterprise.Name)
		} else {
			auditMessage = data.AuditContentEnterpriseAdd
		}

		addAuditLog(r, audit.AuditModuleManageEnterprise, audit.AuditOperationAdd,
			auditMessage, err)
	}()

	name := r.FormValue("name")
	if name == "" {
		util.ReturnError(w, adminerrors.ErrnoRequestError, "name")
		return
	}

	adminUser := r.FormValue("admin")
	if adminUser == "" {
		util.ReturnError(w, adminerrors.ErrnoRequestError, "admin")
		return
	}

	adminReq := data.EnterpriseAdminRequestV3{}
	err = json.Unmarshal([]byte(adminUser), &adminReq)
	if err != nil {
		util.ReturnError(w, adminerrors.ErrnoRequestError, "admin")
		return
	}

	status, err := strconv.Atoi(r.FormValue("status"))
	if err != nil {
		err = nil
		status = 0
	}

	enterpriseAdmin := data.UserDetailV3{
		UserV3: data.UserV3{
			UserName:    adminReq.Account,
			DisplayName: adminReq.Name,
			Email:       adminReq.Email,
			Type:        enum.AdminUser,
		},
		Password: &adminReq.Password,
	}
	temp, _ := json.Marshal(enterpriseAdmin)
	util.LogTrace.Println("Add admin user:", string(temp))

	description := r.FormValue("description")

	var modules []string
	if r.FormValue("modules") == "" {
		modules = []string{}
	} else {
		err = json.Unmarshal([]byte(r.FormValue("modules")), &modules)
		if err != nil {
			util.LogInfo.Println("Parse json fail: ", err.Error())
			util.ReturnError(w, adminerrors.ErrnoRequestError, "modules")
			return
		}
	}

	enterprise = &data.EnterpriseV3{
		Name:        name,
		Description: description,
	}

	id, err := service.AddEnterpriseV4(enterprise, modules, &enterpriseAdmin, false, status > 0)
	if err != nil {
		switch err {
		case util.ErrEnterpriseInfoExists:
			util.ReturnError(w, adminerrors.ErrnoRequestError, "name")
		case util.ErrUserEmailExists:
			util.ReturnError(w, adminerrors.ErrnoRequestError, "admin emails")
		case util.ErrUserNameExists:
			util.ReturnError(w, adminerrors.ErrnoRequestError, "admin username")
		default:
			util.ReturnError(w, adminerrors.ErrnoDBError, err.Error())
		}
		return
	}

	newEnterprise, err := service.GetEnterpriseV3(id)
	if err != nil || newEnterprise == nil {
		util.ReturnError(w, adminerrors.ErrnoDBError, err.Error())
		return
	}

	util.Return(w, nil, newEnterprise)
}

func EnterpriseTryAddHandlerV4(w http.ResponseWriter, r *http.Request) {
	var enterprise *data.EnterpriseV3
	var err error

	name := r.FormValue("name")
	if name == "" {
		util.ReturnError(w, adminerrors.ErrnoRequestError, "name")
		return
	}

	adminUser := r.FormValue("admin")
	if adminUser == "" {
		util.ReturnError(w, adminerrors.ErrnoRequestError, "admin")
		return
	}

	adminReq := data.EnterpriseAdminRequestV3{}
	err = json.Unmarshal([]byte(adminUser), &adminReq)
	if err != nil {
		util.ReturnError(w, adminerrors.ErrnoRequestError, "admin")
		return
	}

	enterpriseAdmin := data.UserDetailV3{
		UserV3: data.UserV3{
			UserName:    adminReq.Account,
			DisplayName: adminReq.Name,
			Email:       adminReq.Email,
			Type:        enum.AdminUser,
		},
		Password: &adminReq.Password,
	}
	temp, _ := json.Marshal(enterpriseAdmin)
	util.LogTrace.Println("Add admin user:", string(temp))

	description := r.FormValue("description")

	var modules []string
	if r.FormValue("modules") == "" {
		modules = []string{}
	} else {
		err = json.Unmarshal([]byte(r.FormValue("modules")), &modules)
		if err != nil {
			util.LogInfo.Println("Parse json fail: ", err.Error())
			util.ReturnError(w, adminerrors.ErrnoRequestError, "modules")
			return
		}
	}

	enterprise = &data.EnterpriseV3{
		Name:        name,
		Description: description,
	}

	_, err = service.AddEnterpriseV4(enterprise, modules, &enterpriseAdmin, true, false)
	if err != nil {
		switch err {
		case util.ErrEnterpriseInfoExists:
			util.ReturnError(w, adminerrors.ErrnoRequestError, "name")
		case util.ErrUserEmailExists:
			util.ReturnError(w, adminerrors.ErrnoRequestError, "admin emails")
		case util.ErrUserNameExists:
			util.ReturnError(w, adminerrors.ErrnoRequestError, "admin username")
		default:
			util.ReturnError(w, adminerrors.ErrnoDBError, err.Error())
		}
		return
	}

	util.Return(w, nil, nil)
}

func EnterpriseActivateHandlerV4(w http.ResponseWriter, r *http.Request) {
	updateEnterpriseStatus(w, r, true)
}
func EnterpriseDeactivateHandlerV4(w http.ResponseWriter, r *http.Request) {
	updateEnterpriseStatus(w, r, false)
}

func updateEnterpriseStatus(w http.ResponseWriter, r *http.Request, status bool) {
	vars := mux.Vars(r)

	var enterpriseID string
	var origEnterprise *data.EnterpriseDetailV3
	var err error

	defer func() {
		// Add audit log
		var auditMessage string
		if origEnterprise != nil {
			auditMessage = fmt.Sprintf("[%s]: %s", data.AuditContentEnterpriseUpdate, origEnterprise.Name)
		} else if enterpriseID != "" {
			auditMessage = fmt.Sprintf("[%s]: %s", data.AuditContentEnterpriseUpdate, enterpriseID)
		} else {
			auditMessage = data.AuditContentEnterpriseUpdate
		}

		addAuditLog(r, audit.AuditModuleManageRobot, audit.AuditOperationEdit,
			auditMessage, err)
	}()

	enterpriseID = vars["enterpriseID"]
	if !util.IsValidUUID(enterpriseID) {
		util.ReturnError(w, adminerrors.ErrnoRequestError, "enterprise id")
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	origEnterprise, err = service.GetEnterpriseV3(enterpriseID)
	if err != nil {
		util.ReturnError(w, adminerrors.ErrnoDBError, err.Error())
		return
	} else if origEnterprise == nil {
		util.ReturnError(w, adminerrors.ErrnoNotFound, "enterprise")
		return
	}

	if status {
		err = service.ActivateEnterpriseV4(enterpriseID, username, password)
	} else {
		err = service.UpdateEnterpriseStatusV4(enterpriseID, false)
	}
	if err != nil {
		util.ReturnError(w, adminerrors.ErrnoDBError, err.Error())
		return
	}

	util.Return(w, nil, true)
}

func UsersGetHandlerV4(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	enterpriseID := vars["enterpriseID"]
	if !util.IsValidUUID(enterpriseID) {
		util.ReturnError(w, adminerrors.ErrnoRequestError, "enterprise-id")
		return
	}

	retData, err := service.GetUsersV3(enterpriseID)
	if err != nil {
		util.ReturnError(w, adminerrors.ErrnoDBError, err.Error())
		return
	} else if retData == nil {
		util.ReturnError(w, adminerrors.ErrnoNotFound, err.Error())
		return
	}

	util.Return(w, nil, retData)
}

func UserGetHandlerV4(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	enterpriseID := vars["enterpriseID"]
	if !util.IsValidUUID(enterpriseID) {
		util.ReturnError(w, adminerrors.ErrnoRequestError, "enterprise-id")
		return
	}

	userID := vars["userID"]
	if !util.IsValidUUID(userID) {
		util.ReturnError(w, adminerrors.ErrnoRequestError, "user-id")
		return
	}

	retData, err := service.GetUserV3(enterpriseID, userID)
	if err != nil {
		util.ReturnError(w, adminerrors.ErrnoDBError, err.Error())
		return
	} else if retData == nil {
		util.ReturnError(w, adminerrors.ErrnoNotFound, err.Error())
		return
	}

	util.Return(w, nil, retData)
}
func AppAddHandlerV4(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var app *data.AppDetailV4
	var err error

	defer func() {
		// Add audit log
		var auditMessage string
		if app != nil {
			auditMessage = fmt.Sprintf("[%s]: %s", data.AuditContentAppAdd, app.Name)
		} else {
			auditMessage = data.AuditContentAppAdd
		}

		addAuditLog(r, audit.AuditModuleManageRobot, audit.AuditOperationAdd,
			auditMessage, err)
	}()

	enterpriseID := vars["enterpriseID"]
	if !util.IsValidUUID(enterpriseID) {
		err = util.ErrInvalidParameter
		returnBadRequest(w, "enterprise-id")
		return
	}

	app, err = parseAppFromRequestV4(r)
	if err != nil {
		returnBadRequest(w, err.Error())
		return
	}

	if app.Name == "" {
		err = util.ErrInvalidParameter
		returnBadRequest(w, "name")
		return
	}

	id, err := service.AddAppV4(enterpriseID, app)
	if err != nil {
		switch err {
		case util.ErrAppInfoExists:
			returnBadRequest(w, "name")
		case util.ErrOperationForbidden:
			returnForbiddenWithMsg(w, err.Error())
		default:
			returnInternalError(w, err.Error())
		}
		return
	} else if id == "" {
		returnBadRequest(w, "enterprise-id")
		return
	}

	newApp, err := service.GetAppV3(enterpriseID, id)
	if err != nil {
		returnInternalError(w, err.Error())
		return
	} else if newApp == nil {
		err = util.ErrInteralServer
		returnInternalError(w, err.Error())
		return
	}
	returnSuccess(w, newApp)
}
func parseAppFromRequestV4(r *http.Request) (*data.AppDetailV4, error) {
	name := strings.TrimSpace(r.FormValue("name"))
	description := r.FormValue("description")
	appType, _ := strconv.Atoi(r.FormValue("app_type"))

	ret := data.AppDetailV4{
		AppV4: data.AppV4{
			Name: name,
			AppType: appType,
		},
		Description: description,
	}

	return &ret, nil
}