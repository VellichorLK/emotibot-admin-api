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
	Conversation    = "Conversation"
	TableSegment    = "Segment"
	tblGroup        = "Group"
	tblRelGrpRule   = "Relation_Group_Rule"
	tblRelRuleLogic = "Relation_Rule_Logic"
	tblRule         = "Rule"
	tblLogic        = "Logic"
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

const (
	fldTagID = "tag_id"
)

const (
	fldRuleID          = "rule_id"
	fldRuleIsDelete    = "is_delete"
	fldRuleName        = "rule_name"
	fldRuleMethod      = "method"
	fldRuleScore       = "score"
	fldRuleDescription = "description"
	fldRuleEnterprise  = "enterprise"
)

const (
	fldLogicID              = "logic_id"
	fldLogicName            = "name"
	fldLogicTagDist         = "tag_distance"
	fldLogicRangeConstraint = "range_constraint"
	fldLogicCreateTime      = "create_time"
	fldLogicUpdateTime      = "update_time"
	fldLogicIsDelete        = "is_delete"
	fldLogicEnterprise      = "enterprise"
	fldLogicSpeaker         = "speaker"
)
