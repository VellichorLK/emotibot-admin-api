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

type daoFlowCreate struct {
	typ          int
	leftChannel  int
	rightChannel int
	enterprise   string
	callTime     int64
	uploadTime   int64
	updateTime   int64
	fileName     string
	uuid         string
	user         string
}

type apiFlowAddBody struct {
	StartTime float64 `json:"start_time"`
	EndTime   float64 `json:"end_time"`
	Text      string  `json:"text"`
	Speaker   string  `json:"speaker"`
}

//Segment is vad segment
type Segment struct {
	callID    uint64
	asr       *apiFlowAddBody
	channel   int
	creatTime int64
}

//ConversationInfo is information in Conversation table
type ConversationInfo struct {
	CallID       uint64
	Status       int
	FileName     string
	FilePath     string
	VoiceID      uint64
	CallComment  string
	Transaction  int
	Series       string
	CallTime     int64
	UploadTime   int64
	UpdateTime   int64
	HostID       string
	HostName     string
	Extension    string
	Department   string
	GuestID      string
	GuestName    string
	GuestPhone   string
	CallUUID     string
	Enterprise   string
	User         string
	Duration     int
	ApplyGroup   []uint64
	Type         int
	LeftChannel  int
	RightChannel int
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

//QIFlowResult give the reuslt of qi flow
type QIFlowResult struct {
	FileName  string               `json:"file_name"`
	Result    []*QIFlowGroupResult `json:"cu_result"`
	Sensitive []string             `json:"sensitive"`
}

//QIFlowGroupResult gives the result of check
type QIFlowGroupResult struct {
	ID       uint64      `json:"-"`
	Name     string      `json:"group_name"`
	QIResult []*QIResult `json:"qi_result"`
}

//QIResult gives the result of rule
type QIResult struct {
	ID          uint64         `json:"-"`
	Name        string         `json:"controller_rule"`
	Valid       bool           `json:"valid"`
	LogicResult []*LogicResult `json:"logic_results"`
}

//LogicResult give the result of logic
type LogicResult struct {
	ID        uint64   `json:"-"`
	Name      string   `json:"logic_rule"`
	Valid     bool     `json:"valid"`
	Recommend []string `json:"recommend"`
}
