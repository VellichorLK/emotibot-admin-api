package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"time"

	uuid "github.com/satori/go.uuid"
)

const (
	statusIETraining   = "training"
	statusIETrainReady = "ready"
	statusIETrainError = "error"
)

type trainPayload struct {
	AppID  string `json:"app_id"`
	Reload bool   `json:"auto_reload`
}
type statusPayload struct {
	AppID   string `json:"app_id"`
	ModelID string `json:"model_id"`
}

const (
	fileMap   = "./.map_data"
	fileStart = "./.start_data"
)

func main() {
	modelMap := map[string][]string{}
	modelStart := map[string]int64{}

	_, err := os.Stat(fileMap)
	if !os.IsNotExist(err) {
		data, err := ioutil.ReadFile(fileMap)
		if err == nil {
			err = json.Unmarshal(data, &modelMap)
		}
	}
	_, err = os.Stat(fileStart)
	if !os.IsNotExist(err) {
		data, err := ioutil.ReadFile(fileStart)
		if err == nil {
			err = json.Unmarshal(data, &modelStart)
		}
	}
	fmt.Printf("Model map: %+v\n", modelMap)
	fmt.Printf("Model time: %+v\n", modelStart)

	http.HandleFunc("/train", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		payload := trainPayload{}
		jsonEncoder := json.NewDecoder(r.Body)
		defer r.Body.Close()
		err := jsonEncoder.Decode(&payload)
		if err != nil {
			returnErr(w, err.Error())
			return
		}
		uid, _ := uuid.NewV1()

		if _, ok := modelMap[payload.AppID]; !ok {
			modelMap[payload.AppID] = []string{}
		}

		modelMap[payload.AppID] = append(modelMap[payload.AppID], uid.String())
		ret := map[string]interface{}{
			"status":   statusIETraining,
			"model_id": uid.String(),
		}
		modelStart[uid.String()] = time.Now().Unix()
		returnJSON(w, ret)
	})
	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		payload := statusPayload{}
		jsonEncoder := json.NewDecoder(r.Body)
		defer r.Body.Close()
		err := jsonEncoder.Decode(&payload)
		if err != nil {
			returnErr(w, err.Error())
			return
		}

		uuids, ok := modelMap[payload.AppID]
		if !ok {
			returnErr(w, "Invalid appid")
			return
		}

		find := false
		for _, uid := range uuids {
			if uid == payload.ModelID {
				find = true
			}
		}
		if !find {
			returnErr(w, "Invalid ModelID")
			return
		}
		status := statusIETraining
		now := time.Now().Unix()
		fmt.Printf("Now: %d, check %d\n", now, modelStart[payload.ModelID])
		if now-modelStart[payload.ModelID] > 10 {
			status = statusIETrainReady
		}

		ret := map[string]interface{}{
			"status":   status,
			"model_id": payload.AppID,
		}
		returnJSON(w, ret)
		fmt.Printf("Model map: %+v\n", modelMap)
		fmt.Printf("Model time: %+v\n", modelStart)
	})
	srv := http.Server{Addr: ":15503"}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for _ = range c {
			fmt.Printf("Close server with: %#v\n%#v\n", modelMap, modelStart)
			mapData, _ := json.Marshal(modelMap)
			startData, _ := json.Marshal(modelStart)
			ioutil.WriteFile(fileMap, mapData, 0644)
			ioutil.WriteFile(fileStart, startData, 0644)
			srv.Shutdown(nil)
		}
	}()

	fmt.Println("Start server at 0.0.0.0:15503")
	err = srv.ListenAndServe()
	if err != nil {
		fmt.Println("Err: ", err.Error())
	}
}

func returnErr(w http.ResponseWriter, err string) {
	ret := map[string]string{
		"error": err,
	}
	returnJSON(w, ret)
}

func returnJSON(w http.ResponseWriter, ret interface{}) {
	data, _ := json.Marshal(ret)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(data)
}
