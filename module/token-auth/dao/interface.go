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

	// v3
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

	GetAppsV3(enterpriseID string) ([]*data.AppV3, error)
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

	AddAuditLog(auditLog data.AuditLog) error
}
