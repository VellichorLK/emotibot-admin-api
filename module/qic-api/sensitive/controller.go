package sensitive

import (
	"emotibot.com/emotigo/module/admin-api/ApiError"
	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/admin-api/util/requestheader"
	"emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/pkg/logger"
	"net/http"
)

type ExceptionInReq struct {
	Staff    []string `json:"staff"`
	Customer []string `json:"customer"`
}

type SensitiveWordInReq struct {
	Name      string         `json:"sw_name"`
	Score     int            `json:"score"`
	Exception ExceptionInReq `json:"exception"`
}

func transformSensitiveWordInReqToSensitiveWord(inreq *SensitiveWordInReq) (word *model.SensitiveWord) {
	if inreq == nil {
		return
	}

	word = &model.SensitiveWord{
		Name:  inreq.Name,
		Score: inreq.Score,
	}

	customerException := make([]model.SimpleSentence, len(inreq.Exception.Customer))
	for idx, uid := range inreq.Exception.Customer {
		customerException[idx] = model.SimpleSentence{
			UUID: uid,
		}
	}

	staffException := make([]model.SimpleSentence, len(inreq.Exception.Staff))
	for idx, uid := range inreq.Exception.Staff {
		staffException[idx] = model.SimpleSentence{
			UUID: uid,
		}
	}
	word.CustomerException = customerException
	word.StaffException = staffException
	return
}

func handleCreateSensitiveWord(w http.ResponseWriter, r *http.Request) {
	enterprise := requestheader.GetEnterpriseID(r)

	swInReq := SensitiveWordInReq{}
	err := util.ReadJSON(r, &swInReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	uid, err := CreateSensitiveWord(swInReq.Name, enterprise, swInReq.Score, swInReq.Exception.Customer, swInReq.Exception.Staff)
	if err != nil {
		logger.Error.Printf("create sensitive word failed after CreateSensitiveWord, reason: %s", err.Error())
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}

	type Response struct {
		UUID string `json:"sw_id"`
	}

	response := Response{
		UUID: uid,
	}

	util.WriteJSON(w, response)
	return
}

func handleCreateSensitiveWordCategory(w http.ResponseWriter, r *http.Request) {
	enterprise := requestheader.GetEnterpriseID(r)

	type Req struct {
		Name string `json:"name"`
	}

	reqBody := Req{}
	err := util.ReadJSON(r, &reqBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := CreateSensitiveWordCategory(reqBody.Name, enterprise)
	if err != nil {
		logger.Error.Printf("create sensitive word failed, err: %s", err.Error())
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}

	category := &model.SensitiveWordCategory{
		Name: reqBody.Name,
		ID:   id,
	}

	err = util.WriteJSON(w, category)
	if err != nil {
		logger.Error.Printf("response category failed, err: %s", err.Error())
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}
}

func handleGetCategory(w http.ResponseWriter, r *http.Request) {
	enterprise := requestheader.GetEnterpriseID(r)

	categories, err := GetCategories(enterprise)
	if err != nil {
		logger.Error.Printf("get sensitive word categories failed after GetCategories, reason: %s", err.Error())
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}

	err = util.WriteJSON(w, categories)
	if err != nil {
		logger.Error.Printf("response categories failed, err: %s", err.Error())
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}
}
