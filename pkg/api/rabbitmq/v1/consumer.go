package rabbitmq

import (
	"fmt"

	"emotibot.com/emotigo/pkg/logger"
)

//Consumer is response for listening and taking message from the config queue.
type Consumer struct {
	client   *Client
	config   ConsumerConfig
	isClosed bool
}

type ConsumerConfig struct {
	QueueName string
	maxRetry  int
}

// Task is the task that will be triggered if new message comes from queue.
type Task func(message []byte) error

func (c *Consumer) Subscribe(task Task) error {
	for {
		ch := c.client.channel()
		msgs, err := ch.Consume(
			c.config.QueueName, // queue
			"",                 // consumer
			false,              // auto-ack
			false,              // exclusive
			false,              // no-local
			false,              // no-wait
			nil,                // args
		)

		if err != nil {
			c.client.Conn.Close()
			return fmt.Errorf("failed to register a consumer, %v", err)
		}
		for d := range msgs {
			err := task(d.Body)
			if err != nil {
				logger.Error.Println("Failed to consume message: ", err)
				continue
			}
			d.Ack(false)
		}
		c.client.reconnect()
	}
}
func (c *Consumer) Consume() ([]byte, error) {
	var err error
	maxRetry := c.config.maxRetry
	for i := 0; ; i++ {
		if maxRetry > 0 && i == maxRetry {
			break
		}
		ch := c.client.channel()
		q, err := ch.QueueDeclare(
			c.config.QueueName, // name
			false,              // durable
			false,              // delete when unused
			false,              // exclusive
			false,              // no-wait
			nil,                // arguments
		)

		if err != nil {
			err = fmt.Errorf("failed to declare a queue, %v", err)
			continue
		}
		msg, ok, err := ch.Get(q.Name, true)
		if !ok {
			err = fmt.Errorf("queue is empty")
			continue
		}
		if err != nil {
			err = fmt.Errorf("get message from queue failed, %v", err)
			continue
		}
		return msg.Body, nil
	}
	return nil, fmt.Errorf("exceed max retries %d, error: %v", c.config.maxRetry, err)
}
