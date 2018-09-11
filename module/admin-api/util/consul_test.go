package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"sync"
	"testing"
	"time"

	"emotibot.com/emotigo/pkg/logger"
	"github.com/hashicorp/consul/api"
)

func TestMain(m *testing.M) {
	logger.Init("TEST", bytes.NewBuffer([]byte{}), os.Stdout, os.Stdout, os.Stdout)
	retCode := m.Run()
	os.Exit(retCode)
}

//TestConsulUpdateVal used mocked Http Server to verify the request is valid
//Instead of using a real consul server, it cut the time and dependency
func TestConsulUpdateVal(t *testing.T) {
	type kv struct {
		key string
		val interface{}
	}
	type testObject struct {
		A string `json:"a"`
		B bool   `json:"b"`
	}
	tables := map[string]kv{
		"字串": kv{"test", "hello"},
		//Know Issue: Golang will parse int as float64 in JSON.UnMarshal
		"數值": kv{"test", 1234.0},
		"JSON物件": kv{"test", testObject{
			A: "Hello", B: false,
		}},
	}
	defaultPath := "/v1/kv/idc/"
	for name, tt := range tables {
		t.Run(name, func(t *testing.T) {
			th := func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPut {
					t.Fatalf("Expect HTTP Method be PUT, but got %v", r.Method)
					return
				}
				if uri := r.URL.RequestURI(); uri != defaultPath+tt.key {
					t.Fatalf("Expect URI to be /%s, but got %s", tt.key, uri)
				}
				fmt.Println(r.URL.RequestURI())
				data, err := ioutil.ReadAll(r.Body)
				defer r.Body.Close()
				if err != nil {
					t.Fatal(err)
					return
				}
				var jsonBody interface{}
				//This is needed because golang json type parsing problem.
				//struct became map lead to incomparable situcation
				switch tt.val.(type) {
				case testObject:
					var jsonBody testObject
					err = json.Unmarshal(data, &jsonBody)
					if err != nil {
						t.Fatal(err)
						return
					}
					if !reflect.DeepEqual(jsonBody, tt.val) {
						t.Fatalf("Expect test val be %T of %v, but got %T of %+v", tt.val, tt.val, jsonBody, jsonBody)
					}
				default:
					err = json.Unmarshal(data, &jsonBody)
					if err != nil {
						t.Fatal(err)
						return
					}
					if jsonBody != tt.val {
						t.Fatalf("Expect test val be %T of %v, but got %T of %+v", tt.val, tt.val, jsonBody, jsonBody)
					}
				}

			}
			ts := httptest.NewServer(http.HandlerFunc(th))
			defer ts.Close()
			u, err := url.Parse(ts.URL)
			if err != nil {
				t.Fatal(err)
			}
			u.Path = defaultPath
			DefaultConsulClient.Address = u
			ConsulUpdateVal(tt.key, tt.val)
		})
	}

}

func TestConsulGetVal(t *testing.T) {
	// FIXME: add test for consul get
}

func TestConsulUpdateTaskEngine(t *testing.T) {
	th := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/"+ConsulTEKey {
			t.Fatalf("expect URI to be /%v, but got %v", ConsulTEKey, r.URL.Path)
		}
	}

	ts := httptest.NewServer(http.HandlerFunc(th))
	defer ts.Close()
	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	DefaultConsulClient.Address = u
	ConsulUpdateTaskEngine("", true)
}

func TestConsulUpdateRobotChat(t *testing.T) {
	appid := "vipshop"
	th := func(w http.ResponseWriter, r *http.Request) {
		expectedURI := fmt.Sprintf("/"+ConsulRCKey, appid)
		if r.URL.Path != expectedURI {
			t.Fatalf("expect URI to be /%v, but got %v", expectedURI, r.URL.Path)
		}
	}

	ts := httptest.NewServer(http.HandlerFunc(th))
	defer ts.Close()
	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	DefaultConsulClient.Address = u
	ConsulUpdateRobotChat(appid)
}

type mockedLocker struct {
	c     chan struct{}
	mutex sync.Mutex
}

// Unlock released the lock. It is an error to call this
// if the lock is not currently held.

func (m *mockedLocker) Lock(stopCh <-chan struct{}) (<-chan struct{}, error) {
	m.mutex.Lock()
	if m.c == nil {
		m.c = make(chan struct{})
	} else {
		return nil, api.ErrLockHeld
	}
	m.mutex.Unlock()

	return m.c, nil
}
func (m *mockedLocker) Unlock() error {
	close(m.c)
	m.c = nil
	return nil
}

//TODO: create a fake LockHandler
//Simulate multiple client trying to acquire the lock.
func TestLock(t *testing.T) {
	if _, err := http.Get(DefaultConsulClient.Address.String()); err != nil {
		t.Skipf("Test need DefaultConsulClient reachable, but got %v", err)
	}
	// var m mockedLocker
	// DefaultConsulClient.SetLockHandler(func(key string) (Locker, error) {
	// 	return &m, nil
	// })
	//lock creation is not thread-safe for now.
	lock, err := DefaultConsulClient.Lock("test")
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			signal := make(chan struct{})
			stop, err := lock.Lock(signal)
			if err == api.ErrLockHeld {
				return
			} else if err != nil {
				t.Fatal(err)
			}
			defer lock.Unlock()
			go func() {
				<-stop
				t.Fatal("stopped")
			}()
			time.Sleep(time.Duration(1) * time.Second)
		}()
	}
	wg.Wait()

}
