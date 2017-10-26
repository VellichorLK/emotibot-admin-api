package Switch

import (
	"emotibot.com/emotigo/module/vipshop-admin/ApiError"
	"emotibot.com/emotigo/module/vipshop-admin/util"
	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
)

var (
	// ModuleInfo is needed for module define
	ModuleInfo util.ModuleInfo
)

func init() {
	ModuleInfo = util.ModuleInfo{
		ModuleName: "switch-manage",
		EntryPoints: []util.EntryPoint{
			util.NewEntryPoint("GET", "switches", []string{"view"}, handleList),
			util.NewEntryPoint("GET", "switch/{id:int}", []string{"view"}, handleSwitch),
			util.NewEntryPoint("POST", "switch/{id:int}", []string{"edit"}, handleUpdateSwitch),
			util.NewEntryPoint("DELETE", "switch/{id:int}", []string{"delete"}, handleDeleteSwitch),
			util.NewEntryPoint("PUT", "switch/create", []string{"create"}, handleNewSwitch),
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

// handleList will show onoff list in database
func handleList(ctx context.Context) {
	appid := util.GetAppID(ctx)

	list, errCode, err := GetSwitches(appid)
	errMsg := ApiError.GetErrorMsg(errCode)
	if errCode != ApiError.SUCCESS {
		ctx.JSON(util.GenRetObj(errCode, errMsg, err))
	} else {
		ctx.JSON(util.GenRetObj(errCode, errMsg, list))
	}
}

func handleSwitch(ctx context.Context) {
	id, _ := ctx.Params().GetInt("id")
	appid := util.GetAppID(ctx)

	ret, errCode, err := GetSwitch(appid, id)
	errMsg := ApiError.GetErrorMsg(errCode)
	if errCode != ApiError.SUCCESS {
		ctx.JSON(util.GenRetObj(errCode, errMsg, err))
	} else {
		ctx.JSON(util.GenRetObj(errCode, errMsg, ret))
	}
}

func handleNewSwitch(ctx context.Context) {
	appid := util.GetAppID(ctx)

	input := loadSwitchFromContext(ctx)
	if input == nil {
		ctx.StatusCode(iris.StatusBadRequest)
		return
	}

	errCode, err := InsertSwitch(appid, input)
	errMsg := ApiError.GetErrorMsg(errCode)
	if errCode != ApiError.SUCCESS {
		ctx.JSON(util.GenRetObj(errCode, errMsg, err))
	} else {
		ctx.JSON(util.GenRetObj(errCode, errMsg, input))
	}
}

func handleUpdateSwitch(ctx context.Context) {
	id, _ := ctx.Params().GetInt("id")
	appid := util.GetAppID(ctx)

	input := loadSwitchFromContext(ctx)
	if input == nil {
		ctx.StatusCode(iris.StatusBadRequest)
		return
	}

	errCode, err := UpdateSwitch(appid, id, input)
	errMsg := ApiError.GetErrorMsg(errCode)
	if errCode != ApiError.SUCCESS {
		ctx.JSON(util.GenRetObj(errCode, errMsg, err))
	} else {
		ctx.JSON(util.GenRetObj(errCode, errMsg, input))
	}
}

func handleDeleteSwitch(ctx context.Context) {
	id, _ := ctx.Params().GetInt("id")
	appid := util.GetAppID(ctx)

	errCode, err := DeleteSwitch(appid, id)
	errMsg := ApiError.GetErrorMsg(errCode)
	if errCode != ApiError.SUCCESS {
		ctx.JSON(util.GenRetObj(errCode, errMsg, err))
	} else {
		ctx.JSON(util.GenRetObj(errCode, errMsg, nil))
	}
}

func loadSwitchFromContext(ctx context.Context) *SwitchInfo {
	input := &SwitchInfo{}
	err := ctx.ReadJSON(input)
	if err != nil {
		return nil
	}

	return input
}
