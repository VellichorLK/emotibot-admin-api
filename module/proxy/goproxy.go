package main
import (
	"net/http"
	"net/http/httputil"
	"io/ioutil"
	"math/rand"
	"bytes"
	"emotibot.com/emotigo/module/proxy/traffic"
	"log"
	"flag"
)


var force_dispatch_apppid = map[string]bool{
	"3239a33d48e057e7470a3a5f810e67ce": true, // botname = xinxx
	"1fd273be896ec71d63c15c770225c079": true, // Trencent.com's appid
}

var AddTrafficChan chan string
var ReadDestChan chan *trafficStats.RouteMap


func GoProxy(w http.ResponseWriter, r *http.Request){

	buf,_ := ioutil.ReadAll(r.Body)

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
		log.Printf("Warning: Unknown request type.", r.Host, r.Method, string(buf))
		http.Error(w, "Method Error", http.StatusBadGateway)
		return
	}

	if force_dispatch_apppid[appid] {
		r.Header.Del("X-Lb-Uid")
		log.Println("  Force Dispatch (appid): ", appid)
	}

	r.Header.Set("X-Lb-Uid", userid)
	r.Header.Set("X-Openapi-Appid", appid)
	r.Header.Set("X-Openapi-Cmd", openapiCmd)


	var route *trafficStats.RouteMap

	AddTrafficChan<- userid
	route = <-ReadDestChan

	url,ok:=route.GoRoute[userid]

	if !ok{
		url = route.DefaultRoute[rand.Int()%len(route.DefaultRoute)]
	}

	r.Body = rdr2
	proxy:=httputil.NewSingleHostReverseProxy(url)
	proxy.ServeHTTP(w,r)
	

}

/*

func Reroute(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	})

}
*/

func HealthCheck(w http.ResponseWriter, r *http.Request){
	w.WriteHeader(200)
}

func main(){

	//parse the flag
	duration:= flag.Int("d", 60, "duration of traffic statics (second)")
	maxLimit:= flag.Int("n", 200, "max times to request in the duration")
	flag.Parse()
	log.Printf("Setting max %d request in %d seconds", *maxLimit, *duration)

	//make the channel
	AddTrafficChan = make(chan string)
	ReadDestChan = make(chan *trafficStats.RouteMap)

	trafficStats.Init(*duration,*maxLimit, AddTrafficChan, ReadDestChan)


/*
	mux := http.NewServeMux()
	mux.HandleFunc("/",GoProxy)
	muxHandler := Reroute(mux)
	http.ListenAndServe(":5000",muxHandler)
	*/

	http.HandleFunc("/",GoProxy)
	http.HandleFunc("/healthcheck", HealthCheck)
	log.Fatal(http.ListenAndServe(":9000",nil))

}





