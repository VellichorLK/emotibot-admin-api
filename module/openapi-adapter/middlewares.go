package main

import (
	"net/http"
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
		trafficManager.Summarize(appID, responseLogger, r, responseTime)
	}
}
