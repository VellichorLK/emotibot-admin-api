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

func UpdateSystemAdminV3(admin *data.UserDetailV3, adminID string) (bool, error) {
	err := checkDB()
	if err != nil {
		return false, err
	}
	return useDB.UpdateUserV3("", adminID, admin)
}

func DeleteSystemAdminV3(adminID string) (bool, error) {
	err := checkDB()
	if err != nil {
		return false, err
	}
	return useDB.DeleteUserV3("", adminID)
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
	defer func() {
		if err != nil {
			util.LogError.Printf("Error when get enterprise %s: %s\n", enterpriseID, err.Error())
		}
	}()

	return useDB.GetEnterpriseV3(enterpriseID)
}

func AddEnterpriseV3(name string, description string, modules []string,
	adminUser *data.UserDetailV3) (enterpriseID string, err error) {
	err = checkDB()
	if err != nil {
		return "", err
	}

	if name == "" || adminUser == nil {
		return "", errors.New("Parameter error")
	}
	if !adminUser.IsValid() {
		return "", errors.New("Invalid user")
	}

	enterprise := data.EnterpriseV3{
		Name:        name,
		Description: description,
	}

	return useDB.AddEnterpriseV3(&enterprise, modules, adminUser)
}

func UpdateEnterpriseV3(enterpriseID string, newEnterprise *data.EnterpriseV3,
	modules []string) (bool, error) {
	err := checkDB()
	if err != nil {
		return false, err
	}

	if enterpriseID == "" {
		return false, errors.New("Invalid enterpriseID")
	}

	return useDB.UpdateEnterpriseV3(enterpriseID, newEnterprise, modules)
}

func DeleteEnterpriseV3(enterpriseID string) (bool, error) {
	err := checkDB()
	if err != nil {
		return false, err
	}

	if enterpriseID == "" {
		return false, errors.New("Invalid enterpriseID")
	}

	return useDB.DeleteEnterpriseV3(enterpriseID)
}

