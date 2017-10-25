package ApiError

var (
	errorMsgMap = map[int]string{
		0:   "success",
		-1:  "db error",
		-2:  "io error",
		101: "Uploaded file still processing",
		102: "File extension should be xlsx",
		103: "File size should smaller than 2MB",
		104: "Multicustomer is not available",
	}

	// SUCCESS
	SUCCESS = 0

	// DB error, like db disconnect, schema error, etc
	DB_ERROR = -1

	// IO error, like create dir fail, create file err, etc
	IO_ERROR = -2

	// Dictionary error: last file still running
	DICT_STILL_RUNNING = 101
	// Dictionary error: extension of uploaded file error
	DICT_FORMAT_ERROR = 102
	// Dictionary error: uploaded file too large
	DICT_SIZE_ERROR = 103
	// Dictionary error: multicustomer service is not available
	DICT_SERVICE_ERROR = 104
)

func GetErrorMsg(errno int) string {
	if errMsg, ok := errorMsgMap[errno]; ok {
		return errMsg
	}
	return ""
}
