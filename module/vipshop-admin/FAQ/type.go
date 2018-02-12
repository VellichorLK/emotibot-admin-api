package FAQ

type SimilarQuestion struct {
	Content string `json:"content"`
	Id      string `json:sqid`
}

type SimilarQuestionReqBody struct {
	SimilarQuestions []SimilarQuestion `json:similarQuestions`
}

//StdQuestion is a Standard Question in FAQ Table
type StdQuestion struct {
	QuestionID int    `json:"questionId"`
	Content    string `json:"content"`
	CategoryID int
}

//Category represents sql table vipshop_category
type Category struct {
	ID       int
	Name     string
	ParentID int
}
