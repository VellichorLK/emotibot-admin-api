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
	Segment      string `json:"answer_seg"`
	WordPos      string `json:"answer_word_pos"`
	Keyword      string `json:"answer_keyword"`
	SentenceType string `json:"answer_sentence_type"`
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
}

func (tag *ChatQuestionTagging) convertToQACoreDocs(appID string) []*qaData.QACoreDoc {
	docs := []*qaData.QACoreDoc{}

	qDoc := &qaData.QACoreDoc{
		DocID:        createCustomChatQuestionDocID(appID, tag.QuestionID),
		AppID:        appID,
		Module:       "other",
		Sentence:     tag.Segment,
		SentenceOrig: tag.Question,
		SentenceType: tag.SentenceType,
		SentencePos:  tag.WordPos,
		Keywords:     tag.Keyword,
	}

	docs = append(docs, qDoc)

	if len(tag.Answers) > 0 {
		// Question doc's answers
		answers := []*qaData.Answer{}

		for _, answer := range tag.Answers {
			ans := &qaData.Answer{
				Sentence: answer.Answer,
			}
			answers = append(answers, ans)

			ansDoc := &qaData.QACoreDoc{
				DocID:        createCustomChatAnswerDocID(qDoc.DocID, answer.AnswerID),
				Sentence:     answer.Segment,
				SentenceOrig: answer.Answer,
				SentenceType: answer.SentenceType,
				Keywords:     answer.Keyword,
			}
			docs = append(docs, ansDoc)
		}

		qDoc.Answers = answers
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
	return fmt.Sprintf("%s_other_%s", appID, questionID)
}

func createCustomChatAnswerDocID(questionDocID string, answerID string) string {
	return fmt.Sprintf("%s#%s", questionDocID, answerID)
}
