package Robot

import (
	"net/http"
	"strconv"

	"emotibot.com/emotigo/module/admin-api/ApiError"
	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/admin-api/util/requestheader"
)

func handleChatQAList(w http.ResponseWriter, r *http.Request) {
	appid := requestheader.GetAppID(r)
	keyword := r.FormValue("keyword")
	curPageStr := r.FormValue("curPage")
	pageLimitStr := r.FormValue("pageLimit")

	curPage, err := strconv.Atoi(curPageStr)
	if err != nil {
		curPage = 1
	}
	pageLimit, err := strconv.Atoi(pageLimitStr)
	if err != nil {
		pageLimit = 10
	}

	chatDataList, errCode, err := GetChatQAList(appid, keyword, curPage, pageLimit)

	if err != nil {
		util.WriteJSON(w, util.GenRetObj(errCode, err.Error()))
		return
	}
	util.WriteJSON(w, util.GenRetObj(ApiError.SUCCESS, chatDataList))
}
