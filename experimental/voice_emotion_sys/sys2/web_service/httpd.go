package main

import (
	"handlers"
	"log"
	"net/http"
	"os"
	"strconv"
)

var envs = make(map[string]string)
var variableLists = [...]string{"RABBITMQ_HOST", "RABBITMQ_PORT", "DB_HOST", "DB_PORT", "DB_USER", "DB_PWD", "FILE_PREFIX", "LISTEN_PORT", "RABBITMQ_USER", "RABBITMQ_PWD"}

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

	err := handlers.InitDatabaseCon(envs["DB_HOST"], envs["DB_PORT"], envs["DB_USER"], envs["DB_PWD"])
	if err != nil {
		log.Fatal("Can't connect to database!!!")
	}

	handlers.InitEmotionMap()

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
func TaskReceive() ([]byte, string, string, uint8, chan bool) {
	var task []byte
	var fileID, queueName string
	var priority uint8
	var reply chan bool
	select {
	case taskInfo := <-handlers.TaskQueue:
		task = taskInfo.PackagedTask
		fileID = taskInfo.FileInfo.ID
		queueName = taskInfo.QueueN
		priority = taskInfo.FileInfo.Priority
		reply = handlers.RelyQueue
	case cronInfo := <-handlers.CronQueue:
		task = cronInfo.Task
		fileID = cronInfo.FileID
		queueName = cronInfo.QueueName
		reply = cronInfo.Reply
	}

	return task, fileID, queueName, priority, reply
}

func main() {

	parseEnv()
	//fakeEnv()
	readyHandlers()
	flow := createAPI()
	http.ListenAndServe(envs["LISTEN_PORT"], flow)

}
