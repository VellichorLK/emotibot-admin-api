package main

import (
	"encoding/base64"
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
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

// ContextKey is use for context index.
type ContextKey string

// RawBodyKey is use for reusing request raw body.
var RawBodyKey ContextKey = "rawBody"
var LogBodyKey ContextKey = "logBody"

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
	go func() {
		for {
			time.Sleep(time.Duration(1) * time.Minute)
			textLogger.Flush()
			taskEngineLogger.Flush()
		}
	}()
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
	mux.HandleFunc("/v1/V2Task", LogCSV(SaveAudioBody(voiceToTaskHandler), taskEngineLogger))
	mux.HandleFunc("/v1/V2Text", LogCSV(SaveAudioBody(voiceToTextHandler), textLogger))
	mux.Handle("/logging/", HTTPBasicAuth(http.StripPrefix("/logging", http.FileServer(http.Dir("/app/log")))))
	server := http.Server{
		Addr:    ":80",
		Handler: mux,
	}
	log.Println("Server starting")
	log.Fatal(server.ListenAndServe())
}

// AudioResponsWriter should save input audio file and any log that should be logged down in csv writer.
type AudioResponseWriter struct {
	http.ResponseWriter
	AudioFile []byte
	Logs      []string
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

// SaveAudioBody is a middleware for checking request is a audio/x-wav file and writing down the body.
// It will use current UCT Timestamp as the file name.
func SaveAudioBody(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		aw, ok := w.(*AudioResponseWriter)
		if !ok {
			aw = &AudioResponseWriter{
				ResponseWriter: w,
				AudioFile:      nil,
				Logs:           nil,
			}
		}
		if r.Header.Get("content-type") != "audio/x-wav" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "content-type only support audio/x-wav.")
			return
		}
		var uniqueID = r.Header.Get("uniqueID")
		if r.Header.Get("uniqueID") == "" {
			uniqueID = strconv.FormatInt(time.Now().Unix(), 10)
		}
		r.Header.Set("uniqueID", uniqueID)
		fileName := "/app/log/" + uniqueID + ".wav"
		f, err := os.Create(fileName)
		if err != nil {
			log.Println("create file error," + err.Error())
			h.ServeHTTP(aw, r)
			return
		}
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Printf("copy body failed, reason:%v. continue request.\n", err)
			h.ServeHTTP(aw, r)
			return
		}
		aw.AudioFile = data
		f.Close()
		h.ServeHTTP(aw, r)
	}
}

// LogCSV will use header as
func LogCSV(h http.HandlerFunc, logger *csv.Writer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		aw, ok := w.(*AudioResponseWriter)
		if !ok {
			aw = &AudioResponseWriter{
				ResponseWriter: w,
				AudioFile:      nil,
				Logs:           nil,
			}
		}
		h.ServeHTTP(aw, r)
		var uniqueID string
		if uniqueID = r.Header.Get("uniqueID"); uniqueID == "" {
			log.Println("No uniqueID, skip logging.")
			return
		}
		if aw.Logs == nil || len(aw.Logs) == 0 {
			log.Println("logBody is empty, skip logging.")
			return
		}
		aw.Logs = append([]string{time.Now().Format("Mon Jan 2 15:04:05"), uniqueID}, aw.Logs...)
		err := logger.Write(aw.Logs)
		if err != nil {
			log.Println("log write failed, " + err.Error())
		}
	}
}

//HTTPBasicAuth protect the h by HTTP basic auth protocol.
func HTTPBasicAuth(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("WWW-Authenticate", `Basic realm="Resticted"`)
		s := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
		if len(s) != 2 {
			http.Error(w, "Not authorized", 401)
			return
		}
		b, err := base64.StdEncoding.DecodeString(s[1])
		if err != nil {
			http.Error(w, err.Error(), 401)
			return
		}
		pair := strings.SplitN(string(b), ":", 2)
		if len(pair) != 2 {
			http.Error(w, "Not authorized", 401)
			return
		}

		if pair[0] != "user1" || pair[1] != "12345" {
			http.Error(w, "Not authorized", 401)
			return
		}

		h.ServeHTTP(w, r)
	}
}
