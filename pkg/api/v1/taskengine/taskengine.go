package taskengine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"emotibot.com/emotigo/pkg/api/v1"
)

//Client is api tool for taskengine
type Client struct {
	Location *url.URL
	AppID    string
	Client   api.HTTPClient
}

type argsET struct {
	AppID                string                 `json:"AppID"`
	UserID               string                 `json:"UserID"`
	Text                 string                 `json:"Text"`
	CU                   map[string]interface{} `json:"CU"`
	QualifiedScenarioIDs string                 `json:"QualifiedScenarioIDs,omitempty"`
}

type Response struct {
	Text       string `json:"TTSText"`
	ScenarioID string `json:"HitScenarioID"`
	Flag       int    `json:"FinalFlag"`
}

// ET call api from TaskEngine.
func (c *Client) ET(uid, text string) ([]byte, error) {
	u := &url.URL{
		Path: "/task_engine/ET",
	}
	path := c.Location.ResolveReference(u)
	var body = argsET{
		AppID:  c.AppID,
		UserID: uid,
		CU:     make(map[string]interface{}, 0),
		Text:   text,
	}
	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("post data marshal failed, %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, path.String(), bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("")
	}
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request failed, %v", err)
	}
	defer resp.Body.Close()
	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read failed, %v", err)
	}
	return data, nil
}

// ParseETResponse parse the input data to taskengine's Response.
// -1 失敗(找無場景, 無APPID)
// -2 解析失敗(問妳哪個城市, 你回天氣很好)
func ParseETResponse(data []byte) (*Response, error) {
	var response Response
	err := json.Unmarshal(data, &response)
	if err != nil {
		return nil, fmt.Errorf("json unmarshal failed, %v", err)
	}
	return &response, nil
}
