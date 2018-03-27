package Dictionary

import "time"

const (
	StatusRunning Status = "running"
	StatusFinish  Status = "finish"
	StatusFail    Status = "fail"
)

type Status string

type StatusInfo struct {
	Status    string    `json:"status"`
	StartTime time.Time `json:"start_time"`
	Message   *string   `json:"message"`
}

// DownloadMeta is the struct describe file which can be downloaded
type DownloadMeta struct {
	UploadTime time.Time `json:"time"`
	UploadFile string    `json:"filename"`
}

type WordBank struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	// 0: directory, 1:wordbank
	Type         int         `json:"type"`
	Children     []*WordBank `json:"children"`
	SimilarWords string      `json:"similar_words,omitempty"`
	Answer       string      `json:"answer,omitempty"`
}
