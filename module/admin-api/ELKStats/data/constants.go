package data

const (
	// AuthorizationHeaderKey is header used for auth, content will be app_id only
	AuthorizationHeaderKey = "Authorization"

	// UserIDHeaderKey is user_id in request header
	UserIDHeaderKey = "X-UserID"

	// UserIPHeaderKey is user IP in request header
	UserIPHeaderKey = "X-Real-IP"

	// EnterpriseIDHeaderKey is Enterprise ID in requqest header
	EnterpriseIDHeaderKey = "X-EnterpriseID"

	// AppIDHeaderKey is App ID in requqest header
	AppIDHeaderKey = "X-AppID"
)

const ESRecordsMappingFile = "./configs/records_mapping.json"
const ESRecordsIndex = "records"
const ESRecordType = "doc"

const ESSessionsMappingFile = "./configs/sessions_mapping.json"
const ESSessionsIndex = "sessions"
const ESSessionsType = "doc"

const ESTimeFormat = "2006-01-02 15:04:05"

const ESTermAggSize = 3000000
const ESTermAggShardSize = 3000000

const LogTimeFieldName = "log_time"
const LogTimeFormat = "2006-01-02T15:04:05.000Z"

const SessionEndTimeFieldName = "end_time"

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
