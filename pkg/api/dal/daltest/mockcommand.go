package daltest

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type mockCommand interface {
	NewResponse() *http.Response
}

type closingBuffer struct {
	*bytes.Buffer
}

type response struct {
	ErrNo      string        `json:"errno"`
	ErrMessage string        `json:"errmsg"`
	Results    []interface{} `json:"actualResults"`
	Operation  []string      `json:"results"`
}

func (c *closingBuffer) Close() error {
	return nil
}

type deleteSimilarQuestionsCmd struct {
	appID string
	lq    []string
}

func (d *deleteSimilarQuestionsCmd) NewResponse() *http.Response {
	resp := &response{
		ErrNo: "OK",
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
