package data

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

const ESRecordsTemplate = "emotibot_records_template"
const ESRecordsTemplateFile = "./InitFiles/elasticsearch/configs/records_template.json"
const ESRecordsIndex = "emotibot-records"
const ESRecordType = "doc"

const ESSessionsTemplate = "emotibot_sessions_template"
const ESSessionsTemplateFile = "./InitFiles/elasticsearch/configs/sessions_template.json"
const ESSessionsIndex = "emotibot-sessions"
const ESSessionsType = "doc"

const ESUsersTemplate = "emotibot_users_template"
const ESUsersTemplateFile = "./InitFiles/elasticsearch/configs/users_template.json"
const ESUsersIndex = "emotibot-users"
const ESUsersType = "doc"

const ESTimeFormat = "2006-01-02 15:04:05"

const ESTermAggSize = 3000000
const ESTermAggShardSize = 3000000

const LogTimeFieldName = "log_time"
const LogTimeFormat = "2006-01-02T15:04:05.000Z"

const SessionEndTimeFieldName = "end_time"
const FirstLogTimeFieldName = "first_log_time"

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
