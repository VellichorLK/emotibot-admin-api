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

type Exceptions struct {
	Staff    []model.SimpleSentence `json:"staff"`
	Customer []model.SimpleSentence `json:"customer"`
}

type SensitiveWordInDetail struct {
	UUID       string     `json:"sw_id"`
	Name       string     `json:"sw_name"`
	Score      int        `json:"score"`
	Exception  Exceptions `json:"execption"`
	CategoryID int64
}

type Req struct {
	Name string `json:"name"`
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

	category := &model.CategortInfo{
		Name: reqBody.Name,
		ID:   uint64(id),
	}

	err = util.WriteJSON(w, category)
	if err != nil {
		logger.Error.Printf("response category failed, err: %s", err.Error())
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}
}

func handleGetSensitiveWords(w http.ResponseWriter, r *http.Request) {
	enterprise := requestheader.GetEnterpriseID(r)
	paging := request.Paging(r)
	vars := mux.Vars(r)

	var deleted int8
	filter := &model.SensitiveWordFilter{
		Enterprise: &enterprise,
		Page:       paging.Page,
		Limit:      paging.Limit,
		Deleted:    &deleted,
	}

	if keyword, ok := vars["keyword"]; ok {
		filter.Keyword = keyword
	}

	total, words, err := GetSensitiveWords(filter)
	if err != nil {
		logger.Error.Printf("get sensitive words failed, err: %s", err.Error())
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

func handleGetSensitiveWord(w http.ResponseWriter, r *http.Request) {
	enterprise := requestheader.GetEnterpriseID(r)
	wUUID := general.ParseID(r)

	word, err := GetSensitiveWordInDetail(wUUID, enterprise)
	if err != nil {
		logger.Error.Printf("get sensitive word(%s) failed, err: %s", wUUID, err.Error())
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}

	if word == nil {
		http.NotFound(w, r)
		return
	}
	wordInDetail := SensitiveWordInDetail{
		UUID:  word.UUID,
		Name:  word.Name,
		Score: word.Score,
		Exception: Exceptions{
			Staff:    word.StaffException,
			Customer: word.CustomerException,
		},
		CategoryID: word.CategoryID,
	}

	util.WriteJSON(w, wordInDetail)
	return
}

func handleUpdateSensitiveWord(w http.ResponseWriter, r *http.Request) {
	enterprise := requestheader.GetEnterpriseID(r)
	wUUID := general.ParseID(r)

	swInReq := SensitiveWordInReq{}
	err := util.ReadJSON(r, &swInReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	staffException := []model.SimpleSentence{}
	for _, sUUID := range swInReq.Exception.Staff {
		ss := model.SimpleSentence{
			UUID: sUUID,
		}
		staffException = append(staffException, ss)
	}

	customerException := []model.SimpleSentence{}
	for _, sUUID := range swInReq.Exception.Staff {
		ss := model.SimpleSentence{
			UUID: sUUID,
		}
		customerException = append(customerException, ss)
	}

	word := &model.SensitiveWord{
		UUID:              wUUID,
		Name:              swInReq.Name,
		Enterprise:        enterprise,
		Score:             swInReq.Score,
		CustomerException: customerException,
		StaffException:    staffException,
	}

	err = UpdateSensitiveWord(word)
	if err != nil {
		logger.Error.Printf("update sensitive word failed, err: %s", err.Error())
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}
	return
}

func hanldeDeleteSensitiveWord(w http.ResponseWriter, r *http.Request) {
	enterprise := requestheader.GetEnterpriseID(r)
	id := general.ParseID(r)

	err := DeleteSensitiveWord(id, enterprise)
	if err != nil {
		logger.Error.Printf("update sensitive word failed, err: %s", err.Error())
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}
	return
}

func handleMoveSensitiveWords(w http.ResponseWriter, r *http.Request) {
	enterprise := requestheader.GetEnterpriseID(r)
	categoryId, err := parseID(r)
	if err != nil {
		util.WriteJSONWithStatus(w, err, http.StatusBadRequest)
		return
	}

	var UUID []string
	err = util.ReadJSON(r, &UUID)
	if err != nil {
		util.WriteJSONWithStatus(w, err, http.StatusBadRequest)
		return
	}

	err = MoveSensitiveWord(UUID, enterprise, categoryId)
	if err != nil {
		logger.Error.Printf("move sensitive words failed, err: %s", err.Error())
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}
	return
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

	var deleted int8
	filter := &model.SensitiveWordFilter{
		Category:   &categoryID,
		Enterprise: &enterprise,
		Deleted:    &deleted,
	}
	total, words, err := GetSensitiveWords(filter)
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
