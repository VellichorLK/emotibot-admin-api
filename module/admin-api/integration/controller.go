package integration

import (
	"fmt"
	"net/http"
	"strings"

	"emotibot.com/emotigo/module/admin-api/QA"

	"emotibot.com/emotigo/pkg/services/linebot"

	"github.com/siongui/gojianfan"

	"emotibot.com/emotigo/module/admin-api/util/AdminErrors"
	"emotibot.com/emotigo/pkg/logger"
	"emotibot.com/emotigo/pkg/services/workweixin"

	"emotibot.com/emotigo/module/admin-api/util"
)

var (
	// ModuleInfo is needed for module define
	ModuleInfo    util.ModuleInfo
	textConverter = gojianfan.T2S
)

func init() {
	ModuleInfo = util.ModuleInfo{
		ModuleName: "integration",
		EntryPoints: []util.EntryPoint{
			util.NewEntryPoint("GET", "chat/{platform}/{appid}", []string{}, handlePlatformChat),
			util.NewEntryPoint("POST", "chat/{platform}/{appid}", []string{}, handlePlatformChat),
		},
	}
}

var handlers map[string]func(w http.ResponseWriter, r *http.Request, appid string, config map[string]string)

func init() {
	handlers = map[string]func(w http.ResponseWriter, r *http.Request, appid string, config map[string]string){
		"line":       handleLineReply,
		"workweixin": handleWorkWeixinReply,
	}
}

func handlePlatformChat(w http.ResponseWriter, r *http.Request) {
	platform := util.GetMuxVar(r, "platform")
	if platform == "" {
		util.ReturnError(w, AdminErrors.ErrnoRequestError, "platform invalid")
		return
	}
	appid := util.GetMuxVar(r, "appid")
	if appid == "" {
		util.ReturnError(w, AdminErrors.ErrnoRequestError, "appid invalid")
		return
	}

	config, err := GetPlatformConfig(appid, platform)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Error.Println("Get platform conf fail:", err.Error())
		return
	}

	logger.Trace.Printf("Get platform config of %s, %s: %+v\n", appid, platform, config)

	handler := handlers[platform]
	if handler == nil {
		w.WriteHeader(http.StatusBadRequest)
		logger.Error.Println("Unsupported platform:", platform)
		return
	}
	handler(w, r, appid, config)
}

var lineBots = map[string]*linebot.Client{}
var lineConverters = [](func(answer *QA.BFOPOpenapiAnswer) linebot.SendingMessage){
	createLineTextMessage,
	createLineButtonTemplateMessage,
	createLineFlexMessage,
}
var lineConverterName = []string{"text", "button template", "flex"}
var lineConverterIdx = map[string]int{}

func handleLineReply(w http.ResponseWriter, r *http.Request, appid string, config map[string]string) {
	if config["token"] == "" || config["secret"] == "" {
		return
	}
	locale := r.URL.Query().Get("locale")
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
	if locale == "zhtw" {
		textConverter = gojianfan.S2T
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

				if strings.Index(message.Text, "##") == 0 {
					logger.Trace.Println("command series,", strings.Replace(message.Text, "##", "", 1))
					if strings.Replace(message.Text, "##", "", 1) == "change" {
						logger.Trace.Println("change converter")
						lineConverterIdx[appid]++
						lineConverterIdx[appid] = lineConverterIdx[appid] % len(lineConverters)
						lineAnswers = append(lineAnswers, linebot.NewTextMessage(lineConverterName[lineConverterIdx[appid]]))
					}
				} else {
					answers := GetChatResult(appid, event.Source.UserID, message.Text)
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

				logger.Trace.Printf("Reply %d messages\n", len(lineAnswers))
				if _, err := bot.ReplyMessage(event.ReplyToken, lineAnswers).Do(); err != nil {
					logger.Error.Println("Reply message fail: ", err.Error())
				}
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
		opt := textConverter(d)
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
		opt := textConverter(d)
		options = append(options, linebot.NewMessageAction(opt, opt))
	}
	buttons := linebot.NewButtonsTemplate("", "", textConverter(answer.Value), options...)
	return linebot.NewTemplateMessage(answer.ToString(), buttons)
}

func createLineTextMessage(answer *QA.BFOPOpenapiAnswer) linebot.SendingMessage {
	return linebot.NewTextMessage(textConverter(answer.ToString()))
}

var workWeixinBot = map[string]*workweixin.Client{}

func handleWorkWeixinReply(w http.ResponseWriter, r *http.Request, appid string, config map[string]string) {
	if config["token"] == "" || config["encoded-aes"] == "" {
		return
	}
	if _, ok := workWeixinBot[appid]; !ok {
		bot, err := workweixin.New(config["token"], config["encoded-aes"])
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logger.Error.Println("workWeixinBot init fail: ", err.Error())
			return
		}
		workWeixinBot[appid] = bot
	}

	bot := workWeixinBot[appid]
	if r.Method == http.MethodGet {
		logger.Trace.Println("Start url verify")
		bot.VerifyURL(w, r)
		return
	}
}
