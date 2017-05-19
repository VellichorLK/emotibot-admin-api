package trafficStats

import (
	"log"
	"net/url"
	"time"

	"github.com/paulbellamy/ratecounter"
)

type RouteMap struct {
	GoRoute      map[string]string
	DefaultRoute *url.URL
	timestamp    int64
}

var stats map[string]*ratecounter.RateCounter
var duration int
var maxCon int64
var banPeriod int64
var newestRoute int64

var ReadTrafficChan chan string
var WriteRouteChan chan *RouteMap
var UpdateRouteChan chan *RouteMap

func Init(dur int, max_con int, ban_period int64, readTrafficChan chan string, writeRouteChan chan *RouteMap) {
	stats = make(map[string]*ratecounter.RateCounter)
	duration = dur
	maxCon = int64(max_con)
	banPeriod = ban_period

	ReadTrafficChan = readTrafficChan
	WriteRouteChan = writeRouteChan
	UpdateRouteChan = make(chan *RouteMap, 1)

	go UpdateTraffic()
	go PushRouteMap()
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
	var newRoute *RouteMap = nil
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
