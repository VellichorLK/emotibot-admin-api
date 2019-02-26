package qi

import (
	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/admin-api/util/requestheader"
	"emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/module/qic-api/util/general"
	"emotibot.com/emotigo/module/qic-api/util/request"
	"emotibot.com/emotigo/pkg/logger"
	"net/http"
	"strconv"
)

func handleCreateCategory(w http.ResponseWriter, r *http.Request) {
	enterprise := requestheader.GetEnterpriseID(r)
	if enterprise == "" {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	category := model.CategortInfo{}
	err := util.ReadJSON(r, &category)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	category.Enterprise = enterprise
	categoryID, err := CreateCategory(&category)
	if err != nil {
		logger.Error.Printf("error while create category in handleCreateCategory, reason: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	type Response struct {
		ID int64 `json:"category_id"`
	}

	response := Response{
		ID: categoryID,
	}

	util.WriteJSON(w, response)
}

func handleGetCategoryies(w http.ResponseWriter, r *http.Request) {
	enterprise := requestheader.GetEnterpriseID(r)
	if enterprise == "" {
		http.Error(w, "no enterprise", http.StatusBadRequest)
		return
	}

	isDelete := int8(0)
	var ctype int8
	query := &model.CategoryQuery{
		Enterprise: &enterprise,
		IsDelete:   &isDelete,
		Type:       &ctype,
	}

	categories, err := GetCategories(query)
	if err != nil {
		logger.Error.Printf("error while get categories in handleGetCategories, reason: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	util.WriteJSON(w, categories)
}

func handleGetCategory(w http.ResponseWriter, r *http.Request) {
	enterprise := requestheader.GetEnterpriseID(r)
	idStr := general.ParseID(r)
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	paging := request.Paging(r)

	isDelete := int8(0)
	query := &model.SentenceQuery{
		CategoryID: &id,
		Enterprise: &enterprise,
		IsDelete:   &isDelete,
		Page:       paging.Page,
		Limit:      paging.Limit,
	}

	total, sentences, err := GetCategorySentences(query)
	if err != nil {
		logger.Error.Printf("error while get category sentences in handleGetCategory, reason: %s", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	type Response struct {
		Paging *general.Paging        `json:"paging"`
		Data   []model.SimpleSentence `json:"data"`
	}

	simpleSentences := []model.SimpleSentence{}
	for _, sentence := range sentences {
		simpleSentence := model.SimpleSentence{
			UUID:       sentence.UUID,
			Name:       sentence.Name,
			CategoryID: sentence.CategoryID,
		}
		simpleSentences = append(simpleSentences, simpleSentence)
	}

	paging.Total = int64(total)
	response := Response{
		Paging: paging,
		Data:   simpleSentences,
	}

	util.WriteJSON(w, response)
}

func handleUpdateCatgory(w http.ResponseWriter, r *http.Request) {
	idStr := general.ParseID(r)
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	category := model.CategoryRequest{}
	err = util.ReadJSON(r, &category)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = UpdateCategory(id, &category)
	if err != nil {
		logger.Error.Printf("error while update category in handleUpdateCategory: reason: %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func handleDeleteCategory(w http.ResponseWriter, r *http.Request) {
	idStr := general.ParseID(r)
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = DeleteCategory(id)
	if err != nil {
		logger.Error.Printf("error while update category in handleDeleteCategory: reason: %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
