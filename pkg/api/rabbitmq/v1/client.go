package rabbitmqtool

import (
	"fmt"
	"sync"
	"time"

	"emotibot.com/emotigo/pkg/logger"

	"github.com/streadway/amqp"
)

// DefaultClientConfig is the default configuration of the Client.
// It will be used with Dial func or other Client without setting it with Config.
var DefaultClientConfig = ClientConfig{
	Host:     "127.0.0.1",
	Port:     5672,
	UserName: "guest",
	Password: "guest",
	MaxRetry: 10,
}

// Client can create a Connection to RabbitMQ server.
// Which its setting can base on ClientConfig or amqpURL.
type Client struct {
	Conn    Connection
	amqpURL string
	config  ClientConfig
	Lock    sync.Mutex
}

// ClientConfig can be use to create a Client with NewClient function.
//		MaxRetry: determine how many time. If you want to disable retry, please set it to 1. 0 will result to default config.
type ClientConfig struct {
	Host     string
	Port     int
	UserName string
	Password string
	MaxRetry int
}

// Connection represents the real dialer for the RabbitMQ connection.
type Connection interface {
	// Channel opens a unique, concurrent server channel to process the bulk of
	// AMQP messages. Any error from methods on this receiver will render the
	// receiver invalid and a new Channel should be opened.
	Channel() (*amqp.Channel, error)
	// Close requests and waits for the response to close the AMQP connection.
	//
	//     It's advisable to use this message when publishing to ensure all kernel
	//     buffers have been flushed on the server and client before exiting.
	//
	//     An error indicates that server may not have received this request to close
	//     but the connection should be treated as closed regardless.
	//
	//     After returning from this call, all resources associated with this
	//     connection, including the underlying io, Channels, Notify listeners and
	//     Channel consumers will also be closed.
	Close() error
	// NotifyBlocked(receiver chan amqp.Blocking) chan amqp.Blocking
	// NotifyClose(receiver chan *amqp.Error) chan *amqp.Error
}

// NewClient create the Client based on config. any empty setting will use the default setting.
func NewClient(config ClientConfig) (*Client, error) {
	if config.Host == "" {
		config.Host = DefaultClientConfig.Host
	}
	if config.Port == 0 {
		config.Port = DefaultClientConfig.Port
	}
	if config.UserName == "" {
		config.UserName = DefaultClientConfig.UserName
	}
	if config.Password == "" {
		config.Password = DefaultClientConfig.Password
	}
	if config.MaxRetry == 0 {
		config.MaxRetry = DefaultClientConfig.MaxRetry
	}
	c := &Client{
		config:  config,
		amqpURL: config.amqpURL(),
		Lock:    sync.Mutex{},
	}
	return c, nil
}

// Dial create a quick connection to RabbitMQ server by the given url
func Dial(url string) (*Client, error) {
	c := &Client{
		amqpURL: url,
		Lock:    sync.Mutex{},
	}
	err := connect(c, DefaultClientConfig.MaxRetry)
	if err != nil {
		return nil, fmt.Errorf("dial failed, %v", err)
	}
	return c, nil
}

func (c *ClientConfig) amqpURL() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%d", c.UserName, c.Password, c.Host, c.Port)
}

func connect(c *Client, maxRetry int) error {
	var err error
	for i := 0; ; i++ {
		if maxRetry > 0 && i < maxRetry {
			return fmt.Errorf("connect to RabbitMQ server [%s] failed: %v", c.amqpURL, err)
		}
		c.Conn, err = amqp.Dial(c.amqpURL)
		if err == nil {
			break
		}
		time.Sleep(5 * time.Second)
	}
	logger.Info.Printf("Connect to RabbitMQ server %s:%d success.", c.config.Host, c.config.Port)
	return nil
}

//reconnect will try to reconnect to RabbitMQ.
func (c *Client) reconnect() {
	c.Lock.Lock()
	defer c.Lock.Unlock()
	if c.IsUnreachable() {
		connect(c, 0)
	}

}

// IsUnreachable will try to create a channel with Conn. and return bool for successfulness.
// Any unreachable situation will let the Client to close its Connection.
func (c *Client) IsUnreachable() bool {
	if c.Conn == nil {
		return true
	}
	ch, err := c.Conn.Channel()
	if err != nil {
		c.Conn.Close()
		return true
	}
	defer ch.Close()
	return false
}

// Consume init a new channel, and start to consume task.
//	notice: It will block the caller routine.
func (c *Client) Consume(queueName string, doFunc func(string) string) error {

	for {

		ch, err := c.Conn.Channel()
		if err != nil {
			return fmt.Errorf("failed to open a channel, %v", err)
		}

		q, err := ch.QueueDeclare(
			queueName, // name
			false,     // durable
			false,     // delete when unused
			false,     // exclusive
			false,     // no-wait
			nil,       // arguments
		)

		if err != nil {
			ch.Close()
			return fmt.Errorf("failed to declare a queue, %v", err)
		}

		err = ch.Qos(
			1,     // prefetch count
			0,     // prefetch size
			false, // global
		)
		if err != nil {
			ch.Close()
			return fmt.Errorf("failed to set Qos, %v", err)
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
			return fmt.Errorf("failed to register a consumer, %v", err)
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
				logger.Error.Println("Failed to publish message: ", err)
				break
			}
			d.Ack(false)
		}

		ch.Close()
		c.reconnect()

	}

}
