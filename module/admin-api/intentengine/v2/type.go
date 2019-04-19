package intentenginev2

type SentenceType int

var TypePositive SentenceType = 0
var TypeNegative SentenceType = 1

// SentenceV2 describe a sentence in trainint data
type SentenceV2 struct {
	ID      int64  `json:"id"`
	Content string `json:"content"`
}

// SentenceV2WithType describe a sentence in trainint data
type SentenceV2WithType struct {
	SentenceV2
	Type int `json:"type"`
}

// IntentV2 describe a intent in V2
type IntentV2 struct {
	ID            int64          `json:"id"`
	Name          string         `json:"name"`
	PositiveCount int            `json:"positive_count"`
	NegativeCount int            `json:"negative_count"`
	Positive      *[]*SentenceV2 `json:"positive,omitempty"`
	Negative      *[]*SentenceV2 `json:"negative,omitempty"`
}

// VersionInfo describe a version
type VersionInfoV2 struct {
	Version           int     `json:"version"`
	IntentEngineModel *string `json:"intent_engine_model"`
	RuleEngineModel   *string `json:"rule_engine_model,omitempty"`
	InUse             bool    `json:"in_use"`
	TrainStartTime    *int64  `json:"start_train"`
	TrainEndTime      *int64  `json:"end_train"`
	Progress          int     `json:"progress"`
	TrainResult       int     `json:"train_result`
}

type StatusV2 struct {
	Status           string `json:"status"`
	LastFinishTime   *int64 `json:"last_train,omitempty"`
	CurrentStartTime *int64 `json:"current_start,omitempty"`
	Progress         int    `json:"progress"`
	Version          int    `json:"version"`
}

// IETrainStatus is return structure of intent-engine module
type IETrainStatus struct {
	Status  string `json:"status"`
	ModelID string `json:"model_id"`
}

// TrainDataResponse is response structure for return training data
type TrainDataResponse struct {
	Status     string         `json:"status"`
	AppID      string         `json:"app_id"`
	Intent     []*TrainIntent `json:"intent"`
	IntentDict []*TrainDict   `json:"intent_dict"`
}

func NewTrainDataResponse(appid string) *TrainDataResponse {
	return &TrainDataResponse{
		Status:     "OK",
		AppID:      appid,
		Intent:     []*TrainIntent{},
		IntentDict: []*TrainDict{},
	}
}

type TrainIntent struct {
	Name      string         `json:"intent_name"`
	Sentences *TrainSentence `json:"sentences"`
	Features  *TrainFeature  `json:"features"`
}

func (intent *TrainIntent) Load(input *IntentV2) {
	intent.Name = input.Name
	intent.Features = &TrainFeature{}
	intent.Sentences = &TrainSentence{
		Positive: []string{},
		Negative: []string{},
	}
	if *input.Negative != nil {
		for idx := range *input.Negative {
			sentence := (*input.Negative)[idx]
			intent.Sentences.Negative = append(intent.Sentences.Negative, sentence.Content)
		}
	}
	if *input.Positive != nil {
		for idx := range *input.Positive {
			sentence := (*input.Positive)[idx]
			intent.Sentences.Positive = append(intent.Sentences.Positive, sentence.Content)
		}
	}
	// At least add intent itself as positive sentence to avoid error in intent trainer
	if len(intent.Sentences.Positive) == 0 {
		intent.Sentences.Positive = append(intent.Sentences.Positive, input.Name)
	}
	intent.Features = &TrainFeature{}
	intent.Features.Init()
}

type TrainDict struct {
	ClassName []string `json:"class_name"`
	DictName  string   `json:"dict_name"`
	Words     []string `json:"words"`
}

type TrainSentence struct {
	Positive []string `json:"positive"`
	Negative []string `json:"negative"`
}

type TrainFeature struct {
	Classes interface{}    `json:"classes"`
	Exacts  []interface{}  `json:"exacts"`
	Blacks  []interface{}  `json:"blacks"`
	Rules   []*FeatureRule `json:"rules"`
}

func (feature *TrainFeature) Init() {
	feature.Classes = struct{}{}
	feature.Exacts = []interface{}{}
	feature.Blacks = []interface{}{}
	feature.Rules = []*FeatureRule{}
}

type FeatureRule struct {
	Rule         string        `json:"rule"`
	Black        []interface{} `json:"black"`
	ReturnValues interface{}   `json:"return_values"`
}
