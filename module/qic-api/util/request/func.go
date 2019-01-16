package request

import (
	"emotibot.com/emotigo/module/qic-api/util/general"
	"net/http"
	"strconv"
)

func Paging(r *http.Request) *general.Paging {
	var err error
	vals := r.URL.Query()

	paging := &general.Paging{}

	pageStr := vals["page"]
	if len(pageStr) > 0 {
		paging.Page, err = strconv.Atoi(pageStr[0])
		if err != nil {
			paging.Page = 0
		}
	}

	limitStr := vals["limit"]
	if len(limitStr) > 0 {
		paging.Limit, err = strconv.Atoi(limitStr[0])
		if err != nil {
			paging.Limit = 0
		}
	}
	return paging
}
