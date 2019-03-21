package model

import "errors"

// ErrAutoIDDisabled will only be returned when Insert operation can not found the lastInsertId.
var ErrAutoIDDisabled = errors.New("dao does not support LastInsertId function")

//Table name in QISYS
const (
	tblConversation     = "Conversation"
	tblSegment          = "Segment"
	tblSegmentEmotion   = "SegEmotionScore"
	tblRelGrpRule       = "Relation_RuleGroup_Rule"
	tblRelRuleLogic     = "Relation_Rule_Logic"
	tblRule             = "Rule"
	tblLogic            = "Logic"
	tblCUPredict        = "CUPredict"
	tblRecommend        = "Recommendations"
	tblSentence         = "Sentence"
	tbleRelSentenceTag  = "Relation_Sentence_Tag"
	tblRelSenTag        = "Relation_Sentence_Tag"
	tblRuleGroup        = "RuleGroup"
	tblRGC              = "RuleGroupCondition"
	tblTags             = "Tag"
	tblSetnenceGroup    = "SentenceGroup"
	tblRelSGS           = "Relation_SentenceGroup_Sentence"
	tblConversationflow = "ConversationFlow"
	tblRelCFSG          = "Relation_ConversationFlow_SentenceGroup"
	tblConversationRule = "Rule"
	tblRelCRCF          = "Relation_Rule_ConversationFlow"
	tblInspectTask      = "InspectorTask"
	tblRelITOutline     = "Relation_InspectorTask_Outline"
	tblCall             = "call"
	tblRelCallRuleGrp   = "Relation_Call_RuleGroup"
	tblScoreForm        = "ScoreForm"
	tblOutline          = "Outline"
	tblUsers            = "users"
	tblRelITCallStaff   = "Relation_InspectorTask_Call_Staff"
	tblTask             = "task"
	tblRelITStaff       = "Relation_InspectorTask_Staff"
	tblCategory         = "Category"
	tblPredictResult    = "CUPredictResult"
	tblSegmentPredict   = "SegmentPredict"
	tblTrainedModel     = "TrainedModel"
	tblNavigation       = "Navigation"
	tblRelNavSenGrp     = "Relation_Nav_SenGrp"
	tblSilenceRule      = "SilenceRule"
	tblSpeedRule        = "SpeedRule"
	tblInterposalRule   = "InterposalRule"
	tblRelRGSilence     = "Relation_RuleGroup_Silence"
	tblRelRGInterposal  = "Relation_RuleGroup_Interposal"
	tblRelRGSpeed       = "Relation_RuleGroup_Speed"
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
	fldSegmentID         = "id"
	fldSegmentCallID     = "call_id"
	fldSegmentStartTime  = "start_time"
	fldSegmentEndTime    = "end_time"
	fldSegmentChannel    = "channel"
	fldSegmentCreateTime = "create_time"
	fldSegmentText       = "asr_text"
)

const (
	fldSegEmoID        = "id"
	fldSegEmoSegmentID = "segment_id"
	fldSegEmoType      = "type"
	fldSegEmoScore     = "score"
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
	fldCUPredict = "predict"
	fldAppID     = "app_id"
	fldSentence  = "sentence"
)

const (
	fldRuleID          = "rule_id"
	fldRuleIsDelete    = "is_delete"
	fldRuleName        = "rule_name"
	fldRuleMethod      = "method"
	fldRuleScore       = "score"
	fldRuleDescription = "description"
	fldRuleMin         = "min"
	fldRuleMax         = "max"
	fldRuleSeverity    = "severity"
	fldRuleCreateTime  = "create_time"
	fldRuleUpdateTime  = "update_time"
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
	fldCondID              = "id"
	fldCondGroupID         = "rg_id"
	fldCondType            = "type" // Toggle, if 1 then no condition consider, 0 condition on.
	fldCondFileName        = "file_name"
	fldCondDeal            = "deal"
	fldCondSeries          = "series"
	fldCondUploadTimeStart = "upload_time_start"
	fldCondUploadTimeEnd   = "upload_time_end"
	fldCondStaffID         = "staff_id"
	fldCondStaffName       = "staff_name"
	fldCondExtension       = "extension"
	fldCondDepartment      = "department"
	fldCondCustomerID      = "customer_id"
	fldCondCustomerName    = "customer_name"
	fldCondCustomerPhone   = "customer_phone"
	fldCondCallStart       = "call_start"
	fldCondCallEnd         = "call_end"
	fldCondLeftChanRole    = "left_channel_role"
	fldCondRightChanRole   = "right_channel_role"
)

// field name in RuleGroup
const (
	fldRuleGrpID           = "id"
	fldRuleGrpUUID         = "uuid"
	fldRuleGrpIsDelete     = "is_delete"
	fldRuleGrpName         = "name"
	fldRuleGrpEnterpriseID = "enterprise"
	fldRuleGrpDescription  = "description"
	fldRuleGrpCreateTime   = "create_time"
	fldRuleGrpUpdateTime   = "update_time"
	fldRuleGrpIsEnable     = "is_enable"
	fldRuleGrpLimitSpeed   = "limit_speed"
	fldRuleGrpLimitSilence = "limit_silence"
	fldRuleGrpType         = "type"
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
	SGOptional = "optional"
)

