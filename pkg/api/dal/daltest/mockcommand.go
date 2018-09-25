package daltest

import (
	"bytes"
	"encoding/json"
	"net/http"

	dal "emotibot.com/emotigo/pkg/api/dal/v1"
)

type mockCommand interface {
	NewResponse() *http.Response
}

type closingBuffer struct {
	*bytes.Buffer
}

func (c *closingBuffer) Close() error {
	return nil
}

var errNoSuccess = "OK"

type deleteSimilarQuestionsCmd struct {
	appID string
	lq    []string
}

func (d *deleteSimilarQuestionsCmd) NewResponse() *http.Response {
	resp := &dal.RawResponse{
		ErrNo: errNoSuccess,
	}
	expectResp, _ := json.Marshal(resp)

	r := &closingBuffer{bytes.NewBuffer(expectResp)}
	return &http.Response{
		Status:     "200 OK",
		StatusCode: http.StatusOK,
		Header:     http.Header{},
		Body:       r,
	}
}

type matchCommand struct {
	result *ExpectResult
}

func (c *matchCommand) NewResponse() *http.Response {
	resp := &dal.RawResponse{
		ErrNo:     errNoSuccess,
		Operation: []string{"OK"},
	}
	data, _ := json.Marshal(resp)
	return &http.Response{
		Status:     "200 OK",
		StatusCode: http.StatusOK,
		Body:       &closingBuffer{bytes.NewBuffer(data)},
	}
}

// ExpectResult represent the result from dal module, which can provide find grind controll over mock result.
type ExpectResult struct {
	results    []interface{}
	operations []string
}

func (r *ExpectResult) WillReturn(Results []interface{}, Operations []string) {
	//TODO: implement me
	r.results = Results
	r.operations = Operations
}
