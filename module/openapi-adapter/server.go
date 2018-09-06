package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"time"
)

//ResponseV1 represent version 1 houta response
type ResponseV1 struct {
	ReturnCode int           `json:"return"`
	Message    string        `json:"return_message"`
	Answers    []interface{} `json:"data"`
	Emotion    []Emotion     `json:"emotion"`
}

type ResponseV2 struct {
	Code    int           `json:"status"`
	Message string        `json:"message"`
	Answers []interface{} `json:"data"`
	Info    Info          `json:"info"`
}

//Data represent version 1 Answer response structure
type Data struct {
	Type  string        `json:"type"`
	Cmd   string        `json:"cmd"`
	Value string        `json:"value"`
	Data  []interface{} `json:"data"`
}

//Emotion represent version 1 emotion response structure
type Emotion struct {
	Type  string      `json:"type"`
	Value string      `json:"value"`
	Score string      `json:"score"`
	Data  interface{} `json:"data"`
}

type Info struct {
	EmotionCat   string `json:"emotion"`
	EmotionScore int    `json:"emotionScore"`
}

type v2Body struct {
	Text       string                 `json:"text"`
	SourceID   string                 `json:"sourceId,omitempty"`
	ClientID   string                 `json:"clientId,omitempty"`
	CustomInfo map[string]string      `json:"customInfo,omitempty"`
	ExtendData map[string]interface{} `json:"extendData,omitempty"`
}

var remoteHost string
var client = &http.Client{
	Timeout: time.Duration(2) * time.Second,
}
var proxy *httputil.ReverseProxy

var logger = log.New(ioutil.Discard, "", log.Ltime|log.Lshortfile)

func main() {
	var (
		ok    bool
		level string
	)

	port, ok := os.LookupEnv("OPENAPI_ADAPTER_SERVER_PORT")
	if !ok {
		port = "8080"
	}

	remoteHost, ok = os.LookupEnv("OPENAPI_ADAPTER_EC_HOST")
	if !ok {
		log.Fatal("please specify openapi v2 remote url by os Env OPENAPI_ADAPTER_EC_HOST")
	}

	remoteHostURL, err := url.Parse(remoteHost)
	if err != nil {
		log.Fatalf("remoteHost is not a valid URL, %v\n", err)
	}

	level, ok = os.LookupEnv("OPENAPI_LOG_LEVEL")
	if l, err := strconv.ParseInt(level, 10, 64); ok && err == nil && l == 0 {
		//level=0, equal to trace logger
		logger = log.New(os.Stdout, "", log.Ltime|log.Lshortfile)
	}

	// Reserve proxy
	proxy = httputil.NewSingleHostReverseProxy(remoteHostURL)

	http.HandleFunc("/api/ApiKey/", OpenAPIAdapterHandler)
	http.HandleFunc("/v1/openapi", OpenAPIHandler)

	logger.Printf("Starting server at port: %s", port)
	logger.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}

// OpenAPIAdapterHandler will translate v1 request into v2 Request and then send to BFOP Server
// response will parse to v1 response format
func OpenAPIAdapterHandler(w http.ResponseWriter, v1 *http.Request) {
	var bodystr, err = ioutil.ReadAll(v1.Body)
	if err != nil {
		http.Error(w, "", 400)
		return
	}
	v1.Body.Close()

	var v1Body map[string]string

	err = json.Unmarshal(bodystr, &v1Body)
	if err != nil {
		customError(w, "body formatted err:"+err.Error(), 500)
		return
	}
	var (
		body   io.Reader
		data   []byte
		v2Req  *http.Request
		resp   *http.Response
		v1Resp ResponseV1
		v2Resp ResponseV2
	)
	data, err = json.Marshal(v2Body{
		Text: v1Body["text"],
	})
	if err != nil {
		customError(w, "transform request failed, "+err.Error(), 500)
		return
	}
	body = bytes.NewBuffer(data)

	remoteURL := fmt.Sprintf("%s/v1/openapi", remoteHost)
	v2Req, err = http.NewRequest(http.MethodPost, remoteURL, body)
	if err != nil {
		customError(w, "transform request failed, "+err.Error(), 500)
		return
	}
	v2Req.Header.Set("appId", v1Body["appid"])
	v2Req.Header.Set("userId", v1Body["userid"])

	resp, err = client.Do(v2Req)
	if err != nil {
		customError(w, "http request failed, "+err.Error(), 500)
		return
	}
	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		customError(w, "io failed, "+err.Error(), 500)
		return
	}
	defer resp.Body.Close()
	err = json.Unmarshal(data, &v2Resp)
	logger.Printf("v2 response body: %s\n", data)
	if err != nil {
		customError(w, "unmarshal version 2 body failed, "+err.Error(), 500)
		return
	}
	v1Resp.Answers = v2Resp.Answers
	v1Resp.ReturnCode = v2Resp.Code
	v1Resp.Message = v2Resp.Message
	v1Resp.Emotion = []Emotion{
		newEmotion(v2Resp),
	}
	result, _ := json.Marshal(v1Resp)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(result)
}

// OpenAPIHandler will do nothing but proxy the request/response
// to/from BFOP server directly
func OpenAPIHandler(w http.ResponseWriter, r *http.Request) {
	proxy.ServeHTTP(w, r)
}

func customError(w http.ResponseWriter, message string, code int) {
	//Response need to be 200 OK all the time to mimic behavior in v1
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	var resp = ResponseV1{
		ReturnCode: code,
		Message:    message,
		Answers:    []interface{}{},
		Emotion:    []Emotion{},
	}
	data, err := json.Marshal(resp)
	if err != nil {
		logger.Println("error json marshal failed. ", err)
	}
	w.Write(data)
}

func newEmotion(resp ResponseV2) Emotion {
	var e = Emotion{}
	if resp.Info.EmotionCat != "" {
		e.Type = "text"
		e.Value = resp.Info.EmotionCat
		e.Score = strconv.Itoa(resp.Info.EmotionScore)
	}
	return e
}
