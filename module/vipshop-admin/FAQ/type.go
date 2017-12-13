package FAQ

type SimilarQuestion struct {
	Content string `json:"content"`
	Id string `json:sqid`
}

type SimilarQuestionReqBody struct {
	User string `json:"user"`
	SimilarQuestions []SimilarQuestion `json:similarQuestions`
}