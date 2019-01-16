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
		ch.Close()
		return nil, fmt.Errorf("failed to declare a queue, %v", err)
	}
	msg, ok, err := ch.Get(q.Name, true)
	if !ok {
		return nil, fmt.Errorf("queue is empty")
	}
	if err != nil {
		return nil, fmt.Errorf("get message from queue failed, %v", err)
	}
	return msg.Body, nil

}
