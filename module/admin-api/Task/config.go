package Task

var (
	SheetName = map[string]string{
		"triggerPhrase":    "触发语句",
		"triggerIntent":    "触发意图",
		"entityCollecting": "信息收集",
		"responseMessage":  "回复语句",
		"nerMap":           "自定义实体",
	}
	SlotType = map[string]string{
		"整数": "integer",
		"姓氏": "last-name",
		"时间日期(粒度-时)(未来时间)": "time-hour-future",
	}
)
