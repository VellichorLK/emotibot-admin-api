package auth

import (
	"os"
)

const (
	const_dfl_db_user     string = "root"
	const_dfl_db_pass     string = "password"
	const_dfl_db_url      string = "172.17.0.1:3306"
	const_dfl_db_name     string = "authentication"
	const_dfl_listen_port string = "8088"
	const_dfl_consul_url  string = "172.17.0.1:8500"
)

type Configuration struct {
	DbUrl      string
	DbUser     string
	DbPass     string
	DbName     string
	ListenPort string
	ConsulUrl  string
}

func GetConfig() *Configuration {
	c := Configuration{}
	c.DbUrl = os.Getenv("MYSQL_URL")
	c.DbUser = os.Getenv("MYSQL_USER")
	c.DbPass = os.Getenv("MYSQL_PASS")
	c.DbName = os.Getenv("MYSQL_DB")
	c.ListenPort = os.Getenv("PORT")
	c.ConsulUrl = os.Getenv("CONSUL_URL")
	if c.DbUrl == "" {
		c.DbUrl = const_dfl_db_url
	}
	if c.DbUser == "" {
		c.DbUser = const_dfl_db_user
	}
	if c.DbPass == "" {
		c.DbPass = const_dfl_db_pass
	}

	if c.DbName == "" {
		c.DbName = const_dfl_db_name
	}

	if c.ListenPort == "" {
		c.ListenPort = const_dfl_listen_port
	}

	if c.ConsulUrl == "" {
		c.ConsulUrl = const_dfl_consul_url
	}
	return &c
}
