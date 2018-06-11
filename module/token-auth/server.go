package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"emotibot.com/emotigo/module/token-auth/dao"
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
		Route{"GetEnterprises", "GET", 2, "enterprises", EnterprisesGetHandler, []interface{}{0}},
		Route{"GetEnterprise", "GET", 2, "enterprise/{enterpriseID}", EnterpriseGetHandler, []interface{}{0, 1, 2}},
		Route{"GetUsers", "GET", 2, "enterprise/{enterpriseID}/users", UsersGetHandler, []interface{}{0, 1, 2}},
		Route{"GetUser", "GET", 2, "enterprise/{enterpriseID}/user/{userID}", UserGetHandler, []interface{}{0, 1, 2}},
		Route{"GetApps", "GET", 2, "enterprise/{enterpriseID}/apps", AppsGetHandler, []interface{}{0, 1, 2}},
		Route{"GetApp", "GET", 2, "enterprise/{enterpriseID}/app/{appID}", AppGetHandler, []interface{}{0, 1, 2}},
		Route{"Login", "POST", 2, "login", LoginHandler, []interface{}{}},
		Route{"ValidateToken", "GET", 2, "token/{token}", ValidateTokenHandler, []interface{}{}},
		Route{"ValidateToken", "GET", 2, "token", ValidateTokenHandler, []interface{}{}},

		Route{"AddUser", "POST", 2, "enterprise/{enterpriseID}/user", UserAddHandler, []interface{}{0, 1, 2}},
		Route{"UpdateUser", "PUT", 2, "enterprise/{enterpriseID}/user/{userID}", UserUpdateHandler, []interface{}{0, 1, 2}},
		Route{"DeleteUser", "DELETE", 2, "enterprise/{enterpriseID}/user/{userID}", UserDeleteHandler, []interface{}{0, 1, 2}},

		Route{"GetRoles", "GET", 2, "enterprise/{enterpriseID}/roles", RolesGetHandler, []interface{}{0, 1, 2}},
		Route{"GetRole", "GET", 2, "enterprise/{enterpriseID}/role/{roleID}", RoleGetHandler, []interface{}{0, 1, 2}},
		Route{"AddRole", "POST", 2, "enterprise/{enterpriseID}/role", RoleAddHandler, []interface{}{0, 1, 2}},
		Route{"UpdateRole", "PUT", 2, "enterprise/{enterpriseID}/role/{roleID}", RoleUpdateHandler, []interface{}{0, 1, 2}},
		Route{"DeleteRole", "DELETE", 2, "enterprise/{enterpriseID}/role/{roleID}", RoleDeleteHandler, []interface{}{0, 1, 2}},
		Route{"GetModules", "GET", 2, "enterprise/{enterpriseID}/modules", ModulesGetHandler, []interface{}{0, 1, 2}},

		// v3 API
		Route{"GetSystemAdmins", "GET", 3, "admins", SystemAdminsGetHandlerV3, []interface{}{0}},
		Route{"GetSystemAdmin", "GET", 3, "admin/{adminID}", SystemAdminGetHandlerV3, []interface{}{0}},
		Route{"AddSystemAdmin", "POST", 3, "admin", SystemAdminAddHandlerV3, []interface{}{0}},
		Route{"UpdateSystemAdmin", "PUT", 3, "admin/{adminID}", SystemAdminUpdateHandlerV3, []interface{}{0}},
		Route{"DeleteSystemAdmin", "DELETE", 3, "admin/{adminID}", SystemAdminDeleteHandlerV3, []interface{}{0}},

		Route{"GetEnterprises", "GET", 3, "enterprises", EnterprisesGetHandlerV3, []interface{}{0}},
		Route{"GetEnterprise", "GET", 3, "enterprise/{enterpriseID}", EnterpriseGetHandlerV3, []interface{}{0, 1, 2}},
		Route{"AddEnterprise", "POST", 3, "enterprise", EnterpriseAddHandlerV3, []interface{}{0}},
		Route{"UpdateEnterprise", "PUT", 3, "enterprise/{enterpriseID}", EnterpriseUpdateHandlerV3, []interface{}{0, 1}},
		Route{"DeleteEnterprise", "DELETE", 3, "enterprise/{enterpriseID}", EnterpriseDeleteHandlerV3, []interface{}{0, 1}},

		Route{"GetUsers", "GET", 3, "enterprise/{enterpriseID}/users", UsersGetHandlerV3, []interface{}{0, 1, 2}},
		Route{"GetUser", "GET", 3, "enterprise/{enterpriseID}/user/{userID}", UserGetHandlerV3, []interface{}{0, 1, 2}},
		Route{"AddUser", "POST", 3, "enterprise/{enterpriseID}/user", UserAddHandlerV3, []interface{}{0, 1, 2}},
		Route{"UpdateUser", "PUT", 3, "enterprise/{enterpriseID}/user/{userID}", UserUpdateHandlerV3, []interface{}{0, 1, 2}},
		Route{"DeleteUser", "DELETE", 3, "enterprise/{enterpriseID}/user/{userID}", UserDeleteHandlerV3, []interface{}{0, 1, 2}},

		Route{"GetApps", "GET", 3, "enterprise/{enterpriseID}/apps", AppsGetHandlerV3, []interface{}{0, 1, 2}},
		Route{"GetApp", "GET", 3, "enterprise/{enterpriseID}/app/{appID}", AppGetHandlerV3, []interface{}{0, 1, 2}},
		Route{"AddApp", "POST", 3, "enterprise/{enterpriseID}/app", AppAddHandlerV3, []interface{}{0, 1, 2}},
		Route{"UpdateApp", "PUT", 3, "enterprise/{enterpriseID}/app/{appID}", AppUpdateHandlerV3, []interface{}{0, 1, 2}},
		Route{"DeleteApp", "DELETE", 3, "enterprise/{enterpriseID}/app/{appID}", AppDeleteHandlerV3, []interface{}{0, 1, 2}},

		Route{"GetGroups", "GET", 3, "enterprise/{enterpriseID}/groups", GroupsGetHandlerV3, []interface{}{0, 1, 2}},
		Route{"GetGroup", "GET", 3, "enterprise/{enterpriseID}/group/{groupID}", GroupGetHandlerV3, []interface{}{0, 1, 2}},
		Route{"AddGroup", "POST", 3, "enterprise/{enterpriseID}/group", GroupAddHandlerV3, []interface{}{0, 1, 2}},
		Route{"UpdateGroup", "PUT", 3, "enterprise/{enterpriseID}/group/{groupID}", GroupUpdateHandlerV3, []interface{}{0, 1, 2}},
		Route{"DeleteGroup", "DELETE", 3, "enterprise/{enterpriseID}/group/{groupID}", GroupDeleteHandlerV3, []interface{}{0, 1, 2}},

		Route{"GetRoles", "GET", 3, "enterprise/{enterpriseID}/roles", RolesGetHandlerV3, []interface{}{0, 1, 2}},
		Route{"GetRole", "GET", 3, "enterprise/{enterpriseID}/role/{roleID}", RoleGetHandlerV3, []interface{}{0, 1, 2}},
		Route{"AddRole", "POST", 3, "enterprise/{enterpriseID}/role", RoleAddHandlerV3, []interface{}{0, 1, 2}},
		Route{"UpdateRole", "PUT", 3, "enterprise/{enterpriseID}/role/{roleID}", RoleUpdateHandlerV3, []interface{}{0, 1, 2}},
		Route{"DeleteRole", "DELETE", 3, "enterprise/{enterpriseID}/role/{roleID}", RoleDeleteHandlerV3, []interface{}{0, 1, 2}},

		Route{"Login", "POST", 3, "login", LoginHandlerV3, []interface{}{}},
		Route{"ValidateToken", "GET", 3, "token/{token}", ValidateTokenHandler, []interface{}{}},
		Route{"ValidateToken", "GET", 3, "token", ValidateTokenHandler, []interface{}{}},

		Route{"GetModules", "GET", 3, "enterprise/{enterpriseID}/modules", ModulesGetHandler, []interface{}{0, 1, 2}},
	}
}

