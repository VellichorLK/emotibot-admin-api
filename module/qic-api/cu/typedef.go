package cu

type apiFlowCreateBody struct {
	CreateTime int64  `json:"create_time"`
	FileName   string `json:"file_name"`
}

type apiFlowCreateResp struct {
	UUID string `json:"id"`
}

type apiFlowFinish struct {
	FinishTime int64 `json:"finish_time"`
}

//V1PredictContext is context that sent to cu to predict
type V1PredictContext struct {
	AppID     string                  `json:"app_id"`
	Threshold int                     `json:"threshold"`
	Data      []*V1PredictRequestData `json:"data"`
}

//V1PredictRequestData is data unit for predict
type V1PredictRequestData struct {
	ID       int    `json:"id"`
	Sentence string `json:"sentence"`
}

//V1PredictResult is the prediction result
type V1PredictResult struct {
	Status      string          `json:"status"`
	Threshold   int             `json:"threshold"`
	LogicResult []V1LogicResult `json:"logic_results"`
	//currently we don't use this information, just use interface to catch it
	//IntentResult  interface{} `json:"intent_results"`
	//KeywordResult interface{} `json:"keyword_results"`
}

//V1LogicResult gives the logic result
type V1LogicResult struct {
	LogicRule V1LogicRule `json:"logic_rule"`
	//Predictions [][]V1Prediction `json:"predictions"`
}

//V1LogicRule is the relative logic information
type V1LogicRule struct {
	Name            string            `json:"name"`
	Operator        string            `json:"operator"`
	Tags            []string          `json:"tags"`
	TagDistance     int               `json:"tag_distance"`
	RangeConstraint V1RangeConstraint `json:"range_constraint"`
}

//V1RangeConstraint gives the constraint of the logic
type V1RangeConstraint struct {
	Range int    `json:"range"`
	Type  string `json:"top"`
}

//V1Prediction gives the tag level result
type V1Prediction struct {
	Tag       string `json:"tag"`
	Candidate []V1PredictRequestData
}
