package workweixin

import "errors"

// MessageType is use to tell the type of input message type
type MessageType string

const (
	MessageTypeText     MessageType = "text"
	MessageTypeImage    MessageType = "image"
	MessageTypeVoice    MessageType = "voice"
	MessageTypeVedio    MessageType = "vedio"
	MessageTypeLocation MessageType = "location"
	MessageTypeLink     MessageType = "link"
	MessageTypeFile     MessageType = "file"
)

const (
	MsgSendURL       = "https://qyapi.weixin.qq.com/cgi-bin/message/send"
	TokenValidateURL = "https://open.work.weixin.qq.com/devtool/getInfoByAccessToken"
	TokenIssueURL    = "https://qyapi.weixin.qq.com/cgi-bin/gettoken"
	MediaUploadURL   = "https://qyapi.weixin.qq.com/cgi-bin/media/upload"
)

var (
	ErrInvalidSignature = errors.New("Invalid signature")
)

type rawMessage struct {
	To         string      `xml:"ToUserName,omitempty"`
	From       string      `xml:"FromUserName,omitempty"`
	CreateTime int64       `xml:"CreateTime,omitempty"`
	Type       MessageType `xml:"MsgType,omitempty"`
	AgentID    int         `xml:"AgentID,omitempty"`
	MsgID      string      `xml:"MsgId,omitempty"`
}

type Message interface {
}

type Input struct {
	To        string `xml:"ToUserName"`
	AgentID   string `xml:"AgentID"`
	Encrypted string `xml:"Encrypt"`
}

type Output struct {
	Encrypted string `xml:"Encrypt"`
	Signature string `xml:"MsgSignature"`
	Timestamp int64  `xml:"TimeStamp"`
	Nonce     string `xml:"nonce"`
}

type TextMessage struct {
	Message
	rawMessage
	Content string `xml:"Content"`
}

type ImageMessage struct {
	Message
	rawMessage
	PicURL  string `xml:"PicUrl"`
	MediaID string `xml:"MediaId"`
}

type VoiceMessage struct {
	Message
	rawMessage
	Format  string `xml:"Format"`
	MediaID string `xml:"MediaId"`
}

type VedioMessage struct {
	Message
	rawMessage
	MediaID      string `xml:"MediaId"`
	ThumbMediaID string `xml:"ThumbMediaId"`
}

type LocationMessage struct {
	Message
	rawMessage
	Latitude  float32 `xml:"Location_X:"`
	Longitude float32 `xml:"Location_Y"`
	Scale     int     `xml:"Scale"`
	Label     string  `xml:"Label"`
}

type LinkMessage struct {
	Message
	rawMessage
	Title       string `xml:"Title"`
	Description string `xml:"Description"`
	PicURL      string `xml:"PicUrl"`
}

type SendingMessage interface {
}

type generalSendMessage struct {
	To      string      `json:"touser"`
	Type    MessageType `json:"msgtype"`
	AgentID int         `json:"agentid"`
	safe    int         `json:"safe"`
}

type TextNode struct {
	Content string `json:"content"`
}

type ImageNode struct {
	MediaId string `json:"media_id"`
}

type FileNode struct {
	MediaId string `json:"media_id"`
}

type TextSendMessage struct {
	SendingMessage
	generalSendMessage
	Text *TextNode `json:"text"`
}

type ImageSendMessage struct {
	SendingMessage
	generalSendMessage
	Image *ImageNode `json:"image"`
}

type FileSendMessage struct {
	SendingMessage
	generalSendMessage
	File *FileNode `json:"file"`
}

type APIReturn struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

type APIChatReturn struct {
	APIReturn
	InvalidUser  string `json:"invaliduser"`
	InvalidParty string `json:"invalidparty"`
	InvalidTag   string `json:"invalidtag"`
}

type APIAccessTokenReturn struct {
	APIReturn
	AccessToken string `json:"access_token"`
	Expire      int64  `json:"expires_in"`
}

type APIMediaUploadReturn struct {
	APIReturn
	Type     string `json:"type"`
	MediaId  string `json:"media_id"`
	CreateAt string `json:"create_at"`
}
