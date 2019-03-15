package setting

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gorilla/mux"

	"emotibot.com/emotigo/module/admin-api/util/requestheader"

	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/admin-api/util/AdminErrors"

	"emotibot.com/emotigo/module/qic-api/model/v1"
	"emotibot.com/emotigo/module/qic-api/util/general"
)

// CustomCol is the json response for the custom column
type CustomCol struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	InputName   string `json:"inputname"`
	Type        int8   `json:"type"`
	Description string `json:"description"`
}

func newUserKeyQueryFromQueryString(r *http.Request) (model.UserKeyQuery, error) {
	values, err := url.ParseQuery(r.URL.RawQuery)
	query := model.UserKeyQuery{
		Paging: &model.Pagination{
			Limit: 0,
			Page:  1,
		},
	}
	limit := values.Get("limit")
	if l, err := strconv.Atoi(limit); limit == "" {
		query.Paging.Limit = 10
	} else if err != nil {
		return model.UserKeyQuery{}, fmt.Errorf("limit '%s' is not number", limit)
	} else {
		query.Paging.Limit = l
	}
	page := values.Get("page")
	if query.Paging.Page, err = strconv.Atoi(page); page != "" && err != nil {
		return model.UserKeyQuery{}, fmt.Errorf("parameters page '%s' is not number", page)
	}
	if name := values.Get("name"); name != "" {
		query.FuzzyName = name
	}
	query.Enterprise = requestheader.GetEnterpriseID(r)
	if query.Enterprise == "" {
		return model.UserKeyQuery{}, fmt.Errorf("enterprise ID is required")
	}
	return query, nil
}

// GetCustomColsHandler is the handler to get custom col by the query string.
func GetCustomColsHandler(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Paging general.Paging `json:"paging"`
		Data   []CustomCol    `json:"data"`
	}
	query, err := newUserKeyQueryFromQueryString(r)
	if err != nil {
		util.ReturnError(w, AdminErrors.ErrnoRequestError, fmt.Sprintf("%s", err.Error()))
		return
	}
	data, page, err := GetCustomCols(query)
	if err != nil {
		util.ReturnError(w, AdminErrors.ErrnoDBError, fmt.Sprintf("Get Custom columns failed, %v", err))
		return
	}
	resp := response{
		Paging: page,
		Data:   data,
	}
	util.WriteJSON(w, resp)
}

func CreateCustomColHandler(w http.ResponseWriter, r *http.Request) {
	type requestBody struct {
		Name  string `json:"name"`
		Type  int8   `json:"type"`
		Input string `json:"inputname"`
	}

	var request requestBody
	err := util.ReadJSON(r, &request)
	if err != nil {
		util.ReturnError(w, AdminErrors.ErrnoRequestError, fmt.Sprintf("bad json payload, %v", err))
		return
	}
	enterpriseID := requestheader.GetEnterpriseID(r)
	if enterpriseID == "" {
		util.ReturnError(w, AdminErrors.ErrnoRequestError, fmt.Sprintf("require header of enterprise ID"))
		return
	}
	customCols, err := NewCustomCols([]NewUKRequest{NewUKRequest{
		Enterprise: enterpriseID,
		Name:       request.Name,
		Type:       request.Type,
		InputName:  request.Input,
	}})
	if err != nil {
		util.ReturnError(w, AdminErrors.ErrnoDBError, fmt.Sprintf("new custom columns failed, %v", err))
		return
	}
	cusCol := customCols[0]
	util.WriteJSON(w, CustomCol{
		ID:          cusCol.ID,
		Name:        cusCol.Name,
		InputName:   cusCol.InputName,
		Type:        cusCol.Type,
		Description: "", // Update user key to have description
	})
}

func DeleteCustomColHandler(w http.ResponseWriter, r *http.Request) {
	enterprise := requestheader.GetEnterpriseID(r)
	if enterprise == "" {
		util.ReturnError(w, AdminErrors.ErrnoRequestError, fmt.Sprintf("require enterprise id"))
		return
	}
	inputname := mux.Vars(r)["col_inputname"]
	if inputname == "" {
		util.ReturnError(w, AdminErrors.ErrnoRequestError, fmt.Sprintf("require path parameter"))
		return
	}
	keys, err := userKeys(nil, model.UserKeyQuery{
		InputNames: []string{inputname},
		Enterprise: enterprise,
	})
	if err != nil {
		util.ReturnError(w, AdminErrors.ErrnoDBError, fmt.Sprintf("query userKey failed, %v", err))
		return
	}
	if len(keys) == 0 {
		util.ReturnError(w, AdminErrors.ErrnoRequestError, fmt.Sprintf("can not find user key by the inputname '%s'", inputname))
		return
	}
	key := keys[0]
	_, err = DeleteCustomCols(enterprise, key.InputName)
	if err != nil {
		util.ReturnError(w, AdminErrors.ErrnoDBError, fmt.Sprintf("delete custom col failed, %v", err))
		return
	}
	util.WriteJSON(w, key)

}
