package data

type ModuleHX struct {
	ID     int    `json:"id"`
	Code   string `json:"code"`
	Name   string `json:"name"`
	Status bool   `json:"status"`
	CmdList string `json:"cmdList"`
	Product int `json:"product"`
	Group int `json:"group"`
}

type ResultMap struct {
	ID int `json:"id"`
	Map map[string][]string `json:"map"`
}