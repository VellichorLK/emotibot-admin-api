package trafficStats

import (
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/paulbellamy/ratecounter"
)

type RouteMap struct {
	GoRoute      map[string]string
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

//maxCon ?最大連接數？
var maxCon int64

var banPeriod int64

var newestRoute int64

// ReadTrafficChan UpdateTraffic函式的資料來源
var ReadTrafficChan chan string

// WriteRouteChan MakeNewRoute的資料來源
var WriteRouteChan chan *RouteMap
var UpdateRouteChan chan *RouteMap
var AppidChan chan *AppidIP

var MonitorAppid = map[string]bool{
	"c385e97b0cdce3bdbbee59083ec3b0d0": true, //ecovacs appid

}

func Init(dur int, max_con int, ban_period int64, logPeriod int, readTrafficChan chan string, writeRouteChan chan *RouteMap, appidChan chan *AppidIP, statsdURL string) {
	stats = make(map[string]*ratecounter.RateCounter)
	duration = dur
	maxCon = int64(max_con)
	banPeriod = ban_period

	ReadTrafficChan = readTrafficChan
	WriteRouteChan = writeRouteChan
	UpdateRouteChan = make(chan *RouteMap, 1)
	AppidChan = appidChan

	go UpdateTraffic()
	go PushRouteMap()
	go AppidCounter(logPeriod, statsdURL)
}

func DefaultRoute() *RouteMap {
	routeMap := new(RouteMap)
	routeMap.GoRoute = make(map[string]string)
	u, _ := url.Parse("http://172.17.0.1:9001")
	routeMap.DefaultRoute = u
	routeMap.timestamp = time.Now().UnixNano()
	newestRoute = routeMap.timestamp
	return routeMap
}

//PushRouteMap 把 UpadteRouteChan的route轉給WriteRouteChan
func PushRouteMap() {
	route := DefaultRoute()

	for {
		select {
		case route = <-UpdateRouteChan:
		case WriteRouteChan <- route:
		}
	}
}

//MakeNewRoute 從WriteRouteChan取得一個資料, 並產生新的RouteMap給新的userID
func MakeNewRoute(uid string) *RouteMap {
	curRoute := <-WriteRouteChan

	if curRoute.timestamp < newestRoute {
		return nil
	}

	_, ok := curRoute.GoRoute[uid]
	if ok {
		return nil
	}
	nr := DefaultRoute()
	//copy the map
	for k, v := range curRoute.GoRoute {
		nr.GoRoute[k] = v
	}
	log.Printf("Sets %s redirect\n", uid)
	log.Printf("%s has %d requests in %d seconds\n", uid, maxCon, duration)
	nr.GoRoute[uid] = "go"
	return nr
}

//MakeNewCounter create a ratecounter.RateCounter base on Duration and userID
func MakeNewCounter(uid string, dur int) *ratecounter.RateCounter {
	stat := ratecounter.NewRateCounter(time.Duration(dur) * time.Second)
	//FIXME: Not concurrency-safe write, should use sync.Map instead
	stats[uid] = stat
	return stat
}

//UpdateTraffic Read Traffic from channel and make counter for every user id
func UpdateTraffic() bool {

	var uid string
	var newRoute *RouteMap
	var lastTimestamp, curTimestamp int64

	lastTimestamp = time.Now().Unix()

	for {
		uid = <-ReadTrafficChan

		curTimestamp = time.Now().Unix()

		if curTimestamp-lastTimestamp >= banPeriod {
			newRoute = DefaultRoute()
			lastTimestamp = curTimestamp
		} else {

			stat, ok := stats[uid]

			if ok {

			} else {
				stat = MakeNewCounter(uid, duration)

			}
			stat.Incr(1)

			if stat.Rate() > maxCon {
				tmpRoute := MakeNewRoute(uid)
				if tmpRoute != nil {
					newRoute = tmpRoute
				}
				MakeNewCounter(uid, duration)
			}
		}

		if newRoute != nil {
			select {
			case UpdateRouteChan <- newRoute:
				newRoute = nil
			default:
			}

		}
	}
}

func AppidCounter(period int, statsdURL string) {
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
		case appip = <-AppidChan:
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
