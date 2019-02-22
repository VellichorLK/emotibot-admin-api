package data

type IntentTestResults struct {
	Latest []*IntentTestResult `json:"latest"`
	Saved  []*IntentTestResult `json:"saved"`
}

type IntentTestResult struct {
	IntentTest  IntentTest  `json:"intent_test"`
	IntentModel IntentModel `json:"ie_model"`
}

type IntentTest struct {
	ID                 int64               `json:"id"`
	Name               string              `json:"name,omitempty"`
	UpdatedTime        int64               `json:"updated_time"`
	IEModelUpdateTime  int64               `json:"ie_model_updated_time,omitempty"`
	TestIntentsCount   int64               `json:"test_intents_count"`
	TestSentencesCount int64               `json:"test_sentences_count"`
	IntentsCount       *int64              `json:"intents_count,omitempty"`
	SentencesCount     *int64              `json:"sentences_count,omitempty"`
	TruePositives      int64               `json:"true_positives"`
	FalsePositives     int64               `json:"false_positives"`
	TrueNegatives      int64               `json:"true_negatives"`
	FalseNegatives     int64               `json:"false_negatives"`
	Saved              bool                `json:"saved"`
	Tester             string              `json:"tester"`
	TestIntents        []*IntentTestIntent `json:"test_intents,omitempty"`
}

type IntentModel struct {
	ID             int64 `json:"id"`
	UpdatedTime    int64 `json:"updated_time"`
	IntentsCount   int64 `json:"intents_count"`
	SentencesCount int64 `json:"sentences_count"`
}

type IntentTestIntent struct {
	ID             int64                 `json:"id"`
	IntentName     *string               `json:"name"`
	Version        *int64                `json:"-"`
	UpdatedTime    int64                 `json:"-"`
	SentencesCount int64                 `json:"sentences_count"`
	PositivesCount *int64                `json:"positives_count,omitempty"`
	Sentences      []*IntentTestSentence `json:"-"`
	Type           *bool                 `json:"type,omitempty"`
}

type IntentTestSentence struct {
	ID         int64   `json:"id"`
	TestIntent int64   `json:"-"`
	IntentName string  `json:"-"`
	Sentence   string  `json:"sentence"`
	Result     int64   `json:"result"`
	Score      *int64  `json:"score"`
	Answer     *string `json:"answer"`
	Message    *string `json:"message,omitempty"`
}

type IntentTestStatusResp struct {
	Version        int64 `json:"version"`
	Status         int64 `json:"status"`
	SentencesCount int64 `json:"sentences_count"`
	Progress       int64 `json:"progress"`
}

type TestResult struct {
	TruePositives  int64
	FalsePositives int64
	TrueNegatives  int64
	FalseNegatives int64
	Error          error
}

type UpdateCmd struct {
	ID      int64  `json:"id"`
	Content string `json:"content"`
}

type UseableModels struct {
	InUsed        *IEModel   `json:"in_used"`
	RecentTrained []*IEModel `json:"recent_trained"`
	RecentTested  []*IEModel `json:"recent_tested"`
	RecentSaved   []*IEModel `json:"recent_saved"`
}

func NewUseableModels() *UseableModels {
	return &UseableModels{
		RecentTrained: make([]*IEModel, 0),
		RecentTested:  make([]*IEModel, 0),
		RecentSaved:   make([]*IEModel, 0),
	}
}

type IEModel struct {
	IntentVersion  int64             `json:"version"`
	ModelID        string            `json:"ie_model_id"`
	TrainTime      int64             `json:"train_time"`
	IntentsCount   int64             `json:"intents_count"`
	SentencesCount int64             `json:"sentences_count"`
	Diffs          *TestIntentsDiffs `json:"diffs"`
	DiffsCount     int64             `json:"diffs_count"`
}

func NewIEModel() *IEModel {
	return &IEModel{
		Diffs: &TestIntentsDiffs{
			Intents:     make([]string, 0),
			TestIntents: make([]string, 0),
		},
	}
}

type TestIntentsDiffs struct {
	Intents     []string `json:"intents"`
	TestIntents []string `json:"test_intents"`
}
