package service

import (
	"errors"

	"emotibot.com/emotigo/module/token-auth/dao"
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
