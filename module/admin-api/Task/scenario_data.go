package Task

import "time"

// Scenario defines the column of the table: taskenginescenario
type Scenario struct {
	ScenarioID     string     `json:"scenarioID"`
	UserID         string     `json:"userID"`
	AppID          string     `json:"appID"`
	Content        string     `json:"content"`
	Layout         string     `json:"layout"`
	Public         int        `json:"public"`
	Editing        int        `json:"editing"`
	EditingContent string     `json:"editingContent"`
	EditingLayout  string     `json:"editingLayout"`
	Updatetime     *time.Time `json:"updatetime"`
	OnOff          int        `json:"onOff"`
}

// GetScenarioResponse defines the return structure of the GetScenario API
type GetScenarioResponse struct {
	Result *GetScenarioResult `json:"result"`
}

// GetScenarioResult defines the result structure in GetScenarioResponse
type GetScenarioResult struct {
	Content        string `json:"content"`
	Layout         string `json:"layout"`
	Editing        int    `json:"editing"`
	EditingContent string `json:"editingContent"`
	EditingLayout  string `json:"editingLayout"`
}

// ScenarioContent defines the structure of the content in taskenginescenario for TE 2.X
type ScenarioContent struct {
	Version  string           `json:"version"`
	Nodes    []*ContentNode   `json:"nodes"`
	Metadata *ContentMetadata `json:"metadata"`
	JSCode   *ContentJSCode   `json:"js_code"`
}

// InitialScenarioContent defines the structure of the initial content
type InitialScenarioContent struct {
	Metadata *ContentMetadata `json:"metadata"`
}

// ContentJSCode defines the structure of the js_code in content
type ContentJSCode struct {
	Alias    []string `json:"alias"`
	TextType string   `json:"text_type"`
	Main     string   `json:"main"`
}

// ContentMetadata defines the structure of the metadata in content
type ContentMetadata struct {
	ScenarioName string `json:"scenario_name"`
	UpdateTime   string `json:"update_time"`
	UpdateUser   string `json:"update_user"`
	ScenarioID   string `json:"scenario_id"`
}

// ContentNode defines the structure of the node in content
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

// CreateScenarioResponse defines the return structure of the CreateInitialScenario API
type CreateScenarioResponse struct {
	Template   *TemplateResult `json:"template"`
	ScenarioID string          `json:"scenarioID"`
}

// TemplateResult defines the template result structure in CreateScenarioResponse
type TemplateResult struct {
	Metadata *ContentMetadata `json:"metadata"`
}

// ScenarioInfoListResponse defines the return structure of the GetScenarioInfoList API
type ScenarioInfoListResponse struct {
	Msg []*ScenarioInfo `json:"msg"`
}

// TemplateScenarioInfoListResponse defines the return structure of the GetTemplateScenarioInfoList API
type TemplateScenarioInfoListResponse struct {
	Result []*ScenarioInfo `json:"result"`
}

// ScenarioInfo defines the return structure of the scenario info
type ScenarioInfo struct {
	ScenarioID   string `json:"scenarioID"`
	ScenarioName string `json:"scenarioName"`
	Enable       bool   `json:"enable"`
	Version      string `json:"version"`
}

// ResultMsgResponse defines the return structure of result message
type ResultMsgResponse struct {
	Msg string `json:"msg"`
}
