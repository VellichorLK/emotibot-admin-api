package cu

import (
	"math/rand"
	"net/http"
	"time"

	"emotibot.com/emotigo/pkg/logger"

	"emotibot.com/emotigo/module/admin-api/ApiError"
	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/admin-api/util/requestheader"
)

var (
	// ModuleInfo is needed for module define
	ModuleInfo  util.ModuleInfo
	maxDirDepth int
)

func random(min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max-min) + min
}

func handleTextProcess(w http.ResponseWriter, r *http.Request) {
	mockEmotions := []string{
		"不满",
		"称赞",
		"不喜欢",
		"高兴",
		"伤心",
		"害怕",
	}

	type RequestObj struct {
		Text string `json:"text"`
	}

	type EmotionObj struct {
		Label string `json:"label"`
	}
	type ResponseObj struct {
		Text    string       `json:"text"`
		Emotion []EmotionObj `json:"emotion"`
	}

	reqBody := []RequestObj{}
	err := util.ReadJSON(r, &reqBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	responseBody := []ResponseObj{}
	for _, textObj := range reqBody {
		ind := random(0, 6)
		emotion := mockEmotions[ind]

		responseObj := ResponseObj{
			Text: textObj.Text,
			Emotion: []EmotionObj{
				EmotionObj{
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
