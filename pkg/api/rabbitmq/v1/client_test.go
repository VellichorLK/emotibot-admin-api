package rabbitmq

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
	"testing"

	"github.com/streadway/amqp"
)

type mockConnection struct{}

func (m *mockConnection) NotifyBlocked(receiver chan amqp.Blocking) chan amqp.Blocking {
	return nil
}
func (m *mockConnection) NotifyClose(receiver chan *amqp.Error) chan *amqp.Error {
	return nil
}
func (m *mockConnection) Channel() (*amqp.Channel, error) {
	return nil, nil
}
func (m *mockConnection) Close() error {
	return nil
}

func TestNewClient(t *testing.T) {
	tmp := connect
	defer func() {
		connect = tmp
	}()
	mockConnector := func(c *Client, maxRetry int) error {
		c.Conn = &mockConnection{}
		return nil
	}
	errorConnector := func(c *Client, maxRetry int) error {
		return fmt.Errorf("unexpected error")
	}
	testTables := []struct {
		Name      string
		Input     ClientConfig
		Connector func(*Client, int) error
		//Output should be config or error
		Output interface{}
	}{
		{
			Name:      "default case",
			Input:     ClientConfig{},
			Connector: mockConnector,
			Output:    DefaultClientConfig,
		},
		{
			Name:      "unexpect error",
			Input:     ClientConfig{},
			Connector: errorConnector,
			Output:    errors.New("unexpected error"),
		},
	}

	for _, tc := range testTables {
		t.Run(tc.Name, func(tt *testing.T) {
			connect = tc.Connector
			c, err := NewClient(tc.Input)

			if output, shouldReturnErr := tc.Output.(error); shouldReturnErr {
				if err == nil {
					tt.Fatal("expect error ", output, ", but got nil")
				}
				return
			} else if !shouldReturnErr && err != nil {
				tt.Fatal("expect error to be nil, but got: ", err)
			}
			if !reflect.DeepEqual(c.config, tc.Output) {
				tt.Logf("config: %+v\nexpect: %+v\n", c.config, tc.Output)
				tt.Fatal("expect config value to be exact with tc output")
			}
		})
	}

}

func TestI11Connect(t *testing.T) {
	c := newIntegrationClient(t)
	c.Close()
}

func TestI11ProduceAndConsume(t *testing.T) {
	c := newIntegrationClient(t)
	defer c.Close()
	pc := ProducerConfig{
		QueueName:   "testing",
		ContentType: "text/plain",
		MaxRetry:    10,
	}
	p := c.NewProducer(pc)
	example := []byte("It is a test")
	err := p.Produce(example)
	if err != nil {
		t.Fatal("produce failed, ", err)
	}
	consumer := c.NewConsumer(ConsumerConfig{
		QueueName: "testing",
		MaxRetry:  10,
	})
	received, err := consumer.Consume()
	if err != nil {
		t.Fatalf("consume failed, %v", err)
	}
	if !reflect.DeepEqual(example, received) {
		t.Logf("received:%s\nexpect:%s\n", received, example)
		t.Error("expect example to be same with received")
	}
}

func TestI11Subscribe(t *testing.T) {
	c := newIntegrationClient(t)
	defer c.Close()
	p := c.NewProducer(ProducerConfig{
		QueueName:   "testing",
		ContentType: "text/plain",
	})
	wg := sync.WaitGroup{}
	products := 10
	wg.Add(products)
	go func() {
		for i := 0; i < products; i++ {
			p.Produce([]byte("hello world"))
		}
		// fmt.Println("produce finished")
	}()
	consumer := c.NewConsumer(ConsumerConfig{
		QueueName: "testing",
		MaxRetry:  10,
	})
	received := 0
	consumer.Subscribe(func(response []byte) error {
		received++
		wg.Done()
		return nil
	})
	wg.Wait()
	if received != 10 {
		t.Error("expect subscribe receive 10 message, but got ", received)
	}

}
