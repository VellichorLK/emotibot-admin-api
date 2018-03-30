package cuservice

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"emotibot.com/emotigo/module/systex-controller/api"
)

type Client struct {
	Location *url.URL
	Client   api.HTTPClient
}

func (c *Client) Simplify(sentence string) (string, error) {
	req, err := newSimplifyRequest(c.Location.String(), sentence)
	if err != nil {
		return "", fmt.Errorf("create simplify request failed, %v", err)
	}
	res, err := c.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("dial %s failed, %v", c.Location.String(), err)
	}
	defer res.Body.Close()
	simplifiedSentence, err := parseSimplifyResponse(res.Body)
	if err != nil {
		return "", fmt.Errorf("response body parse failed, %v", err)
	}
	return simplifiedSentence, nil
}

// newSimplifyRequest creates a cuservice api request for client to use.
// return error if http.NewRequest failed.
func newSimplifyRequest(host, sentence string) (*http.Request, error) {
	input := url.Values{
		"src": []string{sentence},
	}
	url := fmt.Sprintf("%s/cuservice/rest/nlp/simplified?%s", host, input.Encode())
	return http.NewRequest(http.MethodGet, url, nil)
}

func parseSimplifyResponse(r io.Reader) (string, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return "", err
	}
	var apiResponse struct {
		Sentence string `json:"sentence"`
	}
	err = json.Unmarshal(data, &apiResponse)
	if err != nil {
		return "", err
	}

	return apiResponse.Sentence, nil
}
