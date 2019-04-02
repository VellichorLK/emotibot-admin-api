package qi

import (
	"net/http"

	"emotibot.com/emotigo/module/admin-api/ApiError"
	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/qic-api/util/general"
	"emotibot.com/emotigo/pkg/logger"
)

func handleGetGroupedCredit(w http.ResponseWriter, r *http.Request) {
	uuid := general.ParseID(r)

	resp, err := RetrieveGroupedCredit(uuid)
	if err != nil {
		logger.Error.Printf("get grouped credit failed.\n")
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}

	err = util.WriteJSON(w, resp)
	if err != nil {
		logger.Error.Printf("%s\n", err)
	}
}
