package trafficStats

import (
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
	Appid string
	IP    string
}

type AppidCount struct {
	FromWho map[string]uint64 //ip->count
	Counter uint64            //counter in assigned period
}

var stats map[string]*ratecounter.RateCounter
var duration int
var maxCon int64
var banPeriod int64
var newestRoute int64

var ReadTrafficChan chan string
var WriteRouteChan chan *RouteMap
var UpdateRouteChan chan *RouteMap
var AppidChan chan *AppidIP

var monitorAppid = map[string]bool{
	"c385e97b0cdce3bdbbee59083ec3b0d0": true,
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

func PushRouteMap() {
	route := DefaultRoute()

	for {
		select {
		case route = <-UpdateRouteChan:
		case WriteRouteChan <- route:
		}
	}
}

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
	log.Println("Sets", uid, "redirect")
	log.Println(uid, " has ", maxCon, " requests in ", duration, " seconds")
	nr.GoRoute[uid] = "go"
	return nr
}

func MakeNewCounter(uid string, dur int) *ratecounter.RateCounter {
	stat := ratecounter.NewRateCounter(time.Duration(dur) * time.Second)
	stats[uid] = stat
	return stat
}

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
			do, ok := monitorAppid[appip.Appid]
			if do && ok {
				ac, ok := flowCount[appip.Appid]
				if !ok {
					ac = new(AppidCount)
					ac.FromWho = make(map[string]uint64)
					flowCount[appip.Appid] = ac
				}

				ac.FromWho[appip.IP] = 0
				ac.Counter++

				err := c.Incr("request.count", []string{appip.Appid}, 1)
				if err != nil {
					log.Println(err)
				}

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
		err := c.Gauge("num.source", float64(numOfIP), []string{appid}, 1)
		if err != nil {
			log.Println(err)
		}
	}
}
