package Stats

import (
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"time"

	"emotibot.com/emotigo/module/vipshop-admin/util"
	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
)

var (
	// ModuleInfo is needed for module define
	ModuleInfo util.ModuleInfo
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

func handleListAudit(ctx context.Context) {
	appid := util.GetAppID(ctx)
	input, err := loadFilter(ctx)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		return
	}

	ret, errCode, err := GetAuditList(appid, input)
	if err != nil {
		ctx.JSON(util.GenRetObj(errCode, err.Error()))
	} else {
		ctx.JSON(util.GenRetObj(errCode, ret))
	}
}

func loadFilter(ctx context.Context) (*AuditInput, error) {
	input := &AuditInput{}
	err := ctx.ReadJSON(input)
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

func handleQuestionStatistic(ctx context.Context) {
	// user query, standard query, score, count
	day, qType, err := getQuestionParam(ctx)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		return
	}

	appid := util.GetAppID(ctx)

	util.LogTrace.Printf("Request Questions Statistic: days=[%d] type=[%s]", day, qType)

	ret, code, err := GetQuestionStatisticResult(appid, day, qType)
	if err != nil {
		ctx.JSON(util.GenRetObj(code, err.Error()))
	} else {
		ctx.JSON(util.GenRetObj(code, ret))
	}
}

func getQuestionParam(ctx context.Context) (int, string, error) {
	dayStr := ctx.FormValue("days")
	questionType := ctx.FormValue("type")
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

	if util.Contains([]string{"unsolved"}, questionType) {
		return 1, questionType, errors.New("Not support type")
	}

	return day, questionType, nil
}
