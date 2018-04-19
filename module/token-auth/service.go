package main

import (
	"errors"

	"emotibot.com/emotigo/module/token-auth/dao"
	"emotibot.com/emotigo/module/token-auth/data"
	"emotibot.com/emotigo/module/token-auth/util"
)

var useDB dao.DB

func setDB(db dao.DB) {
	useDB = db
}

func getEnterprises() (*data.Enterprises, string) {
	enterprises, err := useDB.GetEnterprises()
	if err != nil {
		return nil, err.Error()
	}
	return enterprises, ""
}
func getEnterprise(enterpriseID string) (ret *data.Enterprise, err error) {
	defer func() {
		if err != nil {
			util.LogError.Printf("Error when get enterprise %s: %s\n", enterpriseID, err.Error())
		}
	}()
	ret, err = useDB.GetEnterprise(enterpriseID)
	if err != nil {
		return nil, err
	}
	apps, err := useDB.GetApps(enterpriseID)
	if err != nil {
		return nil, err
	}
	adminUser, err := useDB.GetAdminUser(enterpriseID)
	if err != nil {
		return nil, err
	}

	ret.Apps = apps
	ret.AdminUser = adminUser
	return
}
func getUsers(enterpriseID string) (*data.Users, error) {
	users, err := useDB.GetUsers(enterpriseID)
	if err != nil {
		return nil, err
	}
	return users, nil
}
func getUser(enterpriseID string, userID string) (*data.User, error) {
	user, err := useDB.GetUser(enterpriseID, userID)
	if err != nil {
		return nil, err
	}
	return user, nil
}
func getApps(enterpriseID string) (*data.Apps, string) {
	apps, err := useDB.GetApps(enterpriseID)
	if err != nil {
		return nil, err.Error()
	}
	return apps, ""
}
func getApp(enterpriseID string, appID string) (*data.App, string) {
	app, err := useDB.GetApp(enterpriseID, appID)
	if err != nil {
		return nil, err.Error()
	}
	return app, ""
}

func login(account string, passwd string) (*data.Enterprise, *data.User, string) {
	user, err := useDB.GetAuthUser(account, passwd)
	if err != nil {
		return nil, nil, err.Error()
	}

	if user == nil {
		return nil, nil, ""
	}

	var enterprise *data.Enterprise
	if user.Enterprise != nil {
		enterprise, err = useDB.GetEnterprise(*user.Enterprise)
		if err != nil {
			return nil, nil, err.Error()
		}
	}
	return enterprise, user, ""
}

func addUser(enterpriseID string, user *data.User, roleID string) (string, error) {
	return useDB.AddUser(enterpriseID, user, roleID)
}

func deleteUser(enterpriseID string, userID string) error {
	_, err := useDB.DeleteUser(enterpriseID, userID)
	return err
}

func updateUser(enterpriseID string, user *data.User) error {
	return useDB.UpdateUser(enterpriseID, user)
}

func getRoles(enterpriseID string) ([]*data.Role, error) {
	ret, err := useDB.GetRoles(enterpriseID)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func getRole(enterpriseID string, roleID string) (*data.Role, error) {
	ret, err := useDB.GetRole(enterpriseID, roleID)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func deleteRole(enterpriseID string, roleID string) (bool, error) {
	ret, err := useDB.DeleteRole(enterpriseID, roleID)
	if err != nil {
		return false, err
	}
	return ret, nil
}

func addRole(enterpriseID string, role *data.Role) (string, error) {
	return useDB.AddRole(enterpriseID, role)
}

func updateRole(enterpriseID string, roleID string, role *data.Role) (bool, error) {
	roles, err := useDB.GetUsersOfRole(enterpriseID, roleID)
	if err != nil {
		return false, err
	}
	if roles != nil && len(*roles) > 0 {
		return false, errors.New("Cannot remove role having user")
	}
	return useDB.UpdateRole(enterpriseID, roleID, role)
}

func getModules(enterpriseID string) ([]*data.Module, error) {
	ret, err := useDB.GetModules(enterpriseID)
	if err != nil {
		return nil, err
	}
	return ret, nil
}
