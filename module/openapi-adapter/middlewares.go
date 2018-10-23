package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"emotibot.com/emotigo/module/openapi-adapter/data"
	"emotibot.com/emotigo/module/openapi-adapter/traffic"
	"emotibot.com/emotigo/pkg/logger"
	"github.com/paulbellamy/ratecounter"
)

type middleware func(next http.HandlerFunc) http.HandlerFunc

func chainMiddleWares(mws ...middleware) middleware {
	return func(final http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Start chaining middlewares by wrapping final handler and middlewares
			// in the reverse order where middlewares passed in
			// When executing the last returned handler, the middleware will be executed
			// in the order where they were passed in and then the final handler at last
			handler := final
			for i := len(mws) - 1; i >= 0; i-- {
				handler = mws[i](handler)
			}

			handler(w, r)
		}
	}
}

func logSummarize(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		responseLogger := data.NewResponseLogger(w)
		startTime := time.Now()
		next.ServeHTTP(responseLogger, r)
		responseTime := time.Since(startTime).Seconds() * 1000

		logger.Info.Printf("%s reponse time: %.2f ms", r.RequestURI, responseTime)

		appID := r.Header.Get("X-Openapi-Appid")
		if appID == "" {
			return
		}
		//Because response can be delayed until function finished, it should be forked out as a go routine
		trafficManager.Summarize(appID, responseLogger, r, responseTime)
	}
}

//NewInputValidatMIddleware will check if
func NewMetadataValidateMiddleware() middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			metadata, err := GetMetadata(r)
			if err != nil {
				customError(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if metadata[AppIDKey] == "" {
				customError(w, data.ErrAppIDNotSpecified.Error(), http.StatusBadRequest)
				return
			} else if metadata[UserIDKey] == "" {
				customError(w, data.ErrUserIDNotSpecified.Error(), http.StatusBadRequest)
				return
			}
			next.ServeHTTP(w, r)
		}
	}
}

func NewDailyLimitMiddleWare(globalApps map[string]int64, maximum int64, lock *sync.Mutex) middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			metadata, err := GetMetadata(r)
			appID, _ := metadata[AppIDKey]
			if err != nil {
				resp, _ := json.Marshal(data.ErrorResponse{
					Message: fmt.Sprintf("get meta data failed, %v", err),
				})
				w.WriteHeader(http.StatusBadRequest)
				w.Write(resp)
				return
			} else if appID == "" {
				resp, _ := json.Marshal(data.ErrorResponse{
					Message: data.ErrAppIDNotSpecified.Error(),
				})
				w.WriteHeader(http.StatusBadRequest)
				w.Write(resp)
				return
			}

			lock.Lock()
			count, _ := globalApps[appID]
			globalApps[appID] = count + 1
			lock.Unlock()
			if count+1 > maximum {
				r.Header.Set("X-Filtered", "true")
			}
			next.ServeHTTP(w, r)
		}
	}
}

// NewQueryThresholdMiddleware create a threshold middleware for http.HandlerFunc
// It used manager
func NewQueryThresholdMiddleware(manager *traffic.TrafficManager) middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			meta, err := GetMetadata(r)
			appID, _ := meta[AppIDKey]
			if err != nil {
				resp, _ := json.Marshal(data.ErrorResponse{
					Message: fmt.Sprintf("get meta data failed, %v", err),
				})
				w.WriteHeader(http.StatusBadRequest)
				w.Write(resp)
				return
			} else if appID == "" {
				resp, _ := json.Marshal(data.ErrorResponse{
					Message: data.ErrAppIDNotSpecified.Error(),
				})
				w.WriteHeader(http.StatusBadRequest)
				w.Write(resp)
				return
			}
			if manager.Count(appID) {
				r.Header.Set("X-Filtered", "true")
			}
			next.ServeHTTP(w, r)

		}
	}
}

// Filter check the request's identifier is valid or not and return a bool.
// If this request should be filter out, return true.
// If not, return false.
type Filter func(identifier string) bool

