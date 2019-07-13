package BF

const (
	ZhCn = "zh-cn"
	ZhTw = "zh-tw"
	MaxFileSize = 10 * 1024 *1024
	MaxFileLength = 10000

	IDLength = 16
    CmdFileMinioNamespace = "command"

	CmdPeriodFormatRegex = `^\d{4}(\-|\/|\.)\d{1,2}(\-|\/|\.)\d{1,2}-\d{4}(\-|\/|\.)\d{1,2}(\-|\/|\.)\d{1,2}$`

	StatusOk = 0
	StatusFileOpenError = 100
	StatusSizeExceed = 101
	StatusFileFormat = 102
	StatusTemplateFormat = 103
	StatusRowsExceed = 104


	RecordStatusOk = 0
	RecordStatusClassExceedMax = 200
	RecordStatusNameExceedMax = 201
	RecordStatusNameError = 202
	RecordStatusTargetError = 203
	RecordStatusPriodError = 204
	RecordStatusResponseError = 205
)

var headerTitle = map[string]map[int]string{
	ZhCn: map[int]string{
		0:            "分类",
		1:            "指令名称",
		2:            "规则适用对象",
		3:            "关联标签",
		4:            "关键字",
		5:            "正则式",
		6:            "生效时间",
		7:            "回复内容",
		8:            "回复位置",
	},
	ZhTw: map[int]string{
		0:            "分類",
		1:            "指令名稱",
		2:            "規則適用對象",
		3:            "關聯標簽",
		4:            "關鍵字",
		5:            "正則式",
		6:            "生效時間",
		7:            "回復內容",
		8:            "回復位置",
	},
}
