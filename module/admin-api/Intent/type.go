package Intent

type Intent struct {
	ID        int      `json:"id"`
	AppID     string   `json:"app_id"`
	Name      string   `json:"name"`
	Sentences []string `json:"sentences"`
}

func NewIntent() Intent {
	return Intent{
		Sentences: make([]string, 0),
	}
}

// Responses

type UploadIntentsResponse struct {
	Version int `json:"version"`
}

type TrainResponse struct {
	Status  string `json:"status"`
	ModelID string `json:"model_id"`
}

type IntentEngineGetDataResponse struct {
	Status     string                `json:"status"`
	AppID      string                `json:"app_id"`
	Intent     []*IntentResponse     `json:"intent"`
	IntentDict []*IntentDictResponse `json:"intent_dict"`
}

func NewIntentEngineGetDataResponse() IntentEngineGetDataResponse {
	return IntentEngineGetDataResponse{
		Intent:     make([]*IntentResponse, 0),
		IntentDict: make([]*IntentDictResponse, 0),
	}
}

type IntentResponse struct {
	Name      string                   `json:"intent_name"`
	Sentences *IntentSentencesResponse `json:"sentences"`
	Features  IntentFeaturesResponse   `json:"features"`
}

type IntentDictResponse struct {
	ClassName []string `json:"class_name"`
	DictName  string   `json:"dict_name"`
	Words     []string `json:"words"`
}

func NewIntentDictResponse() IntentDictResponse {
	return IntentDictResponse{
		ClassName: make([]string, 0),
		Words:     make([]string, 0),
	}
}

type IntentSentencesResponse struct {
	Positive []string `json:"positive"`
	Negative []string `json:"negative"`
}

func NewIntentSentencesResponse() IntentSentencesResponse {
	return IntentSentencesResponse{
		Positive: make([]string, 0),
		Negative: make([]string, 0),
	}
}

type IntentFeaturesResponse struct {
	Classes interface{}   					`json:"classes"`
	Exacts  []interface{} 					`json:"exacts"`
	Blacks  []interface{} 					`json:"blacks"`
	Rules   []IntentFeatureRulesResponse 	`json:"rules"`
}

func NewIntentFeaturesResponse() IntentFeaturesResponse {
	return IntentFeaturesResponse{
		Classes: struct{}{},
		Exacts: make([]interface{}, 0),
		Blacks: make([]interface{}, 0),
		Rules: make([]IntentFeatureRulesResponse, 0),
	}
}

type IntentFeatureRulesResponse struct {
	Rule         string        `json:"rule"`
	Black        []interface{} `json:"black"`
	ReturnValues interface{}   `json:"return_values"`
}

type RuleEngineGetDataResponse struct {
	Status string                    `json:"status"`
	AppID  string                    `json:"app_id"`
	Dict   []*RuleEngineDictResponse `json:"dict"`
}

func NewRuleEngineGetDataResponse() RuleEngineGetDataResponse {
	return RuleEngineGetDataResponse{
		Dict: make([]*RuleEngineDictResponse, 0),
	}
}

type RuleEngineDictResponse struct {
	ClassName     []string      `json:"class_name"`
	DictName      string        `json:"dict_name"`
	Words         []string      `json:"words"`
	PositiveRules []interface{} `json:"positive_rules"`
	NegativeRules []interface{} `json:"negative_rules"`
}

func NewRuleEngineDictResponse() RuleEngineDictResponse {
	return RuleEngineDictResponse{
		ClassName:     make([]string, 0),
		Words:         make([]string, 0),
		PositiveRules: make([]interface{}, 0),
		NegativeRules: make([]interface{}, 0),
	}
}

type IntentEngineStatusResponse struct {
	Status string `json:"status"`
}

type RuleEngineStatusResponse struct {
	Status string `json:"status"`
}

type StatusResponse struct {
	IntentEngineStatus string `json:"ie_status"`
	RuleEngineStatus   string `json:"re_status"`
}
