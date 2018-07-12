package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

func getFilter(url string) string {
	filter := strings.SplitN(url, "/", MaxSlash+1)
	return filter[MaxSlash]
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
		groupStr := params.Get(GROUPS)

		var groupVals []interface{}
		if groupStr != "" {
			groups := strings.Split(groupStr, ",")
			groupIDs, err := parseGroup(groups)
			if err != nil {
				http.Error(w, "Bad Request: "+err.Error(), http.StatusBadRequest)
				return
			}

			groupSet, err := getGroups(appid, groupIDs)
			if err != nil {
				log.Printf("Error:%s\n", err)
				http.Error(w, "Internal Server error", http.StatusInternalServerError)
				return
			}

			groupVals = parseGroupSet(groupSet)
		}

		t1, t2, days, err := ParseTime(_t1, _t2)
		if err != nil {
			http.Error(w, "Bad Request: "+err.Error(), http.StatusBadRequest)
			return
		}

		aes, err := dailyAvgEmotion(t1, t2, appid, groupVals)
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

func AverageDuration(w http.ResponseWriter, r *http.Request) {
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

		t2 = uint64(AddTimeUnit(int64(t2), Day) - 1)

		avgDurMap, err := groupAvgDuration(t1, t2, appid, filter)
		if err != nil {
			http.Error(w, "Internal server error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		gr := genGroupDurationReport(avgDurMap)
		gr.Total = days

		sort.Sort(GroupsDuration(gr.Group))

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

func parseGroup(groups []string) ([]uint64, error) {
	num := len(groups)
	var groupIDs []uint64
	var err error
	var id uint64
	if num > 0 {
		groupIDs = make([]uint64, num, num)
		for idx, val := range groups {
			id, err = strconv.ParseUint(val, 10, 64)
			if err != nil {
				return nil, err
			}
			groupIDs[idx] = id
		}

	}
	return groupIDs, err

}

func parseGroupSet(groups []*GroupSet) []interface{} {

	vals := make([]interface{}, 0)
	for _, group := range groups {
		for _, val := range group.GroupVal {
			vals = append(vals, val)
		}
	}
	return vals
}
