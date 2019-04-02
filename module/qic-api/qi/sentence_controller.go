package qi

import (
	"errors"
	"net/http"
	"strconv"

	"emotibot.com/emotigo/module/admin-api/ApiError"
	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/admin-api/util/requestheader"
	"emotibot.com/emotigo/pkg/logger"
)

type sentencesResp struct {
	Page pageResp        `json:"paging"`
	Data []*DataSentence `json:"data"`
}
type pageResp struct {
	Current int    `json:"current"`
	Total   uint64 `json:"total"`
	Limit   int    `json:"limit"`
}

type sentenceReq struct {
	Name       string   `json:"sentence_name"`
	CategoryID uint64   `json:"category_id,string"`
	Tags       []string `json:"tags"`
}

type sentenceResp struct {
	UUID string `json:"sentence_id"`
}

func categoryCheck(enterprise string, categories []uint64) (bool, error) {
	//FIXME, count the category
	return true, nil
}

func handleGetSentences(w http.ResponseWriter, r *http.Request) {
	enterprise := requestheader.GetEnterpriseID(r)
	page, limit, err := getPageLimit(r)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err), http.StatusBadRequest)
		return
	}
	id, err := getQueryCategoryID(r)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err), http.StatusBadRequest)
		return
	}
	var isDelete int8
	total, sentences, err := GetSentenceList(enterprise, page, limit, &isDelete, id, r.FormValue("keyword"))
	if err != nil {
		logger.Error.Printf("get sentence list failed. %s\n", err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}

	var resp sentencesResp
	resp.Page.Current = page
	resp.Page.Limit = limit
	resp.Page.Total = total

	resp.Data = sentences

	err = util.WriteJSON(w, resp)
	if err != nil {
		logger.Error.Printf("%s\n", err)
	}

}

func handleMoveSentence(w http.ResponseWriter, r *http.Request) {
	enterprise := requestheader.GetEnterpriseID(r)
	categoryStr := parseID(r)

	//check the category authorization
	category, err := strconv.ParseUint(categoryStr, 10, 64)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
		return
	}
	valid, err := categoryCheck(enterprise, []uint64{category})
	if err != nil {
		logger.Error.Printf("check category auth failed. %s\n", err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}
	if !valid {
		util.WriteJSONWithStatus(w, "", http.StatusUnauthorized)
		return
	}

	//check sentence authorization
	var sentences []string
	err = util.ReadJSON(r, &sentences)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
		return
	}

	valid, err = CheckSentenceAuth(sentences, enterprise)
	if err != nil {
		logger.Error.Printf("check sentence auth failed. %s\n", err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}
	if !valid {
		util.WriteJSONWithStatus(w, "", http.StatusUnauthorized)
		return
	}

	_, err = MoveCategories(sentences, enterprise, category)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}

}

func handleGetSentence(w http.ResponseWriter, r *http.Request) {
	uuid := parseID(r)
	enterprise := requestheader.GetEnterpriseID(r)
	sentence, err := GetSentence(uuid, enterprise)
	if err != nil {
		logger.Error.Printf("get sentence failed. %s\n", err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}

	err = util.WriteJSON(w, sentence)
	if err != nil {
		logger.Error.Printf("%s\n", err)
	}
}

func handleNewSentence(w http.ResponseWriter, r *http.Request) {
	enterprise := requestheader.GetEnterpriseID(r)

	var requestBody sentenceReq
	err := util.ReadJSON(r, &requestBody)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
		return
	}

	auth, err := categoryCheck(enterprise, []uint64{requestBody.CategoryID})
	if err != nil {
		logger.Error.Printf("get the category failed. %s\n", err)
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}

	if !auth {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, "no such category id"), http.StatusBadRequest)
		return
	}

	d, err := NewSentence(enterprise, requestBody.CategoryID, requestBody.Name, requestBody.Tags)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}

	var resp sentenceResp
	resp.UUID = d.UUID

	err = util.WriteJSON(w, resp)
	if err != nil {
		logger.Error.Printf("%s\n", err)
	}

}

func handleModifySentence(w http.ResponseWriter, r *http.Request) {
	enterprise := requestheader.GetEnterpriseID(r)
	uuid := parseID(r)

	var requestBody sentenceReq
	err := util.ReadJSON(r, &requestBody)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, err.Error()), http.StatusBadRequest)
		return
	}

	affected, err := UpdateSentence(uuid, requestBody.Name, enterprise, requestBody.Tags)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}

	if affected == 0 {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, "No record is deleted"), http.StatusBadRequest)
		return
	}

}

func handleDeleteSentence(w http.ResponseWriter, r *http.Request) {
	enterprise := requestheader.GetEnterpriseID(r)
	uuid := parseID(r)
	affected, err := SoftDeleteSentence(uuid, enterprise)
	if err != nil {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
		return
	}
	if affected == 0 {
		util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, "No record is deleted"), http.StatusBadRequest)
		return
	}
}

var (
	errPage  = errors.New("Error page")
	errLimit = errors.New("Error limit")
)

func getQueryCategoryID(r *http.Request) (*uint64, error) {
	params := r.URL.Query()
	category := params.Get("category_id")
	if category == "" {
		return nil, nil
	}
	var id uint64
	var err error
	id, err = strconv.ParseUint(category, 10, 64)
	return &id, err
}

// getPageLimit find the page and limit in the r's query string.
// If not found, page and limit will be 1, 0
// If a invalid value is given, errLimit or errPage will be returned.
func getPageLimit(r *http.Request) (page int, limit int, err error) {
	params := r.URL.Query()
	limitStr := params.Get("limit")
	pageStr := params.Get("page")

	if limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit < 0 {
			return 0, 0, errLimit
		}
	}

	if pageStr != "" {
		page, err = strconv.Atoi(pageStr)
		if err != nil || page <= 0 {
			return 0, 0, errPage
		}
	} else {
		page = DPage
	}
	return page, limit, nil
}

//WithSenUUIDCheck checks the uuid
func WithSenUUIDCheck(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uuid := parseID(r)
		enterprise := requestheader.GetEnterpriseID(r)
		if len(uuid) != 32 {
			util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, "no such id"), http.StatusBadRequest)
			return
		}

		valid, err := CheckSentenceAuth([]string{uuid}, enterprise)
		if err != nil {
			logger.Error.Printf("check sentence auth failed. %s\n", err)
			util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.DB_ERROR, err.Error()), http.StatusInternalServerError)
			return
		}
		if !valid {
			util.WriteJSONWithStatus(w, "", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	}
}

//WithEnterpriseCheck checks the enterprise
func WithEnterpriseCheck(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		enterprise := requestheader.GetEnterpriseID(r)
		if enterprise == "" {
			util.WriteJSONWithStatus(w, util.GenRetObj(ApiError.REQUEST_ERROR, "no enterprise"), http.StatusBadRequest)
			return
		}
		next.ServeHTTP(w, r)
	}
}
