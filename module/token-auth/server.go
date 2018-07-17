package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"emotibot.com/emotigo/module/token-auth/dao"
	"emotibot.com/emotigo/module/token-auth/internal/audit"
	"emotibot.com/emotigo/module/token-auth/internal/data"
	"emotibot.com/emotigo/module/token-auth/internal/enum"
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
		Route{"GetEnterprises", "GET", 2, "enterprises", nil, EnterprisesGetHandler, []interface{}{0}},
		Route{"GetEnterprise", "GET", 2, "enterprise/{enterpriseID}", nil, EnterpriseGetHandler, []interface{}{0, 1, 2}},
		Route{"GetUsers", "GET", 2, "enterprise/{enterpriseID}/users", nil, UsersGetHandler, []interface{}{0, 1, 2}},
		Route{"GetUser", "GET", 2, "enterprise/{enterpriseID}/user/{userID}", nil, UserGetHandler, []interface{}{0, 1, 2}},
		Route{"GetApps", "GET", 2, "enterprise/{enterpriseID}/apps", nil, AppsGetHandler, []interface{}{0, 1, 2}},
		Route{"GetApp", "GET", 2, "enterprise/{enterpriseID}/app/{appID}", nil, AppGetHandler, []interface{}{0, 1, 2}},
		Route{"Login", "POST", 2, "login", nil, LoginHandler, []interface{}{}},
		Route{"ValidateToken", "GET", 2, "token/{token}", nil, ValidateTokenHandler, []interface{}{}},
		Route{"ValidateToken", "GET", 2, "token", nil, ValidateTokenHandler, []interface{}{}},

		Route{"AddUser", "POST", 2, "enterprise/{enterpriseID}/user", nil, UserAddHandler, []interface{}{0, 1, 2}},
		Route{"UpdateUser", "PUT", 2, "enterprise/{enterpriseID}/user/{userID}", nil, UserUpdateHandler, []interface{}{0, 1, 2}},
		Route{"DeleteUser", "DELETE", 2, "enterprise/{enterpriseID}/user/{userID}", nil, UserDeleteHandler, []interface{}{0, 1, 2}},

		Route{"GetRoles", "GET", 2, "enterprise/{enterpriseID}/roles", nil, RolesGetHandler, []interface{}{0, 1, 2}},
		Route{"GetRole", "GET", 2, "enterprise/{enterpriseID}/role/{roleID}", nil, RoleGetHandler, []interface{}{0, 1, 2}},
		Route{"AddRole", "POST", 2, "enterprise/{enterpriseID}/role", nil, RoleAddHandler, []interface{}{0, 1, 2}},
		Route{"UpdateRole", "PUT", 2, "enterprise/{enterpriseID}/role/{roleID}", nil, RoleUpdateHandler, []interface{}{0, 1, 2}},
		Route{"DeleteRole", "DELETE", 2, "enterprise/{enterpriseID}/role/{roleID}", nil, RoleDeleteHandler, []interface{}{0, 1, 2}},
		Route{"GetModules", "GET", 2, "enterprise/{enterpriseID}/modules", nil, ModulesGetHandler, []interface{}{0, 1, 2}},

		// v3 API
		Route{"GetSystemAdmins", "GET", 3, "admins", nil, SystemAdminsGetHandlerV3, []interface{}{0}},
		Route{"GetSystemAdmin", "GET", 3, "admin/{adminID}", nil, SystemAdminGetHandlerV3, []interface{}{0}},
		Route{"AddSystemAdmin", "POST", 3, "admin", nil, SystemAdminAddHandlerV3, []interface{}{0}},
		Route{"UpdateSystemAdmin", "PUT", 3, "admin/{adminID}", nil, SystemAdminUpdateHandlerV3, []interface{}{0}},
		Route{"DeleteSystemAdmin", "DELETE", 3, "admin/{adminID}", nil, SystemAdminDeleteHandlerV3, []interface{}{0}},

		Route{"GetEnterprises", "GET", 3, "enterprises", nil, EnterprisesGetHandlerV3, []interface{}{0, 1, 2}},
		Route{"GetEnterprise", "GET", 3, "enterprise/{enterpriseID}", nil, EnterpriseGetHandlerV3, []interface{}{0, 1, 2}},
		Route{"AddEnterprise", "POST", 3, "enterprise", nil, EnterpriseAddHandlerV3, []interface{}{0}},
		Route{"UpdateEnterprise", "PUT", 3, "enterprise/{enterpriseID}", nil, EnterpriseUpdateHandlerV3, []interface{}{0, 1}},
		Route{"DeleteEnterprise", "DELETE", 3, "enterprise/{enterpriseID}", nil, EnterpriseDeleteHandlerV3, []interface{}{0}},

		Route{"GetUsers", "GET", 3, "enterprise/{enterpriseID}/users", nil, UsersGetHandlerV3, []interface{}{0, 1, 2}},
		Route{"GetUser", "GET", 3, "enterprise/{enterpriseID}/user/{userID}", nil, UserGetHandlerV3, []interface{}{0, 1, 2}},
		Route{"AddUser", "POST", 3, "enterprise/{enterpriseID}/user", nil, UserAddHandlerV3, []interface{}{0, 1}},
		Route{"UpdateUser", "PUT", 3, "enterprise/{enterpriseID}/user/{userID}", nil, UserUpdateHandlerV3, []interface{}{0, 1, 2}},
		Route{"DeleteUser", "DELETE", 3, "enterprise/{enterpriseID}/user/{userID}", nil, UserDeleteHandlerV3, []interface{}{0, 1}},

		Route{"GetApps", "GET", 3, "enterprise/{enterpriseID}/apps", nil, AppsGetHandlerV3, []interface{}{0, 1, 2}},
		Route{"GetApp", "GET", 3, "enterprise/{enterpriseID}/app/{appID}", nil, AppGetHandlerV3, []interface{}{0, 1, 2}},
		Route{"AddApp", "POST", 3, "enterprise/{enterpriseID}/app", nil, AppAddHandlerV3, []interface{}{0, 1}},
		Route{"UpdateApp", "PUT", 3, "enterprise/{enterpriseID}/app/{appID}", nil, AppUpdateHandlerV3, []interface{}{0, 1}},
		Route{"DeleteApp", "DELETE", 3, "enterprise/{enterpriseID}/app/{appID}", nil, AppDeleteHandlerV3, []interface{}{0, 1}},

		Route{"GetGroups", "GET", 3, "enterprise/{enterpriseID}/groups", nil, GroupsGetHandlerV3, []interface{}{0, 1, 2}},
		Route{"GetGroup", "GET", 3, "enterprise/{enterpriseID}/group/{groupID}", nil, GroupGetHandlerV3, []interface{}{0, 1, 2}},
		Route{"AddGroup", "POST", 3, "enterprise/{enterpriseID}/group", nil, GroupAddHandlerV3, []interface{}{0, 1}},
		Route{"UpdateGroup", "PUT", 3, "enterprise/{enterpriseID}/group/{groupID}", nil, GroupUpdateHandlerV3, []interface{}{0, 1}},
		Route{"DeleteGroup", "DELETE", 3, "enterprise/{enterpriseID}/group/{groupID}", nil, GroupDeleteHandlerV3, []interface{}{0, 1}},

		Route{"GetRoles", "GET", 3, "enterprise/{enterpriseID}/roles", nil, RolesGetHandlerV3, []interface{}{0, 1, 2}},
		Route{"GetRole", "GET", 3, "enterprise/{enterpriseID}/role/{roleID}", nil, RoleGetHandlerV3, []interface{}{0, 1, 2}},
		Route{"AddRole", "POST", 3, "enterprise/{enterpriseID}/role", nil, RoleAddHandlerV3, []interface{}{0, 1}},
		Route{"UpdateRole", "PUT", 3, "enterprise/{enterpriseID}/role/{roleID}", nil, RoleUpdateHandlerV3, []interface{}{0, 1}},
		Route{"DeleteRole", "DELETE", 3, "enterprise/{enterpriseID}/role/{roleID}", nil, RoleDeleteHandlerV3, []interface{}{0, 1}},

		Route{"Login", "POST", 3, "login", nil, LoginHandlerV3, []interface{}{}},
		Route{"ValidateToken", "GET", 3, "token/{token}", nil, ValidateTokenHandler, []interface{}{}},
		Route{"ValidateToken", "GET", 3, "token", nil, ValidateTokenHandler, []interface{}{}},

		Route{"GetModules", "GET", 3, "enterprise/{enterpriseID}/modules", nil, ModulesGetHandlerV3, []interface{}{0, 1, 2}},
		Route{"GetModules", "GET", 3, "modules", nil, GlobalModulesGetHandlerV3, []interface{}{}},

		Route{"GetEnterpriseId", "GET", 3, "getEnterpriseId", []string{"app-id", "{app-id}"}, EnterpriseIDGetHandlerV3, []interface{}{}},
	}
}

