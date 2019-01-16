// Package rabbitmq v1 is the refractor version of the emotibot.com/emotigo/experimental/scalable_service/workers/golang_worker/RabbitMQTool.
// It warp the RabbitMQ communication as a Client. And offer patterns of produce and consume.
package rabbitmq

import (
	"fmt"
	"sync"
	"time"

	"github.com/streadway/amqp"
)

// Client is a persist connection to RabbitMQ server which will maintain a single connection with channel.
// It create the Connection by given config or amqpURL.
// Only get Client by NewClient or Dial function.
// DO NOT CREATE Client by literal, or the monitor part will fail.
type Client struct {
	amqpURL  string
	Conn     Connection
	config   ClientConfig
	rChannel *amqp.Channel
	wChannel *amqp.Channel
	close    chan interface{}
	lock     sync.RWMutex
}

// Connection abstract the dialing to the RabbitMQ connection.
// 	Connection should manage the serialization and deserialization of frames from IO
// 	and dispatches the frames to the appropriate channel. All RPC methods and
// 	asynchronous Publishing, Delivery, Ack, Nack and Return messages are
// 	multiplexed on this channel. There must always be active receivers for every
// 	asynchronous message on this connection.
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
	// NotifyBlocked registers a listener for RabbitMQ specific TCP flow control
	// method extensions connection.blocked and connection.unblocked. Flow control
	// is active with a reason when Blocking.Blocked is true. When a Connection is
	// blocked, all methods will block across all connections until server
	// resources become free again.
	//
	// This optional extension is supported by the server when the
	// "connection.blocked" server capability key is true.
	NotifyBlocked(receiver chan amqp.Blocking) chan amqp.Blocking
	// NotifyClose registers a listener for close events either initiated by an
	// error accompanying a connection.close method or by a normal shutdown.
	//
	// On normal shutdowns, the chan will be closed.
	//
	// To reconnect after a transport or protocol error, register a listener here
	// and re-run your setup process.
	NotifyClose(receiver chan *amqp.Error) chan *amqp.Error
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
		lock:    sync.RWMutex{},
		close:   make(chan interface{}),
	}
	err := connect(c, config.MaxRetry)
	if err != nil {
		return nil, fmt.Errorf("try connecting server failed with retries %d, %v", config.MaxRetry, err)
	}
	go keepConnAlive(c)
	return c, nil
}

// Dial create a quick connection to RabbitMQ server by the given url
func Dial(url string) (*Client, error) {
	c := &Client{
		amqpURL: url,
		lock:    sync.RWMutex{},
		close:   make(chan interface{}),
	}
	err := connect(c, DefaultClientConfig.MaxRetry)
	if err != nil {
		return nil, fmt.Errorf("connect failed, %v", err)
	}
	go keepConnAlive(c)
	return c, nil
}

// Close will cancel persist check and close the connection.
func (c *Client) Close() error {
	c.close <- struct{}{}
	err := c.Conn.Close()
	return err
}

var connect = func(c *Client, maxRetry int) error {
	var err error
	c.lock.Lock()
	defer c.lock.Unlock()

	for i := 0; ; i++ {
		//clean up old resource
		if c.Conn != nil {
			c.Conn.Close()
		}
		if i > 0 {
			time.Sleep(time.Duration(100) * time.Millisecond)
		}
		if maxRetry > 0 && i == maxRetry {
			return fmt.Errorf("connection exceed maxRetry times: %d. err: %v", maxRetry, err)
		}
		c.Conn, err = amqp.Dial(c.amqpURL)
		if err != nil {
			err = fmt.Errorf("dial to amqp url [%s] failed, %v", c.amqpURL, err)
			continue
		}
		c.wChannel, err = c.Conn.Channel()
		if err != nil {
			err = fmt.Errorf("create write channel failed, %v", err)
			continue
		}
		c.rChannel, err = c.Conn.Channel()
		if err != nil {
			err = fmt.Errorf("create read channel failed, %v", err)
			continue
		}
		break
	}
	return err
}

// reconnect will try to reconnect the connection between RabbitMQ server
func (c *Client) reconnect() {
	connect(c, 0)
}

// channel get the most recent working channel from the Client.
func (c *Client) rwChannels() (*amqp.Channel, *amqp.Channel) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.rChannel, c.wChannel
}

// IsUnreachable will try to create a channel with Conn. and return bool for successfulness.
func (c *Client) IsUnreachable() bool {
	if c.Conn == nil {
		return true
	}
	ch, err := c.Conn.Channel()
	if err != nil {
		return true
	}
	defer ch.Close()
	return false
}

// NewConsumer create a consumer with the config and associate the client.
func (c *Client) NewConsumer(config ConsumerConfig) *Consumer {
	return &Consumer{
		client: c,
		config: config,
	}
}

// NewProducer create a producer with the config.
func (c *Client) NewProducer(config ProducerConfig) *Producer {
	p := &Producer{
		client: c,
		config: config,
	}
	return p
}

func keepConnAlive(c *Client) {
	if c.Conn == nil {
		connect(c, 0)
	}
	closeNotifier := c.Conn.NotifyClose(make(chan *amqp.Error))
	for {
		select {
		case <-closeNotifier:
			c.reconnect()
			// renew a notifier
			closeNotifier = c.Conn.NotifyClose(make(chan *amqp.Error))
		case <-c.close:
			return
		}
	}
}
