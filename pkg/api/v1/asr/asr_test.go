package asr

import (
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
)

var intergation bool

func TestMain(m *testing.M) {
	flag.BoolVar(&intergation, "it", false, "run intergation test in specifiy network.")
	flag.Parse()
	os.Exit(m.Run())
}
func TestParseRecognize(t *testing.T) {
	type testCase struct {
		input  string
		expect string
	}
	testCases := map[string]testCase{
		"normal": testCase{
			input:  `{"status": 0, "hypotheses": [{"utterance": "請 消耗 看看 尿性 糧食 念佛 號"}], "id": "83d0b419-c9bb-4542-a84b-a20ee52e33ca"}`,
			expect: "請 消耗 看看 尿性 糧食 念佛 號",
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			result, err := parseRecognize(strings.NewReader(tc.input))
			if err != nil {
				t.Fatal(err)
			}
			if result != tc.expect {
				t.Errorf("expect %s, but got %s", tc.expect, result)
			}
		})
	}
}

func TestIntergationRecognizeAPI(t *testing.T) {
	testing.Short()
	if !intergation {
		t.Skip("This is a intergation test, only run if -it is passed")
	}
	f, err := os.Open("./testdata/test.wav")
	if err != nil {
		t.Fatal(err)
	}
	u, _ := url.Parse("http://192.168.3.191:8080/")
	c := Client{
		Location: u,
		Client:   http.DefaultClient,
	}
	sentence, err := c.Recognize(f)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(sentence)
}
