package service

import (
	"database/sql"
	"errors"

	"emotibot.com/emotigo/module/token-auth/internal/data"
	"emotibot.com/emotigo/module/token-auth/internal/util"
)

func checkDB() error {
	if useDB == nil {
		return errors.New("DB hasn't set")
	}
	return nil
}

func GetSystemAdminsV3() ([]*data.UserV3, error) {
	err := checkDB()
	if err != nil {
		return nil, err
	}
	return useDB.GetUsersV3("", true)
}

func GetSystemAdminV3(adminID string) (*data.UserDetailV3, error) {
	err := checkDB()
	if err != nil {
		return nil, err
	}
	return useDB.GetUserV3("", adminID)
}

func AddSystemAdminV3(admin *data.UserDetailV3) (string, error) {
	err := checkDB()
	if err != nil {
		return "", err
	}

	exists, existedAdminName, _, err := useDB.EnterpriseUserInfoExistsV3(admin.Type, "",
		admin.UserName, admin.Email)
	if err != nil {
		return "", err
	} else if exists {
		if admin.UserName == existedAdminName {
			return "", util.ErrUserNameExists
		}
	}

	return useDB.AddUserV3("", admin)
}

func UpdateSystemAdminV3(origAdmin *data.UserDetailV3, newAdmin *data.UserDetailV3, adminID string) error {
	err := checkDB()
	if err != nil {
		return err
	}

	if newAdmin.UserName != origAdmin.UserName || newAdmin.Email != origAdmin.Email {
		exists, existedAdminName, _, err := useDB.EnterpriseUserInfoExistsV3(
			newAdmin.Type, "", newAdmin.UserName, newAdmin.Email)
		if err != nil {
			return err
		} else if exists {
			if newAdmin.UserName == existedAdminName {
				return util.ErrUserNameExists
			}
		}
	}

	return useDB.UpdateUserV3("", adminID, newAdmin)
}

func DeleteSystemAdminV3(adminID string) (bool, error) {
	err := checkDB()
	if err != nil {
		return false, err
	}

	exists, err := useDB.UserExistsV3(adminID)
	if err != nil {
		return false, err
	} else if !exists {
		return false, nil
	}

	err = useDB.DeleteUserV3("", adminID)
	if err != nil {
		return false, err
	}

	return true, nil
}

func GetEnterprisesV3() ([]*data.EnterpriseV3, error) {
	err := checkDB()
	if err != nil {
		return nil, err
	}

	return useDB.GetEnterprisesV3()
}

func GetEnterpriseV3(enterpriseID string) (*data.EnterpriseDetailV3, error) {
	err := checkDB()
	if err != nil {
		return nil, err
	}

	return useDB.GetEnterpriseV3(enterpriseID)
}

func AddEnterpriseV3(enterprise *data.EnterpriseV3, modules []string,
	adminUser *data.UserDetailV3) (enterpriseID string, err error) {
	err = checkDB()
	if err != nil {
		return "", err
	}

	exists, err := useDB.EnterpriseInfoExistsV3(enterprise.Name)
	if err != nil {
		return "", err
	} else if exists {
		return "", util.ErrEnterpriseInfoExists
	}

	return useDB.AddEnterpriseV3(enterprise, modules, adminUser)
}

func UpdateEnterpriseV3(enterpriseID string, origEnterprise *data.EnterpriseDetailV3,
	newEnterprise *data.EnterpriseDetailV3, modules []string) error {
	err := checkDB()
	if err != nil {
		return err
	}

	if newEnterprise.Name != origEnterprise.Name {
		exists, err := useDB.EnterpriseInfoExistsV3(newEnterprise.Name)
		if err != nil {
			return err
		} else if exists {
			return util.ErrEnterpriseInfoExists
		}
	}

	return useDB.UpdateEnterpriseV3(enterpriseID, newEnterprise, modules)
}

func DeleteEnterpriseV3(enterpriseID string) (bool, error) {
	err := checkDB()
	if err != nil {
		return false, err
	}

	exists, err := useDB.EnterpriseExistsV3(enterpriseID)
	if err != nil {
		return false, err
	} else if !exists {
		return false, nil
	}

	err = useDB.DeleteEnterpriseV3(enterpriseID)
	if err != nil {
		return false, err
	}

	return true, nil
}

