package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"os"
	"reflect"
	"strings"
	"testing"

	"emotibot.com/emotigo/module/openapi-adapter/data"
	"emotibot.com/emotigo/module/openapi-adapter/traffic"
)

func TestMain(m *testing.M) {
	remoteHost = "http://172.16.101.98:8080"
	remoteHostURL, err := url.Parse(remoteHost)
	if err != nil {
		log.Fatalf("remoteURL is not a valid URL, %v\n", err)
	}

	// Make traffic channel
	duration := 10
	maxRequests := 20
	banPeriod := 300

	addTrafficChan = make(chan string)
	appidChan = make(chan *traffic.AppidIP, 1024)
	trafficManager = traffic.NewTrafficManager(duration, int64(maxRequests), int64(banPeriod))

	proxy = httputil.NewSingleHostReverseProxy(remoteHostURL)
	log.SetFlags(log.Ltime | log.Lshortfile)
	os.Exit(m.Run())
}

// TestAdapter will do a integration test for the adapter
func TestAdapter(t *testing.T) {
	type testCase struct {
		input  map[string]string
		expect data.ResponseV1
	}

	var td = map[string]testCase{
		"welcome": testCase{
			input: map[string]string{
				"cmd":    "chat",
				"appid":  "csbot",
				"userid": "IntegrationTestUser",
				"text":   "welcome_tag",
			},
			expect: data.ResponseV1{
				ReturnCode: 200,
				Message:    "success",
				Data: []data.DataV1{
					data.DataV1{
						Type:  "text",
						Cmd:   "",
						Value: "您好，很高兴为您服务",
						Data: []data.Answer{
							data.Answer{
								Type:       "text",
								SubType:    "text",
								Value:      "您好，很高兴为您服务",
								Data:       []interface{}{},
								ExtendData: "",
							},
						},
					},
				},
				Emotion: []data.Emotion{
					data.Emotion{},
				},
			},
		},
		"creditcard": testCase{
			input: map[string]string{
				"cmd":    "chat",
				"appid":  "csbot",
				"userid": "IntegrationTestUser",
				"text":   "我要办信用卡",
			},
			expect: data.ResponseV1{
				ReturnCode: 200,
				Message:    "success",
				Data: []data.DataV1{
					data.DataV1{
						Type:  "text",
						Cmd:   "",
						Value: "近似问: 1.办信用卡有什么优惠 2.办理信用卡有佣金吗",
						Data: []data.Answer{
							data.Answer{
								Type:    "text",
								SubType: "guslist",
								Value:   "近似问",
								Data: []interface{}{
									"办信用卡有什么优惠",
									"办理信用卡有佣金吗",
								},
								ExtendData: "",
							},
						},
					},
				},
				Emotion: []data.Emotion{
					data.Emotion{
						Type:  "text",
						Value: "中性",
						Score: "80",
					},
				},
			},
		},
		"smallchat1": testCase{
			input: map[string]string{
				"cmd":    "chat",
				"appid":  "csbot",
				"userid": "IntegrationTestUser",
				"text":   "你叫什麼名字",
			},
			expect: data.ResponseV1{
				ReturnCode: 200,
				Message:    "success",
				Data: []data.DataV1{
					data.DataV1{
						Type:  "text",
						Cmd:   "",
						Value: "我是信仔,你也可以叫我“小智”,我是您身边的智能理财管家。",
						Data: []data.Answer{
							data.Answer{
								Type:       "text",
								SubType:    "text",
								Value:      "我是信仔,你也可以叫我“小智”,我是您身边的智能理财管家。",
								Data:       []interface{}{},
								ExtendData: "",
							},
						},
					},
				},
				Emotion: []data.Emotion{
					data.Emotion{
						Type:  "text",
						Value: "疑惑",
						Score: "80",
					},
				},
			},
		},
	}

	for name, tc := range td {
		t.Run(name, func(tt *testing.T) {
			form := url.Values{}

			for key, val := range tc.input {
				form.Add(key, val)
			}
			req, _ := http.NewRequest(http.MethodPost, "/api/ApiKey/openapi.php", strings.NewReader(form.Encode()))
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			rr := httptest.NewRecorder()

			OpenAPIAdapterHandler(rr, req)

			var resp data.ResponseV1
			bodyData, _ := ioutil.ReadAll(rr.Body)
			err := json.Unmarshal(bodyData, &resp)
			if err != nil {
				log.Printf("response body: %s\n", bodyData)
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

// TestProxy will do a integration test for the proxy
func TestProxy(t *testing.T) {
	type testCase struct {
		text   string
		appID  string
		userID string
		expect data.ResponseV2
	}

	var td = map[string]testCase{
		"welcome": testCase{
			text:   "welcome_tag",
			appID:  "csbot",
			userID: "IntegrationTestUser",
			expect: data.ResponseV2{
				Code:    200,
				Message: "success",
				Answers: []data.Answer{
					data.Answer{
						Type:       "text",
						SubType:    "text",
						Value:      "您好，很高兴为您服务",
						Data:       []interface{}{},
						ExtendData: "",
					},
				},
				Info: data.Info{},
			},
		},
		"creditcard": testCase{
			text:   "我要办信用卡",
			appID:  "csbot",
			userID: "IntegrationTestUser",
			expect: data.ResponseV2{
				Code:    200,
				Message: "success",
				Answers: []data.Answer{
					data.Answer{
						Type:    "text",
						SubType: "guslist",
						Value:   "近似问",
						Data: []interface{}{
							"办信用卡有什么优惠",
							"办理信用卡有佣金吗",
						},
						ExtendData: "",
					},
				},
				Info: data.Info{
					EmotionCat:   "中性",
					EmotionScore: 80,
				},
			},
		},
		"smallchat1": testCase{
			text:   "你叫什麼名字",
			appID:  "csbot",
			userID: "IntegrationTestUser",
			expect: data.ResponseV2{
				Code:    200,
				Message: "success",
				Answers: []data.Answer{
					data.Answer{
						Type:       "text",
						SubType:    "text",
						Value:      "我是信仔,你也可以叫我“小智”,我是您身边的智能理财管家。",
						Data:       []interface{}{},
						ExtendData: "",
					},
				},
				Info: data.Info{
					EmotionCat:   "疑惑",
					EmotionScore: 80,
				},
			},
		},
	}

	for name, tc := range td {
		t.Run(name, func(tt *testing.T) {
			text := data.V2Body{
				Text: tc.text,
			}
			textData, _ := json.Marshal(text)
			body := bytes.NewBuffer(textData)
			req, _ := http.NewRequest(http.MethodPost, "/v1/openapi", body)
			req.Header.Add("Content-Type", "application/json")
			req.Header.Add("appid", tc.appID)
			req.Header.Add("userid", tc.userID)
			rr := httptest.NewRecorder()

			OpenAPIHandler(rr, req)

			var resp data.ResponseV2
			bodyData, _ := ioutil.ReadAll(rr.Body)
			err := json.Unmarshal(bodyData, &resp)
			if err != nil {
				log.Printf("response body: %s\n", bodyData)
				tt.Fatalf("expect body format to be v2Response, but got error, %v", err)
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
