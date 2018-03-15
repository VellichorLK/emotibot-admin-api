package Dictionary

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"path"
	"time"

	"github.com/tealeg/xlsx"

	"emotibot.com/emotigo/module/vipshop-admin/ApiError"
	"emotibot.com/emotigo/module/vipshop-admin/util"
)

// CheckProcessStatus will Check wordbank status
func CheckProcessStatus(appid string) (string, error) {
	status, err := getProcessStatus(appid)
	if err != nil {
		return "", err
	}

	return status, nil
}

// CheckFullProcessStatus will return full wordbank status
func CheckFullProcessStatus(appid string) (*StatusInfo, error) {
	status, err := getFullProcessStatus(appid)
	if err != nil {
		return nil, err
	}

	return status, nil
}

// GetDownloadMeta will return latest two success process status
func GetDownloadMeta(appid string) (map[string]*DownloadMeta, error) {
	metas, err := getLastTwoSuccess(appid)
	if err != nil {
		return nil, err
	}

	ret := make(map[string]*DownloadMeta)
	util.LogTrace.Printf("Get download meta: (%d) %+v", len(ret), metas)

	if len(metas) >= 1 {
		ret["currentFile"] = metas[0]
	}

	if len(metas) >= 2 {
		ret["lastFile"] = metas[1]
	}
	util.LogInfo.Printf("Transfor finish")

	return ret, nil
}

func CheckUploadFile(appid string, file multipart.File, info *multipart.FileHeader) (string, int, error) {
	// 1. check is uploaded file still running
	ret, err := getProcessStatus(appid)
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
		errMsg := fmt.Sprintf("%s%s%s", util.Msg["File"], util.Msg["Format"], util.Msg["Error"])
		insertProcess(appid, StatusFail, info.Filename, errMsg)
		return "", ApiError.DICT_FORMAT_ERROR, errors.New(errMsg)
	}

	now := time.Now()
	filename := fmt.Sprintf("wordbank_%s.xlsx", now.Format("20060102150405"))
	size, err := util.SaveDictionaryFile(appid, filename, file)
	if err != nil {
		errMsg := fmt.Sprintf("%s%s%s", util.Msg["Save"], util.Msg["File"], util.Msg["Error"])
		insertProcess(appid, StatusFail, filename, errMsg)
		util.LogError.Printf("save dict io error: %s", err.Error())
		return "", ApiError.IO_ERROR, errors.New(errMsg)
	}

	if size < 0 || size > 2*1024*1024 {
		errMsg := fmt.Sprintf("%s%s%s", util.Msg["File"], util.Msg["Size"], util.Msg["Error"])
		insertProcess(appid, StatusFail, info.Filename, errMsg)
		return "", ApiError.DICT_SIZE_ERROR, errors.New(errMsg)
	}

	// Check if xlsx format is correct or not
	file.Seek(0, io.SeekStart)
	buf := make([]byte, size)
	if _, err := file.Read(buf); err != nil {
		errMsg := fmt.Sprintf("%s%s%s", util.Msg["Read"], util.Msg["File"], util.Msg["Error"])
		insertProcess(appid, StatusFail, filename, errMsg)
		util.LogError.Printf("read dict io error: %s", err.Error())
		return "", ApiError.IO_ERROR, errors.New(errMsg)
	}
	_, err = xlsx.OpenBinary(buf)
	if err != nil {
		errMsg := fmt.Sprintf("%s%s%s, %s xlsx", util.Msg["File"], util.Msg["Format"], util.Msg["Error"], util.Msg["Not"])
		insertProcess(appid, StatusFail, info.Filename, errMsg)
		util.LogError.Printf("Not correct xlsx: %s", err.Error())
		return "", ApiError.DICT_FORMAT_ERROR, errors.New(errMsg)
	}
	// 4. insert to db the running file
	// Note: running record will be added from multicustomer, WTF

	return filename, ApiError.SUCCESS, nil
}

func GetEntities(appid string) ([]*WordBank, error) {
	wordbanks, err := getEntities(appid)
	return wordbanks, err
}
