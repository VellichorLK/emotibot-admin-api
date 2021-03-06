package handlers

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/streadway/amqp"
)

func randomString(l int) string {
	bytes := make([]byte, l)
	for i := 0; i < l; i++ {
		bytes[i] = byte(randInt(65, 90))
	}
	return string(bytes)
}

func randInt(min int, max int) int {
	return min + rand.Intn(max-min)
}

func failOnError(err error, msg string) {
	log.Println(msg, err)

}

//ConnectToRabbitMQ connect to rabbitmq server
func (rc *RabbitmqConnection) ConnectToRabbitMQ() {
	rc.connect()
}

func (rc *RabbitmqConnection) connect() {
	amqpURL := "amqp://" + rc.User + ":" + rc.Pwd + "@" + rc.Host + ":" + strconv.Itoa(rc.Port)
	conn, err := amqp.Dial(amqpURL)
	for err != nil {
		failOnError(err, "Failed to connect to RabbitMQ")
		time.Sleep(5 * time.Second)
		conn, err = amqp.Dial(amqpURL)
	}
	rc.Conn = conn
	log.Printf("Connect to RabbitMQ server %s:%d success.", rc.Host, rc.Port)
}

//Reconnect try to recoonect to rabbitmq
func (rc *RabbitmqConnection) Reconnect() {
	getLock := rc.Lock.TryLock()
	if getLock {
		defer rc.Lock.Unlock()
		if rc.testConnection() {
			rc.Conn.Close()
			rc.connect()
		}
	}
}

//Close the rabbitmq connection
func (rc *RabbitmqConnection) Close() {
	rc.Conn.Close()
}

func (rc *RabbitmqConnection) testConnection() bool {

	ch, err := rc.Conn.Channel()
	if err != nil {
		return true
	}
	ch.Close()
	return false
}

func createQueue(ch *amqp.Channel) {
	args := make(map[string]interface{})
	args["x-max-priority"] = int16(10)

	//declare the queue
	for _, queue := range QUEUEMAP {

		var setProrities map[string]interface{}
		if queue.HasPriority {
			setProrities = args
		}

		_, err := ch.QueueDeclare(
			queue.Name,   // name
			true,         // durable
			false,        // delete when usused
			false,        // exclusive
			false,        // no-wait
			setProrities, // arguments
		)
		if err != nil {
			failOnError(err, "Failed to create queue "+queue.Name)
			panic(fmt.Sprintf("%s: %s (%s)", err, "Failed to create queue", queue.Name))
		}
	}

}

type TaskReceive func() (string, string, string, uint8)

//StartSendTaskService receive the task from web, format it to json and push to rabbitmq
func StartSendTaskService(host string, port int, user string, pwd string, taskReceive TaskReceive) {
	rc := &RabbitmqConnection{Host: host, Port: port, User: user, Pwd: pwd}
	rc.ConnectToRabbitMQ()
	defer rc.Close()

	ch, err := rc.Conn.Channel()

	if err != nil {
		failOnError(err, "Failed to open a channel")
		panic(fmt.Sprintf("%s: %s", err, "Failed to open a channel"))
	}
	defer ch.Close()

	createQueue(ch)

	for {
		/*
			taskInfo := <-TaskQueue
			task := taskInfo.packagedTask
			corrID := taskInfo.fileInfo.FileID
			queueName := taskInfo.QueueN
		*/
		task, corrID, queueName, priority := taskReceive()
		//log.Printf("task:%s, corrID:%s, queueName:%s, priority:%v\n ", task, corrID, queueName, priority)

		err = ch.Publish(
			"",        // exchange
			queueName, // routing key
			false,     // mandatory
			false,     // immediate
			amqp.Publishing{
				ContentType:   "text/plain",
				CorrelationId: corrID,
				Body:          []byte(task),
				Priority:      priority,
				DeliveryMode:  amqp.Persistent,
			})
		if err != nil {
			RelyQueue <- false
			failOnError(err, "Failed to publish a message. Reconnect to rabbitmq")
			ch = rebuildChannel(rc)
		} else {
			RelyQueue <- true
		}
	}

}

type DoFunc func(string) (string, string)

func StartReceiveTaskService(host string, port int, user string, pwd string, queueN string, doFunc DoFunc) {
	rc := &RabbitmqConnection{Host: host, Port: port, User: user, Pwd: pwd}

	rc.ConnectToRabbitMQ()
	defer rc.Close()

	ch, err := rc.Conn.Channel()

	if err != nil {
		failOnError(err, "Failed to open a channel")
		panic(fmt.Sprintf("%s: %s", err, "Failed to open a channel"))
	}
	defer ch.Close()

	createQueue(ch)

	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		failOnError(err, "Failed to set QoS")
		panic(fmt.Sprintf("%s: %s (%s)", err, "Failed to set QoS", queueN))
	}

	msgs, err := ch.Consume(
		queueN, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)

	if err != nil {
		failOnError(err, "Failed to register a consumer")
		panic(fmt.Sprintf("%s: %s (%s)", err, "Failed to register a consumer", queueN))
	}

	for d := range msgs {
		response, replyQueue := doFunc(string(d.Body[:]))

		if replyQueue != "" {
			err = ch.Publish(
				"",         // exchange
				replyQueue, // routing key
				false,      // mandatory
				false,      // immediate
				amqp.Publishing{
					ContentType:   "text/plain",
					CorrelationId: d.CorrelationId,
					Body:          []byte(response),
				})
			if err != nil {
				failOnError(err, "Failed to publish a message")
				break
			}
		}
		d.Ack(false)
	}

}

func rebuildChannel(rc *RabbitmqConnection) *amqp.Channel {
	rc.Reconnect()
	ch, err := rc.Conn.Channel()

	if err != nil {
		failOnError(err, "Failed to open a channel")
		return nil
	}
	return ch
}
