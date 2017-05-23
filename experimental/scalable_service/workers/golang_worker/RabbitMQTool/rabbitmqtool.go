package rabbitmqtool

import (
	"log"
	"strconv"
	"time"

	"github.com/streadway/amqp"
)

func failOnError(err error, msg string) {
	log.Println(msg, err)
}

//CreateConnection create a new connection to rabbitmq
func CreateConnection(host string, port int) *RabbitmqConnection {
	rc := &RabbitmqConnection{Host: host, Port: port}
	rc.connectToRabbitMQ()
	return rc
}

//ConnectToRabbitMQ connect to rabbitmq server
func (rc *RabbitmqConnection) connectToRabbitMQ() {
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
	rc.Lock.Lock()
	defer rc.Lock.Unlock()
	if rc.testConnection() {
		rc.Conn.Close()
		rc.connect()
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

//InitChannelAndStartConsuming init a new channel, start to consume task
func (rc *RabbitmqConnection) InitChannelAndStartConsuming(taskQueue string, doFunc func(string) string) bool {

	for {

		ch, err := rc.Conn.Channel()
		if err != nil {
			failOnError(err, "Failed to open a channel")
			return false
		}

		q, err := ch.QueueDeclare(
			taskQueue, // name
			false,     // durable
			false,     // delete when unused
			false,     // exclusive
			false,     // no-wait
			nil,       // arguments
		)

		if err != nil {
			ch.Close()
			failOnError(err, "Failed to declare a queue")
			return false
		}

		err = ch.Qos(
			1,     // prefetch count
			0,     // prefetch size
			false, // global
		)
		if err != nil {
			ch.Close()
			failOnError(err, "Failed to set QoS")
			return false
		}

		msgs, err := ch.Consume(
			q.Name, // queue
			"",     // consumer
			false,  // auto-ack
			false,  // exclusive
			false,  // no-local
			false,  // no-wait
			nil,    // args
		)

		if err != nil {
			ch.Close()
			failOnError(err, "Failed to register a consumer")
			return false
		}

		for d := range msgs {
			response := doFunc(string(d.Body[:]))

			err = ch.Publish(
				"",        // exchange
				d.ReplyTo, // routing key
				false,     // mandatory
				false,     // immediate
				amqp.Publishing{
					ContentType:   "text/plain",
					CorrelationId: d.CorrelationId,
					Body:          []byte(response),
				})
			if err != nil {
				failOnError(err, "Failed to publish a message")
				break
			}
			d.Ack(false)
		}

		ch.Close()
		rc.Reconnect()

	}

}
