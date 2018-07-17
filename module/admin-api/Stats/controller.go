package Stats

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"emotibot.com/emotigo/module/admin-api/ApiError"
	"emotibot.com/emotigo/module/admin-api/util"
	"github.com/gorilla/mux"
)

var (
	// ModuleInfo is needed for module define
	ModuleInfo   util.ModuleInfo
	cache        map[string]*StatRet
	cacheTimeout *time.Time
)

const (
	DefaultListPerPage = 20
)

func init() {
	ModuleInfo = util.ModuleInfo{
		ModuleName: "statistic",
		EntryPoints: []util.EntryPoint{
			util.NewEntryPoint("POST", "audit", []string{"view"}, handleListAudit),

			util.NewEntryPoint("GET", "question", []string{"view"}, handleQuestionStatistic),

			util.NewEntryPoint("GET", "dialogOneDay", []string{"view"}, handleDialogOneDayStatistic),
			util.NewEntryPoint("GET", "traffics", []string{"view"}, handleRobotsTraffic),
			util.NewEntryPoint("GET", "responses", []string{"view"}, handleRobotsResponse),
			util.NewEntryPoint("GET", "brands/{name}/detail", []string{"view"}, handleMonitor),
			util.NewEntryPoint("GET", "users/last_visit", []string{"view"}, handleLastVisit),
			util.NewEntryPoint("GET", "users/{uid}/records", []string{"view"}, handleRecords),
			util.NewEntryPoint("GET", "faq", []string{"view"}, handleFAQStats),
			util.NewEntryPoint("POST", "sessions/query", []string{}, handleSessionQuery),
		},
	}
	cacheTimeout = nil
	cache = make(map[string]*StatRet)
}

func InitDB() {
	url := getEnvironment("MYSQL_URL")
	user := getEnvironment("MYSQL_USER")
	pass := getEnvironment("MYSQL_PASS")
	db := getEnvironment("MYSQL_DB")
	dao, err := initStatDB(url, user, pass, db)
	if err != nil {
		util.LogError.Printf("Cannot init statistic db, [%s:%s@%s:%s]: %s\n", user, pass, url, db, err.Error())
	}

	util.SetDB(ModuleInfo.ModuleName, dao)
}

func getStatsDB() *sql.DB {
	return util.GetDB(ModuleInfo.ModuleName)
}

func getEnvironments() map[string]string {
	return util.GetEnvOf(ModuleInfo.ModuleName)
}

func getEnvironment(key string) string {
	envs := util.GetEnvOf(ModuleInfo.ModuleName)
	if envs != nil {
		if val, ok := envs[key]; ok {
			return val
		}
	}
	return ""
}

func getGlobalEnv(key string) string {
	envs := util.GetEnvOf("server")
	if envs != nil {
		if val, ok := envs[key]; ok {
			return val
		}
	}
	return ""
}

func handleListAudit(w http.ResponseWriter, r *http.Request) {
	appid := util.GetAppID(r)
	input, err := loadFilter(r)
	if err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	ret, errCode, err := GetAuditList(appid, input)
	if err != nil {
		util.WriteJSON(w, util.GenRetObj(errCode, err.Error()))
	} else {
		util.WriteJSON(w, util.GenRetObj(errCode, ret))
	}
}

func loadFilter(r *http.Request) (*AuditInput, error) {
	input := &AuditInput{}
	err := util.ReadJSON(r, input)
	if err != nil {
		return nil, err
	}

	if input.Filter != nil {
		input.Filter.Module = strings.Trim(input.Filter.Module, " ")
		input.Filter.Operation = strings.Trim(input.Filter.Operation, " ")
		input.Filter.UserID = strings.Trim(input.Filter.UserID, " ")
	}

	if input.Page == 0 {
		input.Page = 1
	}

	if input.ListPerPage == 0 {
		input.ListPerPage = DefaultListPerPage
	}

	if input.End == 0 {
		input.End = int(time.Now().Unix())
	}

	if input.Start == 0 {
		input.Start = input.End - 60*60*24
	}
	return input, nil
}

func handleQuestionStatistic(w http.ResponseWriter, r *http.Request) {
	// user query, standard query, score, count
	day, qType, err := getQuestionParam(r)
	if err != nil {
		http.Error(w, "", http.StatusBadRequest)
		util.WriteJSON(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()))
		return
	}

	appid := util.GetAppID(r)

	util.LogTrace.Printf("Request Questions Statistic: days=[%d] type=[%s]", day, qType)

	code := ApiError.SUCCESS
	ret := getRetInCache(day, qType)
	if ret == nil {
		ret, code, err = GetQuestionStatisticResult(appid, day, qType)
		setRetCache(day, qType, ret)
	}
	if err != nil {
		util.WriteJSON(w, util.GenRetObj(code, err.Error()))
	} else {
		util.WriteJSON(w, util.GenRetObj(code, ret))
	}
}

