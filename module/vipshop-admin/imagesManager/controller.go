package imagesManager

import (
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"emotibot.com/emotigo/module/vipshop-admin/ApiError"
	"emotibot.com/emotigo/module/vipshop-admin/util"
	"github.com/kataras/iris/context"
)

var (
	ModuleInfo util.ModuleInfo
	Envs       map[string]string
	LocalID    uint64
	Volume     string
	db         *sql.DB
)

func init() {
	ModuleInfo = util.ModuleInfo{
		ModuleName: "mediabase",
		EntryPoints: []util.EntryPoint{
			util.NewEntryPoint("POST", "images", []string{}, receiveImage),
			util.NewEntryPoint("GET", "images", []string{}, handleImageList),
			util.NewEntryPoint("DELETE", "images/{id:int}", []string{}, handleDeleteImage),
			util.NewEntryPoint("POST", "images/{id:int}", []string{}, updateImage),
			util.NewEntryPoint("PUT", "images/{id:int}", []string{}, copyImage),
			util.NewEntryPoint("POST", "images/delete", []string{}, handleDeleteImages),
			util.NewEntryPoint("POST", "images/download", []string{}, downloadImages),
			util.NewEntryPoint("POST", "images/search", []string{}, searchImageHandler),
		},
	}
}

//InitDB init the database connection
func InitDB() {

	Envs = util.GetModuleEnvironments(ModuleInfo.ModuleName)

	url := util.GetEnviroment(Envs, "MYSQL_URL")
	user := util.GetEnviroment(Envs, "MYSQL_USER")
	pass := util.GetEnviroment(Envs, "MYSQL_PASS")
	dbName := util.GetEnviroment(Envs, "MYSQL_DB")
	dao, err := util.InitDB(url, user, pass, dbName)
	if err != nil {
		util.LogError.Printf("Cannot init self learning db, [%s:%s@%s:%s]: %s\n", user, pass, url, dbName, err.Error())
		return
	}
	util.SetDB(ModuleInfo.ModuleName, dao)
	db = dao

	mediaLocal := strings.Trim(util.GetEnviroment(Envs, "LOCATION"), "/")
	Volume = filepath.Clean(util.GetEnviroment(Envs, "VOLUME"))
	LocalID, err = getLocationID(mediaLocal)
	if err != nil {
		util.LogError.Println(err)
	}
}

func receiveImage(ctx context.Context) {
	args := make([]*uploadArg, 0)
	err := ctx.ReadJSON(&args)
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(util.GenRetObj(ApiError.JSON_PARSE_ERROR, err.Error()))
		return
	}
	tx, err := db.Begin()
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(util.GenRetObj(ApiError.OPENAPI_URL_ERROR, err.Error()))
		util.LogError.Println(err)
		return
	}
	defer tx.Rollback()

	auditLog := "[图片素材]:"
	auditRet := 1
	first := true

	userID := util.GetUserID(ctx)
	userIP := util.GetUserIP(ctx)
	defer func() {
		util.AddAuditLog(userID, userIP, util.AuditModuleMedia, util.AuditOperationAdd, auditLog, auditRet)
	}()

	for _, file := range args {
		if first {
			auditLog = fmt.Sprintf("%s%s", auditLog, file.FileName)
			first = false
		} else {
			auditLog = fmt.Sprintf("%s:%s", auditLog, file.FileName)
		}
		content, err := base64.StdEncoding.DecodeString(file.Content)
		if err != nil {
			ctx.StatusCode(http.StatusBadRequest)
			ctx.JSON(util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()))
			auditRet = 0
			return
		}

		_, err = storeImage(tx, file.FileName, content)
		if err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(util.GenRetObj(ApiError.OPENAPI_URL_ERROR, err.Error()))
			util.LogError.Println(err)
			auditRet = 0
			return
		}
	}

	err = tx.Commit()
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(util.GenRetObj(ApiError.OPENAPI_URL_ERROR, err.Error()))
		util.LogError.Println(err)
		auditRet = 0
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
	lastOperation := time.Now()

	list, err := getImageList(listArgs)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(util.GenRetObj(ApiError.OPENAPI_URL_ERROR, err.Error()))
		util.LogError.Println(err)
		return

	}
	util.LogInfo.Printf("get image list in handleImageList took: %s\n", time.Since(lastOperation))
	ctx.StatusCode(http.StatusOK)
	ctx.JSON(list)
	return

}

