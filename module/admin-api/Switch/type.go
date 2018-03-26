package Switch

import "time"

type SwitchInfo struct {
	ID         int       `json:"id"`
	Code       string    `json:"code"`
	Name       string    `json:"name"`
	Status     int       `json:"status"`
	Remark     string    `json:"remark"`
	Scenario   string    `json:"scenario"`
	NumType    string    `json:"num_type"`
	Num        int       `json:"num"`
	Msg        string    `json:"msg"`
	Flow       int       `json:"flow"`
	WhiteList  string    `json:"white_list"`
	BlackList  string    `json:"black_list"`
	UpdateTime time.Time `json:"update_time"`
}
