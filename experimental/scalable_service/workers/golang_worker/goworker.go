package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"

	"strings"

	"emotibot.com/emotigo/experimental/scalable_service/workers/golang_worker/RabbitMQTool"
)

//Task the json format of task
type Task struct {
	Path   string `json:"path"`
	Method string `json:"method"`
	Query  string `json:"query"`
	Body   string `json:"body"`
}

func fib(n int) int {
	if n == 0 {
		return 0
	} else if n == 1 {
		return 1
	} else {
		return fib(n-1) + fib(n-2)
	}
}

func doFunc(jsonTask string) string {

	var task Task
	var res string
	json.Unmarshal([]byte(jsonTask), &task)
	if task.Method == "GET" {
		querys := strings.Split(task.Query, "&")
		for _, query := range querys {
			if strings.HasPrefix(query, "n=") {
				numString := strings.Trim(query, "n=")

				n, err := strconv.Atoi(string(numString))
				if err != nil {
					res = fmt.Sprintf("Wrong query, n is not number from golang %s", os.Getenv("HOSTNAME"))
				} else {
					res = strconv.Itoa(fib(n))
				}

			}
		}
		if res == "" {
			res = fmt.Sprintf("Wrong query format %s from golang %s", task.Query, os.Getenv("HOSTNAME"))
		}

	} else {
		res = fmt.Sprintf("Done %s %s %s %s from golang %s", task.Path, task.Method, task.Query, task.Body, os.Getenv("HOSTNAME"))
	}

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
