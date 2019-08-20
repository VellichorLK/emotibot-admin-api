package integration

import (
	"net/http"
	"strings"

	"emotibot.com/emotigo/module/admin-api/QA"
	"emotibot.com/emotigo/module/admin-api/util"
	"emotibot.com/emotigo/pkg/logger"
	"emotibot.com/emotigo/pkg/services/workweixin"
	"github.com/PuerkitoBio/goquery"
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
	bot.GetNewAccessToken()
	logger.Info.Println("bot config: ", bot)

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
			answers := GetChatResult(appid, message.From, message.Content, PlatformWorkWeixin)

			var replyMessages []workweixin.SendingMessage
			for _, answer := range answers {
				if answer == nil {
					continue
				}
				var objMessage interface{}
				if answer.Type == "text" && answer.SubType == "text" {
					answerText := textConverter(answer.ToString())

					html, err := goquery.NewDocumentFromReader(strings.NewReader(answerText))
					if err != nil {
						logger.Info.Println("answerText: ", answerText)
						objMessage = workweixin.NewTextMessage(message.From, message.AgentID, answerText)
					} else {
						filterTags := html.Find("p, img")
						if len(filterTags.Nodes) == 0 {
							rText := html.Find("body").Text()
							logger.Info.Println("rText: ", rText)
							objMessage = workweixin.NewTextMessage(message.From, message.AgentID, strings.TrimSpace(rText))
						} else {
							filterTags.Each(func(i int, s *goquery.Selection) {
								nodeTag := s.Nodes[0].Data
								node := s.Nodes[0]
								logger.Info.Println("nodeTag: ", nodeTag)
								logger.Info.Println(node)
								if nodeTag == "img" {
									imgAnswer := QA.BFOPOpenapiAnswer{}
									imgAnswer.Type = "url"
									imgAnswer.SubType = "image"
									imgAnswer.Value = node.Attr[0].Val
									imgAnswer.Data = answer.Data
									logger.Info.Println(imgAnswer)
									mediaId, err := bot.UploadMedia(&imgAnswer)
									if err == nil {
										imgMessage := workweixin.NewImageMessage(message.From, message.AgentID, mediaId)
										logger.Info.Println("imgMessage: ", imgMessage)
										replyMessages = append(replyMessages, imgMessage)
									}
								}
								if nodeTag == "p" {
									txtMessage := workweixin.NewTextMessage(message.From, message.AgentID, strings.TrimSpace(s.Text()))
									logger.Info.Println("txtMessage: ", txtMessage)
									replyMessages = append(replyMessages, txtMessage)
								}
							})
						}
					}
				}
				// relatelist相关问，guslist相似问，guslist推荐问回答处理
				if answer.Type == "text" && (answer.SubType == "relatelist" || answer.SubType == "guslist") {
					objMessage = workweixin.NewTextMessage(message.From, message.AgentID, answer.ToString())
				}
				if answer.Type == "cmd" {
					objMessage = workweixin.NewTextMessage(message.From, message.AgentID, textConverter(answer.ToString()))
				}
				if answer.Type == "url" && answer.SubType == "image" {
					mediaId, err := bot.UploadMedia(answer)
					if err == nil {
						objMessage = workweixin.NewImageMessage(message.From, message.AgentID, mediaId)
					}
				}
				if answer.Type == "url" && answer.SubType == "docs" {
					mediaId, err := bot.UploadMedia(answer)
					if err == nil {
						objMessage = workweixin.NewFileMessage(message.From, message.AgentID, mediaId)
					}
				}

				if objMessage != nil {
					replyMessages = append(replyMessages, objMessage)
				}
			}

			logger.Info.Println("replyMessages: ", replyMessages)

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
