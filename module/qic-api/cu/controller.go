package cu

import (
	"math/rand"
	"net/http"
	"time"

	"emotibot.com/emotigo/module/admin-api/util/AdminErrors"

	"emotibot.com/emotigo/module/admin-api/ApiError"
	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/admin-api/util/requestheader"
	"emotibot.com/emotigo/pkg/logger"
)

func random(min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max-min) + min
}

var adminErrInitialFailed = AdminErrors.New(AdminErrors.ErrnoInitfailed, "cu package init failed")

func handleTextProcess(w http.ResponseWriter, r *http.Request) {
	if emotionPredict == nil {
		util.Return(w, adminErrInitialFailed, nil)
	}

	type processRequest struct {
		Text string `json:"text"`
	}

	type emotionData struct {
		Label string `json:"label"`
	}
	type result struct {
		Text    string        `json:"text"`
		Emotion []emotionData `json:"emotion"`
	}

	reqBody := []processRequest{}
	err := util.ReadJSON(r, &reqBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	responseBody := []result{}
	for _, textObj := range reqBody {
		emotion := "test"

		responseObj := result{
			Text: textObj.Text,
			Emotion: []emotionData{
				emotionData{
					Label: emotion,
				},
			},
		}

		responseBody = append(responseBody, responseObj)
	}

	util.WriteJSON(w, responseBody)
}

func handleFlowCreate(w http.ResponseWriter, r *http.Request) {
	//get the first available bot and its first scenario
	enterprise := requestheader.GetEnterpriseID(r)
	user := requestheader.GetUserID(r)
	//appid := requestheader.GetAppID(r)

	var requestBody apiFlowCreateBody
	err := util.ReadJSON(r, &requestBody)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
		return
	}

	uuid, err := createFlowConversation(enterprise, user, &requestBody)
	if err != nil {
		logger.Error.Printf("%s\n", err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}

	resp := &apiFlowCreateResp{UUID: uuid}
	err = util.WriteJSON(w, resp)
	if err != nil {
		logger.Error.Printf("%s\n", err)
	}
}
