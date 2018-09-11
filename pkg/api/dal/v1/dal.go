/*
Package dal contains api with dal module.
dal module itself is a service layer for SSM(FAQ) datastore.
Use `NewClientWithHTTPClient` to create a Client to call api.
DetailError contains detail response from dal module.
*/
package dal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"emotibot.com/emotigo/pkg/api"
)

// Client can call dal module api by
type Client struct {
	client  api.HTTPClient
	address string
}

// NewClientWithHTTPClient create a Client based on address & client given.
// address should contain schema://hostname:port/ only, the path will be cleanup
// error is returned if address can not be parsed as url.
func NewClientWithHTTPClient(address string, client api.HTTPClient) (*Client, error) {
	a, err := url.Parse(address)
	if err != nil {
		return nil, fmt.Errorf("url parsed error: %v", err)
	}
	return &Client{
		client:  client,
		address: a.Scheme + "://" + a.Host + "/dal",
	}, nil
}

// IsStandardQuestion determine given content is a standard question (sq) in SSM datastore
func (c *Client) IsStandardQuestion(appID, content string) (bool, error) {
	var (
		err  error
		body bytes.Buffer
		req  *http.Request
		resp *http.Response
	)
	json.NewEncoder(&body).Encode(request{
		Op:           "query",
		Category:     "sq",
		AppID:        appID,
		UserRecordID: 0,
		Data: data{
			Subop: "defaultSubop",
			Entities: []entity{
				entity{
					Content: content,
				},
			},
			// Segments: []string{"recordId"},
		},
	})
	req, err = http.NewRequest(http.MethodPost, c.address, &body)
	if err != nil {
		return false, fmt.Errorf("new request error, %v", err)
	}
	resp, err = c.client.Do(req)
	if err != nil {
		return false, fmt.Errorf("do request failed, %v", err)
	}
	defer resp.Body.Close()
	if code := resp.StatusCode; code != http.StatusOK {
		return false, fmt.Errorf("response status is %d(not healthy)", code)
	}
	var respContent response
	err = json.NewDecoder(resp.Body).Decode(&respContent)
	if err != nil {
		return false, fmt.Errorf("response format error, %v", err)
	}
	if len(respContent.Operation) > 0 && respContent.Operation[0] == "OK" {
		return true, nil
	}
	return false, nil
}

// IsSimilarQuestion determine given content is a similar question (lq) in SSM datastore
func (c *Client) IsSimilarQuestion(appID, lq string) (bool, error) {
	var (
		err  error
		body bytes.Buffer
		req  *http.Request
		resp *http.Response
	)
	json.NewEncoder(&body).Encode(request{
		Op:           "query",
		Category:     "lq",
		AppID:        appID,
		UserRecordID: 0,
		Data: data{
			Subop: "defaultSubop",
			Entities: []entity{
				entity{
					Content: lq,
				},
			},
		},
	})
	req, err = http.NewRequest(http.MethodPost, c.address, &body)
	if err != nil {
		return false, fmt.Errorf("new request error, %v", err)
	}
	resp, err = c.client.Do(req)
	if err != nil {
		return false, fmt.Errorf("do request failed, %v", err)
	}
	defer resp.Body.Close()
	if code := resp.StatusCode; code != http.StatusOK {
		return false, fmt.Errorf("response status is %d(not healthy)", code)
	}
	var respContent response
	err = json.NewDecoder(resp.Body).Decode(&respContent)
	if err != nil {
		return false, fmt.Errorf("response format error, %v", err)
	}
	if len(respContent.Operation) > 0 && respContent.Operation[0] == "OK" {
		return true, nil
	}
	return false, nil
}

// Questions retrive all of the standard question(sq) by given app ID.
func (c *Client) Questions(appID string) ([]string, error) {
	var (
		body bytes.Buffer
		err  error
		req  *http.Request
		resp *http.Response
	)
	json.NewEncoder(&body).Encode(request{
		Op:           "query",
		Category:     "sq",
		AppID:        "csbot",
		UserRecordID: 0,
		Data: data{
			Subop:    "conditionsSubop",
			Segments: []string{"recordId", "content"},
		},
	})
	req, err = http.NewRequest(http.MethodPost, c.address, &body)
	if err != nil {
		return nil, fmt.Errorf("new request error, %v", err)
	}
	resp, err = c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request failed, %v", err)
	}
	defer resp.Body.Close()
	if code := resp.StatusCode; code != http.StatusOK {
		return nil, fmt.Errorf("response status is %d(not healthy)", code)
	}
	var content response
	err = json.NewDecoder(resp.Body).Decode(&content)
	if err != nil {
		return nil, fmt.Errorf("response format error, %v", err)
	}
	var questions = make([]string, len(content.Results))
	for i, record := range content.Results {
		r, ok := record.(map[string]interface{})
		if !ok {
			continue
		}
		questions[i] = r["content"].(string)
	}
	return questions, nil
}

