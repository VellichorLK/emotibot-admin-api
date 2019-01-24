package controller

import (
	"crypto/md5"
	"fmt"
	"html/template"
	"net/http"
	"net/url"

	"emotibot.com/emotigo/module/admin-api/util/AdminErrors"

	"emotibot.com/emotigo/module/token-auth/cache"

	"emotibot.com/emotigo/module/token-auth/internal/audit"
	"emotibot.com/emotigo/module/token-auth/internal/data"
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
		util.Return(w, AdminErrors.New(AdminErrors.ErrnoRequestError, "invalid oauth client info"), nil)
		return
	}

	userObj := authCache.Get("auth", code)
	if userObj == nil {
		util.Return(w, AdminErrors.New(AdminErrors.ErrnoRequestError, "invalid code"), nil)
		return
	}
	user, ok := userObj.(*data.UserDetailV3)
	if !ok {
		util.Return(w, AdminErrors.New(AdminErrors.ErrnoTypeConvert, "invalid type conversion"), nil)
		return
	}

	token, err := user.GenerateToken()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		util.Return(w, AdminErrors.New(AdminErrors.ErrnoRequestError, "invalid oauth client info"), nil)
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
