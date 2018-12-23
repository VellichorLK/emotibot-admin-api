package datatype

import "encoding/base64"

type WorkWeixinRequest struct {
	User    string `xml:"ToUserName"`
	Agent   string `xml:"AgentID"`
	Encrypt string `xml:"Encrypt"`
}

type WorkWeiXinConfig struct {
	CorpID      string
	Token       string
	EncodingAES string
	AESKey      []byte
}

func (c *WorkWeiXinConfig) Load(config map[string]string) {
	c.CorpID = config["corpid"]
	c.Token = config["token"]
	c.EncodingAES = config["encodedaes"]
	c.AESKey, _ = base64.StdEncoding.DecodeString(c.EncodingAES)
}

func (c *WorkWeiXinConfig) IsValid() bool {
	var err error
	c.AESKey, err = base64.StdEncoding.DecodeString(c.EncodingAES)
	if err != nil {
		return false
	}
	return c.CorpID != "" && c.Token != "" && c.EncodingAES != ""
}
