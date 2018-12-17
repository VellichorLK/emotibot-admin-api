package cu

//the Conversation type
const (
	AudioFile = iota
	Flow
)

//the speaker
const (
	SpeakerDefault = iota
	Host
	Guest
)

//Table name in QISYS
const (
	Conversation = "Conversation"
)

//field name in Conversation table
const (
	ConFieldCallTime     = "call_time"
	ConFieldUpdateTime   = "update_time"
	ConFieldUploadTime   = "upload_time"
	ConFieldType         = "type"
	ConFieldLeftChannel  = "left_channel"
	ConFieldRightChannel = "right_channel"
	ConFieldEnterprise   = "enterprise"
	ConFieldFileName     = "file_name"
	ConFieldUUID         = "call_uuid"
	ConFieldUser         = "user"
)
