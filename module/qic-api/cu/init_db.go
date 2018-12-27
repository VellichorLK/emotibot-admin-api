package cu

import (
	"database/sql"

	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/pkg/logger"
)

//InitDB uses the env as the connection setting
func InitDB() error {

	//envs := config.GetEnvOf("admin")

	envs := ModuleInfo.Environments

	url := envs["MYSQL_URL"]
	user := envs["MYSQL_USER"]
	pass := envs["MYSQL_PASS"]
	db := envs["MYSQL_DB"]
	dao, err := util.InitDB(url, user, pass, db)
	if err != nil {
		logger.Error.Printf("Cannot init qi db, [%s:%s@%s:%s]: %s\n", user, pass, url, db, err.Error())
		return err
	}
	util.SetDB(ModuleInfo.ModuleName, dao)
	SetupServiceDB(dao)
	SetUpTimeCache()
	return nil
}

//GetDB gets the db with this module name
func GetDB() *sql.DB {
	return util.GetDB(ModuleInfo.ModuleName)
}
