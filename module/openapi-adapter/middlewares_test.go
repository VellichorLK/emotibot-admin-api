package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
)

func TestDailyLimitMiddleWare(t *testing.T) {
	var count int64
	handler := func(w http.ResponseWriter, r *http.Request) {
		if filtered := r.Header.Get("X-Filtered"); filtered == "true" {
			return
		}
		count++
	}
	var appGroup = map[string]int64{}
	var limit int64 = 10
	dailyLimit := NewDailyLimitMiddleWare(appGroup, limit, &sync.Mutex{})(handler)
	var i int64
	for ; i <= 1000; i++ {
		w := httptest.NewRecorder()
		form := url.Values{}
		form.Set("appid", "emotibot")
		form.Set("cmd", "TEST")
		form.Set("userid", "taylor")
		r := httptest.NewRequest(http.MethodPost, "/test", nil)
		r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		r.Form = form
		dailyLimit(w, r)
	}
	if c, _ := appGroup["emotibot"]; c != i {
		t.Fatal("Expect app count to be ", i, "but got ", c)
	}
	if count != limit {
		t.Fatal("expect inner visit only to limit size", limit, "but got ", count)
	}
}
