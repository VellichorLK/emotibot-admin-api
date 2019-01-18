package cu

import (
	"errors"
	"net/http"
	"strconv"

	"emotibot.com/emotigo/module/qic-api/model/v1"

	"emotibot.com/emotigo/module/admin-api/util/AdminErrors"
	"emotibot.com/emotigo/module/qic-api/sensitive"

	"emotibot.com/emotigo/module/admin-api/ApiError"
	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/admin-api/util/requestheader"
	emotionengine "emotibot.com/emotigo/pkg/api/emotion-engine/v1"
	"emotibot.com/emotigo/pkg/logger"
)

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

func newAsrContent(content apiFlowAddBody) *model.AsrContent {
	return &model.AsrContent{
		StartTime: content.StartTime,
		EndTime:   content.EndTime,
		Text:      content.Text,
		Speaker:   content.Speaker,
	}
}

func handleFlowAdd(w http.ResponseWriter, r *http.Request) {
	uuid := util.GetMuxVar(r, "id")

	var requestBody []apiFlowAddBody
	err := util.ReadJSON(r, &requestBody)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
		return
	}

	if len(requestBody) == 0 {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, "empty sentence"), http.StatusBadRequest)
		return
	}
	var asrContents = make([]*model.AsrContent, 0)
	for _, b := range requestBody {
		asrContents = append(asrContents, newAsrContent(b))
	}
	//insert the segment
	err = insertSegmentByUUID(uuid, asrContents)
	if err != nil {
		if err == ErrSpeaker {
			util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
		} else {
			logger.Error.Printf("%s\n", err)
			util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		}
		return
	}
	enterprise := requestheader.GetEnterpriseID(r)
	resp, err := getCurrentQIFlowResult(w, enterprise, uuid)
	if err != nil {
		return
	}

	callID, err := getIDByUUID(uuid)
	if err != nil {
		logger.Error.Printf("%s\n", err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}
	if callID == 0 {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, "no such id is found"), http.StatusBadRequest)
		return
	}

	_, err = UpdateFlowResult(callID, resp)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}

	resp.Sensitive = make([]string, 0)

	for i := 0; i < len(requestBody); i++ {

		words, err := sensitive.IsSensitive(requestBody[i].Text)
		if err != nil {
			logger.Error.Printf("get sensitive words failed. %s\n", err)
			util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
			return
		}

		resp.Sensitive = append(resp.Sensitive, words...)
	}

	err = util.WriteJSON(w, resp)
	if err != nil {
		logger.Error.Printf("%s\n", err)
	}
}

func getCurrentQIFlowResult(w http.ResponseWriter, enterprise string, uuid string) (*model.QIFlowResult, error) {

	conversation, err := getConversation(uuid)

	//get the current segment by order
	segments, err := getFlowSentences(uuid)
	if err != nil {
		logger.Error.Printf("%s\n", err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return nil, err
	}

	if err != nil {
		logger.Error.Printf("%s\n", err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return nil, err
	}

	//transform the segment to cu-predict-request
	predictReq := segmentToV1PredictRequest(segments)

	//get the group info for flow usage
	groups, err := GetFlowGroup(enterprise)
	if err != nil {
		logger.Error.Printf("%s\n", err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return nil, err
	}

	if len(groups) == 0 {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, "no group for flow usage"), http.StatusBadRequest)
		return nil, errors.New("no group for flow usage")
	}

	resp := &model.QIFlowResult{FileName: conversation.FileName}

	for i := 0; i < len(groups); i++ {
		appIDStr := strconv.FormatInt(groups[i].ID, 10)
		predictContext := &V1PredictContext{AppID: appIDStr, Threshold: 50, Data: predictReq}

		predictResult, err := predictByV1CuModule(predictContext)
		if err != nil {
			util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.OPENAPI_URL_ERROR, err.Error()), http.StatusInternalServerError)
			return nil, err
		}

		qiResult, err := GetRuleLogic(uint64(groups[0].ID))
		if err != nil {
			util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
			return nil, err
		}

		err = FillCUCheckResult(predictResult, qiResult)
		if err != nil {
			util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
			return nil, err
		}

		groupRes := &model.QIFlowGroupResult{Name: groups[i].Name}
		groupRes.QIResult = qiResult

		resp.Result = append(resp.Result, groupRes)
	}

	return resp, nil

}

func handleFlowResult(w http.ResponseWriter, r *http.Request) {
	uuid := util.GetMuxVar(r, "id")

	callID, err := getIDByUUID(uuid)
	if err != nil {
		logger.Error.Printf("%s\n", err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}
	if callID == 0 {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, "no such id is found"), http.StatusBadRequest)
		return
	}
	resp, err := GetFlowResult(callID)
	if err != nil {
		logger.Error.Printf("%s\n", err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}

	err = util.WriteJSON(w, resp)
	if err != nil {
		logger.Error.Printf("%s\n", err)
	}
}

func handleFlowFinish(w http.ResponseWriter, r *http.Request) {

	var requestBody apiFlowFinish
	err := util.ReadJSON(r, &requestBody)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
		return
	}

	uuid := util.GetMuxVar(r, "id")
	enterprise := requestheader.GetEnterpriseID(r)
	resp, err := getCurrentQIFlowResult(w, enterprise, uuid)
	if err != nil {
		return
	}
	_ = resp

	err = FinishFlowQI(&requestBody, uuid, resp)
	if err != nil {
		if err == ErrEndTimeSmaller {
			util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
		} else {
			logger.Error.Printf("Finish the qi flow failed. %s\n", err)
			util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		}

		return
	}

}
