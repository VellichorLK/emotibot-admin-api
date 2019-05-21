package Robot

import (
	"fmt"
	"time"

	qaData "emotibot.com/emotigo/module/admin-api/QADoc/data"
)

// FunctionInfo store info about robot's function
type FunctionInfo struct {
	Status bool `json:"status"`
}

// QAInfo store info about robot's qa pair
// First question in questions is main question
type QAInfo struct {
	ID               int       `json:"id"`
	Question         string    `json:"main_question"`
	RelatedQuestions []string  `json:"relate_questions"`
	Answers          []string  `json:"answers"`
	CreatedTime      time.Time `json:"created_time"`
}

// RetQAInfo is the struct in api return
type RetQAInfo struct {
	Count int       `json:"count"`
	Infos []*QAInfo `json:"qa_infos"`
}

// ChatInfo store info about robot chat setting
type ChatInfo struct {
	Type     int      `json:"type"`
	Contents []string `json:"contents"`
}

type ChatDescription struct {
	Type    int    `json:"type"`
	Name    string `json:"name"`
	Comment string `json:"comment"`
}

// ChatInfoInput is used when update robot chat setting
type ChatInfoInput struct {
	Type     int      `json:"type"`
	Contents []string `json:"contents"`
	Name     string   `json:"name"`
}

type Function struct {
	Code   string `json:"code"`
	Name   string `json:"name"`
	Active bool   `json:"active"`
	Remark string `json:"remark"`
	Intent string `json:"intent"`
}

type ChatQA struct {
	Question string   `json:"question"`
	Answers  []string `json:"answer"`
}

type ChatQAList struct {
	TotalQACnt int      `json:"totalQACnt"`
	ChatQAs    []ChatQA `json:"chatQAs"`
}

type SolrQueryResponse struct {
	ResponseHeader SolrResponseHeader `json:"responseHeader"`
	Response       SolrResponse       `json:"response"`
}

type SolrResponseHeader struct {
	Status int `json:"status"`
}
type SolrResponse struct {
	NumFound int      `json:"numFound"`
	QAs      []SolrQA `json:"docs"`
}
type SolrQA struct {
	Question string   `json:"sentence_original"`
	Answers  []string `json:"related_sentences"`
}

type ChatContentInfoV2 struct {
	ID      int    `json:"id"`
	Content string `json:"content"`
}

type ChatInfoV2 struct {
	Type     int                  `json:"type"`
	Name     string               `json:"name"`
	Comment  string               `json:"comment"`
	Contents []*ChatContentInfoV2 `json:"contents"`
	Limit    int                  `json:"limit"`
}

type InfoV3 struct {
	ID      int    `json:"id"`
	Content string `json:"content"`
}

type QAInfoV3 struct {
	ID               int       `json:"id"`
	Question         string    `json:"main_question"`
	RelatedQuestions []*InfoV3 `json:"relate_questions"`
	Answers          []*InfoV3 `json:"answers"`
}

// ManualAnswerTagging is used when updating robot profile
type ManualAnswerTagging struct {
	DocID        string `json:"answer_id"`
	Answer       string `json:"content"`
}

// ManualTagging is used when updating robot profile
type ManualTagging struct {
	DocID        string                 `json:"id"`
	Question     string                 `json:"question"`
	Segment      string                 `json:"question_seg"`
	WordPos      string                 `json:"question_word_pos"`
	Keyword      string                 `json:"question_keyword"`
	SentenceType string                 `json:"question_sentence_type"`
	Answers      []*ManualAnswerTagging `json:"answers"`
	AppID        string                 `json:"app_id"`
	StdQID       string                 `json:"std_q_id"`
	StdQContent  string                 `json:"std_q_content"`
}

func (tag *ManualTagging) convertToQACoreDocs(appID string) []*qaData.QACoreDoc {
	docs := []*qaData.QACoreDoc{}

	qDoc := &qaData.QACoreDoc{
		DocID:        createRobotQuestionDocID(appID, tag.DocID),
		AppID:        appID,
		Module:       "robot_custom",
		Domain:       "",
		Sentence:     tag.Segment,
		SentenceOrig: tag.Question,
		SentenceType: tag.SentenceType,
		SentencePos:  tag.WordPos,
		Keywords:     tag.Keyword,
		StdQID:       tag.StdQID,
		StdQContent:  tag.StdQContent,
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
		}

		qDoc.Answers = answers
	}

	return docs
}

type ManualTaggings []*ManualTagging

func (tags ManualTaggings) convertToQACoreDocs() []*qaData.QACoreDoc {
	docs := []*qaData.QACoreDoc{}

	for _, tag := range tags {
		docs = append(docs, tag.convertToQACoreDocs(tag.AppID)...)
	}

	return docs
}

func createRobotQuestionDocID(appID string, questionID string) string {
	return fmt.Sprintf("%s_robot_custom_%s", appID, questionID)
}

func createRobotAnswerDocID(questionDocID string, answerID string) string {
	return fmt.Sprintf("%s#%s", questionDocID, answerID)
}
