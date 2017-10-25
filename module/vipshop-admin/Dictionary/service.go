package Dictionary

import (
	"fmt"
	"mime/multipart"
	"path"
	"time"

	"emotibot.com/emotigo/module/vipshop-admin/ApiError"
	"emotibot.com/emotigo/module/vipshop-admin/util"
)

// Check wordbank status, 0: success, 1: running, -1: fail
func CheckProcessStatus(appid string) (string, error) {
	status, err := GetProcessStatus(appid)
	if err != nil {
		return "", err
	}

	return status, nil
}

func GetDownloadMeta(appid string) (map[string]*DownloadMeta, error) {
	metas, err := GetLastTwoSuccess(appid)
	if err != nil {
		return nil, err
	}

	ret := make(map[string]*DownloadMeta)
	util.LogTrace.Printf("Get download meta: (%d) %+v", len(ret), metas)

	if len(metas) >= 1 {
		ret["lastFile"] = metas[0]
	}

	if len(metas) >= 2 {
		ret["currentFile"] = metas[1]
	}
	util.LogInfo.Printf("Transfor finish")

	return ret, nil
}

func CheckUploadFile(appid string, file multipart.File, info *multipart.FileHeader) (string, int, error) {
	// 1. check is uploaded file still running
	ret, err := GetProcessStatus(appid)
	if err != nil {
		return "", ApiError.DB_ERROR, err
	}
	if ret == string(StatusRunning) {
		return "", ApiError.DICT_STILL_RUNNING, nil
	}

	// 2. Upload extension, and size should > 0 and < 2 * 1024 * 1024
	// 3. save file in settings/appid/wordbank_YYYYMMDD.xlsx
	ext := path.Ext(info.Filename)
	util.LogTrace.Printf("upload file ext: [%s]", ext)
	if ext != ".xlsx" {
		InsertProcess(appid, StatusFail, info.Filename, "format error")
		return "", ApiError.DICT_FORMAT_ERROR, nil
	}

	now := time.Now()
	filename := fmt.Sprintf("wordbank_%s.xlsx", now.Format("20060102150405"))
	size, err := util.SaveDictionaryFile(appid, filename, file)
	if err != nil {
		InsertProcess(appid, StatusFail, filename, "io error")
		return "", ApiError.IO_ERROR, err
	}

	if size < 0 || size > 2*1024*1024 {
		InsertProcess(appid, StatusFail, filename, "size error")
		return "", ApiError.DICT_SIZE_ERROR, nil
	}

	// 4. insert to db the running file
	// Note: running record will be added from multicustomer, WTF

	return filename, ApiError.SUCCESS, nil
}
