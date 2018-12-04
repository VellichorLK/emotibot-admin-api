package feedback

import "errors"

var (
	ErrDBNotInit = errors.New("DB is not init")
)

// Reason is a basic structure of feedback reason
type Reason struct {
	ID      int64  `json:"id"`
	Content string `json:"content"`
}
