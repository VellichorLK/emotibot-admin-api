package qi

import (
	"database/sql"

	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/pkg/logger"
)

var (
	// ModuleInfo is needed for module define
	ModuleInfo util.ModuleInfo
	db         *sql.DB
)

func init() {
	ModuleInfo = util.ModuleInfo{
		ModuleName: "qi",
		EntryPoints: []util.EntryPoint{
			util.NewEntryPoint("POST", "groups", []string{}, handleCreateGroup),
			util.NewEntryPoint("GET", "groups", []string{}, handleGetGroups),
			util.NewEntryPoint("GET", "groups/{id:[0-9]+}", []string{}, handleGetGroup),
			util.NewEntryPoint("PUT", "groups/{id:[0-9]+}", []string{}, handleUpdateGroup),
			util.NewEntryPoint("DELETE", "groups/{id:[0-9]+}", []string{}, handleDeleteGroup),
		},
		OneTimeFunc: map[string]func(){
			"init db": func() {
				envs := ModuleInfo.Environments

				url := envs["MYSQL_URL"]
				user := envs["MYSQL_USER"]
				pass := envs["MYSQL_PASS"]
				db := envs["MYSQL_DB"]

				conn, err := util.InitDB(url, user, pass, db)
				if err != nil {
					logger.Error.Printf("Cannot init qi db, [%s:%s@%s:%s]: %s\n", user, pass, url, db, err.Error())
					return
				}
				serviceDAO = &sqlDAO{conn: conn}
			},
		},
	}
}
