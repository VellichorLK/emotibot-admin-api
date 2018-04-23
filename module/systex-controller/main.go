package main

import (
	"encoding/csv"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"emotibot.com/emotigo/module/systex-controller/api/asr"
	"emotibot.com/emotigo/module/systex-controller/api/cuservice"
	"emotibot.com/emotigo/module/systex-controller/api/taskengine"
)

//api global client for ASR module
var asrClient asr.Client
var teClient taskengine.Client
var csClient cuservice.Client

// requiredEnvs
var requiredEnvs = []string{
	"ASR_URL", //ASR Hostname and port
	"TE_URL",  //Task Engine Hostname and port
	"CS_URL",  //CuService Hostname and port
	"APPID",   //APPID for task engine
}

var textLogger *csv.Writer
var taskEngineLogger *csv.Writer

func main() {
	csvFile, err := GetStreamingFile("/app/log/v2Text.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer csvFile.Close()
	csvFile2, err := GetStreamingFile("/app/log/v2TE.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer csvFile.Close()
	textLogger = csv.NewWriter(csvFile)
	taskEngineLogger = csv.NewWriter(csvFile2)

	var env = make(map[string]string)
	for _, key := range requiredEnvs {
		value, ok := os.LookupEnv(key)
		if !ok {
			log.Fatalf("Need env: %s", key)
		}
		env[key] = value
	}

	c := &http.Client{
		Timeout: time.Duration(15) * time.Second,
	}

	asrURL, err := url.Parse(env["ASR_URL"])
	if err != nil {
		log.Fatal(err)
	}
	asrClient.Client = c
	asrClient.Location = asrURL

	teURL, err := url.Parse(env["TE_URL"])
	if err != nil {
		log.Fatal(err)
	}
	teClient = taskengine.Client{
		Client:   c,
		Location: teURL,
		AppID:    env["APPID"],
	}

	csURL, err := url.Parse(env["CS_URL"])
	if err != nil {
		log.Fatal(err)
	}
	csClient = cuservice.Client{
		Location: csURL,
		Client:   c,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/V2Task", voiceToTaskHandler)
	mux.HandleFunc("/v1/V2Text", voiceToTextHandler)
	mux.Handle("/logging/", http.StripPrefix("/logging", http.FileServer(http.Dir("/app/log"))))
	server := http.Server{
		Addr:    ":80",
		Handler: mux,
	}
	log.Println("Server starting")
	log.Fatal(server.ListenAndServe())
}

func GetStreamingFile(filePath string) (*os.File, error) {
	logFile, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0644)
	if os.IsNotExist(err) {
		logFile, err = os.Create(filePath)
	}
	if err != nil {
		log.Fatalln(err)
	}
	return logFile, nil
}
