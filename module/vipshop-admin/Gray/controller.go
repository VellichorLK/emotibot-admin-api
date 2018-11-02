package Gray

import (
	"net/http"
	"strconv"
	"time"
	"fmt"

	"emotibot.com/emotigo/module/vipshop-admin/util"
	"github.com/kataras/iris/context"
)

var (
	// ModuleInfo is needed for module define
	ModuleInfo     util.ModuleInfo

)

type errorJSON struct {
	Message string `json:"message"`
}

func init() {
	ModuleInfo = util.ModuleInfo{
		ModuleName: "faq",
		EntryPoints: []util.EntryPoint{
			util.NewEntryPoint("POST", "te/gray/update/{percent:int}", []string{"edit"}, handleUpdateGray),
			util.NewEntryPoint("POST", "te/gray/query", []string{"edit"}, handleQueryGray),
			util.NewEntryPoint("POST", "te/white/query", []string{"edit"}, handleQueryWhite),
			util.NewEntryPoint("POST", "te/white/create", []string{"edit"}, handleCreateWhite),
			util.NewEntryPoint("POST", "te/white/delete", []string{"edit"}, handleDeleteWhite),
		},
	}

}

func updateConsul() {
	unixTime := time.Now().UnixNano() / 1000000
	_, err := util.ConsulUpdateVal("te/gray", unixTime)
	if err != nil {
		util.LogError.Println("consul update failed, %v", err)
	}
}

func handleUpdateGray(ctx context.Context) {
	type successJSON struct {
		
	}

	db := util.GetAuditDB()
	if db == nil {
		util.LogError.Println("Audit DB Connection failed")
		ctx.StatusCode(500)
		ctx.JSON("Audit DB Connection failed")
	}
	percent, err := strconv.Atoi(ctx.Params().Get("percent"))

	util.LogInfo.Println("percent=", percent);

	rows, err := db.Query("select `percent` from gray")
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(errorJSON{Message: err.Error()})
		return
	}
	
	if !rows.Next() {
		ctx.StatusCode(http.StatusNotFound)
		return
	}
	var oldPercent = 0;
	rows.Scan(&oldPercent);

	defer rows.Close()

	ret, err := db.Exec("update gray set percent = ?", percent)
	var returnJSON = successJSON{}
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(errorJSON{Message: err.Error()})
		addAuditGray(ctx, util.AuditOperationEdit, fmt.Sprintf("[灰度发布]：%d => %d", oldPercent, percent), 0)
		return
	} else {
		addAuditGray(ctx, util.AuditOperationEdit, fmt.Sprintf("[灰度发布]：%d => %d", oldPercent, percent), 1)
	}
	util.LogInfo.Println("ret = ", ret);
	util.LogInfo.Println("returnJSON = ", returnJSON);
	updateConsul()
	ctx.JSON(returnJSON)
}

func handleQueryGray(ctx context.Context) {
	type successJSON struct {
		Percent     float64        `json:"percent"`
	}

	db := util.GetAuditDB()
	if db == nil {
		util.LogError.Println("Audit DB Connection failed")
		ctx.StatusCode(500)
		ctx.JSON("Audit DB Connection failed")
	}
	
	rows, err := db.Query("select `percent` from gray")
	var returnJSON = successJSON{}
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(errorJSON{Message: err.Error()})
		return
	}
	
	if !rows.Next() {
		ctx.StatusCode(http.StatusNotFound)
		return
	}
	rows.Scan(&returnJSON.Percent)

	defer rows.Close()

	if err = rows.Err(); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
	}
	util.LogInfo.Println("returnJSON = ", returnJSON);

	ctx.JSON(returnJSON)
}

func handleQueryWhite(ctx context.Context) {
	appid := util.GetAppID(ctx)
	// parse QueryCondition
	condition, err := ParseCondition(ctx)
	if err != nil {
		util.LogError.Printf("Error happened while parsing query options %s", err.Error())
		ctx.StatusCode(http.StatusBadRequest)
		return
	}

	whites, err := FetchWhites(condition, appid)
	if err != nil {
		util.LogError.Printf("Error FetchWhites %s", err.Error())
		ctx.StatusCode(http.StatusBadRequest)
		return
	}
	util.LogInfo.Println("whites = ", whites);


	type successJSON struct {
		QueryResult []White    `json:"QueryResult"`
	}

	//totalPage := math.Floor(float64(total / condition.Limit))

	response := successJSON{
		QueryResult:  whites,
	}

	ctx.JSON(response)
}


func handleCreateWhite(ctx context.Context) {
	appid := util.GetAppID(ctx)
	// parse QueryCondition
	userId := ctx.FormValue("userIds")
	util.LogInfo.Println("create userId", userId);
	
	total, err := BatchInsertWhite(userId, appid);
	util.LogInfo.Println("total = ", total)
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		addAuditWhite(ctx, util.AuditOperationAdd, fmt.Sprintf("[白名单]：%s", userId), 0);
		return
	} else {
		addAuditWhite(ctx, util.AuditOperationAdd, fmt.Sprintf("[白名单]：%s", userId), 1);
	}
	updateConsul()
	//ctx.JSON("{}")
}

func handleDeleteWhite(ctx context.Context) {
	appid := util.GetAppID(ctx)
	// parse QueryCondition
	userId := ctx.FormValue("userIds")
	util.LogInfo.Println("del userId", userId);
	
	total, err := BatchDeleteWhite(userId, appid);
	util.LogInfo.Println("total = ", total)
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		addAuditWhite(ctx, util.AuditOperationDelete, fmt.Sprintf("[白名单]：%s", userId), 0);
		return
	} else {
		addAuditWhite(ctx, util.AuditOperationDelete, fmt.Sprintf("[白名单]：%s", userId), 1);
	}
	updateConsul()
	
}

func addAuditWhite(ctx context.Context, operation string, msg string, result int) {
	userID := util.GetUserID(ctx)
	userIP := util.GetUserIP(ctx)
	util.AddAuditLog(userID, userIP, util.AuditModuleTE, operation, msg, result)
}

func addAuditGray(ctx context.Context, operation string, msg string, result int) {
	userID := util.GetUserID(ctx)
	userIP := util.GetUserIP(ctx)
	util.AddAuditLog(userID, userIP, util.AuditModuleTE, operation, msg, result)
}