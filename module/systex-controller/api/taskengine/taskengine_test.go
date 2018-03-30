package taskengine

import (
	"flag"
	"net/http"
	"net/url"
	"testing"
)

var intergation = flag.Bool("it", false, "run intergation test in specifiy network.")

func TestIntergationET(t *testing.T) {
	var c = Client{
		Client: http.DefaultClient,
		AppID:  "0acf81d19b8f9dd6cce363bea9b4810f",
		Location: &url.URL{
			Scheme: "http",
			Host:   "192.168.3.191:14101",
		},
	}
	data, err := c.ET("123", "订餐厅")
	if err != nil {
		t.Fatal(err)
	}
	resp, err := ParseETResponse(data)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Flag != 0 {
		t.Errorf("expect flag to be 0, but got %d", resp.Flag)
	}
	if resp.Text != "请问您要订哪间餐厅？" {
		t.Errorf("expect text to be \"请问您要订哪间餐厅？\", but got %s", resp.Text)
	}
}
