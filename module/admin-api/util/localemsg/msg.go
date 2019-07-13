package localemsg

import "emotibot.com/emotigo/pkg/logger"

const (
	ZhCn = "zh-cn"
	ZhTw = "zh-tw"
)

var localeMsg = map[string]map[string]string{
	ZhCn: map[string]string{
		"Success": "成功",
		"Fail":    "失败",
	},
	ZhTw: map[string]string{
		"Success": "成功",
		"Fail":    "失敗",
	},
}

func init() {
	allMsg := []map[string]map[string]string{
		intentMsg, intentTestMsg, auditMsg, dictionaryMsg, statsMsg, taskengineMsg, recordsMsg, sessionsMsg, customChatMsg,cmdMsg,
	}

	// merge all module lang map
	for _, msgMap := range allMsg {
		for lang := range msgMap {
			if _, ok := localeMsg[lang]; !ok {
				localeMsg[lang] = map[string]string{}
			}
			for key, val := range msgMap[lang] {
				localeMsg[lang][key] = val
			}
		}
	}
}

func Get(locale string, key string) string {
	if _, ok := localeMsg[locale]; !ok {
		return localeMsg[ZhCn][key]
	}

	if _, ok := localeMsg[locale][key]; !ok {
		logger.Trace.Printf("Key [%s] in locale [%s] is not existed\n", key, locale)
		return ""
	}
	return localeMsg[locale][key]
}
