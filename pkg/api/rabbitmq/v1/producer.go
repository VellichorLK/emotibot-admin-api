package rabbitmq

import (
	"fmt"

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
		err error
		q   amqp.Queue
	)
	for i := 0; i < p.config.MaxRetry; i++ {
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
		if err != nil {
			err = fmt.Errorf("publish to queue failed, %v", err)
			continue
		} else {
			break
		}
	}

	return err
}

// Close will prevent any new product to be published by the producer.
// It is a low-cost operation. which it's client connection WILL NOT BE CLOSED.
func (p *Producer) Close() {
	p.isClosed = true
}
