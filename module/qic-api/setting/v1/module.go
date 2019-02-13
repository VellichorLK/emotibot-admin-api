// Package setting is the http api of setting subdomain and it's related functions.
package setting

import (
	"net/http"

	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/pkg/logger"
)

var (
	//ModuleInfo is the main entrypoint of the setting package
	ModuleInfo util.ModuleInfo
	userkeyDao = &model.UserKeySQLDao{}
	db         model.DBLike
)

func init() {
	ModuleInfo = util.ModuleInfo{
		ModuleName: "setting",
		EntryPoints: []util.EntryPoint{
			util.NewEntryPoint(http.MethodGet, "custom-column", []string{}, GetCustomColsHandler),
			util.NewEntryPoint(http.MethodPost, "custom-column", []string{}, CreateCustomColHandler),
			util.NewEntryPoint(http.MethodDelete, "custom-column/{col_inputname}", []string{}, DeleteCustomColHandler),
		},
		OneTimeFunc: map[string]func(){
			"init db": func() {
				logger.Info.Println("start init setting db")
				envs := ModuleInfo.Environments
				url := envs["MYSQL_URL"]
				user := envs["MYSQL_USER"]
				pass := envs["MYSQL_PASS"]
				dbName := envs["MYSQL_DB"]
				newConn, err := util.InitDB(url, user, pass, dbName)
				if err != nil {
					logger.Error.Println("init setting db failed, ", err)
				}
				userkeyDao = model.NewUserKeyDao(newConn)
				db = &model.DefaultDBLike{
					DB: newConn,
				}
				newUserKey = userkeyDao.NewUserKey
				userKeys = userkeyDao.UserKeys
				countUserKeys = userkeyDao.CountUserKeys
				deleteUserKey = userkeyDao.DeleteUserKeys
				logger.Info.Println("init setting db succeed")
			},
		},
	}
}
