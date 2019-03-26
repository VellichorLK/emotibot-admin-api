package main

import (
	"emotibot.com/emotigo/module/token-auth/internal/enum"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"emotibot.com/emotigo/module/token-auth/controller"
	"emotibot.com/emotigo/module/token-auth/dao"
	"emotibot.com/emotigo/module/token-auth/internal/audit"
	"emotibot.com/emotigo/module/token-auth/internal/data"
	"emotibot.com/emotigo/module/token-auth/internal/util"
	"emotibot.com/emotigo/module/token-auth/service"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

const (
	prefixURL = "/auth"
)

// Route define the end point of server
type Route struct {
	Name        string
	Method      string
	Version     int
	Pattern     string
	Queries     []string
	HandlerFunc http.HandlerFunc

	// 0 means super admin can use this API
	// 1 means enterprise admin can use this API
	// 2 means user in enterprise can use this API
	GrantType []interface{}
}

type Routes []Route

var routes Routes

func setUpRoutes() {
	routes = Routes{
		// v2 API
		Route{"GetEnterprises", "GET", 2, "enterprises", nil, controller.EnterprisesGetHandler, []interface{}{0}},
		Route{"GetEnterprise", "GET", 2, "enterprise/{enterpriseID}", nil, controller.EnterpriseGetHandler, []interface{}{0, 1, 2}},
		Route{"GetUsers", "GET", 2, "enterprise/{enterpriseID}/users", nil, controller.UsersGetHandler, []interface{}{0, 1, 2}},
		Route{"GetUser", "GET", 2, "enterprise/{enterpriseID}/user/{userID}", nil, controller.UserGetHandler, []interface{}{0, 1, 2}},
		Route{"GetApps", "GET", 2, "enterprise/{enterpriseID}/apps", nil, controller.AppsGetHandler, []interface{}{0, 1, 2}},
		Route{"GetApp", "GET", 2, "enterprise/{enterpriseID}/app/{appID}", nil, controller.AppGetHandler, []interface{}{0, 1, 2}},
		Route{"Login", "POST", 2, "login", nil, controller.LoginHandler, []interface{}{}},
		Route{"ValidateToken", "GET", 2, "token/{token}", nil, controller.ValidateTokenHandler, []interface{}{}},
		Route{"ValidateToken", "GET", 2, "token", nil, controller.ValidateTokenHandler, []interface{}{}},

		Route{"AddUser", "POST", 2, "enterprise/{enterpriseID}/user", nil, controller.UserAddHandler, []interface{}{0, 1, 2}},
		Route{"UpdateUser", "PUT", 2, "enterprise/{enterpriseID}/user/{userID}", nil, controller.UserUpdateHandler, []interface{}{0, 1, 2}},
		Route{"DeleteUser", "DELETE", 2, "enterprise/{enterpriseID}/user/{userID}", nil, controller.UserDeleteHandler, []interface{}{0, 1, 2}},

		Route{"GetRoles", "GET", 2, "enterprise/{enterpriseID}/roles", nil, controller.RolesGetHandler, []interface{}{0, 1, 2}},
		Route{"GetRole", "GET", 2, "enterprise/{enterpriseID}/role/{roleID}", nil, controller.RoleGetHandler, []interface{}{0, 1, 2}},
		Route{"AddRole", "POST", 2, "enterprise/{enterpriseID}/role", nil, controller.RoleAddHandler, []interface{}{0, 1, 2}},
		Route{"UpdateRole", "PUT", 2, "enterprise/{enterpriseID}/role/{roleID}", nil, controller.RoleUpdateHandler, []interface{}{0, 1, 2}},
		Route{"DeleteRole", "DELETE", 2, "enterprise/{enterpriseID}/role/{roleID}", nil, controller.RoleDeleteHandler, []interface{}{0, 1, 2}},
		Route{"GetModules", "GET", 2, "enterprise/{enterpriseID}/modules", nil, controller.ModulesGetHandler, []interface{}{0, 1, 2}},

		// v3 API
		Route{"GetSystemAdmins", "GET", 3, "admins", nil, controller.SystemAdminsGetHandlerV3, []interface{}{0}},
		Route{"GetSystemAdmin", "GET", 3, "admin/{adminID}", nil, controller.SystemAdminGetHandlerV3, []interface{}{0}},
		Route{"AddSystemAdmin", "POST", 3, "admin", nil, controller.SystemAdminAddHandlerV3, []interface{}{0}},
		Route{"UpdateSystemAdmin", "PUT", 3, "admin/{adminID}", nil, controller.SystemAdminUpdateHandlerV3, []interface{}{0}},
		Route{"DeleteSystemAdmin", "DELETE", 3, "admin/{adminID}", nil, controller.SystemAdminDeleteHandlerV3, []interface{}{0}},

		Route{"GetEnterprises", "GET", 3, "enterprises", nil, controller.EnterprisesGetHandlerV3, []interface{}{0, 1, 2}},
		Route{"GetEnterprise", "GET", 3, "enterprise/{enterpriseID}", nil, controller.EnterpriseGetHandlerV3, []interface{}{0, 1, 2}},
		Route{"AddEnterprise", "POST", 3, "enterprise", nil, controller.EnterpriseAddHandlerV3, []interface{}{0}},
		Route{"UpdateEnterprise", "PUT", 3, "enterprise/{enterpriseID}", nil, controller.EnterpriseUpdateHandlerV3, []interface{}{0, 1}},
		Route{"DeleteEnterprise", "DELETE", 3, "enterprise/{enterpriseID}", nil, controller.EnterpriseDeleteHandlerV3, []interface{}{0}},
		Route{"GetEnterpriseSecretKey", "GET", 3, "enterprise/{enterpriseID}/secret", nil, controller.EnterpriseGetSecretHandlerV3, []interface{}{0, 1, 2}},
		Route{"GetEnterpriseSecretKey", "POST", 3, "enterprise/{enterpriseID}/secret", nil, controller.EnterpriseRenewSecretHandlerV3, []interface{}{0, 1, 2}},

		Route{"GetUsers", "GET", 3, "enterprise/{enterpriseID}/users", nil, controller.UsersGetHandlerV3, []interface{}{0, 1, 2}},
		Route{"GetUser", "GET", 3, "enterprise/{enterpriseID}/user/{userID}", nil, controller.UserGetHandlerV3, []interface{}{0, 1, 2}},
		Route{"AddUser", "POST", 3, "enterprise/{enterpriseID}/user", nil, controller.UserAddHandlerV3, []interface{}{0, 1}},
		Route{"UpdateUser", "PUT", 3, "enterprise/{enterpriseID}/user/{userID}", nil, controller.UserUpdateHandlerV3, []interface{}{0, 1, 2}},
		Route{"DeleteUser", "DELETE", 3, "enterprise/{enterpriseID}/user/{userID}", nil, controller.UserDeleteHandlerV3, []interface{}{0, 1}},

		Route{"GetApps", "GET", 3, "enterprise/{enterpriseID}/apps", nil, controller.AppsGetHandlerV3, []interface{}{0, 1, 2}},
		Route{"GetApp", "GET", 3, "enterprise/{enterpriseID}/app/{appID}", nil, controller.AppGetHandlerV3, []interface{}{0, 1, 2}},
		Route{"AddApp", "POST", 3, "enterprise/{enterpriseID}/app", nil, controller.AppAddHandlerV3, []interface{}{0, 1}},
		Route{"UpdateApp", "PUT", 3, "enterprise/{enterpriseID}/app/{appID}", nil, controller.AppUpdateHandlerV3, []interface{}{0, 1}},
		Route{"DeleteApp", "DELETE", 3, "enterprise/{enterpriseID}/app/{appID}", nil, controller.AppDeleteHandlerV3, []interface{}{0, 1}},
		Route{"GetAppSecretKey", "GET", 3, "enterprise/{enterpriseID}/app/{appID}/secret", nil, controller.AppGetSecretHandlerV3, []interface{}{0, 1, 2}},
		Route{"GetAppSecretKey", "POST", 3, "enterprise/{enterpriseID}/app/{appID}/secret", nil, controller.AppRenewSecretHandlerV3, []interface{}{0, 1, 2}},

		Route{"GetGroups", "GET", 3, "enterprise/{enterpriseID}/groups", nil, controller.GroupsGetHandlerV3, []interface{}{0, 1, 2}},
		Route{"GetGroup", "GET", 3, "enterprise/{enterpriseID}/group/{groupID}", nil, controller.GroupGetHandlerV3, []interface{}{0, 1, 2}},
		Route{"AddGroup", "POST", 3, "enterprise/{enterpriseID}/group", nil, controller.GroupAddHandlerV3, []interface{}{0, 1}},
		Route{"UpdateGroup", "PUT", 3, "enterprise/{enterpriseID}/group/{groupID}", nil, controller.GroupUpdateHandlerV3, []interface{}{0, 1}},
		Route{"DeleteGroup", "DELETE", 3, "enterprise/{enterpriseID}/group/{groupID}", nil, controller.GroupDeleteHandlerV3, []interface{}{0, 1}},

		Route{"GetRoles", "GET", 3, "enterprise/{enterpriseID}/roles", nil, controller.RolesGetHandlerV3, []interface{}{0, 1, 2}},
		Route{"GetRole", "GET", 3, "enterprise/{enterpriseID}/role/{roleID}", nil, controller.RoleGetHandlerV3, []interface{}{0, 1, 2}},
		Route{"AddRole", "POST", 3, "enterprise/{enterpriseID}/role", nil, controller.RoleAddHandlerV3, []interface{}{0, 1}},
		Route{"UpdateRole", "PUT", 3, "enterprise/{enterpriseID}/role/{roleID}", nil, controller.RoleUpdateHandlerV3, []interface{}{0, 1}},
		Route{"DeleteRole", "DELETE", 3, "enterprise/{enterpriseID}/role/{roleID}", nil, controller.RoleDeleteHandlerV3, []interface{}{0, 1}},

		Route{"Login", "POST", 3, "login", nil, controller.LoginHandlerV3, []interface{}{}},
		Route{"ValidateToken", "GET", 3, "token/{token}", nil, controller.ValidateTokenHandlerV3, []interface{}{}},
		Route{"ValidateToken", "GET", 3, "token", nil, controller.ValidateTokenHandlerV3, []interface{}{}},
		Route{"IssueApiKey", "POST", 3, "apikey/issue", nil, controller.IssueApiKeyHandler, []interface{}{}},
		Route{"ValidateApiKey", "GET", 3, "apikey", nil, controller.ValidateApiKey, []interface{}{}},

		Route{"GetModules", "GET", 3, "enterprise/{enterpriseID}/modules", nil, controller.ModulesGetHandlerV3, []interface{}{0, 1, 2}},
		Route{"GetModules", "GET", 3, "modules", nil, controller.GlobalModulesGetHandlerV3, []interface{}{}},

		Route{"GetEnterpriseId", "GET", 3, "getEnterpriseId", []string{"app-id", "{app-id}"}, controller.EnterpriseIDGetHandlerV3, []interface{}{}},
		Route{"GetUserBelong", "GET", 3, "user/{userID}/info", nil, controller.UserInfoGetHandler, []interface{}{0, 1, 2}},
		Route{"GetAllEnterpriseRobot", "GET", 3, "enterpriselist", nil, controller.EnterpriseAppGetHandlerV3, []interface{}{0, 1, 2}},

		Route{"AddIMUser", "POST", 3, "enterprise/{enterpriseID}/imuser", nil, controller.IMUserAddHandlerV3, []interface{}{0, 1}},
		Route{"UpdateIMUser", "PUT", 3, "enterprise/{enterpriseID}/imuser/{userID}", nil, controller.IMUserUpdateHandlerV3, []interface{}{0, 1, 2}},
		Route{"DeleteIMUser", "DELETE", 3, "enterprise/{enterpriseID}/imuser/{userID}", nil, controller.IMUserDeleteHandlerV3, []interface{}{0, 1, 2}},
		Route{"ValidateIMToken", "GET", 3, "imtoken", nil, controller.IMValidateTokenHandler, []interface{}{}},
		Route{"GetIMRobots", "GET", 3, "enterprise/{enterpriseID}/imrobots", nil, controller.IMAppsGetHandlerV3, []interface{}{0, 1, 2}},

		Route{"GetCaptcha", "GET", 1, "captcha", nil, controller.CaptchaGetHandler, []interface{}{}},

		Route{"ValidateToken", "GET", 3, "trace/token/{token}", nil, controller.TraceValidateTokenHandlerV3, []interface{}{}},
		Route{"ValidateToken", "GET", 3, "trace/token", nil, controller.TraceValidateTokenHandlerV3, []interface{}{}},

		Route{"GetOauthLogin", "GET", 4, "oauth/login", nil, controller.GetOAuthLoginPage, []interface{}{}},
		Route{"DoOauthLogin", "POST", 4, "oauth/login", nil, controller.HandleOAuthLoginPage, []interface{}{}},
		Route{"GetUserTokenFromCode", "GET", 4, "oauth/token", nil, controller.GetOAuthTokenViaCode, []interface{}{}},

		Route{"AddEnterprise", "POST", 4, "enterprise", nil, controller.EnterpriseAddHandlerV4, []interface{}{0}},
		Route{"TrAddEnterprise", "POST", 4, "enterprise/try-create", nil, controller.EnterpriseTryAddHandlerV4, []interface{}{0}},
		Route{"ActivateEnterprise", "POST", 4, "enterprise/{enterpriseID}/active", nil, controller.EnterpriseActivateHandlerV4, []interface{}{0}},
		Route{"DeactivateEnterprise", "POST", 4, "enterprise/{enterpriseID}/deactive", nil, controller.EnterpriseDeactivateHandlerV4, []interface{}{0}},

		//华夏 API
		Route{"GetRolesHX", "GET", 4, "enterprise/{enterpriseID}/roles", nil, controller.RolesGetHandlerHX, []interface{}{0, 1, 2}},
		Route{"GetModulesHX", "GET", 4, "modules", nil, controller.ModulesGetHandlerHX, []interface{}{0, 1, 2}},
		Route{"GetRolePrivileges", "GET", 4, "enterprise/{enterpriseID}/{roleId}/privileges", nil, controller.PrivilegesGetHandlerHX, []interface{}{0, 1, 2}},
		Route{"UpdateRolePrivileges", "POST", 4, "enterprise/{enterpriseID}/{roleId}/privileges", nil, controller.PrivilegesUpdateHandlerHX, []interface{}{0, 1, 2}},

	}
}

func setUpDB() {
	db := dao.MYSQLController{}

	url, port, user, passwd, dbName := util.GetMySQLConfig()
	util.LogInfo.Printf("Init mysql: %s:%s@%s:%d/%s\n", user, passwd, url, port, dbName)
	db.InitDB(url, port, dbName, user, passwd)
	service.SetDB(&db)
	service.SetDBV3(&db)
	service.SetDBV4(&db)
	service.SetDBHX(&db)

	url, port, user, passwd, dbName = util.GetAuditMySQLConfig()
	util.LogInfo.Printf("Init audit mysql: %s:%s@%s:%d/%s\n", user, passwd, url, port, dbName)
	db.InitAuditDB(url, port, dbName, user, passwd)
	audit.SetDB(&db)

	url, port, user, passwd, dbName = util.GetBFMySQLConfig()
	util.LogInfo.Printf("Init bf mysql: %s:%s@%s:%d/%s\n", user, passwd, url, port, dbName)
	db.InitBFDB(url, port, dbName, user, passwd)
}

func checkAuth(r *http.Request, route Route) bool {
	util.LogInfo.Printf("Access: %s %s", r.Method, r.RequestURI)
	if len(route.GrantType) == 0 {
		util.LogTrace.Println("[Auth check] pass: no need")
		return true
	}

	authorization := r.Header.Get("Authorization")
	if authorization == "Bearer EMOTIBOTDEBUGGER" {
		return true
	}

	vals := strings.Split(authorization, " ")
	if len(vals) < 2 {
		util.LogError.Println("[Auth check] Auth fail: no header")
		return false
	}

	switch route.Version {
	case 2:
		userInfo := data.User{}
		err := userInfo.SetValueWithToken(vals[1])
		if err != nil {
			util.LogInfo.Printf("[Auth check] Auth fail: no valid token [%s]\n", err.Error())
			return false
		}

		if !util.IsInSlice(userInfo.Type, route.GrantType) {
			util.LogInfo.Printf("[Auth check] Need user be [%v], get [%d]\n", route.GrantType, userInfo.Type)
			return false
		}

		vars := mux.Vars(r)
		// Enterprise admin user can only check enterprise of itself
		// Enterprise normal can only check enterprise of itself and user info of itself
		if userInfo.Type == enum.AdminUser || userInfo.Type == enum.NormalUser {
			if userInfo.Enterprise == nil {
				return false
			}

			enterpriseID := vars["enterpriseID"]
			if enterpriseID != "" && enterpriseID != *userInfo.Enterprise {
				util.LogInfo.Printf("[Auth check] user of [%s] can not access [%s]\n", *userInfo.Enterprise, enterpriseID)
				return false
			}
		}

		if userInfo.Type == enum.NormalUser {
			userID := vars["userID"]
			if userID != "" && userID != userInfo.ID {
				util.LogInfo.Printf("[Auth check] user [%s] can not access other users' info\n", userInfo.ID)
				return false
			}
		}
	case 3:
		userInfo := data.UserDetailV3{}
		err := userInfo.SetValueWithToken(vals[1])
		if err != nil {
			util.LogInfo.Printf("[Auth check] Auth fail: no valid token [%s]\n", err.Error())
			return false
		}

		if !util.IsInSlice(userInfo.Type, route.GrantType) {
			util.LogInfo.Printf("[Auth check] Need user be [%v], get [%d]\n", route.GrantType, userInfo.Type)
			return false
		}

		vars := mux.Vars(r)
		// Enterprise admin user can only check enterprise of itself
		// Enterprise normal can only check enterprise of itself and user info of itself
		if userInfo.Type == enum.AdminUser || userInfo.Type == enum.NormalUser {
			if userInfo.Enterprise == nil {
				return false
			}

			enterpriseID := vars["enterpriseID"]
			if enterpriseID != "" && enterpriseID != *userInfo.Enterprise {
				util.LogInfo.Printf("[Auth check] user of [%s] can not access [%s]\n", *userInfo.Enterprise, enterpriseID)
				return false
			}
		}

		if userInfo.Type == enum.NormalUser {
			userID := vars["userID"]
			if userID != "" && userID != userInfo.ID {
				util.LogInfo.Printf("[Auth check] user [%s] can not access other users' info\n", userInfo.ID)
				return false
			}
		}
	}

	return true
}

func setUpLog() {
}

func main() {
	util.LogInit(os.Stderr, os.Stdout, os.Stdout, os.Stderr, "AUTH")
	setUpRoutes()
	setUpDB()
	setUpLog()
	setupRoutines()

	router := mux.NewRouter().StrictSlash(true)

	for idx := range routes {
		route := routes[idx]
		path := fmt.Sprintf("%s/v%d/%s", prefixURL, route.Version, route.Pattern)
		router.
			Methods(route.Method).
			Path(path).
			Name(route.Name).
			HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer func() {
					if err := recover(); err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						util.PrintRuntimeStack(10)

						errMsg := fmt.Sprintf("%#v", err)
						util.LogError.Println("Panic error:", errMsg)
					}
				}()
				if checkAuth(r, route) {
					if route.HandlerFunc != nil {
						route.HandlerFunc(w, r)
					}
				} else {
					controller.ReturnUnauthorized(w)
				}
			})

		if route.Queries != nil {
			router.Queries(route.Queries...)
		}

		util.LogInfo.Printf("Setup for path [%s:%s], %+v", route.Method, path, route.GrantType)
	}

	util.SetJWTExpireTime(util.GetJWTExpireTimeConfig())

	url, port := util.GetServerConfig()
	serverBind := fmt.Sprintf("%s:%d", url, port)
	util.LogInfo.Printf("Start auth server on %s\n", serverBind)
	err := http.ListenAndServe(serverBind, router)
	if err != nil {
		util.LogError.Panicln(err.Error())
		os.Exit(1)
	}
}

func setupRoutines() {
	ticker := time.NewTicker(time.Hour * 12)
	go func() {
		for t := range ticker.C {
			service.ClearExpireToken()
			util.LogInfo.Println("Clear expired token at", t.Unix())
		}
	}()
}

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

	userInfo := data.UserDetailV3{}
	err := userInfo.SetValueWithToken(vals[1])
	if err != nil {
		return nil
	}

	return &userInfo
}
