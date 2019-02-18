package sensitive

import (
	"emotibot.com/emotigo/module/admin-api/ApiError"
	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/admin-api/util/requestheader"
	"emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/module/qic-api/util/general"
	"emotibot.com/emotigo/module/qic-api/util/request"
	"emotibot.com/emotigo/pkg/logger"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
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

type SensitiveWordInRes struct {
	UUID     string `json:"sw_id"`
	Name     string `json:"sw_name"`
	Category int64  `json:"category_id"`
}

type Req struct {
	Name string `json:"name"`
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

func toSensitiveWordInRes(words []model.SensitiveWord) (inres []SensitiveWordInRes) {
	inres = []SensitiveWordInRes{}
	for _, w := range words {
		inres = append(
			inres,
			SensitiveWordInRes{
				UUID:     w.UUID,
				Name:     w.Name,
				Category: w.CategoryID,
			},
		)
	}
	return
}

func parseID(r *http.Request) (int64, error) {
	vars := mux.Vars(r)
	return strconv.ParseInt(vars["id"], 10, 64)
}

func handleGetWordsUnderCategory(w http.ResponseWriter, r *http.Request) {
	enterprise := requestheader.GetEnterpriseID(r)
	paging := request.Paging(r)

	categoryID, err := parseID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	total, words, err := GetWordsUnderCategory(categoryID, enterprise)
	if err != nil {
		logger.Error.Printf("get words under categories failed, err: %s", err.Error())
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}

	type Response struct {
		Paging *general.Paging      `json:"paging"`
		Data   []SensitiveWordInRes `json:"data"`
	}

	wordsInRes := toSensitiveWordInRes(words)
	paging.Total = total
	response := Response{
		Paging: paging,
		Data:   wordsInRes,
	}

	util.WriteJSON(w, response)
	return
}

func handleUpdateCategory(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	reqBody := Req{}
	err = util.ReadJSON(r, &reqBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	category := &model.CategoryRequest{
		Name: reqBody.Name,
	}

	err = UpdateCategory(id, category)
	if err != nil {
		logger.Error.Printf("update category failed, err: %s", err.Error())
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}
}

func handleDeleteCategory(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	enterprise := requestheader.GetEnterpriseID(r)

	_, err = DeleteCategory(id, enterprise)
	if err != nil {
		logger.Error.Printf("delete category failed, err: %s", err.Error())
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}
}