func getRetInCache(day int, qType string) *StatRet {
	if cacheTimeout == nil {
		util.LogTrace.Printf("No cache")
		return nil
	} else if time.Now().After(*cacheTimeout) {
		util.LogTrace.Printf("Cache timeout")
		return nil
	}
	key := fmt.Sprintf("%d-%s", day, qType)
	return cache[key]
}

func setRetCache(day int, qType string, ret *StatRet) {
	if ret != nil {
		key := fmt.Sprintf("%d-%s", day, qType)
		cache[key] = ret

		now := time.Now().Local()
		dayEnd := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999, now.Location())
		cacheTimeout = &dayEnd
		util.LogTrace.Printf("Update cache of %s: %s", key, dayEnd.Format(time.RFC3339))
	}
}

func getQuestionParam(r *http.Request) (int, string, error) {
	dayStr := r.FormValue("days")
	questionType := r.FormValue("type")
	day, err := strconv.Atoi(dayStr)
	if err != nil {
		return 1, "", err
	}

	if day < 0 {
		day = 1
	}
	if day > 30 {
		day = 30
	}

	return day, questionType, nil
}

func handleDialogOneDayStatistic(w http.ResponseWriter, r *http.Request) {
	appid := util.GetAppID(r)
	start, end, err := getInputTime(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	tagType := r.FormValue("type")
	ret, errCode, err := GetDialogOneDayStatistic(appid, start.Unix(), end.Unix(), tagType)
	if err != nil {
		util.WriteJSON(w, util.GenRetObj(errCode, err.Error()))
	} else {
		util.WriteJSON(w, ret)
	}
}

func handleRobotsTraffic(w http.ResponseWriter, r *http.Request) {
	appid := util.GetAppID(r)
	if appid == "" {
		http.Error(w, "appid is empty", http.StatusBadRequest)
		return
	}

	db := util.GetDB(ModuleInfo.ModuleName)
	if db == nil {
		http.Error(w, "get db failed", http.StatusInternalServerError)
		return
	}
	start, end, err := getInputTime(r)
	if err != nil {
		http.Error(w, "Bad Request: "+err.Error(), http.StatusBadRequest)
		return
	}
	typ, err := getType(r)
	if err != nil {
		http.Error(w, "Bad Request: "+err.Error(), http.StatusBadRequest)
		return
	}
	rows, err := RobotTrafficsTable.GetGroupedRows(appid, typ, "name", []string{"resolved_rate"}, start, end)
	if err != nil {
		http.Error(w, "Get rows failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	var output = StatResponse{
		Headers: RobotTrafficsTable.Columns,
		Data:    rows,
	}
	err = util.WriteJSON(w, output)
	if err != nil {
		http.Error(w, "io failed, "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func handleRobotsResponse(w http.ResponseWriter, r *http.Request) {
	appid := util.GetAppID(r)
	if appid == "" {
		http.Error(w, "appid is empty", http.StatusBadRequest)
		return
	}
	db := util.GetDB(ModuleInfo.ModuleName)
	if db == nil {
		http.Error(w, "get db failed", http.StatusInternalServerError)
		return
	}
	start, end, err := getInputTime(r)
	if err != nil {
		http.Error(w, "Bad Request: "+err.Error(), http.StatusBadRequest)
		return
	}
	typ, err := getType(r)
	if err != nil {
		http.Error(w, "Bad Request: "+err.Error(), http.StatusBadRequest)
		return
	}
	rows, err := RobotResponseTable.GetGroupedRows(appid, typ, "name", nil, start, end)
	if err != nil {
		http.Error(w, "Get rows failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	var output = StatResponse{
		Headers: RobotResponseTable.Columns,
		Data:    rows,
	}
	err = util.WriteJSON(w, output)
	if err != nil {
		http.Error(w, "io failed, "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func handleMonitor(w http.ResponseWriter, r *http.Request) {
	appid := util.GetAppID(r)
	if appid == "" {
		http.Error(w, "appid is empty", http.StatusBadRequest)
		return
	}
	db := util.GetDB(ModuleInfo.ModuleName)
	if db == nil {
		http.Error(w, "get db failed", http.StatusInternalServerError)
		return
	}
	start, end, err := getInputTime(r)
	if err != nil {
		http.Error(w, "Bad Request: "+err.Error(), http.StatusBadRequest)
		return
	}

	brandName, ok := mux.Vars(r)["name"]
	if !ok || brandName == "" {
		http.Error(w, "Bad Request: name should be provided in query string", http.StatusBadRequest)
		return
	}
	tags, err := getTagValue(appid, 2)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	brandName, ok = tags[brandName]
	if !ok {
		http.Error(w, fmt.Sprintf("%v", tags), http.StatusBadRequest)
		return
	}

	var rows []statsRow
	var st StatTable
	var selector func(appID string, start, end time.Time, eq ...whereEqual) ([]statsRow, error)
	if end.Sub(start).Hours() <= 24.0 {
		st = HourlyMonitortable
		selector = NewStatsSelector(st, "cache_hour")
	} else {
		st = DailyMonitorTable
		selector = NewStatsSelector(st, "cache_day")
	}

	rows, err = selector(appid, start, end, whereEqual{"name", brandName}, whereEqual{"type", 2})
	if err != nil {
		http.Error(w, "query error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var output = StatResponse{
		Headers: st.Columns,
		Data:    rows,
	}

	err = util.WriteJSON(w, output)
	if err != nil {
		http.Error(w, "io failed, "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func handleLastVisit(w http.ResponseWriter, r *http.Request) {
	appID := util.GetAppID(r)
	start, end, err := getInputTime(r)
	if err != nil {
		http.Error(w, "Bad Request: "+err.Error(), http.StatusBadRequest)
		return
	}
	qs := r.URL.Query()
	eq := []whereEqual{
		//type is fixed to 2 now.
		whereEqual{"type", 2},
	}
	brand := qs.Get("brand")
	if brand == "" {
		http.Error(w, "No brand name is given.", http.StatusBadRequest)
		return
	}
	tags, err := getTagValue(appID, 2)
	if err != nil {
		http.Error(w, "tag not found "+err.Error(), http.StatusInternalServerError)
		return
	}
	var ok bool
	brand, ok = tags[brand]
	if !ok {
		http.Error(w, "brand is invalid", http.StatusBadRequest)
		return
	}
	eq = append(eq, whereEqual{"name", brand})
	uid := qs.Get("uid")
	if uid != "" {
		eq = append(eq, whereEqual{"user_id", uid})
	}
	//TODO: use phone_number in somewhere search condiction
	// phone = qs.Get("phone_number")

	//It is a quick dirty fix for UserContact
	//因為最後訪問時間的渠道沒有全部的概念, 但現在在api端寫找出max方法來不及了, 當渠道為all先把name拔掉。
	var st StatTable
	if brand == "all" {
		st = StatTable{
			Name:    UserContactsTable.Name,
			Columns: UserContactsTable.Columns[1:],
		}
	} else {
		st = UserContactsTable
	}

	selector := NewStatsSelector(st, "last_chat")
	rows, err := selector(appID, start, end, eq...)
	if err != nil {
		http.Error(w, "query error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	var output = StatResponse{
		Headers: st.Columns,
		Data:    rows,
	}

	err = util.WriteJSON(w, output)
	if err != nil {
		http.Error(w, "io failed, "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func handleRecords(w http.ResponseWriter, r *http.Request) {
	appID := util.GetAppID(r)
	start, end, err := getInputTime(r)
	if err != nil {
		http.Error(w, "bad request: "+err.Error(), http.StatusBadRequest)
		return
	}
	userID, ok := mux.Vars(r)["uid"]
	if !ok {
		http.Error(w, "uid is not given on path", http.StatusBadRequest)
		return
	}
	output, err := GetChatRecords(appID, start, end, userID)
	if err != nil {
		http.Error(w, "query failed:"+err.Error(), http.StatusInternalServerError)
		return
	}
	err = util.WriteJSON(w, output)
	if err != nil {
		http.Error(w, "io failed, "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func handleFAQStats(w http.ResponseWriter, r *http.Request) {
	appID := util.GetAppID(r)
	start, end, err := getInputTime(r)
	if err != nil {
		http.Error(w, "bad request: "+err.Error(), http.StatusBadRequest)
		return
	}
	qs := r.URL.Query()
	brandName := qs.Get("brand")
	tags, err := getTagValue(appID, 2)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var ok bool
	brandName, ok = tags[brandName]
	if !ok {
		http.Error(w, "brand is invalid", http.StatusBadRequest)
		return
	}

	keyword := qs.Get("keyword")
	output, err := GetFAQStats(appID, start, end, brandName, keyword)
	if err != nil {
		http.Error(w, "query failed:"+err.Error(), http.StatusInternalServerError)
		return
	}
	err = util.WriteJSON(w, output)
	if err != nil {
		http.Error(w, "io failed, "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func handleSessionQuery(w http.ResponseWriter, r *http.Request) {
	appID := util.GetAppID(r)
	var condition SessionCondition
	err := util.ReadJSON(r, &condition)
	if err != nil {
		http.Error(w, "Body bad formatted, "+err.Error(), http.StatusBadRequest)
		return
	}
	// Check  filter Conditition
	if condition.Limit == nil {
		http.Error(w, "need limit condition, "+err.Error(), http.StatusBadRequest)
		return
	}
	totalSize, sessions, err := GetSessions(appID, condition)
	if err != nil {
		http.Error(w, "Get sessions failed, "+err.Error(), http.StatusInternalServerError)
		return
	}
	var response = struct {
		Sessions []Session `json:"sessions"`
		Size     int       `json:"total_size"`
	}{
		Sessions: sessions,
		Size:     totalSize,
	}
	err = util.WriteJSON(w, response)
	if err != nil {
		util.LogError.Printf("IO Error %v\n", err)
	}
}

