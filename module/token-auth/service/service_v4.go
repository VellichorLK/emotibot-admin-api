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

func UpdateEnterpriseStatusV4(enterpriseID string, status bool) error {
	return useDBV4.UpdateEnterpriseStatusV4(enterpriseID, status)
}
