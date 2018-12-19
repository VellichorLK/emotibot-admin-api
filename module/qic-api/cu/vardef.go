package cu

//the Conversation type
const (
	AudioFile = iota
	Flow
)

//the speaker
const (
	ChannelSilence = iota
	ChannelHost
	ChannelGuest
)

//the speaker in wording
const (
	WordHost    = "host"
	WordGuest   = "guest"
	WordSilence = "silence"
)

//Table name in QISYS
const (
	Conversation = "Conversation"
	TableSegment = "Segment"
)

//field name in Conversation table
const (
	ConFieldID          = "call_id"
	ConFieldStatus      = "status"
	ConFieldFileName    = "file_name"
	ConFieldPath        = "file_path"
	ConFieldVoiceID     = "voice_id"
	ConFieldCallComment = "call_comment"
	ConFieldTransaction = "transaction"
	ConFieldSeries      = "series"
	ConFieldCallTime    = "call_time"
	ConFieldUpdateTime  = "update_time"
	ConFieldUploadTime  = "upload_time"

	ConFieldHostID     = "host_id"
	ConFieldHostName   = "host_name"
	ConFieldExtenstion = "extension"
	ConFieldDepartment = "department"
	ConFieldGuestID    = "guest_id"
	ConFieldGuestName  = "guest_name"
	ConFieldGuestPhone = "guest_phone"

	ConFieldUUID       = "call_uuid"
	ConFieldEnterprise = "enterprise"
	ConFieldUser       = "user"
	ConFieldDuration   = "duration"

	ConFieldApplyGroup = "apply_group_list"

	ConFieldType         = "type"
	ConFieldLeftChannel  = "left_channel"
	ConFieldRightChannel = "right_channel"
)

//field name in Segment table
const (
	SegFieldID         = "segment_id"
	SegFieldCallID     = "call_id"
	SegFieldStartTime  = "start_time"
	SegFieldEndTime    = "end_time"
	SegFieldChannel    = "channel"
	SegFieldCreateTiem = "create_time"
	SegFieldAsrText    = "asr_text"
)
