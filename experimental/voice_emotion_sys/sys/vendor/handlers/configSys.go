package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/hashicorp/consul/api"
)

//MinimumSecond minimum second of system
func MinimumSecond(w http.ResponseWriter, r *http.Request) {
	appid := r.Header.Get(NUAPPID)
	if appid == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method == "PUT" {

		type miniSecond struct {
			Sec int `json:"second"`
		}

		minSec := &miniSecond{}
		err := json.NewDecoder(r.Body).Decode(minSec)
		if err != nil {
			http.Error(w, "Bad Request: "+err.Error(), http.StatusBadRequest)
			return
		}

		if minSec.Sec < 0 {
			http.Error(w, "Bad Request second < 0", http.StatusBadRequest)
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

	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

}
