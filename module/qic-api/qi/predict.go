package qi

import (
	"errors"

	"emotibot.com/emotigo/module/qic-api/util/logicaccess"
)

var predictor *logicaccess.Client

//error message
var (
	ErrNoPredictConn   = errors.New("No connection to cu module")
	ErrNeedSentence    = errors.New("Need sentences")
	ErrThreshold       = errors.New("Wrong threshold")
	ErrNeedSession     = errors.New("No Session is assigned")
	ErrNeedIdxSentence = errors.New("Need sentence id")
)

//BatchPredict predicts the sentences by appID. Basically appID is tagID, maybe RuleGroup id in the future.
func BatchPredict(appID uint64, threshold int, sentences []string) (*logicaccess.PredictResult, error) {
	if predictor == nil {
		return nil, ErrNoPredictConn
	}
	if len(sentences) == 0 {
		return nil, ErrNeedSentence
	}
	if threshold < 0 {
		return nil, ErrThreshold
	}

	var r logicaccess.BatchPredictRequest
	r.ID = appID
	r.Threshold = threshold
	for i := 0; i < len(sentences); i++ {
		s := &logicaccess.PredictData{SentenceID: i + 1, Sentence: sentences[i]}
		r.Data = append(r.Data, s)
	}
	return predictor.BatchPredictAndUnMarshal(&r)
}

//CreateSession creates the session to do the qi on line
func CreateSession(appID uint64, session string, threshold int) error {
	if predictor == nil {
		return ErrNoPredictConn
	}
	if session == "" {
		return ErrNeedSession
	}
	if threshold < 0 {
		return ErrThreshold
	}
	r := &logicaccess.SessionRequest{ID: appID, Session: session, Threshold: threshold}
	return predictor.SessionCreate(r)
}

//Predict predict a single sentence, iTHsentence is i'th sentence in the whole context
func Predict(appID uint64, session string, iTHsentence int, senstence string) (*logicaccess.PredictResult, error) {
	if predictor == nil {
		return nil, ErrNoPredictConn
	}
	if session == "" {
		return nil, ErrNeedSession
	}
	if iTHsentence <= 0 {
		return nil, ErrNeedIdxSentence
	}
	if senstence == "" {
		return nil, ErrNeedSentence
	}

	d := &logicaccess.PredictData{SentenceID: iTHsentence, Sentence: senstence}
	r := &logicaccess.PredictRequest{ID: appID, Session: session, Data: d}
	return predictor.PredictAndUnMarshal(r)
}

//DeleteSession deletes the session. Must be called after finishing the session
func DeleteSession(appID uint64, session string) error {
	if predictor == nil {
		return ErrNoPredictConn
	}
	if session == "" {
		return ErrNeedSession
	}
	r := &logicaccess.SessionRequest{ID: appID, Session: session}
	return predictor.SessionDelete(r)
}
