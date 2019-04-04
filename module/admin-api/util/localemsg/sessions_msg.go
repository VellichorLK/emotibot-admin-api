package localemsg

var sessionsMsg = map[string]map[string]string{
	"zh-cn": map[string]string{
		// Export records
		"sessionId":      "会话ID",
		"startTime":      "会话开始时间",
		"endTime":        "会话结束时间",
		"userId":         "用户ID",
		"rating":         "满意度",
		"customInfo":     "客制化资讯",
		"feedback":       "反馈选择",
		"customFeedback": "反馈文字",
		"feedbackTime":   "反馈时间",
	},
	"zh-tw": map[string]string{
		// Export records
		"sessionId":      "會話ID",
		"startTime":      "會話開始時間",
		"endTime":        "會話結束時間",
		"userId":         "用戶ID",
		"rating":         "滿意度",
		"customInfo":     "客制化資訊",
		"robotAnswer":    "機器人回答",
		"feedback":       "反饋選擇",
		"customFeedback": "反饋文字",
		"feedbackTime":   "反饋時間",
	},
}