// field name in Relation_SentenceGroup_Sentece
const (
	RSGSSGID = "sg_id"
	RSGSSID  = "s_id"
)

//common field name
const (
	fldID          = "id"
	fldIsDelete    = "is_delete"
	fldName        = "name"
	fldEnterprise  = "enterprise"
	fldUUID        = "uuid"
	fldCreateTime  = "create_time"
	fldUpdateTime  = "update_time"
	fldDescription = "description"
	fldCreator     = "creator"
	fldStatus      = "status"
	fldType        = "type"
	fldCategoryID  = "category_id"
	fldMin         = "min"
	fldMax         = "max"
	fldNodeOrder   = "node_order"

	fldSilSecond   = "silence_sec"
	fldSilTime     = "silence_time"
	fldExcptBefore = "exception_before"
	fldExcptAfter  = "exception_after"

	fldExcptUnder = "exception_under"
	fldExcptOver  = "exception_over"

	fldOverLappedSec   = "overlapped_sec"
	fldOverLappedTimes = "overlapped_time"

	fldRGUUID         = "rg_uuid"
	fldInterposalUUID = "int_uuid"
	fldSpeedUUID      = "spe_uuid"
	fldSilenceUUID    = "sil_uuid"
)

//relation field name
const (
	fldRelTagID = "tag_id"
	fldRelSenID = "s_id"
)

// fields in ConversationFlow
const (
	CFType       = "type"
	CFExpression = "expression"
)

// fields in Relation_ConversationFlow_SentenceGroup
const (
	RCFSGCFID = "cf_id"
	RCFSGSGID = "sg_id"
)

// fields in Rule
const (
	CRMethod      = "method"
	CRScore       = "score"
	CRDescription = "description"
	CRMin         = "min"
	CRMax         = "max"
	CRSeverity    = "severity"
)

// fields in Relation_Rule_ConversationFlow
const (
	CRCFRID  = "rule_id"
	CRCFCFID = "cf_id"
)

// fields in InspectTask
const (
	ITCallStart         = "call_start"
	ITCallEnd           = "call_end"
	ITInspectPercentage = "inspector_percentage"
	ITInspectByPerson   = "inspector_byperson"
	ITPublishTime       = "publish_time"
	ITReviewPercentage  = "review_percentage"
	ITReviewByPerson    = "review_byperson"
	ITExcluedInspected  = "exclude_inspected"
	ITFormID            = "form_id"
)

// fields in Relation_InspectorTask_Outline
const (
	RITOTaskID     = "task_id"
	RITOTOutlineID = "outline_id"
)

const (
	fldCallID               = "call_id"
	fldCallUUID             = "call_uuid"
	fldCallFileName         = "file_name"
	fldCallFilePath         = "file_path"
	fldCallDescription      = "description"
	fldCallDuration         = "duration"
	fldCallUploadTime       = "upload_time"
	fldCallCallTime         = "call_time"
	fldCallStaffID          = "staff_id"
	fldCallStaffName        = "staff_name"
	fldCallExt              = "extension"
	fldCallDepartment       = "department"
	fldCallCustomerID       = "customer_id"
	fldCallCustomerName     = "customer_name"
	fldCallCustomerPhone    = "customer_phone"
	fldCallEnterprise       = "enterprise"
	fldCallUploadedUser     = "uploader"
	fldCallLeftSilenceTime  = "left_silence_time"
	fldCallRightSilenceTime = "right_silence_time"
	fldCallLeftSpeed        = "left_speed"
	fldCallRightSpeed       = "right_speed"
	fldCallType             = "type"
	fldCallLeftChan         = "left_channel"
	fldCallRightChan        = "right_channel"
	fldCallStatus           = "status"
	fldCallDemoFilePath     = "demo_file_path"
	fldCallTaskID           = "task_id"
	fldWhos                 = "whos"
)

// fields for Relation_Call_RuleGroup table
const (
	fldCRGRelID          = "id"
	fldCRGRelRuleGroupID = "rg_id"
	fldCRGRelCallID      = "call_id"
)

const (
	fldLinkID = "link_id"
)

// fields for Relation_InspectorTask_Call_Staff
const (
	RITCSTaskID  = "task_id"
	RITCSStaffID = "staff_id"
	RITCSCallID  = "call_id"
)

// fields for users
const (
	USERDisplayName = "display_name"
)
const (
	fldTaskID          = "task_id"
	fldTaskStatus      = "status"
	fldTaskDescription = "description"
	fldTaskDeal        = "deal"
	fldTaskSeries      = "series"
	fldTaskCreateTime  = "create_time"
	fldTaskUpdateTime  = "update_time"
)

// fields for Relation_InspectorTask_Staff
const (
	RITStaffTaskID  = "task_id"
	RITStaffStaffID = "staff_id"
)

const (
	fldParentID = "parent_id"
	fldOrgID    = "org_id"
	fldValid    = "valid"
	fldRevise   = "revise"
	fldScore    = "score"
)

const (
	fldSegID     = "segment_id"
	fldMatch     = "match"
	fldMatchText = "match_text"
)

const (
	fldIntenName   = "intent_name"
	fldIgoreIntent = "ignore_intent"
	fldIntentLink  = "intent_link_id"
)

const (
	fldNavID    = "nav_id"
	fldSenGrpID = "sg_id"
)
