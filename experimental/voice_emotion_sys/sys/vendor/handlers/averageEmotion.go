package handlers

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"
)

func getFilter(url string) string {
	filter := strings.SplitN(url, "/", MaxSlash)
	return filter[MaxSlash-1]
}

func GroupAverageEmotion(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		appid := r.Header.Get(NUAPPID)
		if appid == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		filter := getFilter(r.URL.Path)

		switch filter {
		case NTAG:
		case NTAG2:
		default:
			http.Error(w, "Bad Request: wrong filter assigned", http.StatusBadRequest)
			return
		}

		params := r.URL.Query()
		_t1 := params.Get(NT1)
		_t2 := params.Get(NT2)

		t1, t2, days, err := ParseTime(_t1, _t2)
		if err != nil {
			http.Error(w, "Bad Request: "+err.Error(), http.StatusBadRequest)
			return
		}

		avgEmotionMap, err := groupAvgEmotion(t1, t2, appid, filter)
		if err != nil {
			http.Error(w, "Internal server error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		gr := genGroupReport(avgEmotionMap)
		gr.Total = days

		sort.Sort(GroupsEmotion(gr.Group))

		resp, err := json.Marshal(gr)
		if err != nil {
			http.Error(w, "Internal server error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		contentType := "application/json; charset=utf-8"

		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(http.StatusOK)
		w.Write(resp)

	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
}

func CallAverageEmotion(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {

		appid := r.Header.Get(NUAPPID)
		if appid == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		params := r.URL.Query()
		_t1 := params.Get(NT1)
		_t2 := params.Get(NT2)

		t1, t2, days, err := ParseTime(_t1, _t2)
		if err != nil {
			http.Error(w, "Bad Request: "+err.Error(), http.StatusBadRequest)
			return
		}

		aes, err := dailyAvgEmotion(t1, t2, appid)
		if err != nil {
			http.Error(w, "Internal server error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		dr := &DailyReport{Total: days, AvgEmotions: aes}

		resp, err := json.Marshal(dr)
		if err != nil {
			http.Error(w, "Internal server error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		contentType := "application/json; charset=utf-8"

		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(http.StatusOK)
		w.Write(resp)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
}
