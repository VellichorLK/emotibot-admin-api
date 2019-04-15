package controllers

import (
	"encoding/json"
	"net/http"

	"emotibot.com/emotigo/module/admin-api/QADoc/data"
	"emotibot.com/emotigo/module/admin-api/QADoc/services"
	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/module/admin-api/util/AdminErrors"
)

func CreateQADocHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	request := data.QACoreDoc{}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		util.ReturnError(w, AdminErrors.ErrnoJSONParse, err.Error())
		return
	}

	_, err = services.CreateQADoc(&request)
	if err != nil {
		util.ReturnError(w, AdminErrors.ErrnoIOError, err.Error())
		return
	}

	util.Return(w, nil, "Success")
}

func CreateQADocsHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	request := []*data.QACoreDoc{}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		util.ReturnError(w, AdminErrors.ErrnoJSONParse, err.Error())
		return
	}

	_, err = services.BulkCreateQADocs(request)
	if err != nil {
		util.ReturnError(w, AdminErrors.ErrnoIOError, err.Error())
		return
	}

	util.Return(w, nil, "Success")
}

func DeleteQADocsHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	request := data.DeleteQADocsRequest{}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		util.ReturnError(w, AdminErrors.ErrnoJSONParse, err.Error())
		return
	}

	_, err = services.DeleteQADocs(request.AppID, request.Module)
	if err != nil {
		util.ReturnError(w, AdminErrors.ErrnoIOError, err.Error())
		return
	}

	util.Return(w, nil, "Success")
}

func DeleteQADocsByIDsHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	request := data.DeleteQADocsByIDsRequest{}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		util.ReturnError(w, AdminErrors.ErrnoJSONParse, err.Error())
		return
	}

	_, err = services.DeleteQADocsByIds(request.IDs)
	if err != nil {
		util.ReturnError(w, AdminErrors.ErrnoIOError, err.Error())
		return
	}

	util.Return(w, nil, "Success")
}
