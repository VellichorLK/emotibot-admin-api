package QA

import (
	"emotibot.com/emotigo/module/vipshop-admin/ApiError"
	"emotibot.com/emotigo/module/vipshop-admin/util"
	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
)

var (
	// ModuleInfo is needed for module define
	TestModuleInfo util.ModuleInfo
)

const (
	DefaultListPerPage = 20
)

func init() {
	TestModuleInfo = util.ModuleInfo{
		ModuleName: "qatest",
		EntryPoints: []util.EntryPoint{
			util.NewEntryPoint("POST", "chat-test", []string{"view"}, hadleChatTest),
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

func hadleChatTest(ctx context.Context) {
	appid := util.GetAppID(ctx)
	user := util.GetUserID(ctx)
	input, err := loadQATestInput(ctx)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.JSON(util.GenRetObj(ApiError.JSON_PARSE_ERROR, err.Error()))
		return
	}

	ret, errCode, err := DoChatRequestWithDC(appid, user, input)
	if err != nil {
		ctx.JSON(util.GenRetObj(errCode, err.Error()))
	} else {
		ctx.JSON(util.GenRetObj(errCode, ret))
	}
}

func loadQATestInput(ctx context.Context) (*QATestInput, error) {
	input := &QATestInput{}
	err := ctx.ReadJSON(input)
	if err != nil {
		return nil, err
	}
	return input, nil
}

func getOpenAPIURL() string {
	return getEnvironment("OPENAPI_TEST_URL")
}

func getTestURL() string {
	return getEnvironment("TEST_URL")
}
