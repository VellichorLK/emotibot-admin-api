package data

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"emotibot.com/emotigo/module/vipshop-admin/SelfLearning/internal/model"
	"emotibot.com/emotigo/module/vipshop-admin/util"
	json2 "github.com/bitly/go-simplejson"
)

type WordCount struct {
	mutex  *sync.Mutex
	counts map[string]int
	total  int
}

func (w *WordCount) GetWordCount() map[string]int {
	return w.counts
}

const nluQuerySize = 20

func (l *NativeLog) GetWordPos(nluURL string, questions []string) {
	start := time.Now()
	util.LogTrace.Println("Load Word-Pos start.")

	size := len(questions)
	for i := 0; i < size/nluQuerySize; i++ {
		partial := questions[i*nluQuerySize : (i+1)*nluQuerySize]
		dalItems := parseFromWordPosRsp(
			callWordPosService(
				getWordPosRequest(partial), nluURL))
		for _, dalItem := range dalItems {
			l.Logs = append(l.Logs, dalItem)
		}
	}
	util.LogTrace.Printf("Load Word-Pos ends. Time consuming: [%v]s",
		time.Since(start).Seconds())
}

func getWordPosRequest(questions []string) string {
	jsonArray, err := json.Marshal(questions)
	if err != nil {
		return ""
	}

	return fmt.Sprintf(`{
			"queries": %s,
			"flags": "segment,keyword",
			"time": "false",
			"appid": "0"
		}`, string(jsonArray))
}

func callWordPosService(request string, nluURL string) []byte {
	response, err := http.Post(
		nluURL,
		"application/json",
		strings.NewReader(request))

	if err != nil {
		util.LogError.Printf("Cannot fetch NLU service:%s, %+v\n ", nluURL, err)
		return nil
	}

	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		util.LogError.Printf("IO read error %s\n", err)
	}

	return body
}

func parseFromWordPosRsp(rsp []byte) []*DalItem {
	jsonRsp, err := json2.NewJson(rsp)
	if err != nil {
		util.LogWarn.Println("fail to parse response")
		return nil
	}
	var dalItems []*DalItem
	jsonArray := jsonRsp.MustArray()

	for i := range jsonArray {
		dalItem := DalItem{}
		dalItem.Word2Vec = make(map[string]model.Vector)
		dalItem.Embedding = make(model.Vector, 0)
		singleResult := jsonRsp.GetIndex(i)

		segment := singleResult.Get("segment").MustArray()
		if err != nil {
			util.LogWarn.Println("parse_segment_fail")
			continue
		}
		var wordPos []string

		for _, obj := range segment {
			jsonObj, _ := json.Marshal(obj)
			json2Obj, _ := json2.NewJson(jsonObj)
			pos, _ := json2Obj.Get("word").String()
			if StopWordList[pos] == 1 || len(pos) == 0 {
				continue
			}
			wordPos = append(wordPos, pos)
		}

		keyword := singleResult.Get("keyword").MustArray()
		if err != nil {
			util.LogWarn.Println("parse_segment_fail")
			continue
		}
		var keys []string

		for _, obj := range keyword {
			jsonObj, _ := json.Marshal(obj)
			json2Obj, _ := json2.NewJson(jsonObj)
			pos, _ := json2Obj.Get("word").String()
			if StopWordList[pos] == 1 || len(pos) == 0 {
				continue
			}
			keys = append(keys, pos)
		}

		if len(wordPos) == 0 || len(keys) == 0 {
			continue
		}

		dalItem.Content, err = singleResult.Get("query").String()
		if err != nil {
			util.LogWarn.Println("parse_query_fail: %s\n ", singleResult)
			continue
		}

		dalItem.Tokens = wordPos
		dalItem.KeyWords = keys
		dalItems = append(dalItems, &dalItem)
	}
	return dalItems
}
