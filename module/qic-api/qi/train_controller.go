package qi

import (
	"net/http"

	"emotibot.com/emotigo/module/admin-api/ApiError"
	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/pkg/logger"
)

func handleTrainAllTags(w http.ResponseWriter, r *http.Request) {
	err := TrainAllTags()
	if err != nil {
		logger.Error.Printf("train failed. %s\n", err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return

	}
}

func handleUnload(w http.ResponseWriter, r *http.Request) {
	err := UnloadAllTags()
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}
}
