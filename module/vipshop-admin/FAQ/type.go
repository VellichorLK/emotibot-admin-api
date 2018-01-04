package FAQ

type SimilarQuestion struct {
	Content string `json:"content"`
	Id      string `json:sqid`
}

type SimilarQuestionReqBody struct {
	SimilarQuestions []SimilarQuestion `json:similarQuestions`
}
