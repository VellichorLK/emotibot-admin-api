package Dictionary

import (
	"fmt"

	"emotibot.com/emotigo/module/admin-api/ApiError"
	"emotibot.com/emotigo/module/admin-api/util"
)

func GetWordbank(appid string, id int) (*WordBank, error) {
	return getWordbank(appid, id)
}

// UpdateWordbank will add a update if wordbank is nil, or add wordbank
func UpdateWordbank(appid string, newWordBank *WordBank) (int, error) {
	err := updateWordbank(appid, newWordBank)
	if err != nil {
		return ApiError.DB_ERROR, err
	}
	return ApiError.SUCCESS, nil
}

// AddWordbank will add a category if wordbank is nil, or add wordbank
func AddWordbank(appid string, paths []string, newWordBank *WordBank) (int, error) {
	if newWordBank == nil {
		// Add dir
		exist, err := checkDirExist(appid, paths)
		if err != nil {
			return ApiError.DB_ERROR, err
		} else if exist {
			return ApiError.REQUEST_ERROR, fmt.Errorf("dir existed")
		}
		err = addWordbankDir(appid, paths)
		if err != nil {
			return ApiError.DB_ERROR, err
		}
	} else {
		// Add wordbank
		exist, err := checkWordbankExist(appid, paths, newWordBank.Name)
		if err != nil {
			return ApiError.DB_ERROR, err
		} else if exist {
			return ApiError.REQUEST_ERROR, fmt.Errorf("wordbank existed")
		}
		err = addWordbank(appid, paths, newWordBank)
		if err != nil {
			return ApiError.DB_ERROR, err
		}
	}
	return ApiError.SUCCESS, nil
}

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

func GetEntities(appid string) ([]*WordBank, error) {
	wordbanks, err := getEntities(appid)
	return wordbanks, err
}

func SyncWordbank(appid string) {
	wordbanks, err := getWordbankRows(appid)
	if err != nil {
		return
	}
	TriggerUpdateWordbank(appid, wordbanks)
}

func DeleteWordbankDir(appid string, paths []string) (int, error) {
	return deleteWordbankDir(appid, paths)
}
func DeleteWordbank(appid string, id int) error {
	return deleteWordbank(appid, id)
}

func GetWordbankRow(appid string, id int) (*WordBankRow, error) {
	return getWordbankRow(appid, id)
}

// GetWordbankV3 will get wordbank from new table
func GetWordbanksV3(appid string) (*WordBankClassV3, error) {
	return getWordbanksV3(appid)
}

func GetWordbankV3(appid string, id int) (*WordBankV3, error) {
	return getWordbankV3(appid, id)
}

func GetWordbankClassV3(appid string, id int) (*WordBankClassV3, error) {
	return getWordbankClassV3(appid, id)
}

func DeleteWordbankV3(appid string, id int) error {
	return deleteWordbankV3(appid, id)
}
func DeleteWordbankClassV3(appid string, id int) error {
	return deleteWordbankClassV3(appid, id)
}
