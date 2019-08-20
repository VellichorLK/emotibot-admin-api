package service

import (
	"emotibot.com/emotigo/module/token-auth/dao"
	"emotibot.com/emotigo/module/token-auth/internal/data"
	"emotibot.com/emotigo/module/token-auth/internal/util"
)

var useDBV5 dao.DBV5

func SetDBV5(db dao.DBV5) {
	useDBV5 = db
}

// AddAppV5 is to add app to service with chat_lang info
func AddAppV5(enterpriseID string, app *data.AppDetailV5) (appID string, err error) {
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

	return useDBV5.AddAppV5(enterpriseID, app)
}

func AppPropsGetV5(pKey string) ([]*data.AppPropV5, error) {
	return useDBV5.AppPropsGetV5(pKey)
}

func GetAppV5(enterpriseID string, appID string) (*data.AppDetailV5, error) {
	err := checkDB()
	if err != nil {
		return nil, err
	}

	return useDBV5.GetAppV5(enterpriseID, appID)
}

func GetAppsV5(enterpriseID string) ([]*data.AppDetailV5, error) {
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

	return useDBV5.GetAppsV5(enterpriseID)
}
