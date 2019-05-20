package integration

import (
	"encoding/json"
	"fmt"
	"net/http"

	robotConfig "emotibot.com/emotigo/module/admin-api/Robot/config.v1"

	"emotibot.com/emotigo/pkg/misc/adminerrors"

	"emotibot.com/emotigo/module/admin-api/util/requestheader"

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
			util.NewEntryPoint("GET", "config/{platform}", []string{"view"}, handleGetConfig),
			//util.NewEntryPoint("GET", "config/{platform}/{appid}", []string{"view"}, handleGetConfig),
			util.NewEntryPoint("POST", "config/{platform}", []string{}, handleSetConfig),
			util.NewEntryPoint("DELETE", "config/{platform}", []string{}, handleDeleteConfig),
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
				logger.Trace.Printf("Send %d line reply\n", len(task.Messages))
				if _, err := task.Bot.ReplyMessage(task.ReplyToken, task.Messages).Do(); err != nil {
					logger.Error.Println("Reply message fail: ", err.Error())
				}
			}
		case task := <-workWeixinQueue:
			if task != nil {
				logger.Trace.Printf("Send %d work weixin reply\n", len(task.Messages))
				if err := task.Bot.SendMessages(task.Messages); err != nil {
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

func handleGetConfig(w http.ResponseWriter, r *http.Request) {
	appid := requestheader.GetAppID(r)
	platform := util.GetMuxVar(r, "platform")
	configs, err := GetPlatformConfig(appid, platform)
	util.Return(w, err, configs)
}

func handleSetConfig(w http.ResponseWriter, r *http.Request) {
	appid := requestheader.GetAppID(r)
	platform := util.GetMuxVar(r, "platform")
	values := map[string]string{}
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	decodeError := decoder.Decode(&values)
	if decodeError != nil {
		util.ReturnError(w, adminerrors.ErrnoRequestError, decodeError.Error())
		return
	}

	values = generateConfigOfPlatform(platform, values)

	configs, err := SetPlatformConfig(appid, platform, values)
	if err != nil {
		util.Return(w, err, nil)
		return
	}

	configs["url"] = getWebhookURL(appid, platform, requestheader.GetOrigin(r))

	// 更新配置缓存
	key := fmt.Sprintf("%s-%s", appid, platform)
	configCache[key] = configs

	util.Return(w, err, configs)
}

func handleDeleteConfig(w http.ResponseWriter, r *http.Request) {
	appid := requestheader.GetAppID(r)
	platform := util.GetMuxVar(r, "platform")
	err := DeletePlatformConfig(appid, platform)
	util.Return(w, err, nil)
}

func generateConfigOfPlatform(platform string, values map[string]string) map[string]string {
	if platform == "workweixin" {
		return generateWorkWeixinConfig(values)
	}
	return values
}

func getWebhookURL(appid, platform, dftHost string) string {
	server := dftHost

	config, err := robotConfig.GetConfig(appid, "uploadimg_server")
	if err == nil && config != nil {
		if config.Value != "" {
			server = config.Value
		}
	}

	return fmt.Sprintf("%s/api/v1/integration/chat/%s/%s",
		server, platform, appid)
}
