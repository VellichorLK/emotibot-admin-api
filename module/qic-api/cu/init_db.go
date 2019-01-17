package cu

import (
	"database/sql"

	"emotibot.com/emotigo/module/admin-api/util"
)

//GetDB gets the db with this module name
func GetDB() *sql.DB {
	return util.GetDB(ModuleInfo.ModuleName)
}
