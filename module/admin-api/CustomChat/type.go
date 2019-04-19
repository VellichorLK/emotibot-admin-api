package CustomChat


type Answer struct {
	ID             int64          `json:"id"`
	QID            int64          `json:"qid"`
	Content        string         `json:"content"`
	Status         int         `json:"status"`
}

type Extend struct {
	ID             int64          `json:"id"`
	QID            int64          `json:"qid"`
	Content        string         `json:"content"`
	Status         int         `json:"status"`
}

type Question struct {
	ID            int64          `json:"id"`
	AnswerCount   int            `json:"answercount"`
	ExtendCount   int            `json:"extendcount"`
	Content       string         `json:"content"`
	Answers       []Answer       `json:"answers,omitempty"`
	Extends        []Extend       `json:"extend,omitempty"`
	Category      string         `json:"category"`
	Status        int         `json:"status"`
}

type CustomQuestions struct {
	QuestionCount   int            `json:"questioncount"`
	Questions       []Question     `json:"questions,omitempty"`
	Category        string         `json:"category"`
}

type ChatAnswerTagging struct {
	AnswerID     string                 `json:"answer_id"`
	Answer       string                 `json:"content"`
	Segment      string                 `json:"answer_seg"`
	WordPos      string                 `json:"answer_word_pos"`
	Keyword      string                 `json:"answer_keyword"`
	SentenceType string                 `json:"answer_sentence_type"`
}

type ChatQuestionTagging struct {
	QuestionID   string                 `json:"id"`
	Question     string                 `json:"question"`
	Segment      string                 `json:"question_seg"`
	WordPos      string                 `json:"question_word_pos"`
	Keyword      string                 `json:"question_keyword"`
	SentenceType string                 `json:"question_sentence_type"`
	Answers      []*ChatAnswerTagging   `json:"answers"`
	AppID        string                 `json:"app_id"`
}
