package LocaleMsg

import "emotibot.com/emotigo/module/admin-api/util"

var localeMsg = map[string]map[string]string{
	"zh-cn": map[string]string{
		"Fail":                   "失败",
		"IntentID":               "意图编号",
		"DeleteIntent":           "删除意图",
		"AddIntent":              "新增意图",
		"UpdateIntent":           "更新意图",
		"IntentName":             "意图名称",
		"IntentPositive":         "意图正例语料",
		"IntentNegative":         "意图反例语料",
		"AddIntentSummaryTpl":    ": %s, 正例语料 %d 条, 反例语料 %d 条",
		"UpdateIntentSummaryTpl": ": %s, 修改语料 %d 条, 删除语料 %d 条",
		"IntentModifyUpdate":     "语料变更内容",
		"IntentModifyDelete":     "欲删除语料",
	},
	"zh-tw": map[string]string{
		"Fail":                   "失敗",
		"IntentID":               "意圖編號",
		"DeleteIntent":           "刪除意圖",
		"AddIntent":              "新增意圖",
		"UpdateIntent":           "更新意圖",
		"IntentName":             "意圖名稱",
		"IntentPositive":         "意圖正例語料",
		"IntentNegative":         "意圖反例語料",
		"AddIntentSummaryTpl":    ": %s, 正例語料 %d 條, 反例語料 %d 條",
		"UpdateIntentSummaryTpl": ": %s, 修改語料 %d 條, 刪除語料 %d 條",
		"IntentModifyUpdate":     "語料變更內容",
		"IntentModifyDelete":     "欲刪除語料",
	},
}

func Get(locale string, key string) string {
	if _, ok := localeMsg[locale]; !ok {
		return localeMsg["zh-cn"][key]
	}

	if _, ok := localeMsg[locale][key]; !ok {
		util.LogTrace.Printf("Key [%s] in locale [%s] is not existed\n", key, locale)
		return ""
	}
	return localeMsg[locale][key]
}
