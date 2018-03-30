package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"emotibot.com/emotigo/module/systex-controller/api/asr"
	"emotibot.com/emotigo/module/systex-controller/api/taskengine"
)

//api global client for ASR module
var asrClient asr.Client
var teClient taskengine.Client
var requiredEnvs = []string{
	"ASR_URL", //ASR Hostname and port
	"TE_URL",  //Task Engine Hostname and port
}

type v2TextResponse struct {
	Text string `json:"text"`
}

func voiceToTextHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintln(w, "Only support POST method.")
		return
	}

	if r.Header.Get("content-type") != "audio/x-wav" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "content-type only support audio/x-wav.")
		return
	}
	sentence, err := asrClient.Recognize(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, err)
		return
	}
	var resp v2TextResponse
	log.Printf("v2Text: %s", sentence)
	resp.Text = sentence
	data, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintln(w, string(data))
}

func voiceToTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintln(w, "Only support POST method.")
		return
	}

	if r.Header.Get("content-type") != "audio/x-wav" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "content-type only support audio/x-wav.")
		return
	}
	var userID = r.Header.Get("X-UserID")
	if userID == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "UserID is empty")
		return
	}

	sentence, err := asrClient.Recognize(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, err)
		return
	}
	resp, err := teClient.ET(userID, sentence)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)

}

func main() {
	var env = make(map[string]string)
	for _, key := range requiredEnvs {
		value, ok := os.LookupEnv(key)
		if !ok {
			log.Fatalf("Need env: %s", key)
		}
		env[key] = value
	}

	c := &http.Client{
		Timeout: time.Duration(5) * time.Second,
	}
	asrURL, err := url.Parse(env["ASR_URL"])
	if err != nil {
		log.Fatal(err)
	}
	asrClient.Client = c
	asrClient.Location = asrURL
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/V2Task", voiceToTaskHandler)
	mux.HandleFunc("/v1/V2Text", voiceToTextHandler)
	server := http.Server{
		Addr:    ":80",
		Handler: mux,
	}
	log.Println("Server starting")
	log.Fatal(server.ListenAndServe())
}
