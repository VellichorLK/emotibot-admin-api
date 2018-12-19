package cu

import (
	"math/rand"
	"net/http"
	"time"

	"emotibot.com/emotigo/module/admin-api/util/AdminErrors"

	"emotibot.com/emotigo/module/admin-api/ApiError"
	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/admin-api/util/requestheader"
	emotionengine "emotibot.com/emotigo/pkg/api/emotion-engine/v1"
	"emotibot.com/emotigo/pkg/logger"
)

func random(min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max-min) + min
}

var adminErrInitialFailed = AdminErrors.New(AdminErrors.ErrnoInitfailed, "cu package init failed")

type processResponseBody []processedText

type processedText struct {
	Text    string        `json:"text"`
	Emotion []emotionData `json:"emotion"`
}
type emotionData struct {
	Label string `json:"label"`
}

func handleTextProcess(w http.ResponseWriter, r *http.Request) {
	if emotionPredict == nil {
		util.Return(w, adminErrInitialFailed, nil)
		return
	}

	type processRequest struct {
		Text string `json:"text"`
	}

	sentences := []processRequest{}
	err := util.ReadJSON(r, &sentences)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	predictsGroup := make(map[string][]emotionengine.Predict, len(sentences))
	for _, s := range sentences {
		req := emotionengine.PredictRequest{
			AppID:    "demo",
			Sentence: s.Text,
		}
		predictions, err := emotionPredict(req)
		if err != nil {
			wrappederr := AdminErrors.New(AdminErrors.ErrnoAPIError, "call prediction failed, "+err.Error())
			util.Return(w, wrappederr, nil)
			return
		}
		predictsGroup[s.Text] = predictions
	}

	responseBody := processResponseBody{}
	for _, s := range sentences {
		predictions := predictsGroup[s.Text]
		dataGroup := make([]emotionData, 0, 0)
		for _, p := range predictions {
			if p.Score < filterScore {
				continue
			}
			data := emotionData{
				Label: p.Label,
			}
			dataGroup = append(dataGroup, data)
		}

		re := processedText{
			Text:    s.Text,
			Emotion: dataGroup,
		}

		responseBody = append(responseBody, re)
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

func handleFlowAdd(w http.ResponseWriter, r *http.Request) {
	uuid := util.GetMuxVar(r, "id")

	var requestBody []*apiFlowAddBody
	err := util.ReadJSON(r, &requestBody)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
		return
	}

	if len(requestBody) == 0 {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, "empty sentence"), http.StatusBadRequest)
		return
	}

	//insert the segment
	err = insertSegmentByUUID(uuid, requestBody)
	if err != nil {
		if err == ErrSpeaker {
			util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
		} else {
			logger.Error.Printf("%s\n", err)
			util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		}
		return
	}

	//get the current segment by order
	segments, err := getFlowSentences(uuid)
	if err != nil {
		logger.Error.Printf("%s\n", err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}

	//transform the segment to cu-predict-request
	predictReq := segmentToV1PredictRequest(segments)

	//FIXME get app_id here
	predictContext := &V1PredictContext{AppID: 1, Threshold: 50, Data: predictReq}

	predictResult, err := predictByV1CuModule(predictContext)
	if err != nil {
		logger.Error.Printf("%s\n", err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.OPENAPI_URL_ERROR, err.Error()), http.StatusInternalServerError)
	}

	err = util.WriteJSON(w, predictResult)
	if err != nil {
		logger.Error.Printf("%s\n", err)
	}

}
