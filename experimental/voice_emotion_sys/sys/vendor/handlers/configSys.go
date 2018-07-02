package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/hashicorp/consul/api"
)

type miniSecond struct {
	Sec int `json:"second,omitempty"`
}

//MinimumSecond minimum second of system
func MinimumSecond(w http.ResponseWriter, r *http.Request) {
	appid := r.Header.Get(NUAPPID)
	if appid == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method == "PUT" {
		putMinHandleSec(w, r)
	} else if r.Method == "GET" {
		getMinHandleSec(w, r)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

}

func putMinHandleSec(w http.ResponseWriter, r *http.Request) {
	appid := r.Header.Get(NUAPPID)
	minSec := &miniSecond{}
	err := json.NewDecoder(r.Body).Decode(minSec)
	if err != nil {
		http.Error(w, "Bad Request: "+err.Error(), http.StatusBadRequest)
		return
	}

	if minSec.Sec < 1 {
		http.Error(w, "Bad Request second < 1", http.StatusBadRequest)
		return
	}

	consulKey := ConsulMinimumSecKey + "/" + appid
	kv := consulClient.KV()

	val := strconv.Itoa(minSec.Sec)

	// PUT a new KV pair
	p := &api.KVPair{Key: consulKey, Value: []byte(val)}
	_, err = kv.Put(p, nil)
	if err != nil {
		log.Printf("Error when put kv to consul %s\n", err)
		http.Error(w, "Internal server error ", http.StatusInternalServerError)
		return
	}
}

func getMinHandleSec(w http.ResponseWriter, r *http.Request) {
	appid := r.Header.Get(NUAPPID)
	consulKey := ConsulMinimumSecKey + "/" + appid
	kv := consulClient.KV()

	pair, _, err := kv.Get(consulKey, nil)
	if err != nil {
		log.Printf("Error  getting kv %s from consul %s\n", consulKey, err)
		http.Error(w, "Internal server error ", http.StatusInternalServerError)
		return
	}
	var sec int

	if pair != nil {
		sec, err = strconv.Atoi(string(pair.Value))
		if err != nil {
			log.Printf("transform minimum second error %s,%s", consulKey, err)
			http.Error(w, "Internal server error ", http.StatusInternalServerError)
			return
		}
	}

	miniSec := miniSecond{Sec: sec}

	encodeRes, err := json.Marshal(miniSec)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal server error ", http.StatusInternalServerError)
		return
	}
	contentType := "application/json; charset=utf-8"

	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)
	w.Write(encodeRes)
}
