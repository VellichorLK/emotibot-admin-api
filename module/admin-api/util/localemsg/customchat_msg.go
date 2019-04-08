package localemsg

var customChatMsg = map[string]map[string]string{
	ZhCn: map[string]string{
		"CustomChatUploadSheetErr":            "上传的资料表数量错误",
		"CustomChatQuestionSheetName":             "闲聊答案",
		"CustomChatExtendSheetName":             "闲聊语料",
		"CustomChatUploadNoHeaderTpl":         "上传的资料表 %s 中标头遗失",
		"CustomChatCategory":                  "分类",
		"CustomChatQuestion":                  "问题",
		"CustomChatAnswer":                    "答案",
		"CustomChatExtend":                    "相似问",
		"CustomChatUploadQuestionRowInvalidTpl":    "资料表 %s 中第 %d 行缺少栏位",
		"CustomChatUploadQuestionRowNoQuestionTpl":     "资料表 %s 中第 %d 行问题为空",
		"CustomChatUploadQuestionRowNoAnswerTpl": "资料表 %s 中第 %d 行答案为空",
		"CustomChatUploadQuestionRowNoExtendTpl": "资料表 %s 中第 %d 行答案为空",
		"CustomChatExport": "导出闲聊",
		"CustomChatUploadQuestionExceedLimit": "资料表 %s 中第 %d 行问题超过%d",
		"CustomChatUploadAnswerExceedLimit": "资料表 %s 中第 %d 行答案超过%d",
	},
	ZhTw: map[string]string{
		"CustomChatUploadSheetErr":            "上传的资料表数量错误",
		"CustomChatQuestionSheetName":             "閒聊答案",
		"CustomChatExtendSheetName":             "閒聊語料",
		"CustomChatUploadNoHeaderTpl":         "上傳的資料表 %s 中標頭遺失",
		"CustomChatCategory":                  "分類",
		"CustomChatQuestion":                  "問題",
		"CustomChatAnswer":                    "答案",
		"CustomChatExtend":                    "相似問",
		"CustomChatUploadQuestionRowInvalidTpl":    "資料表 %s 中第 %d 行缺少欄位",
		"CustomChatUploadQuestionRowNoQuestionTpl":     "資料表 %s 中第 %d 行問題為空",
		"CustomChatUploadQuestionRowNoAnswerTpl": "資料表 %s 中第 %d 行答案為空",
		"CustomChatUploadQuestionRowNoExtendTpl": "資料表 %s 中第 %d 行相似問為空",
		"CustomChatExport": "導出閒聊",
		"CustomChatUploadQuestionExceedLimit": "資料表 %s 中第 %d 行問題超过%d",
		"CustomChatUploadAnswerExceedLimit": "資料表 %s 中第 %d 行答案超过%d",
	},
}

