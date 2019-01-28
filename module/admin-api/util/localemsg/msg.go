package localemsg

import "emotibot.com/emotigo/pkg/logger"

var localeMsg = map[string]map[string]string{
	"zh-cn": map[string]string{
		"Success": "成功",
		"Fail":    "失败",
	},
	"zh-tw": map[string]string{
		"Success": "成功",
		"Fail":    "失敗",
	},
}

func init() {
	allMsg := []map[string]map[string]string{
		intentMsg, auditMsg, dictionaryMsg,
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
		return localeMsg["zh-cn"][key]
	}

	if _, ok := localeMsg[locale][key]; !ok {
		logger.Trace.Printf("Key [%s] in locale [%s] is not existed\n", key, locale)
		return ""
	}
	return localeMsg[locale][key]
}
