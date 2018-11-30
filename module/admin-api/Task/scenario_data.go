package Task

// ScenarioContent define the structure of the content in taskenginescenario for TE 2.X
type ScenarioContent struct {
	Version  string          `json:"version"`
	Nodes    []ContentNode   `json:"nodes"`
	Metadata ContentMetadata `json:"metadata"`
	//	Setting    Setting       `json:"setting"`
	//	MsgConfirm []interface{} `json:"msg_confirm"`
}

// ContentMetadata define the structure of the metadata in content
type ContentMetadata struct {
	ScenarioName string `json:"scenario_name"`
	UpdateTime   string `json:"update_time"`
	UpdateUser   string `json:"update_user"`
	ScenarioID   string `json:"scenario_id"`
}

// ContentNode define the structure of the node in content
type ContentNode struct {
	NodeType string `json:"node_type"`
	NodeID   string `json:"node_id"`
	//Description string    `json:"description"`
	//Warnings    []Warning `json:"warnings"`
	//GlobalVars           []interface{}          `json:"global_vars"`
	//Edges                []Edge                 `json:"edges"`
	// EntryConditionRules [][]EntryConditionRule `json:"entry_condition_rules"`
	//NodeDialogueCntLimit *int64                 `json:"node_dialogue_cnt_limit,omitempty"`
	//Content              *NodeContent           `json:"content,omitempty"`
}

// ScenarioInfoListResponse define the return structure of the GetScenarioInfoList API
type ScenarioInfoListResponse struct {
	Msg []ScenarioInfo `json:"msg"`
}

// ScenarioInfo define the return structure of the scenario info
type ScenarioInfo struct {
	ScenarioID   string `json:"scenarioID"`
	ScenarioName string `json:"scenarioName"`
	Enable       bool   `json:"enable"`
	Version      string `json:"version"`
}
