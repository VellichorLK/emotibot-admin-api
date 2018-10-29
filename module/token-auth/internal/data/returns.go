package data

type Return struct {
	ReturnMessage string      `json:"ret_msg"`
	ReturnObj     interface{} `json:"result"`
}

type BFReturn struct {
	ErrorCode int         `json:"error_code"`
	ErrorMsg  string      `json:"error_message"`
	Data      interface{} `json:"data"`
}
