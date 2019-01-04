//Package faqcluster is a client package for communicate with faq-platform-clustering module
package faqcluster

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"emotibot.com/emotigo/pkg/api"
	"emotibot.com/emotigo/pkg/logger"
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
func NewClient(address *url.URL) *Client {
	return NewClientWithHTTPClient(address, http.DefaultClient)
}

//NewClientWithHTTPClient create a Client with given address & custom HTTPClient.
func NewClientWithHTTPClient(address *url.URL, delegator api.HTTPClient) *Client {
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
	Result     *result                `json:"result"`
}

//clusterRequest is the request struct indicate in the [document](http://wiki.emotibot.com/pages/viewpage.action?pageId=9574324).
type clusterRequest struct {
	Parameters map[string]interface{} `json:"para"`
	Data       []interface{}          `json:"data"`
}

//Result is a Clustering result which contains of several Clusters and Filtered data which does not used in
type Result struct {
	Clusters []Cluster
	Filtered []Data
}

//Cluster represent multiple similar data points, which will have zero to many Tags. the center points will be in CenterQuestions
type Cluster struct {
	CenterQuestions map[string]struct{}
	Data            []Data
	Tags            []string
}

type Data struct {
	Value  string
	Others map[string]interface{}
}

type result struct {
	Clusters []cluster                `json:"data"`
	Filtered []map[string]interface{} `json:"removed"`
}

type cluster struct {
	CenterQuestions []string                 `json:"centerQuestion"`
	Data            []map[string]interface{} `json:"cluster"`
	Tags            []string                 `json:"clusterTag"`
}

type RawError struct {
	StatusCode int
	Msg        string
	Input      interface{}
	Body       []byte
}

func (e *RawError) Error() string {
	return e.Msg
}

//Clustering is an api to do clustering with data.
//Context ctx will be used with http request.
//parameters list can be found at [here](http://wiki.emotibot.com/pages/viewpage.action?pageId=9574324)
func (c *Client) Clustering(ctx context.Context, parameters map[string]interface{}, data []interface{}) (*Result, error) {
	input := clusterRequest{
		Parameters: parameters,
		Data:       data}
	reqbody, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("body marshal failed, %v", err)
	}
	logger.Trace.Printf("faqcluster: send request, %s\n", reqbody)
	req, err := http.NewRequest(http.MethodPost, c.clusterEndpoint, bytes.NewBuffer(reqbody))
	if err != nil {
		return nil, fmt.Errorf("new cluster request failed, %v", err)
	}
	req = req.WithContext(ctx)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request failed, %v", err)
	}
	defer resp.Body.Close()
	rawBody, _ := ioutil.ReadAll(resp.Body)
	logger.Trace.Println("faqcluster: Request Done, detail body. ", string(rawBody))
	if resp.StatusCode != http.StatusOK {
		return nil, &RawError{
			StatusCode: resp.StatusCode,
			Msg:        fmt.Sprintf("request status code is not OK, but %d", resp.StatusCode),
			Input:      reqbody,
			Body:       rawBody,
		}
	}
	var response clusteringResponse
	err = json.Unmarshal(rawBody, &response)
	if err != nil {
		return nil, &RawError{
			StatusCode: resp.StatusCode,
			Msg:        fmt.Sprintf("body decode failed, %v", err),
			Input:      reqbody,
			Body:       rawBody,
		}
	}
	if response.Operation != StatusSuccess {
		return nil, &RawError{
			StatusCode: resp.StatusCode,
			Msg:        fmt.Sprintf("operation %s as failed, %s", response.Operation, response.ErrorMsg),
			Input:      reqbody,
			Body:       rawBody,
		}
	}
	if response.Result == nil {
		return nil, &RawError{
			StatusCode: resp.StatusCode,
			Msg:        "result is null",
			Input:      reqbody,
			Body:       rawBody,
		}
	}
	var result = Result{
		Clusters: []Cluster{},
		Filtered: []Data{},
	}
	for _, c := range response.Result.Clusters {
		var centralQuestions = map[string]struct{}{}
		for _, q := range c.CenterQuestions {
			centralQuestions[q] = struct{}{}
		}
		var receivedData = []Data{}
		for _, d := range c.Data {
			value, found := d["value"]
			if !found {
				return nil, fmt.Errorf("data no value")
			} else {
				delete(d, "value")
			}
			valueStr, ok := value.(string)
			if !ok {
				return nil, fmt.Errorf("data's value is not string")
			}
			receivedData = append(receivedData, Data{Value: valueStr, Others: d})
		}
		cluster := Cluster{
			CenterQuestions: centralQuestions,
			Data:            receivedData,
			Tags:            c.Tags,
		}
		result.Clusters = append(result.Clusters, cluster)
	}
	for _, d := range response.Result.Filtered {
		value, found := d["value"]
		if !found {
			return nil, fmt.Errorf("data no value")
		} else {
			delete(d, "value")
		}
		valueStr, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("data's value is not string")
		}
		result.Filtered = append(result.Filtered, Data{Value: valueStr, Others: d})
	}
	return &result, nil
}
