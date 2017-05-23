package handlers

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

//DefaultHandler default handler for http request
func DefaultHandler(w http.ResponseWriter, r *http.Request) {

	module, ok := funcSupport[r.URL.Path]
	if !ok {
		http.Error(w, "Url Error", http.StatusBadRequest)
		return
	}

	contentType, ok := module.Method[r.Method]
	if !ok {
		http.Error(w, "Method unsupported", http.StatusBadRequest)
		return
	}

	buf, _ := ioutil.ReadAll(r.Body)
	query := r.URL.RawQuery
	decodeQuery, err := url.QueryUnescape(query)

	if err != nil {
		http.Error(w, "Query parameter wrong format", http.StatusBadRequest)
		log.Println(query, " wrong format encoding query")
		return
	}

	task := &TaskBlock{Method: r.Method, Path: r.URL.Path, Body: string(buf), Query: decodeQuery}

	encodeTask, err := json.Marshal(task)

	if err != nil {
		http.Error(w, "Json convert err", http.StatusInternalServerError)
		log.Printf("path:%s, method:%s, body:%s, query:%s", r.URL.Path, r.Method, string(buf), decodeQuery)
	} else {
		//log.Println(string(encodeTask))
		result, errContentType, statusCode := PushTask(string(encodeTask), module.Queue, 1000)

		if errContentType != "" {
			contentType = errContentType
		}

		contentType += "; charset=utf-8"

		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(statusCode)
		w.Write([]byte(result))
	}

}
