package Stats

import (
	"fmt"
	"time"

	"emotibot.com/emotigo/module/admin-api/util"
)

//StatTable represent a sql table as a HTML table.
type StatTable struct {
	Columns []Column
	Name    string
}

var typDict = map[string]int{
	"all":      0,
	"platform": 1,
	"brand":    2,
	"category": 3,
}

//RobotTrafficRow 代表机器人服务量统计的快取表
var RobotTrafficsTable = StatTable{
	Name: "robot_traffic_stats",
	Columns: []Column{
		Column{ID: "name", Text: "名稱", typ: str},
		Column{ID: "unique_users", Text: "接入客户量", typ: integer},
		Column{ID: "effective_users", Text: "有效处理客户量", typ: integer},
		Column{ID: "total_messages", Text: "接入消息量", typ: integer},
		Column{ID: "resolved_messages", Text: "有效處理消息量", typ: integer},
		Column{ID: "unresolved_messages", Text: "轉人工量", typ: integer},
		Column{ID: "resolved_rate", Text: "成功解決率", typ: flt},
	},
}

var RobotResponseTable = StatTable{
	Name: "robot_response_stats",
	Columns: []Column{
		Column{ID: "name", Text: "名稱", typ: str},
		Column{ID: "precision_match", Text: "精確匹配", typ: integer},
		Column{ID: "fuzzy_match", Text: "模糊匹配", typ: integer},
		Column{ID: "default_match", Text: "默認回覆", typ: integer},
		Column{ID: "system_errors", Text: "系統異常", typ: integer},
		Column{ID: "sensitive_match", Text: "敏感詞", typ: integer},
		Column{ID: "business_default_match", Text: "業務默認回覆", typ: integer},
		Column{ID: "chat_module", Text: "寒暄", typ: integer},
		Column{ID: "on_list_match", Text: "列表回覆", typ: integer},
		Column{ID: "title_match", Text: "標題提問", typ: integer},
		Column{ID: "meaningless_responses", Text: "無意義回覆", typ: integer},
		Column{ID: "common_responses", Text: "通用句式", typ: integer},
		Column{ID: "direction_responses", Text: "引導用句", typ: integer},
		Column{ID: "unknown_responses", Text: "未知回答", typ: integer},
	},
}

var HourlyMonitortable = StatTable{
	Name: "monitor_hourly",
	Columns: []Column{
		Column{ID: "cache_hour", Text: "时间", typ: datetime},
		Column{ID: "manually_users", Text: "人工接入客户量", typ: integer},
		Column{ID: "manually_messages", Text: "人工接入会话量", typ: integer},
		Column{ID: "unique_users", Text: "機器人接入客户量", typ: integer},
		Column{ID: "total_messages", Text: "機器人接入會話量", typ: integer},
		Column{ID: "unresolved_rate", Text: "轉人工率", typ: flt},
	},
}

var DailyMonitorTable = StatTable{
	Name: "monitor_daily",
	Columns: []Column{
		Column{ID: "cache_day", Text: "時間", typ: datetime},
		Column{ID: "manually_users", Text: "人工接入客户量", typ: integer},
		Column{ID: "manually_messages", Text: "人工接入会话量", typ: integer},
		Column{ID: "unique_users", Text: "機器人接入客户量", typ: integer},
		Column{ID: "total_messages", Text: "機器人接入會話量", typ: integer},
		Column{ID: "unresolved_rate", Text: "轉人工率", typ: flt},
	},
}

var UserContactsTable = StatTable{
	Name: "user_contacts",
	Columns: []Column{
		Column{ID: "name", Text: "渠道", typ: str},
		Column{ID: "user_id", Text: "用戶ID", typ: str},
		Column{ID: "last_chat", Text: "最後訪問時間", typ: datetime},
	},
}

var FAQStatsTable = StatTable{
	Name: "faq_stats",
	Columns: []Column{
		Column{ID: "question_name", Text: "标准问题", typ: str},
		Column{ID: "brand", Text: "渠道", typ: str},
		Column{ID: "categoryL1", Text: "业务分类一级", typ: str},
		Column{ID: "categoryL2", Text: "业务分类二级", typ: str},
		Column{ID: "categoryL3", Text: "业务分类三级", typ: str},
		Column{ID: "categoryL4", Text: "业务分类四级", typ: str},
		Column{ID: "categoryL5", Text: "业务分类五级", typ: str},
		Column{ID: "total_count", Text: "访问总数", typ: str},
		Column{ID: "hit_count", Text: "命中次数", typ: str},
		Column{ID: "accuracy", Text: "命中率", typ: str},
	},
}

