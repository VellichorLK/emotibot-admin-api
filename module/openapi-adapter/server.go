package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"emotibot.com/emotigo/module/openapi-adapter/data"
	"emotibot.com/emotigo/module/openapi-adapter/traffic"
	"emotibot.com/emotigo/module/openapi-adapter/util"
	"emotibot.com/emotigo/pkg/logger"
)

var remoteHost string
var client = &http.Client{
	Timeout: time.Duration(2) * time.Second,
}
var proxy *httputil.ReverseProxy

var trafficManager *traffic.TrafficManager

func init() {
	envFile := flag.String("f", "", "environment variables file")
	flag.Parse()

	if *envFile != "" {
		err := util.LoadConfigFromFile(*envFile)
		if err != nil {
			logger.Error.Printf(err.Error())
			os.Exit(-1)
		}
	} else {
		err := util.LoadConfigFromOSEnv()
		if err != nil {
			logger.Error.Printf(err.Error())
			os.Exit(-1)
		}
	}
}

func main() {
	logLevel, ok := util.GetEnv("SERVER_LOG_LEVEL")
	if !ok {
		logLevel = "INFO"
	}
	logger.SetLevel(logLevel)
	logger.Info.Printf("Set log level %s\n", logLevel)

	port, ok := util.GetEnv("SERVER_PORT")
	if !ok {
		port = "8080"
	}

	remoteHost, ok = util.GetEnv("EC_HOST")
	if !ok {
		logger.Error.Println("please specify openapi v2 remote url by os Env OPENAPI_ADAPTER_EC_HOST")
		return
	}

	remoteHostURL, err := url.Parse(remoteHost)
	if err != nil {
		logger.Error.Printf("remoteHost is not a valid URL, %v\n", err)
		return
	}

	duration, err := util.GetIntEnv("DURATION")
	checkerr(err, "OPENAPI_ADAPTER_DURATION")
	maxRequests, err := util.GetIntEnv("MAXREQUESTS")
	checkerr(err, "OPENAPI_ADAPTER_MAXREQUESTS")
	banPeriod, err := util.GetIntEnv("BANPERIOD")
	checkerr(err, "OPENAPI_ADAPTER_BANPERIOD")
	logPeriod, err := util.GetIntEnv("LOGPERIOD")
	checkerr(err, "OPENAPI_ADAPTER_LOGPERIOD")

	logger.Info.Printf("Setting max %d request in %d seconds, "+
		"banned period: %d seconds, log period: %d seconds\n",
		maxRequests, duration, banPeriod, logPeriod)

	// Check statsd host
	monitorTraffic := true

	statsdHost, ok := util.GetEnv("STATSD_HOST")
	if !ok || statsdHost == "" {
		logger.Warn.Println("No STATSD_HOST")
		monitorTraffic = false
	}

	statsdPort, ok := util.GetEnv("STATSD_PORT")
	if !ok || statsdPort == "" {
		logger.Warn.Println("No STATSD_PORT")
		monitorTraffic = false
	}

	statsdPortInt, err := strconv.Atoi(statsdPort)
	if err != nil {
		logger.Warn.Printf("Invalidate STATSD_PORT: %s\n", statsdPort)
		monitorTraffic = false
	}

	// Make traffic channel
	trafficManager = traffic.NewTrafficManager(monitorTraffic, statsdHost, statsdPortInt,
		duration, int64(maxRequests), int64(banPeriod), int64(logPeriod))

	// Reserve proxy
	proxy = httputil.NewSingleHostReverseProxy(remoteHostURL)

	middleWares := chainMiddleWares(logSummarize)

	http.HandleFunc("/api/ApiKey/", middleWares(OpenAPIAdapterHandler))
	http.HandleFunc("/v1/openapi", middleWares(OpenAPIHandler))
	http.HandleFunc("/_health_check", HealthCheck)

	logger.Info.Printf("Starting server at port: %s\n", port)
	logger.Error.Fatalln(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}

// OpenAPIAdapterHandler will translate v1 request into v2 Request and then send to BFOP Server
// response will parse to v1 response format
func OpenAPIAdapterHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	err := extractHeadersFromBody(r)
	if err != nil {
		switch err {
		case data.ErrUnsupportedMethod:
			customError(w, err.Error(), http.StatusBadRequest)
		default:
			customError(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	ok, err := trafficManager.Monitor(w, r)
	if !ok {
		switch err {
		case data.ErrAppIDNotSpecified:
			customError(w, err.Error(), http.StatusBadRequest)
		case data.ErrUserIDNotSpecified:
			customError(w, err.Error(), http.StatusBadRequest)
		default:
			customError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	err = r.ParseForm()
	if err != nil {
		customError(w, "request body format error: "+err.Error(), http.StatusBadRequest)
		return
	}

	var (
		body     io.Reader
		bodyData []byte
		v2Req    *http.Request
		resp     *http.Response
		v1Resp   data.ResponseV1
		v2Resp   data.ResponseV2
	)

	text := r.FormValue("text")

	bodyData, err = json.Marshal(data.V2Body{
		Text: text,
	})
	if err != nil {
		customError(w, "request body format error: "+err.Error(), http.StatusBadRequest)
		return
	}

	body = bytes.NewBuffer(bodyData)

	remoteURL := fmt.Sprintf("%s/v1/openapi", remoteHost)
	v2Req, err = http.NewRequest(http.MethodPost, remoteURL, body)
	if err != nil {
		customError(w, "transform request failed, "+err.Error(), http.StatusInternalServerError)
		return
	}
	v2Req.Header.Set("appId", r.FormValue("appid"))
	v2Req.Header.Set("userId", r.FormValue("userid"))

	// Add headers for load balancing
	v2Req.Header.Set("X-Lb-Uid", r.Header.Get("X-Lb-Uid"))
	v2Req.Header.Set("X-Openapi-Appid", r.Header.Get("X-Openapi-Appid"))

	resp, err = client.Do(v2Req)
	if err != nil {
		customError(w, "http request failed, "+err.Error(), http.StatusInternalServerError)
		return
	}
	bodyData, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		customError(w, "io failed, "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	err = json.Unmarshal(bodyData, &v2Resp)
	logger.Info.Printf("v2 response body: %s\n", bodyData)
	if err != nil {
		customError(w, "unmarshal version 2 body failed, "+err.Error(), http.StatusInternalServerError)
		return
	}

	v1RespData := make([]data.DataV1, len(v2Resp.Answers))

	for i, answer := range v2Resp.Answers {
		var value string
		dataStr := ""
		index := 1

		for j, d := range answer.Data {
			if str, ok := d.(string); ok {
				dataStr += fmt.Sprintf("%d.%s", index, str)
				if j != len(answer.Data)-1 {
					dataStr += " "
				}
				index++
			}
		}

		if dataStr != "" {
			value = fmt.Sprintf("%s: %s", answer.Value, dataStr)
		} else {
			value = answer.Value
		}

		v1RespData[i] = data.DataV1{
			Type:  "text",
			Cmd:   "",
			Value: value,
			Data:  []data.Answer{answer},
		}
	}

	v1Resp.ReturnCode = v2Resp.Code
	v1Resp.Message = v2Resp.Message
	v1Resp.Data = v1RespData
	v1Resp.Emotion = []data.Emotion{
		newEmotion(v2Resp),
	}

	result, _ := json.Marshal(v1Resp)
	w.Write(result)
}

// OpenAPIHandler will do nothing but proxy the request/response
// to/from BFOP server directly
func OpenAPIHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID := r.Header.Get("userId")
	appID := r.Header.Get("appId")

	r.Header.Set("X-Lb-Uid", userID)
	r.Header.Set("X-Openapi-Appid", appID)

	ok, err := trafficManager.Monitor(w, r)
	if !ok {
		switch err {
		case data.ErrAppIDNotSpecified:
			resp, _ := json.Marshal(data.ErrorResponse{
				Message: err.Error(),
			})
			w.WriteHeader(http.StatusBadRequest)
			w.Write(resp)
		case data.ErrUserIDNotSpecified:
			resp, _ := json.Marshal(data.ErrorResponse{
				Message: err.Error(),
			})
			w.WriteHeader(http.StatusBadRequest)
			w.Write(resp)
		default:
			resp, _ := json.Marshal(data.ErrorResponse{
				Message: err.Error(),
			})
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(resp)
		}
		return
	}

	proxy.ServeHTTP(w, r)
}

// HealtCheck returns service health status
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

//MetaDataKey is the meta data defined which can be used after calling GetMetadata
type MetaDataKey string

const (
	UserIDKey     MetaDataKey = "userId"
	AppIDKey      MetaDataKey = "appId"
	OpenAPICmdKey MetaDataKey = "openAPICmd"
)

// GetMetadata retrive metadata from request
// It
func GetMetadata(r *http.Request) (map[MetaDataKey]string, error) {
	var metadata = make(map[MetaDataKey]string)

	buf, _ := ioutil.ReadAll(r.Body)

	rdr1 := ioutil.NopCloser(bytes.NewBuffer(buf))
	rdr2 := ioutil.NopCloser(bytes.NewBuffer(buf))

	r.Body = rdr1
	err := r.ParseForm()
	if err != nil {
		return nil, err
	}

	appid := ""
	userid := ""
	openapiCmd := ""

	if r.Method == "GET" || r.Method == "POST" {
		appid = r.FormValue("appid")
		openapiCmd = r.FormValue("cmd")
		// userid: OpenAPI
		userid = r.FormValue("userid")
		// UserID: /api/APP/chat.php  # FreemeOS
		userid += r.FormValue("UserID")
		// All other IDs:
		// phthon OpenID WeChatID wechatid user_id
		userid += r.FormValue("phthon")
		userid += r.FormValue("OpenID")
		userid += r.FormValue("WeChatID")
		userid += r.FormValue("wechatid")
		userid += r.FormValue("user_id")
	} else {
		// FIXME: Should we drop non GET/POST requests?
		logger.Warn.Printf("Unknown request type. %s %s %s\n", r.Host, r.Method, string(buf))
		return nil, data.ErrUnsupportedMethod
	}
	metadata[AppIDKey] = appid
	metadata[UserIDKey] = userid
	metadata[OpenAPICmdKey] = openapiCmd
	r.Body = rdr2
	return metadata, nil
}

// extractHeadersFromBody will extract the neccessary fields from request body and add to headers
// extractHeadersFromBody is only called by OpenAPI v1
// It will change the request itself, be care when using it.
func extractHeadersFromBody(r *http.Request) error {

	data, err := GetMetadata(r)
	if err != nil {
		return err
	}
	logger.Info.Printf("%+v\n", r)
	userid, _ := data[UserIDKey]
	appid, _ := data[AppIDKey]
	openapiCmd, _ := data[OpenAPICmdKey]

	r.Header.Set("X-Lb-Uid", userid)
	r.Header.Set("X-Openapi-Appid", appid)
	r.Header.Set("X-Openapi-Cmd", openapiCmd)

	r.Body = rdr2
	return nil
}

func checkerr(err error, who string) {
	if err != nil {
		logger.Error.Fatalf("No %s env variable, %v\n", who, err)
	}
}

func readList(r io.Reader) (map[string]bool, error) {
	bufData, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read reader failed, %v", err)
	}
	lists := strings.Split(string(bufData), "\n")
	list := make(map[string]bool, len(lists))
	for _, item := range lists {
		//Skip # comment
		if strings.HasPrefix(item, "#") {
			continue
		}
		list[item] = true
	}
	return list, nil
}

func customError(w http.ResponseWriter, message string, code int) {
	// Response need to be 200 OK all the time to mimic behavior in v1
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	var resp = data.ResponseV1{
		ReturnCode: code,
		Message:    message,
		Data:       []data.DataV1{},
		Emotion:    []data.Emotion{},
	}
	data, err := json.Marshal(resp)
	if err != nil {
		logger.Error.Printf("JSON marshal failed: %s\n", err.Error())
	}
	w.Write(data)
}

func newEmotion(resp data.ResponseV2) data.Emotion {
	var e = data.Emotion{}
	if resp.Info.EmotionCat != "" {
		e.Type = "text"
		e.Value = resp.Info.EmotionCat
		e.Score = strconv.Itoa(resp.Info.EmotionScore)
	}
	return e
}
