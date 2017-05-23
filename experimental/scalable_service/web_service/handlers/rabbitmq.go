package handlers

import (
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"sync/atomic"
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
	amqpURL := "amqp://guest:guest@" + rc.Host + ":" + strconv.Itoa(rc.Port)
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

func (rc *RabbitmqConnection) testConnection() bool {

	ch, err := rc.Conn.Channel()
	if err != nil {
		return true
	}
	ch.Close()
	return false
}

//InitChannelAndSendTask create channel, send task and wait result
func InitChannelAndSendTask(rc *RabbitmqConnection, task string, taskQueue string, ms int) (string, string, int) {

	var res string

	ch, err := rc.Conn.Channel()

	if err != nil {
		failOnError(err, "Failed to open a channel")
		return "Failed to open a channel", "text/plain", http.StatusServiceUnavailable
	}

	atomic.AddUint32(&rc.Count, 1)
	defer atomic.AddUint32(&rc.Count, ^uint32(0))
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		true,  // delete when usused
		true,  // exclusive
		false, // noWait
		nil,   // arguments
	)
	if err != nil {
		failOnError(err, "Failed to declare a queue")
		return "Failed to declare a queue", "text/plain", http.StatusServiceUnavailable
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		failOnError(err, "Failed to register a consumer")
		return "Failed to register a consumer", "text/plain", http.StatusServiceUnavailable
	}

	corrID := randomString(32)

	err = ch.Publish(
		"",        // exchange
		taskQueue, // routing key
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType:   "text/plain",
			CorrelationId: corrID,
			ReplyTo:       q.Name,
			Body:          []byte(task),
		})
	if err != nil {
		failOnError(err, "Failed to publish a message")
		return "Failed to publish a message", "text/plain", http.StatusServiceUnavailable
	}

	for {
		select {
		case d := <-msgs:
			if corrID == d.CorrelationId {
				res = string(d.Body)
				//contentType = d.ContentType
				return res, "", http.StatusOK
			}
		case <-time.After(time.Millisecond * time.Duration(ms)):
			return "Timeout", "text/plain", http.StatusRequestTimeout
		}
	}
}
