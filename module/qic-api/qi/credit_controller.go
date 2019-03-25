package qi

import (
	"net/http"

	"emotibot.com/emotigo/module/admin-api/ApiError"
	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/admin-api/util/requestheader"
	model "emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/module/qic-api/util/general"
	"emotibot.com/emotigo/pkg/logger"
)

func handleGetCredit(w http.ResponseWriter, r *http.Request) {
	uuid := general.ParseID(r)

	resp, err := RetrieveCredit(uuid)
	if err != nil {
		logger.Error.Printf("get credit failed.\n")
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}

	err = util.WriteJSON(w, resp)
	if err != nil {
		logger.Error.Printf("%s\n", err)
	}
}

//WithCallIDCheck checks the call id and it's belongings
func WithCallIDCheck(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		enterprise := requestheader.GetEnterpriseID(r)
		uuid := general.ParseID(r)

		q := model.CallQuery{UUID: []string{uuid}, EnterpriseID: &enterprise}
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