// newAppIDLimitMiddleware create a traffic limiter middleware based on filter.
// limitMiddleware controll how many request should pass this middleware by filter.
// If request appID can't be found or error raise, it will return a data.ErrAppIDNotSpecified error.
// It use the request's appID as identifier. and determine the flow by fallowing state.
// If filter return true, the request will be dispatch to the retryHandler and terminate processing.
// If filter return false, r will process to next http.HandlerFunc.
func newAppIDLimitMiddleware(retryHandler http.HandlerFunc, filter Filter) middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			metadata, err := GetMetadata(r)
			appID, _ := metadata[AppIDKey]
			if err != nil {
				resp, _ := json.Marshal(data.ErrorResponse{
					Message: fmt.Sprintf("get meta data failed, %v", err),
				})
				w.WriteHeader(http.StatusBadRequest)
				w.Write(resp)
				return
			} else if appID == "" {
				resp, _ := json.Marshal(data.ErrorResponse{
					Message: data.ErrAppIDNotSpecified.Error(),
				})
				w.WriteHeader(http.StatusBadRequest)
				w.Write(resp)
				return
			}

			if filter(appID) {
				retryHandler(w, r)
				return
			}
			next.ServeHTTP(w, r)
		}
	}
}

type filterConfig struct {
	Apps    map[string]appConfig
	Default appConfig
}
type appConfig struct {
	DailyLimit int64
	QPSLimit   int64
}

func createFilterConfig(data []byte) (*filterConfig, error) {
	var config = &filterConfig{
		Apps: map[string]appConfig{},
	}
	lines := strings.Split(string(data), "\n")
	for i, l := range lines {
		if len(l) == 0 || strings.HasPrefix(l, "#") {
			continue
		}
		col := strings.SplitN(l, "\t", 3)
		if len(col) < 3 {
			return nil, fmt.Errorf("line %d parsed failed: column should contain at least 3 tab", i+1)
		}
		fmt.Println("Columns: ", col)
		appID := col[0]
		dayLimit, err := strconv.ParseInt(col[1], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("line %d parsed failed: can not parse column 1 as int64, %v", i+1, err)
		}
		qpsLimit, err := strconv.ParseInt(col[2], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("line %d parsed failed: can not parse column 2 as int64, %v", i+1, err)
		}
		app := appConfig{
			DailyLimit: dayLimit,
			QPSLimit:   qpsLimit,
		}
		if appID == "*" {
			config.Default = app
		} else {
			config.Apps[appID] = app
		}
	}

	return config, nil
}

//newAppFilterByConfig will create a filter and counter by config and a write locker.
//return filter is a simple counting method for identifier, it is multiroutine-safe by the given locker.
//It also return the counter used in the filter func. Be sure to lock the locker if tries to write data
//If app config can not be found, filter will use the default setting.
func newAppFilterByConfig(config *filterConfig, lock *sync.Mutex) (filter Filter, counter map[string]int64) {
	counter = map[string]int64{}
	filter = func(identifier string) bool {
		lock.Lock()
		count, _ := counter[identifier]
		count++
		counter[identifier] = count
		lock.Unlock()
		app, found := config.Apps[identifier]
		if !found {
			app = config.Default
		}
		if count > app.DailyLimit {
			return true
		}
		return false
	}
	return filter, counter
}

func newQPSFilterByConfig(config *filterConfig) (filter Filter, currentQPS *sync.Map) {
	currentQPS = &sync.Map{}
	filter = func(identifier string) bool {
		counter, _ := currentQPS.LoadOrStore(identifier, ratecounter.NewRateCounter(time.Second))
		c := counter.(*ratecounter.RateCounter)
		app, found := config.Apps[identifier]
		if !found {
			app = config.Default
		}
		if c.Incr(1); c.Rate() > app.QPSLimit {
			return true
		}
		return false
	}
	return filter, currentQPS
}
