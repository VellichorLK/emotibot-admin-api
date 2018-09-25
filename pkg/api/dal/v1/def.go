package dal

type request struct {
	Op           string `json:"op"`
	Category     string `json:"category"`
	AppID        string `json:"appid"`
	UserRecordID int64  `json:"userRecordId"`
	Data         data   `json:"data"`
}

type data struct {
	Subop      string      `json:"subop"`
	Conditions *conditions `json:"conditions,omitempty"`
	Entities   []entity    `json:"entities,omitempty"`
	Groupby    string      `json:"groupby,omitempty"`
	Algo       string      `json:"algo,omitempty"`
	StartIndex int64       `json:"startIndex,omitempty"`
	EndIndex   int64       `json:"endIndex,omitempty"`
	Segments   []string    `json:"segments,omitempty"`
}

type entity struct {
	Content       string `json:"content"`
	ParentContent string `json:"parentContent"`
}

type conditions struct {
	Labels        []string `json:"labels,omitempty"`
	Content       string   `json:"content,omitempty"`
	ParentContent string   `json:"parentContent,omitempty"`
}

// RawResponse is the general response retrive from dal api endpoint.
// It is exported only for test package to avoid duplicated struct.
type RawResponse struct {
	ErrNo      string        `json:"errno"`
	ErrMessage string        `json:"errmsg"`
	Results    []interface{} `json:"actualResults"`
	Operation  []string      `json:"results"`
}

//DetailError should contain detail dal response message for debug.
type DetailError struct {
	ErrMsg  string
	Results []string
}

func (e *DetailError) Error() string {
	return "dal error: " + e.ErrMsg
}
