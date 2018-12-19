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
	tblGroup     = "Group"
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

//field name in Group Table
const (
	fldGroupAppID          = "app_id"
	fldGroupIsDeleted      = "is_delete"
	fldGroupIsEnabled      = "is_enable"
	fldGroupName           = "group_name"
	fldGroupEnterprise     = "enterprise"
	fldGroupDescription    = "description"
	fldGroupCreatedTime    = "create_time"
	fldGroupUpdatedTime    = "update_time"
	fldGroupLimitedSpeed   = "limit_speed"
	fldGroupLimitedSilence = "limit_silence"
	fldGroupType           = "type"
)
