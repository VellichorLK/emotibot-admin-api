package data

type Return struct {
	ReturnMessage string      `json:"ret_msg"`
	ReturnObj     interface{} `json:"result"`
	BFErrorCode   int         `json:"error_code"`
	BFErrorMsg    string      `json:"error_message"`
	BFData        interface{} `json:"data"`
}
