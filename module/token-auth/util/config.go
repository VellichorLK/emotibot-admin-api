package util

import (
	"strconv"
	"strings"
)

const (
	mysqlSQLURLKey      = "ADMIN_AUTH_MYSQL_URL"
	mysqlSQLUserKey     = "ADMIN_AUTH_MYSQL_USER"
	mysqlSQLPasswordKey = "ADMIN_AUTH_MYSQL_PASS"
	mysqlSQLDatabaseKey = "ADMIN_AUTH_MYSQL_DB"

	serverPortKey = "ADMIN_AUTH_PORT"
	serverURLKey  = "ADMIN_AUTH_URL"
)

// GetMySQLConfig will get db init config from env
func GetMySQLConfig() (url string, port int, user string, password string, database string) {
	envURL := GetStrEnv(mysqlSQLURLKey, "127.0.0.1:3306")
	params := strings.Split(envURL, ":")
	if len(params) <= 1 {
		url = params[0]
		port = 3306
	} else {
		url = params[0]
		port, _ = strconv.Atoi(params[1])
	}
	user = GetStrEnv(mysqlSQLUserKey, "root")
	password = GetStrEnv(mysqlSQLPasswordKey, "password")
	database = GetStrEnv(mysqlSQLDatabaseKey, "auth")
	return
}

// GetServerConfig will get server binding config from env
func GetServerConfig() (url string, port int) {
	port = GetIntEnv(serverPortKey, 8088)
	url = GetStrEnv(serverURLKey, "0.0.0.0")
	return
}
