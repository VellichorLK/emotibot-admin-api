package workweixin

type WorkWeixinMessageType string

const (
	WorkWeixinMessageTypeText = "text"
)

type WorkWeixinMessage struct {
	To         string                `xml:"ToUserName"`
	From       string                `xml:"FromUserName"`
	CreateTime int                   `xml:"CreateTime"`
	Type       WorkWeixinMessageType `xml:"MsgType"`
}

type WorkWeixinTextMessage struct {
	WorkWeixinMessage
}