var ChatRecordTable = StatTable{
	Name: "record",
	Columns: []Column{
		Column{ID: "user_id", Text: "用戶ID"},
		Column{ID: "name", Text: "渠道"},
		Column{ID: "input", Text: "用戶問"},
		Column{ID: "output", Text: "機器人答"},
		Column{ID: "conversation_time", Text: "對話時間"},
	},
}

type statsRow map[string]interface{}

//Column represent StatTable's column header and it's data type
//	ID should be exact match to sql table column.
//	Text will be used on UI representation.
//	typ should defined which value it should hold, list should have all type we defined. default: string
type Column struct {
	ID   string `json:"id"`
	Text string `json:"text"`
	typ  int
}

//We have define three kind of data type will be in row's value
const (
	str = iota
	integer
	flt
	datetime
)

//GetGroupedRows will retrive StatTable's cache data group by groupCol.ID
func (st StatTable) GetGroupedRows(appID string, typ int, groupColName string, avgCol []string, start, end time.Time) ([]statsRow, error) {
	var sqlCol string
	for i, col := range st.Columns {
		if i != 0 {
			sqlCol += ", "
		}
		if col.typ == integer || col.typ == flt {
			var shouldAvg bool
			for _, colID := range avgCol {
				if colID == col.ID {
					shouldAvg = true
					break
				}
			}
			if shouldAvg {
				sqlCol += "ROUND(AVG(" + col.ID + "), 4)"
			} else {
				sqlCol += "SUM(" + col.ID + ")"
			}
		} else {
			sqlCol += col.ID
		}
	}

	query := "SELECT " + sqlCol + " FROM " + st.Name + " WHERE (cache_day BETWEEN ? AND ?) AND type = ? AND app_id = ? GROUP BY " + groupColName
	db := util.GetDB(ModuleInfo.ModuleName)
	if db == nil {
		return nil, fmt.Errorf("can not get db of " + ModuleInfo.ModuleName)
	}
	rows, err := db.Query(query, start, end, typ, appID)
	if err != nil {
		return nil, fmt.Errorf("query failed, %v\n origin query: %s", err, query)
	}
	defer rows.Close()

	var tRows = []statsRow{}
	for rows.Next() {
		var tRow = make(map[string]interface{}, len(st.Columns))
		values := make([]interface{}, 0, len(st.Columns))
		for _, col := range st.Columns {
			switch col.typ {
			case str:
				values = append(values, new(string))
			case integer:
				values = append(values, new(int64))
			case flt:
				values = append(values, new(float64))
			case datetime:
				values = append(values, &time.Time{})
			}
		}

		rows.Scan(values...)
		for i, v := range values {
			tRow[st.Columns[i].ID] = v
		}
		tRows = append(tRows, tRow)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("scan err" + err.Error())
	}

	return tRows, nil
}

// NewStatsSelector return a selector can be called to select every rows.
func NewStatsSelector(st StatTable, dateCol string) func(appID string, start, end time.Time, eq ...whereEqual) ([]statsRow, error) {
	return func(appID string, start, end time.Time, eq ...whereEqual) ([]statsRow, error) {
		var sqlCol string
		for index, col := range st.Columns {
			if index != 0 {
				sqlCol += ", "
			}
			sqlCol += col.ID
		}
		query := "SELECT " + sqlCol + " FROM " + st.Name + " WHERE (" + dateCol + " BETWEEN ? AND ?) AND app_id = ? "
		var input = []interface{}{start, end, appID}
		for _, e := range eq {
			query += " AND " + e.ColName + " = ?"
			input = append(input, e.value)
		}
		db := util.GetDB(ModuleInfo.ModuleName)
		if db == nil {
			return nil, fmt.Errorf("can not get db of " + ModuleInfo.ModuleName)
		}

		rows, err := db.Query(query, input...)
		if err != nil {
			util.LogError.Printf("query failed, detail query: %s. input: %v \n", query, input)
			return nil, fmt.Errorf("query failed, %v", err)
		}
		defer rows.Close()
		var tRows = []statsRow{}
		for rows.Next() {
			var tRow = make(map[string]interface{}, len(st.Columns))
			values := make([]interface{}, 0, len(st.Columns))
			for _, col := range st.Columns {
				switch col.typ {
				case str:
					values = append(values, new(string))
				case integer:
					values = append(values, new(int64))
				case flt:
					values = append(values, new(float64))
				case datetime:
					values = append(values, &time.Time{})
				}
			}

			rows.Scan(values...)
			for i, v := range values {
				tRow[st.Columns[i].ID] = v
			}
			tRows = append(tRows, tRow)
		}
		if err = rows.Err(); err != nil {
			return nil, fmt.Errorf("scan err" + err.Error())
		}

		return tRows, nil
	}

}

type whereEqual struct {
	ColName string
	value   interface{}
}
