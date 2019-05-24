package Service

import (
	"fmt"
	"net/http"
	"time"

	"emotibot.com/emotigo/module/admin-api/util/AdminErrors"

	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/admin-api/util/requestheader"
	"emotibot.com/emotigo/pkg/api/dal/v1"
	dac "emotibot.com/emotigo/pkg/api/dac/v1"
)

var (
	// ModuleInfo is needed for module define
	ModuleInfo util.ModuleInfo
	dalClient  *dal.Client
	dacClient  *dac.Client
)

func init() {
	ModuleInfo = util.ModuleInfo{
		ModuleName: "service",
		EntryPoints: []util.EntryPoint{
			util.NewEntryPoint("GET", "recommend", []string{}, handleGetRecommandQuestion),
			util.NewEntryPointWithVer("GET", "recommend",[]string{}, handleGetRecommandQuestionV2, 2),
		},
	}
}

func Init() error {
	var err error
	dalURL, found := util.GetEnvOf("server")["DAL_URL"]
	if !found {
		return fmt.Errorf("CAN NOT FOUND SERVER ENV \"DAL_URL\"")
	}

	dalClient, err = dal.NewClientWithHTTPClient(dalURL, &http.Client{Timeout: time.Duration(5) * time.Second})
	if err != nil {
		return fmt.Errorf("new dal client failed, %v", err)
	}

	dacURL, found := util.GetEnvOf("server")["DAC_URL"]
	if !found {
		return fmt.Errorf("CAN NOT FOUND SERVER ENV \"DAC_URL\"")
	}

	dacClient, err = dac.NewClientWithHTTPClient(dacURL, &http.Client{Timeout: time.Duration(5) * time.Second})
	if err != nil {
		return fmt.Errorf("new dac client failed, %v", err)
	}

	return nil
}

func handleGetRecommandQuestion(w http.ResponseWriter, r *http.Request) {
	appid := requestheader.GetAppID(r)
	if appid == "" {
		appid = r.URL.Query().Get("appid")
	}

	userInput := r.URL.Query().Get("user_input")
	limit, _ := util.GetParamInt(r, "limit")

	if userInput == "" {
		util.ReturnError(w, AdminErrors.ErrnoRequestError, "user_input")
		return
	}
	if limit == 0 {
		util.ReturnError(w, AdminErrors.ErrnoRequestError, "limit")
		return
	}

	ret, err := GetRecommandStdQuestion(appid, userInput, limit)
	if err != nil {
		util.ReturnError(w, err.Errno(), err.Error())
		return
	}
	util.Return(w, nil, ret)
}
func handleGetRecommandQuestionV2(w http.ResponseWriter, r *http.Request) {
	appid := requestheader.GetAppID(r)
	if appid == "" {
		appid = r.URL.Query().Get("appid")
	}

	userInput := r.URL.Query().Get("user_input")
	limit, _ := util.GetParamInt(r, "limit")

	if userInput == "" {
		util.ReturnError(w, AdminErrors.ErrnoRequestError, "user_input")
		return
	}
	if limit == 0 {
		util.ReturnError(w, AdminErrors.ErrnoRequestError, "limit")
		return
	}

	ret, err := GetRecommandStdQuestionFromDac(appid, userInput, limit)
	if err != nil {
		util.ReturnError(w, err.Errno(), err.Error())
		return
	}
	util.Return(w, nil, ret)
}