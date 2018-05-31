package Stats

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"emotibot.com/emotigo/module/admin-api/util"
)

// func TestRetrieveHeader(t *testing.T) {
// 	var table = RobotTrafficRow{}
// 	var cHeaders, err = RetrieveHeader(&table)
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	log.Printf("%+v\n", cHeaders)
// 	json.Marshal(table)
// }

func TestGetStatsTable(t *testing.T) {
	db, err := sql.Open("mysql", "root:password@tcp(172.16.101.47:3306)/backend_log?parseTime=true&loc=Asia%2FShanghai")
	if err != nil {
		t.Fatal(err)
	}
	util.SetDB(ModuleInfo.ModuleName, db)
	st := RobotTrafficsTable
	theDay := time.Now()
	if err != nil {
		t.Fatal(err)
	}
	beforeTheDay := theDay.AddDate(0, 0, -30)
	d, err := st.GetGroupedRows("vipshop", 1, "name", nil, beforeTheDay, theDay)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(theDay)
	// fmt.Println(beforeTheDay.Unix())
	ds, _ := json.Marshal(d)
	fmt.Println(string(ds))
}
