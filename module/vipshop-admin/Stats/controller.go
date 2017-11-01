package Stats

import (
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
		},
	}
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
