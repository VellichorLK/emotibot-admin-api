package logicaccess

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"
)

func mockHTTPPostMedhod(url string, d interface{}, timeout time.Duration) (int, []byte, error) {
	paths := strings.Split(url, "/")
	num := len(paths)
	if num == 0 {
		return http.StatusOK, nil, nil
	}

	callAPI := paths[num-1]

	okStaus, _ := json.Marshal(struct {
		Status string `json:"string"`
	}{
		Status: "OK",
	})

	readyStatus, _ := json.Marshal(struct {
		Status string `json:"string"`
	}{
		Status: "ready",
	})

	var resp PredictResult

	resp.Dialogue = make([]AttrResult, 0)
	resp.Keyword = make([]AttrResult, 0)
	resp.UsrResponse = make([]AttrResult, 0)
	resp.Logic = make([]LogicResult, 0)
	resp.Status = "OK"

	okRsp, _ := json.Marshal(resp)

	switch callAPI {
	case "predict":
		fallthrough
	case "batch_predict":
		return http.StatusOK, okRsp, nil
	case "status":
		return http.StatusOK, readyStatus, nil
	case "train":
		fallthrough
	case "create":
		fallthrough
	case "delete":
		fallthrough
	case "unload_model":
		return http.StatusOK, okStaus, nil
	default:
		return 404, nil, nil
	}

}

func SetupMockPostMethod() {
	caller = mockHTTPPostMedhod
}

func TestTrain(t *testing.T) {
	SetupMockPostMethod()
	c := &Client{URL: "http://127.0.0.1"}
	d := &TrainUnit{}
	err := c.Train(d)
	if err == nil {
		t.Errorf("expecting has error, but get nil\n")
	}

	d.Logic = &TrainLogic{}
	err = c.Train(d)
	if err == nil {
		t.Errorf("expecting has error, but get nil\n")
	}

	d.Keyword = &TrainKeyword{}
	err = c.Train(d)
	if err != nil {
		t.Errorf("expecting no error, but get %s\n", err)
	}
}

func TestPredictAndUnMarshal(t *testing.T) {
	SetupMockPostMethod()
	c := &Client{URL: "http://127.0.0.1"}
	var r PredictRequest
	resp, err := c.PredictAndUnMarshal(&r)
	if err != nil {
		t.Errorf("expecting no error, but get %s\n", err)
	}

	if resp.Status != "OK" {
		t.Errorf("expecting OK, but get %s\n", resp.Status)
	}
}

func TestSessionCreate(t *testing.T) {
	SetupMockPostMethod()
	c := &Client{URL: "http://127.0.0.1"}
	var r SessionRequest
	err := c.SessionCreate(&r)
	if err == nil {
		t.Errorf("expecting error, but get no error\n")
	}

	r.Session = "12345"
	r.Threshold = 70

	err = c.SessionCreate(&r)
	if err != nil {
		t.Errorf("expecting no error, but get %s\n", err)
	}
}

func TestSessionDelete(t *testing.T) {
	SetupMockPostMethod()
	c := &Client{URL: "http://127.0.0.1"}
	var r SessionRequest
	err := c.SessionDelete(&r)
	if err == nil {
		t.Errorf("expecting error, but get no error\n")
	}

	r.Session = "12345"
	r.Threshold = 70

	err = c.SessionDelete(&r)
	if err != nil {
		t.Errorf("expecting no error, but get %s\n", err)
	}
}
