package dao

import (
	"emotibot.com/emotigo/module/token-auth/internal/data"
)

// DB define interface for different dao modules
type DB interface {
	GetEnterprises() (*data.Enterprises, error)
	GetEnterprise(enterpriseID string) (*data.Enterprise, error)

	GetUsers(enterpriseID string) (*data.Users, error)
	GetUser(enterpriseID string, userID string) (*data.User, error)
	GetAdminUser(enterpriseID string) (*data.User, error)
	GetAuthUser(account string, passwd string) (user *data.User, err error)

	AddUser(enterpriseID string, user *data.User, roleID string) (userID string, err error)
	UpdateUser(enterpriseID string, user *data.User) error
	DeleteUser(enterpriseID string, userID string) (bool, error)

	GetRoles(enterpriseID string) ([]*data.Role, error)
	GetRole(enterpriseID string, roleID string) (*data.Role, error)
	AddRole(enterprise string, role *data.Role) (string, error)
	UpdateRole(enterprise string, roleID string, role *data.Role) (bool, error)
	DeleteRole(enterprise string, roleID string) (bool, error)
	GetUsersOfRole(enterpriseID string, roleID string) (*data.Users, error)

	GetModules(enterpriseID string) ([]*data.Module, error)
	AddEnterprise(enterprise *data.Enterprise, adminUser *data.User) (string, error)
	AddApp(enterpriseID string, app *data.App) (string, error)

	// TODO
	DisableUser(enterpriseID string, userID string) (bool, error)

	DeleteEnterprise(enterpriseID string) (bool, error)
	GetApps(enterpriseID string) (*data.Apps, error)
	GetApp(enterpriseID string, AppID string) (*data.App, error)
	UpdateApp(enterpriseID string, app data.App) (*data.App, error)
	DisableApp(enterpriseID string, AppID string) (bool, error)
	DeleteApp(enterpriseID string, AppID string) (bool, error)
}

type DBV3 interface {
	// v3
	GetEnterpriseAppListV3(enterpriseID *string, userID *string) ([]*data.EnterpriseAppListV3, error)
	GetEnterprisesV3() ([]*data.EnterpriseV3, error)
	GetEnterpriseV3(enterpriseID string) (*data.EnterpriseDetailV3, error)
	AddEnterpriseV3(enterprise *data.EnterpriseV3, modules []string, adminUser *data.UserDetailV3) (string, error)
	UpdateEnterpriseV3(enterprsieID string, newEnterprise *data.EnterpriseDetailV3, modules []string) error
	DeleteEnterpriseV3(enterprsieID string) error

	EnterpriseExistsV3(enterpriseID string) (bool, error)
	EnterpriseInfoExistsV3(enterpriseName string) (bool, error)

	GetUsersV3(enterpriseID string, admin bool) ([]*data.UserV3, error)
	GetUserV3(enterpriseID string, userID string) (*data.UserDetailV3, error)
	AddUserV3(enterpriseID string, user *data.UserDetailV3) (userID string, err error)
	UpdateUserV3(enterpriseID string, userID string, user *data.UserDetailV3) error
	DeleteUserV3(enterpriseID string, userID string) error

	GetAuthUserV3(account string, passwd string) (user *data.UserDetailV3, err error)
	GetUserPasswordV3(userID string) (string, error)
	UserExistsV3(userID string) (bool, error)
	EnterpriseUserInfoExistsV3(userType int, enterpriseID string,
		userName string, userEmail string) (bool, string, string, error)

	GetAppsV3(enterpriseID string) ([]*data.AppDetailV3, error)
	GetAppV3(enterpriseID string, appID string) (*data.AppDetailV3, error)
	AddAppV3(enterpriseID string, app *data.AppDetailV3) (string, error)
	UpdateAppV3(enterpriseID string, appID string, app *data.AppDetailV3) error
	DeleteAppV3(enterpriseID string, appID string) error

	AppExistsV3(appID string) (bool, error)
	EnterpriseAppInfoExistsV3(enterpriseID string, appName string) (bool, error)

	GetGroupsV3(enterpriseID string) ([]*data.GroupDetailV3, error)
	GetGroupV3(enterpriseID string, groupID string) (*data.GroupDetailV3, error)
	AddGroupV3(enterpriseID string, group *data.GroupDetailV3, apps []string) (string, error)
	UpdateGroupV3(enterpriseID string, groupID string, group *data.GroupDetailV3, apps []string) error
	DeleteGroupV3(enterpriseID string, groupID string) error

	GroupExistsV3(groupID string) (bool, error)
	EnterpriseGroupInfoExistsV3(enterpriseID string, groupName string) (bool, error)

	GetRolesV3(enterpriseID string) ([]*data.RoleV3, error)
	GetRoleV3(enterpriseID string, roleID string) (*data.RoleV3, error)
	AddRoleV3(enterpriseID string, role *data.RoleV3) (string, error)
	UpdateRoleV3(enterpriseID string, roleID string, role *data.RoleV3) error
	DeleteRoleV3(enterpriseID string, roleID string) error

	RoleExistsV3(roleID string) (bool, error)
	EnterpriseRoleInfoExistsV3(enterpriseID string, roleName string) (bool, error)

	GetModulesV3(enterpriseID string) ([]*data.ModuleDetailV3, error)

	GetEnterpriseIDV3(appID string) (string, error)
	GetUserV3ByKeyValue(key string, value string) (*data.UserDetailV3, error)

	GetAppSecretV3(appid string) (string, error)
	GetEnterpriseSecretV3(enterprise string) (string, error)
	RenewAppSecretV3(appid string) (string, error)
	RenewEnterpriseSecretV3(enterprise string) (string, error)
	GenerateAppApiKeyV3(enterprise, appid string, expired int) (string, error)
	GetAppViaApiKey(apiKey string) (string, error)
	GetEnterpriseViaApiKey(apiKey string) (string, error)
	RemoveAppApiKeyV3(appid, token string) error
	RemoveAppAllApiKeyV3(appid string) error
	ClearExpireToken()

	AddAuditLog(auditLog data.AuditLog) error
}