func handleDeleteImage(ctx context.Context) {
	imageID, err := strconv.ParseUint(ctx.Params().GetEscape("id"), 10, 64)
	if err != nil || imageID <= 0 {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(util.GenRetObj(ApiError.REQUEST_ERROR, "Invalid id "+ctx.Params().GetEscape("id")))
		return
	}
	lastOperation := time.Now()

	imageIDs := []interface{}{imageID}

	auditRet := 1
	files, _ := getFileNameByImageID(imageIDs)
	util.LogInfo.Printf("get file name in handleDeleteImage took: %s\n", time.Since(lastOperation))
	lastOperation = time.Now()

	err = deleteImagesByID(ctx, imageIDs)
	if err != nil {
		auditRet = 0
	}
	util.LogInfo.Printf("delete image in handleDeleteImage took: %s\n", time.Since(lastOperation))

	writeImageAduitLog(ctx, files, auditRet, util.AuditOperationDelete)
}

func handleDeleteImages(ctx context.Context) {
	ids := make([]uint64, 0)
	err := ctx.ReadJSON(&ids)
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(util.GenRetObj(ApiError.JSON_PARSE_ERROR, err.Error()))
		return
	}
	lastOperation := time.Now()

	imageIDs := make([]interface{}, len(ids))
	for i := 0; i < len(ids); i++ {
		imageIDs[i] = ids[i]
	}

	auditRet := 1
	files, _ := getFileNameByImageID(imageIDs)
	util.LogInfo.Printf("get file name in handleDeleteImages took: %s\n", time.Since(lastOperation))
	lastOperation = time.Now()

	err = deleteImagesByID(ctx, imageIDs)
	if err != nil {
		auditRet = 0
	}
	util.LogInfo.Printf("delete images in handleDeleteImages took: %s\n", time.Since(lastOperation))

	writeImageAduitLog(ctx, files, auditRet, util.AuditOperationDelete)

}

func writeImageAduitLog(ctx context.Context, files map[uint64]string, result int, operation string) (err error) {
	userID := util.GetUserID(ctx)
	userIP := util.GetUserIP(ctx)

	auditLog := "[图片素材]:"
	first := true

	for _, fileName := range files {
		if first {
			auditLog = fmt.Sprintf("%s%s", auditLog, fileName)
			first = false
		} else {
			auditLog = fmt.Sprintf("%s;%s", auditLog, fileName)
		}
	}

	err = util.AddAuditLog(userID, userIP, util.AuditModuleMedia, operation, auditLog, result)
	return
}

func deleteImagesByID(ctx context.Context, imageIDs []interface{}) (err error) {
	lastOperation := time.Now()
	relationMap, err := getRelationByID(imageIDs)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		util.LogError.Println(err)
		return
	}
	util.LogInfo.Printf("get image relation in deleteImagesByID took: %s\n", time.Since(lastOperation))
	lastOperation = time.Now()

	if len(relationMap) > 0 {
		ctx.StatusCode(http.StatusConflict)
		err = fmt.Errorf("Conflict")
		return
	}

	_, err = deleteImages(imageIDs)
	util.LogInfo.Printf("delete images in deleteImagesByID took: %s\n", time.Since(lastOperation))


	if err != nil {
		if err == errImageNotExist {
			ctx.StatusCode(http.StatusBadRequest)
		} else {
			ctx.StatusCode(http.StatusInternalServerError)
			util.LogError.Println(err)
		}
		ctx.JSON(util.GenRetObj(ApiError.OPENAPI_URL_ERROR, err.Error()))
		return
	}
	return
}

