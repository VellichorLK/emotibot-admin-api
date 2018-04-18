package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"emotibot.com/emotigo/module/systex-controller/api/taskengine"

	"github.com/siongui/gojianfan"
)

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
		log.Println(err)
		return
	}
	var resp v2TextResponse
	log.Printf("v2Text: %s\n", sentence)
	textLogger.Write([]string{time.Now().String(), sentence})
	textLogger.Flush()
	resp.Text = sentence
	data, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
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
		CustomError(w, "asr api error:"+err.Error())
		return
	}
	sentence, err = csClient.Simplify(sentence)
	if err != nil {
		CustomError(w, "Cu Service api error:"+err.Error())
		return
	}

	resp, err := teClient.ET(userID, sentence)
	if err != nil {
		CustomError(w, "task engine error:"+err.Error())
		return
	}
	teResp, err := taskengine.ParseETResponse(resp)
	if err != nil {
		CustomError(w, "task engine response err:"+err.Error())
		return
	}
	var out = Response{
		Text:       gojianfan.S2T(teResp.Text),
		ScenarioID: teResp.ScenarioID,
		Flag:       teResp.Flag,
		AsrText:    sentence,
	}
	outJSONData, err := json.Marshal(out)
	if err != nil {
		CustomError(w, "response json err: "+err.Error())
		return
	}

	log.Printf("v2Task: %s, te result: %v\n", sentence, out)
	taskEngineLogger.Write([]string{time.Now().String(), sentence, string(outJSONData)})
	taskEngineLogger.Flush()
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(outJSONData)
}

type Response struct {
	Text       string `json:"TTSText"`
	ScenarioID string `json:"HitScenarioID"`
	Flag       int    `json:"FinalFlag"`
	AsrText    string `json:"asr_text"`
}

//Custom
func CustomError(w http.ResponseWriter, message string) {
	w.WriteHeader(http.StatusInternalServerError)
	log.Println(message)
	fmt.Fprintln(w, message)
}
