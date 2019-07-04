package service

import (
	"errors"

	"emotibot.com/emotigo/module/token-auth/dao"
	"emotibot.com/emotigo/module/token-auth/internal/data"
	"emotibot.com/emotigo/module/token-auth/internal/util"
)

var useDBV4 dao.DBV4

// SetDBV4 will setup a db implement interface of dao.DBV4
func SetDBV4(db dao.DBV4) {
	useDBV4 = db
}

// CheckOauthLoginRequest will check if clientID and redirectURL is setup in system or not
// If clientID and redirectURL is valid, return true.
func CheckOauthLoginRequest(clientID, redirectURL string) (bool, error) {
	if useDBV4 == nil {
		return false, errors.New("DB hasn't set")
	}

	oauthClient, err := useDBV4.GetOAuthClient(clientID)
	if err != nil {
		return false, err
	}

	if oauthClient == nil {
		return false, nil
	}

	if oauthClient.ID != clientID ||
		oauthClient.RedirectURI != redirectURL ||
		!oauthClient.Active {
		return false, nil
	}

	return true, nil
}

// CheckOauthRequest will check if clientID, clientSecret and redirectURL is setup in system or not
// If it is valid, return true.
func CheckOauthRequest(clientID, clientSecret, redirectURL string) (bool, error) {
	if useDBV4 == nil {
		return false, errors.New("DB hasn't set")
	}

	oauthClient, err := useDBV4.GetOAuthClient(clientID)
	if err != nil {
		return false, err
	}

	if oauthClient == nil {
		return false, nil
	}

	if oauthClient.ID != clientID ||
		oauthClient.RedirectURI != redirectURL ||
		oauthClient.Secret != clientSecret ||
		!oauthClient.Active {
		return false, nil
	}

	return true, nil
}

// AddAppV4 is to add app to service with app_type info
func AddAppV4(enterpriseID string, app *data.AppDetailV4) (appID string, err error) {
	err = checkDB()
	if err != nil {
		return "", err
	}

	exists, err := useDBV3.EnterpriseExistsV3(enterpriseID)
	if err != nil {
		return "", err
	} else if !exists {
		return "", nil
	}

	exists, err = useDBV3.EnterpriseAppInfoExistsV3(enterpriseID, app.Name)
	if err != nil {
		return "", err
	} else if exists {
		return "", util.ErrAppInfoExists
	}

	return useDBV4.AddAppV4(enterpriseID, app)
}

// AddEnterpriseV4 will add enterprise into system. If dryRun is true, it will only run for check
func AddEnterpriseV4(enterprise *data.EnterpriseV3, modules []string,
	adminUser *data.UserDetailV3, dryRun bool, active bool) (enterpriseID string, err error) {
	err = checkDB()
	if err != nil {
		return "", err
	}

	exists, err := useDBV3.EnterpriseInfoExistsV3(enterprise.Name)
	if err != nil {
		return "", err
	} else if exists {
		return "", util.ErrEnterpriseInfoExists
	}

	return useDBV4.AddEnterpriseV4(enterprise, modules, adminUser, dryRun, active)
}

func GetAppsV4(enterpriseID string) ([]*data.AppDetailV4, error) {
	err := checkDB()
	if err != nil {
		return nil, err
	}

	exists, err := useDBV3.EnterpriseExistsV3(enterpriseID)
	if err != nil {
		return nil, err
	} else if !exists {
		return nil, nil
	}

	return useDBV4.GetAppsV4(enterpriseID)
}

func UpdateEnterpriseStatusV4(enterpriseID string, status bool) error {
	return useDBV4.UpdateEnterpriseStatusV4(enterpriseID, status)
}

func ActivateEnterpriseV4(enterpriseID string, username string, password string) error {
	return useDBV4.ActivateEnterpriseV4(enterpriseID, username, password)
}

func GetModulesV4(enterpriseID string) ([]*data.ModuleDetailV4, error) {
	err := checkDB()
	if err != nil {
		return nil, err
	}

	exists, err := useDBV3.EnterpriseExistsV3(enterpriseID)
	if err != nil {
		return nil, err
	} else if !exists {
		return nil, nil
	}

	return useDBV4.GetModulesV4(enterpriseID)
}

func GetRolesV4(enterpriseID string) ([]*data.RoleV4, error) {
	err := checkDB()
	if err != nil {
		return nil, err
	}

	exists, err := useDBV3.EnterpriseExistsV3(enterpriseID)
	if err != nil {
		return nil, err
	} else if !exists {
		return nil, nil
	}

	return useDBV4.GetRolesV4(enterpriseID)
}