type DBV4 interface {
	// all interface with DBV3
	// GetEnterpriseAppList(enterpriseID *string, userID *string) ([]*data.EnterpriseAppListV3, error)
	// GetEnterprises() ([]*data.EnterpriseV3, error)
	// GetEnterprise(enterpriseID string) (*data.EnterpriseDetailV3, error)
	AddEnterpriseV4(enterprise *data.EnterpriseV3, modules []string, adminUser *data.UserDetailV3, dryRun, active bool) (string, error)
	// UpdateEnterprise(enterprsieID string, newEnterprise *data.EnterpriseDetailV3, modules []string) error
	// DeleteEnterprise(enterprsieID string) error
	UpdateEnterpriseStatusV4(enterpriseID string, status bool) error
	ActivateEnterpriseV4(enterpriseID string, username string, password string) error

	// EnterpriseExists(enterpriseID string) (bool, error)
	// EnterpriseInfoExists(enterpriseName string) (bool, error)

	// GetUsers(enterpriseID string, admin bool) ([]*data.UserV3, error)
	// GetUser(enterpriseID string, userID string) (*data.UserDetailV3, error)
	// AddUser(enterpriseID string, user *data.UserDetailV3) (userID string, err error)
	// UpdateUser(enterpriseID string, userID string, user *data.UserDetailV3) error
	// DeleteUser(enterpriseID string, userID string) error

	// GetAuthUser(account string, passwd string) (user *data.UserDetailV3, err error)
	// GetUserPassword(userID string) (string, error)
	// UserExists(userID string) (bool, error)
	// EnterpriseUserInfoExists(userType int, enterpriseID string,
	// 	userName string, userEmail string) (bool, string, string, error)

	// GetApps(enterpriseID string) ([]*data.AppDetailV3, error)
	// GetApp(enterpriseID string, appID string) (*data.AppDetailV3, error)
	// AddApp(enterpriseID string, app *data.AppDetailV3) (string, error)
	// UpdateApp(enterpriseID string, appID string, app *data.AppDetailV3) error
	// DeleteApp(enterpriseID string, appID string) error

	// AppExists(appID string) (bool, error)
	// EnterpriseAppInfoExists(enterpriseID string, appName string) (bool, error)

	// GetGroups(enterpriseID string) ([]*data.GroupDetailV3, error)
	// GetGroup(enterpriseID string, groupID string) (*data.GroupDetailV3, error)
	// AddGroup(enterpriseID string, group *data.GroupDetailV3, apps []string) (string, error)
	// UpdateGroup(enterpriseID string, groupID string, group *data.GroupDetailV3, apps []string) error
	// DeleteGroup(enterpriseID string, groupID string) error

	// GroupExists(groupID string) (bool, error)
	// EnterpriseGroupInfoExists(enterpriseID string, groupName string) (bool, error)

	// GetRoles(enterpriseID string) ([]*data.RoleV3, error)
	// GetRole(enterpriseID string, roleID string) (*data.RoleV3, error)
	// AddRole(enterpriseID string, role *data.RoleV3) (string, error)
	// UpdateRole(enterpriseID string, roleID string, role *data.RoleV3) error
	// DeleteRole(enterpriseID string, roleID string) error

	// RoleExists(roleID string) (bool, error)
	// EnterpriseRoleInfoExists(enterpriseID string, roleName string) (bool, error)

	// GetModules(enterpriseID string) ([]*data.ModuleDetailV3, error)

	// GetEnterpriseID(appID string) (string, error)
	// GetUserV3ByKeyValue(key string, value string) (*data.UserDetailV3, error)

	// GetAppSecret(appid string) (string, error)
	// RenewAppSecret(appid string) (string, error)
	// GenerateAppApiKey(appid string, expired int) (string, error)
	// GetAppViaApiKey(apiKey string) (string, error)
	// RemoveAppApiKey(appid, token string) error
	// RemoveAppAllApiKey(appid string) error
	// ClearExpireToken()

	// AddAuditLog(auditLog data.AuditLog) error

	AddAppV4(enterpriseID string, app *data.AppDetailV4) (string, error)
	GetAppsV4(enterpriseID string) ([]*data.AppDetailV4, error)

	// OAuth part
	GetOAuthClient(clientID string) (*data.OAuthClient, error)
}
