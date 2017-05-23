package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"

	"emotibot.com/emotigo/experimental/scalable_service/workers/golang_worker/RabbitMQTool"
)

//Task the json format of task
type Task struct {
	Path   string `json:"path"`
	Method string `json:"method"`
	Query  string `json:"query"`
	Body   string `json:"body"`
}

func doFunc(jsonTask string) string {

	var task Task
	json.Unmarshal([]byte(jsonTask), &task)
	res := fmt.Sprintf("Done %s %s %s %s from golang %s", task.Path, task.Method, task.Query, task.Body, os.Getenv("HOSTNAME"))
	return res
}

func main() {
	rabbitmqHost := os.Getenv("RABBITMQ_HOST")
	rabbitmqPort, err := strconv.Atoi(os.Getenv("RABBITMQ_PORT"))

	if err != nil {
		log.Fatalf("%s: %s", "Convert port error", err)
	}

	var queueName = "golang_task"

	rc := rabbitmqtool.CreateConnection(rabbitmqHost, rabbitmqPort)

	rc.InitChannelAndStartConsuming(queueName, doFunc)

}
