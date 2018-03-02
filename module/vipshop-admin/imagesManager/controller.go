package imagesManager

import (
	"encoding/base64"
	"net/http"
	"path/filepath"

	"emotibot.com/emotigo/module/vipshop-admin/ApiError"
	"emotibot.com/emotigo/module/vipshop-admin/util"
	"github.com/kataras/iris/context"
)

var (
	ModuleInfo util.ModuleInfo
	Envs       map[string]string
	LocalID    uint64
	Volume     string
)

func init() {
	ModuleInfo = util.ModuleInfo{
		ModuleName: "mediabase",
		EntryPoints: []util.EntryPoint{
			util.NewEntryPoint("POST", "images", []string{}, receiveImage),
			util.NewEntryPoint("GET", "images", []string{}, handleImageList),
		},
	}
}

//InitDB init the database connection
func InitDB() {

	Envs = util.GetModuleEnvironments(ModuleInfo.ModuleName)

	url := util.GetEnviroment(Envs, "MYSQL_URL")
	user := util.GetEnviroment(Envs, "MYSQL_USER")
	pass := util.GetEnviroment(Envs, "MYSQL_PASS")
	db := util.GetEnviroment(Envs, "MYSQL_DB")
	dao, err := util.InitDB(url, user, pass, db)
	if err != nil {
		util.LogError.Printf("Cannot init self learning db, [%s:%s@%s:%s]: %s\n", user, pass, url, db, err.Error())
		return
	}
	util.SetDB(ModuleInfo.ModuleName, dao)

	mediaLocal := filepath.Clean(util.GetEnviroment(Envs, "LOCATION"))
	Volume = filepath.Clean(util.GetEnviroment(Envs, "VOLUME"))
	LocalID, _ = getLocationID(mediaLocal)
}

func receiveImage(ctx context.Context) {
	args := make([]*uploadArg, 0)
	err := ctx.ReadJSON(&args)
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(util.GenRetObj(ApiError.JSON_PARSE_ERROR, err.Error()))
		return
	}

	for _, file := range args {

		content, err := base64.StdEncoding.DecodeString(file.Content)
		if err != nil {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()))
			return
		}

		err = storeImage(file.FileName, content)
		if err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(util.GenRetObj(ApiError.OPENAPI_URL_ERROR, err.Error()))
			return
		}
	}
}

func handleImageList(ctx context.Context) {
	params := ctx.URLParams()

	listArgs, err := getImagesParams(params)
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()))
		return
	}
	ctx.JSON(listArgs)

}
