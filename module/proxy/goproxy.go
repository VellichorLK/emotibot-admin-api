package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"os"
	"strconv"

	"strings"

	"emotibot.com/emotigo/module/proxy/traffic"
)

var AddTrafficChan chan string
var ReadDestChan chan *trafficStats.RouteMap
var AppidChan chan *trafficStats.AppidIP

func GoProxy(w http.ResponseWriter, r *http.Request) {

	buf, _ := ioutil.ReadAll(r.Body)

	rdr1 := ioutil.NopCloser(bytes.NewBuffer(buf))
	rdr2 := ioutil.NopCloser(bytes.NewBuffer(buf))

	r.Body = rdr1
	r.ParseMultipartForm(0)

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
		log.Printf("Warning: Unknown request type. %s %s %s", r.Host, r.Method, string(buf))
		http.Error(w, "Method Error", http.StatusBadGateway)
		return
	}

	ipPort := strings.Split(r.RemoteAddr, ":")
	if len(ipPort) == 2 {
		appidIP := new(trafficStats.AppidIP)
		appidIP.Appid = appid
		appidIP.IP = ipPort[0]
		appidIP.Userid = userid
		AppidChan <- appidIP
	} else {
		log.Printf("Warning: ip:port not fit. %s\n", r.RemoteAddr)
	}

	var route *trafficStats.RouteMap

	AddTrafficChan <- userid
	route = <-ReadDestChan

	_, ok := route.GoRoute[userid]
	url := route.DefaultRoute

	if ok {
		userid = userid + strconv.Itoa(rand.Intn(1000))
	}

	r.Header.Set("X-Lb-Uid", userid)
	r.Header.Set("X-Openapi-Appid", appid)
	r.Header.Set("X-Openapi-Cmd", openapiCmd)

	r.Body = rdr2
	proxy := httputil.NewSingleHostReverseProxy(url)
	proxy.ServeHTTP(w, r)

}

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
}

func checkerr(err error, who string) {
	if err != nil {
		log.Println("No ", who, " enviroments variable!", err)
		panic(1)
	}
}

func main() {

	duration, err := strconv.Atoi(os.Getenv("DURATION"))
	checkerr(err, "DURATION")
	maxLimit, err := strconv.Atoi(os.Getenv("MAXREQUESTS"))
	checkerr(err, "MAXREQUESTS")
	banPeriod, err := strconv.Atoi(os.Getenv("BANPERIOD"))
	checkerr(err, "BANPERIOD")
	logPeriod, err := strconv.Atoi(os.Getenv("LOGPERIOD"))
	checkerr(err, "LOGPERIOD")
	statsdHost := os.Getenv("STATSDHOST")
	if statsdHost == "" {
		log.Println("No STATSDHOST")
		panic(1)
	}
	statsdPort := os.Getenv("STATSDPORT")
	if statsdPort == "" {
		log.Println("No STATSDPORT")
		panic(1)
	}

	log.Printf("Setting max %d request in %d seconds, banned period %d, log period:%d\n", maxLimit, duration, banPeriod, logPeriod)

	//make the channel
	AddTrafficChan = make(chan string)
	ReadDestChan = make(chan *trafficStats.RouteMap)
	AppidChan = make(chan *trafficStats.AppidIP, 1024)
	trafficStats.Init(duration, maxLimit, int64(banPeriod), logPeriod, AddTrafficChan, ReadDestChan, AppidChan, statsdHost+":"+statsdPort)

	http.HandleFunc("/", GoProxy)
	http.HandleFunc("/_health_check", HealthCheck)
	log.Fatal(http.ListenAndServe(":9000", nil))

}
