package Task

import "emotibot.com/emotigo/module/admin-api/util"

type MapTuple struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type MapMeta struct {
	UpdateTime       string `json:"update_time"`
	UpdateUser       string `json:"update_user"`
	MappingTableName string `json:"mapping_table_name"`
}

type MappingTable struct {
	MappingData []*MapTuple `json:"mapping_table"`
	Metadata    *MapMeta    `json:"metadata"`
}

var SlotTypeMap = map[string]string{
	"整数": "integer",
	"姓氏": "last-name",
	"时间日期(粒度-时)(未来时间)": "time-hour-future",
}

type SpreadsheetEntity struct {
	EntityName     string `xlsx:"0"`
	EntityCategory string `xlsx:"1"`
	EntityTypt     string `xlsx:"2"`
	Prompt         string `xlsx:"3"`
}

func (s *SpreadsheetEntity) ToEntity() Entity {
	ner := Ner{
		EntityType:     s.EntityTypt,
		SlotType:       SlotTypeMap[s.EntityTypt],
		EntityCategory: s.EntityCategory,
		SourceType:     "system",
	}
	entity := Entity{
		EntityName:     s.EntityName,
		EntityCategory: s.EntityCategory,
		Prompt:         s.Prompt,
		ID:             util.GenRandomUUIDSameAsOpenAPI(),
		Required:       true,
		MustRetry:      true,
		RetryNum:       3,
		Ner:            &ner,
	}
	return entity
}

type Entity struct {
	EntityName     string `json:"entityName"`
	EntityCategory string `json:"entityCategory"`
	Prompt         string `json:"prompt"`
	ID             string `json:"id"`
	Required       bool   `json:"required"`
	MustRetry      bool   `json:"must_retry"`
	RetryNum       int    `json:"retry_num"`
	Ner            *Ner   `json:"ner"`
}

type Ner struct {
	EntityType     string `json:"entityType"`
	SlotType       string `json:"slotType"`
	EntityCategory string `json:"entityCategory"`
	SourceType     string `json:"sourceType"`
}

type Scenario struct {
	EditingContent *ScenarioContent        `json:"editingContent"`
	EditingLayout  map[string]*interface{} `json:"editingLayout"`
}

type ScenarioContent struct {
	Version    string                  `json:"version"`
	Metadata   map[string]string       `json:"metadata"`
	Setting    map[string]int          `json:"setting"`
	MsgConfirm []*interface{}          `json:"msg_confirm"`
	Nodes      []*interface{}          `json:"nodes"`
	IDToNerMap map[string]*interface{} `json:"idToNerMap"`
	Skills     map[string]*Skill       `json:"skills"`
}

type Skill struct {
	SkillName           string                  `json:"skillName"`
	TriggerList         []*interface{}          `json:"triggerList"`
	EntityCollectorList []*Entity               `json:"entityCollectorList"`
	ActionGroupList     []*interface{}          `json:"actionGroupList"`
	RelatedEntities     map[string]*interface{} `json:"relatedEntities"`
	ReParsers           []*interface{}          `json:"re_parsers"`
	TDESetting          map[string]*interface{} `json:"tde_setting"`
}
