package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
)

func TestMain(m *testing.M) {
	remoteURL = "http://172.16.101.98:8080/v1/openapi"
	log.SetFlags(log.Ltime | log.Lshortfile)
	os.Exit(m.Run())
}

//TestIntegrated will do a integration test for the adapter
func TestIntegrated(t *testing.T) {
	type testCase struct {
		input  map[string]string
		expect ResponseV1
	}
	var td = map[string]testCase{
		"welcome": testCase{
			input: map[string]string{
				"text":   "welcome_tag",
				"appid":  "csbot",
				"userid": "IntegrationTestUser",
			},
			expect: ResponseV1{
				ReturnCode: 200,
				Message:    "success",
				Answers: []interface{}{
					map[string]interface{}{
						"type":       "text",
						"subType":    "text",
						"value":      "您好，很高兴为您服务",
						"data":       []interface{}{},
						"extendData": "",
					},
				},
				Emotion: []Emotion{
					Emotion{},
				},
			},
		},
		"smallchat1": testCase{
			input: map[string]string{
				"text":   "你叫什麼名字",
				"appid":  "csbot",
				"userid": "IntegrationTestUser",
			},
			expect: ResponseV1{
				ReturnCode: 200,
				Message:    "success",
				Answers: []interface{}{
					map[string]interface{}{
						"type":       "text",
						"subType":    "text",
						"value":      "我是信仔,你也可以叫我“小智”,我是您身边的智能理财管家。",
						"data":       []interface{}{},
						"extendData": "",
					},
				},
				Emotion: []Emotion{
					Emotion{
						Type:  "text",
						Value: "疑惑",
						Score: "0",
					},
				},
			},
		},
	}
	for name, tc := range td {
		t.Run(name, func(tt *testing.T) {
			data, _ := json.Marshal(tc.input)
			body := bytes.NewBuffer(data)
			req, _ := http.NewRequest(http.MethodPost, "/", body)
			rr := httptest.NewRecorder()

			OpenAPIAdapterHandler(rr, req)

			var resp ResponseV1
			data, _ = ioutil.ReadAll(rr.Body)
			err := json.Unmarshal(data, &resp)
			if err != nil {
				log.Printf("response body: %s\n", data)
				tt.Fatalf("expect body format to be v1Response, but got error, %v", err)
			}
			if rr.Code != http.StatusOK {

				tt.Fatalf("expect status code OK but got %d, message: %s", rr.Code, resp.Message)
			}

			if !reflect.DeepEqual(tc.expect, resp) {
				tt.Fatalf("expect response to be %+v, but got %+v", tc.expect, resp)
			}
		})
	}
}
