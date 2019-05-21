package data

import (
	"time"
)

const (
	HourInSeconds   = 60 * 60
	MinuteInSeconds = 60
)

var DurationDay = time.Hour * 24

const (
	// AuthorizationHeaderKey is header used for auth, content will be app_id only
	AuthorizationHeaderKey = "Authorization"

	// UserIDHeaderKey is user_id in request header
	UserIDHeaderKey = "X-UserID"

	// UserIPHeaderKey is user IP in request header
	UserIPHeaderKey = "X-Real-IP"

	// AppIDHeaderKey is App ID in requqest header
	AppIDHeaderKey = "X-AppID"
)

const StandardTimeFormat = "2006-01-02 15:04:05"

var ESTimeFormat = StandardTimeFormat

const ESTermAggSize = 3000000
const ESTermAggShardSize = 3000000

const LogTimeFieldName = "log_time"
const LogTimeFormat = "2006-01-02T15:04:05.000Z"

const SessionStartTimeFieldName = "start_time"
const SessionEndTimeFieldName = "end_time"
const FirstLogTimeFieldName = "first_log_time"

const TERecordsTriggerTimeFieldName = "trigger_time"
const TERecordsFinishTimeFieldName = "finish_time"

const SessionRatingFieldName = "rating"
const SessionFeedbackFieldName = "feedback"
const SessionFeedbackTimeFieldName = "feedback_time"

const (
	IntervalYear   = "year"
	IntervalMonth  = "month"
	IntervalDay    = "day"
	IntervalHour   = "hour"
	IntervalMinute = "minute"
	IntervalSecond = "second"
)

const (
	AggByTime = "time"
	AggByTag  = "tag"
)

const ESMaxResultWindow = 10000

const ESScrollSize = 10000
const ESScrollKeepAlive = "30s"

const MaxNumRecordsPerXlsx = 100000

const (
	XlsxExportDir           = "exports"
	XlsxDirTimestampFormat  = "20060102"
	XlsxFileTimestampFormat = "20060102_150405"
)

const (
	CodeExportTaskRunning = iota
	CodeExportTaskCompleted
	CodeExportTaskFailed
	CodeExportTaskEmpty
)

var ExportTaskCodesMap = map[int]string{
	CodeExportTaskRunning:   "RUNNING",
	CodeExportTaskCompleted: "COMPLETED",
	CodeExportTaskFailed:    "FAILED",
	CodeExportTaskEmpty:     "EMPTY",
}
