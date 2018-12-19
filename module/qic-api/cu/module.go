package cu

import (
	"database/sql"
	"time"

	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/qic-api/util/timecache"
)

func init() {
	ModuleInfo = util.ModuleInfo{
		ModuleName: "cu",
		EntryPoints: []util.EntryPoint{
			util.NewEntryPoint("POST", "text/process", []string{}, handleTextProcess),
			util.NewEntryPoint("POST", "conversation", []string{}, handleFlowCreate),
			util.NewEntryPoint("POST", "conversation/{id}/append", []string{}, handleFlowAdd),
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

func SetUpTimeCache() {
	config := &timecache.TCacheConfig{}
	config.SetCollectionDuration(30 * time.Second)
	config.SetCollectionMethod(timecache.OnUpdate)
	cache.Activate(config)
}
