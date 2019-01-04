package clustering

import (
	"errors"
	"time"

	"emotibot.com/emotigo/pkg/logger"
)

//ReportsService define the operations of the Reports
type ReportsService interface {
	NewReport(report Report) (uint64, error)
	GetReport(id uint64) (Report, error)
	QueryReports(query ReportQuery) ([]Report, error)
	UpdateReportStatus(id uint64, status ReportStatus) error
	NewReportError(err ReportError) (uint64, error)
}

//ReportClusterService define the operations of the clusters.
type ReportClustersService interface {
	NewCluster(clusters Cluster) (uint64, error)
	GetCluster(id uint64) (Cluster, error)
}

//ReportRecordsService define the operations of the report's record.
type ReportRecordsService interface {
	NewRecords(records ...ReportRecord) error
	GetRecords(reportID uint64) ([]ReportRecord, error)
}

//SimpleFTService define the operation of simpleFT
type SimpleFTService interface {
	//GetFTModel retrive the model_name by given app_id
	GetFTModel(appID string) (string, error)
}

// ErrNotAvailable is used for indicating out of resource, since clustering is a resource intensive operation.
var ErrNotAvailable = errors.New("clustering error: resource is not available yet. please try again")

const nonClusterID uint64 = 0

func newReportError(s ReportsService, errMsg string, id uint64) {
	newReportErrorWithStatus(s, errMsg, id, ReportStatusError)
}

func newReportErrorWithStatus(s ReportsService, errMsg string, id uint64, status ReportStatus) {
	logger.Error.Println("Report handle error: " + errMsg)
	if id > 0 {
		s.UpdateReportStatus(id, status)
	}
	err := ReportError{
		ReportID:   id,
		Cause:      errMsg,
		CreateTime: time.Now().Unix(),
	}
	s.NewReportError(err)
}
