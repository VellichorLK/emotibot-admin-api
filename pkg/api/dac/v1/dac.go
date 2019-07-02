package v1


import (
	"emotibot.com/emotigo/pkg/api"
	"net/url"
	"fmt"
	"net/http"
	"encoding/json"
	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/pkg/logger"
)

// Client can call dac module api by
type Client struct {
	client  api.HTTPClient
	address string
}


func NewClientWithHTTPClient(address string, client api.HTTPClient) (*Client, error) {
	a, err := url.Parse(address)
	if err != nil {
		return nil, fmt.Errorf("url parsed error: %v", err)
	}
	return &Client{
		client:  client,
		address: a.Scheme + "://" + a.Host + "/ssm/dac/openapi/dac",
	}, nil
}


// IsStandardQuestion determine given content is a standard question (sq) in SSM datastore
func (c *Client) IsStandardQuestion(appID, content string) (bool, error) {
	var (
		err  error
	)

	params := map[string]string{
		"op":   "query",
		"category": "sq",
		"appId": appID,
		"sqContent": content,
	}

	status, rets, err := util.HTTPPostJSONWithStatus(c.address, params, 30)
	if err != nil {
		return false, err
	}

	if status != http.StatusOK {
		return false, fmt.Errorf("response status is %d(not healthy)", status)
	}

	result := SQResult{}

	err = json.Unmarshal([]byte(rets), &result)
	if err != nil {
		logger.Error.Println("Resolve result from json fail:", err.Error())
		return false, err
	}

	if len(result.ActualResults) > 0 && result.Errno == "OK" {
		return true, nil
	}
	return false, nil
}


// IsSimilarQuestion determine given content is a similar question (lq) in SSM datastore
func (c *Client) IsSimilarQuestion(appID, lq string) (bool, error) {
	var (
		err  error
	)

	params := map[string]string{
		"op":   "query",
		"category": "lq",
		"appId": appID,
		"lqContent": lq,
	}

	status, rets, err := util.HTTPPostJSONWithStatus(c.address, params, 30)
	if err != nil {
		return false, err
	}

	if status != http.StatusOK {
		return false, fmt.Errorf("response status is %d(not healthy)", status)
	}

	result := LQResult{}

	err = json.Unmarshal([]byte(rets), &result)
	if err != nil {
		logger.Error.Println("Resolve result from json fail:", err.Error())
		return false, err
	}

	if len(result.ActualResults) > 0 && result.Errno == "OK" {
		return true, nil
	}
	return false, nil
}


// Questions retrive all of the standard question(sq) by given app ID.
func (c *Client) Questions(appID string) ([]string, error) {

	var (
		err  error
	)

	params := map[string]string{
		"op":   "query",
		"category": "sq",
		"appId": appID,
	}

	status, rets, err := util.HTTPPostJSONWithStatus(c.address, params, 30)
	if err != nil {
		return nil, err
	}

	if status != http.StatusOK {
		return nil, fmt.Errorf("response status is %d(not healthy)", status)
	}

	result := SQResult{}

	err = json.Unmarshal([]byte(rets), &result)
	if err != nil {
		logger.Error.Println("Resolve result from json fail:", err.Error())
		return nil, err
	}
	var questions = make([]string, len(result.ActualResults))
	if len(result.ActualResults) > 0 && result.Errno == "OK" {

		for i, item := range result.ActualResults {
			questions[i] = item.Content
		}
	}
	return questions, nil
}

// Questions retrive all of the standard question(sq) by given app ID.
func (c *Client) GetQuestionsMap(appID string) (map[string]string, error) {

	var (
		err  error
	)

	params := map[string]string{
		"op":   "query",
		"category": "sq",
		"appId": appID,
	}

	status, rets, err := util.HTTPPostJSONWithStatus(c.address, params, 30)
	if err != nil {
		return nil, err
	}

	if status != http.StatusOK {
		return nil, fmt.Errorf("response status is %d(not healthy)", status)
	}

	result := SQResult{}

	err = json.Unmarshal([]byte(rets), &result)
	if err != nil {
		logger.Error.Println("Resolve result from json fail:", err.Error())
		return nil, err
	}
	//var questions = make([]string, len(result.ActualResults))
	var questions = make(map[string]string)
	if len(result.ActualResults) > 0 && result.Errno == "OK" {

		for _, item := range result.ActualResults {
			questions[item.Content] = item.Content
		}
	}
	return questions, nil
}

//SimilarQuestions retrive lq (similar question) of given sq(standard question) & app ID.
func (c *Client) SimilarQuestions(appID string, sq string) ([]string, error) {

	var (
		err  error
	)

	params := map[string]string{
		"op":   "query",
		"category": "lq",
		"appId": appID,
		"sqContent": sq,
	}

	status, rets, err := util.HTTPPostJSONWithStatus(c.address, params, 30)
	if err != nil {
		return nil, err
	}

	if status != http.StatusOK {
		return nil, fmt.Errorf("response status is %d(not healthy)", status)
	}

	result := LQResult{}

	err = json.Unmarshal([]byte(rets), &result)
	if err != nil {
		logger.Error.Println("Resolve result from json fail:", err.Error())
		return nil, err
	}
	var lq = make([]string, 0)
	if len(result.ActualResults) > 0 && result.Errno == "OK" {
		for _, item := range result.ActualResults {
			lq = append(lq, item.LqContent)
		}
	}
	return lq, nil

}




