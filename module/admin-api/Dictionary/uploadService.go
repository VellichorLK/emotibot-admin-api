package Dictionary

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"path"
	"strings"
	"time"

	"emotibot.com/emotigo/module/admin-api/ApiError"
	"emotibot.com/emotigo/module/admin-api/util"
	"github.com/tealeg/xlsx"
)

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

func parseDictionaryFromXLSX(buf []byte) (ret []*WordBankRow, err error) {
	exceptSheetNum := 3
	xlsxFile, err := xlsx.OpenBinary(buf)
	if err != nil {
		return
	}

	sheets := xlsxFile.Sheets
	if sheets == nil {
		err = errors.New(util.Msg["SheetError"])
		return
	}
	if len(sheets) != exceptSheetNum {
		err = errors.New(util.Msg["SheetError"])
		return
	}

	sheet := sheets[exceptSheetNum-1]
	if sheet.Name != util.Msg["TemplateXLSXName"] {
		err = errors.New(util.Msg["SheetError"])
		return
	}

	ret = []*WordBankRow{}
	rows := sheet.Rows
	if rows == nil {
		err = errors.New(util.Msg["EmptyRows"])
		return
	}
	if len(rows) <= 1 {
		err = errors.New(util.Msg["EmptyRows"])
		return
	}

	for idx, row := range rows {
		wordbank := WordBankRow{}
		if row.Cells == nil {
			util.LogError.Printf("Cannot get cell from row %d\n", idx)
			continue
		}
		if len(row.Cells) == 0 {
			continue
		}
		rowCellStr := make([]string, len(row.Cells))
		for cellIdx, cell := range row.Cells {
			rowCellStr[cellIdx] = strings.TrimSpace(cell.Value)
		}
		if strings.Join(rowCellStr, "") == "" {
			util.LogTrace.Printf("Skip empty row %d\n", idx+1)
			continue
		}

		for cellIdx, value := range rowCellStr {
			switch cellIdx {
			case 0:
				wordbank.Level1 = value
			case 1:
				wordbank.Level2 = value
			case 2:
				wordbank.Level3 = value
			case 3:
				wordbank.Level4 = value
			case 4:
				wordbank.Name = strings.TrimSpace(value)
			case 5:
				wordbank.SimilarWords = strings.Replace(value, "，", ",", -1)
			case 6:
				wordbank.Answer = value
			}
		}
		ret = append(ret, &wordbank)
	}
	fillWordbanksRowValue(ret)
	result, errs := checkWordbanksRowValue(ret)
	if !result {
		err = errors.New(strings.Join(errs, "\n"))
	}
	ret = ret[1:]
	return
}

func fillWordbanksRowValue(wordbanks []*WordBankRow) {
	for idx := 1; idx < len(wordbanks); idx++ {
		last := wordbanks[idx-1]
		now := wordbanks[idx]
		// If levelN change, level(N+n) will reset, not inherit last row
		if now.Level1 != "" {
			continue
		} else {
			now.Level1 = last.Level1
		}
		if now.Level2 != "" {
			continue
		} else {
			now.Level2 = last.Level2
		}
		if now.Level3 != "" {
			continue
		} else {
			now.Level3 = last.Level3
		}
		if now.Level4 != "" {
			continue
		} else {
			now.Level4 = last.Level4
		}
	}
	return
}

