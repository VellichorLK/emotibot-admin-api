package qi

import (
	"net/http"
	"strconv"

	"emotibot.com/emotigo/pkg/logger"

	"emotibot.com/emotigo/module/admin-api/ApiError"
	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/admin-api/util/requestheader"
	model "emotibot.com/emotigo/module/qic-api/model/v1"
)

func handleGetCredit(w http.ResponseWriter, r *http.Request) {
	enterprise := requestheader.GetEnterpriseID(r)
	callStr := parseID(r)

	//check the category authorization
	call, err := strconv.ParseUint(callStr, 10, 64)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
		return
	}

	resp, err := RetrieveCredit(call)
	if err != nil {
		logger.Error.Printf("get credit failed.\n")
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}

	_ = enterprise

	err = util.WriteJSON(w, resp)
	if err != nil {
		logger.Error.Printf("%s\n", err)
	}
}

//WithCallIDCheck checks the call id and it's belongings
func WithCallIDCheck(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		enterprise := requestheader.GetEnterpriseID(r)
		callStr := parseID(r)
		call, err := strconv.ParseInt(callStr, 10, 64)
		if err != nil {
			util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
			return
		}
		q := model.CallQuery{ID: []int64{call}, EnterpriseID: &enterprise}
		count, err := callDao.Count(nil, q)
		if err != nil {
			logger.Error.Printf("count call failed. %s\n", err)
			util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
			return
		}
		if count != 1 {
			util.WriteJSONWithStatus(w, "", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	}
}
