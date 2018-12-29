package model

import "errors"

// ErrAutoIDDisabled will only be returned when Insert operation can not found the lastInsertId.
var ErrAutoIDDisabled = errors.New("dao does not support LastInsertId function")

//Table name in QISYS
const (
	tblConversation    = "Conversation"
	tblSegment         = "Segment"
	tblGroup           = "Group"
	tblRelGrpRule      = "Relation_Group_Rule"
	tblRelRuleLogic    = "Relation_Rule_Logic"
	tblRule            = "Rule"
	tblLogic           = "Logic"
	tblCUPredict       = "CUPredict"
	tblRecommend       = "Recommendations"
	tblSentence        = "Sentence"
	tbleRelSentenceTag = "Relation_Sentence_Tag"
	tblRelSenTag       = "Relation_Sentence_Tag"
	tblRuleGroup       = "RuleGroup"
	tblRGC             = "RuleGroupCondition"
	tblTags            = "Tag"
	tblSetnenceGroup   = "SentenceGroup"
	tblRelSGS          = "Relation_SentenceGroup_Sentence"
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

//field name in Tag Table
const (
	fldTagID         = "id"
	fldTagUUID       = "uuid"
	fldTagIsDeleted  = "is_delete"
	fldTagName       = "name"
	fldTagType       = "type"
	fldTagPosSen     = "pos_sentences"
	fldTagNegSen     = "neg_sentences"
	fldTagCreateTime = "create_time"
	fldTagUpdateTime = "update_time"
	fldTagEnterprise = "enterprise"
)
const (
	fldCallID    = "call_id"
	fldCUPredict = "predict"
	fldSentence  = "sentence"
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

// field name in RuleGroupCondition table
const (
	RGCID            = "id"
	RGCGroupID       = "rg_id"
	RGCType          = "type"
	RGCFileName      = "file_name"
	RGCDeal          = "deal"
	RGCSeries        = "series"
	RGCUploadTime    = "upload_time"
	RGCStaffID       = "staff_id"
	RGCStaffName     = "staff_name"
	RGCExtension     = "extension"
	RGCDepartment    = "department"
	RGCCustomerID    = "customer_id"
	RGCCustomerName  = "customer_name"
	RGCCustomerPhone = "customer_phone"
	RGCCategory      = "category"
	RGCCallStart     = "call_start"
	RGCCallEnd       = "call_end"
	RGCLeftChannel   = "left_channel"
	RGCRightChannel  = "right_channel"
)

// field name in RuleGroup
const (
	RGID           = "id"
	RGUUID         = "uuid"
	RGIsDelete     = "is_delete"
	RGName         = "name"
	RGEnterprise   = "enterprise"
	RGDescription  = "description"
	RGCreateTime   = "create_time"
	RGUpdateTime   = "update_time"
	RGIsEnable     = "is_enable"
	RGLimitSpeed   = "limit_speed"
	RGLimitSilence = "limit_silence"
	RGType         = "type"
)

// field name Relation_RuleGroup_Rule
const (
	RRRGroupID = "rg_id"
	RRRRuleID  = "rule_id"
)

// field name in SentenceGroup
const (
	SGRole     = "role"
	SGPoistion = "position"
	SGRange    = "range"
)

// field name in Relation_SentenceGroup_Sentece
const (
	RSGSSGID = "sg_id"
	RSGSSID  = "s_id"
)

//common field name
const (
	fldID         = "id"
	fldIsDelete   = "is_delete"
	fldName       = "name"
	fldEnterprise = "enterprise"
	fldUUID       = "uuid"
	fldCreateTime = "create_time"
	fldUpdateTime = "update_time"
)

//relation field name
const (
	fldRelTagID = "tag_id"
	fldRelSenID = "s_id"
)
