package emotionengine

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
)

var isIntegration bool

func TestMain(m *testing.M) {
	flag.BoolVar(&isIntegration, "integration", false, "")
	flag.Parse()
	os.Exit(m.Run())
}
func TestClientIntegration(t *testing.T) {
	if !isIntegration {

	}
}

type mockTransporter [][]byte

func (mt mockTransporter) Do(*http.Request) (*http.Response, error) {
	if len(mt) == 0 {
		return nil, fmt.Errorf("you call too many time mock data")
	}
	resp := &http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewReader(mt[len(mt)-1])),
	}
	mt = mt[:len(mt)-1]
	return resp, nil
}

const exampleTrainRequest = `
{
"status": "OK",
"model_id": "5a3a101dc812670fa745073d"
}`

const examplePredictResult = `
{
"status": "OK", "predictions": [{"label": "不满", "other_info": [], "score": 95}]
}
`

func TestClientTrain(t *testing.T) {
	var transporter = &mockTransporter{
		[]byte(exampleTrainRequest),
	}
	c := Client{
		Transport: transporter,
		ServerURL: "",
	}
	var m = Model{
		AppID:        "csbot",
		IsAutoReload: true,
		Data: map[string]Emotion{
			"不滿": Emotion{
				Name:             "不滿",
				PositiveSentence: []string{"不爽", "你怎麼當客服的"},
				NegativeSentence: []string{},
			},
		},
	}
	id, err := c.Train(m)
	if err != nil {
		t.Fatal("expect OK query but got ", err)
	}
	if id != "5a3a101dc812670fa745073d" {
		t.Fatal("expect id to be 5a3a101dc812670fa745073d, but got ", id)
	}

	c.Transport = nil
	e := m.Data["不滿"]
	e.Name = "亂換的"
	m.Data["不滿"] = e
	c.Train(m)
}

func TestClientPredict(t *testing.T) {
	mt := mockTransporter{
		[]byte(examplePredictResult),
	}
	c := Client{
		Transport: mt,
		ServerURL: "",
	}
	req := PredictRequest{
		AppID:    "csbot",
		Sentence: "測試",
	}
	predictions, err := c.Predict(req)
	if err != nil {
		t.Fatal("expect predict result to be OK, but got ", err)
	}
	if predictions == nil || len(predictions) != 1 {
		t.Fatal("expect predictions to has one element, but got predictions: ", predictions)
	}
	if predictions[0].Label != "不满" {
		t.Error("expect predition 0's labe is 不满, but got ", predictions[0].Label)
	}
	if predictions[0].Score != 95 {
		t.Error("expect prediction 0's Score to be 95, but got ", predictions[0].Score)
	}

}
