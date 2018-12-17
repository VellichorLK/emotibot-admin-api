package cu

import (
	"database/sql"

	"emotibot.com/emotigo/module/admin-api/util"
)

func init() {
	ModuleInfo = util.ModuleInfo{
		ModuleName: "cu",
		EntryPoints: []util.EntryPoint{
			util.NewEntryPoint("POST", "text/process", []string{}, handleTextProcess),
			util.NewEntryPoint("POST", "conversation", []string{}, handleFlowCreate),
		},
	}
	maxDirDepth = 4
}

//SetupServiceDB sets up the db structure
func SetupServiceDB(db *sql.DB) {
	serviceDao = SQLDao{
		conn: db,
	}
}
