package qi

import (
	"database/sql"

	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/pkg/logger"
)

var (
	// ModuleInfo is needed for module define
	ModuleInfo util.ModuleInfo
	tagDao     TagDao
	sqlConn    *sql.DB
	dbLike     model.DBLike
)

func init() {
	ModuleInfo = util.ModuleInfo{
		ModuleName: "qi",
		EntryPoints: []util.EntryPoint{
			util.NewEntryPoint("POST", "groups", []string{}, handleCreateGroup),
			util.NewEntryPoint("GET", "groups", []string{}, handleGetGroups),
			util.NewEntryPoint("GET", "groups/filters", []string{}, handleGetGroupsByFilter),
			util.NewEntryPoint("GET", "groups/{id}", []string{}, handleGetGroup),
			util.NewEntryPoint("PUT", "groups/{id}", []string{}, handleUpdateGroup),
			util.NewEntryPoint("DELETE", "groups/{id}", []string{}, handleDeleteGroup),
		},
		OneTimeFunc: map[string]func(){
			"init db": func() {
				envs := ModuleInfo.Environments

				url := envs["MYSQL_URL"]
				user := envs["MYSQL_USER"]
				pass := envs["MYSQL_PASS"]
				db := envs["MYSQL_DB"]

				sqlConn, err := util.InitDB(url, user, pass, db)
				if err != nil {
					logger.Error.Printf("Cannot init qi db, [%s:%s@%s:%s]: %s\n", user, pass, url, db, err.Error())
					return
				}
				dbLike = &model.DefaultDBLike{
					DB: sqlConn,
				}

				serviceDAO = model.NewGroupSQLDao(sqlConn)
				tagDao, err = model.NewTagSQLDao(sqlConn)
				if err != nil {
					logger.Error.Printf("init tag dao failed, %v", err)
					return
				}
			},
		},
	}
}
