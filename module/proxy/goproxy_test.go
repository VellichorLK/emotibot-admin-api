package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"sync"
	"sync/atomic"
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

var localURL *url.URL

func TestMain(m *testing.M) {
	AddTrafficChan = make(chan string)
	ReadDestChan = make(chan *trafficStats.RouteMap)
	AppidChan = make(chan *trafficStats.AppidIP, 1024)
	trafficStats.DefaultRoutingURL = "http://127.0.0.1:9001"
	localURL, _ = url.Parse("http://127.0.0.1:9001")
	retCode := m.Run()
	os.Exit(retCode)
}

func TestReadList(t *testing.T) {
	t.Parallel()
	testdata := strings.NewReader("#井字號開頭應該被程式省略\n73d2d21af6d8146692069f88b4406b88")
	list, err := ReadList(testdata)
	if err != nil {
		t.Fatal(err)
	}
	if !list["73d2d21af6d8146692069f88b4406b88"] {
		t.Errorf("list should contains 73d2d21af6d8146692069f88b4406b88, but got %+v", list)
	}
	if len(list) != 1 {
		t.Errorf("list should have 1 item, but got %d", len(list))
	}
}

func TestGoProxy(t *testing.T) {
	trafficManager = trafficStats.NewTrafficManager(1000, 100, int64(1000), *localURL)
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
			w.WriteHeader(500)
			t.Fail()
			fmt.Printf("response should have origin user header, but got [%v]\n", uIDHeader)
			return
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
	defer s.Close()
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

}

func BenchmarkGoProxy(b *testing.B) {
	trafficManager = trafficStats.NewTrafficManager(60, int64(10), int64(1000), *localURL)
	for i := 0; i <= b.N; i++ {
		id := string(i % 1000000)
		trafficManager.CheckOverFlowed(id)
	}
}

//TestGoProxyForcedByPass 測試GoProxy函式是否正確分流
func TestGoProxyForcedByPass(t *testing.T) {
	maxConnection := 10
	trafficManager = trafficStats.NewTrafficManager(60, int64(10), int64(1000), *localURL)
	var overFlowedCounter int32
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		defer w.Write([]byte{})

		uIDHeader := r.Header.Get("X-Lb-Uid")
		if uIDHeader != "123" {
			atomic.AddInt32(&overFlowedCounter, 1)
		}
		w.WriteHeader(200)
		w.Write([]byte{})
	})

	go http.ListenAndServe(":9001", nil)
	time.Sleep(1 * time.Second)
	var wg sync.WaitGroup
	for i := 0; i < maxConnection*2; i++ {
		wg.Add(1)
		go func() {
			rr := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/", nil)
			values := url.Values{}
			values.Set("userid", "123")
			r.URL.RawQuery = values.Encode()
			GoProxy(rr, r)
			wg.Done()
			if r := rr.Result(); r.StatusCode != 200 {
				t.Fatalf("response should be OK, but got %v", r.Status)
			}

		}()

	}
	wg.Wait()
	// 使用了兩倍的maxConnection request, 去掉保護前的maxConnection 應該還要有maxConnection 個被強制分流
	if c := atomic.LoadInt32(&overFlowedCounter); c != int32(maxConnection) {
		t.Fatalf("should have %d of requests ByPassed, but got %d of requests", maxConnection, c)
	}

}