// SetSimilarQuestion set given slice of lq to a sq.
func (c *Client) SetSimilarQuestion(appID, sq string, lq ...string) error {

	var (
		err  error
	)
	lqs := make([]string, len(lq))
	for i, content := range lq {
		lqs[i] = content
	}

	params := map[string]interface{}{
		"op":   "add",
		"category": "lq",
		"appId": appID,
		"sqContent": sq,
		"lqContentList": lqs,
	}

	status, rets, err := util.HTTPPostJSONWithStatus(c.address, params, 30)
	if err != nil {
		return err
	}

	if status != http.StatusOK {
		return fmt.Errorf("response status is %d(not healthy)", status)
	}

	result := LQResult{}

	err = json.Unmarshal([]byte(rets), &result)
	if err != nil {
		logger.Error.Println("Resolve result from json fail:", err.Error())
		return err
	}

	if result.Errno != "OK" {
		return fmt.Errorf(result.Errno)
	}
	return nil

}


// DeleteSimilarQuestions Delete lq from the ssm datastore.
func (c *Client) DeleteSimilarQuestions(appID string, lq ...string) error {
	var (
		err  error
	)
	lqs := make([]string, len(lq))
	for i, content := range lq {
		lqs[i] = content
	}

	params := map[string]interface{}{
		"op":   "delete",
		"category": "lq",
		"appId": appID,
		"lqContentList": lqs,
	}

	status, rets, err := util.HTTPPostJSONWithStatus(c.address, params, 30)
	if err != nil {
		return err
	}

	if status != http.StatusOK {
		return fmt.Errorf("response status is %d(not healthy)", status)
	}

	result := LQResult{}

	err = json.Unmarshal([]byte(rets), &result)
	if err != nil {
		logger.Error.Println("Resolve result from json fail:", err.Error())
		return err
	}

	if result.Errno != "OK" {
		return fmt.Errorf(result.Errno)
	}
	return nil
}

// SetSimilarQuestion set given slice of lq to a sq.
func (c *Client) SetSimilarQuestionWithUser(appID, userId, sq string, lq ...string) error {

	var (
		err  error
	)
	lqs := make([]string, len(lq))
	for i, content := range lq {
		lqs[i] = content
	}

	params := map[string]interface{}{
		"op":   "add",
		"category": "lq",
		"appId": appID,
		"sqContent": sq,
		"lqContentList": lqs,
		"userId": userId,
	}

	status, rets, err := util.HTTPPostJSONWithStatus(c.address, params, 30)
	if err != nil {
		return err
	}

	if status != http.StatusOK {
		return fmt.Errorf("response status is %d(not healthy)", status)
	}

	result := LQResult{}

	err = json.Unmarshal([]byte(rets), &result)
	if err != nil {
		logger.Error.Println("Resolve result from json fail:", err.Error())
		return err
	}

	if result.Errno != "OK" {
		return fmt.Errorf(result.Errno)
	}
	return nil

}


// DeleteSimilarQuestions Delete lq from the ssm datastore.
func (c *Client) DeleteSimilarQuestionsWithUser(appID, userId string, lq ...string) error {
	var (
		err  error
	)
	lqs := make([]string, len(lq))
	for i, content := range lq {
		lqs[i] = content
	}

	params := map[string]interface{}{
		"op":   "delete",
		"category": "lq",
		"appId": appID,
		"lqContentList": lqs,
		"userId": userId,
	}

	status, rets, err := util.HTTPPostJSONWithStatus(c.address, params, 30)
	if err != nil {
		return err
	}

	if status != http.StatusOK {
		return fmt.Errorf("response status is %d(not healthy)", status)
	}

	result := LQResult{}

	err = json.Unmarshal([]byte(rets), &result)
	if err != nil {
		logger.Error.Println("Resolve result from json fail:", err.Error())
		return err
	}

	if result.Errno != "OK" {
		return fmt.Errorf(result.Errno)
	}
	return nil
}
// Question retrive sq based on given lq and appID.
func (c *Client) Question(appID, lq string) (string, error) {
	var (
		err  error
	)
	var sq = ""
	params := map[string]string{
		"op":   "query",
		"category": "lq",
		"appId": appID,
		"lqContent": lq,
	}

	status, rets, err := util.HTTPPostJSONWithStatus(c.address, params, 30)
	if err != nil {
		return sq, err
	}

	if status != http.StatusOK {
		return sq, fmt.Errorf("response status is %d(not healthy)", status)
	}

	result := LQResult{}

	err = json.Unmarshal([]byte(rets), &result)
	if err != nil {
		logger.Error.Println("Resolve result from json fail:", err.Error())
		return sq, err
	}

	if len(result.ActualResults) > 0 && result.Errno == "OK" {
		sq = result.ActualResults[0].SqContent
	}
	return sq, nil
}