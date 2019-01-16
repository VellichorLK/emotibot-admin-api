package rabbitmq

import (
	"fmt"
	"time"

	"github.com/streadway/amqp"
)

// Producer is the RabbitMQ queue publisher.
type Producer struct {
	client   *Client
	config   ProducerConfig
	isClosed bool
}

type ProducerConfig struct {
	QueueName   string
	ContentType string
	MaxRetry    int
}

// Produce send the product to the configured queue.
// It is a synchronize method to publish the product.
// If the producer is already closed or publish failed, It will return a error.
func (p *Producer) Produce(product []byte) error {
	var (
		err      error
		q        amqp.Queue
		maxRetry = p.config.MaxRetry
	)
	for i := 0; ; i++ {
		if maxRetry > 0 && i == maxRetry {
			break
		}
		if i > 0 {
			time.Sleep(time.Duration(100) * time.Millisecond)
		}
		if p.isClosed {
			return fmt.Errorf("producer is already closed")
		}
		ch := p.client.channel()
		q, err = ch.QueueDeclare(
			p.config.QueueName, // name
			false,              // durable
			false,              // delete when unused
			false,              // exclusive
			false,              // no-wait
			nil,                // arguments
		)
		if err != nil {
			err = fmt.Errorf("queue declare failed, %v", err)
			continue
		}

		publish := amqp.Publishing{
			ContentType: p.config.ContentType,
			Body:        product,
		}
		err = ch.Publish(
			"",     // exchange
			q.Name, // routing key
			false,  // mandatory
			false,  // immediate
			publish,
		)
		if err == nil {
			return nil
		}
		err = fmt.Errorf("publish to queue failed, %v", err)
	}
	return fmt.Errorf("retry max times %d reached, error: %v", maxRetry, err)

}

// Close will prevent any new product to be published by the producer.
// It is a low-cost operation. which it's client connection WILL NOT BE CLOSED.
func (p *Producer) Close() {
	p.isClosed = true
}