func setUpDB() {
	db := dao.MYSQLController{}
	url, port, user, passwd, dbName := util.GetMySQLConfig()
	util.LogInfo.Printf("Init mysql: %s:%s@%s:%d/%s\n", user, passwd, url, port, dbName)
	db.InitDB(url, port, dbName, user, passwd)
	service.SetDB(&db)
}

func checkAuth(r *http.Request, route Route) bool {
	util.LogInfo.Printf("Access: %s %s", r.Method, r.RequestURI)
	if len(route.GrantType) == 0 {
		util.LogError.Println("[Auth check] pass: no need")
		return true
	}

	authorization := r.Header.Get("Authorization")
	vals := strings.Split(authorization, " ")
	if len(vals) < 2 {
		util.LogError.Println("[Auth check] Auth fail: no header")
		return false
	}

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
	// Type 1 can only check enterprise of itself
	// Type 2 can only check enterprise of itself and user info of itself
	if userInfo.Type == 1 || userInfo.Type == 2 {
		enterpriseID := vars["enterpriseID"]
		if enterpriseID != *userInfo.Enterprise {
			util.LogInfo.Printf("[Auth check] user of [%s] can not access [%s]\n", *userInfo.Enterprise, enterpriseID)
			return false
		}
	}

	if userInfo.Type == 2 {
		userID := vars["userID"]
		if userID != "" && userID != userInfo.ID {
			util.LogInfo.Printf("[Auth check] user [%s] can not access other users' info\n", userInfo.ID)
			return false
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
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
				}
			})
		util.LogInfo.Printf("Setup for path [%s:%s], %+v", route.Method, path, route.GrantType)
	}

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
