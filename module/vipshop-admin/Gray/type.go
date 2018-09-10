package Gray

type White struct {
	UserId      int      `json:"userId"`
}

type QueryCondition struct {
	Keyword                string
	Limit                  int
	CurPage                int
}

type Parameter interface {
	FormValue(name string) string
}