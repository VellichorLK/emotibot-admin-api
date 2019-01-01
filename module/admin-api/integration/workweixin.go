package integration

import (
	"net/http"

	"emotibot.com/emotigo/pkg/logger"
	"emotibot.com/emotigo/pkg/services/workweixin"
	"github.com/siongui/gojianfan"
)

var workWeixinBot = map[string]*workweixin.Client{}

func handleWorkWeixinReply(w http.ResponseWriter, r *http.Request, appid string, config map[string]string) {
	if config["token"] == "" || config["encoded-aes"] == "" || config["cropid"] == "" || config["secret"] == "" {
		return
	}
	if _, ok := workWeixinBot[appid]; !ok {
		bot, err := workweixin.New(config["cropid"], config["secret"], config["token"], config["encoded-aes"])
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logger.Error.Println("workWeixinBot init fail: ", err.Error())
			return
		}
		workWeixinBot[appid] = bot
	}
	locale := r.URL.Query().Get("locale")
	if locale == "zhtw" {
		textConverter = gojianfan.S2T
	} else {
		textConverter = gojianfan.T2S
	}

	bot := workWeixinBot[appid]
	if r.Method == http.MethodGet {
		bot.VerifyURL(w, r)
		return
	} else if r.Method == http.MethodPost {
		msg, err := bot.ParseRequest(r)
		if err != nil {
			logger.Error.Println("workWeixinBot parse fail: ", err.Error())
		}
		switch message := msg.(type) {
		case *workweixin.TextMessage:
			logger.Trace.Printf("Receive: %s\n", message.Content)
			answers := GetChatResult(appid, message.From, message.Content)

			replyMessages := []workweixin.SendingMessage{}
			for _, answer := range answers {
				if answer == nil {
					continue
				}
				if answer.Type == "text" {
					replyMessages = append(replyMessages, workweixin.NewTextMessage(
						message.From, message.AgentID, textConverter(answer.ToString())))
				}
			}
			workWeixinQueue <- &workWeixinTask{
				Bot:      bot,
				Messages: replyMessages,
			}
		}
		w.WriteHeader(http.StatusOK)
	}
}
