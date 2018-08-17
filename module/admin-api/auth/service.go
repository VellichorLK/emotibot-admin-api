package auth

import (
	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/pkg/logger"
)

var (
	// ModuleInfo is needed for module define
	ModuleInfo util.ModuleInfo
)

func init() {
	ModuleInfo = util.ModuleInfo{
		ModuleName:  "auth",
		EntryPoints: []util.EntryPoint{},
	}
}

func InitDB() {
	envs := getEnvironments()
	url := envs["MYSQL_URL"]
	user := envs["MYSQL_USER"]
	pass := envs["MYSQL_PASS"]
	db := envs["MYSQL_DB"]
	dao, err := util.InitDB(url, user, pass, db)
	if err != nil {
		logger.Error.Printf("Cannot init auth db, [%s:%s@%s:%s]: %s\n", user, pass, url, db, err.Error())
	}

	util.SetDB(ModuleInfo.ModuleName, dao)
}

func getEnvironments() map[string]string {
	return util.GetEnvOf(ModuleInfo.ModuleName)
}