func updateImage(ctx context.Context) {
	imageID, err := strconv.ParseUint(ctx.Params().GetEscape("id"), 10, 64)
	if err != nil || imageID <= 0 {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(util.GenRetObj(ApiError.REQUEST_ERROR, "Invalid id "+ctx.Params().GetEscape("id")))
		return
	}
	lastOperation := time.Now()

	files, err := getFileNameByImageID([]interface{}{imageID})
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(util.GenRetObj(ApiError.OPENAPI_URL_ERROR, err.Error()))
		util.LogError.Println(err)
		return
	}
	util.LogInfo.Printf("get file name in updateImage took: %s\n", time.Since(lastOperation))
	lastOperation = time.Now()


	if _, ok := files[imageID]; !ok {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(util.GenRetObj(ApiError.REQUEST_ERROR, "No such id"))
		return
	}

	title := ctx.FormValue(TITLE)
	var tx *sql.Tx

	auditRet := 1
	auditLog := fmt.Sprintf("[图片素材]:%s => %s", files[imageID], title)
	userID := util.GetUserID(ctx)
	userIP := util.GetUserIP(ctx)

	if title != "" && title != files[imageID] {

		orgExt := path.Ext(files[imageID])
		newExt := path.Ext(title)
		if orgExt != newExt {
			ctx.StatusCode(http.StatusConflict)
			ctx.JSON(util.GenRetObj(ApiError.OPENAPI_URL_ERROR, errors.New("extension can't not be changed")))
			auditRet = 0
			util.AddAuditLog(userID, userIP, util.AuditModuleMedia, util.AuditOperationEdit, auditLog, auditRet)
			return
		}

		tx, err = db.Begin()
		if err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(util.GenRetObj(ApiError.OPENAPI_URL_ERROR, err.Error()))
			util.LogError.Println(err)
			auditRet = 0
			util.AddAuditLog(userID, userIP, util.AuditModuleMedia, util.AuditOperationEdit, auditLog, auditRet)
			return
		}
		defer tx.Rollback()
		sqlString := "update " + imageTable + " set " + attrFileName + "=? where " + attrID + "=?"
		_, _, _, err = inputUniqueFileName(tx, sqlString, title, []interface{}{title, imageID})
		if err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(util.GenRetObj(ApiError.OPENAPI_URL_ERROR, err.Error()))
			util.LogError.Println(err)
			auditRet = 0
			util.AddAuditLog(userID, userIP, util.AuditModuleMedia, util.AuditOperationEdit, auditLog, auditRet)
			return
		}
	}
	util.LogInfo.Printf("update images in updateImage took: %s\n", time.Since(lastOperation))
	lastOperation = time.Now()

	file, fileHeader, err := ctx.FormFile(IMAGE)
	if fileHeader != nil {

		rawFileName, err := getImageName(imageID)
		if err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(util.GenRetObj(ApiError.OPENAPI_URL_ERROR, err.Error()))
			util.LogError.Println(err)
			auditRet = 0
			util.AddAuditLog(userID, userIP, util.AuditModuleMedia, util.AuditOperationEdit, auditLog, auditRet)
			return
		}

		dstFile, err := os.Create(Volume + "/" + rawFileName)
		if err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(util.GenRetObj(ApiError.OPENAPI_URL_ERROR, err.Error()))
			util.LogError.Println(err)
			auditRet = 0
			util.AddAuditLog(userID, userIP, util.AuditModuleMedia, util.AuditOperationEdit, auditLog, auditRet)
			return
		}

		_, err = io.Copy(dstFile, file)
		if err != nil {
			ctx.StatusCode(http.StatusInternalServerError)
			ctx.JSON(util.GenRetObj(ApiError.OPENAPI_URL_ERROR, err.Error()))
			util.LogError.Println(err)
			auditRet = 0
			util.AddAuditLog(userID, userIP, util.AuditModuleMedia, util.AuditOperationEdit, auditLog, auditRet)
			return
		}
	}
	util.LogInfo.Printf("replace old images in updateImage took: %s\n", time.Since(lastOperation))

	if tx != nil {
		tx.Commit()
	}

	util.AddAuditLog(userID, userIP, util.AuditModuleMedia, util.AuditOperationEdit, auditLog, auditRet)

}

