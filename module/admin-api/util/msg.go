package util

var (
	Msg = map[string]string{
		"To":             "到",
		"Move":           "移动",
		"DownloadFile":   "下载",
		"UploadFile":     "上传",
		"Delete":         "删除",
		"Deleted":        "已删除",
		"Modify":         "修改",
		"User":           "用户",
		"Role":           "角色",
		"Add":            "新增",
		"Size":           "档案大小",
		"Error":          "错误",
		"Systen":         "系统",
		"Read":           "读取",
		"Save":           "写入",
		"File":           "档案",
		"Format":         "格式",
		"Server":         "伺服器",
		"Not":            "非",
		"Cannot":         "无法",
		"Has":            "拥有",
		"Success":        "成功",
		"Open":           "开启",
		"Close":          "关闭",
		"Request":        "请求",
		"Rebuild":        "重新建模",
		"RobotProfile":   "机器人形象",
		"Question":       "问题",
		"RelateQuestion": "相关问题",
		"Origin":         "原始",
		"Updated":        "更新后",
		"Category":       "分类",
		"Row":            "列",
		"IDError":        "ID错误",
		"Name":           "名称",
		"Content":        "内容",
		"Start":          "开始",
		"ParseError":     "档案格式错误",
		"MarshalError":   "JSON转换错误",
		"ServerError":    "伺服器错误",

		"NotExistTemplate":   "%s不存在",
		"DuplicateTemplate":  "已存在相同%s的%s",
		"BadRequestTemplate": "无效的栏位：%s",

		// General Error
		"ErrorReadFileError":   "读取档案失败",
		"ErrorUploadEmptyFile": "上传档案为大小为0",

		// Used in robot words
		"RobotWords": "话术",

		// Used in Switch
		"Status":   "状态",
		"Remark":   "描述",
		"Scenario": "业务场景",
		"Num":      "次数配置",
		"Msg":      "文案",

		// Used in Robot Skill
		"function_weathermodule":     "天气模块",
		"function_computationmodule": "数学计算模块",
		"function_jokemodule":        "笑话模块",
		"function_storymodule":       "讲故事模块",
		"function_riddlemodule":      "猜谜语模块",
		"function_chengyumodule":     "成语模块",

		// Used in FAQ
		"DeleteCategoryDesc": "一并删除分类内 %d 个标准问题",

		// Wordbank
		"Wordbank":               "词库",
		"WordbankDir":            "词库目录",
		"SimilarWord":            "相似词",
		"Answer":                 "答案",
		"TemplateXLSXName":       "词库模板",
		"SensitiveWordbank":      "敏感词库",
		"ProperNounsWordbank":    "专有词库",
		"SheetError":             "获取词库模板资料表错误",
		"EmptyRows":              "资料表中无资料",
		"DirectoryError":         "目录错误",
		"Level1Error":            "第一层目录错误",
		"BelowRows":              "以下行数",
		"ErrorEmptyNameTpl":      "行 %d: 词库名为空",
		"ErrorNameTooLongTpl":    "行 %d: 词库名超过35字",
		"ErrorSimilarTooLongTpl": "行 %d: 同义词超过35字",
		"ErrorPathTooLongTpl":    "行 %d: 目錄名超过20字",
		"ErrorRowErrorTpl":       "行 %d：%s",
		"ErrorPathLevelTpl":      "路径 %d 级内容错误",
		"ErrorNotEditable":       "该词库不可编辑",

		// Used in Command and command-class
		"Cmd":                    "指令",
		"CmdClass":               "指令目录",
		"CmdClassName":           "指令目录名称",
		"CmdParentID":            "指令目录ID",
		"ErrorCmdParentNotFound": "目标目录获取失败",

		// Used in task-engine
		"MappingTable":       "转换数据",
		"MappingTableName":   "转换数据名称",
		"TaskEngineScenario": "任务引擎场景",
		"Spreadsheet":        "Spreadsheet场景",

		// Used in robot profile
		"RobotProfileQuestion":              "形象问题",
		"RobotProfileAnswer":                "形象问题答案",
		"RobotProfileRelateQuestion":        "形象问题相关问",
		"AddRobotProfileAnswerTemplate":     "新增形象问题 [%s] 的答案",
		"EditRobotProfileAnswerTemplate":    "编辑形象问题 [%s] 的答案",
		"DelRobotProfileAnswerTemplate":     "删除形象问题 [%s] 的答案",
		"AddRobotProfileRQuestionTemplate":  "新增形象问题 [%s] 的相似问",
		"EditRobotProfileRQuestionTemplate": "编辑形象问题 [%s] 的相似问",
		"DelRobotProfileRQuestionTemplate":  "删除形象问题 [%s] 的相似问",
	}
	ModuleName = map[string]string{
		"user-manage":        "用户管理",
		"role-manage":        "角色管理",
		"task-engine":        "场景编辑",
		"task-engine-upload": "上传转换数据",
		"statistic-dash":     "统计概览",
		"statistic-daily":    "日志管理",
		"statistic-analysis": "统计分析管理",
		"statistic-audit":    "安全日志",
		"qalist":             "问答库",
		"qatest":             "对话测试",
		"dictionary":         "词库管理",
		"robot-profile":      "形象设置",
		"robot-skill":        "技能设置",
		"robot-chat":         "话术设置",
		"switch-manage":      "开关管理",
	}
)
