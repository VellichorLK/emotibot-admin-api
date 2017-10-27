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
	Message   string    `json:"message"`
}

// DownloadMeta is the struct describe file which can be downloaded
type DownloadMeta struct {
	UploadTime time.Time `json:"time"`
	UploadFile string    `json:"filename"`
}
