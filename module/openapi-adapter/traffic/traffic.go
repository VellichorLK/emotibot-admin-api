package traffic

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
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
	UserID     string
	IP         string
	RequestURI string
}

type AppIDCount struct {
	FromWho  map[string]uint64 //ip -> count
	FromUser map[string]uint64 //userid -> count
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
}

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
	}

	go m.trafficStats()

	return &m
}

func (m *TrafficManager) Monitor(w http.ResponseWriter, r *http.Request) (bool, error) {
	appID := r.Header.Get("X-Openapi-Appid")
	userID := r.Header.Get("X-Lb-Uid")

	if appID == "" {
		return false, data.ErrAppIDNotSpecified
	}

	if userID == "" {
		return false, data.ErrUserIDNotSpecified
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
	return true, nil
}

func (m *TrafficManager) Summarize(appID string, w *data.ResponseLogger, r *http.Request, responseTime float64) {
	var metric string
	re := regexp.MustCompile("\\.")
	requestURI := re.ReplaceAllString(r.RequestURI, "-")

	// Request count
	metric = fmt.Sprintf("%s.%s.%s.%s", data.StatsdNamespace, appID, requestURI, "request.count")
	m.StatsdClient.IncrementCounter(metric)

	// Response time
	metric = fmt.Sprintf("%s.%s.%s.%s", data.StatsdNamespace, appID, requestURI, "response.time")
	m.StatsdClient.Timing(metric, int64(responseTime))

	// Response status code
	switch {
	case w.StatusCode >= 200 && w.StatusCode < 300:
		metric = fmt.Sprintf("%s.%s.%s.%s", data.StatsdNamespace, appID, requestURI, "response.2xx")
		m.StatsdClient.IncrementCounter(metric)
	case w.StatusCode >= 400 && w.StatusCode < 500:
		metric = fmt.Sprintf("%s.%s.%s.%s", data.StatsdNamespace, appID, requestURI, "response.4xx")
		m.StatsdClient.IncrementCounter(metric)
	case w.StatusCode >= 500 && w.StatusCode < 600:
		metric = fmt.Sprintf("%s.%s.%s.%s", data.StatsdNamespace, appID, requestURI, "response.5xx")
		m.StatsdClient.IncrementCounter(metric)
	}
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
	flowCountLock := &sync.Mutex{}
	timeCh := time.After(time.Duration(m.LogPeriod) * time.Second)

	for {
		select {
		case trafficStat = <-m.TrafficStatsChan:
			flowCountLock.Lock()
			ac, ok := flowCount[trafficStat.AppID]
			if !ok {
				ac = new(AppIDCount)
				ac.FromWho = make(map[string]uint64)
				ac.FromUser = make(map[string]uint64)
				flowCount[trafficStat.AppID] = ac
			}

			// We count only the number of unique source IPs and users within log period,
			// not the exact value of the counters, so simply set them with zeros
			ac.FromUser[trafficStat.UserID] = 0
			ac.FromWho[trafficStat.IP] = 0
			flowCountLock.Unlock()
		case <-timeCh:
			goStatsd(m.StatsdClient, flowCount)

			// Reset flow counter
			flowCountLock.Lock()
			flowCount = make(map[string]*AppIDCount)
			flowCountLock.Unlock()

			timeCh = time.After(time.Duration(m.LogPeriod) * time.Second)
		}
	}
}

func goStatsd(c *statsd.Client, flowCount map[string]*AppIDCount) {
	if c == nil {
		return
	}

	var metric string

	for appID, v := range flowCount {
		// Number of unique source IPs within log period
		numOfIP := len(v.FromWho)
		metric = fmt.Sprintf("%s.%s.%s", data.StatsdNamespace, appID, "num.source")
		c.IncrementCounterByValue(metric, numOfIP)

		if numOfIP > 0 {
			log.Printf("[%s] has following unique IPs (%d):\n", appID, numOfIP)
			for ip := range v.FromWho {
				log.Printf("%s\n", ip)
			}
			log.Printf("-----------------------------------------\n")
		}

		// Number of unique users within log period
		numOfUserID := len(v.FromUser)
		metric = fmt.Sprintf("%s.%s.%s", data.StatsdNamespace, appID, "num.userid")
		c.IncrementCounterByValue(metric, numOfUserID)

		if numOfUserID > 0 {
			log.Printf("[%s] has following unique users (%d):\n", appID, numOfUserID)
			for userID := range v.FromUser {
				log.Printf("%s\n", userID)
			}
			log.Printf("-----------------------------------------\n")
		}
	}
}
