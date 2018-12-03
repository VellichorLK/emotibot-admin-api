package util

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"emotibot.com/emotigo/module/admin-api/ApiError"
	"emotibot.com/emotigo/pkg/logger"
	"github.com/hashicorp/consul/api"
)

const (
	//ConsulTEKey is a helper value used in ConsulUpdateTaskEngine
	ConsulTEKey                = "te/enabled"
	ConsulTEMappingTableKey    = "te/mapping_table"
	ConsulTEMappingTableAllKey = "te/mapping_table_all"
	ConsulTEScenarioKey        = "te/scenario"
	ConsulTEScenarioAllKey     = "te/scenario_all"
	ConsulTEAppKey             = "te/app"
	// ConsulRCKey is a helper value used in ConsulUpdateRobotChat
	ConsulRCKey = "chat/%s"
	// ConsulFunctionKey is a helper value used in ConsulUpdateFunctionStatus
	ConsulFunctionKey          = "function/%s"
	ConsulFAQKey               = "faq/%s"
	ConsulEntityKey            = "cnlu/%s"
	ConsulRuleKey              = "rule/%s"
	ConsulCmdKey               = "cmd/%s"
	ConsulProfileKey           = "profile/%s"
	ConsulIntentKey            = "intent/%s"
	ConsulControllerSettingKey = "setting/controller"
	ConsulReleaseInfoKey       = "release_versions"
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

// ConsulGetHandler should handle get kv store in consul.
type ConsulGetHandler func(key string) (string, int, error)

// ConsulGetTreeHandler should handle get kv store recursively in consul.
type ConsulGetTreeHandler func(key string) (map[string]string, int, error)

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

// RootConsulClient is a used for access consul values from root
var RootConsulClient = NewConsulClient(&url.URL{
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
		k, err := url.Parse(key)
		if err != nil {
			logger.Error.Printf("Get error when parse url: %s\n", err.Error())
			return 0, err
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

func newDefaultGetTreeHandler(c *http.Client, u *url.URL) ConsulGetTreeHandler {
	return func(key string) (map[string]string, int, error) {
		key = strings.TrimPrefix(key, "/")
		k, err := url.Parse(key)
		if err != nil {
			logger.Error.Printf("Get error when parse url: %s\n", err.Error())
			return nil, 0, err
		}
		temp := u.ResolveReference(k)
		request, err := http.NewRequest(http.MethodGet, temp.String(), nil)
		if err != nil {
			//TODO: should Logging behavior be done at Controller level or API level?
			logConsulError(err)
			return nil, ApiError.REQUEST_ERROR, err
		}
		q := request.URL.Query()
		q.Add("recurse", "true")
		request.URL.RawQuery = q.Encode()

		logTraceConsul("get", request.URL.String())
		response, err := c.Do(request)
		if err != nil {
			logConsulError(err)
			return nil, ApiError.CONSUL_SERVICE_ERROR, err
		}
		defer response.Body.Close()
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return nil, ApiError.IO_ERROR, err
		}
		if response.StatusCode == http.StatusNotFound {
			return nil, ApiError.SUCCESS, nil
		}

		objs := []map[string]interface{}{}
		err = json.Unmarshal(body, &objs)
		if len(objs) <= 0 {
			return nil, ApiError.JSON_PARSE_ERROR, err
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

				logger.Trace.Printf("Get [%s]: %s\n", key, string(value))
				ret[moduleName] = strValue
			}
		}

		return ret, ApiError.SUCCESS, nil
	}
}

func newDefaultGetHandler(c *http.Client, u *url.URL) ConsulGetHandler {
	return func(key string) (string, int, error) {
		key = strings.TrimPrefix(key, "/")
		k, err := url.Parse(key)
		if err != nil {
			logger.Error.Printf("Get error when parse url: %s\n", err.Error())
			return "", 0, err
		}
		temp := u.ResolveReference(k)
		request, err := http.NewRequest(http.MethodGet, temp.String(), nil)
		if err != nil {
			//TODO: should Logging behavior be done at Controller level or API level?
			logConsulError(err)
			return "", ApiError.REQUEST_ERROR, err
		}
		logTraceConsul("get", temp.String())
		response, err := c.Do(request)
		if err != nil {
			logConsulError(err)
			return "", ApiError.CONSUL_SERVICE_ERROR, err
		}
		defer response.Body.Close()
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return "", ApiError.IO_ERROR, err
		}
		if response.StatusCode == http.StatusNotFound {
			return "", ApiError.SUCCESS, nil
		}

		obj := []map[string]interface{}{}
		err = json.Unmarshal(body, &obj)
		if len(obj) <= 0 {
			return "", ApiError.JSON_PARSE_ERROR, err
		}
		if b64Val, ok := obj[0]["Value"]; ok {
			value, err := base64.StdEncoding.DecodeString(b64Val.(string))
			if err != nil {
				return "", ApiError.BASE64_PARSE_ERROR, err
			}
			return string(value), ApiError.SUCCESS, nil
		}

		return string(body), ApiError.SUCCESS, nil
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

// ConsulGetVal get Consul KV Store by the given key, return value in string format
func (c ConsulClient) ConsulGetVal(key string) (string, int, error) {
	return c.getHandler(key)
}

// ConsulGetVal get Consul KV Store by the given key, return value in string format
func (c ConsulClient) ConsulGetTreeVal(key string) (map[string]string, int, error) {
	return c.getTreeHandler(key)
}

//ConsulUpdateEntity is a convenient function for updating wordbank's Consul Key
func ConsulUpdateEntity(appid string, value interface{}) (int, error) {
	key := fmt.Sprintf(ConsulEntityKey, appid)
	return ConsulUpdateVal(key, value)
}

//ConsulUpdateFAQ is a convenient function for updating Task Engine's Consul Key
func ConsulUpdateFAQ(appid string) (int, error) {
	now := time.Now().Unix()
	key := fmt.Sprintf(ConsulFAQKey, appid)
	return ConsulUpdateVal(key, now)
}

//ConsulUpdateTaskEngine update Task Engine enabling Consul Key
//to enable/disable Task Engine
func ConsulUpdateTaskEngine(appid string, val bool) (int, error) {
	//contains no appid, becaues this can only be use in vipshop for now
	return ConsulUpdateVal(ConsulTEKey, val)
}

//ConsulUpdateTaskEngineMappingTable update the mapping table consul key
//to inform TaskEngine to reload new tables
func ConsulUpdateTaskEngineMappingTable() (int, error) {
	t := time.Now()
	val := t.Format("2006-01-02 15:04:05")
	return ConsulUpdateVal(ConsulTEMappingTableKey, val)
}

//ConsulUpdateTaskEngineMappingTableAll update the mapping table all consul key
//to inform TaskEngine to reload all tables
func ConsulUpdateTaskEngineMappingTableAll() (int, error) {
	t := time.Now()
	val := t.Format("2006-01-02 15:04:05")
	return ConsulUpdateVal(ConsulTEMappingTableAllKey, val)
}

//ConsulUpdateTaskEngineScenario update the scenario consul key to inform TE to reload scenario
func ConsulUpdateTaskEngineScenario() (int, error) {
	t := time.Now()
	val := t.Format("2006-01-02 15:04:05")
	return ConsulUpdateVal(ConsulTEScenarioKey, val)
}

//ConsulUpdateTaskEngineScenarioAll update the scenario all consul key to inform TE to reload all scenario
func ConsulUpdateTaskEngineScenarioAll() (int, error) {
	t := time.Now()
	val := t.Format("2006-01-02 15:04:05")
	return ConsulUpdateVal(ConsulTEScenarioAllKey, val)
}

//ConsulUpdateTaskEngineApp update the app-scenario pair to consul for TE to reload
func ConsulUpdateTaskEngineApp(appid, val string) (int, error) {
	key := fmt.Sprintf("%s/%s", ConsulTEAppKey, appid)
	return ConsulUpdateVal(key, val)
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

//ConsulUpdateRule is a convenient function for updating Robot Chat's Consul Key
func ConsulUpdateRule(appid string) (int, error) {
	key := fmt.Sprintf(ConsulRuleKey, appid)
	now := time.Now().Unix()
	return ConsulUpdateVal(key, now)
}

//ConsulUpdateCmd is a convenient function for updating Robot Chat's Consul Key
func ConsulUpdateCmd(appid string) (int, error) {
	key := fmt.Sprintf(ConsulCmdKey, appid)
	now := time.Now().Unix()
	return ConsulUpdateVal(key, now)
}

//ConsulUpdateProfile is a convenient function for updating Robot Chat's Consul Key
func ConsulUpdateProfile(appid string) (int, error) {
	key := fmt.Sprintf(ConsulProfileKey, appid)
	now := time.Now().Unix()
	return ConsulUpdateVal(key, now)
}

//ConsulUpdateIntent is a convenient function for updating Robot Chat's Consul Key
func ConsulUpdateIntent(appid string) (int, error) {
	key := fmt.Sprintf(ConsulIntentKey, appid)
	now := time.Now().Unix()
	return ConsulUpdateVal(key, now)
}

func ConsulGetControllerSetting() (string, int, error) {
	key := ConsulControllerSettingKey
	return ConsulGetVal(key)
}
func ConsulSetControllerSetting(val string) (int, error) {
	key := ConsulControllerSettingKey
	return ConsulUpdateVal(key, val)
}

func ConsulGetReleaseSetting() (map[string]string, int, error) {
	key := ConsulReleaseInfoKey
	return ConsulGetTreeFromRoot(key)
}

// ConsulUpdateVal is a convenient function for updating Consul KV Store.
// ConsulUpdateVal update Consul KV Store by the given key, value pair.
// value will be formatted by json.Marshal(val), and send to consul's web api by PUT Method.
// It is a wrapper around DefaultConsulClient.ConsulUpdateVal(key, val).
func ConsulUpdateVal(key string, val interface{}) (int, error) {
	return DefaultConsulClient.ConsulUpdateVal(key, val)
}

func ConsulGetVal(key string) (string, int, error) {
	return DefaultConsulClient.ConsulGetVal(key)
}

func ConsulGetTreeFromRoot(key string) (map[string]string, int, error) {
	return RootConsulClient.ConsulGetTreeVal(key)
}

func logTraceConsul(function string, msg string) {
	logger.Trace.Printf("[CONSUL][%s]:%s", function, msg)
}

func logConsulError(err error) {
	logTraceConsul("connect error", err.Error())
}
