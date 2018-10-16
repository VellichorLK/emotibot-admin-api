package controllers

import (
	"encoding/json"
	"net/http"

	"emotibot.com/emotigo/module/admin-api/ELKStats/data"
)

func returnOK(w http.ResponseWriter, resp interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	writeResponseJSON(w, resp)
}

func returnBadRequest(w http.ResponseWriter, errResp data.ErrorResponseWithCode) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	writeResponseJSON(w, errResp)
}

func returnUnprocessableEntity(w http.ResponseWriter, errResp data.ErrorResponseWithCode) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnprocessableEntity)
	writeResponseJSON(w, errResp)
}

func returnInternalServerError(w http.ResponseWriter, errResp data.ErrorResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	writeResponseJSON(w, errResp)
}

func writeResponseJSON(w http.ResponseWriter, response interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&response)
}
