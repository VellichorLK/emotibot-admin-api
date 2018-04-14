package Dictionary

import (
	"fmt"
	"strings"
	"time"
)

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
	ID   *int   `json:"id,omitempty"`
	Name string `json:"name"`
	// 0: directory, 1:wordbank
	Type         int         `json:"type"`
	Children     []*WordBank `json:"children"`
	SimilarWords string      `json:"similar_words,omitempty"`
	Answer       string      `json:"answer,omitempty"`
}

type WordBankRow struct {
	Level1       string
	Level2       string
	Level3       string
	Level4       string
	Name         string
	SimilarWords string
	Answer       string
}

func (row WordBankRow) ToString() string {
	similars := strings.Split(strings.Replace(row.SimilarWords, "，", ",", -1), ",")
	trimSimilars := make([]string, len(similars))
	for idx := range trimSimilars {
		trimSimilars[idx] = strings.TrimSpace(similars[idx])
	}
	return fmt.Sprintf("%s>%s>%s>%s\t%s\t%s",
		row.Level1, row.Level2, row.Level3, row.Level4, row.Name,
		strings.Join(trimSimilars, "\t"))
}
