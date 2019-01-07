//Package logicaccess
package logicaccess

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"emotibot.com/emotigo/pkg/logger"

	"emotibot.com/emotigo/module/admin-api/util"
)

//TrainUnit is unit used to train the logic in cu module
type TrainUnit struct {
	Logic       *TrainLogic      `json:"logic"`
	Keyword     *TrainKeyword    `json:"keyword"`
	Dialog      *CommonTrainData `json:"dialogue_act"`
	UsrResponse *CommonTrainData `json:"user_response"`
}

//TrainLogic is the logic attribute in request body for training cu module
type TrainLogic struct {
	ID   uint64            `json:"app_id,string"`
	Data []*TrainLogicData `json:"data"`
}

//CommonTrainData is common usage data block used to train cu module
type CommonTrainData struct {
	ID   uint64          `json:"app_id,string"`
	Data []*TrainTagData `json:"data"`
}

//TrainKeyword is the keyword attribute in request body for training cu module
type TrainKeyword struct {
	ID   uint64              `json:"app_id,string"`
	Data []*TrainKeywordData `json:"data"`
}

//TrainAPPID is common attribute, app_id, in request body for training cu module
type TrainAPPID struct {
	ID uint64 `json:"app_id,string"`
}

//TrainLogicData is the data attribute in logic which is the attribute in request body for training cu module
type TrainLogicData struct {
	Name        string           `json:"name"`
	Operator    string           `json:"operator"`
	Tags        []string         `json:"tags"`
	TagDistance int              `json:"tag_distance"`
	Constraint  *RangeConstraint `json:"range_constraint,omitempty"`
}

//RangeConstraint is the range_constraint object used in data which is in logic
type RangeConstraint struct {
	Range int    `json:"range"`
	Type  string `json:"type"`
}

//TrainTagData is the structure of tag usage
type TrainTagData struct {
	Tag         string   `json:"tag"`
	PosSentence []string `json:"pos_sentences"`
	NegSentence []string `json:"neg_sentences"`
}

//TrainKeywordData is the structure keyword
type TrainKeywordData struct {
	Tag   string   `json:"tag"`
	Words []string `json:"words"`
}

//StatusResp is the error response from cu module
type StatusResp struct {
	Status string `json:"status"`
	Err    string `json:"error"`
}

//PredictRequest is the structure of input request for prediction
type PredictRequest struct {
	ID      uint64       `json:"app_id,string"`
	Session string       `json:"session_id"`
	Data    *PredictData `json:"data"`
}

//BatchPredictRequest is the structure to predict the whole context result
type BatchPredictRequest struct {
	ID        uint64         `json:"app_id,string"`
	Threshold int            `json:"threshold"`
	Data      []*PredictData `json:"data"`
}

//SessionRequest is used to create a new session
type SessionRequest struct {
	ID        uint64 `json:"app_id,string"`
	Session   string `json:"session_id"`
	Threshold int    `json:"threshold"`
}

//PredictData is the structure used to store sentence data
type PredictData struct {
	SentenceID int    `json:"sentence_id"`
	Sentence   string `json:"sentence"`
}

//PredictResult stores the prediction result
type PredictResult struct {
	Status      string        `json:"status"`
	Threshold   int           `json:"threshold"`
	Logic       []LogicResult `json:"logic_results"`
	Dialogue    []AttrResult  `json:"dialogue_act_results"`
	UsrResponse []AttrResult  `json:"user_response_results"`
	Keyword     []AttrResult  `json:"keyword_results"`
}

//LogicResult gives the logic result
type LogicResult struct {
	LogicRule   TrainLogic     `json:"logic_rule"`
	Predictions [][]Prediction `json:"predictions"`
}

//Prediction gives the tag level result
type Prediction struct {
	Tag       string        `json:"tag"`
	Candidate []PredictData `json:"candidates"`
}

//AttrResult stores the dialogue_act_results, user_response_results, keyword_results data
type AttrResult struct {
	PredictData
	Score     int    `json:"score"`
	Tag       string `json:"tag"`
	Match     string `json:"match,omitempty"`
	MatchText string `json:"match_text,omitempty"`
}

//Client is the struct to implement the communicatoin with cu module
type Client struct {
	URL     string
	Timeout time.Duration
}

//error message
var (
	ErrNoRequest = errors.New("No request data")
	ErrRequest   = errors.New("Wrong request value")
)

type callHTTPPost func(string, interface{}, time.Duration) (int, []byte, error)

var caller callHTTPPost = util.HTTPPostJSONWithStatusByteResp