func setUpDB() {
	db := dao.MYSQLController{}

	url, port, user, passwd, dbName := util.GetMySQLConfig()
	util.LogInfo.Printf("Init mysql: %s:%s@%s:%d/%s\n", user, passwd, url, port, dbName)
	db.InitDB(url, port, dbName, user, passwd)
	service.SetDB(&db)

	url, port, user, passwd, dbName = util.GetMySQLAuditConfig()
	util.LogInfo.Printf("Init audit mysql: %s:%s@%s:%d/%s\n", user, passwd, url, port, dbName)
	db.InitAuditDB(url, port, dbName, user, passwd)
	audit.SetDB(&db)
}

func checkAuth(r *http.Request, route Route) bool {
	util.LogInfo.Printf("Access: %s %s", r.Method, r.RequestURI)
	if len(route.GrantType) == 0 {
		util.LogTrace.Println("[Auth check] pass: no need")
		return true
	}

	authorization := r.Header.Get("Authorization")
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

	router := mux.NewRouter().StrictSlash(true)

	for idx := range routes {
		route := routes[idx]
		path := fmt.Sprintf("%s/v%d/%s", prefixURL, route.Version, route.Pattern)
		router.
			Methods(route.Method).
			Path(path).
			Name(route.Name).
			HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if checkAuth(r, route) {
					if route.HandlerFunc != nil {
						route.HandlerFunc(w, r)
					}
				} else {
					returnUnauthorized(w)
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
