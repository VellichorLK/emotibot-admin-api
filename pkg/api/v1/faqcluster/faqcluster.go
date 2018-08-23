//Package faqcluster is a client package for communicate with faq-platform-clustering module
package faqcluster

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"emotibot.com/emotigo/pkg/api/v1"
)

//StatusSuccess is the status for ClusteringResponse
const StatusSuccess = "Success"

//Client use internal api.HTTPClient to communicate with different api endpoints.
type Client struct {
	client          api.HTTPClient
	clusterEndpoint string
	resultEndpoint  string
}

//NewClient create a Client with given address.
//It is a convention Client creation which will use http.DefaultClient as delegator.
//So be care with issues happened a lot in http.DefaultClient
func NewClient(address url.URL) *Client {
	return NewClientWithDelegator(address, http.DefaultClient)
}

//NewClientWithDelegator create a Client with given address & http delegator.
func NewClientWithDelegator(address url.URL, delegator api.HTTPClient) *Client {
	clusterAddr, _ := address.Parse("clustering/post")
	resultAddr, _ := address.Parse("get_result")
	return &Client{
		client:          delegator,
		clusterEndpoint: clusterAddr.String(),
		resultEndpoint:  resultAddr.String(),
	}
}

//ClusteringResponse use
type ClusteringResponse struct {
	Status  string `json:"status"`
	Message string `json:"reason"`
	Done    bool   `json:"processDone"`
	TaskID  string `json:"taskId"`
	Batch   int    `json:"batch"`
}

type clusterRequest struct {
	Data []interface{} `json:"data"`
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

//ErrNotDone is return by GetResult when result api return a unfinished stats.
var ErrNotDone = errors.New("faqcluster error: process is not done")

//Clustering is an api to create a clustering task with input data.
//It will be an Async function, use ClusteringResponse to check GetResult function
func (c *Client) Clustering(data []interface{}) (ClusteringResponse, error) {
	var response ClusteringResponse
	reqbody, err := json.Marshal(clusterRequest{Data: data})
	if err != nil {
		return ClusteringResponse{}, fmt.Errorf("body marshal failed, %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, c.clusterEndpoint, bytes.NewBuffer(reqbody))
	if err != nil {
		return ClusteringResponse{}, fmt.Errorf("new cluster request failed, %v", err)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return ClusteringResponse{}, fmt.Errorf("http request failed, %v", err)
	}
	log.Println(req.URL)
	if resp.StatusCode != http.StatusOK {
		return ClusteringResponse{}, fmt.Errorf("request status code is not OK, but %d", resp.StatusCode)
	}
	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ClusteringResponse{}, fmt.Errorf("read io failed, %v", err)
	}
	defer resp.Body.Close()
	err = json.Unmarshal(respData, &response)
	if err != nil {
		log.Printf("receive raw body: %s\n", respData)
		return ClusteringResponse{}, fmt.Errorf("response body is not correctly formated, %v", err)
	}
	return response, nil
}

//GetResult is an api to retrive Result with input taskID. return ErrNotDone if Result have not produced yet.
func (c *Client) GetResult(taskID string) (*Result, error) {
	query := url.Values{
		"reportId": []string{taskID},
	}.Encode()
	req, err := http.NewRequest(http.MethodGet, c.resultEndpoint+"?"+query, nil)
	if err != nil {
		return nil, fmt.Errorf("new result request failed, %v", err)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request failed, %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("response http status %d is not OK", resp.StatusCode)
	}
	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read io failed, %v", err)
	}
	rawResponse := resultRawResponse{}
	err = json.Unmarshal(respData, &rawResponse)
	if err != nil {
		log.Printf("raw response: %s\n", respData)
		return nil, fmt.Errorf("response body is not correctly formated, %v", err)
	}

	if !rawResponse.Done {
		return nil, ErrNotDone
	}

	if rawResponse.ErrorCode != 0 {
		return nil, fmt.Errorf("task is not success, error code %d", rawResponse.ErrorCode)
	}

	return &rawResponse.Result, nil
}
