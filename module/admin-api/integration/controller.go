package integration

import (
	"net/http"

	"emotibot.com/emotigo/pkg/services/linebot"

	"github.com/siongui/gojianfan"

	"emotibot.com/emotigo/module/admin-api/util/AdminErrors"
	"emotibot.com/emotigo/pkg/logger"
	"emotibot.com/emotigo/pkg/services/workweixin"

	"emotibot.com/emotigo/module/admin-api/util"
)

var (
	// ModuleInfo is needed for module define
	ModuleInfo util.ModuleInfo
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
	for _, event := range events {
		switch event.Type {
		case linebot.EventTypeMessage:
			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				answer := GetChatResult(appid, event.Source.UserID, message.Text)
				if locale == "zhtw" {
					answer = gojianfan.S2T(answer)
				} else {
					answer = gojianfan.T2S(answer)
				}

				if _, err := bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(answer)).Do(); err != nil {
					logger.Error.Println("Reply message fail: ", err.Error())
				}
			}
		}
	}
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