func copyImage(ctx context.Context) {
	imageID, err := strconv.ParseUint(ctx.Params().GetEscape("id"), 10, 64)
	if err != nil || imageID <= 0 {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(util.GenRetObj(ApiError.REQUEST_ERROR, "Invalid id "+ctx.Params().GetEscape("id")))
		return
	}
	lastOperation := time.Now()

	title := ctx.FormValue(TITLE)
	if title == "" {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(util.GenRetObj(ApiError.REQUEST_ERROR, "No need parameter"))
		return
	}

	auditRet := 1
	var auditLog string = "[图片素材]"
	userID := util.GetUserID(ctx)
	userIP := util.GetUserIP(ctx)

	nameList, err := getRealFileNameByImageID([]interface{}{imageID})
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(util.GenRetObj(ApiError.OPENAPI_URL_ERROR, err.Error()))
		util.LogError.Println(err)
		auditRet = 0
		auditLog = fmt.Sprintf("%s:%s", auditLog, title)
		util.AddAuditLog(userID, userIP, util.AuditModuleMedia, util.AuditOperationAdd, auditLog, auditRet)
		return
	}
	util.LogInfo.Printf("get file name in copyImage took: %s\n", time.Since(lastOperation))
	lastOperation = time.Now()

	if len(nameList) <= 0 {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(util.GenRetObj(ApiError.REQUEST_ERROR, "No such id "))
		return
	}

	fileName := nameList[0]

	content, err := ioutil.ReadFile(Volume + "/" + fileName)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(util.GenRetObj(ApiError.OPENAPI_URL_ERROR, err.Error()))
		util.LogError.Println(err)
		auditRet = 0
		auditLog = fmt.Sprintf("%s:%s", auditLog, title)
		util.AddAuditLog(userID, userIP, util.AuditModuleMedia, util.AuditOperationAdd, auditLog, auditRet)
		return
	}
	util.LogInfo.Printf("copy image in copyImage took: %s\n", time.Since(lastOperation))
	lastOperation = time.Now()

	tx, err := db.Begin()
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(util.GenRetObj(ApiError.OPENAPI_URL_ERROR, err.Error()))
		util.LogError.Println(err)
		auditRet = 0
		auditLog = fmt.Sprintf("%s:%s", auditLog, title)
		util.AddAuditLog(userID, userIP, util.AuditModuleMedia, util.AuditOperationAdd, auditLog, auditRet)
		return
	}
	defer tx.Rollback()

	fileName, err = storeImage(tx, title, content)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(util.GenRetObj(ApiError.OPENAPI_URL_ERROR, err.Error()))
		util.LogError.Println(err)
		auditRet = 0
		auditLog = fmt.Sprintf("%s:%s", auditLog, fileName)
		util.AddAuditLog(userID, userIP, util.AuditModuleMedia, util.AuditOperationAdd, auditLog, auditRet)
		return
	}
	util.LogInfo.Printf("store image in copyImage took: %s\n", time.Since(lastOperation))

	tx.Commit()

	auditLog = fmt.Sprintf("%s:%s", auditLog, fileName)
	util.AddAuditLog(userID, userIP, util.AuditModuleMedia, util.AuditOperationAdd, auditLog, auditRet)
}

func downloadImages(ctx context.Context) {
	ids := make([]uint64, 0)
	err := ctx.ReadJSON(&ids)
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(util.GenRetObj(ApiError.JSON_PARSE_ERROR, err.Error()))
		return
	}
	if len(ids) == 0 {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(util.GenRetObj(ApiError.JSON_PARSE_ERROR, "No assigned id"))
		return
	}
	lastOperation := time.Now()

	imageIDs := make([]interface{}, len(ids))
	for i := 0; i < len(ids); i++ {
		imageIDs[i] = ids[i]
	}

	auditRet := 1
	files, err := getFileNameByImageID(imageIDs)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(util.GenRetObj(ApiError.OPENAPI_URL_ERROR, err.Error()))
		util.LogError.Println(err)
		return
	}
	util.LogInfo.Printf("get file name in downloadImages took: %s\n", time.Since(lastOperation))
	lastOperation = time.Now()

	b, err := packageImages(imageIDs)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(util.GenRetObj(ApiError.OPENAPI_URL_ERROR, err.Error()))
		util.LogError.Println(err)
		auditRet = 0
	}
	util.LogInfo.Printf("package images in downloadImages took: %s\n", time.Since(lastOperation))

	writeImageAduitLog(ctx, files, auditRet, util.AuditOperationExport)
	if auditRet == 0 {
		return
	}

	ctx.Header("Content-Disposition", "attachment; filename=download.zip")
	ctx.Header("Cache-Control", "public")
	ctx.Binary(b.Bytes())

}

func searchImageHandler(ctx context.Context) {

	var inputArgs = struct {
		AnswerIDs []uint64 `json:"answers"`
	}{}

	err := ctx.ReadJSON(&inputArgs)
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(util.GenRetObj(ApiError.JSON_PARSE_ERROR, err.Error()))
		return
	}

	if len(inputArgs.AnswerIDs) == 0 {
		ctx.StatusCode(http.StatusBadRequest)
		ctx.JSON(util.GenRetObj(ApiError.JSON_PARSE_ERROR, "No assigned id"))
		return
	}
	lastOperation := time.Now()

	answerImages, err := getAnswerImage(inputArgs.AnswerIDs)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(util.GenRetObj(ApiError.OPENAPI_URL_ERROR, "Internal server error"))
		util.LogError.Println(err)
		return
	}
	util.LogInfo.Printf("get answer images in searchImageHandler took: %s\n", time.Since(lastOperation))

	ctx.StatusCode(http.StatusOK)
	ctx.JSON(answerImages)
	return

}
