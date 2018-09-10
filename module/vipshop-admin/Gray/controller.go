package Gray

import (
	"fmt"
	"net/http"
	"strconv"
	//"math"

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
	fmt.Println("percent=", percent);
	ret, err := db.Exec("update gray set percent = ?", percent)
	var returnJSON = successJSON{}
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.JSON(errorJSON{Message: err.Error()})
		return
	}
	
	//rows.Scan()
	//defer rows.Close()

	/**if err = rows.Err(); err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
	}*/
	fmt.Println("ret = ", ret);
	fmt.Println("returnJSON = ", returnJSON);

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
	fmt.Println("returnJSON = ", returnJSON);

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
	/**total, err := QueryTotalWhite(condition, appid)
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		return
	}
	fmt.Println("total = ", total);*/

	whites, err := FetchWhites(condition, appid)
	if err != nil {
		fmt.Println(err);
		ctx.StatusCode(http.StatusBadRequest)
		return
	}
	fmt.Println("whites = ", whites);


	type successJSON struct {
		CurPage     string     `json:"CurPage"`
		QueryResult []White    `json:"QueryResult"`
		TotalPage   float64        `json:"TotalPage"`
		TotalNum    int        `json:"TotalNum"`
	}

	//totalPage := math.Floor(float64(total / condition.Limit))

	response := successJSON{
		CurPage:      "0",
		QueryResult:  whites,
		TotalPage:    0,
		TotalNum: 	  0,
	}

	ctx.JSON(response)
}


func handleCreateWhite(ctx context.Context) {
	appid := util.GetAppID(ctx)
	// parse QueryCondition
	userId := ctx.FormValue("userIds")
	fmt.Println("create userId", userId);
	
	total, err := BatchInsertWhite(userId, appid);
	fmt.Println("total = ", total)
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		return
	}

	//ctx.JSON("{}")
}

func handleDeleteWhite(ctx context.Context) {
	appid := util.GetAppID(ctx)
	// parse QueryCondition
	userId := ctx.FormValue("userIds")
	fmt.Println("del userId", userId);
	
	total, err := BatchDeleteWhite(userId, appid);
	fmt.Println("total = ", total)
	if err != nil {
		ctx.StatusCode(http.StatusBadRequest)
		return
	}

	//ctx.JSON("{}")
}