func (a *Client) postToModule(url string, d interface{}) ([]byte, error) {
	if d == nil {
		return nil, ErrNoRequest
	}

	status, resp, err := caller(url, d, a.Timeout)
	if err != nil {
		logger.Error.Printf("Call cu module failed. %s\n", err)
		return nil, err
	}

	if status != http.StatusOK {
		var errResp StatusResp
		err = json.Unmarshal(resp, &errResp)
		if err != nil {
			logger.Error.Printf("Unmarshal error message failed. %s\n %s\n", err, errResp)
			return nil, err
		}
		return resp, errors.New(errResp.Err)
	}
	return resp, nil
}

//Train calls api /train in cu
func (a *Client) Train(d *TrainUnit) error {
	if d == nil || d.Logic == nil {
		return ErrNoRequest
	}

	//at least one of the attribute has the data
	if d.Dialog == nil && d.Keyword == nil && d.UsrResponse == nil {
		return ErrNoRequest
	}

	//pre-allocating data for calling cu module
	if d.Dialog == nil {
		d.Dialog = &CommonTrainData{ID: d.Logic.ID, Data: make([]*TrainTagData, 0)}
	}
	if d.UsrResponse == nil {
		d.UsrResponse = &CommonTrainData{ID: d.Logic.ID, Data: make([]*TrainTagData, 0)}
	}
	if d.Keyword == nil {
		d.Keyword = &TrainKeyword{ID: d.Logic.ID, Data: make([]*TrainKeywordData, 0)}
	}

	_, err := a.postToModule(a.URL+"/train", d)
	return err
}

//Status calls api /status to get the training status
func (a *Client) Status(d *TrainAPPID) (string, error) {
	resp, err := a.postToModule(a.URL+"/status", d)
	if err != nil {
		return "", err
	}
	var statusMsg StatusResp
	err = json.Unmarshal(resp, &statusMsg)
	if err != nil {
		logger.Error.Printf("Unmarshal failed. %s. %s\n", err, resp)
		return "", err
	}
	return statusMsg.Status, errors.New(statusMsg.Err)
}

//PredictAndUnMarshal calls the cu module to predict the result and unmarshal
func (a *Client) PredictAndUnMarshal(d *PredictRequest) (*PredictResult, error) {
	if d == nil || d.Session == "" || d.Data == nil {
		return nil, ErrRequest
	}
	respBytes, err := a.postToModule(a.URL+"/predict", d)
	if err != nil {
		return nil, err
	}
	return a.unmarshalResp(respBytes)
}

//Predict calls the cu module to predict the result
func (a *Client) Predict(d *PredictRequest) ([]byte, error) {
	if d == nil || d.Session == "" || d.Data == nil {
		return nil, ErrRequest
	}
	return a.postToModule(a.URL+"/predict", d)
}

//BatchPredict predicts the batch reuslt
func (a *Client) BatchPredict(d *BatchPredictRequest) ([]byte, error) {
	if d == nil || d.Threshold < 0 || len(d.Data) == 0 {
		return nil, ErrRequest
	}
	return a.postToModule(a.URL+"/batch_predict", d)
}

//BatchPredictAndUnMarshal predicts the batch reuslt
func (a *Client) BatchPredictAndUnMarshal(d *BatchPredictRequest) (*PredictResult, error) {
	if d == nil || d.Threshold < 0 || len(d.Data) == 0 {
		return nil, ErrRequest
	}
	respBytes, err := a.postToModule(a.URL+"/batch_predict", d)
	if err != nil {
		return nil, err
	}
	return a.unmarshalResp(respBytes)
}

func (a *Client) unmarshalResp(d []byte) (*PredictResult, error) {
	var resp PredictResult
	err := json.Unmarshal(d, &resp)
	if err != nil {
		logger.Error.Printf("unmarshal failed. %s\n", err)
	}
	return &resp, err
}

//SessionCreate is used on the fly
func (a *Client) SessionCreate(d *SessionRequest) error {
	if d == nil || d.Session == "" || d.Threshold < 0 {
		return ErrRequest
	}
	_, err := a.postToModule(a.URL+"/session/create", d)
	return err
}

//SessionDelete is called when the session is finished on the fly
func (a *Client) SessionDelete(d *SessionRequest) error {
	if d == nil || d.Session == "" || d.Threshold < 0 {
		return ErrRequest
	}
	_, err := a.postToModule(a.URL+"/session/delete",
		struct {
			*SessionRequest
			Threshold bool `json:"threshold,omitempty"`
		}{
			SessionRequest: d,
		})
	return err
}

//UnloadModel unloads the model when is no use
func (a *Client) UnloadModel(d *TrainAPPID) error {
	_, err := a.postToModule(a.URL+"/unload_model", d)
	return err
}
