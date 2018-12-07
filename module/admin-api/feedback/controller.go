package feedback

import (
	"fmt"
	"net/http"

	"emotibot.com/emotigo/module/admin-api/util/AdminErrors"

	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/admin-api/util/requestheader"
)

var (
	getReasonService = GetReasons
	addReasonService = AddReason
	delReasonService = DeleteReason
)

func handleGetFeedbackReasons(w http.ResponseWriter, r *http.Request) {
	appid := requestheader.GetAppID(r)
	reasons, err := getReasonService(appid)
	if err != nil {
		util.ReturnError(w, err.Errno(), err.Error())
	} else {
		util.Return(w, nil, reasons)
	}
}

func handleAddFeedbackReason(w http.ResponseWriter, r *http.Request) {
	appid := requestheader.GetAppID(r)

	type input struct {
		Content string `json:"content"`
	}
	params := input{}
	jsonErr := util.ReadJSON(r, &params)
	if jsonErr != nil {
		util.ReturnError(w, AdminErrors.ErrnoRequestError,
			fmt.Sprintf("Invalid json: %s", jsonErr.Error()))
	}

	reasons, err := addReasonService(appid, params.Content)
	if err != nil {
		util.ReturnError(w, err.Errno(), err.Error())
	} else {
		util.Return(w, nil, reasons)
	}
}

func handleDeleteFeedbackReason(w http.ResponseWriter, r *http.Request) {
	appid := requestheader.GetAppID(r)
	id, muxErr := util.GetMuxInt64Var(r, "id")

	if muxErr != nil {
		util.ReturnError(w, AdminErrors.ErrnoRequestError,
			fmt.Sprintf("Invalid json: %s", muxErr.Error()))
	}

	err := delReasonService(appid, id)
	if err != nil {
		util.ReturnError(w, err.Errno(), err.Error())
	} else {
		util.Return(w, nil, nil)
	}
}