func GetUsersV3(enterpriseID string) ([]*data.UserV3, error) {
	err := checkDB()
	if err != nil {
		return nil, err
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

	user, err := useDB.GetUserV3(enterpriseID, userID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func AddUserV3(enterpriseID string, user *data.UserDetailV3, roleID string) (string, error) {
	err := checkDB()
	if err != nil {
		return "", err
	}

	return useDB.AddUserV3(enterpriseID, user)
}

func UpdateUserV3(enterpriseID string, userID string, user *data.UserDetailV3) (bool, error) {
	err := checkDB()
	if err != nil {
		return false, err
	}

	return useDB.UpdateUserV3(enterpriseID, userID, user)
}

func DeleteUserV3(enterpriseID string, userID string) (bool, error) {
	err := checkDB()
	if err != nil {
		return false, err
	}

	return useDB.DeleteUserV3(enterpriseID, userID)
}

func GetAppsV3(enterpriseID string) ([]*data.AppV3, error) {
	err := checkDB()
	if err != nil {
		return nil, err
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

	if enterpriseID == "" {
		return "", errors.New("Invalid enterpriseID")
	}
	if app == nil {
		return "", errors.New("Invalid app")
	}

	return useDB.AddAppV3(enterpriseID, app)
}

func UpdateAppV3(enterpriseID string, appID string, app *data.AppDetailV3) (bool, error) {
	err := checkDB()
	if err != nil {
		return false, err
	}

	if enterpriseID == "" {
		return false, errors.New("Invalid enterpriseID")
	}
	if appID == "" {
		return false, errors.New("Invalid appID")
	}

	return useDB.UpdateAppV3(enterpriseID, appID, app)
}

func DeleteAppV3(enterpriseID string, appID string) (bool, error) {
	err := checkDB()
	if err != nil {
		return false, err
	}

	if enterpriseID == "" {
		return false, errors.New("Invalid enterpriseID")
	}
	if appID == "" {
		return false, errors.New("Invalid appID")
	}

	return useDB.DeleteAppV3(enterpriseID, appID)
}

func GetGroupsV3(enterpriseID string) ([]*data.GroupDetailV3, error) {
	err := checkDB()
	if err != nil {
		return nil, err
	}

	if enterpriseID == "" {
		return nil, errors.New("INvalid enterpriseID")
	}

	return useDB.GetGroupsV3(enterpriseID)
}

func GetGroupV3(enterpriseID string, groupID string) (*data.GroupDetailV3, error) {
	err := checkDB()
	if err != nil {
		return nil, err
	}

	if enterpriseID == "" {
		return nil, errors.New("Invalid enterpriseID")
	}
	if groupID == "" {
		return nil, errors.New("Invalid groupID")
	}

	return useDB.GetGroupV3(enterpriseID, groupID)
}

func AddGroupV3(enterpriseID string, group *data.GroupDetailV3,
	apps []string) (groupID string, err error) {
	err = checkDB()
	if err != nil {
		return "", err
	}

	if enterpriseID == "" {
		return "", errors.New("Invalid enterpriseID")
	}
	if group == nil {
		return "", errors.New("Invalid group")
	}

	return useDB.AddGroupV3(enterpriseID, group, apps)
}

func UpdateGroupV3(enterpriseID string, groupID string,
	group *data.GroupDetailV3, apps []string) (bool, error) {
	err := checkDB()
	if err != nil {
		return false, err
	}

	if enterpriseID == "" {
		return false, errors.New("Invalid enterpriseID")
	}
	if groupID == "" {
		return false, errors.New("Invalid groupID")
	}
	if group == nil {
		return false, errors.New("Invalid group")
	}

	return useDB.UpdateGroupV3(enterpriseID, groupID, group, apps)
}

func DeleteGroupV3(enterpriseID string, groupID string) (bool, error) {
	err := checkDB()
	if err != nil {
		return false, err
	}

	if enterpriseID == "" {
		return false, errors.New("Invalid enterpriseID")
	}
	if groupID == "" {
		return false, errors.New("Invalid groupID")
	}

	return useDB.DeleteGroupV3(enterpriseID, groupID)
}

func GetRolesV3(enterpriseID string) ([]*data.RoleV3, error) {
	err := checkDB()
	if err != nil {
		return nil, err
	}

	if enterpriseID == "" {
		return nil, errors.New("Invalid enterpriseID")
	}

	return useDB.GetRolesV3(enterpriseID)
}

func GetRoleV3(enterpriseID string, roleID string) (*data.RoleV3, error) {
	err := checkDB()
	if err != nil {
		return nil, err
	}

	if enterpriseID == "" {
		return nil, errors.New("Invalid enterpriseID")
	}
	if roleID == "" {
		return nil, errors.New("Invalid roleID")
	}

	return useDB.GetRoleV3(enterpriseID, roleID)
}

func AddRoleV3(enterpriseID string, role *data.RoleV3) (string, error) {
	err := checkDB()
	if err != nil {
		return "", err
	}

	if enterpriseID == "" {
		return "", errors.New("Invalid enterpriseID")
	}
	if role == nil {
		return "", errors.New("Invalid role")
	}

	return useDB.AddRoleV3(enterpriseID, role)
}

func UpdateRoleV3(enterpriseID string, roleID string, role *data.RoleV3) (bool, error) {
	err := checkDB()
	if err != nil {
		return false, err
	}

	if enterpriseID == "" {
		return false, errors.New("Invalid enterpriseID")
	}
	if roleID == "" {
		return false, errors.New("Invalid roleID")
	}

	return useDB.UpdateRoleV3(enterpriseID, roleID, role)
}

func DeleteRoleV3(enterpriseID string, roleID string) (bool, error) {
	if useDB == nil {
		return false, errors.New("DB hasn't set")
	}

	if enterpriseID == "" {
		return false, errors.New("Invalid enterpriseID")
	}
	if roleID == "" {
		return false, errors.New("Invalid roleID")
	}

	usersCount, err := useDB.GetUsersCountOfRoleV3(roleID)
	if err != nil {
		return false, err
	}
	
	if usersCount > 0 {
		return false, errors.New("Cannot remove role having user")
	}

	return useDB.DeleteRoleV3(enterpriseID, roleID)
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

	return useDB.GetModulesV3(enterpriseID)
}
