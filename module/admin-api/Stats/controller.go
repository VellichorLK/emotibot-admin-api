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
