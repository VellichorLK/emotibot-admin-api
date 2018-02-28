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
	Children []int
}

type Question struct {
	QuestionId int `json:"questionId"`
	SQuestionConunt int `json:"sQuesCount"`
	Content string `json:"questionContent"`
	CategoryName string `json:"categoryName"`
	CategoryId int `json:"categoryId"`
	Answers []Answer `json:"answerItem"`
}

type Answer struct {
	QuestionId int `json:"Question_Id"`
	AnswerId int `json:"Answer_Id"`
	Content string `json:"Content_String"`
	RelatedQuestion string `json:"RelatedQuestion"`
	DynamicMenu string `json:"DynamicMenu"`
	NotShow int `json:"Not_Show_In_Relative_Q"`
	BeginTime string `json:"Begin_Time"`
	EndTime string `json:"End_Time"`
	AnswerCmd string `json:"Answer_CMD"`
	AnswerCmdMsg string `json:"Answer_CMD_Msg"`
	Dimension []string `json:"dimension"`
}

type QueryCondition struct {
	TimeSet bool
	BeginTime string
	EndTime string
	Keyword string
	SearchQuestion bool
	SearchAnswer bool
	SearchDynamicMenu bool
	SearchRelativeQuestion bool
	NotShow bool
	Dimension []DimensionGroup
	CategoryId int
	Limit int
	CurPage int
}

type DimensionGroup struct {
	TypeId int `json:"typeId"`
	Content string `json:"tagContent"`
}

type Parameter interface {
	FormValue(name string) string
}

type Tag struct {
	Type int
	Content string
}
