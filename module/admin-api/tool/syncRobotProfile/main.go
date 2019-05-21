package main

import (
	"os"

	"emotibot.com/emotigo/module/admin-api/Robot"
	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/pkg/logger"
)

func main() {
	logger.Init("SYNC", os.Stdout, os.Stdout, os.Stdout, os.Stdout)
	if len(os.Args) > 1 {
		err := util.LoadConfigFromFile(os.Args[1])
		if err != nil {
			logger.Error.Printf(err.Error())
			os.Exit(-1)
		}
	}
	initDB()
	Robot.SyncOnce()
}

func initDB() {
	url := getServerEnv("MYSQL_URL")
	user := getServerEnv("MYSQL_USER")
	pass := getServerEnv("MYSQL_PASS")
	db := getServerEnv("MYSQL_DB")
	err := util.InitMainDB(url, user, pass, db)
	if err != nil {
		logger.Trace.Printf("Init DB Error: %s\n", err.Error())
	}
}

func getServerEnv(key string) string {
	envs := util.GetEnvOf("server")
	if envs != nil {
		if val, ok := envs[key]; ok {
			return val
		}
	}
	return ""
}