func checkWordbanksRowValue(wordbanks []*WordBankRow) (bool, []string) {
	// TODO: check name length, similary word lenth and number
	pathErrList := []string{}
	level1ErrList := []string{}
	for idx, wordbank := range wordbanks {
		if idx == 0 {
			continue
		}
		// Name empty means directory only, not an error
		// if wordbank.Name == "" {
		// 	ret = append(ret, fmt.Sprintf("Row %d: empty name", idx+1))
		// 	continue
		// }
		if wordbank.Level1 != util.Msg["SensitiveWordbank"] &&
			wordbank.Level1 != util.Msg["ProperNounsWordbank"] {
			level1ErrList = append(level1ErrList, fmt.Sprintf("%d", idx+1))
			continue
		}

		checks := []*string{&wordbank.Level1, &wordbank.Level2, &wordbank.Level3, &wordbank.Level4}
		shouldBlank := false
		for _, check := range checks {
			if shouldBlank && *check != "" {
				pathErrList = append(pathErrList, fmt.Sprintf("%d", idx+1))
				break
			}
			if *check == "" {
				shouldBlank = true
			}
		}
	}
	ret := []string{}
	if len(pathErrList) > 0 {
		ret = append(ret, fmt.Sprintf("%s%s: %s",
			util.Msg["BelowRows"], util.Msg["DirectoryError"],
			strings.Join(pathErrList, ",")))
	}
	if len(level1ErrList) > 0 {
		ret = append(ret, fmt.Sprintf("%s%s: %s",
			util.Msg["BelowRows"], util.Msg["Level1Error"],
			strings.Join(level1ErrList, ",")))
	}
	if len(ret) > 0 {
		return false, ret
	}
	return true, nil
}

func SaveWordbankRows(appid string, wordbanks []*WordBankRow) error {
	return saveWordbankRows(appid, wordbanks)
}

func SaveWordbankToFile(appid string, wordbanks []*WordBankRow) (error, string, string) {
	wordLines := []string{}
	synonyms := []string{}

	for _, wordbank := range wordbanks {
		isSensitive := wordbank.Level1 == util.Msg["SensitiveWordbank"]
		if wordbank.Name == "" {
			continue
		}

		words := []string{}
		if wordbank.SimilarWords != "" {
			similars := strings.Split(wordbank.SimilarWords, ",")
			for _, s := range similars {
				words = append(words, strings.TrimSpace(s))
			}
		}
		words = append(words, wordbank.Name)
		for _, w := range words {
			if isSensitive {
				// strange word mgc for sensitive words
				wordLines = append(wordLines, fmt.Sprintf("%s\tmgc", w))
			} else {
				wordLines = append(wordLines, w)
			}
		}
		synonyms = append(synonyms, wordbank.ToString())
	}

	util.LogTrace.Printf("word files %s.txt\n%s\n", appid, strings.Join(wordLines, "\n"))
	util.LogTrace.Printf("synonym files %s_synonym.txt\n%s\n", appid, strings.Join(synonyms, "\n"))

	err, md5Words, md5Synonyms := util.SaveNLUFileFromEntity(appid, wordLines, synonyms)
	if err != nil {
		return err, "", ""
	}

	return nil, md5Words, md5Synonyms
}

func RecordDictionaryImportProcess(appid string, filename string, buf []byte, importErr error) {
	message := ""
	if importErr != nil {
		message = importErr.Error()
	}
	insertImportProcess(appid, filename, importErr == nil, message)
	insertEntityFile(appid, filename, buf)
}

func GetWordbankFile(appid string, filename string) ([]byte, error) {
	return getWordbankFile(appid, filename)
}

func TriggerUpdateWordbank(appid string, wordbanks []*WordBankRow) (err error) {
	// 1. save to local file which can be get from url
	err, md5Words, md5Synonyms := SaveWordbankToFile(appid, wordbanks)
	if err != nil {
		return
	}

	// 2. Update consul key
	// TODO: use relative to compose the url
	url := getEnvironment("INTERNAL_URL")
	if url == "" {
		url = defaultInternalURL
	}
	now := time.Now()
	consulJSON := map[string]interface{}{
		"url":         fmt.Sprintf("%s/Files/settings/%s/%s.txt", url, appid, appid),
		"md5":         md5Words,
		"synonym-url": fmt.Sprintf("%s/Files/settings/%s/%s_synonyms.txt", url, appid, appid),
		"synonym-md5": md5Synonyms,
		"timestamp":   now.UnixNano() / 1000000,
	}
	util.ConsulUpdateEntity(appid, consulJSON)
	util.LogInfo.Printf("Update to consul:\n%+v\n", consulJSON)
	return
}
