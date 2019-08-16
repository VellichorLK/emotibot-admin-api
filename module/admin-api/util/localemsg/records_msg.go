package localemsg

var recordsMsg = map[string]map[string]string{
	"zh-cn": map[string]string{
		// Modules
		"backfill":    "未知回复",
		"chat":        "闲聊",
		"keyword":     "其他",
		"function":    "机器人技能",
		"faq":         "常用标准问题",
		"task_engine": "多轮对话引擎",
		"to_human":    "转人工",
		"knowledge":   "知识百科",
		"domain_kg":   "知识推理引擎",
		"command":     "机器人话术",
		"emotion":     "情绪话术",

		// Export records
		"unique_id":      "对话ID",
		"sessionId":      "会话ID",
		"taskEngineId":   "多轮ID",
		"userId":         "用户ID",
		"userQ":          "用户问题",
		"FAQ":            "标准问题",
		"robotAnswer":    "机器人回答",
		"robotRawAnswer": "机器人回答 JSON",
		"matchScore":     "匹配分数",
		"module":         "出话模块",
		"source":         "出话来源",
		"logTime":        "访问时间",
		"emotionCol":     "情感",
		"emotionScore":   "情感分数",
		"intent":         "意图",
		"intentScore":    "意图分数",
		"customInfo":     "客制化资讯",
		"FAQCategory":    "FAQ分类",
		"FAQLabel":       "FAQ标签",
		"feedback":       "反馈选择",
		"customFeedback": "反馈文字",
		"feedbackTime":   "反馈时间",
		"threshold":      "出话阈值",
		"respondTime":    "相应时间",
	},
	"zh-tw": map[string]string{
		// Modules
		"backfill":    "未知回覆",
		"chat":        "閒聊",
		"keyword":     "其他",
		"function":    "機器人技能",
		"faq":         "常用標準問題",
		"task_engine": "多輪對話引擎",
		"to_human":    "轉人工",
		"knowledge":   "知識百科",
		"domain_kg":   "知識推理引擎",
		"command":     "機器人話術",
		"emotion":     "情緒話術",

		// Export records
		"unique_id":      "對話ID",
		"sessionId":      "會話ID",
		"taskEngineId":   "多輪場景ID",
		"userId":         "用戶ID",
		"userQ":          "用戶問題",
		"FAQ":            "標準問題",
		"robotAnswer":    "機器人回答",
		"robotRawAnswer": "機器人回答 JSON",
		"matchScore":     "匹配分數",
		"module":         "出話模組",
		"source":         "出話來源",
		"logTime":        "訪問時間",
		"emotionCol":     "情感",
		"emotionScore":   "情感分數",
		"intent":         "意圖",
		"intentScore":    "意圖分數",
		"customInfo":     "客制化資訊",
		"FAQCategory":    "FAQ分類",
		"FAQLabel":       "FAQ標籤",
		"feedback":       "反饋選擇",
		"customFeedback": "反饋文字",
		"feedbackTime":   "反饋時間",
		"threshold":      "出話閾值",
		"respondTime":    "響應時間",
	},
}
