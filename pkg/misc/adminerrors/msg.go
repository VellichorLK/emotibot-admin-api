package adminerrors

var msg = map[string]map[int]string{
	"zh-cn": map[int]string{
		ErrnoDBError:       "资料库错误",
		ErrnoIOError:       "IO错误",
		ErrnoRequestError:  "请求参数错误",
		ErrnoConsulService: "呼叫Consul服务失败",
		ErrnoOpenAPI:       "呼叫OpenAPI失败",
		ErrnoJSONParse:     "解析JSON格式错误",
		ErrnoAPIError:      "呼叫外部API错误",
		ErrnoNotFound:      "资源不存在",
		ErrnoBase64Decode:  "解析Base64错误",
	},
	"zh-tw": map[int]string{
		ErrnoDBError:       "資料庫錯誤",
		ErrnoIOError:       "IO錯誤",
		ErrnoRequestError:  "請求參數錯誤",
		ErrnoConsulService: "呼叫Consul服務失敗",
		ErrnoOpenAPI:       "呼叫OpenAPI失敗",
		ErrnoJSONParse:     "解析JSON格式錯誤",
		ErrnoAPIError:      "呼叫外部API錯誤",
		ErrnoNotFound:      "資源不存在",
		ErrnoBase64Decode:  "解析Base64錯誤",
	},
}

var unknownMsg = map[string]string{
	"zh-cn": "未知错误",
	"zh-tw": "未知錯誤",
}
