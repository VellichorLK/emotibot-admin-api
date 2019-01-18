package rabbitmq

import (
	"flag"
	"os"
	"testing"
)

var isIntegration bool

func TestMain(m *testing.M) {
	flag.BoolVar(&isIntegration, "integration", false, "flag for running integration test")
	flag.Parse()
	os.Exit(m.Run())
}

func skipIntergartion(t *testing.T) {
	if !isIntegration {
		t.Skip("skip intergration test, please specify -intergation flag.")
	}
	return
}

func newIntegrationClient(t *testing.T) *Client {
	c, err := Dial("amqp://guest:guest@127.0.0.1:5672")
	if err != nil {
		t.Fatal("expect client connect success but got error: ", err)
	}
	return c
}
