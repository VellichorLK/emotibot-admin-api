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

func TestNewAppFilterByConfig(t *testing.T) {
	testConfig := filterConfig{
		Default: appConfig{
			QPSLimit:   3,
			DailyLimit: 50,
		},
		Apps: map[string]appConfig{
			"csbot": appConfig{
				QPSLimit:   10,
				DailyLimit: 150,
			},
		},
	}
	mw, counter := newAppFilterByConfig(&testConfig, &sync.Mutex{})
	appList := []string{"ABC", "csbot"}
	for _, appID := range appList {
		i := 0
		for ; i < 100; i++ {
			shouldFilter := mw(appID)
			if appID == "csbot" && shouldFilter {
				t.Fatal("expect csbot do not filter any, but return shouldFilter as true")
			}
			if appID != "csbot" {
				if i < 50 && shouldFilter {
					t.Fatal("expect it does not filter first 50, but return shouldFilter as true")
				}
				if i >= 50 && !shouldFilter {
					t.Fatal("expect it does filter latter 50, but return shouldFilter as false")
				}
			}

		}
		if counter[appID] != int64(i) {
			t.Error("expect appID ", appID, " receive ", i, "count, but got ", counter[appID])
		}
	}

}

func TestNewAppFilterByConfigMultiple(t *testing.T) {
	config := filterConfig{
		Apps: map[string]appConfig{
			"csbot": appConfig{
				DailyLimit: 100,
			},
		},
	}
	lock := sync.Mutex{}
	filter, counter := newAppFilterByConfig(&config, &lock)
	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			filter("csbot")
			wg.Done()
		}()
	}
	wg.Wait()
	if counter["csbot"] != 1000 {
		t.Fatal("expect csbot count to 1000 but got ", counter["csbot"])
	}
}
func TestCreateFilterConfig(t *testing.T) {
	exampleConfig := `#This is a example Config
csbot	100	10
test1	50	5
bcbe21fe4bf94081ba30ca375df81907	20000000	50
bcbe21fe4bf94081ba30ca375df81907	50	5
*	1000	10
`
	data := []byte(exampleConfig)
	config, err := createFilterConfig(data)
	if err != nil {
		t.Fatal(err)
	}
	if config.Default.DailyLimit != 1000 {
		t.Error("expect default daily limit to be 1000, but got", config.Default.DailyLimit)
	}
	if config.Default.QPSLimit != 10 {
		t.Error("expect default qps limit to be 10, but got", config.Default.QPSLimit)
	}
	app, found := config.Apps["csbot"]
	if !found {
		t.Fatal("expect to found app csbot")
	}
	if app.DailyLimit != 100 {
		t.Error("expect csbot daily limit to be 100 but got ", app.DailyLimit)
	}
	if app.QPSLimit != 10 {
		t.Error("expect csbot qps limit to be 10 but got ", app.QPSLimit)
	}
}

func TestNewAppIDLimitMiddleware(t *testing.T) {
	var retryCount, passCount int64
	mw := newAppIDLimitMiddleware(func(w http.ResponseWriter, r *http.Request) {
		retryCount++
	}, func(identifier string) bool {
		if passCount >= 30 {
			return true
		}
		return false
	})
	handler := mw(func(w http.ResponseWriter, r *http.Request) {
		passCount++
	})
	for i := 0; i < 100; i++ {
		w := httptest.NewRecorder()
		var form = url.Values{}
		form.Set("appid", "emotibot")
		form.Set("cmd", "GET")
		form.Set("userid", "taylor")
		r := httptest.NewRequest(http.MethodPost, "/test", nil)
		r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		r.Form = form
		handler.ServeHTTP(w, r)
	}
	if passCount != 30 {
		t.Error("expect 30 request dispatch to inner handler, but got ", passCount)
	}
	if retryCount != 70 {
		t.Error("expect 70 request redirect to retry handler, but got ", retryCount)
	}

}
