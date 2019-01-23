package qi

import (
	"net/http"
	"strconv"

	"emotibot.com/emotigo/module/admin-api/ApiError"
	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/admin-api/util/requestheader"
	"emotibot.com/emotigo/pkg/logger"
)

func handleTrainAllTags(w http.ResponseWriter, r *http.Request) {
	enterprise := requestheader.GetEnterpriseID(r)
	models, err := TrainModelByEnterprise(enterprise)
	if err != nil {
		logger.Error.Printf("train failed. %s\n", err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}
	w.Write([]byte(strconv.FormatInt(models, 10)))
}

func handleUnload(w http.ResponseWriter, r *http.Request) {
	err := UnloadAllTags()
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}
}
