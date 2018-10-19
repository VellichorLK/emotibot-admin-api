package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"emotibot.com/emotigo/module/openapi-adapter/data"
	"emotibot.com/emotigo/pkg/logger"
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

func NewDailyLimitMiddleWare(next http.HandlerFunc, globalApps map[string]int64, maximum int64) http.HandlerFunc {
	var lock = sync.Mutex{}
	return func(w http.ResponseWriter, r *http.Request) {
		metadata, err := GetMetadata(r)
		appID, _ := metadata[AppIDKey]
		if err != nil {
			resp, _ := json.Marshal(data.ErrorResponse{
				Message: fmt.Sprintf("get metad data failed, %v", err),
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
		if count++; count > maximum {
			r.Header.Set("X-Filtered", "true")
		}
		globalApps[appID] = count
		lock.Unlock()
		next.ServeHTTP(w, r)
	}
}
