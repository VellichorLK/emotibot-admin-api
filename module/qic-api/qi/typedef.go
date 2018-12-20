package qi

type SimpleGroup struct {
	ID int64 `json:"group_id"`
	Name string `json:"group_name"`
}
type Group struct {
	ID int64 `json:"group_id,omitempty"`
	Name string `json:"group_name,omitempty"`
	Enterprise string `json:",omitempty"`
	Enabled int `json:is_enable,omitempty`
	Speed float64 `json:"limit_speed,omitempty"`
	SlienceDuration float64 `json:"limit_silence,omitempty"`
	Condition *GroupCondition `json:"other,omitempty"`
}

type GroupCondition struct {
	FileName string `json:"file_name"`
	CallDuration int64 `json:"call_time"`
	CallComment string `json:"call_comment"`
	Deal int `json:"transcation"`
	Series string `json:"series"`
	StaffID string `json:"host_id"`
	StaffName string `json:"host_name"`
	Extension string `json:"extension"`
	Department string `json:"department"`
	ClientID string `json:"guest_id"`
	ClientName string `json:"guest_name"`
	ClientPhone string `json:"guest_phone"`
	LeftChannel string `json:"left_channel"`
	LeftChannelCode int
	RightChannel string `json:"right_channel"`
	RightChannelCode int
	CallStart int64 `json:"call_from"`
	CallEnd int64 `json:"call_end"`
}