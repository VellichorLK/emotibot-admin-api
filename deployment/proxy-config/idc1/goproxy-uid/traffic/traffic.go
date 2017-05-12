package trafficStats

import (
	"time"
	"github.com/paulbellamy/ratecounter"
	"net/url"
	"log"
)



type RouteMap struct{
	GoRoute map[string] string
	DefaultRoute *url.URL
}


var stats map[string] *ratecounter.RateCounter
var duration int
var maxCon int64


var ReadTrafficChan chan string
var WriteRouteChan chan *RouteMap
var UpdateRouteChan chan *RouteMap

func Init(dur int, max_con int, readTrafficChan chan string, writeRouteChan chan *RouteMap){
	stats = make(map[string] *ratecounter.RateCounter)
	duration = dur
	maxCon = int64(max_con)

	ReadTrafficChan = readTrafficChan
	WriteRouteChan = writeRouteChan
	UpdateRouteChan = make(chan *RouteMap, 1)

	go UpdateTraffic()
	go PushRouteMap()
}


func DefaultRoute() *RouteMap{
	routeMap := new(RouteMap)
	routeMap.GoRoute = make(map[string] string)
	u,_ := url.Parse("http://172.17.0.1:9001")
	routeMap.DefaultRoute = u
	return routeMap
}

func PushRouteMap(){
	route := DefaultRoute()	

	for {
		select {
			case WriteRouteChan <- route:
			case route = <-UpdateRouteChan:
		}
	}
}


func MakeNewRoute(uid string) *RouteMap{
        curRoute:= <-WriteRouteChan
        _,ok:= curRoute.GoRoute[uid]
        if ok{
                return nil
        }
        nr:=DefaultRoute()
                //copy the map
	for k,v := range curRoute.GoRoute{
		nr.GoRoute[k] = v
	}
	log.Println("Sets", uid, "redirect")
	log.Println(uid, " has ", maxCon, " requests in ", duration, " seconds")
	nr.GoRoute[uid] = "go"
	return nr
/*
	curRoute:= <-WriteRouteChan
	_,ok:= curRoute.GoRoute[uid]
	if ok{
		return nil
	}
	nr:=DefaultRoute()
	u,err := url.Parse("http://127.0.0.1:4000")
	if err == nil{
		//copy the map
		for k,v := range curRoute.GoRoute{
			nr.GoRoute[k] = v
		}
		log.Println("Sets", uid, "redirect to ", u)
		log.Println(uid, " has ", maxCon, " requests in ", duration, " seconds")
		nr.GoRoute[uid] = u
	}else{
		return nil
	}
	return nr
	*/
}

func MakeNewCounter(uid string, dur int) *ratecounter.RateCounter{
	stat:= ratecounter.NewRateCounter(time.Duration(dur) * time.Second)
	stats[uid] = stat
	return stat
}

func UpdateTraffic() bool{

	var uid string
	var newRoute *RouteMap = nil

	for{
		uid = <-ReadTrafficChan

		stat, ok:= stats[uid]

		if ok {

		}else{
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

		if newRoute != nil {
			select{
				case UpdateRouteChan<-newRoute:
					newRoute = nil
				default:
			}

		}
	}
}

