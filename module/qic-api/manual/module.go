package manual

import (
	_ "database/sql"

	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/pkg/logger"
)

var (
	// ModuleInfo is needed for module define
	ModuleInfo util.ModuleInfo
	// tagDao     TagDao
	// sqlConn    *sql.DB
	manualDB model.DBLike
	authDB   model.DBLike
)

func init() {
	ModuleInfo = util.ModuleInfo{
		ModuleName: "manual",
		EntryPoints: []util.EntryPoint{
			util.NewEntryPoint("POST", "sampling/tasks", []string{}, handleCreateTask),
			util.NewEntryPoint("GET", "sampling/tasks", []string{}, handleGetTasks),
			util.NewEntryPoint("GET", "sampling/tasks/{id}", []string{}, handleGetTask),
			util.NewEntryPoint("PATCH", "sampling/tasks/{id}", []string{}, handleUpdateTask),
		},
		OneTimeFunc: map[string]func(){
			"init db": func() {
				envs := ModuleInfo.Environments

				// init qi db connection
				url := envs["MYSQL_URL"]
				user := envs["MYSQL_USER"]
				pass := envs["MYSQL_PASS"]
				db := envs["MYSQL_DB"]

				newConn, err := util.InitDB(url, user, pass, db)
				if err != nil {
					logger.Error.Printf("Cannot init qi db, [%s:%s@%s:%s]: %s\n", user, pass, url, db, err.Error())
					return
				}

				manualDB = &model.DefaultDBLike{
					DB: newConn,
				}

				// init auth connection
				url = envs["AUTH_MYSQL_URL"]
				user = envs["AUTH_MYSQL_USER"]
				pass = envs["AUTH_MYSQL_PASS"]
				db = envs["AUTH_MYSQL_DB"]

				newConn, err = util.InitDB(url, user, pass, db)
				if err != nil {
					logger.Error.Printf("Cannot init auth db, [%s:%s@%s:%s]: %s\n", user, pass, url, db, err.Error())
					return
				}

				authDB = &model.DefaultDBLike{
					DB: newConn,
				}
			},
		},
	}
}
