package rabbitmq

import "fmt"

// ClientConfig can be use to create a Client with NewClient function.
//		MaxRetry: determine how many time. If you want to disable retry, please set it to 1. 0 will result to default config.
type ClientConfig struct {
	Host     string
	Port     int
	UserName string
	Password string
	MaxRetry int
}

// DefaultClientConfig is the default configuration of the Client.
// It will be used with Dial func or other Client without setting it with Config.
var DefaultClientConfig = ClientConfig{
	Host:     "127.0.0.1",
	Port:     5672,
	UserName: "guest",
	Password: "guest",
	MaxRetry: 10,
}

func (c *ClientConfig) amqpURL() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%d", c.UserName, c.Password, c.Host, c.Port)
}
