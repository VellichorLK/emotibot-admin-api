package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"emotibot.com/emotigo/module/vipshop-admin/ApiError"
)

const (
	//ConsulTEKey is a helper value used in ConsulUpdateTaskEngine
	ConsulTEKey = "te/enabled"
	//ConsulRCKey is a helper value used in ConsulUpdateRobotChat
	ConsulRCKey = "%sdata/%s"
)

// ConsulAPI define the method should be implemented by ConsulClient.
// By defining the ConsulAPI interface, we can mock the ConsulClient for Unit Test.
// TODO: Put Get Method into it.
type ConsulAPI interface {
	ConsulUpdateVal(key string, val interface{}) (int, error)
}

// ConsulClient is a homemade client used for communicate with Consul API.
type ConsulClient struct {
	Address string
	client  *http.Client
}

// DefaultConsulClient is a used for convenient function packed in package.
var DefaultConsulClient = &ConsulClient{
	Address: "127.0.0.1",
	client:  http.DefaultClient,
}

// NewConsulClient create a consul client with http.DefaultClient in http package.
// Be care with the DefaultClient's Timeout value.
func NewConsulClient(address string) *ConsulClient {
	return NewConsulClientWithCustomHTTP(address, http.DefaultClient)
}

// NewConsulClientWithCustomHTTP create a client with given http.Client.
func NewConsulClientWithCustomHTTP(address string, client *http.Client) *ConsulClient {
	address = strings.TrimLeft(address, "/")
	return &ConsulClient{
		Address: address,
		client:  client,
	}
}

// ConsulUpdateVal update Consul KV Store by the given key, value pair.
// value will be formatted by json.Marshal(val), and send to consul's web api by PUT Method.
func (c *ConsulClient) ConsulUpdateVal(key string, val interface{}) (int, error) {
	reqURL := fmt.Sprintf("%s/%s", c.Address, key)
	body, err := json.Marshal(val)
	request, err := http.NewRequest(http.MethodPut, reqURL, bytes.NewReader(body))
	if err != nil {
		//TODO: should Logging behavior be done at Controller level or API level?
		logConsulError(err)
		return ApiError.REQUEST_ERROR, err
	}

	logTraceConsul("update", reqURL)

	_, err = c.client.Do(request)
	if err != nil {
		logConsulError(err)
		return ApiError.CONSUL_SERVICE_ERROR, err
	}
	return ApiError.SUCCESS, nil
}

//ConsulUpdateTaskEngine is a convenient function for updating Task Engine's Consul Key
func ConsulUpdateTaskEngine(appid string, val bool) (int, error) {
	//contains no appid, becaues this can be use in vipshop for now
	return ConsulUpdateVal(ConsulTEKey, val)
}

//ConsulUpdateRobotChat is a convenient function for updating Robot Chat's Consul Key
func ConsulUpdateRobotChat(appid string) (int, error) {
	key := fmt.Sprintf(ConsulRCKey, appid, appid)
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
