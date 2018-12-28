package general

type Paging struct {
	Total int64 `json:"total"`
	Page  int   `json:"page"`
	Limit int   `sjon:"limit"`
}
