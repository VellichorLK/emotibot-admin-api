package SelfLearning

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type responseClient struct {
	URL    string
	client *http.Client
}

//rresponse is response's response
type rresponse struct {
	OtherInfo *responseInfo `json:"response_other_info"`
}
type responseInfo struct {
	Custom *responseCustom `json:"response_custom_return"`
}
type responseCustom struct {
	RelatedQ []*relatedQ `json:"relatedQ"`
}
type relatedQ struct {
	UserQ string  `json:"userQ"`
	StdQ  string  `json:"stdQ"`
	Score float64 `json:"score"`
}
type responseArg struct {
	UniqueID   string      `json:"UniqueID"`
	UserID     string      `json:"UserID"`
	Text0      string      `json:"Text0"`
	Text1Old   string      `json:"Text1_Old"`
	Text1      string      `json:"Text1"`
	Robot      string      `json:"robot"`
	CreateTime string      `json:"createdtime"`
	ModuleType string      `json:"module_type"`
	ModuleList string      `json:"module_list"`
	CustomInfo *customInfo `json:"customInfo"`
}

type customInfo struct {
	QType    string `json:"qtype"`
	Top      string `json:"top"`
	Platform string `json:"platform"`
	Brand    string `json:"brand"`
	Sex      string `json:"sex"`
	Age      string `json:"age"`
	Hobbies  string `json:"hobbies"`
}

type sortMapKey struct {
	mapData   map[string]float64
	sliceData []interface{}
}

func (mk *sortMapKey) keyToSlice() {
	if mk.mapData != nil {
		size := len(mk.mapData)
		mk.sliceData = make([]interface{}, size, size)
		counter := 0
		for key := range mk.mapData {
			mk.sliceData[counter] = key
			counter++
		}
	}
}

func (mk *sortMapKey) Len() int {
	return len(mk.sliceData)
}
func (mk *sortMapKey) Less(i, j int) bool {
	return mk.mapData[mk.sliceData[i].(string)] > mk.mapData[mk.sliceData[j].(string)]
}
func (mk *sortMapKey) Swap(i, j int) {
	mk.sliceData[i], mk.sliceData[j] = mk.sliceData[j], mk.sliceData[i]
}

func newResponseArg(appid, text string) *responseArg {
	return &responseArg{
		Text0:      text,
		Text1:      text,
		Text1Old:   text,
		UniqueID:   "123456",
		UserID:     "789",
		Robot:      appid,
		CreateTime: "2039-11-25 19:54:36",
		ModuleType: "enable",
		ModuleList: "faq",
		CustomInfo: &customInfo{
			QType:    "debug",
			Top:      "5",
			Platform: "all",
			Brand:    "",
			Sex:      "",
			Age:      "",
			Hobbies:  "",
		},
	}
}

func (r *responseClient) newRequest(appid, sentence string) (*http.Request, error) {
	arg := newResponseArg(appid, sentence)
	content, err := json.Marshal(&arg)
	if err != nil {
		return nil, err
	}
	buf := bytes.NewReader(content)
	return http.NewRequest(http.MethodPost, r.URL, buf)
}
func (r *responseClient) Post(appid, sentence string) (*rresponse, error) {
	req, err := r.newRequest(appid, sentence)
	if err != nil {
		return nil, fmt.Errorf("request formatting error, ")
	}
	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	response := &rresponse{}
	err = json.Unmarshal(body, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}
