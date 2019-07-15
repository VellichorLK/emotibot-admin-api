package BF

import "time"

type CommandRecord struct {
	Class             string               `json:"class"`
	Name              string               `json:"name"`
	Target            string               `json:"target"`
	Tags              string               `json:"tags"`
	Keywords          string               `json:"keywords"`
	Regex             string               `json:"regex"`
	Period            string               `json:"period"`
	Answer            string               `json:"answer"`
	ResponseType      string               `json:"response_type"`
	CheckStatus       RecordStatus         `json:"status"`
}

type RecordStatus struct {
	Status            int                  `json:"status"`
	Content           string               `json:"content"`
}

type CommandObj struct {
	Id                int                  `json:"id"`
	ClassId           int                  `json:"cid"`
	Name              string               `json:"name"`
	Target            int                  `json:"target"`
	Tags              []int                `json:"tags"`
	Rule              string               `json:"rule"`
	BeginTime         *time.Time           `json:"begin_time"`
	EndTime           *time.Time           `json:"end_time"`
	Answer            string               `json:"answer"`
	ResponseType      int                  `json:"response_type"`
	Status            int                  `json:"status"`
}

type CmdRule struct {
	Type              string                `json:"type"`
	Value             []string              `json:"value"`
}

type CmdTag struct {
	CmdId             int                `json:"cmd_id"`
	RobotTagId        int              `json:"robot_tag_id"`
}