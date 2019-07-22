package integration

import (
	"net/http"
	"strings"

	"emotibot.com/emotigo/module/admin-api/util"

	"emotibot.com/emotigo/pkg/logger"
	"emotibot.com/emotigo/pkg/services/workweixin"
	"github.com/siongui/gojianfan"
)

var workWeixinBot = map[string]*workweixin.Client{}

func handleWorkWeixinReply(w http.ResponseWriter, r *http.Request, appid string, config map[string]string) {
	if config["token"] == "" || config["encoded-aes"] == "" || config["corpid"] == "" || config["secret"] == "" {
		return
	}
	// If client is not created, create it with config
	if _, ok := workWeixinBot[appid]; !ok {
		fullEncoded := strings.Replace(config["encoded-aes"], "=", "", -1) + "="
		bot, err := workweixin.New(config["corpid"], config["secret"], config["token"], fullEncoded)
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
		// Work weixin will use GET method to validate webhook validation
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
				if answer.Type == "cmd" {
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

func generateWorkWeixinConfig(values map[string]string) map[string]string {
	token := util.GenRandomString(32)
	workWeixinEncoded := util.GenRandomString(43)

	values["token"] = token
	values["encoded-aes"] = workWeixinEncoded

	return values
}
