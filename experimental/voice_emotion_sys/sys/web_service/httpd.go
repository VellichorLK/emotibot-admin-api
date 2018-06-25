package main

import (
	"handlers"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

var envs = make(map[string]string)
var variableLists = [...]string{"RABBITMQ_HOST", "RABBITMQ_PORT", "DB_HOST", "DB_PORT",
	"DB_USER", "DB_PWD", "FILE_PREFIX", "LISTEN_PORT", "RABBITMQ_USER", "RABBITMQ_PWD",
	"CONSUL_IP", "CONSUL_PORT"}

func parseEnv() {
	for _, v := range variableLists {
		if os.Getenv(v) == "" {
			log.Fatalf("%s is empty!", v)
		}
		envs[v] = os.Getenv(v)
	}
	envs["LISTEN_PORT"] = ":" + envs["LISTEN_PORT"]

}

func readyHandlers() {

	err := handlers.InitConsulClient(envs["CONSUL_IP"]+":"+envs["CONSUL_PORT"], 1*time.Second)
	if err != nil {
		log.Fatalf("Init consul client error:%s\n", err)
	}

	err = handlers.InitDatabaseCon(envs["DB_HOST"], envs["DB_PORT"], envs["DB_USER"], envs["DB_PWD"], "voice_emotion")
	//err := handlers.InitDatabaseCon(envs["DB_HOST"], envs["DB_PORT"], envs["DB_USER"], envs["DB_PWD"], "mydb")
	if err != nil {
		log.Fatal("Can't connect to database!!!")
	}

	handlers.InitEmotionMap()

	err = handlers.LoadUsrField()
	if err != nil {
		log.Fatal(err)
	}

	port, err := strconv.Atoi(envs["RABBITMQ_PORT"])
	if err != nil {
		log.Fatalf("Can't conver  RABBITMQ_PORT(%s) to int!!!", envs["RABBITMQ_PORT"])
	}
	go handlers.StartSendTaskService(envs["RABBITMQ_HOST"], port, envs["RABBITMQ_USER"], envs["RABBITMQ_PWD"], TaskReceive)

	handlers.FilePrefix = envs["FILE_PREFIX"]
}

func createAPI() http.Handler {
	muxer := http.NewServeMux()
	for i := 0; i < len(services); i++ {

		for path, doer := range services[i] {
			h := http.HandlerFunc(doer)
			muxer.Handle(path, h)
		}
	}

	var hs http.Handler

	cbs := len(MiddleServices)

	if cbs > 0 {
		cbs--
		hs = MiddleServices[cbs](muxer)
		for i := 0; i < cbs; i++ {
			hs = MiddleServices[cbs-1-i](hs)
		}

	} else {
		return muxer
	}

	return hs
}

//TaskReceive return finished channel,task as string, fileID, queueName, priority
func TaskReceive() (string, string, string, uint8) {
	taskInfo := <-handlers.TaskQueue
	return taskInfo.PackagedTask, taskInfo.FileInfo.ID, taskInfo.QueueN, taskInfo.FileInfo.Priority
}

func main() {

	parseEnv()
	//fakeEnv()
	readyHandlers()
	flow := createAPI()
	http.ListenAndServe(envs["LISTEN_PORT"], flow)

}
