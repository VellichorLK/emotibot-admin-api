package rabbitmq

import (
	"errors"
	"fmt"
	"reflect"
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
	newIntegrationClient(t)
}
