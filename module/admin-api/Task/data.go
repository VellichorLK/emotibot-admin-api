package Task

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

// TEConfig defines the structure of TE config json
type TEConfig struct {
	System     string      `json:"system"`
	TEv2Config *TEv2Config `json:"task_engine_v2"`
}

// TEv2Config defines the structure of TEv2 config json
type TEv2Config struct {
	EnableJSCode bool        `json:"enable_js_code"`
	EnableNode   *EnableNode `json:"enable_node"`
}

// EnableNode defines the structure of the enabled nodes in TEv2 config json
type EnableNode struct {
	NluPc  bool `json:"nlu_pc"`
	Action bool `json:"action"`
}
