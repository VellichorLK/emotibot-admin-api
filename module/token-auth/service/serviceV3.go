package service

import (
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
	return useDB.AddUserV3("", admin)
}

func UpdateSystemAdminV3(admin *data.UserDetailV3, adminID string) error {
	err := checkDB()
	if err != nil {
		return err
	}

	return useDB.UpdateUserV3("", adminID, admin)
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

	return useDB.AddEnterpriseV3(enterprise, modules, adminUser)
}

func UpdateEnterpriseV3(enterpriseID string, newEnterprise *data.EnterpriseV3,
	modules []string) error {
	err := checkDB()
	if err != nil {
		return err
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

	return useDB.GetUserV3(enterpriseID, userID)
}

func AddUserV3(enterpriseID string, user *data.UserDetailV3) (string, error) {
	err := checkDB()
	if err != nil {
		return "", err
	}

	return useDB.AddUserV3(enterpriseID, user)
}

func UpdateUserV3(enterpriseID string, userID string, user *data.UserDetailV3) error {
	err := checkDB()
	if err != nil {
		return err
	}

	return useDB.UpdateUserV3(enterpriseID, userID, user)
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

	return useDB.AddAppV3(enterpriseID, app)
}

func UpdateAppV3(enterpriseID string, appID string, app *data.AppDetailV3) error {
	err := checkDB()
	if err != nil {
		return err
	}

	return useDB.UpdateAppV3(enterpriseID, appID, app)
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

	return useDB.AddGroupV3(enterpriseID, group, apps)
}

func UpdateGroupV3(enterpriseID string, groupID string,
	group *data.GroupDetailV3, apps []string) error {
	err := checkDB()
	if err != nil {
		return err
	}

	return useDB.UpdateGroupV3(enterpriseID, groupID, group, apps)
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

	return useDB.AddRoleV3(enterpriseID, role)
}

func UpdateRoleV3(enterpriseID string, roleID string, role *data.RoleV3) error {
	err := checkDB()
	if err != nil {
		return err
	}

	return useDB.UpdateRoleV3(enterpriseID, roleID, role)
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

	usersCount, err := useDB.GetUsersCountOfRoleV3(roleID)
	if err != nil {
		return false, err
	}

	if usersCount > 0 {
		return false, util.ErrRoleUsersNotEmpty
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
