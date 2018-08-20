package QA

import (
	"net/http"

	"emotibot.com/emotigo/module/admin-api/ApiError"
	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/admin-api/util/requestheader"
)

var (
	// ModuleInfo is needed for module define
	TestModuleInfo util.ModuleInfo
	HandleFuncMap  map[string]func(appid string, user string, input *QATestInput) (*RetData, int, error)
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
	HandleFuncMap = map[string]func(appid string, user string, input *QATestInput) (*RetData, int, error){
		"DC":         DoChatRequestWithDC,
		"OPENAPI":    DoChatRequestWithOpenAPI,
		"CONTROLLER": DoChatRequestWithController,
		"BFOP":       DoChatRequestWithBFOPController,
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

func hadleChatTest(w http.ResponseWriter, r *http.Request) {
	appid := requestheader.GetAppID(r)
	user := requestheader.GetUserID(r)
	input, err := loadQATestInput(r)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.JSON_PARSE_ERROR, err.Error()), http.StatusBadRequest)
		return
	}
	var ret *RetData
	var errCode int

	if handleFunc, ok := HandleFuncMap[getQATestType()]; ok {
		ret, errCode, err = handleFunc(appid, user, input)
	} else {
		// default use dc
		ret, errCode, err = DoChatRequestWithDC(appid, user, input)
	}
	if err != nil {
		util.WriteJSON(w, util.GenRetObj(errCode, err.Error()))
	} else {
		util.WriteJSON(w, util.GenRetObj(errCode, ret))
	}
}

func loadQATestInput(r *http.Request) (*QATestInput, error) {
	input := &QATestInput{}
	err := util.ReadJSON(r, input)
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

func getControllerURL() string {
	return getEnvironment("CONTROLLER_URL")
}

func getQATestType() string {
	return getEnvironment("TEST_TYPE")
}
