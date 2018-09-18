package util

import (
	"emotibot.com/emotigo/module/openapi-adapter/statsd"
)

var client *statsd.Client

// GetClient returns client of statsd host and creates one if not existed
func GetStatsdClient(host string, port int) *statsd.Client {
	if client == nil {
		client = statsd.New(host, port)
	}
	return client
}
