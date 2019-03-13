package redis

import (
	"github.com/mediocregopher/radix"
	"strings"
	"time"
)

func NewCluster(address []string, password string) (cluster radix.Client, err error) {
	// connection factory
	connFunc := func(network, addr string) (radix.Conn, error) {
		return radix.Dial(network, addr,
			radix.DialTimeout(1*time.Minute),
			radix.DialAuthPass(password),
		)
	}

	// connection pool factroy which use connection factory to create connection
	poolFunc := func(network, addr string) (radix.Client, error) {
		return radix.NewPool(network, addr, 5, radix.PoolConnFunc(connFunc))
	}

	// create redis cluster client
	cluster, err = radix.NewCluster(address, radix.ClusterPoolFunc(poolFunc))
	return
}

func NewClusterFromEnvs(envs map[string]string) (radix.Client, error) {
	address, password := ConfigFromEnv(envs)

	return NewCluster(address, password)
}

func ConfigFromEnv(envs map[string]string) (address []string, password string) {
	redisAddresses := envs["REDIS_URLS"]
	address = strings.Split(redisAddresses, ",")

	password = envs["REDIS_PASSWORD"]
	return address, password
}
