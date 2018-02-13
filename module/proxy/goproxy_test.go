package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"emotibot.com/emotigo/module/proxy/traffic"
)

var envs = map[string]string{
	"DURATION":    "10",
	"MAXREQUESTS": "10",
	"BANPERIOD":   "10",
	"LOGPERIOD":   "10",
	"STATSDHOST":  "127.0.0.1",
	"STATSDPORT":  "8500",
}

func TestMain(m *testing.M) {
	AddTrafficChan = make(chan string)
	ReadDestChan = make(chan *trafficStats.RouteMap)
	AppidChan = make(chan *trafficStats.AppidIP, 1024)
	trafficStats.DefaultRoutingURL = "http://127.0.0.1:9001"

	retCode := m.Run()
	os.Exit(retCode)
}
func TestGoProxy(t *testing.T) {
	trafficStats.Init(10, 1024, 100, 100, AddTrafficChan, ReadDestChan, AppidChan, "127.0.0.1:8500")
	rr := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/", nil)
	values := url.Values{}
	values.Set("userid", "123")
	r.URL.RawQuery = values.Encode()
	var counter = 0
	mux := http.NewServeMux()
	expectedResult := "Test"
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		counter++
		if uIDHeader := r.Header.Get("X-Lb-Uid"); uIDHeader != "123" {
			fmt.Printf("%+v\n", r)
			w.WriteHeader(500)
			t.Fatalf("response should have origin user header, but got [%v]", uIDHeader)
		}
		w.WriteHeader(200)
		fmt.Fprintf(w, "%s", expectedResult)
	})
	s := &http.Server{
		Addr:         ":9001",
		Handler:      mux,
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 10,
	}
	go s.ListenAndServe()
	GoProxy(rr, r)

	if counter != 1 {
		t.Fail()
	}
	response := rr.Result()
	d, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("response should be OK, but got %v", response.Status)
	}
	if text := string(d); strings.Compare(text, expectedResult) != 0 {
		t.Fatalf("response should be %s, but got %v\n", expectedResult, text)
	}
	rr = httptest.NewRecorder()
	GoProxy(rr, r)
	response = rr.Result()
}

//TestGoProxyForcedByPass 測試GoProxy函式是否正確分流
func TestGoProxyForcedByPass(t *testing.T) {
	maxConnection := 10
	trafficStats.Init(60, maxConnection, 100000, 100000, AddTrafficChan, ReadDestChan, AppidChan, "127.0.0.1:8500")
	//counter 用1開始的 代表第一個request 就已經算上connection
	counter := 1
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		defer w.Write([]byte{})
		uIDHeader := r.Header.Get("X-Lb-Uid")
		if counter <= maxConnection && uIDHeader != "123" {
			w.WriteHeader(500)
			t.Fatalf("response should have origin user header, but got [%v]", uIDHeader)
		}
		if counter > maxConnection && uIDHeader == "123" {
			w.WriteHeader(500)
			fmt.Printf("Number of %d response should active bypassing now, but still got Header as 123\n", counter)
			return
		}
		w.WriteHeader(200)
		w.Write([]byte{})
		counter++
	})
	go http.ListenAndServe(":9001", nil)
	for i := 0; i <= maxConnection*2; i++ {
		rr := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/", nil)
		values := url.Values{}
		values.Set("userid", "123")
		r.URL.RawQuery = values.Encode()
		GoProxy(rr, r)
		if r := rr.Result(); r.StatusCode != 200 {
			t.Fatalf("response should be OK, but got %v", r.Status)
		}
	}

}
