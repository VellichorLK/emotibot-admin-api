package Gray

import (
	//"encoding/json"
	//"fmt"
	"strconv"
	//"strings"
	//"sort"
	// "strings"

	//"emotibot.com/emotigo/module/vipshop-admin/util"
)

func ParseCondition(param Parameter) (QueryCondition, error) {
	curPage := param.FormValue("cur_page")
	limit := param.FormValue("page_limit")

	var condition = QueryCondition{
		Keyword:                param.FormValue("key_word"),
		Limit:                  10,
		CurPage:                0,
	}

	page, err := strconv.Atoi(curPage)
	if err == nil {
		if page < 0 {
			page = 0
		}
		condition.CurPage = page
	}

	pageLimit, err := strconv.Atoi(limit)
	if err == nil {
		condition.Limit = pageLimit
	}

	return condition, nil
}