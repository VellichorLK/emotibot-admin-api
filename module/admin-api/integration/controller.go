package integration

import (
	"fmt"
	"net/http"

	"github.com/siongui/gojianfan"

	"emotibot.com/emotigo/module/admin-api/util/AdminErrors"
	"emotibot.com/emotigo/pkg/logger"

	"emotibot.com/emotigo/module/admin-api/util"
)

var (
	// ModuleInfo is needed for module define
	ModuleInfo      util.ModuleInfo
	textConverter   = gojianfan.T2S
	queueSize       = 10
	lineTaskQueue   = make(chan *lineTask, queueSize)
	workWeixinQueue = make(chan *workWeixinTask, queueSize)
)

func init() {
	ModuleInfo = util.ModuleInfo{
		ModuleName: "integration",
		EntryPoints: []util.EntryPoint{
			util.NewEntryPoint("GET", "chat/{platform}/{appid}", []string{}, handlePlatformChat),
			util.NewEntryPoint("POST", "chat/{platform}/{appid}", []string{}, handlePlatformChat),
			util.NewEntryPoint("GET", "configs/reload", []string{}, handleReloadPlatformConfig),
			// util.NewEntryPoint("GET", "configs", []string{"view"}, handleGetConfigs),
			// util.NewEntryPoint("Get", "config/{platform}", []string{"view"}, handleGetConfigTemplate),
			// util.NewEntryPoint("Get", "config/{platform}/{appid}", []string{"view"}, handleGetConfig),
		},
	}
	go sendFromQueue()
}

// sendFromQueue will get reply task from queue to avoid webhook timeout
func sendFromQueue() {
	for {
		select {
		case task := <-lineTaskQueue:
			if task != nil {
				logger.Trace.Println("Send line reply")
				if _, err := task.Bot.ReplyMessage(task.ReplyToken, task.Messages).Do(); err != nil {
					logger.Error.Println("Reply message fail: ", err.Error())
				}
			}
		case task := <-workWeixinQueue:
			if task != nil {
				logger.Trace.Println("Send work weixin reply")
				if _, err := task.Bot.SendMessages(task.Messages); err != nil {
					logger.Error.Println("Reply message fail: ", err.Error())
				}
			}
		}
	}
}

var handlers = map[string]func(w http.ResponseWriter, r *http.Request, appid string, config map[string]string){
	"line":       handleLineReply,
	"workweixin": handleWorkWeixinReply,
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

	// Check if config cache is valid of not
	var err error
	key := fmt.Sprintf("%s-%s", appid, platform)
	config, ok := configCache[key]
	if !ok {
		config, err = GetPlatformConfig(appid, platform)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logger.Error.Println("Get platform conf fail:", err.Error())
			return
		}
		configCache[key] = config
		logger.Trace.Println("Add cache")
	}

	logger.Trace.Printf("Get platform config of %s, %s: %+v\n", appid, platform, config)

	// Get handler via platfrom value, which is get from URL var
	handler := handlers[platform]
	if handler == nil {
		w.WriteHeader(http.StatusBadRequest)
		logger.Error.Println("Unsupported platform:", platform)
		return
	}
	handler(w, r, appid, config)
}

var configCache = map[string]map[string]string{}

func handleReloadPlatformConfig(w http.ResponseWriter, r *http.Request) {
	configCache = map[string]map[string]string{}
}
