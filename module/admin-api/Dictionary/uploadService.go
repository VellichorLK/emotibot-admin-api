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

func GetWordDataFromWordbanks(wordbanks []*WordBankRow) (error, []string, []string) {
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

	return nil, wordLines, synonyms
}

func SaveWordbankToFile(appid string, wordbanks []*WordBankRow) (error, string, string) {
	err, words, synonyms := GetWordDataFromWordbanks(wordbanks)
	util.LogTrace.Printf("word files %s.txt\n%s\n", appid, words)
	util.LogTrace.Printf("synonym files %s_synonym.txt\n%s\n", appid, synonyms)

	err, md5Words, md5Synonyms := util.SaveNLUFileFromEntity(appid, words, synonyms)
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
	err := insertEntityFile(appid, filename, buf)
	if err != nil {
		util.LogError.Println("Cannot write file into mysql: ", err.Error())
	}
}

func GetWordbankFile(appid string, filename string) ([]byte, error) {
	return getWordbankFile(appid, filename)
}

func TriggerUpdateWordbank(appid string, wordbanks []*WordBankRow, version int) (err error) {
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
		"url":         fmt.Sprintf("%s/api/v%d/dictionary/words/%s", url, version, appid),
		"md5":         md5Words,
		"synonym-url": fmt.Sprintf("%s/api/v%d/dictionary/synonyms/%s", url, version, appid),
		"synonym-md5": md5Synonyms,
		"timestamp":   now.UnixNano() / 1000000,
	}
	util.ConsulUpdateEntity(appid, consulJSON)
	util.LogInfo.Printf("Update to consul:\n%+v\n", consulJSON)
	return
}

func GetWordData(appid string) (error, []string, []string) {
	wordbanks, err := getWordbankRows(appid)
	if err != nil {
		return err, nil, nil
	}
	return GetWordDataFromWordbanks(wordbanks)
}

// below functions is for dictionary V3
func parseDictionaryFromXLSXV3(buf []byte) (root *WordBankClassV3, err error) {
	defer func() {
		if err != nil {
			util.LogError.Println("Parse xlsx fail: ", err.Error())
		}
	}()

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
	rows := sheet.Rows

	// Check if sheet is correct and format is correct
	switch {
	case sheet.Name != util.Msg["TemplateXLSXName"]:
		err = errors.New(util.Msg["SheetError"])
		return
	case rows == nil:
		err = errors.New(util.Msg["EmptyRows"])
		return
	case len(rows) <= 1:
		err = errors.New(util.Msg["EmptyRows"])
		return
	}
	rows = rows[1:]

	// classReadOny's key is class path
	classReadOny := map[string]bool{}
	// classWordbank store wordbanks in each path
	classWordbank := map[string][]*WordBankV3{}
	var lastWordbankRow *WordBankRow
	for idx, row := range rows {
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

		currentWordbankRow := &WordBankRow{}
		currentWordbankRow.ReadFromRow(rowCellStr)
		err = fillV3RowWithLast(currentWordbankRow, lastWordbankRow)
		if err != nil {
			err = fmt.Errorf("Invalid row %d, %s, %v", idx+1, err.Error(), rowCellStr)
			return
		}

		path := currentWordbankRow.GetPath()
		if _, ok := classReadOny[path]; !ok {
			isReadOnly := currentWordbankRow.IsReadOnly()
			classReadOny[path] = isReadOnly
		}

		if currentWordbankRow.Name != "" {
			wordbank := &WordBankV3{
				Name:   currentWordbankRow.Name,
				Answer: currentWordbankRow.Answer,
			}
			if currentWordbankRow.SimilarWords != "" {
				wordbank.SimilarWords = strings.Split(currentWordbankRow.SimilarWords, ",")
			}

			if _, ok := classWordbank[path]; !ok {
				classWordbank[path] = []*WordBankV3{}
			}
			classWordbank[path] = append(classWordbank[path], wordbank)
		}

		lastWordbankRow = currentWordbankRow
	}

	// log for debug
	util.LogTrace.Println("classReadOny:")
	for path, readOnly := range classReadOny {
		util.LogTrace.Printf("\t%s: %t\n", path, readOnly)
	}
	util.LogTrace.Println("classWordbank:")
	for path, wbs := range classWordbank {
		util.LogTrace.Printf("\t%s: \n", path)
		for _, wb := range wbs {
			util.LogTrace.Printf("\t\t%+v\n", *wb)
		}
	}

	root, err = createV3ObjsFromParseContent(classReadOny, classWordbank)
	return
}

