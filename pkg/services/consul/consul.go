package consul

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/hashicorp/consul/api"
)

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
type ConsulUpdateHandler func(key string, val interface{}) error

// ConsulGetHandler should handle get kv store in consul.
type ConsulGetHandler func(key string) (string, error)

// ConsulGetTreeHandler should handle get kv store recursively in consul.
type ConsulGetTreeHandler func(key string) (map[string]string, error)

// ConsulClient is an adapter used for communicate with Consul API.
type ConsulClient struct {
	lockHandler    ConsulLockHandler
	updateHandler  ConsulUpdateHandler
	getHandler     ConsulGetHandler
	getTreeHandler ConsulGetTreeHandler
	Address        *url.URL //address should be a valid URL string, ex: http://127.0.0.1:8500/
	client         *http.Client
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
	c.getHandler = newDefaultGetHandler(client, address)
	c.getTreeHandler = newDefaultGetTreeHandler(client, address)
	return c
}

//SetLockHandler set the handler function for the func Lock in this ConsulClient.
func (c *ConsulClient) SetLockHandler(handler ConsulLockHandler) {
	c.lockHandler = handler
}

//SetUpdateHandler set the handler for the Update value operation in ConsulClient.
func (c *ConsulClient) SetUpdateHandler(handler ConsulUpdateHandler) {
	c.updateHandler = handler
}

func newDefaultUpdateHandler(c *http.Client, u *url.URL) ConsulUpdateHandler {
	return func(key string, val interface{}) error {
		key = strings.TrimPrefix(key, "/")
		k, err := url.Parse(key)
		if err != nil {
			return fmt.Errorf("Get error when parse url: %v", err)
		}
		temp := u.ResolveReference(k)
		var body []byte
		if str, ok := val.(string); ok {
			body = []byte(str)
		} else {
			body, err = json.Marshal(val)
		}
		request, err := http.NewRequest(http.MethodPut, temp.String(), bytes.NewReader(body))
		if err != nil {
			return err
		}
		_, err = c.Do(request)
		if err != nil {
			return err
		}

		return nil
	}
}

//ErrNotFound is an kind of error
var ErrKeyNotFound = errors.New("consul: key not found")

func newDefaultGetTreeHandler(c *http.Client, u *url.URL) ConsulGetTreeHandler {
	return func(key string) (map[string]string, error) {
		key = strings.TrimPrefix(key, "/")
		k, err := url.Parse(key)
		if err != nil {
			return nil, err
		}
		temp := u.ResolveReference(k)
		request, err := http.NewRequest(http.MethodGet, temp.String(), nil)
		if err != nil {
			return nil, err
		}
		q := request.URL.Query()
		q.Add("recurse", "true")
		request.URL.RawQuery = q.Encode()

		response, err := c.Do(request)
		if err != nil {
			return nil, err
		}
		defer response.Body.Close()
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}
		if response.StatusCode == http.StatusNotFound {
			//TODO
			return nil, ErrKeyNotFound
		}

		objs := []map[string]interface{}{}
		err = json.Unmarshal(body, &objs)
		if len(objs) <= 0 {
			return nil, err
		}

		ret := map[string]string{}
		for idx := range objs {
			if b64Val, ok := objs[idx]["Value"]; ok {
				value, err := base64.StdEncoding.DecodeString(b64Val.(string))
				if err != nil {
					continue
				}
				origKey := objs[idx]["Key"].(string)
				moduleName := strings.TrimPrefix(origKey, key+"/")
				strValue := strings.TrimPrefix(string(value), moduleName+":")

				ret[moduleName] = strValue
			}
		}

		return ret, nil
	}
}

func newDefaultGetHandler(c *http.Client, u *url.URL) ConsulGetHandler {
	return func(key string) (string, error) {
		key = strings.TrimPrefix(key, "/")
		k, err := url.Parse(key)
		if err != nil {
			return "", err
		}
		temp := u.ResolveReference(k)
		request, err := http.NewRequest(http.MethodGet, temp.String(), nil)
		if err != nil {
			return "", err
		}
		response, err := c.Do(request)
		if err != nil {
			return "", err
		}
		defer response.Body.Close()
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return "", err
		}
		if response.StatusCode == http.StatusNotFound {
			return "", nil
		}

		obj := []map[string]interface{}{}
		err = json.Unmarshal(body, &obj)
		if len(obj) <= 0 {
			return "", err
		}
		if b64Val, ok := obj[0]["Value"]; ok {
			value, err := base64.StdEncoding.DecodeString(b64Val.(string))
			if err != nil {
				return "", err
			}
			return string(value), nil
		}

		return string(body), nil
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
func (c *ConsulClient) ConsulUpdateVal(key string, val interface{}) error {
	return c.updateHandler(key, val)
}

// ConsulGetVal get Consul KV Store by the given key, return value in string format
func (c ConsulClient) ConsulGetVal(key string) (string, error) {
	return c.getHandler(key)
}

// ConsulGetVal get Consul KV Store by the given key, return value in string format
func (c ConsulClient) ConsulGetTreeVal(key string) (map[string]string, error) {
	return c.getTreeHandler(key)
}

// ConsulUpdateVal is a convenient function for updating Consul KV Store.
// ConsulUpdateVal update Consul KV Store by the given key, value pair.
// value will be formatted by json.Marshal(val), and send to consul's web api by PUT Method.
// It is a wrapper around DefaultConsulClient.ConsulUpdateVal(key, val).
func ConsulUpdateVal(key string, val interface{}) error {
	return DefaultConsulClient.ConsulUpdateVal(key, val)
}

func ConsulGetVal(key string) (string, error) {
	return DefaultConsulClient.ConsulGetVal(key)
}
