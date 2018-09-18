package traffic

import (
	"fmt"
	"log"
	"net/url"
	"sync"
	"time"

	"emotibot.com/emotigo/module/openapi-adapter/statsd"
	"emotibot.com/emotigo/module/openapi-adapter/util"
	"github.com/paulbellamy/ratecounter"
)

type RouteMap struct {
	GoRoute      map[string]bool
	DefaultRoute *url.URL
	timestamp    int64
}

type AppidIP struct {
	Appid      string
	IP         string
	RequestURI string
	Userid     string
}

type AppidCount struct {
	FromWho  map[string]uint64 //ip -> count
	FromUser map[string]uint64 //userid -> count
	Counter  uint64            //counter in assigned period
}

type RequestCount struct {
	URI     string
	Counter uint64
}

// TrafficManager is a manager to determine which user is overload.
type TrafficManager struct {
	Stats          *sync.Map
	BufferDuration time.Duration
	maxConnection  int64
	banPeriod      int64
}

var statsdNamespace = "emotibot.grproxy"

// NewTrafficManager init a new TrafficManager by setting
// duration RateCounter計算的時間長度
func NewTrafficManager(duration int, maxConnection, banPeriod int64) *TrafficManager {
	return &TrafficManager{
		Stats:          new(sync.Map),
		BufferDuration: time.Duration(duration),
		maxConnection:  maxConnection,
		banPeriod:      banPeriod,
	}
}

// CheckOverFlowed 檢查該使用者是否已經超過流量
func (m *TrafficManager) CheckOverFlowed(uid string) bool {
	stat, _ := m.Stats.LoadOrStore(uid, ratecounter.NewRateCounter(m.BufferDuration*time.Second))
	counter := stat.(*ratecounter.RateCounter)
	// FIXME: Counter must be increased and evaluated at the same time (or using mutex),
	// otherwise, there might have some out of sync issue.
	if counter.Incr(1); counter.Rate() > m.maxConnection {
		return true
	}
	return false
}

// AppidCounter 計算 AppID 統計資訊
func AppidCounter(appipChan <-chan *AppidIP, period int, statsdHost string, statsdPort int) {
	var appip *AppidIP

	c := util.GetStatsdClient(statsdHost, statsdPort)

	flowCount := make(map[string]*AppidCount)
	requestCount := make(map[string]*RequestCount)
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

			// Here we don't count how many times the ip or useid used this request
			// so simply set it to zero
			ac.FromUser[appip.Userid] = 0
			ac.FromWho[appip.IP] = 0

			ac.Counter++

			metric := fmt.Sprintf("%s.%s.%s", statsdNamespace, appip.Appid, "request.count")
			c.IncrementCounter(metric)

			req, ok := requestCount[appip.Appid]
			if !ok {
				req = new(RequestCount)
				req.URI = appip.RequestURI
				requestCount[appip.Appid] = req
			}

			req.Counter++
		case <-timeCh:
			goStatsd(c, flowCount, requestCount)
			timeCh = time.After(time.Duration(period) * time.Second)
		}
	}
}

func goStatsd(c *statsd.Client, flowCount map[string]*AppidCount,
	requestCount map[string]*RequestCount) {
	var metric string

	for appid, v := range flowCount {
		numOfIP := len(v.FromWho)
		v.Counter = 0

		metric = fmt.Sprintf("%s.%s.%s", statsdNamespace, appid, "num.source")
		c.IncrementGaugeByValue(metric, numOfIP)

		if numOfIP > 0 {
			log.Printf("[%s] has below ip (%d):\n", appid, numOfIP)
			for ip := range v.FromWho {
				log.Printf("%s\n", ip)
			}
			log.Printf("-----------------------------------------\n")
		}

		numOfUserID := len(v.FromUser)

		metric = fmt.Sprintf("%s.%s.%s", statsdNamespace, appid, "num.userid")
		c.IncrementGaugeByValue(metric, numOfUserID)
	}

	for appid, req := range requestCount {
		metric = fmt.Sprintf("%s.%s.%s.%s", statsdNamespace, appid, req.URI, "num.uri.request")
		c.IncrementGaugeByValue(metric, int(req.Counter))

		// Reset the counter
		req.Counter = 0
	}
}
