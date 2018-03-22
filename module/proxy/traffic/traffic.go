package trafficStats

import (
	"fmt"
	"log"
	"net/url"
	"sync"
	"time"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/paulbellamy/ratecounter"
)

//RouteMap is a singleton struct represnting every blacklist users and the time of this list created.
type RouteMap struct {
	GoRoute      map[string]bool
	DefaultRoute *url.URL
	timestamp    int64
}

type AppidIP struct {
	Appid  string
	IP     string
	Userid string
}

type AppidCount struct {
	FromWho  map[string]uint64 //ip -> count
	FromUser map[string]uint64 //userid -> count
	Counter  uint64            //counter in assigned period
}

var stats map[string]*ratecounter.RateCounter

//duration RateCounter計算的時間長度
var duration int

//maxCon 代表counter生成時要以多少為最大強制分流值
var maxCon int64

var banPeriod int64

var newestRoute int64

var MonitorAppid = map[string]bool{
	"c385e97b0cdce3bdbbee59083ec3b0d0": true, //ecovacs appid

}

//TrafficManager is a manager to determine which user is overload.
type TrafficManager struct {
	Stats          *sync.Map
	BufferDuration time.Duration
	Route          url.URL
	maxConnection  int64
	banPeriod      int64
}

//LogManager DO NOT USE IT! not implement yet.
type LogManager struct {
	LogPeriod    int64
	AgentAddress string
	AppidChan    chan *AppidIP
}

//NewTrafficManager init a new TrafficManager by setting
func NewTrafficManager(duration int, maxConnection, banPeriod int64, route url.URL) *TrafficManager {
	return &TrafficManager{
		Stats:          new(sync.Map),
		BufferDuration: time.Duration(duration),
		Route:          route,
		maxConnection:  maxConnection,
		banPeriod:      banPeriod,
	}
}

//NewLogManager init a new LogManager with parameters setting
func NewLogManager(logPeriod int64, channel chan *AppidIP) *LogManager {
	return &LogManager{
		LogPeriod: logPeriod,
		AppidChan: channel,
	}
}

//DefaultRoutingURL 定義了導向的LB網址
var DefaultRoutingURL = "http://172.17.0.1:9001"

//CheckOverFlowed 檢查該使用者是否已經超過流量
func (m *TrafficManager) CheckOverFlowed(uid string) bool {
	stat, _ := m.Stats.LoadOrStore(uid, ratecounter.NewRateCounter(m.BufferDuration*time.Second))
	counter := stat.(*ratecounter.RateCounter)
	//FIXME: counter must be incre and evalue at the same line(or use mutext) or there might have some out of sync issue.
	if counter.Incr(1); counter.Rate() > m.maxConnection {
		return true
	}
	return false
}

//計算datadog資訊
func AppidCounter(appipChan <-chan *AppidIP, period int, statsdURL string) {
	var appip *AppidIP

	c, err := statsd.New(statsdURL)
	if err != nil {
		log.Fatal(err)
	}
	c.Namespace = "emotibot."
	c.Tags = append(c.Tags, "module:goproxy")

	flowCount := make(map[string]*AppidCount)
	timeCh := time.After(time.Duration(period) * time.Second)
	for {
		select {
		case appip = <-appipChan:
			ac, ok := flowCount[appip.Appid]
			if !ok {
				ac = new(AppidCount)
				ac.FromWho = make(map[string]uint64)
				ac.FromUser = make(map[string]uint64)
				flowCount[appip.Appid] = ac
			}

			//here we don't count how many times the ip or useid used this request
			//so simply set it to zero
			ac.FromUser[appip.Userid] = 0
			ac.FromWho[appip.IP] = 0

			ac.Counter++

			err := c.Incr("request.count", []string{"appid:" + appip.Appid}, 1)
			if err != nil {
				log.Println(err)
			}

		case <-timeCh:
			goDatadog(c, flowCount)
			timeCh = time.After(time.Duration(period) * time.Second)
		}

	}
}

func goDatadog(c *statsd.Client, flowCount map[string]*AppidCount) {
	for appid, v := range flowCount {
		numOfIP := len(v.FromWho)
		v.Counter = 0
		err := c.Gauge("num.source", float64(numOfIP), []string{"appid:" + appid}, 1)
		if err != nil {
			log.Println(err)
		}
		if numOfIP > 0 {
			fmt.Printf("[%s] has below ip (%d):\n", appid, numOfIP)
			for ip := range v.FromWho {
				fmt.Printf("%s\n", ip)
			}
			fmt.Printf("-----------------------------------------\n")
		}

		numOfUserID := len(v.FromUser)
		err = c.Gauge("num.userid", float64(numOfUserID), []string{"appid:" + appid}, 1)
		if err != nil {
			log.Println(err)
		}
	}

}
