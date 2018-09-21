package traffic

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	"emotibot.com/emotigo/module/openapi-adapter/data"
	"emotibot.com/emotigo/module/openapi-adapter/statsd"
	"emotibot.com/emotigo/module/openapi-adapter/util"
	"emotibot.com/emotigo/pkg/logger"
	"github.com/paulbellamy/ratecounter"
)

type RouteMap struct {
	GoRoute      map[string]bool
	DefaultRoute *url.URL
	timestamp    int64
}

type TrafficStat struct {
	AppID      string
	IP         string
	RequestURI string
	UserID     string
}

type AppIDCount struct {
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
	MonitorTraffic   bool
	StatsdClient     *statsd.Client
	Stats            *sync.Map
	BufferDuration   time.Duration
	MaxConnection    int64
	BanPeriod        int64
	LogPeriod        int64
	TrafficStatsChan chan *TrafficStat
	AppIDCountChan   chan *AppIDCount
}

const statsdNamespace = "emotibot.goproxy"

// NewTrafficManager init a new TrafficManager by setting
// duration RateCounter計算的時間長度
func NewTrafficManager(monitorTraffic bool, statsdHost string, statsdPort int,
	duration int, maxConnection int64, banPeriod int64, logPeriod int64) *TrafficManager {

	var client *statsd.Client

	if monitorTraffic {
		client = statsd.New(statsdHost, statsdPort)
	}

	m := TrafficManager{
		MonitorTraffic:   monitorTraffic,
		StatsdClient:     client,
		Stats:            new(sync.Map),
		BufferDuration:   time.Duration(duration),
		MaxConnection:    maxConnection,
		BanPeriod:        banPeriod,
		LogPeriod:        logPeriod,
		TrafficStatsChan: make(chan *TrafficStat),
		AppIDCountChan:   make(chan *AppIDCount),
	}

	go m.trafficStats()

	return &m
}

func (m *TrafficManager) Monitor(w http.ResponseWriter, r *http.Request) bool {
	userID := r.Header.Get("X-Lb-Uid")
	appID := r.Header.Get("X-Openapi-Appid")

	if userID == "" {
		resp, err := json.Marshal(data.ErrorResponse{
			Message: "User ID not specified",
		})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return false
		}

		w.WriteHeader(http.StatusBadRequest)
		w.Write(resp)
		return false
	}

	if appID == "" {
		resp, err := json.Marshal(data.ErrorResponse{
			Message: "App ID not specified",
		})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return false
		}

		w.WriteHeader(http.StatusBadRequest)
		w.Write(resp)
		return false
	}

	logger.Info.Printf("Header [X-Lb-Uid]: %s\n", r.Header.Get("X-Lb-Uid"))
	logger.Info.Printf("Header [X-Openapi-Appid]: %s\n", r.Header.Get("X-Openapi-Appid"))

	if m.checkOverFlowed(userID) {
		// TODO: Reject the connection
		logger.Warn.Printf("User ID:%s is overflowed\n", userID)
		//w.WriteHeader(http.StatusTooManyRequests)
		//return false
	}

	userIP, err := util.ParseRemoteIP(r.RemoteAddr)
	if err != nil {
		logger.Warn.Printf("Warning: ip:port not fit. %s\n", r.RemoteAddr)
	}

	trafficStat := TrafficStat{
		AppID:      appID,
		IP:         userIP,
		UserID:     userID,
		RequestURI: r.RequestURI,
	}

	m.TrafficStatsChan <- &trafficStat
	return true
}

// checkOverFlowed 檢查該使用者是否已經超過流量
func (m *TrafficManager) checkOverFlowed(uid string) bool {
	stat, _ := m.Stats.LoadOrStore(uid, ratecounter.NewRateCounter(m.BufferDuration*time.Second))
	counter := stat.(*ratecounter.RateCounter)
	// FIXME: Counter must be increased and evaluated at the same time (or using mutex),
	// otherwise, there might have some out of sync issue.
	if counter.Incr(1); counter.Rate() > m.MaxConnection {
		return true
	}
	return false
}

// trafficStats 計算流量相關統計資訊
func (m *TrafficManager) trafficStats() {
	var trafficStat *TrafficStat

	flowCount := make(map[string]*AppIDCount)
	requestCount := make(map[string]*RequestCount)
	timeCh := time.After(time.Duration(m.LogPeriod) * time.Second)

	for {
		select {
		case trafficStat = <-m.TrafficStatsChan:
			ac, ok := flowCount[trafficStat.AppID]
			if !ok {
				ac = new(AppIDCount)
				ac.FromWho = make(map[string]uint64)
				ac.FromUser = make(map[string]uint64)
				flowCount[trafficStat.AppID] = ac
			}

			// Here we don't count how many times the ip or useid used this request
			// so simply set it to zero
			ac.FromUser[trafficStat.UserID] = 0
			ac.FromWho[trafficStat.IP] = 0

			ac.Counter++

			metric := fmt.Sprintf("%s.%s.%s", statsdNamespace, trafficStat.AppID, "request.count")
			m.StatsdClient.IncrementCounter(metric)

			req, ok := requestCount[trafficStat.AppID]
			if !ok {
				req = new(RequestCount)
				req.URI = trafficStat.RequestURI
				requestCount[trafficStat.AppID] = req
			}

			req.Counter++
		case <-timeCh:
			goStatsd(m.StatsdClient, flowCount, requestCount)
			timeCh = time.After(time.Duration(m.LogPeriod) * time.Second)
		}
	}
}

func goStatsd(c *statsd.Client, flowCount map[string]*AppIDCount,
	requestCount map[string]*RequestCount) {

	if c == nil {
		return
	}

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
