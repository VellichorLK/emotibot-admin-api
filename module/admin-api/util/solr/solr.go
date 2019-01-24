package solr

import (
	"fmt"
)

var (
	host string
	port string
)

func Setup(envHost string, envPort string) {
	host = envHost
	port = envPort
}

func GetBaseURL() string {
	return fmt.Sprintf("http://%s:%s/solr", host, port)
}
