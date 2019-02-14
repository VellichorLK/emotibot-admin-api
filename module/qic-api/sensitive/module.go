package sensitive

import (
	"database/sql"

	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/pkg/logger"
)

var (
	// ModuleInfo is needed for module define
	ModuleInfo  util.ModuleInfo
	swDao       model.SensitiveWordDao
	sentenceDao model.SentenceDao
	sqlConn     *sql.DB
	dbLike      model.DBLike
)

func init() {
	ModuleInfo = util.ModuleInfo{
		ModuleName: "sensitive",
		EntryPoints: []util.EntryPoint{
			util.NewEntryPoint("POST", "word", []string{}, handleCreateSensitiveWord),

			// category apis
			util.NewEntryPoint("POST", "category", []string{}, handleCreateSensitiveWordCategory),
		},
		OneTimeFunc: map[string]func(){
			"init db": func() {
				envs := ModuleInfo.Environments

				url := envs["MYSQL_URL"]
				user := envs["MYSQL_USER"]
				pass := envs["MYSQL_PASS"]
				db := envs["MYSQL_DB"]

				newConn, err := util.InitDB(url, user, pass, db)
				sqlConn = newConn
				if err != nil {
					logger.Error.Printf("Cannot init sensitive db, [%s:%s@%s:%s]: %s\n", user, pass, url, db, err.Error())
					return
				}

				dbLike = &model.DefaultDBLike{
					DB: sqlConn,
				}

				swDao = &model.SensitiveWordSqlDao{}
				sentenceDao = model.NewSentenceSQLDao(sqlConn)
			},
		},
	}
}