func fillV3RowWithLast(current *WordBankRow, last *WordBankRow) error {
	if current == nil {
		return errors.New("Invalid func param")
	}
	// If levelN change, level(N+n) will reset,
	// not inherit from last row
	for last != nil {
		if current.Level1 != "" {
			break
		} else {
			current.Level1 = last.Level1
		}
		if current.Level2 != "" {
			break
		} else {
			current.Level2 = last.Level2
		}
		if current.Level3 != "" {
			break
		} else {
			current.Level3 = last.Level3
		}
		if current.Level4 != "" {
			break
		} else {
			current.Level4 = last.Level4
		}
		break
	}

	// check row's path is valid or not
	// if LevelN is empty, Level(N+n) must be empty
	// Level 1 cannot be empty
	shouldBeBlank := false
	hasErr := false
	level := 1
	for true {
		if current.Level1 == "" {
			hasErr = true
			break
		}

		level = 2
		if current.Level2 == "" {
			shouldBeBlank = true
		}

		level = 3
		if shouldBeBlank && (current.Level3 != "") {
			hasErr = true
			break
		} else if current.Level3 == "" {
			shouldBeBlank = true
		}

		level = 4
		if shouldBeBlank && (current.Level4 != "") {
			hasErr = true
			break
		} else if current.Level4 == "" {
			shouldBeBlank = true
		}
		break
	}
	if hasErr {
		util.LogTrace.Printf("Check for %#v\n", current)
		return fmt.Errorf("Invalid path in level %d", level)
	}

	return nil
}

func createV3ObjsFromParseContent(classReadOny map[string]bool, classWordbank map[string][]*WordBankV3) (*WordBankClassV3, error) {
	root := &WordBankClassV3{
		ID:           -1,
		Name:         "",
		Wordbank:     []*WordBankV3{},
		Children:     []*WordBankClassV3{},
		Editable:     false,
		IntentEngine: true,
		RuleEngine:   true,
	}
	classMap := map[string]*WordBankClassV3{}
	for classPath := range classReadOny {
		// Check form root to self
		names := strings.Split(classPath, "/")
		currentClass := root
		for idx, name := range names {
			selfPath := strings.Join(names[0:idx+1], "/")

			if class, ok := classMap[selfPath]; ok {
				currentClass = class
				continue
			}

			editable := true
			if val, ok := classReadOny[selfPath]; ok && val {
				editable = false
			}
			newWordbankClass := &WordBankClassV3{
				Name:         name,
				Editable:     editable,
				Wordbank:     []*WordBankV3{},
				Children:     []*WordBankClassV3{},
				IntentEngine: true,
				RuleEngine:   true,
			}
			classMap[selfPath] = newWordbankClass
			currentClass.Children = append(currentClass.Children, newWordbankClass)
			currentClass = classMap[selfPath]
		}
	}

	for classPath, wordbanks := range classWordbank {
		class, ok := classMap[classPath]
		if !ok {
			util.LogError.Printf("Wordbank's class not exist: %s", classPath)
			continue
		}

		for _, wordbank := range wordbanks {
			class.Wordbank = append(class.Wordbank, &WordBankV3{
				Name:         wordbank.Name,
				SimilarWords: wordbank.SimilarWords,
				Answer:       wordbank.Answer,
			})
		}
	}

	return root, nil
}

func SaveWordbankV3Rows(appid string, root *WordBankClassV3) error {
	return saveWordbankV3Rows(appid, root)
}

func GetWordDataV3(appid string) (error, []string, []string) {
	root, err := getWordbanksV3(appid)
	if err != nil {
		return err, nil, nil
	}
	return GetWordDataFromWordbanksV3(root)
}

func GetWordDataFromWordbanksV3(root *WordBankClassV3) (error, []string, []string) {
	var getClassData func(class *WordBankClassV3, path []string) (words []string, synonyms []string)
	getClassData = func(class *WordBankClassV3, path []string) (words []string, synonyms []string) {
		words = []string{}
		synonyms = []string{}
		if class == nil {
			return
		}

		util.LogTrace.Printf("Get data of [%s] in [%s]\n", class.Name, strings.Join(path, "/"))

		inSensitive := false
		if len(path) > 0 && path[0] == util.Msg["SensitiveWordbank"] {
			inSensitive = true
		}

		for _, child := range class.Children {
			var childWords []string
			var childSynonyms []string
			if class.ID == -1 {
				// path didn't need append when current class is virtual root
				childWords, childSynonyms = getClassData(child, path)
			} else {
				childWords, childSynonyms = getClassData(child, append(path, class.Name))
			}
			words = append(words, childWords...)
			synonyms = append(synonyms, childSynonyms...)
		}

		for _, wordbank := range class.Wordbank {
			util.LogTrace.Printf("Handle wordbank [%s]\n", wordbank.Name)
			for _, word := range wordbank.SimilarWords {
				var w string
				if inSensitive {
					w = fmt.Sprintf("%s\tmgc", word)
				} else {
					w = word
				}
				words = append(words, w)
			}

			// hack for now, the format need to fix with NLU
			synonymPaths := make([]string, 4)
			for idx := range synonymPaths {
				switch {
				case idx < len(path):
					synonymPaths[idx] = path[idx]
				case idx == len(path):
					synonymPaths[idx] = class.Name
				default:
					break
				}
			}
			synonymLine := fmt.Sprintf("%s\t%s\t%s",
				strings.Join(synonymPaths, ">"), wordbank.Name,
				strings.Join(wordbank.SimilarWords, ","))
			synonyms = append(synonyms, synonymLine)
		}
		return
	}

	words, syonyms := getClassData(root, []string{})

	return nil, words, syonyms
}
