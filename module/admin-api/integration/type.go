package integration

import (
	"emotibot.com/emotigo/pkg/services/linebot"
	"emotibot.com/emotigo/pkg/services/workweixin"
)

type lineTask struct {
	Bot        *linebot.Client
	ReplyToken string
	Messages   []linebot.SendingMessage
}

type workWeixinTask struct {
	Bot      *workweixin.Client
	Messages []workweixin.SendingMessage
}
