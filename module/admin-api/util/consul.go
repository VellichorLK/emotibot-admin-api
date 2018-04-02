package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"emotibot.com/emotigo/module/admin-api/ApiError"
	"github.com/hashicorp/consul/api"
)

const (
	//ConsulTEKey is a helper value used in ConsulUpdateTaskEngine
	ConsulTEKey = "te/enabled"
	// ConsulRCKey is a helper value used in ConsulUpdateRobotChat
	// vipshopdata will be modified, change it back for demo day
	ConsulRCKey = "chat/%s"
	// ConsulRCKey is a helper value used in ConsulUpdateFunctionStatus
	ConsulFunctionKey = "function/%s"
	ConsulFAQKey      = "faq/%s"
)

// ConsulAPI define the method should be implemented by ConsulClient.
// By defining the ConsulAPI interface, we can mock the ConsulClient for Unit Test.
// TODO: Put Get Method into it.
type ConsulAPI interface {
	update(key string, val interface{}) (int, error)
}

type Locker interface {
	Lock(stopCh <-chan struct{}) (<-chan struct{}, error)
	// Unlock released the lock. It is an error to call this
	// if the lock is not currently held.
	Unlock() error
}

// ConsulLockHandler should returns a handle to a lock struct which can be used
// to acquire and release the mutex. The key used must have
// write permissions.
// It use the definition of consul/api.LockOpts,
// so it should return api.ErrLockHeld if it can't acquiring lock.
type ConsulLockHandler func(key string) (Locker, error)

// ConsulUpdateHandler should handle update kv store in consul.
// val should be json encoded and return int as ApiError defined for backword compability.
type ConsulUpdateHandler func(key string, val interface{}) (int, error)

// ConsulClient is an adapter used for communicate with Consul API.
type ConsulClient struct {
	lockHandler   ConsulLockHandler
	updateHandler ConsulUpdateHandler
	Address       *url.URL //address should be a valid URL string, ex: http://127.0.0.1:8500/
	client        *http.Client
}

// DefaultConsulClient is a used for convenient function packed in package.
var DefaultConsulClient = NewConsulClient(&url.URL{
	Host:   "127.0.0.1:8500",
	Scheme: "http",
})

// NewConsulClient create a consul client with http.DefaultClient in http package.
// Be care with the DefaultClient's Timeout value.
func NewConsulClient(address *url.URL) *ConsulClient {
	return NewConsulClientWithCustomHTTP(address, http.DefaultClient)
}

// NewConsulClientWithCustomHTTP create a client with given http.Client.
func NewConsulClientWithCustomHTTP(address *url.URL, client *http.Client) *ConsulClient {
	c := &ConsulClient{
		Address: address,
		client:  client,
	}
	c.updateHandler = newDefaultUpdateHandler(client, address)
	c.lockHandler = newDefaultLockHandler(client, address)
	return c
}

//SetLockHandler set the handler function for the func Lock in this ConsulClient.
func (c *ConsulClient) SetLockHandler(handler func(key string) (Locker, error)) {
	c.lockHandler = handler
}

//SetUpdateHandler set the handler for the Update value operation in ConsulClient.
func (c *ConsulClient) SetUpdateHandler(handler func(key string, val interface{}) (int, error)) {
	c.updateHandler = handler
}

func newDefaultUpdateHandler(c *http.Client, u *url.URL) ConsulUpdateHandler {
	return func(key string, val interface{}) (int, error) {
		key = strings.TrimPrefix(key, "/")
		k, _ := url.Parse(key)
		temp := u.ResolveReference(k)
		body, err := json.Marshal(val)
		request, err := http.NewRequest(http.MethodPut, temp.String(), bytes.NewReader(body))
		if err != nil {
			//TODO: should Logging behavior be done at Controller level or API level?
			logConsulError(err)
			return ApiError.REQUEST_ERROR, err
		}
		logTraceConsul("update", temp.String())
		_, err = c.Do(request)
		if err != nil {
			logConsulError(err)
			return ApiError.CONSUL_SERVICE_ERROR, err
		}

		return ApiError.SUCCESS, nil
	}
}

//newDefaultLockHandler generate a default behavior for acquiring lock. It basically use consul/api client for lock
func newDefaultLockHandler(client *http.Client, addr *url.URL) ConsulLockHandler {
	a, err := api.NewClient(&api.Config{
		Address:    addr.Host,
		Scheme:     addr.Scheme,
		HttpClient: client,
	})
	if err != nil {
		//Return a no-op handler
		return func(key string) (Locker, error) {
			return nil, err
		}
	}

	return func(key string) (Locker, error) {
		opt := &api.LockOptions{
			Key: key,
		}
		return a.LockOpts(opt)
	}
}

// Lock will call the client's lockHandler which handle the implemented work.
func (c *ConsulClient) Lock(key string) (Locker, error) {
	return c.lockHandler(key)
}

// ConsulUpdateVal update Consul KV Store by the given key, value pair.
// value will be formatted by json.Marshal(val), and send to consul's web api by PUT Method.
func (c *ConsulClient) ConsulUpdateVal(key string, val interface{}) (int, error) {
	return c.updateHandler(key, val)
}

//ConsulUpdateFAQ is a convenient function for updating Task Engine's Consul Key
func ConsulUpdateFAQ(appid string) (int, error) {
	//contains no appid, becaues this can be use in vipshop for now
	now := time.Now().Unix()
	return ConsulUpdateVal(ConsulFAQKey, now)
}

//ConsulUpdateTaskEngine is a convenient function for updating Task Engine's Consul Key
func ConsulUpdateTaskEngine(appid string, val bool) (int, error) {
	//contains no appid, becaues this can be use in vipshop for now
	return ConsulUpdateVal(ConsulTEKey, val)
}

//ConsulUpdateFunctionStatus is a convenient function for updating Robot Chat's Consul Key
func ConsulUpdateFunctionStatus(appid string) (int, error) {
	key := fmt.Sprintf(ConsulFunctionKey, appid)
	now := time.Now().Unix()
	return ConsulUpdateVal(key, now)
}

//ConsulUpdateRobotChat is a convenient function for updating Robot Chat's Consul Key
func ConsulUpdateRobotChat(appid string) (int, error) {
	key := fmt.Sprintf(ConsulRCKey, appid)
	now := time.Now().Unix()
	return ConsulUpdateVal(key, now)
}

// ConsulUpdateVal is a convenient function for updating Consul KV Store.
// ConsulUpdateVal update Consul KV Store by the given key, value pair.
// value will be formatted by json.Marshal(val), and send to consul's web api by PUT Method.
// It is a wrapper around DefaultConsulClient.ConsulUpdateVal(key, val).
func ConsulUpdateVal(key string, val interface{}) (int, error) {
	return DefaultConsulClient.ConsulUpdateVal(key, val)
}

func logTraceConsul(function string, msg string) {
	LogTrace.Printf("[CONSUL][%s]:%s", function, msg)
}

func logConsulError(err error) {
	logTraceConsul("connect error", err.Error())
}
