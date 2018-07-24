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
		"时间日期":             "time",
		"时间日期(粒度-分)(未来时间)": "time-minute-future",
		"时间日期(粒度-时)(未来时间)": "time-hour-future",
		"时间日期(粒度-天)(未来时间)": "time-day-future",
		"整数":            "integer",
		"人数":            "person-number",
		"姓氏":            "last-name",
		"来电人姓氏":         "your-last-name",
		"联络人姓氏":         "his-last-name",
		"电子邮件":          "email",
		"大陆手机号码":        "mobile-mainland",
		"大陆固定电话号码+手机号码": "phone-mainland",
		"台湾固定电话号码+手机号码": "phone-taiwan",
		"是否":     "binary",
		"车牌号码":   "car-plate",
		"包厢还是大堂": "room-type",
		"宝宝椅":    "baby-chair",
		"是否排号":   "take-ticket",
		"特殊需求":   "other-require",
		"信用卡号":   "bank-card",
		"金钱":     "money",
	}
)
