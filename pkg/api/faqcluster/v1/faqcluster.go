//Package faqcluster is a client package for communicate with faq-platform-clustering module
package faqcluster

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"emotibot.com/emotigo/pkg/api"
)

//StatusSuccess is the status for ClusteringResponse
const StatusSuccess = "success"

//Client use internal api.HTTPClient to communicate with different api endpoints.
type Client struct {
	client          api.HTTPClient
	clusterEndpoint string
}

// NewClient create a Client with given address.
// It is a convention Client creation which will use http.DefaultClient as delegator.
// Not recommend use in production, because the issues with http.DefaultClient.
func NewClient(address url.URL) *Client {
	return NewClientWithHTTPClient(address, http.DefaultClient)
}

//NewClientWithHTTPClient create a Client with given address & custom HTTPClient.
func NewClientWithHTTPClient(address url.URL, delegator api.HTTPClient) *Client {
	clusterAddr, _ := address.Parse("clustering")
	return &Client{
		client:          delegator,
		clusterEndpoint: clusterAddr.String(),
	}
}

//clusteringResponse is the response struct indicate in the [document](http://wiki.emotibot.com/pages/viewpage.action?pageId=9574324).
type clusteringResponse struct {
	Operation  string                 `json:"errno"`
	ErrorMsg   string                 `json:"error_message"`
	Parameters map[string]interface{} `json:"para"`
	Result     Result                 `json:"results"`
}

//clusterRequest is the request struct indicate in the [document](http://wiki.emotibot.com/pages/viewpage.action?pageId=9574324).
type clusterRequest struct {
	Parameters map[string]interface{} `json:"para"`
	Data       []interface{}          `json:"data"`
}

//Result is a Clustering result which contains of several Clusters and Filtered data which does not used in
type Result struct {
	Clusters []Cluster     `json:"data"`
	Filtered []interface{} `json:"removed"`
}

//Cluster represent multiple similar data points, which will have zero to many Tags. the center points will be in CenterQuestions
type Cluster struct {
	CenterQuestions []string      `json:"centerQuestion"`
	Data            []interface{} `json:"cluster"`
	Tags            []string      `json:"clusterTag"`
}

type resultRawResponse struct {
	ErrorCode int    `json:"error"`
	Done      bool   `json:"processDone"`
	Result    Result `json:"results"`
}

type RawError struct {
	Msg   string
	Input interface{}
	Body  io.Reader
}

func (e *RawError) Error() string {
	return e.Msg
}

//Clustering is an api to do clustering with data.
//Context ctx will be used with http request.
//parameters list can be found at [here](http://wiki.emotibot.com/pages/viewpage.action?pageId=9574324)
func (c *Client) Clustering(ctx context.Context, parameters map[string]interface{}, data []interface{}) (*Result, error) {
	var response clusteringResponse
	input := clusterRequest{
		Parameters: parameters,
		Data:       data}
	reqbody, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("body marshal failed, %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, c.clusterEndpoint, bytes.NewBuffer(reqbody))
	if err != nil {
		return nil, fmt.Errorf("new cluster request failed, %v", err)
	}
	req = req.WithContext(ctx)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request failed, %v", err)
	}
	var rawBody = &bytes.Buffer{}
	io.Copy(rawBody, resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, &RawError{
			Msg:   fmt.Sprintf("request status code is not OK, but %d", resp.StatusCode),
			Input: reqbody,
			Body:  rawBody,
		}
	}
	err = json.NewDecoder(rawBody).Decode(&response)
	if err != nil {
		return nil, &RawError{
			Msg:   fmt.Sprintf("body decode failed, %v", err),
			Input: reqbody,
			Body:  rawBody,
		}
	}
	if response.Operation != StatusSuccess {
		return nil, &RawError{
			Msg:   fmt.Sprintf("operation %s as failed, %s", response.Operation, response.ErrorMsg),
			Input: reqbody,
			Body:  rawBody,
		}
	}

	return &response.Result, nil
}
