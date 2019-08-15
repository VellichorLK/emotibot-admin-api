package integration

import (
	"fmt"
	"net/http"
	"strings"

	"emotibot.com/emotigo/module/admin-api/QA"
	"emotibot.com/emotigo/pkg/logger"
	"emotibot.com/emotigo/pkg/services/linebot"
	"github.com/siongui/gojianfan"
)

// lineBots is the cache of linebot client, key is appid, value is bot client
var lineBots = map[string]*linebot.Client{}

// lineConverters is the converter of reply message type
var lineConverters = [](func(answer *QA.BFOPOpenapiAnswer) linebot.SendingMessage){
	createLineTextMessage,
	createLineButtonTemplateMessage,
	createLineFlexMessage,
}

// lineConverterName is used to show the converter name to bot command
var lineConverterName = []string{"text", "button template", "flex"}

// lineConverterIdx is setting of converter, key is appid
var lineConverterIdx = map[string]int{}

func handleLineReply(w http.ResponseWriter, r *http.Request, appid string, config map[string]string) {
	if config["token"] == "" || config["secret"] == "" {
		return
	}
	// If client is not created, create it with config
	if _, ok := lineBots[appid]; !ok {
		bot, err := linebot.New(config["secret"], config["token"])
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logger.Error.Println("Linebot init fail: ", err.Error())
			return
		}
		lineBots[appid] = bot
	}

	bot := lineBots[appid]
	events, err := bot.ParseRequest(r)
	if err != nil {
		if err == linebot.ErrInvalidSignature {
			logger.Error.Println("Request signature check fail: ", err.Error())
		} else {
			logger.Error.Println("Unknown error: ", err.Error())
		}
		return
	}

	locale := r.URL.Query().Get("locale")
	if locale == "zhtw" {
		textConverter = gojianfan.S2T
	} else {
		textConverter = gojianfan.T2S
	}
	if _, ok := lineConverterIdx[appid]; !ok {
		lineConverterIdx[appid] = 0
	}

	for _, event := range events {
		switch event.Type {
		case linebot.EventTypeMessage:
			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				lineAnswers := []linebot.SendingMessage{}

				// message with '##' prefix will be commands for users for now
				if strings.Index(message.Text, "##") == 0 {
					logger.Trace.Println("command series,", strings.Replace(message.Text, "##", "", 1))
					// 'change' command will change message format converter
					if strings.Replace(message.Text, "##", "", 1) == "change" {
						logger.Trace.Println("change converter")
						lineConverterIdx[appid]++
						lineConverterIdx[appid] = lineConverterIdx[appid] % len(lineConverters)
						lineAnswers = append(lineAnswers, linebot.NewTextMessage(lineConverterName[lineConverterIdx[appid]]))
					}
				} else {
					answers := GetChatResult(appid, event.Source.UserID, message.Text, PlatformLine)
					for _, answer := range answers {
						if answer == nil {
							continue
						}
						if answer.Type == "text" {
							if (answer.SubType == "relatelist" || answer.SubType == "guslist") && len(answer.Data) > 0 {
								lineAnswers = append(lineAnswers, lineConverters[lineConverterIdx[appid]](answer))
							} else {
								lineAnswers = append(lineAnswers, linebot.NewTextMessage(textConverter(answer.ToString())))
							}
						}
					}
				}

				sourceID := ""
				switch event.Source.Type {
				case linebot.EventSourceTypeGroup:
					sourceID = event.Source.GroupID
				case linebot.EventSourceTypeRoom:
					sourceID = event.Source.RoomID
				case linebot.EventSourceTypeUser:
					sourceID = event.Source.UserID
				}
				logger.Trace.Printf("Reply %d messages to %s\n", len(lineAnswers), sourceID)
				lineTaskQueue <- &lineTask{
					Bot:        bot,
					ReplyToken: event.ReplyToken,
					Messages:   lineAnswers,
				}

				// if _, err := bot.ReplyMessage(event.ReplyToken, lineAnswers).Do(); err != nil {
				// 	logger.Error.Println("Reply message fail: ", err.Error())
				// }
			}
		}
	}
}

func createLineFlexMessage(answer *QA.BFOPOpenapiAnswer) linebot.SendingMessage {
	contents := []linebot.FlexComponent{
		&linebot.TextComponent{
			Type:  linebot.FlexComponentTypeText,
			Text:  textConverter(answer.Value),
			Align: linebot.FlexComponentAlignTypeStart,
			Wrap:  true,
		},
	}
	for idx, d := range answer.Data {
		opt := textConverter(d["name"])
		contents = append(contents,
			&linebot.TextComponent{
				Type:   linebot.FlexComponentTypeText,
				Text:   fmt.Sprintf("%d. %s", idx+1, opt),
				Action: linebot.NewMessageAction(opt, opt),
				Align:  linebot.FlexComponentAlignTypeStart,
				Wrap:   true,
				Color:  "#1875f0",
			})
	}
	return linebot.NewFlexMessage(
		textConverter(answer.ToString()),
		&linebot.BubbleContainer{
			Type: linebot.FlexContainerTypeBubble,
			Body: &linebot.BoxComponent{
				Type:     linebot.FlexComponentTypeBox,
				Layout:   linebot.FlexBoxLayoutTypeVertical,
				Contents: contents,
				Spacing:  linebot.FlexComponentSpacingTypeMd,
			},
		},
	)
}

func createLineButtonTemplateMessage(answer *QA.BFOPOpenapiAnswer) linebot.SendingMessage {
	options := []linebot.TemplateAction{}
	for _, d := range answer.Data {
		opt := textConverter(d["name"])
		options = append(options, linebot.NewMessageAction(opt, opt))
	}
	buttons := linebot.NewButtonsTemplate("", "", textConverter(answer.Value), options...)
	return linebot.NewTemplateMessage(answer.ToString(), buttons)
}

func createLineTextMessage(answer *QA.BFOPOpenapiAnswer) linebot.SendingMessage {
	return linebot.NewTextMessage(textConverter(answer.ToString()))
}
