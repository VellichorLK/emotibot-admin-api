package localemsg

import "emotibot.com/emotigo/pkg/logger"

var localeMsg = map[string]map[string]string{
	"zh-cn": map[string]string{
		"Fail":                            "失败",
		"IntentID":                        "意图编号",
		"DeleteIntent":                    "删除意图",
		"AddIntent":                       "新增意图",
		"UpdateIntent":                    "更新意图",
		"IntentName":                      "意图名称",
		"IntentPositive":                  "意图正例语料",
		"IntentNegative":                  "意图反例语料",
		"IntentSentence":                  "语句",
		"AddIntentSummaryTpl":             ": %s, 正例语料 %d 条, 反例语料 %d 条",
		"UpdateIntentSummaryTpl":          ": %s, 修改语料 %d 条, 删除语料 %d 条",
		"IntentModifyUpdate":              "语料变更内容",
		"IntentModifyDelete":              "欲删除语料",
		"IntentUploadSheetErr":            "上传的资料表数量错误",
		"IntentBF2Sheet1Name":             "请在此页填写你的意图名称和positive语句",
		"IntentBF2Sheet2Name":             "请在此页填写你的意图名称和negative语句",
		"IntentUploadNoHeaderTpl":         "上传的资料表 %s 中标头遗失",
		"IntentUploadHeaderErrTpl":        "上传的资料表 %s 中标头格式错误",
		"IntentUploadBF2RowInvalidTpl":    "资料表 %s 中第 %d 行缺少栏位",
		"IntentUploadBF2RowNoNameTpl":     "资料表 %s 中第 %d 行名字为空",
		"IntentUploadBF2RowNoSentenceTpl": "资料表 %s 中第 %d 行语句为空",
		"IntentExport":                    "导出意图",
	},
	"zh-tw": map[string]string{
		"Fail":                            "失敗",
		"IntentID":                        "意圖編號",
		"DeleteIntent":                    "刪除意圖",
		"AddIntent":                       "新增意圖",
		"UpdateIntent":                    "更新意圖",
		"IntentName":                      "意圖名稱",
		"IntentPositive":                  "意圖正例語料",
		"IntentNegative":                  "意圖反例語料",
		"IntentSentence":                  "語句",
		"AddIntentSummaryTpl":             ": %s, 正例語料 %d 條, 反例語料 %d 條",
		"UpdateIntentSummaryTpl":          ": %s, 修改語料 %d 條, 刪除語料 %d 條",
		"IntentModifyUpdate":              "語料變更內容",
		"IntentModifyDelete":              "欲刪除語料",
		"IntentUploadSheetErr":            "上傳的資料表數量錯誤",
		"IntentBF2Sheet1Name":             "請在此頁填寫你的意圖名稱和positive語句",
		"IntentBF2Sheet2Name":             "請在此頁填寫你的意圖名稱和negative語句",
		"IntentUploadNoHeaderTpl":         "上傳的資料表 %s 中標頭遺失",
		"IntentUploadHeaderErrTpl":        "上傳的資料表 %s 中標頭格式錯誤",
		"IntentUploadBF2RowInvalidTpl":    "資料表 %s 中第 %d 行缺少欄位",
		"IntentUploadBF2RowNoNameTpl":     "資料表 %s 中第 %d 行名字為空",
		"IntentUploadBF2RowNoSentenceTpl": "資料表 %s 中第 %d 行語句為空",
		"IntentExport":                    "導出意圖",
	},
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
