package asr

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"emotibot.com/emotigo/pkg/api"
)

//API is for communicate with ASR Module for Systex project.
type Client struct {
	Location *url.URL
	Client   api.HTTPClient
}

// Recognize is the api of ASR module. will transform input voice data into sentence.
// Notice: api only support .wav and 8k sample rate now.
func (c *Client) Recognize(voice io.Reader) (string, error) {
	p := url.URL{
		Path: "/client/dynamic/recognize",
	}
	req, err := http.NewRequest(http.MethodPost, c.Location.ResolveReference(&p).String(), voice)
	if err != nil {
		return "", fmt.Errorf("new request failed, %v", err)
	}
	response, err := c.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("do request failed, %v", err)
	}
	defer response.Body.Close()
	return parseRecognize(response.Body)
}

//recognizeBody only used as internal struct for decode json string.
type recognizeBody struct {
	Status     int `json:"status"`
	Hypotheses []struct {
		Utterance string `json:"utterance"`
	} `json:"hypotheses"`
}

//parseRecognize should be used to parsing the sentence out of response's body from recognize api.
func parseRecognize(body io.Reader) (string, error) {
	data, err := ioutil.ReadAll(body)
	if err != nil {
		return "", err
	}
	var b recognizeBody
	err = json.Unmarshal(data, &b)
	if err != nil {
		return "", fmt.Errorf("json unmarshal failed, %v, data: %s", err, string(data))
	}
	if b.Status != 0 {
		return "", fmt.Errorf("bad recognize response: status is %d", b.Status)
	}
	if len(b.Hypotheses) == 0 {
		return "", fmt.Errorf("bad recognize response: empty hypothese")
	}
	return b.Hypotheses[0].Utterance, nil
}