//SimilarQuestions retrive lq (similar question) of given sq(standard question) & app ID.
func (c *Client) SimilarQuestions(appID string, sq string) ([]string, error) {
	var buffer bytes.Buffer
	json.NewEncoder(&buffer).Encode(request{
		Op:           "query",
		Category:     "lq",
		AppID:        appID,
		UserRecordID: 0,
		Data: data{
			Subop: "conditionsSubop",
			Conditions: &conditions{
				ParentContent: sq,
			},
			Segments: []string{"recordId", "content"},
		},
	})
	r, err := http.NewRequest(http.MethodPost, c.address, &buffer)
	if err != nil {
		return nil, fmt.Errorf("new request error, %v", err)
	}
	var resp *http.Response
	resp, err = c.client.Do(r)
	if err != nil {
		return nil, fmt.Errorf("do request failed, %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request status is %d(not healthy)", resp.StatusCode)
	}
	var respBody response
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		return nil, &DetailError{
			ErrMsg: "response body error" + err.Error(),
		}
	}
	var lq = make([]string, 0)
	for _, result := range respBody.Results {
		r, ok := result.(map[string]interface{})
		if !ok {
			continue
		}
		lq = append(lq, r["content"].(string))
	}
	return lq, nil
}

// SetSimilarQuestion set given slice of lq to a sq.
func (c *Client) SetSimilarQuestion(appID, sq string, lq ...string) error {
	entities := make([]entity, len(lq))
	for i, content := range lq {
		entities[i] = entity{
			Content:       content,
			ParentContent: sq,
		}
	}
	var buffer bytes.Buffer
	json.NewEncoder(&buffer).Encode(request{
		Op:           "insert",
		Category:     "lq",
		AppID:        appID,
		UserRecordID: 0,
		Data: data{
			Subop:    "defaultSubop",
			Entities: entities,
		},
	})
	r, err := http.NewRequest(http.MethodPost, c.address, &buffer)
	if err != nil {
		return fmt.Errorf("new request error, %v", err)
	}
	var resp *http.Response
	resp, err = c.client.Do(r)
	if err != nil {
		return fmt.Errorf("do request failed, %v", err)
	}
	defer resp.Body.Close()
	var respBody response
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		return fmt.Errorf("dal error: response body format failed, %v", err)
	}
	if respBody.ErrNo != "OK" {
		return &DetailError{
			ErrMsg:  "got response error " + respBody.ErrNo + " with message " + respBody.ErrMessage,
			Results: respBody.Operation,
		}
	}
	return nil
}

// DeleteSimilarQuestions Delete lq from the ssm datastore.
func (c *Client) DeleteSimilarQuestions(appID string, lq ...string) error {
	var entities = make([]entity, len(lq))
	for i, q := range lq {
		entities[i] = entity{
			Content: q,
		}
	}
	var buffer bytes.Buffer
	json.NewEncoder(&buffer).Encode(request{
		Op:           "delete",
		Category:     "lq",
		AppID:        appID,
		UserRecordID: 0,
		Data: data{
			Subop:    "defaultSubop",
			Entities: entities,
		},
	})
	r, err := http.NewRequest(http.MethodPost, c.address, &buffer)
	if err != nil {
		return fmt.Errorf("new request error, %v", err)
	}
	var resp *http.Response
	resp, err = c.client.Do(r)
	if err != nil {
		return fmt.Errorf("do request failed, %v", err)
	}
	defer resp.Body.Close()
	var respBody response
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		return fmt.Errorf("decode dal body failed")
	}
	if respBody.ErrNo != "OK" {
		return &DetailError{
			ErrMsg:  "response error " + respBody.ErrNo + " with msg " + respBody.ErrMessage,
			Results: respBody.Operation,
		}
	}
	return nil
}

// Question retrive sq based on given lq and appID.
func (c *Client) Question(appID, lq string) (string, error) {
	var buffer bytes.Buffer
	json.NewEncoder(&buffer).Encode(request{
		Op:           "query",
		Category:     "lq",
		AppID:        appID,
		UserRecordID: 0,
		Data: data{
			Subop: "conditionsSubop",
			Conditions: &conditions{
				Content: lq,
			},
			Segments: []string{"parentContent"},
		},
	})
	req, err := http.NewRequest(http.MethodPost, c.address, &buffer)
	if err != nil {
		return "", fmt.Errorf("new request error, %v", err)
	}
	resp, err := c.client.Do(req)
	defer resp.Body.Close()
	var respBody response
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		return "", fmt.Errorf("decode dal body failed")
	}
	r, ok := respBody.Results[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("response convert failed")
	}
	content, ok := r["parentContent"].(string)
	if !ok {
		return "", fmt.Errorf("result content is not string")
	}
	return content, nil
}
