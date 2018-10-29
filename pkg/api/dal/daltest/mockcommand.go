package daltest

import (
	"bytes"
	"encoding/json"
	"net/http"

	dal "emotibot.com/emotigo/pkg/api/dal/v1"
)

// mockCommand should create the correct http.Response based on inject result and expect behavior.
// dal client then can interprate it.
// TODO: Add a verify method to verify correct input.
type mockCommand interface {
	NewResponse() *http.Response
}

//closingBuffer is a warp struct to allow bytes.Buffer satisfy io.ReadCloser interface.
type closingBuffer struct {
	*bytes.Buffer
}

func (c *closingBuffer) Close() error {
	return nil
}

var errNoSuccess = "OK"

type execCommand struct {
	result *ExpectResult
}
type deleteSimilarQuestionsCmd struct {
	*execCommand
	appID string
	lq    []string
}

func (d *execCommand) NewResponse() *http.Response {
	resp := &dal.RawResponse{
		ErrNo: errNoSuccess,
	}
	expectResp, _ := json.Marshal(resp)
	r := &closingBuffer{bytes.NewBuffer(expectResp)}
	if d.result.hasErr {
		return &http.Response{
			Status:     "400 Bad Request",
			StatusCode: http.StatusBadRequest,
			Body:       &closingBuffer{&bytes.Buffer{}},
		}
	}
	return &http.Response{
		Status:     "200 OK",
		StatusCode: http.StatusOK,
		Body:       r,
	}
}

//matchCommand handle request of boolean response.
type matchCommand struct {
	result *ExpectResult
	input  []interface{}
}

func (c *matchCommand) NewResponse() *http.Response {
	var resp *dal.RawResponse
	if c.result.hasErr {
		return &http.Response{
			Status:     "500 Internal Server Error",
			StatusCode: http.StatusInternalServerError,
			Body:       &closingBuffer{&bytes.Buffer{}},
		}
	}
	if c.result.values != nil && len(c.result.values) > 0 && c.result.values[0] == true {
		resp = &dal.RawResponse{
			ErrNo:     errNoSuccess,
			Operation: []string{"OK"},
		}
	} else {
		resp = &dal.RawResponse{
			ErrNo:     "OTHER",
			Operation: []string{"FAILED"},
		}

	}

	data, _ := json.Marshal(resp)
	return &http.Response{
		Status:     "200 OK",
		StatusCode: http.StatusOK,
		Body:       &closingBuffer{bytes.NewBuffer(data)},
	}
}

type queryQuestionsCommand struct {
	result *ExpectResult
	input  []interface{}
}

func (c *queryQuestionsCommand) NewResponse() *http.Response {
	var response dal.RawResponse
	if c.result.values != nil {
		response.ErrNo = errNoSuccess
		for _, value := range c.result.values {
			var record = map[string]interface{}{
				"content": value,
			}
			response.Results = append(response.Results, record)
		}
		data, _ := json.Marshal(response)
		cb := &closingBuffer{bytes.NewBuffer(data)}
		return &http.Response{
			Status:     "200 OK",
			StatusCode: http.StatusOK,
			Body:       cb,
		}
	}

	return &http.Response{
		Status:     "400 Bad Request",
		StatusCode: http.StatusBadRequest,
		Body:       &closingBuffer{&bytes.Buffer{}},
	}
}

type questionByLQCommand struct {
	result *ExpectResult
}

func (c *questionByLQCommand) NewResponse() *http.Response {
	var response dal.RawResponse
	if c.result.values != nil {
		response.ErrNo = errNoSuccess
		for _, value := range c.result.values {
			var record = map[string]interface{}{
				"parentContent": value,
			}
			response.Results = append(response.Results, record)
		}
		data, _ := json.Marshal(response)
		cb := &closingBuffer{bytes.NewBuffer(data)}
		return &http.Response{
			Status:     "200 OK",
			StatusCode: http.StatusOK,
			Body:       cb,
		}
	}

	return &http.Response{
		Status:     "400 Bad Request",
		StatusCode: http.StatusBadRequest,
		Body:       &closingBuffer{&bytes.Buffer{}},
	}
}
