package BF

import (
	"errors"

	"emotibot.com/emotigo/module/admin-api/auth"
)

func GetSSMCategories(appid string) (*Category, error) {
	return getSSMCategories(appid, false)
}

func GetSSMLabels(appid string) ([]*SSMLabel, error) {
	return getSSMLabels(appid)
}

func GetBFAccessToken(userid string) (string, error) {
	return getBFAccessToken(userid)
}

func GetNewAccessToken() (string, error) {
	admins, err := auth.GetSystemAdminID()
	if err != nil {
		return "", err
	}
	if len(admins) == 0 {
		return "", errors.New("Need one system admin for api-key usage")
	}
	return getBFAccessToken(admins[0])
}

func GetNewAccessTokenOfAppid(appid string) (string, error) {
	admins, err := auth.GetEnterpriseAdminOfRobot(appid)
	if err != nil {
		return "", err
	}
	if len(admins) == 0 {
		return "", errors.New("Need one system admin for api-key usage")
	}
	return getBFAccessToken(admins[0])
}
