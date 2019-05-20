package CustomChat

import (
	"fmt"

	qaData "emotibot.com/emotigo/module/admin-api/QADoc/data"
)

type Answer struct {
	ID      int64  `json:"id"`
	QID     int64  `json:"qid"`
	Content string `json:"content"`
	Status  int    `json:"status"`
}

type Extend struct {
	ID      int64  `json:"id"`
	QID     int64  `json:"qid"`
	Content string `json:"content"`
	Status  int    `json:"status"`
}

type Question struct {
	ID          int64    `json:"id"`
	AnswerCount int      `json:"answercount"`
	ExtendCount int      `json:"extendcount"`
	Content     string   `json:"content"`
	Answers     []Answer `json:"answers,omitempty"`
	Extends     []Extend `json:"extend,omitempty"`
	Category    string   `json:"category"`
	Status      int      `json:"status"`
}

type CustomQuestions struct {
	QuestionCount int        `json:"questioncount"`
	Questions     []Question `json:"questions,omitempty"`
	Category      string     `json:"category"`
}

type ChatAnswerTagging struct {
	AnswerID     string `json:"answer_id"`
	Answer       string `json:"content"`
}

type ChatQuestionTagging struct {
	QuestionID   string               `json:"id"`
	Question     string               `json:"question"`
	Segment      string               `json:"question_seg"`
	WordPos      string               `json:"question_word_pos"`
	Keyword      string               `json:"question_keyword"`
	SentenceType string               `json:"question_sentence_type"`
	Answers      []*ChatAnswerTagging `json:"answers"`
	AppID        string               `json:"app_id"`
	StdQID       string               `json:"std_q_id"`
	StdQContent  string               `json:"std_q_content"`
}

func (tag *ChatQuestionTagging) convertToQACoreDocs(appID string) []*qaData.QACoreDoc {
	docs := []*qaData.QACoreDoc{}

	qDoc := &qaData.QACoreDoc {
		DocID:        createCustomChatQuestionDocID(appID, tag.QuestionID),
		AppID:        appID,
		Module:       "editroial_custom",
		Domain:       "",
		Sentence:     tag.Segment,
		SentenceOrig: tag.Question,
		SentenceType: tag.SentenceType,
		SentencePos:  tag.WordPos,
		Keywords:     tag.Keyword,
		StdQID:       tag.StdQID,
		StdQContent:  tag.StdQContent,
	}
	if len(tag.Answers) > 0 {
		// Question doc's answers
		answers := []*qaData.Answer{}

		for _, answer := range tag.Answers {
			ans := &qaData.Answer{
				Sentence: answer.Answer,
			}
			answers = append(answers, ans)
		}

		qDoc.Answers = answers
		docs = append(docs, qDoc)
	}
	return docs
}

type ChatQuestionTaggings []*ChatQuestionTagging

func (tags ChatQuestionTaggings) convertToQACoreDocs() []*qaData.QACoreDoc {
	docs := []*qaData.QACoreDoc{}

	for _, tag := range tags {
		docs = append(docs, tag.convertToQACoreDocs(tag.AppID)...)
	}

	return docs
}

func createCustomChatQuestionDocID(appID string, questionID string) string {
	return fmt.Sprintf("%s_editorial_custom_%s", appID, questionID)
}

func createCustomChatAnswerDocID(questionDocID string, answerID string) string {
	return fmt.Sprintf("%s#%s", questionDocID, answerID)
}