func GetUsersV3(enterpriseID string) ([]*data.UserV3, error) {
	err := checkDB()
	if err != nil {
		return nil, err
	}

	exists, err := useDB.EnterpriseExistsV3(enterpriseID)
	if err != nil {
		return nil, err
	} else if !exists {
		return nil, nil
	}

	users, err := useDB.GetUsersV3(enterpriseID, false)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func GetUserV3(enterpriseID string, userID string) (*data.UserDetailV3, error) {
	err := checkDB()
	if err != nil {
		return nil, err
	}

	info, err := useDB.GetUserV3(enterpriseID, userID)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return info, nil
}

func AddUserV3(enterpriseID string, user *data.UserDetailV3) (string, error) {
	err := checkDB()
	if err != nil {
		return "", err
	}

	exists, err := useDB.EnterpriseExistsV3(enterpriseID)
	if err != nil {
		return "", err
	} else if !exists {
		return "", nil
	}

	exists, existedUserName, _, err := useDB.EnterpriseUserInfoExistsV3(user.Type,
		enterpriseID, user.UserName, user.Email)
	if err != nil {
		return "", err
	} else if exists {
		if user.UserName == existedUserName {
			return "", util.ErrUserNameExists
		}
	}

	err = checkUserRoles(user, enterpriseID)
	if err != nil {
		return "", err
	}

	return useDB.AddUserV3(enterpriseID, user)
}

func UpdateUserV3(enterpriseID string, userID string,
	origUser *data.UserDetailV3, newUser *data.UserDetailV3) error {
	err := checkDB()
	if err != nil {
		return err
	}

	if newUser.UserName != origUser.UserName || newUser.Email != origUser.Email {
		exists, existedUserName, _, err := useDB.EnterpriseUserInfoExistsV3(newUser.Type, enterpriseID,
			newUser.UserName, newUser.Email)
		if err != nil {
			return err
		} else if exists {
			if newUser.UserName != origUser.UserName && newUser.UserName == existedUserName {
				return util.ErrUserNameExists
			}
		}
	}

	err = checkUserRoles(newUser, enterpriseID)
	if err != nil {
		return err
	}

	return useDB.UpdateUserV3(enterpriseID, userID, newUser)
}

func DeleteUserV3(enterpriseID string, userID string) (bool, error) {
	err := checkDB()
	if err != nil {
		return false, err
	}

	exists, err := useDB.UserExistsV3(userID)
	if err != nil {
		return false, err
	} else if !exists {
		return false, nil
	}

	err = useDB.DeleteUserV3(enterpriseID, userID)
	if err != nil {
		return false, err
	}

	return true, nil
}

func GetAppsV3(enterpriseID string) ([]*data.AppV3, error) {
	err := checkDB()
	if err != nil {
		return nil, err
	}

	exists, err := useDB.EnterpriseExistsV3(enterpriseID)
	if err != nil {
		return nil, err
	} else if !exists {
		return nil, nil
	}

	return useDB.GetAppsV3(enterpriseID)
}

func GetAppV3(enterpriseID string, appID string) (*data.AppDetailV3, error) {
	err := checkDB()
	if err != nil {
		return nil, err
	}

	return useDB.GetAppV3(enterpriseID, appID)
}

func AddAppV3(enterpriseID string, app *data.AppDetailV3) (appID string, err error) {
	err = checkDB()
	if err != nil {
		return "", err
	}

	exists, err := useDB.EnterpriseExistsV3(enterpriseID)
	if err != nil {
		return "", err
	} else if !exists {
		return "", nil
	}

	exists, err = useDB.EnterpriseAppInfoExistsV3(enterpriseID, app.Name)
	if err != nil {
		return "", err
	} else if exists {
		return "", util.ErrAppInfoExists
	}

	return useDB.AddAppV3(enterpriseID, app)
}

func UpdateAppV3(enterpriseID string, appID string,
	origApp *data.AppDetailV3, newApp *data.AppDetailV3) error {
	err := checkDB()
	if err != nil {
		return err
	}

	if newApp.Name != origApp.Name {
		exists, err := useDB.EnterpriseAppInfoExistsV3(enterpriseID, newApp.Name)
		if err != nil {
			return err
		} else if exists {
			return util.ErrAppInfoExists
		}
	}

	return useDB.UpdateAppV3(enterpriseID, appID, newApp)
}

func DeleteAppV3(enterpriseID string, appID string) (bool, error) {
	err := checkDB()
	if err != nil {
		return false, err
	}

	exists, err := useDB.AppExistsV3(appID)
	if err != nil {
		return false, err
	} else if !exists {
		return false, nil
	}

	err = useDB.DeleteAppV3(enterpriseID, appID)
	if err != nil {
		return false, err
	}

	return true, nil
}

func GetGroupsV3(enterpriseID string) ([]*data.GroupDetailV3, error) {
	err := checkDB()
	if err != nil {
		return nil, err
	}

	exists, err := useDB.EnterpriseExistsV3(enterpriseID)
	if err != nil {
		return nil, err
	} else if !exists {
		return nil, nil
	}

	return useDB.GetGroupsV3(enterpriseID)
}

func GetGroupV3(enterpriseID string, groupID string) (*data.GroupDetailV3, error) {
	err := checkDB()
	if err != nil {
		return nil, err
	}

	return useDB.GetGroupV3(enterpriseID, groupID)
}

func AddGroupV3(enterpriseID string, group *data.GroupDetailV3,
	apps []string) (groupID string, err error) {
	err = checkDB()
	if err != nil {
		return "", err
	}

	exists, err := useDB.EnterpriseExistsV3(enterpriseID)
	if err != nil {
		return "", err
	} else if !exists {
		return "", nil
	}

	exists, err = useDB.EnterpriseGroupInfoExistsV3(enterpriseID, group.Name)
	if err != nil {
		return "", err
	} else if exists {
		return "", util.ErrGroupInfoExists
	}

	return useDB.AddGroupV3(enterpriseID, group, apps)
}

func UpdateGroupV3(enterpriseID string, groupID string, origGroup *data.GroupDetailV3,
	newGroup *data.GroupDetailV3, apps []string) error {
	err := checkDB()
	if err != nil {
		return err
	}

	if newGroup.Name != origGroup.Name {
		exists, err := useDB.EnterpriseGroupInfoExistsV3(enterpriseID, newGroup.Name)
		if err != nil {
			return err
		} else if exists {
			return util.ErrGroupInfoExists
		}
	}

	return useDB.UpdateGroupV3(enterpriseID, groupID, newGroup, apps)
}

func DeleteGroupV3(enterpriseID string, groupID string) (bool, error) {
	err := checkDB()
	if err != nil {
		return false, err
	}

	exists, err := useDB.GroupExistsV3(groupID)
	if err != nil {
		return false, err
	} else if !exists {
		return false, nil
	}

	err = useDB.DeleteGroupV3(enterpriseID, groupID)
	if err != nil {
		return false, err
	}

	return true, nil
}

func GetRolesV3(enterpriseID string) ([]*data.RoleV3, error) {
	err := checkDB()
	if err != nil {
		return nil, err
	}

	exists, err := useDB.EnterpriseExistsV3(enterpriseID)
	if err != nil {
		return nil, err
	} else if !exists {
		return nil, nil
	}

	return useDB.GetRolesV3(enterpriseID)
}

func GetRoleV3(enterpriseID string, roleID string) (*data.RoleV3, error) {
	err := checkDB()
	if err != nil {
		return nil, err
	}

	return useDB.GetRoleV3(enterpriseID, roleID)
}

func AddRoleV3(enterpriseID string, role *data.RoleV3) (string, error) {
	err := checkDB()
	if err != nil {
		return "", err
	}

	exists, err := useDB.EnterpriseExistsV3(enterpriseID)
	if err != nil {
		return "", err
	} else if !exists {
		return "", nil
	}

	exists, err = useDB.EnterpriseRoleInfoExistsV3(enterpriseID, role.Name)
	if err != nil {
		return "", err
	} else if exists {
		return "", util.ErrRoleInfoExists
	}

	return useDB.AddRoleV3(enterpriseID, role)
}

func UpdateRoleV3(enterpriseID string, roleID string, origRole *data.RoleV3,
	newRole *data.RoleV3) error {
	err := checkDB()
	if err != nil {
		return err
	}

	if newRole.Name != origRole.Name {
		exists, err := useDB.EnterpriseRoleInfoExistsV3(enterpriseID, newRole.Name)
		if err != nil {
			return err
		} else if exists {
			return util.ErrRoleInfoExists
		}
	}

	return useDB.UpdateRoleV3(enterpriseID, roleID, newRole)
}

func DeleteRoleV3(enterpriseID string, roleID string) (bool, error) {
	err := checkDB()
	if err != nil {
		return false, err
	}

	exists, err := useDB.RoleExistsV3(roleID)
	if err != nil {
		return false, err
	} else if !exists {
		return false, nil
	}

	err = useDB.DeleteRoleV3(enterpriseID, roleID)
	if err != nil {
		return false, err
	}

	return true, nil
}

func LoginV3(account string, passwd string) (*data.UserDetailV3, error) {
	err := checkDB()
	if err != nil {
		return nil, err
	}

	return useDB.GetAuthUserV3(account, passwd)
}

func GetModulesV3(enterpriseID string) ([]*data.ModuleDetailV3, error) {
	err := checkDB()
	if err != nil {
		return nil, err
	}

	exists, err := useDB.EnterpriseExistsV3(enterpriseID)
	if err != nil {
		return nil, err
	} else if !exists {
		return nil, nil
	}

	return useDB.GetModulesV3(enterpriseID)
}

func GetGlobalModulesV3() ([]*data.ModuleDetailV3, error) {
	err := checkDB()
	if err != nil {
		return nil, err
	}
	return useDB.GetModulesV3("")
}

func GetUserPasswordV3(userID string) (string, error) {
	err := checkDB()
	if err != nil {
		return "", err
	}

	exists, err := useDB.UserExistsV3(userID)
	if err != nil {
		return "", err
	} else if !exists {
		return "", nil
	}

	return useDB.GetUserPasswordV3(userID)
}

func GetEnterpriseIDV3(appID string) (string, error) {
	cacheMod := "app-enterprise"
	err := checkDB()
	if err != nil {
		return "", err
	}

	id := util.GetCacheValue(cacheMod, appID)
	if id != "" {
		util.LogTrace.Printf("Hit cache for appid [%s]: [%s]", appID, id)
		return id, nil
	}

	exists, err := useDB.AppExistsV3(appID)
	if err != nil {
		return "", err
	} else if !exists {
		return "", nil
	}

	id, err = useDB.GetEnterpriseIDV3(appID)
	if err != nil {
		return "", err
	}
	util.SetCache(cacheMod, appID, id)
	return id, nil
}

func checkUserRoles(user *data.UserDetailV3, enterpriseID string) error {
	if user.Roles == nil {
		return nil
	}

	if user.Roles.GroupRoles != nil {
		for _, groupRole := range user.Roles.GroupRoles {
			exist, err := useDB.GroupExistsV3(groupRole.ID)
			if err != nil {
				return err
			} else if !exist {
				return util.ErrRobotGroupNotExist
			}

			exist, err = useDB.RoleExistsV3(groupRole.Role)
			if err != nil {
				return err
			} else if !exist {
				return util.ErrRoleNotExist
			}
		}
	}

	if user.Roles.AppRoles != nil {
		for _, appRole := range user.Roles.AppRoles {
			exist, err := useDB.AppExistsV3(appRole.ID)
			if err != nil {
				return err
			} else if !exist {
				return util.ErrRobotNotExist
			}

			exist, err = useDB.RoleExistsV3(appRole.Role)
			if err != nil {
				return err
			} else if !exist {
				return util.ErrRoleNotExist
			}
		}
	}

	return nil
}

func GetEnterpriseApp(enterpriseID *string, userID *string) ([]*data.EnterpriseAppListV3, error) {
	return useDB.GetEnterpriseAppListV3(enterpriseID, userID)
}

func GetUserV3ByKeyValue(key string, value string) (*data.UserDetailV3, error) {
	return useDB.GetUserV3ByKeyValue(key, value)
}
