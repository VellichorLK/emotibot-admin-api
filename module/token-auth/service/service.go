package service

import (
	"database/sql"
	"errors"

	"emotibot.com/emotigo/module/token-auth/dao"
	"emotibot.com/emotigo/module/token-auth/internal/data"
	"emotibot.com/emotigo/module/token-auth/internal/util"
)

var useDB dao.DB

func SetDB(db dao.DB) {
	useDB = db
}

func GetEnterprises() (*data.Enterprises, error) {
	if useDB == nil {
		return nil, errors.New("DB hasn't set")
	}
	return useDB.GetEnterprises()
}
func GetEnterprise(enterpriseID string) (ret *data.Enterprise, err error) {
	if useDB == nil {
		return nil, errors.New("DB hasn't set")
	}
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
func GetUsers(enterpriseID string) (*data.Users, error) {
	if useDB == nil {
		return nil, errors.New("DB hasn't set")
	}
	users, err := useDB.GetUsers(enterpriseID)
	if err != nil {
		return nil, err
	}
	return users, nil
}
func GetUser(enterpriseID string, userID string) (*data.User, error) {
	if useDB == nil {
		return nil, errors.New("DB hasn't set")
	}
	user, err := useDB.GetUser(enterpriseID, userID)
	if err != nil {
		return nil, err
	}
	return user, nil
}
func GetApps(enterpriseID string) (*data.Apps, string) {
	if useDB == nil {
		return nil, "DB hasn't set"
	}
	apps, err := useDB.GetApps(enterpriseID)
	if err != nil {
		return nil, err.Error()
	}
	return apps, ""
}
func GetApp(enterpriseID string, appID string) (*data.App, string) {
	if useDB == nil {
		return nil, "DB hasn't set"
	}
	app, err := useDB.GetApp(enterpriseID, appID)
	if err != nil {
		return nil, err.Error()
	}
	return app, ""
}

func Login(account string, passwd string) (*data.Enterprise, *data.User, string) {
	if useDB == nil {
		return nil, nil, "DB hasn't set"
	}
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

func AddUser(enterpriseID string, user *data.User, roleID string) (string, error) {
	if useDB == nil {
		return "", errors.New("DB hasn't set")
	}
	return useDB.AddUser(enterpriseID, user, roleID)
}

func DeleteUser(enterpriseID string, userID string) error {
	if useDB == nil {
		return errors.New("DB hasn't set")
	}
	_, err := useDB.DeleteUser(enterpriseID, userID)
	return err
}

func UpdateUser(enterpriseID string, user *data.User) error {
	if useDB == nil {
		return errors.New("DB hasn't set")
	}
	return useDB.UpdateUser(enterpriseID, user)
}

func GetRoles(enterpriseID string) ([]*data.Role, error) {
	if useDB == nil {
		return nil, errors.New("DB hasn't set")
	}
	ret, err := useDB.GetRoles(enterpriseID)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func GetRole(enterpriseID string, roleID string) (*data.Role, error) {
	if useDB == nil {
		return nil, errors.New("DB hasn't set")
	}
	ret, err := useDB.GetRole(enterpriseID, roleID)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func DeleteRole(enterpriseID string, roleID string) (bool, error) {
	if useDB == nil {
		return false, errors.New("DB hasn't set")
	}
	users, err := useDB.GetUsersOfRole(enterpriseID, roleID)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}
	if users != nil && len(*users) > 0 {
		return false, errors.New("Cannot remove role having user")
	}
	ret, err := useDB.DeleteRole(enterpriseID, roleID)
	if err != nil {
		return false, err
	}
	return ret, nil
}

func AddRole(enterpriseID string, role *data.Role) (string, error) {
	if useDB == nil {
		return "", errors.New("DB hasn't set")
	}
	return useDB.AddRole(enterpriseID, role)
}

func UpdateRole(enterpriseID string, roleID string, role *data.Role) (bool, error) {
	if useDB == nil {
		return false, errors.New("DB hasn't set")
	}
	return useDB.UpdateRole(enterpriseID, roleID, role)
}

func GetModules(enterpriseID string) ([]*data.Module, error) {
	if useDB == nil {
		return nil, errors.New("DB hasn't set")
	}
	ret, err := useDB.GetModules(enterpriseID)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func AddEnterprise(enterprise *data.Enterprise, adminUser *data.User) (enterpriseID string, err error) {
	if useDB == nil {
		return "", errors.New("DB hasn't set")
	}
	if enterprise == nil || adminUser == nil {
		return "", errors.New("Parameter error")
	}
	if !adminUser.IsValid() {
		return "", errors.New("Invalid user")
	}
	return useDB.AddEnterprise(enterprise, adminUser)
}

func AddApp(enterpriseID string, app *data.App) (appid string, err error) {
	if useDB == nil {
		return "", errors.New("DB hasn't set")
	}
	if enterpriseID == "" {
		return "", errors.New("Invalid enterpriseID")
	}
	if app == nil {
		return "", errors.New("Invalid app")
	}
	return useDB.AddApp(enterpriseID, app)
}

func DeleteEnterprise(enterpriseID string) error {
	if useDB == nil {
		return errors.New("DB hasn't set")
	}
	if enterpriseID == "" {
		return errors.New("Invalid enterpriseID")
	}
	_, err := useDB.DeleteEnterprise(enterpriseID)
	return err
}
