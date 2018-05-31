package Stats

import (
	"fmt"
	"net/http"
	"time"
)

//asiaTaipe is used for Location identify in parsing time, *If runtime environment have no datetime info, the getInputTime may crashed*
var asiaTaipei, _ = time.LoadLocation("Asia/Taipei")

// getInputTime is a util function for retreiving start_time & end_time from request.
// It will return error if request input is invalid or missing.
func getInputTime(r *http.Request) (start, end time.Time, err error) {
	qs := r.URL.Query()
	start, err = time.ParseInLocation("20060102", qs.Get("start_time"), asiaTaipei)
	if err != nil {
		return start, end, fmt.Errorf("start_time is invalid, %v", err)
	}
	end, err = time.ParseInLocation("20060102", qs.Get("end_time"), asiaTaipei)
	if err != nil {
		return start, end, fmt.Errorf("end_time is invalid, %v", err)
	}
	//endtime default should be XXX 23:59:59
	end = end.AddDate(0, 0, 1).Add(-time.Second * 1)
	if end.Before(start) {
		return start, end, fmt.Errorf("start should alway ahead of end")
	}
	return
}

// getType is a util function for get query string key "type" from request.
// If no value is found, error is returned.
func getType(r *http.Request) (typ int, err error) {
	var found bool
	typStr := r.URL.Query().Get("type")
	if typStr == "" {
		return 0, fmt.Errorf("can not found type in request's query string")
	}
	typ, found = typDict[typStr]
	if !found {
		return 0, fmt.Errorf("invalid type in request's query string")
	}
	return
}
