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
